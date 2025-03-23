package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
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
	NewMailbox string `glazed.parameter:"new-mailbox"`
	Force      bool   `glazed.parameter:"force"`
	
	// IMAP settings
	imap.IMAPSettings
}

func NewCreateMailboxCommand() (*CreateMailboxCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed parameter layer: %w", err)
	}

	imapLayer, err := imap.NewIMAPParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("failed to create IMAP layer: %w", err)
	}

	return &CreateMailboxCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create-mailbox",
			cmds.WithShort("Create a new mailbox on an IMAP server"),
			cmds.WithLong("This command creates a new mailbox on an IMAP server if it doesn't already exist"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"new-mailbox",
					parameters.ParameterTypeString,
					parameters.WithHelp("Name of the mailbox to create"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"force",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Force creation even if mailbox already exists (will delete and recreate)"),
					parameters.WithDefault(false),
				),
			),
			cmds.WithLayersList(
				glazedParameterLayer,
				imapLayer,
			),
		),
	}, nil
}

func (c *CreateMailboxCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &CreateMailboxSettings{}
	if err := parsedLayers.InitializeStruct("default", settings); err != nil {
		return err
	}
	if err := parsedLayers.InitializeStruct("imap", &settings.IMAPSettings); err != nil {
		return err
	}

	// Check if password is provided
	if settings.Password == "" {
		return fmt.Errorf("password is required (provide via --password flag or IMAP_PASSWORD environment variable)")
	}

	// Connect to IMAP server
	log.Debug().Msg("Connecting to IMAP server")
	client, err := settings.IMAPSettings.ConnectToIMAPServer()
	if err != nil {
		return fmt.Errorf("error connecting to IMAP server: %w", err)
	}
	defer client.Close()

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