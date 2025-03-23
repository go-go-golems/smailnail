package commands

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
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

type StoreAttachmentCommand struct {
	*cmds.CommandDescription
}

type StoreAttachmentSettings struct {
	From           string `glazed.parameter:"from"`
	To             string `glazed.parameter:"to"`
	Subject        string `glazed.parameter:"subject"`
	Body           string `glazed.parameter:"body"`
	AttachmentPath string `glazed.parameter:"attachment-path"`

	// IMAP flags
	Seen     bool `glazed.parameter:"seen"`
	Flagged  bool `glazed.parameter:"flagged"`
	Answered bool `glazed.parameter:"answered"`
	Draft    bool `glazed.parameter:"draft"`
	Deleted  bool `glazed.parameter:"deleted"`

	// IMAP settings
	smailnail_imap.IMAPSettings
}

func NewStoreAttachmentCommand() (*StoreAttachmentCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed parameter layer: %w", err)
	}

	imapLayer, err := smailnail_imap.NewIMAPParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("failed to create IMAP layer: %w", err)
	}

	return &StoreAttachmentCommand{
		CommandDescription: cmds.NewCommandDescription(
			"store-attachment",
			cmds.WithShort("Store an email message with attachment in an IMAP mailbox"),
			cmds.WithLong("This command creates an email message with an attachment and stores it in an IMAP mailbox"),
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
					parameters.WithDefault("Test email with attachment"),
				),
				parameters.NewParameterDefinition(
					"body",
					parameters.ParameterTypeString,
					parameters.WithHelp("Email body content"),
					parameters.WithDefault("This is a test email with an attachment sent using smailnail."),
				),
				parameters.NewParameterDefinition(
					"attachment-path",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to the file to attach"),
					parameters.WithRequired(true),
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

func (c *StoreAttachmentCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &StoreAttachmentSettings{}
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

	// Check if attachment file exists
	if _, err := os.Stat(settings.AttachmentPath); os.IsNotExist(err) {
		return fmt.Errorf("attachment file does not exist: %s", settings.AttachmentPath)
	}

	// Read attachment file
	fileContent, err := os.ReadFile(settings.AttachmentPath)
	if err != nil {
		return fmt.Errorf("error reading attachment file: %w", err)
	}

	// Determine content type (simplistic implementation)
	contentType := "application/octet-stream"
	ext := filepath.Ext(settings.AttachmentPath)
	switch ext {
	case ".txt":
		contentType = "text/plain"
	case ".pdf":
		contentType = "application/pdf"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".doc", ".docx":
		contentType = "application/msword"
	case ".xls", ".xlsx":
		contentType = "application/vnd.ms-excel"
	case ".zip":
		contentType = "application/zip"
	}

	// Connect to IMAP server
	log.Debug().Msg("Connecting to IMAP server")
	client, err := settings.IMAPSettings.ConnectToIMAPServer()
	if err != nil {
		return fmt.Errorf("error connecting to IMAP server: %w", err)
	}
	defer client.Close()

	// Create the message
	messageData, err := createMessageWithAttachment(
		settings.From,
		settings.To,
		settings.Subject,
		settings.Body,
		filepath.Base(settings.AttachmentPath),
		fileContent,
		contentType,
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
		types.MRP("body_length", len(settings.Body)),
		types.MRP("attachment", filepath.Base(settings.AttachmentPath)),
		types.MRP("attachment_size", len(fileContent)),
		types.MRP("attachment_type", contentType),
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
func createMessageWithAttachment(from, to, subject, body string,
	filename string, fileContent []byte, contentType string) ([]byte, error) {

	var buf bytes.Buffer

	// Create the mail header
	h := mail.Header{}
	h.SetDate(time.Now())
	h.SetAddressList("From", []*mail.Address{{Address: from}})
	h.SetAddressList("To", []*mail.Address{{Address: to}})
	h.SetSubject(subject)

	// Create the multipart message
	mw, err := mail.CreateWriter(&buf, h)
	if err != nil {
		return nil, err
	}

	// Create the text part
	th := mail.InlineHeader{}
	th.Set("Content-Type", "text/plain; charset=utf-8")
	tw, err := mw.CreateSingleInline(th)
	if err != nil {
		return nil, err
	}
	if _, err := tw.Write([]byte(body)); err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}

	// Create the attachment part
	ah := mail.AttachmentHeader{}
	ah.Set("Content-Type", contentType)
	ah.SetFilename(filename)
	aw, err := mw.CreateAttachment(ah)
	if err != nil {
		return nil, err
	}
	if _, err := aw.Write(fileContent); err != nil {
		return nil, err
	}
	if err := aw.Close(); err != nil {
		return nil, err
	}

	// Close the message writer
	if err := mw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
