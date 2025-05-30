package dsl

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

// Operator represents a boolean logic operator
type Operator string

const (
	OperatorAnd Operator = "and"
	OperatorOr  Operator = "or"
	OperatorNot Operator = "not"
)

// Rule represents a complete IMAP DSL rule
type Rule struct {
	Name        string       `yaml:"name"`
	Description string       `yaml:"description"`
	Search      SearchConfig `yaml:"search"`
	Output      OutputConfig `yaml:"output"`
	Actions     ActionConfig `yaml:"actions,omitempty"`
}

// Validate checks if the rule is valid
func (r *Rule) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("rule name is required")
	}

	if err := r.Search.Validate(); err != nil {
		return fmt.Errorf("invalid search config: %w", err)
	}

	if err := r.Output.Validate(); err != nil {
		return fmt.Errorf("invalid output config: %w", err)
	}

	// Validate actions if present
	if err := r.Actions.Validate(); err != nil {
		return fmt.Errorf("invalid actions config: %w", err)
	}

	return nil
}

// SearchConfig defines search criteria
type SearchConfig struct {
	// Date-based search
	Since      string `yaml:"since,omitempty"`
	Before     string `yaml:"before,omitempty"`
	On         string `yaml:"on,omitempty"`
	WithinDays int    `yaml:"within_days,omitempty"`

	// Header-based search
	From            string          `yaml:"from,omitempty"`
	To              string          `yaml:"to,omitempty"`
	Cc              string          `yaml:"cc,omitempty"`
	Bcc             string          `yaml:"bcc,omitempty"`
	Subject         string          `yaml:"subject,omitempty"`
	SubjectContains string          `yaml:"subject_contains,omitempty"`
	Header          *HeaderCriteria `yaml:"header,omitempty"`

	// Content-based search
	BodyContains string `yaml:"body_contains,omitempty"`
	Text         string `yaml:"text,omitempty"`

	// Flag-based search
	Flags *FlagCriteria `yaml:"flags,omitempty"`

	// Size-based search
	Size *SizeCriteria `yaml:"size,omitempty"`

	// Complex conditions with boolean operators
	Operator   Operator              `yaml:"operator,omitempty"`
	Conditions []ComplexSearchConfig `yaml:"conditions,omitempty"`
}

// ComplexSearchConfig defines a search condition that can contain nested conditions
type ComplexSearchConfig struct {
	// Base search criteria fields (same as SearchConfig fields)
	SearchConfig `yaml:",inline"`
}

// HeaderCriteria defines criteria for searching specific headers
type HeaderCriteria struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

// FlagCriteria defines criteria for searching by flags
type FlagCriteria struct {
	Has    []string `yaml:"has,omitempty"`
	NotHas []string `yaml:"not_has,omitempty"`
}

// SizeCriteria defines criteria for searching by message size
type SizeCriteria struct {
	LargerThan  string `yaml:"larger_than,omitempty"`
	SmallerThan string `yaml:"smaller_than,omitempty"`
}

// Validate checks if the search config is valid
func (s *SearchConfig) Validate() error {
	// Check date criteria
	if s.Since != "" {
		if _, err := parseDate(s.Since); err != nil {
			return fmt.Errorf("invalid 'since' date: %w", err)
		}
	}

	if s.Before != "" {
		if _, err := parseDate(s.Before); err != nil {
			return fmt.Errorf("invalid 'before' date: %w", err)
		}
	}

	if s.On != "" {
		if _, err := parseDate(s.On); err != nil {
			return fmt.Errorf("invalid 'on' date: %w", err)
		}
	}

	// Check header criteria
	if s.Header != nil {
		if s.Header.Name == "" {
			return fmt.Errorf("header name is required when using header search")
		}
	}

	// Check flag criteria
	if s.Flags != nil {
		for _, flag := range s.Flags.Has {
			if !isValidFlag(flag) {
				return fmt.Errorf("invalid flag in 'has' list: %s", flag)
			}
		}

		for _, flag := range s.Flags.NotHas {
			if !isValidFlag(flag) {
				return fmt.Errorf("invalid flag in 'not_has' list: %s", flag)
			}
		}
	}

	// Check size criteria
	if s.Size != nil {
		if s.Size.LargerThan != "" {
			if _, err := parseSize(s.Size.LargerThan); err != nil {
				return fmt.Errorf("invalid 'larger_than' size: %w", err)
			}
		}

		if s.Size.SmallerThan != "" {
			if _, err := parseSize(s.Size.SmallerThan); err != nil {
				return fmt.Errorf("invalid 'smaller_than' size: %w", err)
			}
		}
	}

	// Validate complex conditions
	if s.Operator != "" {
		if s.Operator != OperatorAnd && s.Operator != OperatorOr && s.Operator != OperatorNot {
			return fmt.Errorf("invalid operator: %s (must be 'and', 'or', or 'not')", s.Operator)
		}

		if len(s.Conditions) == 0 {
			return fmt.Errorf("operator %s specified but no conditions provided", s.Operator)
		}

		// NOT operator should have exactly one condition
		if s.Operator == OperatorNot && len(s.Conditions) > 1 {
			return fmt.Errorf("operator 'not' can only have one condition, but %d were provided", len(s.Conditions))
		}

		// Validate each nested condition
		for i, condition := range s.Conditions {
			if err := condition.Validate(); err != nil {
				return fmt.Errorf("invalid condition at index %d: %w", i, err)
			}
		}
	}

	return nil
}

