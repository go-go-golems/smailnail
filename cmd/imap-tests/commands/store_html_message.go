package commands

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-message/mail"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	smailnail_imap "github.com/go-go-golems/smailnail/pkg/imap"
	"github.com/go-go-golems/smailnail/pkg/mailutil"
	"github.com/rs/zerolog/log"
)

type StoreHTMLMessageCommand struct {
	*cmds.CommandDescription
}

type StoreHTMLMessageSettings struct {
	From     string `glazed:"from"`
	To       string `glazed:"to"`
	Subject  string `glazed:"subject"`
	TextBody string `glazed:"text-body"`
	HTMLBody string `glazed:"html-body"`

	// IMAP flags
	Seen     bool `glazed:"seen"`
	Flagged  bool `glazed:"flagged"`
	Answered bool `glazed:"answered"`
	Draft    bool `glazed:"draft"`
	Deleted  bool `glazed:"deleted"`

	// IMAP settings
	smailnail_imap.IMAPSettings
}

func NewStoreHTMLMessageCommand() (*StoreHTMLMessageCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed section: %w", err)
	}

	imapSection, err := smailnail_imap.NewIMAPSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create IMAP section: %w", err)
	}

	return &StoreHTMLMessageCommand{
		CommandDescription: cmds.NewCommandDescription(
			"store-html-message",
			cmds.WithShort("Store an HTML email message in an IMAP mailbox"),
			cmds.WithLong("This command creates an HTML email message with plain text alternative and stores it in an IMAP mailbox"),
			cmds.WithFlags(
				fields.New(
					"from",
					fields.TypeString,
					fields.WithHelp("Sender email address"),
					fields.WithRequired(true),
				),
				fields.New(
					"to",
					fields.TypeString,
					fields.WithHelp("Recipient email address"),
					fields.WithRequired(true),
				),
				fields.New(
					"subject",
					fields.TypeString,
					fields.WithHelp("Email subject"),
					fields.WithDefault("Test HTML email"),
				),
				fields.New(
					"text-body",
					fields.TypeString,
					fields.WithHelp("Plain text email body content"),
					fields.WithDefault("This is a test email sent using smailnail."),
				),
				fields.New(
					"html-body",
					fields.TypeString,
					fields.WithHelp("HTML email body content"),
					fields.WithDefault("<html><body><h1>Test Email</h1><p>This is a <strong>test email</strong> sent using smailnail.</p></body></html>"),
				),
				// IMAP flags
				fields.New(
					"seen",
					fields.TypeBool,
					fields.WithHelp("Mark message as seen"),
					fields.WithDefault(false),
				),
				fields.New(
					"flagged",
					fields.TypeBool,
					fields.WithHelp("Mark message as flagged"),
					fields.WithDefault(false),
				),
				fields.New(
					"answered",
					fields.TypeBool,
					fields.WithHelp("Mark message as answered"),
					fields.WithDefault(false),
				),
				fields.New(
					"draft",
					fields.TypeBool,
					fields.WithHelp("Mark message as draft"),
					fields.WithDefault(false),
				),
				fields.New(
					"deleted",
					fields.TypeBool,
					fields.WithHelp("Mark message as deleted"),
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

func (c *StoreHTMLMessageCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedValues *values.Values,
	gp middlewares.Processor,
) error {
	settings := &StoreHTMLMessageSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return err
	}
	if err := parsedValues.DecodeSectionInto(smailnail_imap.IMAPSectionSlug, &settings.IMAPSettings); err != nil {
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

	// Create the message
	messageData, err := createHTMLMessage(
		settings.From,
		settings.To,
		settings.Subject,
		settings.TextBody,
		settings.HTMLBody,
	)
	if err != nil {
		return fmt.Errorf("error creating message: %w", err)
	}

	// Prepare flags
	var flags []imap.Flag
	if settings.Seen {
		flags = append(flags, imap.FlagSeen)
	}
	if settings.Flagged {
		flags = append(flags, imap.FlagFlagged)
	}
	if settings.Answered {
		flags = append(flags, imap.FlagAnswered)
	}
	if settings.Draft {
		flags = append(flags, imap.FlagDraft)
	}
	if settings.Deleted {
		flags = append(flags, imap.FlagDeleted)
	}

	// Store the message
	err = storeMessage(client, settings.Mailbox, messageData, flags, time.Now())
	if err != nil {
		return fmt.Errorf("error storing message: %w", err)
	}

	// Output success information
	row := types.NewRow(
		types.MRP("status", "success"),
		types.MRP("server", settings.Server),
		types.MRP("mailbox", settings.Mailbox),
		types.MRP("from", settings.From),
		types.MRP("to", settings.To),
		types.MRP("subject", settings.Subject),
		types.MRP("text_body_length", len(settings.TextBody)),
		types.MRP("html_body_length", len(settings.HTMLBody)),
		types.MRP("message_size", len(messageData)),
		types.MRP("flags", flags),
		types.MRP("timestamp", time.Now().Format(time.RFC3339)),
	)

	if err := gp.AddRow(ctx, row); err != nil {
		return fmt.Errorf("error adding row to output: %w", err)
	}

	return nil
}

// Helper function
func createHTMLMessage(from, to, subject, textBody, htmlBody string) ([]byte, error) {
	var buf bytes.Buffer

	// Create a new mail message
	h := mail.Header{}
	h.SetDate(time.Now())
	if err := mailutil.SetSingleAddress(&h, "From", from); err != nil {
		return nil, err
	}
	if err := mailutil.SetSingleAddress(&h, "To", to); err != nil {
		return nil, err
	}
	h.SetSubject(subject)

	// Create a multipart message with alternatives
	mw, err := mail.CreateWriter(&buf, h)
	if err != nil {
		return nil, err
	}

	// Create the alternative part
	altw, err := mw.CreateInline()
	if err != nil {
		return nil, err
	}

	// Add the plain text part
	th := mail.InlineHeader{}
	th.Set("Content-Type", "text/plain; charset=utf-8")
	tw, err := altw.CreatePart(th)
	if err != nil {
		return nil, err
	}
	if _, err := tw.Write([]byte(textBody)); err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}

	// Add the HTML part
	hh := mail.InlineHeader{}
	hh.Set("Content-Type", "text/html; charset=utf-8")
	hw, err := altw.CreatePart(hh)
	if err != nil {
		return nil, err
	}
	if _, err := hw.Write([]byte(htmlBody)); err != nil {
		return nil, err
	}
	if err := hw.Close(); err != nil {
		return nil, err
	}

	// Close the alternative and message writers
	if err := altw.Close(); err != nil {
		return nil, err
	}
	if err := mw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
