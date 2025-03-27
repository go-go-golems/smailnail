package dsl

import (
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/rs/zerolog/log"
)

// BuildSearchCriteria converts SearchConfig to imap.SearchCriteria and returns appropriate SearchOptions
func BuildSearchCriteria(config SearchConfig, outputConfig *OutputConfig) (*imap.SearchCriteria, *imap.SearchOptions, error) {
	criteria := &imap.SearchCriteria{}
	options := &imap.SearchOptions{}

	// Process complex conditions first
	if config.Operator != "" {
		// Validate that we have conditions for the operator
		if len(config.Conditions) == 0 {
			switch config.Operator {
			case OperatorAnd:
				return nil, nil, fmt.Errorf("empty conditions list for AND operator")
			case OperatorOr:
				return nil, nil, fmt.Errorf("empty conditions list for OR operator")
			case OperatorNot:
				return nil, nil, fmt.Errorf("NOT operator requires at least one condition")
			default:
				return nil, nil, fmt.Errorf("unsupported operator: %s", config.Operator)
			}
		}

		return buildComplexSearchCriteria(config, outputConfig)
	}

	// Process date criteria
	if config.Since != "" {
		since, err := parseDate(config.Since)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid 'since' date: %w", err)
		}
		criteria.Since = since
	}

	if config.Before != "" {
		before, err := parseDate(config.Before)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid 'before' date: %w", err)
		}
		criteria.Before = before
	}

	if config.On != "" {
		on, err := parseDate(config.On)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid 'on' date: %w", err)
		}

		// For "on" date, we need to set both since and before to cover the entire day
		// Since = start of the day, Before = start of the next day
		startOfDay := time.Date(on.Year(), on.Month(), on.Day(), 0, 0, 0, 0, on.Location())
		endOfDay := startOfDay.AddDate(0, 0, 1)

		criteria.Since = startOfDay
		criteria.Before = endOfDay
	}

	if config.WithinDays > 0 {
		// Calculate date from N days ago
		since := time.Now().AddDate(0, 0, -config.WithinDays)
		// Set to start of that day
		since = time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, since.Location())
		criteria.Since = since
	}

	// Process header-based search criteria
	if config.From != "" {
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{
			Key:   "From",
			Value: config.From,
		})
	}

	if config.To != "" {
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{
			Key:   "To",
			Value: config.To,
		})
	}

	if config.Cc != "" {
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{
			Key:   "Cc",
			Value: config.Cc,
		})
	}

	if config.Bcc != "" {
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{
			Key:   "Bcc",
			Value: config.Bcc,
		})
	}

	if config.Subject != "" {
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{
			Key:   "Subject",
			Value: config.Subject,
		})
	}

	if config.SubjectContains != "" {
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{
			Key:   "Subject",
			Value: config.SubjectContains,
		})
	}

	if config.Header != nil && config.Header.Name != "" {
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{
			Key:   config.Header.Name,
			Value: config.Header.Value,
		})
	}

	// Process content-based search criteria
	if config.BodyContains != "" {
		criteria.Body = []string{config.BodyContains}
	}

	if config.Text != "" {
		criteria.Text = []string{config.Text}
	}

	// Process flag-based search criteria
	if config.Flags != nil {
		if len(config.Flags.Has) > 0 {
			for _, flag := range config.Flags.Has {
				// Convert flag name to IMAP format if needed
				imapFlag := convertToIMAPFlag(flag)
				criteria.Flag = append(criteria.Flag, imap.Flag(imapFlag))
			}
		}

		if len(config.Flags.NotHas) > 0 {
			for _, flag := range config.Flags.NotHas {
				// Convert flag name to IMAP format if needed
				imapFlag := convertToIMAPFlag(flag)
				criteria.NotFlag = append(criteria.NotFlag, imap.Flag(imapFlag))
			}
		}
	}

	// Process size-based search criteria
	if config.Size != nil {
		if config.Size.LargerThan != "" {
			size, err := parseSize(config.Size.LargerThan)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid 'larger_than' size: %w", err)
			}

			criteria.Larger = int64(size)
		}

		if config.Size.SmallerThan != "" {
			size, err := parseSize(config.Size.SmallerThan)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid 'smaller_than' size: %w", err)
			}

			criteria.Smaller = int64(size)
		}
	}

	// Process UID-based pagination if provided in the output config
	if outputConfig != nil {
		if outputConfig.AfterUID > 0 || outputConfig.BeforeUID > 0 {
			// Create a UID range for pagination
			uidSet := imap.UIDSet{}

			if outputConfig.AfterUID > 0 && outputConfig.BeforeUID > 0 {
				// Between AfterUID+1 and BeforeUID-1
				uidSet.AddRange(imap.UID(outputConfig.AfterUID+1), imap.UID(outputConfig.BeforeUID-1))
			} else if outputConfig.AfterUID > 0 {
				// Greater than AfterUID
				uidSet.AddRange(imap.UID(outputConfig.AfterUID+1), 0) // 0 means "*" (unlimited) in go-imap
			} else if outputConfig.BeforeUID > 0 {
				// Less than BeforeUID
				uidSet.AddRange(imap.UID(1), imap.UID(outputConfig.BeforeUID-1))
			}

			criteria.UID = []imap.UIDSet{uidSet}
		}

		// Set search options to optimize the search
		// Only request as many results as needed (limit + offset)
		if outputConfig.Limit > 0 {
			// We need to always set ReturnAll to true to get sequence numbers
			// that we can use for fetching the messages
			options.ReturnAll = true

			// We also want to get a count of total results if possible
			options.ReturnCount = true

			log.Debug().
				Int("limit", outputConfig.Limit).
				Int("offset", outputConfig.Offset).
				Bool("return_all", options.ReturnAll).
				Bool("return_count", options.ReturnCount).
				Msg("Search options set")
		}
	}

	return criteria, options, nil
}

