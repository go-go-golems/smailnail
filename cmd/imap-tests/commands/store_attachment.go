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

type StoreAttachmentCommand struct {
	*cmds.CommandDescription
}

type StoreAttachmentSettings struct {
	From           string `glazed:"from"`
	To             string `glazed:"to"`
	Subject        string `glazed:"subject"`
	Body           string `glazed:"body"`
	AttachmentPath string `glazed:"attachment-path"`

	// IMAP flags
	Seen     bool `glazed:"seen"`
	Flagged  bool `glazed:"flagged"`
	Answered bool `glazed:"answered"`
	Draft    bool `glazed:"draft"`
	Deleted  bool `glazed:"deleted"`

	// IMAP settings
	smailnail_imap.IMAPSettings
}

func NewStoreAttachmentCommand() (*StoreAttachmentCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed section: %w", err)
	}

	imapSection, err := smailnail_imap.NewIMAPSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create IMAP section: %w", err)
	}

	return &StoreAttachmentCommand{
		CommandDescription: cmds.NewCommandDescription(
			"store-attachment",
			cmds.WithShort("Store an email message with attachment in an IMAP mailbox"),
			cmds.WithLong("This command creates an email message with an attachment and stores it in an IMAP mailbox"),
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
					fields.WithDefault("Test email with attachment"),
				),
				fields.New(
					"body",
					fields.TypeString,
					fields.WithHelp("Email body content"),
					fields.WithDefault("This is a test email with an attachment sent using smailnail."),
				),
				fields.New(
					"attachment-path",
					fields.TypeString,
					fields.WithHelp("Path to the file to attach"),
					fields.WithRequired(true),
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

func (c *StoreAttachmentCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedValues *values.Values,
	gp middlewares.Processor,
) error {
	settings := &StoreAttachmentSettings{}
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
	if err := mailutil.SetSingleAddress(&h, "From", from); err != nil {
		return nil, err
	}
	if err := mailutil.SetSingleAddress(&h, "To", to); err != nil {
		return nil, err
	}
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
