package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/smailnail/pkg/dsl"
	"github.com/go-go-golems/smailnail/pkg/imap"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type FetchMailCommand struct {
	*cmds.CommandDescription
}

type FetchMailSettings struct {
	// Search criteria settings
	Since            string   `glazed:"since"`
	Before           string   `glazed:"before"`
	WithinDays       int      `glazed:"within-days"`
	From             string   `glazed:"from"`
	To               string   `glazed:"to"`
	Subject          string   `glazed:"subject"`
	SubjectContains  string   `glazed:"subject-contains"`
	BodyContains     string   `glazed:"body-contains"`
	HasFlags         []string `glazed:"has-flags"`
	DoesNotHaveFlags []string `glazed:"not-has-flags"`
	LargerThan       string   `glazed:"larger-than"`
	SmallerThan      string   `glazed:"smaller-than"`

	// Output settings
	Limit                int    `glazed:"limit"`
	Offset               int    `glazed:"offset"`
	AfterUID             uint32 `glazed:"after-uid"`
	BeforeUID            uint32 `glazed:"before-uid"`
	Format               string `glazed:"format"`
	IncludeContent       bool   `glazed:"include-content"`
	ConcatenateMimeParts bool   `glazed:"concatenate-mime-parts"`
	ContentMaxLength     int    `glazed:"content-max-length"`
	ContentType          string `glazed:"content-type"`
	PrintRule            bool   `glazed:"print-rule"`

	// IMAP settings
	imap.IMAPSettings
}

func NewFetchMailCommand() (*FetchMailCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed section: %w", err)
	}

	imapSection, err := imap.NewIMAPSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create IMAP section: %w", err)
	}

	return &FetchMailCommand{
		CommandDescription: cmds.NewCommandDescription(
			"fetch-mail",
			cmds.WithShort("Fetch emails from an IMAP server using CLI arguments"),
			cmds.WithLong("This command connects to an IMAP server and fetches emails based on search criteria provided as command line arguments"),
			cmds.WithFlags(
				// Search criteria flags
				fields.New(
					"since",
					fields.TypeString,
					fields.WithHelp("Fetch emails since date (YYYY-MM-DD)"),
				),
				fields.New(
					"before",
					fields.TypeString,
					fields.WithHelp("Fetch emails before date (YYYY-MM-DD)"),
				),
				fields.New(
					"within-days",
					fields.TypeInteger,
					fields.WithHelp("Fetch emails within the last N days"),
					fields.WithDefault(0),
				),
				fields.New(
					"from",
					fields.TypeString,
					fields.WithHelp("Fetch emails from a specific sender"),
				),
				fields.New(
					"to",
					fields.TypeString,
					fields.WithHelp("Fetch emails sent to a specific recipient"),
				),
				fields.New(
					"subject",
					fields.TypeString,
					fields.WithHelp("Fetch emails with an exact subject match"),
				),
				fields.New(
					"subject-contains",
					fields.TypeString,
					fields.WithHelp("Fetch emails with subject containing a string"),
				),
				fields.New(
					"body-contains",
					fields.TypeString,
					fields.WithHelp("Fetch emails with body containing a string"),
				),
				fields.New(
					"has-flags",
					fields.TypeStringList,
					fields.WithHelp("Fetch emails with specific flags (comma-separated)"),
				),
				fields.New(
					"not-has-flags",
					fields.TypeStringList,
					fields.WithHelp("Fetch emails without specific flags (comma-separated)"),
				),
				fields.New(
					"larger-than",
					fields.TypeString,
					fields.WithHelp("Fetch emails larger than size (e.g., '1M', '500K')"),
				),
				fields.New(
					"smaller-than",
					fields.TypeString,
					fields.WithHelp("Fetch emails smaller than size (e.g., '1M', '500K')"),
				),

				// Output flags
				fields.New(
					"limit",
					fields.TypeInteger,
					fields.WithHelp("Maximum number of emails to fetch"),
					fields.WithDefault(10),
				),
				fields.New(
					"offset",
					fields.TypeInteger,
					fields.WithHelp("Number of messages to skip (for pagination)"),
					fields.WithDefault(0),
				),
				fields.New(
					"format",
					fields.TypeString,
					fields.WithHelp("Output format (json, text, table)"),
					fields.WithDefault("text"),
				),
				fields.New(
					"include-content",
					fields.TypeBool,
					fields.WithHelp("Include email content in output"),
					fields.WithDefault(true),
				),
				fields.New(
					"concatenate-mime-parts",
					fields.TypeBool,
					fields.WithHelp("Concatenate all MIME parts into a single content string instead of showing structured output"),
					fields.WithDefault(true),
				),
				fields.New(
					"content-max-length",
					fields.TypeInteger,
					fields.WithHelp("Maximum length of content to display"),
					fields.WithDefault(1000),
				),
				fields.New(
					"content-type",
					fields.TypeString,
					fields.WithHelp("MIME type to filter content (e.g., 'text/plain', 'text/*')"),
					fields.WithDefault("text/plain"),
				),
				fields.New(
					"print-rule",
					fields.TypeBool,
					fields.WithHelp("Print the equivalent YAML rule instead of executing it"),
					fields.WithDefault(false),
				),
				fields.New(
					"after-uid",
					fields.TypeInteger,
					fields.WithHelp("Fetch messages with UIDs greater than this value"),
					fields.WithDefault(0),
				),
				fields.New(
					"before-uid",
					fields.TypeInteger,
					fields.WithHelp("Fetch messages with UIDs less than this value"),
					fields.WithDefault(0),
				),
			),
			cmds.WithSections(glazedSection, imapSection),
		),
	}, nil
}

