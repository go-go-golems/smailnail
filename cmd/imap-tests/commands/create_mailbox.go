package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/smailnail/pkg/imap"
	"github.com/rs/zerolog/log"
)

type CreateMailboxCommand struct {
	*cmds.CommandDescription
}

type CreateMailboxSettings struct {
	NewMailbox string `glazed:"new-mailbox"`
	Force      bool   `glazed:"force"`

	// IMAP settings
	imap.IMAPSettings
}

func NewCreateMailboxCommand() (*CreateMailboxCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed section: %w", err)
	}

	imapSection, err := imap.NewIMAPSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create IMAP section: %w", err)
	}

	return &CreateMailboxCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create-mailbox",
			cmds.WithShort("Create a new mailbox on an IMAP server"),
			cmds.WithLong("This command creates a new mailbox on an IMAP server if it doesn't already exist"),
			cmds.WithFlags(
				fields.New(
					"new-mailbox",
					fields.TypeString,
					fields.WithHelp("Name of the mailbox to create"),
					fields.WithRequired(true),
				),
				fields.New(
					"force",
					fields.TypeBool,
					fields.WithHelp("Force creation even if mailbox already exists (will delete and recreate)"),
					fields.WithDefault(false),
				),
			),
			cmds.WithSections(
				glazedSection,
				imapSection,
			),
		),
	}, nil
}

func (c *CreateMailboxCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedValues *values.Values,
	gp middlewares.Processor,
) error {
	settings := &CreateMailboxSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return err
	}
	if err := parsedValues.DecodeSectionInto(imap.IMAPSectionSlug, &settings.IMAPSettings); err != nil {
		return err
	}

	// Check if password is provided
	if settings.Password == "" {
		return fmt.Errorf("password is required (provide via --password flag or IMAP_PASSWORD environment variable)")
	}

	// Connect to IMAP server
	log.Debug().Msg("Connecting to IMAP server")
	client, err := settings.ConnectToIMAPServer()
	if err != nil {
		return fmt.Errorf("error connecting to IMAP server: %w", err)
	}
	defer func() {
		_ = client.Close()
	}()

	// First try to list mailboxes to see if it already exists
	log.Debug().Msg("Listing existing mailboxes")
	existingMailboxes, err := listMailboxes(client)
	if err != nil {
		return fmt.Errorf("error listing mailboxes: %w", err)
	}

	// Check if mailbox already exists
	mailboxExists := false
	for _, name := range existingMailboxes {
		if name == settings.NewMailbox {
			mailboxExists = true
			break
		}
	}

	// Handle existing mailbox
	if mailboxExists {
		if settings.Force {
			// Delete the existing mailbox
			log.Debug().Str("mailbox", settings.NewMailbox).Msg("Deleting existing mailbox")
			if err := client.Delete(settings.NewMailbox).Wait(); err != nil {
				return fmt.Errorf("error deleting existing mailbox: %w", err)
			}
		} else {
			// Output information about existing mailbox
			row := types.NewRow(
				types.MRP("status", "skipped"),
				types.MRP("server", settings.Server),
				types.MRP("mailbox", settings.NewMailbox),
				types.MRP("reason", "Mailbox already exists"),
				types.MRP("existing_mailboxes", existingMailboxes),
				types.MRP("timestamp", time.Now().Format(time.RFC3339)),
			)

			if err := gp.AddRow(ctx, row); err != nil {
				return fmt.Errorf("error adding row to output: %w", err)
			}

			return nil
		}
	}

	// Create the mailbox
	log.Debug().Str("mailbox", settings.NewMailbox).Msg("Creating mailbox")
	if err := client.Create(settings.NewMailbox, nil).Wait(); err != nil {
		return fmt.Errorf("error creating mailbox: %w", err)
	}

	action := "created"
	if mailboxExists {
		action = "recreated"
	}

	// Output success information
	row := types.NewRow(
		types.MRP("status", "success"),
		types.MRP("server", settings.Server),
		types.MRP("mailbox", settings.NewMailbox),
		types.MRP("action", action),
		types.MRP("existing_mailboxes", existingMailboxes),
		types.MRP("timestamp", time.Now().Format(time.RFC3339)),
	)

	if err := gp.AddRow(ctx, row); err != nil {
		return fmt.Errorf("error adding row to output: %w", err)
	}

	return nil
}

// Helper function to list all mailboxes
func listMailboxes(client *imapclient.Client) ([]string, error) {
	var mailboxes []string

	cmd := client.List("", "*", nil)
	for {
		mailbox := cmd.Next()
		if mailbox == nil {
			break
		}
		mailboxes = append(mailboxes, mailbox.Mailbox)
	}

	if err := cmd.Close(); err != nil {
		return nil, err
	}

	return mailboxes, nil
}
