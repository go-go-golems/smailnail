package cmds

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/smailnail/pkg/mailgen"
	mailgenTypes "github.com/go-go-golems/smailnail/pkg/types"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// GenerateCommand represents the generate command
type GenerateCommand struct {
	*cmds.CommandDescription
}

// NewGenerateCommand creates a new generate command
func NewGenerateCommand() (*GenerateCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create glazed parameter layers")
	}

	return &GenerateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"generate",
			cmds.WithShort("Generate emails from template"),
			cmds.WithLong("Generate emails from template using YAML configuration"),
			cmds.WithLayersList(glazedLayer),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"config",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to YAML config file"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"output-dir",
					parameters.ParameterTypeString,
					parameters.WithHelp("Directory to output generated emails"),
					parameters.WithDefault("./output"),
				),
				parameters.NewParameterDefinition(
					"write-files",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Write emails to files"),
					parameters.WithDefault(false),
				),
			),
		),
	}, nil
}

// RunIntoGlazeProcessor generates emails and outputs them as structured data
func (c *GenerateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse command settings
	type GenerateSettings struct {
		ConfigFile string `glazed.parameter:"config"`
		OutputDir  string `glazed.parameter:"output-dir"`
		WriteFiles bool   `glazed.parameter:"write-files"`
	}

	settings := &GenerateSettings{}
	if err := parsedLayers.InitializeStruct("default", settings); err != nil {
		return err
	}

	// Read and parse config file
	configData, err := os.ReadFile(settings.ConfigFile)
	if err != nil {
		return errors.Wrapf(err, "failed to read config file '%s'", settings.ConfigFile)
	}

	var config mailgenTypes.TemplateConfig
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return errors.Wrapf(err, "failed to parse config file '%s'", settings.ConfigFile)
	}

	// Create mail generator
	generator := mailgen.NewMailGenerator(&config)

	// Generate emails
	emails, err := generator.Generate(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to generate emails")
	}

	// Create output directory if needed
	if settings.WriteFiles {
		if err := os.MkdirAll(settings.OutputDir, 0755); err != nil {
			return errors.Wrapf(err, "failed to create output directory '%s'", settings.OutputDir)
		}
	}

	// Process generated emails
	for i, email := range emails {
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
			if err := os.WriteFile(filePath, []byte(emailText), 0644); err != nil {
				return errors.Wrapf(err, "failed to write email %d to file '%s'", i, filePath)
			}
		}
	}

	return nil
}