// buildComplexSearchCriteria handles the conversion of complex nested conditions
func buildComplexSearchCriteria(config SearchConfig, outputConfig *OutputConfig) (*imap.SearchCriteria, *imap.SearchOptions, error) {
	options := &imap.SearchOptions{}

	// Set search options for pagination
	if outputConfig != nil {
		// Copy the same pagination logic as in BuildSearchCriteria
		if outputConfig.Limit > 0 {
			options.ReturnAll = true
			options.ReturnCount = true
		}
	}

	// Process nested conditions based on operator
	switch config.Operator {
	case OperatorAnd:
		return buildAndCondition(config.Conditions, outputConfig)
	case OperatorOr:
		return buildOrCondition(config.Conditions, outputConfig)
	case OperatorNot:
		if len(config.Conditions) == 0 {
			return nil, nil, fmt.Errorf("NOT operator requires at least one condition")
		}

		// NOT operator should have exactly one condition
		if len(config.Conditions) > 1 {
			return nil, nil, fmt.Errorf("operator 'not' can only have one condition, but %d were provided", len(config.Conditions))
		}

		return buildNotCondition(config.Conditions[0], outputConfig)
	default:
		return nil, nil, fmt.Errorf("unsupported operator: %s", config.Operator)
	}
}

// buildAndCondition creates a criteria with AND logic for multiple conditions
func buildAndCondition(conditions []ComplexSearchConfig, outputConfig *OutputConfig) (*imap.SearchCriteria, *imap.SearchOptions, error) {
	if len(conditions) == 0 {
		return nil, nil, fmt.Errorf("empty conditions list for AND operator")
	}

	// Start with the first condition
	mainCriteria, options, err := buildSingleCondition(conditions[0], outputConfig)
	if err != nil {
		return nil, nil, err
	}

	// AND with each subsequent condition
	for i := 1; i < len(conditions); i++ {
		subCriteria, _, err := buildSingleCondition(conditions[i], nil)
		if err != nil {
			return nil, nil, err
		}

		// Use the And method to combine criteria
		mainCriteria.And(subCriteria)
	}

	return mainCriteria, options, nil
}

