# Mail Generation Tool Implementation Plan

## Phase 1: Core Functionality

### 2. CLI Structure (Cobra)
```go
// cmd/root.go
var rootCmd = &cobra.Command{
  Use:   "mailgen",
  Short: "Generate test emails from YAML templates",
}

func Execute() {
  if err := rootCmd.Execute(); err != nil {
    os.Exit(1)
  }
}
```

### 3. YAML Parsing Structure
```go
// pkg/mailgen/spec.go
type EmailSpec struct {
  Variables map[string]interface{} `yaml:"variables"`
  Templates map[string]EmailTemplate `yaml:"templates"`
  Rules     map[string]GenerationRule `yaml:"rules"`
  Generate  []GenerationConfig `yaml:"generate"`
}

type EmailTemplate struct {
  Subject  string `yaml:"subject"`
  From     string `yaml:"from"`
  Body     string `yaml:"body"`
  // ... other fields
}

type GenerationRule struct {
  Template   string         `yaml:"template"`
  Variations []RuleVariation `yaml:"variations"`
}

type RuleVariation map[string]interface{}
```

### 4. Template Processing Engine
```go
// pkg/mailgen/engine.go
type TemplateEngine struct {
  funcMap template.FuncMap
}

func NewTemplateEngine() *TemplateEngine {
  return &TemplateEngine{
    funcMap: sprig.TxtFuncMap(),
  }
}

func (e *TemplateEngine) RenderTemplate(tplText string, data interface{}) (string, error) {
  tpl := template.New("").Funcs(e.funcMap)
  tpl, err := tpl.Parse(tplText)
  // ... error handling
}
```

### 5. Generation Context
```go
// pkg/mailgen/context.go
type GenerationContext struct {
  Variables map[string]interface{}
  Index     int
  Template  *EmailTemplate
  Rule      *GenerationRule
  // Additional context fields
}

func BuildContext(spec *EmailSpec, config *GenerationConfig) *GenerationContext {
  // ...
}
```

### 6. Main Generation Logic
```go
// pkg/mailgen/generator.go
func GenerateEmails(spec *EmailSpec, outputDir string) error {
  engine := NewTemplateEngine()
  
  for _, genConfig := range spec.Generate {
    rule := spec.Rules[genConfig.Rule]
    template := spec.Templates[rule.Template]
    
    for i := 0; i < genConfig.Count; i++ {
      ctx := BuildContext(spec, genConfig, i)
      
      // Process each variation
      for _, variation := range rule.Variations {
        email, err := processVariation(engine, ctx, variation)
        // ... handle error
      }
    }
  }
  return nil
}
```

## Phase 2: Implementation Steps

### 1. Basic YAML Loading
- [ ] Implement YAML parsing with validation
- [ ] Add basic error handling for missing sections

### 2. Template Processing
- [ ] Implement core template rendering with Sprig functions
- [ ] Handle nested variable resolution
- [ ] Add template caching for performance

### 3. Variation Processing
- [ ] Implement property resolution order:
  1. Global variables
  2. Rule-level variables
  3. Variation-specific properties

### 4. Output Generation
```go
// pkg/mailgen/output.go
func writeEmail(outputPath string, email *Email) error {
  // Create directory structure if needed
  // Write email parts to file
}
```

### 5. CLI Integration
```go
// cmd/generate.go
var generateCmd = &cobra.Command{
  Use:   "generate [YAML_FILE]",
  Args:  cobra.ExactArgs(1),
  Run: func(cmd *cobra.Command, args []string) {
    spec := loadSpec(args[0])
    err := mailgen.GenerateEmails(spec)
    // ... error handling
  },
}
```

## Phase 3: Advanced Features

### 1. Template Validation
- [ ] Pre-validate all templates at load time
- [ ] Check for missing variables
- [ ] Verify template syntax

### 2. Contextual Functions
```go
// Add custom template functions
funcMap["pick"] = func(items ...interface{}) interface{} {
  // implementation
}

funcMap["randInt"] = func(min, max int) int {
  // implementation
}
```