// Validate checks if the complex search config is valid
func (c *ComplexSearchConfig) Validate() error {
	// Validate base criteria
	return c.SearchConfig.Validate()
}

// OutputConfig defines output formatting
type OutputConfig struct {
	Format    string        `yaml:"format,omitempty"`     // json, text, table
	Limit     int           `yaml:"limit,omitempty"`      // Maximum number of messages to return
	Offset    int           `yaml:"offset,omitempty"`     // Number of messages to skip for pagination
	AfterUID  uint32        `yaml:"after_uid,omitempty"`  // Fetch messages with UIDs greater than this value
	BeforeUID uint32        `yaml:"before_uid,omitempty"` // Fetch messages with UIDs less than this value
	Fields    []interface{} `yaml:"fields,omitempty"`
}

// Validate checks if the output config is valid
func (o *OutputConfig) Validate() error {
	if o.Format != "" && o.Format != "json" && o.Format != "text" && o.Format != "table" {
		return fmt.Errorf("invalid format: %s (must be 'json', 'text', or 'table')", o.Format)
	}

	if len(o.Fields) == 0 {
		return fmt.Errorf("at least one output field is required")
	}

	if o.Limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}

	// Validate fields
	for _, fieldInterface := range o.Fields {
		field, ok := fieldInterface.(Field)
		if !ok {
			continue
		}

		// Validate mime_parts field
		if field.Name == "mime_parts" && field.Content != nil {
			if field.Content.Mode != "" &&
				field.Content.Mode != "text_only" &&
				field.Content.Mode != "full" &&
				field.Content.Mode != "filter" {
				return fmt.Errorf("invalid mime_parts mode: %s (must be 'text_only', 'full', or 'filter')", field.Content.Mode)
			}

			if field.Content.Mode == "filter" && len(field.Content.Types) == 0 {
				return fmt.Errorf("mime_parts types must be specified when mode is 'filter'")
			}
		}
	}

	return nil
}

