package cmds

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/mail"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	smailnail_imap "github.com/go-go-golems/smailnail/pkg/imap"
	"github.com/go-go-golems/smailnail/pkg/mailgen"
	"github.com/go-go-golems/smailnail/pkg/mailutil"
	mailgenTypes "github.com/go-go-golems/smailnail/pkg/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// GenerateCommand represents the generate command
type GenerateCommand struct {
	*cmds.CommandDescription
}

// NewGenerateCommand creates a new generate command
func NewGenerateCommand() (*GenerateCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create glazed section")
	}

	imapSection, err := smailnail_imap.NewIMAPSection()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create IMAP section")
	}

	return &GenerateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"generate",
			cmds.WithShort("Generate emails from template"),
			cmds.WithLong("Generate emails from template using YAML configuration"),
			cmds.WithSections(glazedSection, imapSection),
			cmds.WithFlags(
				fields.New(
					"configs",
					fields.TypeStringList,
					fields.WithHelp("Path to YAML config files"),
					fields.WithRequired(true),
				),
				fields.New(
					"output-dir",
					fields.TypeString,
					fields.WithHelp("Directory to output generated emails"),
					fields.WithDefault("./output"),
				),
				fields.New(
					"write-files",
					fields.TypeBool,
					fields.WithHelp("Write emails to files"),
					fields.WithDefault(false),
				),
				fields.New(
					"store-imap",
					fields.TypeBool,
					fields.WithHelp("Store generated emails in IMAP server"),
					fields.WithDefault(false),
				),
			),
		),
	}, nil
}

type GenerateSettings struct {
	ConfigFile []string `glazed:"configs"`
	OutputDir  string   `glazed:"output-dir"`
	WriteFiles bool     `glazed:"write-files"`
	StoreIMAP  bool     `glazed:"store-imap"`
	smailnail_imap.IMAPSettings
}

