# Email Generation Tool (mailgen) Implementation Plan

## Overview

This document outlines the technical plan for implementing the `mailgen` tool, which will generate test emails based on the YAML-based DSL as specified in the Email Generation DSL document. The tool will be implemented in Go and will initially focus on generating email files from YAML template files.

## 1. Project Structure

```
smailnail/
├── cmd/
│   ├── mailgen/
│   │   ├── main.go             # Entry point of the application
│   │   ├── commands/           # Command implementations
│   │   │   ├── generate.go     # Generate command implementation
│   ├── ...
├── pkg/
│   ├── mailgen/                # Package for email generation logic
│   │   ├── mailgen.go          # DSL implementation
│   ├── ...
```

## 2. Implementation Steps

### 2.1 Define Core Data Structures

- [ ] Create data structures to represent the DSL components:

```go
// pkg/mailgen/email.go

package mailgen

// DSLConfig represents the complete DSL configuration from a YAML file
type DSLConfig struct {
    Variables map[string]interface{}        `yaml:"variables"`
    Templates map[string]EmailTemplate      `yaml:"templates"`
    Rules     map[string]Rule               `yaml:"rules"`
    Generate  []GenerationConfig            `yaml:"generate"`
}

// EmailTemplate defines an email template structure
type EmailTemplate struct {
    Subject  string `yaml:"subject"`
    From     string `yaml:"from"`
    To       string `yaml:"to,omitempty"`
    CC       string `yaml:"cc,omitempty"`
    BCC      string `yaml:"bcc,omitempty"`
    ReplyTo  string `yaml:"reply_to,omitempty"`
    Body     string `yaml:"body"`
}

// Rule defines how to generate variations of emails
type Rule struct {
    Template   string                   `yaml:"template"`
    Variations []map[string]interface{} `yaml:"variations"`
}

// GenerationConfig specifies generation settings
type GenerationConfig struct {
    Rule   string `yaml:"rule"`
    Count  int    `yaml:"count"`
    Output string `yaml:"output,omitempty"`
}

// Email represents a generated email
type Email struct {
    Subject  string
    From     string
    To       string
    CC       string
    BCC      string
    ReplyTo  string
    Body     string
    Headers  map[string]string
}
```

### 2.2 Implement YAML Parser

- [ ] Create a parser to load and validate DSL configurations:

```go
// pkg/mailgen/parser.go

package mailgen

import (
    "fmt"
    "io/ioutil"
    "os"

    "gopkg.in/yaml.v3"
)

// ParseYAML parses a YAML file into a DSLConfig structure
func ParseYAML(filePath string) (*DSLConfig, error) {
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("error reading file: %w", err)
    }

    var config DSLConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("error parsing YAML: %w", err)
    }

    if err := validateConfig(&config); err != nil {
        return nil, err
    }

    return &config, nil
}

// validateConfig validates the DSL configuration
func validateConfig(config *DSLConfig) error {
    // Validate templates
    for id, template := range config.Templates {
        if template.Subject == "" {
            return fmt.Errorf("template '%s' missing required field 'subject'", id)
        }
        if template.From == "" {
            return fmt.Errorf("template '%s' missing required field 'from'", id)
        }
        if template.Body == "" {
            return fmt.Errorf("template '%s' missing required field 'body'", id)
        }
    }

    // Validate rules
    for id, rule := range config.Rules {
        if rule.Template == "" {
            return fmt.Errorf("rule '%s' missing required field 'template'", id)
        }
        if _, exists := config.Templates[rule.Template]; !exists {
            return fmt.Errorf("rule '%s' references non-existent template '%s'", id, rule.Template)
        }
        if len(rule.Variations) == 0 {
            return fmt.Errorf("rule '%s' has no variations", id)
        }
    }

    // Validate generation configs
    for i, gen := range config.Generate {
        if gen.Rule == "" {
            return fmt.Errorf("generation config #%d missing required field 'rule'", i+1)
        }
        if _, exists := config.Rules[gen.Rule]; !exists {
            return fmt.Errorf("generation config #%d references non-existent rule '%s'", i+1, gen.Rule)
        }
        if gen.Count <= 0 {
            return fmt.Errorf("generation config #%d must have count > 0", i+1)
        }
    }

    return nil
}
```

### 2.3 Implement Template Context

- [ ] Create a context object for template rendering:

```go
// pkg/mailgen/context.go

package mailgen

import (
    "fmt"
)

// TemplateContext represents the context available during template rendering
type TemplateContext struct {
    Variables  map[string]interface{}  // Global variables
    Template   string                  // Current template ID
    Rule       string                  // Current rule ID
    Index      int                     // Current generation index
    Properties map[string]interface{}  // Variation properties
}

// NewTemplateContext creates a new template context
func NewTemplateContext(
    variables map[string]interface{},
    templateID string,
    ruleID string,
    index int,
    properties map[string]interface{},
) *TemplateContext {
    return &TemplateContext{
        Variables:  variables,
        Template:   templateID,
        Rule:       ruleID,
        Index:      index,
        Properties: properties,
    }
}

// Get returns a value from the context
func (c *TemplateContext) Get(key string) (interface{}, error) {
    // First check properties
    if val, exists := c.Properties[key]; exists {
        return val, nil
    }

    // Then check top-level context
    switch key {
    case "variables":
        return c.Variables, nil
    case "template":
        return c.Template, nil
    case "rule":
        return c.Rule, nil
    case "index":
        return c.Index, nil
    }

    return nil, fmt.Errorf("key '%s' not found in context", key)
}

// ToMap converts the context to a map for template rendering
func (c *TemplateContext) ToMap() map[string]interface{} {
    result := make(map[string]interface{})

    // Add top-level context
    result["variables"] = c.Variables
    result["template"] = c.Template
    result["rule"] = c.Rule
    result["index"] = c.Index

    // Add all properties to root level
    for k, v := range c.Properties {
        result[k] = v
    }

    return result
}
```

### 2.4 Implement Template Renderer

- [ ] Create a renderer to process templates with Sprig functions:

```go
// pkg/mailgen/renderer.go

package mailgen

import (
    "bytes"
    "fmt"
    "text/template"

    "github.com/Masterminds/sprig/v3"
)

// TemplateRenderer handles template rendering with Sprig functions
type TemplateRenderer struct {
    funcMap template.FuncMap
}

// NewTemplateRenderer creates a new template renderer
func NewTemplateRenderer() *TemplateRenderer {
    // Get all sprig functions
    funcMap := sprig.FuncMap()

    // Add any custom functions
    funcMap["pick"] = pickRandom

    return &TemplateRenderer{
        funcMap: funcMap,
    }
}

// RenderString renders a template string with the given context
func (r *TemplateRenderer) RenderString(templateStr string, context map[string]interface{}) (string, error) {
    tmpl, err := template.New("").Funcs(r.funcMap).Parse(templateStr)
    if err != nil {
        return "", fmt.Errorf("template parse error: %w", err)
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, context); err != nil {
        return "", fmt.Errorf("template execution error: %w", err)
    }

    return buf.String(), nil
}

// RenderEmail renders all parts of an email template with the given context
func (r *TemplateRenderer) RenderEmail(tmpl EmailTemplate, context map[string]interface{}) (*Email, error) {
    email := &Email{
        Headers: make(map[string]string),
    }

    // Render subject
    subject, err := r.RenderString(tmpl.Subject, context)
    if err != nil {
        return nil, fmt.Errorf("subject rendering error: %w", err)
    }
    email.Subject = subject

    // Render from
    from, err := r.RenderString(tmpl.From, context)
    if err != nil {
        return nil, fmt.Errorf("from rendering error: %w", err)
    }
    email.From = from

    // Render optional to
    if tmpl.To != "" {
        to, err := r.RenderString(tmpl.To, context)
        if err != nil {
            return nil, fmt.Errorf("to rendering error: %w", err)
        }
        email.To = to
    }

    // Render optional CC
    if tmpl.CC != "" {
        cc, err := r.RenderString(tmpl.CC, context)
        if err != nil {
            return nil, fmt.Errorf("cc rendering error: %w", err)
        }
        email.CC = cc
    }

    // Render optional BCC
    if tmpl.BCC != "" {
        bcc, err := r.RenderString(tmpl.BCC, context)
        if err != nil {
            return nil, fmt.Errorf("bcc rendering error: %w", err)
        }
        email.BCC = bcc
    }

    // Render optional reply-to
    if tmpl.ReplyTo != "" {
        replyTo, err := r.RenderString(tmpl.ReplyTo, context)
        if err != nil {
            return nil, fmt.Errorf("reply-to rendering error: %w", err)
        }
        email.ReplyTo = replyTo
    }

    // Render body
    body, err := r.RenderString(tmpl.Body, context)
    if err != nil {
        return nil, fmt.Errorf("body rendering error: %w", err)
    }
    email.Body = body

    return email, nil
}

// Helper functions for template rendering
func pickRandom(slice interface{}) (interface{}, error) {
    // Implementation for pick function
    // This is a simplified version that will need to be expanded
    return nil, nil
}
```