// UnmarshalYAML implements custom unmarshaling for fields
func (o *OutputConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Define a temporary struct to unmarshal into
	type tempOutputConfig struct {
		Format string        `yaml:"format"`
		Limit  int           `yaml:"limit"`
		Fields []interface{} `yaml:"fields"`
	}

	// Unmarshal into the temporary struct
	var temp tempOutputConfig
	if err := unmarshal(&temp); err != nil {
		return err
	}

	// Copy the simple fields
	o.Format = temp.Format
	o.Limit = temp.Limit
	o.Fields = make([]interface{}, len(temp.Fields))

	// Process each field
	for i, field := range temp.Fields {
		switch f := field.(type) {
		case string:
			// Simple field like "subject", "from", etc.
			o.Fields[i] = Field{Name: f}
		case map[string]interface{}:
			// Complex field like body: {type: "text/plain", max_length: 1000}
			if contentMap, ok := f["body"].(map[string]interface{}); ok {
				contentField := &ContentField{
					ShowContent: true, // Default to showing content for body
				}
				if t, ok := contentMap["type"].(string); ok {
					contentField.Type = t
				}
				if ml, ok := contentMap["max_length"].(int); ok {
					contentField.MaxLength = ml
				}
				if ml, ok := contentMap["min_length"].(int); ok {
					contentField.MinLength = ml
				}
				o.Fields[i] = Field{Name: "body", Content: contentField}
			} else if contentMap, ok := f["mime_parts"].(map[string]interface{}); ok {
				contentField := &ContentField{
					ShowTypes: true, // Default to showing types for mime_parts
				}
				if sc, ok := contentMap["show_content"].(bool); ok {
					contentField.ShowContent = sc
				}
				if st, ok := contentMap["show_types"].(bool); ok {
					contentField.ShowTypes = st
				}
				if mode, ok := contentMap["mode"].(string); ok {
					contentField.Mode = mode
				}
				if t, ok := contentMap["type"].(string); ok {
					contentField.Type = t
				}
				if ml, ok := contentMap["max_length"].(int); ok {
					contentField.MaxLength = ml
				}
				if ml, ok := contentMap["min_length"].(int); ok {
					contentField.MinLength = ml
				}
				if types, ok := contentMap["types"].([]interface{}); ok {
					contentField.Types = make([]string, len(types))
					for j, t := range types {
						contentField.Types[j] = t.(string)
					}
				}
				o.Fields[i] = Field{Name: "mime_parts", Content: contentField}
			} else {
				// Just store as is for now
				o.Fields[i] = field
			}
		default:
			// Just store as is
			o.Fields[i] = field
		}
	}

	return nil
}

// Field represents an output field, which can be a simple string or complex field
type Field struct {
	Name    string        `yaml:"name"`
	Content *ContentField `yaml:"content,omitempty"`
	// More field types will be added later
}

// ContentField represents content output configuration for both body and MIME parts
type ContentField struct {
	Type        string   `yaml:"type,omitempty"`         // MIME type for body or filter for MIME parts
	MaxLength   int      `yaml:"max_length,omitempty"`   // Maximum length of content to return
	MinLength   int      `yaml:"min_length,omitempty"`   // Minimum length of content to return
	Mode        string   `yaml:"mode,omitempty"`         // "text_only", "full", "filter", or empty for body
	Types       []string `yaml:"types,omitempty"`        // List of MIME types to include when mode is "filter"
	ShowTypes   bool     `yaml:"show_types,omitempty"`   // Whether to show MIME types in output
	ShowContent bool     `yaml:"show_content,omitempty"` // Whether to show content in output (default true)
}

func (c *ContentField) ShouldInclude(mediaType string) bool {
	log.Debug().
		Str("media_type", mediaType).
		Str("mode", c.Mode).
		Strs("allowed_types", c.Types).
		Msg("Checking if MIME type should be included")

	switch c.Mode {
	case "text_only":
		result := strings.HasPrefix(mediaType, "text/plain")
		log.Debug().
			Str("media_type", mediaType).
			Bool("result", result).
			Msg("text_only mode check")
		return result
	case "filter":
		if len(c.Types) == 0 {
			log.Debug().Msg("No types specified in filter mode, including all")
			return true
		}
		for _, allowedType := range c.Types {
			log.Debug().
				Str("media_type", mediaType).
				Str("allowed_type", allowedType).
				Msg("Checking type match")

			if strings.HasSuffix(allowedType, "/*") {
				prefix := strings.TrimSuffix(allowedType, "/*")
				if strings.HasPrefix(mediaType, prefix+"/") {
					log.Debug().
						Str("media_type", mediaType).
						Str("prefix", prefix).
						Msg("Wildcard match found")
					return true
				}
			} else if mediaType == allowedType {
				log.Debug().
					Str("media_type", mediaType).
					Str("allowed_type", allowedType).
					Msg("Exact match found")
				return true
			}
		}
		log.Debug().
			Str("media_type", mediaType).
			Strs("allowed_types", c.Types).
			Msg("No matching type found")
		return false
	default: // "full" or empty
		log.Debug().
			Str("mode", c.Mode).
			Msg("Using default mode (full), including all types")
		return true
	}
}

// Helper functions

