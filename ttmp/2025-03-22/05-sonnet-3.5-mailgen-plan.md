# Email Generation Implementation Plan

## Overview
This plan outlines the step-by-step implementation of a command-line tool for generating test emails based on the YAML DSL specification. The tool will be built incrementally, starting with basic functionality and growing in complexity.

The base package of our project is 	"github.com/go-go-golems/smailnail"

## Phase 1: Basic Email Generation
- [ ] Setup project structure
  ```
  cmd/mailgen/
    ├── main.go           # Main entry point
    ├── cmds/         # Command implementations
    │   └── generate.go   # Generate command
    └── pkg/
        ├── mailgen/mailgen.go # DSL implementation
  ```

### 1.1 Core Data Structures
- [ ] Define core types in `pkg/types/config.go`:
  ```go
  type EmailTemplate struct {
      Subject   string            `yaml:"subject"`
      From      string            `yaml:"from"`
      To        string            `yaml:"to,omitempty"`
      Cc        string            `yaml:"cc,omitempty"`
      Bcc       string            `yaml:"bcc,omitempty"`
      ReplyTo   string            `yaml:"reply_to,omitempty"`
      Body      string            `yaml:"body"`
  }

  type TemplateConfig struct {
      Variables  map[string]interface{} `yaml:"variables"` // All values must be strings or []string
      Templates  map[string]EmailTemplate `yaml:"templates"`
      Rules      map[string]RuleConfig    `yaml:"rules"`
      Generate   []GenerateConfig         `yaml:"generate"`
  }

  type RuleConfig struct {
      Template    string                   `yaml:"template"`
      Variations  []map[string]string      `yaml:"variations"` // All values must be strings
  }

  type GenerateConfig struct {
      Rule    string `yaml:"rule"`
      Count   int    `yaml:"count"`
      Output  string `yaml:"output,omitempty"`
  }
  ```

### 1.2 Command Line Interface
- [ ] Implement basic cobra command in `cmd/mailgen/main.go`:
  ```go
  func main() {
      rootCmd := &cobra.Command{
          Use:   "mailgen",
          Short: "Generate test emails from YAML templates",
      }
      
      rootCmd.AddCommand(commands.NewGenerateCommand())
      
      if err := rootCmd.Execute(); err != nil {
          fmt.Fprintf(os.Stderr, "Error: %v\n", err)
          os.Exit(1)
      }
  }
  ```

### 1.3 Generate Command
- [ ] Implement generate command in `cmd/mailgen/cmds/generate.go`:
  ```go
  type GenerateCommand struct {
      *cmds.CommandDescription
  }

  func NewGenerateCommand() *GenerateCommand {
      return &GenerateCommand{
          CommandDescription: cmds.NewCommandDescription(
              "generate",
              cmds.WithShort("Generate emails from template"),
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
              ),
          ),
      }
  }
  ```

### 1.4 Template Processing
- [ ] Implement template processor 
  ```go
  type MailGenerator struct {
      config *types.TemplateConfig
      funcs  template.FuncMap
  }

  func NewMailGenerator(config *types.TemplateConfig) *MailGenerator {
      return &MailGenerator{
          config: config,
          funcs:  sprig.TxtFuncMap(),
      }
  }

  func (g *MailGenerator) Generate(ctx context.Context) ([]*types.Email, error) {
      // Process templates and generate emails
  }
  ```

## Phase 2: Enhanced Features
- [ ] Implement template helpers if not present in sprig
  - Date/time functions
  - Random data generation
  - String manipulation functions

## Phase 3: Advanced Features FOR LATER
- [ ] Add support for attachments
  - File attachments
  - Generated attachments
  - MIME types

### Template Processing
1. Use Go's `text/template` with Sprig functions for template processing
3. Use structured error handling with `pkg/errors`

## Usage Example
```yaml
# example.yaml
variables:
  sender_name: "John Doe"
  sender_email: "john@example.com"
  subjects:
    - "Hello there!"
    - "Important update"
  greetings:
    - "Hello"
    - "Hi"

templates:
  basic:
    subject: "{{ .subject }}"
    from: "{{ .sender_name }} <{{ .sender_email }}>"
    body: |
      {{ .greeting }} {{ .recipient }},
      
      {{ .content }}
      
      Best regards,
      {{ .sender_name }}

rules:
  welcome:
    template: basic
    variations:
      - sender_name: "{{ .variables.sender_name }}"
        sender_email: "{{ .variables.sender_email }}"
        subject: "{{ index .variables.subjects 0 }}"
        greeting: "{{ index .variables.greetings 0 }}"
        recipient: "User"
        content: "Welcome to our service!"

generate:
  - rule: welcome
    count: 5
    output: "emails/welcome_{{ .index }}.txt"
```

```bash
# Usage
mailgen generate --config example.yaml --output-dir ./generated_emails
``` 