### 2.5 Implement Generator

- [ ] Create the main generator to orchestrate the process:

```go
// pkg/mailgen/generator.go

package mailgen

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/Masterminds/sprig/v3"
)

// Generator manages the email generation process
type Generator struct {
    Config   *DSLConfig
    Renderer *TemplateRenderer
}

// NewGenerator creates a new generator with the given configuration
func NewGenerator(config *DSLConfig) *Generator {
    return &Generator{
        Config:   config,
        Renderer: NewTemplateRenderer(),
    }
}

// Generate executes the email generation process
func (g *Generator) Generate() error {
    for _, genConfig := range g.Config.Generate {
        rule, exists := g.Config.Rules[genConfig.Rule]
        if !exists {
            return fmt.Errorf("rule '%s' not found", genConfig.Rule)
        }

        template, exists := g.Config.Templates[rule.Template]
        if !exists {
            return fmt.Errorf("template '%s' not found", rule.Template)
        }

        // Generate specified number of emails
        for i := 0; i < genConfig.Count; i++ {
            // Select a variation (for simplicity, we'll just cycle through them)
            variation := rule.Variations[i % len(rule.Variations)]

            // Create context for the template
            ctx := NewTemplateContext(
                g.Config.Variables,
                rule.Template,
                genConfig.Rule,
                i+1, // Use 1-based index for output
                variation,
            )

            // Pre-process variation properties that might contain templates
            processedVariation, err := g.preprocessVariation(variation, ctx.ToMap())
            if err != nil {
                return fmt.Errorf("error preprocessing variation: %w", err)
            }

            // Update context with processed values
            for k, v := range processedVariation {
                ctx.Properties[k] = v
            }

            // Render the email
            email, err := g.Renderer.RenderEmail(template, ctx.ToMap())
            if err != nil {
                return fmt.Errorf("error rendering email: %w", err)
            }

            // Save the email
            if err := g.saveEmail(email, genConfig.Output, ctx.ToMap()); err != nil {
                return fmt.Errorf("error saving email: %w", err)
            }
        }
    }

    return nil
}

// preprocessVariation processes all variation properties that might contain templates
func (g *Generator) preprocessVariation(variation map[string]interface{}, context map[string]interface{}) (map[string]interface{}, error) {
    processed := make(map[string]interface{})

    for key, value := range variation {
        // Only process string values as templates
        if strValue, ok := value.(string); ok {
            // Check if it looks like a template
            if strings.Contains(strValue, "{{") {
                rendered, err := g.Renderer.RenderString(strValue, context)
                if err != nil {
                    return nil, fmt.Errorf("error rendering property '%s': %w", key, err)
                }
                processed[key] = rendered
            } else {
                processed[key] = strValue
            }
        } else {
            // Non-string values are kept as-is
            processed[key] = value
        }
    }

    return processed, nil
}

// saveEmail saves the generated email to a file
func (g *Generator) saveEmail(email *Email, outputTemplate string, context map[string]interface{}) error {
    // If no output template, just print to stdout
    if outputTemplate == "" {
        fmt.Println("Subject:", email.Subject)
        fmt.Println("From:", email.From)
        if email.To != "" {
            fmt.Println("To:", email.To)
        }
        if email.CC != "" {
            fmt.Println("CC:", email.CC)
        }
        if email.BCC != "" {
            fmt.Println("BCC:", email.BCC)
        }
        if email.ReplyTo != "" {
            fmt.Println("Reply-To:", email.ReplyTo)
        }
        fmt.Println()
        fmt.Println(email.Body)
        fmt.Println("----------------------------")
        return nil
    }

    // Render the output path template
    path, err := g.Renderer.RenderString(outputTemplate, context)
    if err != nil {
        return fmt.Errorf("error rendering output path: %w", err)
    }

    // Create parent directories if needed
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("error creating directories: %w", err)
    }

    // Open the file for writing
    file, err := os.Create(path)
    if err != nil {
        return fmt.Errorf("error creating file: %w", err)
    }
    defer file.Close()

    // Write email headers
    fmt.Fprintf(file, "Subject: %s\n", email.Subject)
    fmt.Fprintf(file, "From: %s\n", email.From)
    if email.To != "" {
        fmt.Fprintf(file, "To: %s\n", email.To)
    }
    if email.CC != "" {
        fmt.Fprintf(file, "CC: %s\n", email.CC)
    }
    if email.BCC != "" {
        fmt.Fprintf(file, "BCC: %s\n", email.BCC)
    }
    if email.ReplyTo != "" {
        fmt.Fprintf(file, "Reply-To: %s\n", email.ReplyTo)
    }
    fmt.Fprintf(file, "Date: %s\n", time.Now().Format(time.RFC822Z))

    // Add any custom headers
    for name, value := range email.Headers {
        fmt.Fprintf(file, "%s: %s\n", name, value)
    }

    // Add empty line before body
    fmt.Fprintln(file)

    // Write email body
    fmt.Fprintln(file, email.Body)

    return nil
}
```