// RunIntoGlazeProcessor generates emails and outputs them as structured data
func (c *GenerateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedValues *values.Values,
	gp middlewares.Processor,
) error {
	// Parse command settings
	settings := &GenerateSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return err
	}
	if err := parsedValues.DecodeSectionInto(smailnail_imap.IMAPSectionSlug, &settings.IMAPSettings); err != nil {
		return err
	}

	log.Info().Msgf("Settings: %+v", settings)

	var allEmails []*mailgenTypes.Email

	log.Info().Msgf("Generating emails from %d config files", len(settings.ConfigFile))

	// Process each config file independently
	for _, configFile := range settings.ConfigFile {
		// Read and parse config file
		// #nosec G304 -- the CLI intentionally accepts user-specified config file paths.
		configData, err := os.ReadFile(configFile)
		if err != nil {
			return errors.Wrapf(err, "failed to read config file '%s'", configFile)
		}

		var config mailgenTypes.TemplateConfig
		if err := yaml.Unmarshal(configData, &config); err != nil {
			return errors.Wrapf(err, "failed to parse config file '%s'", configFile)
		}

		// Create mail generator for this config
		generator := mailgen.NewMailGenerator(&config)

		// Generate emails for this config
		emails, err := generator.Generate(ctx)
		if err != nil {
			return errors.Wrapf(err, "failed to generate emails from config file '%s'", configFile)
		}

		allEmails = append(allEmails, emails...)
	}

	// Create output directory if needed
	if settings.WriteFiles {
		if err := os.MkdirAll(settings.OutputDir, 0700); err != nil {
			return errors.Wrapf(err, "failed to create output directory '%s'", settings.OutputDir)
		}
	}

	// Connect to IMAP server if needed
	var imapClient *imapclient.Client
	if settings.StoreIMAP {
		var err error
		imapClient, err = settings.ConnectToIMAPServer()
		if err != nil {
			return errors.Wrap(err, "failed to connect to IMAP server")
		}
		defer func() {
			_ = imapClient.Close()
		}()

		// Select the target mailbox
		if _, err := imapClient.Select(settings.Mailbox, nil).Wait(); err != nil {
			return errors.Wrapf(err, "failed to select mailbox '%s'", settings.Mailbox)
		}
	}

	// Process all generated emails
	for i, email := range allEmails {
		// Create a glazed row for each email
		row := types.NewRow(
			types.MRP("index", i),
			types.MRP("subject", email.Subject),
			types.MRP("from", email.From),
			types.MRP("to", email.To),
			types.MRP("cc", email.Cc),
			types.MRP("bcc", email.Bcc),
			types.MRP("reply_to", email.ReplyTo),
			types.MRP("body", email.Body),
		)

		// Add row to processor
		if err := gp.AddRow(ctx, row); err != nil {
			return errors.Wrapf(err, "failed to process email %d", i)
		}

		// Write email to file if requested
		if settings.WriteFiles {
			fileName := fmt.Sprintf("email_%d.txt", i)
			filePath := filepath.Join(settings.OutputDir, fileName)

			// Format email as text
			emailText := fmt.Sprintf("Subject: %s\nFrom: %s\n", email.Subject, email.From)
			if email.To != "" {
				emailText += fmt.Sprintf("To: %s\n", email.To)
			}
			if email.Cc != "" {
				emailText += fmt.Sprintf("Cc: %s\n", email.Cc)
			}
			if email.Bcc != "" {
				emailText += fmt.Sprintf("Bcc: %s\n", email.Bcc)
			}
			if email.ReplyTo != "" {
				emailText += fmt.Sprintf("Reply-To: %s\n", email.ReplyTo)
			}
			emailText += fmt.Sprintf("\n%s", email.Body)

			// Write to file
			if err := os.WriteFile(filePath, []byte(emailText), 0600); err != nil {
				return errors.Wrapf(err, "failed to write email %d to file '%s'", i, filePath)
			}
		}

		// Store email in IMAP server if requested
		if settings.StoreIMAP {
			// Create IMAP message
			var buf bytes.Buffer

			// Create mail header
			h := mail.Header{}
			h.SetDate(time.Now())
			if err := mailutil.SetSingleAddress(&h, "From", email.From); err != nil {
				return errors.Wrapf(err, "failed to parse From address for email %d", i)
			}
			if email.To != "" {
				if err := mailutil.SetSingleAddress(&h, "To", email.To); err != nil {
					return errors.Wrapf(err, "failed to parse To address for email %d", i)
				}
			}
			if email.Cc != "" {
				if err := mailutil.SetSingleAddress(&h, "Cc", email.Cc); err != nil {
					return errors.Wrapf(err, "failed to parse Cc address for email %d", i)
				}
			}
			if email.Bcc != "" {
				if err := mailutil.SetSingleAddress(&h, "Bcc", email.Bcc); err != nil {
					return errors.Wrapf(err, "failed to parse Bcc address for email %d", i)
				}
			}
			if email.ReplyTo != "" {
				if err := mailutil.SetSingleAddress(&h, "Reply-To", email.ReplyTo); err != nil {
					return errors.Wrapf(err, "failed to parse Reply-To address for email %d", i)
				}
			}
			h.SetSubject(email.Subject)

			// Create message writer
			w, err := mail.CreateSingleInlineWriter(&buf, h)
			if err != nil {
				return errors.Wrapf(err, "failed to create message writer for email %d", i)
			}

			// Write body
			if _, err := w.Write([]byte(email.Body)); err != nil {
				return errors.Wrapf(err, "failed to write message body for email %d", i)
			}

			// Close writer
			if err := w.Close(); err != nil {
				return errors.Wrapf(err, "failed to close message writer for email %d", i)
			}

			messageData := buf.Bytes()

			// Prepare flags
			var flags []imap.Flag
			flags = append(flags, imap.FlagSeen)

			// Set the append options
			options := &imap.AppendOptions{
				Flags: flags,
				Time:  time.Now(),
			}

			// Create append command
			cmd := imapClient.Append(settings.Mailbox, int64(len(messageData)), options)

			// Write message data
			if _, err := cmd.Write(messageData); err != nil {
				return errors.Wrapf(err, "failed to write message data for email %d", i)
			}

			// Close writer
			if err := cmd.Close(); err != nil {
				return errors.Wrapf(err, "failed to close append command for email %d", i)
			}

			// Wait for command to complete
			if _, err := cmd.Wait(); err != nil {
				return errors.Wrapf(err, "failed to store email %d in IMAP server", i)
			}
		}
	}

	return nil
}