func (c *FetchMailCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedValues *values.Values,
	gp middlewares.Processor,
) error {
	settings := &FetchMailSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return err
	}
	if err := parsedValues.DecodeSectionInto(imap.IMAPSectionSlug, &settings.IMAPSettings); err != nil {
		return err
	}

	log.Debug().Interface("settings", settings).Msg("Settings")

	log.Debug().Msg("Building rule from settings")
	// Build rule from command line arguments
	rule, err := c.buildRuleFromSettings(settings)
	if err != nil {
		return fmt.Errorf("error building rule from settings: %w", err)
	}

	// If print-rule is set, output the rule as YAML and return
	if settings.PrintRule {
		yamlData, err := yaml.Marshal(rule)
		if err != nil {
			return fmt.Errorf("error marshaling rule to YAML: %w", err)
		}

		// Create a row with the YAML data
		row := types.NewRow()
		row.Set("rule", string(yamlData))
		if err := gp.AddRow(ctx, row); err != nil {
			return fmt.Errorf("error adding rule to output: %w", err)
		}
		return nil
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

	// Select mailbox
	log.Debug().Msg("Selecting mailbox")
	if err := c.selectMailbox(client, settings.Mailbox); err != nil {
		return fmt.Errorf("error selecting mailbox: %w", err)
	}

	// Fetch messages
	log.Debug().Msg("Fetching messages")
	msgs, err := rule.FetchMessages(client)
	if err != nil {
		return fmt.Errorf("error fetching messages: %w", err)
	}

	// Process messages
	for _, msg := range msgs {
		// Create a new row for each message
		row := types.NewRow()

		// Always include UID
		row.Set("uid", msg.UID)

		// Always include basic email fields
		if msg.Envelope != nil {
			row.Set("subject", msg.Envelope.Subject)

			if len(msg.Envelope.From) > 0 {
				from := msg.Envelope.From[0]
				row.Set("from", fmt.Sprintf("%s <%s>", from.Name, from.Address))
			}

			if len(msg.Envelope.To) > 0 {
				var toAddresses []string
				for _, to := range msg.Envelope.To {
					toAddresses = append(toAddresses, fmt.Sprintf("%s <%s>", to.Name, to.Address))
				}
				row.Set("to", strings.Join(toAddresses, ", "))
			}

			row.Set("date", msg.Envelope.Date.Format(time.RFC3339))
		}

		// Always include flags and size
		row.Set("flags", strings.Join(msg.Flags, ", "))
		row.Set("size", msg.Size)

		// Handle content if requested
		if settings.IncludeContent && len(msg.MimeParts) > 0 {
			if settings.ConcatenateMimeParts {
				// Concatenate all matching MIME parts into a single content string
				var contents []string
				for _, part := range msg.MimeParts {
					// Fix: Only add slash if Subtype is not empty
					mimeType := part.Type
					if part.Subtype != "" {
						mimeType = part.Type + "/" + part.Subtype
					}

					if c.shouldIncludeMimeType(mimeType, settings.ContentType) {
						contents = append(contents, part.Content)
						log.Debug().
							Str("mime_type", mimeType).
							Str("filter", settings.ContentType).
							Bool("included", true).
							Int("content_length", len(part.Content)).
							Msg("Added MIME part content")
					} else {
						log.Debug().
							Str("mime_type", mimeType).
							Str("filter", settings.ContentType).
							Bool("included", false).
							Msg("Excluded MIME part content")
					}
				}
				content := strings.Join(contents, "\n\n")
				if settings.ContentMaxLength > 0 && len(content) > settings.ContentMaxLength {
					content = content[:settings.ContentMaxLength] + "..."
				}
				row.Set("content", content)
				log.Debug().
					Int("total_parts", len(msg.MimeParts)).
					Int("matched_parts", len(contents)).
					Int("final_content_length", len(content)).
					Msg("Finished processing MIME parts")
			} else {
				// Structured MIME parts output
				var parts []map[string]interface{}
				for _, part := range msg.MimeParts {
					// Fix: Only add slash if Subtype is not empty
					mimeType := part.Type
					if part.Subtype != "" {
						mimeType = part.Type + "/" + part.Subtype
					}

					if c.shouldIncludeMimeType(mimeType, settings.ContentType) {
						partMap := map[string]interface{}{
							"type":    mimeType,
							"size":    part.Size,
							"charset": part.Charset,
						}
						if part.Filename != "" {
							partMap["filename"] = part.Filename
						}

						content := part.Content
						if settings.ContentMaxLength > 0 && len(content) > settings.ContentMaxLength {
							content = content[:settings.ContentMaxLength] + "..."
						}
						partMap["content"] = content

						parts = append(parts, partMap)
						log.Debug().
							Str("mime_type", mimeType).
							Str("filter", settings.ContentType).
							Bool("included", true).
							Int("content_length", len(content)).
							Msg("Added structured MIME part")
					} else {
						log.Debug().
							Str("mime_type", mimeType).
							Str("filter", settings.ContentType).
							Bool("included", false).
							Msg("Excluded structured MIME part")
					}
				}
				row.Set("mime_parts", parts)
				log.Debug().
					Int("total_parts", len(msg.MimeParts)).
					Int("matched_parts", len(parts)).
					Msg("Finished processing structured MIME parts")
			}
		}

		// Add the row to the processor
		if err := gp.AddRow(ctx, row); err != nil {
			return fmt.Errorf("error adding row to processor: %w", err)
		}
	}

	// Add pagination metadata to the last row
	if len(msgs) > 0 {
		// Get total count from first message (all should have the same count)
		totalMessagesFound := int(msgs[0].TotalCount)
		if totalMessagesFound == 0 {
			// Fallback to number of returned messages
			totalMessagesFound = len(msgs)
		}

		// Create pagination metadata row
		paginationRow := types.NewRow()

		// Add regular pagination info
		paginationRow.Set("type", "pagination_metadata")
		paginationRow.Set("total_results", totalMessagesFound)
		paginationRow.Set("fetched_results", len(msgs))
		paginationRow.Set("limit", settings.Limit)
		paginationRow.Set("offset", settings.Offset)
		// Calculate if there are more results
		paginationRow.Set("has_more", totalMessagesFound > (settings.Offset+len(msgs)))
		if totalMessagesFound > (settings.Offset + len(msgs)) {
			paginationRow.Set("next_offset", settings.Offset+len(msgs))
		}

		// Add UID-based pagination info
		var lowestUID, highestUID uint32
		for i, msg := range msgs {
			uid := uint32(msg.UID)
			if i == 0 || uid < lowestUID {
				lowestUID = uid
			}
			if i == 0 || uid > highestUID {
				highestUID = uid
			}
		}

		paginationRow.Set("lowest_uid", lowestUID)
		paginationRow.Set("highest_uid", highestUID)

		// Add the pagination metadata row
		if err := gp.AddRow(ctx, paginationRow); err != nil {
			log.Error().Err(err).Msg("Failed to add pagination metadata row")
		}
	}

	return nil
}