### 2.6 Implement CLI Commands

- [ ] Create the root command:

```go
// cmd/mailgen/commands/root.go

package commands

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
    Use:   "mailgen",
    Short: "Generate test email content from YAML templates",
    Long: `mailgen is a CLI tool for generating test email content from YAML templates
following a defined DSL. It allows creating multiple variations of emails
with random or predefined content.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

func init() {
    // Here you will define your flags and configuration settings
    rootCmd.PersistentFlags().StringP("output-dir", "o", "output", "Directory to save generated emails")
    rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
}
```

- [ ] Create the generate command:

```go
// cmd/mailgen/commands/generate.go

package commands

import (
    "fmt"
    "path/filepath"

    "github.com/spf13/cobra"
    "smailnail/pkg/mailgen"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
    Use:   "generate [template-file]",
    Short: "Generate emails from a template file",
    Long: `Generate emails based on a YAML template file following the Email Generation DSL.
Multiple email variations can be generated based on the rules defined in the template.`,
    Args: cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        templateFile := args[0]

        // Parse output directory flag
        outputDir, _ := cmd.Flags().GetString("output-dir")
        verbose, _ := cmd.Flags().GetBool("verbose")

        // Parse the template file
        config, err := mailgen.ParseYAML(templateFile)
        if err != nil {
            return fmt.Errorf("failed to parse template file: %w", err)
        }

        // Set default output path if not specified
        for i, genConfig := range config.Generate {
            if genConfig.Output == "" {
                // Default output path: {output-dir}/{rule}-{template}-{index}.eml
                ruleName := genConfig.Rule
                templateName := config.Rules[genConfig.Rule].Template
                config.Generate[i].Output = filepath.Join(
                    outputDir,
                    fmt.Sprintf("%s-%s-{{ .index }}.eml", ruleName, templateName),
                )
            }
        }

        // Create the generator
        generator := mailgen.NewGenerator(config)

        // Generate emails
        if err := generator.Generate(); err != nil {
            return fmt.Errorf("generation failed: %w", err)
        }

        if verbose {
            fmt.Println("Email generation completed successfully")
        }

        return nil
    },
}

func init() {
    rootCmd.AddCommand(generateCmd)

    // Add local flags for the generate command
    generateCmd.Flags().StringP("format", "f", "eml", "Output format (eml, txt, json)")
}
```

- [ ] Create the main entry point:

```go
// cmd/mailgen/main.go

package main

import (
    "smailnail/cmd/mailgen/commands"
)

func main() {
    commands.Execute()
}
```

### 2.7 Implement Custom Sprig Wrapper

- [ ] Create a wrapper for Sprig functions with extensions:

```go
// pkg/mailgen/sprig_wrapper.go

package mailgen

import (
    "math/rand"
    "reflect"
    "time"

    "github.com/Masterminds/sprig/v3"
)

// init initializes the random seed
func init() {
    rand.Seed(time.Now().UnixNano())
}

// PickRandom picks a random element from a slice
func PickRandom(slice interface{}) (interface{}, error) {
    val := reflect.ValueOf(slice)
    if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
        return nil, nil // Return nil for non-slice types
    }

    length := val.Len()
    if length == 0 {
        return nil, nil // Return nil for empty slices
    }

    // Get a random index
    index := rand.Intn(length)
    return val.Index(index).Interface(), nil
}

