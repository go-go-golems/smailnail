package imap

import (
	"crypto/tls"
	"fmt"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
)

// IMAPSettings represents the settings for connecting to an IMAP server
type IMAPSettings struct {
	Server   string `glazed:"server"`
	Port     int    `glazed:"port"`
	Username string `glazed:"username"`
	Password string `glazed:"password"`
	Mailbox  string `glazed:"mailbox"`
	Insecure bool   `glazed:"insecure"`
}

const IMAPSectionSlug = "imap"

// NewIMAPSection creates a new section for IMAP server settings.
func NewIMAPSection() (schema.Section, error) {
	return schema.NewSection(
		IMAPSectionSlug,
		"IMAP Server Connection Settings",
		schema.WithFields(
			fields.New(
				"server",
				fields.TypeString,
				fields.WithHelp("IMAP server address"),
			),
			fields.New(
				"port",
				fields.TypeInteger,
				fields.WithHelp("IMAP server port"),
				fields.WithDefault(993),
			),
			fields.New(
				"username",
				fields.TypeString,
				fields.WithHelp("IMAP username"),
			),
			fields.New(
				"password",
				fields.TypeString,
				fields.WithHelp("IMAP password"),
			),
			fields.New(
				"mailbox",
				fields.TypeString,
				fields.WithHelp("Mailbox to search in"),
				fields.WithDefault("INBOX"),
			),
			fields.New(
				"insecure",
				fields.TypeBool,
				fields.WithHelp("Skip TLS verification"),
				fields.WithDefault(false),
			),
		),
	)
}

func (s *IMAPSettings) ConnectToIMAPServer() (*imapclient.Client, error) {
	serverAddr := fmt.Sprintf("%s:%d", s.Server, s.Port)

	options := &imapclient.Options{
		TLSConfig: &tls.Config{
			// #nosec G402 -- this is an explicit user-controlled dev/test escape hatch exposed as --insecure.
			InsecureSkipVerify: s.Insecure,
		},
	}

	client, err := imapclient.DialTLS(serverAddr, options)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to IMAP server: %w", err)
	}

	if err := client.Login(s.Username, s.Password).Wait(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	return client, nil
}
