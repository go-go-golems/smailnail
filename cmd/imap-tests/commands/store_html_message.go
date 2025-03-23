package commands

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-message/mail"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	smailnail_imap "github.com/go-go-golems/smailnail/pkg/imap"
	"github.com/rs/zerolog/log"
)

type StoreHTMLMessageCommand struct {
	*cmds.CommandDescription
}

type StoreHTMLMessageSettings struct {
	From     string `glazed.parameter:"from"`
	To       string `glazed.parameter:"to"`
	Subject  string `glazed.parameter:"subject"`
	TextBody string `glazed.parameter:"text-body"`
	HTMLBody string `glazed.parameter:"html-body"`

	// IMAP flags
	Seen     bool `glazed.parameter:"seen"`
	Flagged  bool `glazed.parameter:"flagged"`
	Answered bool `glazed.parameter:"answered"`
	Draft    bool `glazed.parameter:"draft"`
	Deleted  bool `glazed.parameter:"deleted"`

	// IMAP settings
	smailnail_imap.IMAPSettings
}

func NewStoreHTMLMessageCommand() (*StoreHTMLMessageCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed parameter layer: %w", err)
	}

	imapLayer, err := smailnail_imap.NewIMAPParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("failed to create IMAP layer: %w", err)
	}

	return &StoreHTMLMessageCommand{
		CommandDescription: cmds.NewCommandDescription(
			"store-html-message",
			cmds.WithShort("Store an HTML email message in an IMAP mailbox"),
			cmds.WithLong("This command creates an HTML email message with plain text alternative and stores it in an IMAP mailbox"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"from",
					parameters.ParameterTypeString,
					parameters.WithHelp("Sender email address"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"to",
					parameters.ParameterTypeString,
					parameters.WithHelp("Recipient email address"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"subject",
					parameters.ParameterTypeString,
					parameters.WithHelp("Email subject"),
					parameters.WithDefault("Test HTML email"),
				),
				parameters.NewParameterDefinition(
					"text-body",
					parameters.ParameterTypeString,
					parameters.WithHelp("Plain text email body content"),
					parameters.WithDefault("This is a test email sent using smailnail."),
				),
				parameters.NewParameterDefinition(
					"html-body",
					parameters.ParameterTypeString,
					parameters.WithHelp("HTML email body content"),
					parameters.WithDefault("<html><body><h1>Test Email</h1><p>This is a <strong>test email</strong> sent using smailnail.</p></body></html>"),
				),
				// IMAP flags
				parameters.NewParameterDefinition(
					"seen",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Mark message as seen"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"flagged",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Mark message as flagged"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"answered",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Mark message as answered"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"draft",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Mark message as draft"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"deleted",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Mark message as deleted"),
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

func (c *StoreHTMLMessageCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &StoreHTMLMessageSettings{}
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
	h.SetAddressList("From", []*mail.Address{{Address: from}})
	h.SetAddressList("To", []*mail.Address{{Address: to}})
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