// buildOrCondition creates a criteria with OR logic for multiple conditions
func buildOrCondition(conditions []ComplexSearchConfig, outputConfig *OutputConfig) (*imap.SearchCriteria, *imap.SearchOptions, error) {
	if len(conditions) == 0 {
		return nil, nil, fmt.Errorf("empty conditions list for OR operator")
	}

	// Handle special case of a single condition
	if len(conditions) == 1 {
		return buildSingleCondition(conditions[0], outputConfig)
	}

	// Create the result criteria
	resultCriteria := &imap.SearchCriteria{}
	var options *imap.SearchOptions

	// Process each pair of conditions and create OR operations
	for i := 0; i < len(conditions); i += 2 {
		// Get first condition
		c1, opts, err := buildSingleCondition(conditions[i], nil)
		if err != nil {
			return nil, nil, err
		}

		// Save options from first iteration
		if i == 0 && outputConfig != nil {
			options = opts
		}

		// If we have an odd number of conditions and this is the last one
		if i == len(conditions)-1 {
			// Handle special case: last single condition in odd-length list
			orPair := [2]imap.SearchCriteria{*c1, 
			resultCriteria.Or = append(resultCriteria.Or, orPair)
			continue
		}

		// Get second condition
		c2, _, err := buildSingleCondition(conditions[i+1], nil)
		if err != nil {
			return nil, nil, err
		}

		// Create OR pair
		orPair := [2]imap.SearchCriteria{*c1, *c2}
		resultCriteria.Or = append(resultCriteria.Or, orPair)
	}

	return resultCriteria, options, nil
}

// buildNotCondition creates a criteria with NOT logic for a condition
func buildNotCondition(condition ComplexSearchConfig, outputConfig *OutputConfig) (*imap.SearchCriteria, *imap.SearchOptions, error) {
	subCriteria, options, err := buildSingleCondition(condition, outputConfig)
	if err != nil {
		return nil, nil, err
	}

	// Create the result with NOT logic
	resultCriteria := &imap.SearchCriteria{
		Not: []imap.SearchCriteria{*subCriteria},
	}

	return resultCriteria, options, nil
}

// buildSingleCondition builds search criteria for a single complex condition
func buildSingleCondition(condition ComplexSearchConfig, outputConfig *OutputConfig) (*imap.SearchCriteria, *imap.SearchOptions, error) {
	// If this is itself a complex condition with an operator, recursively process it
	if condition.Operator != "" && len(condition.Conditions) > 0 {
		return buildComplexSearchCriteria(condition.SearchConfig, outputConfig)
	}

	// Otherwise, treat it as a flat condition
	return BuildSearchCriteria(condition.SearchConfig, outputConfig)
}

// parseDate parses a date string in RFC3339 or ISO8601 format
func parseDate(dateStr string) (time.Time, error) {
	// Try RFC3339 format first
	t, err := time.Parse(time.RFC3339, dateStr)
	if err == nil {
		return t, nil
	}

	// Try ISO8601 date-only format
	t, err = time.Parse("2006-01-02", dateStr)
	if err == nil {
		return t, nil
	}

	// Try a few more common formats
	formats := []string{
		"2006/01/02",
		"01/02/2006",
		"02/01/2006",
		"Jan 2, 2006",
		"2 Jan 2006",
		time.RFC822,
		time.RFC1123,
	}

	for _, format := range formats {
		t, err = time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse date: %s", dateStr)
}

// convertToIMAPFlag converts a user-friendly flag name to IMAP format
func convertToIMAPFlag(flag string) string {
	// If it already starts with \ or $, return as is
	if strings.HasPrefix(flag, "\\") || strings.HasPrefix(flag, "$") {
		return flag
	}

	// Map of standard flag names to IMAP format
	standardFlags := map[string]string{
		"seen":      "\\Seen",
		"answered":  "\\Answered",
		"flagged":   "\\Flagged",
		"deleted":   "\\Deleted",
		"draft":     "\\Draft",
		"recent":    "\\Recent",
		"important": "$Important",
	}

	// Convert to lowercase for case-insensitive comparison
	flagLower := strings.ToLower(flag)

	// Check if it's a standard flag
	if imapFlag, ok := standardFlags[flagLower]; ok {
		return imapFlag
	}

	// Return as is for custom flags
	return flag
}