### 3. Output Formats
- [ ] Support multiple output formats (text, EML, JSON)
- [ ] Implement output template system
- [ ] Add progress reporting

## Testing Strategy

1. Unit Tests:
   - Template rendering with mock data
   - YAML parsing edge cases
   - Variable resolution order

2. Integration Tests:
   - Full pipeline tests with sample YAML files
   - Output validation

3. Example Files:
   ```yaml
   # examples/basic.yaml
   variables:
     domains: ["example.com", "test.org"]
   templates:
     basic:
       subject: "Test Email {{.index}}"
       body: "Hello from {{pick .variables.domains}}"
   generate:
     - rule: basic_rule
       count: 5
   ```

## Dependencies

- **Cobra**: CLI structure and argument handling
- **Sprig**: Template functions
- **YAML.v3**: Configuration parsing
- **zerolog**: Structured logging (optional)

## Next Steps

1. Implement core YAML parsing
2. Set up basic template rendering
3. Add first working version of generate command

Would you like me to proceed with implementing any specific part of this plan first? 

### 5. CLI Integration (Glazed Style)

```go
// cmd/generate.go
import (
  "github.com/go-go-golems/glazed/pkg/cmds"
  "github.com/go-go-golems/glazed/pkg/cmds/layers"
  "github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

type GenerateCommand struct {
  *cmds.CommandDescription
}

func NewGenerateCommand() (*GenerateCommand, error) {
  return &GenerateCommand{
    CommandDescription: cmds.NewCommandDescription(
      "generate",
      cmds.WithShort("Generate emails from YAML spec"),
      cmds.WithFlags(
        parameters.NewParameterDefinition(
          "output-dir",
          parameters.ParameterTypeString,
          parameters.WithShortFlag("o"),
          parameters.WithHelp("Output directory for generated emails"),
          parameters.WithDefault("./output"),
        ),
        parameters.NewParameterDefinition(
          "validate",
          parameters.ParameterTypeBool,
          parameters.WithHelp("Validate templates before generation"),
          parameters.WithDefault(true),
        ),
      ),
      cmds.WithArguments(
        parameters.NewParameterDefinition(
          "spec-file",
          parameters.ParameterTypeString,
          parameters.WithHelp("YAML specification file"),
          parameters.WithRequired(true),
        ),
      ),
    ),
  }, nil
}

func (c *GenerateCommand) RunIntoGlazeProcessor(
  ctx context.Context,
  parsedLayers *layers.ParsedLayers,
  gp middlewares.Processor,
) error {
  type Settings struct {
    OutputDir string `glazed.parameter:"output-dir"`
    Validate  bool   `glazed.parameter:"validate"`
    SpecFile  string `glazed.parameter:"spec-file"`
  }
  
  settings := &Settings{}
  if err := parsedLayers.InitializeStruct("default", settings); err != nil {
    return err
  }

  // Load and process YAML spec
  spec, err := loadSpec(settings.SpecFile)
  if err != nil {
    return err
  }

  if settings.Validate {
    if err := validateTemplates(spec); err != nil {
      return err
    }
  }


  // ...
}

// In main.go setup:
func main() {
  rootCmd := cli.NewRootCommand()
  
  generateCmd, _ := NewGenerateCommand()
  glazedCmd, _ := cli.BuildCobraCommandFromCommand(generateCmd)
  
  rootCmd.AddCommand(glazedCmd)
  
  if err := rootCmd.Execute(); err != nil {
    os.Exit(1)
  }
}
```

Key changes from tutorial integration:
1. Using Glazed's parameter system for flags/arguments
2. Structured command description with validation
3. Layer parsing for settings
4. Glaze processor integration for output handling
5. Proper error handling patterns

The command now supports:
- Standardized flag parsing with help/validation
- Structured output through Glaze processors
- Layer-based configuration system
- Better error handling and template validation
- Consistent CLI structure with other Glazed commands

Example usage:
```bash
mailgen generate -o ./emails spec.yaml
mailgen generate --validate=false ./test-spec.yml
```

Would you like me to update any other sections or add more Glazed-specific implementation details?