// parseSize parses a size string with units (B, K, M, G)
func parseSize(sizeStr string) (int64, error) {
	re := regexp.MustCompile(`^(\d+)([BKMG])?$`)
	matches := re.FindStringSubmatch(sizeStr)

	if matches == nil {
		return 0, fmt.Errorf("invalid size format: %s (expected format: 100B, 10K, 5M, 1G)", sizeStr)
	}

	size, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size number: %s", matches[1])
	}

	unit := matches[2]
	switch unit {
	case "B", "":
		// Size is already in bytes
	case "K":
		size *= 1024
	case "M":
		size *= 1024 * 1024
	case "G":
		size *= 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("invalid size unit: %s (expected B, K, M, or G)", unit)
	}

	return size, nil
}

// isValidFlag checks if a flag name is valid
func isValidFlag(flag string) bool {
	// Standard IMAP flags
	standardFlags := map[string]bool{
		"seen":      true,
		"answered":  true,
		"flagged":   true,
		"deleted":   true,
		"draft":     true,
		"recent":    true,
		"important": true,
	}

	// Convert to lowercase for case-insensitive comparison
	flagLower := strings.ToLower(flag)

	// Check if it's a standard flag
	if standardFlags[flagLower] {
		return true
	}

	// Check if it's a custom flag (starts with backslash or dollar sign)
	if strings.HasPrefix(flag, "\\") || strings.HasPrefix(flag, "$") {
		return true
	}

	// Allow keywords (alphanumeric plus some special chars)
	match, _ := regexp.MatchString(`^[a-zA-Z0-9_.-]+$`, flag)
	return match
}

// ActionConfig defines actions to perform on matched messages
type ActionConfig struct {
	// Flag operations
	Flags *FlagActions `yaml:"flags,omitempty"`

	// Move/Copy operations
	MoveTo string `yaml:"move_to,omitempty"`
	CopyTo string `yaml:"copy_to,omitempty"`

	// Delete operation
	Delete interface{} `yaml:"delete,omitempty"` // Can be bool or DeleteConfig

	// Export operation
	Export *ExportConfig `yaml:"export,omitempty"`
}

// FlagActions defines add/remove flag operations
type FlagActions struct {
	Add    []string `yaml:"add,omitempty"`
	Remove []string `yaml:"remove,omitempty"`
}

// Validate checks if the action config is valid
func (a *ActionConfig) Validate() error {
	// Validate flag actions
	if a.Flags != nil {
		if err := a.Flags.Validate(); err != nil {
			return fmt.Errorf("invalid flag actions: %w", err)
		}
	}

	// Validate export config
	if a.Export != nil {
		if err := a.Export.Validate(); err != nil {
			return fmt.Errorf("invalid export config: %w", err)
		}
	}

	// Validate delete configuration
	if a.Delete != nil {
		switch deleteConfig := a.Delete.(type) {
		case bool:
			// Boolean is always valid
		case map[string]interface{}:
			// If it's a map, it should be convertible to DeleteConfig
			if _, ok := deleteConfig["trash"]; !ok {
				return fmt.Errorf("delete config must have a 'trash' field")
			}
		default:
			return fmt.Errorf("delete config must be a boolean or an object with a 'trash' field")
		}
	}

	return nil
}

// Validate checks if the flag actions are valid
func (f *FlagActions) Validate() error {
	// Validate add flags
	for _, flag := range f.Add {
		if !isValidFlag(flag) {
			return fmt.Errorf("invalid flag in 'add' list: %s", flag)
		}
	}

	// Validate remove flags
	for _, flag := range f.Remove {
		if !isValidFlag(flag) {
			return fmt.Errorf("invalid flag in 'remove' list: %s", flag)
		}
	}

	return nil
}

// Validate checks if the export config is valid
func (e *ExportConfig) Validate() error {
	// Validate format
	if e.Format != "" && e.Format != "eml" && e.Format != "mbox" {
		return fmt.Errorf("invalid format: %s (must be 'eml' or 'mbox')", e.Format)
	}

	// If no format is specified, default to "eml"
	if e.Format == "" {
		e.Format = "eml"
	}

	return nil
}

// DeleteConfig provides options for delete operations
type DeleteConfig struct {
	Trash bool `yaml:"trash,omitempty"` // If true, move to trash; if false, delete permanently
}

// ExportConfig defines options for exporting messages
type ExportConfig struct {
	Format           string `yaml:"format,omitempty"`            // eml, mbox
	Directory        string `yaml:"directory,omitempty"`         // Where to save files
	FilenameTemplate string `yaml:"filename_template,omitempty"` // Template for filenames
}