// RegisterCustomFunctions registers custom functions to augment the Sprig function map
func RegisterCustomFunctions(funcMap map[string]interface{}) {
    // Add our custom functions
    funcMap["pick"] = PickRandom

    // Add any other custom functions here
}

// GetFuncMap returns a complete function map with Sprig and custom functions
func GetFuncMap() map[string]interface{} {
    funcMap := sprig.FuncMap()
    RegisterCustomFunctions(funcMap)
    return funcMap
}
```

## 3. Testing Plan

### 3.1 Unit Tests

- [ ] Test YAML parsing and validation
- [ ] Test template rendering with Sprig functions
- [ ] Test email generation process
- [ ] Test template context functionality

### 3.2 Integration Tests

- [ ] Test end-to-end process from YAML to email output
- [ ] Test with various template examples

## 4. Example Usage

Once implemented, the tool can be used as follows:

```bash
# Generate emails from a template file
mailgen generate templates/example.yaml

# Specify output directory
mailgen generate --output-dir emails/ templates/example.yaml

# Enable verbose output
mailgen generate --verbose templates/example.yaml
```

## 5. Example Template

Here's a simple example of a valid template file:

```yaml
variables:
  first_names:
    - John
    - Jane
    - Bob
    - Alice
  last_names:
    - Smith
    - Doe
    - Johnson
    - Williams
  domains:
    - example.com
    - mailtest.org
    - testmail.net
  greetings:
    - Hello
    - Hi
    - Hey
    - Dear
  closings:
    - Best regards
    - Cheers
    - Thanks
    - Sincerely

templates:
  simple_email:
    subject: "Test Email: {{ .subject_line }}"
    from: "{{ .sender_name }} <{{ .sender_email }}>"
    to: "{{ .recipient_name }} <{{ .recipient_email }}>"
    body: |
      {{ .greeting }} {{ .recipient_first_name }},
      
      {{ .body_content }}
      
      {{ .closing }},
      {{ .sender_name }}

rules:
  test_emails:
    template: simple_email
    variations:
      - sender_name: "{{ pick .variables.first_names }} {{ pick .variables.last_names }}"
        sender_email: "sender@example.com"
        recipient_name: "{{ pick .variables.first_names }} {{ pick .variables.last_names }}"
        recipient_first_name: "{{ pick .variables.first_names }}"
        recipient_email: "{{ lower (printf \"%s.%s@%s\" (pick .variables.first_names) (pick .variables.last_names) (pick .variables.domains)) }}"
        greeting: "{{ pick .variables.greetings }}"
        subject_line: "Important Update {{ now | date \"2006-01-02\" }}"
        body_content: "This is a test email generated by the mailgen tool."
        closing: "{{ pick .variables.closings }}"

generate:
  - rule: test_emails
    count: 5
    output: "emails/test-{{ now | date \"2006-01-02\" }}/email-{{ .index }}.eml"
```

## 6. Phased Implementation Approach

### 6.1 Phase 1: Basic Email Generation

- [ ] Implement basic DSL parser and validator
- [ ] Implement template rendering with simple Sprig functions
- [ ] Implement email generation and output to files
- [ ] Implement CLI commands

### 6.2 Phase 2: Enhanced Features

- [ ] Add support for more complex template contexts
- [ ] Add support for nested templates
- [ ] Add support for more Sprig functions
- [ ] Add support for multiple output formats

### 6.3 Phase 3: Advanced Features

- [ ] Add support for MIME attachments
- [ ] Add support for HTML emails
- [ ] Add support for direct SMTP sending
- [ ] Add support for scripting/programming logic in templates

## 7. Dependencies

- `github.com/spf13/cobra` - Command line interface framework
- `github.com/Masterminds/sprig/v3` - Template functions library
- `gopkg.in/yaml.v3` - YAML parsing library
- `text/template` - Go standard library for templating

## 8. Considerations and Challenges

1. **Template Error Handling**: Provide clear error messages for template rendering issues
2. **Performance**: Optimize for large-scale generation of emails
3. **Security**: Ensure template execution is secure (avoid code execution vulnerabilities)
4. **Extensibility**: Design for future extensions like SMTP integration

## 9. Next Steps

1. Implement the core data structures
2. Implement the YAML parser and validator
3. Implement the template rendering engine
4. Implement the generator
5. Implement the CLI commands
6. Test with example templates
7. Document usage and examples 