// Build a Rule struct from command line settings
func (c *FetchMailCommand) buildRuleFromSettings(settings *FetchMailSettings) (*dsl.Rule, error) {
	// Start building the search config
	searchConfig := dsl.SearchConfig{
		Since:           settings.Since,
		Before:          settings.Before,
		WithinDays:      settings.WithinDays,
		From:            settings.From,
		To:              settings.To,
		Subject:         settings.Subject,
		SubjectContains: settings.SubjectContains,
		BodyContains:    settings.BodyContains,
	}

	// Add flag criteria if specified
	if len(settings.HasFlags) > 0 || len(settings.DoesNotHaveFlags) > 0 {
		searchConfig.Flags = &dsl.FlagCriteria{
			Has:    settings.HasFlags,
			NotHas: settings.DoesNotHaveFlags,
		}
	}

	// Add size criteria if specified
	if settings.LargerThan != "" || settings.SmallerThan != "" {
		searchConfig.Size = &dsl.SizeCriteria{
			LargerThan:  settings.LargerThan,
			SmallerThan: settings.SmallerThan,
		}
	}

	// Build fields for output config
	var fields []interface{}

	// Always include basic email fields
	fields = append(fields,
		dsl.Field{Name: "uid"},
		dsl.Field{Name: "subject"},
		dsl.Field{Name: "from"},
		dsl.Field{Name: "to"},
		dsl.Field{Name: "date"},
		dsl.Field{Name: "flags"},
		dsl.Field{Name: "size"},
	)

	// Add content field if needed
	if settings.IncludeContent {
		contentField := &dsl.ContentField{
			ShowContent: true,
			MaxLength:   settings.ContentMaxLength,
		}

		// Set types for filtering
		if settings.ContentType != "" {
			contentField.Mode = "filter"
			contentField.Types = []string{settings.ContentType}
		}

		fields = append(fields, dsl.Field{
			Name:    "mime_parts",
			Content: contentField,
		})
	}

	// Create output config
	outputConfig := dsl.OutputConfig{
		Format:    settings.Format,
		Limit:     settings.Limit,
		Offset:    settings.Offset,
		AfterUID:  settings.AfterUID,
		BeforeUID: settings.BeforeUID,
		Fields:    fields,
	}

	// Create the rule
	rule := &dsl.Rule{
		Name:        "cli-rule",
		Description: "Rule generated from command line arguments",
		Search:      searchConfig,
		Output:      outputConfig,
	}

	// Validate the rule
	if err := rule.Validate(); err != nil {
		return nil, fmt.Errorf("invalid rule: %w", err)
	}

	return rule, nil
}

func (c *FetchMailCommand) selectMailbox(client *imapclient.Client, mailbox string) error {
	if _, err := client.Select(mailbox, nil).Wait(); err != nil {
		return fmt.Errorf("failed to select mailbox %q: %w", mailbox, err)
	}
	return nil
}

func (c *FetchMailCommand) shouldIncludeMimeType(mimeType string, filter string) bool {
	// If no filter, include all
	if filter == "" {
		return true
	}

	log.Debug().
		Str("mime_type", mimeType).
		Str("filter", filter).
		Msg("Checking MIME type match")

	// Exact match
	if mimeType == filter {
		log.Debug().Msg("Exact match")
		return true
	}

	// Wildcard match (e.g., text/*)
	if strings.HasSuffix(filter, "/*") {
		prefix := strings.TrimSuffix(filter, "/*")
		result := strings.HasPrefix(mimeType, prefix+"/")
		log.Debug().
			Str("prefix", prefix).
			Bool("wildcard_match", result).
			Msg("Checking wildcard match")
		return result
	}

	log.Debug().Msg("No match found")
	return false
}
