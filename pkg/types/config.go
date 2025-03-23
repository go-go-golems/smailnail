package types

import (
	"fmt"

	"github.com/pkg/errors"
)

// EmailTemplate defines the structure of an email template
type EmailTemplate struct {
	Subject string `yaml:"subject"`
	From    string `yaml:"from"`
	To      string `yaml:"to,omitempty"`
	Cc      string `yaml:"cc,omitempty"`
	Bcc     string `yaml:"bcc,omitempty"`
	ReplyTo string `yaml:"reply_to,omitempty"`
	Body    string `yaml:"body"`
}

// TemplateConfig defines the structure of the YAML configuration file
type TemplateConfig struct {
	Variables map[string]interface{}   `yaml:"variables"` // All values must be strings or []string
	Templates map[string]EmailTemplate `yaml:"templates"`
	Rules     map[string]RuleConfig    `yaml:"rules"`
	Generate  []GenerateConfig         `yaml:"generate"`
}

// RuleConfig defines a rule for generating email variations
type RuleConfig struct {
	Template   string              `yaml:"template"`
	Variations []map[string]string `yaml:"variations"` // All values must be strings
}

// GenerateConfig defines how many emails to generate using a particular rule
type GenerateConfig struct {
	Rule   string `yaml:"rule"`
	Count  int    `yaml:"count"`
	Output string `yaml:"output,omitempty"`
}

// Email represents a generated email
type Email struct {
	Subject string `json:"subject"`
	From    string `json:"from"`
	To      string `json:"to,omitempty"`
	Cc      string `json:"cc,omitempty"`
	Bcc     string `json:"bcc,omitempty"`
	ReplyTo string `json:"reply_to,omitempty"`
	Body    string `json:"body"`
}

// validateVariables ensures all values in the variables map are either strings or []string
func validateVariables(vars map[string]interface{}, path string) error {
	for key, value := range vars {
		currentPath := path
		if path != "" {
			currentPath = path + "." + key
		} else {
			currentPath = key
		}

		switch v := value.(type) {
		case string:
			// String values are allowed
			continue
		case []interface{}:
			// Check each element in the slice is a string
			for i, elem := range v {
				if _, ok := elem.(string); !ok {
					return fmt.Errorf("variable at path '%s[%d]' must be a string, got %T", currentPath, i, elem)
				}
			}
		case map[string]interface{}:
			// Recursively validate nested maps
			if err := validateVariables(v, currentPath); err != nil {
				return err
			}
		default:
			return fmt.Errorf("variable at path '%s' must be a string, string slice, or map of strings, got %T", currentPath, value)
		}
	}
	return nil
}

// Validate checks if the template configuration is valid
func (tc *TemplateConfig) Validate() error {
	// Validate variables structure
	if err := validateVariables(tc.Variables, ""); err != nil {
		return errors.Wrap(err, "invalid variables")
	}

	// Check if templates exist
	if len(tc.Templates) == 0 {
		return errors.New("no templates defined")
	}

	// Check if rules exist
	if len(tc.Rules) == 0 {
		return errors.New("no rules defined")
	}

	// Check if generate section exists
	if len(tc.Generate) == 0 {
		return errors.New("no generate section defined")
	}

	// Check if each rule refers to a valid template
	for ruleName, rule := range tc.Rules {
		if _, exists := tc.Templates[rule.Template]; !exists {
			return errors.Errorf("rule '%s' refers to undefined template '%s'", ruleName, rule.Template)
		}

		// Validate that all variation values are strings
		for i, variation := range rule.Variations {
			for key, value := range variation {
				if value == "" {
					return errors.Errorf("rule '%s' variation %d has empty value for key '%s'", ruleName, i, key)
				}
			}
		}
	}

	// Check if each generate entry refers to a valid rule
	for i, gen := range tc.Generate {
		if _, exists := tc.Rules[gen.Rule]; !exists {
			return errors.Errorf("generate entry %d refers to undefined rule '%s'", i, gen.Rule)
		}

		if gen.Count <= 0 {
			return errors.Errorf("generate entry %d has invalid count %d", i, gen.Count)
		}
	}

	return nil
}
