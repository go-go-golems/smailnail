package dsl

import (
	"testing"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/stretchr/testify/assert"
)

func TestBuildComplexSearchCriteria(t *testing.T) {
	testDate, _ := time.Parse("2006-01-02", "2024-01-15")
	startOfDay := time.Date(testDate.Year(), testDate.Month(), testDate.Day(), 0, 0, 0, 0, testDate.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)

	tests := []struct {
		name              string
		config            SearchConfig
		validateCriteria  func(*testing.T, *imap.SearchCriteria)
		shouldError       bool
		expectedErrorText string
	}{
		{
			name: "Simple AND condition",
			config: SearchConfig{
				Operator: OperatorAnd,
				Conditions: []ComplexSearchConfig{
					{SearchConfig: SearchConfig{From: "person@example.com"}},
					{SearchConfig: SearchConfig{Subject: "Important"}},
				},
			},
			validateCriteria: func(t *testing.T, criteria *imap.SearchCriteria) {
				// AND creates a combined criteria with all conditions applied
				assert.Len(t, criteria.Header, 2)
				assert.Equal(t, "From", criteria.Header[0].Key)
				assert.Equal(t, "person@example.com", criteria.Header[0].Value)
				assert.Equal(t, "Subject", criteria.Header[1].Key)
				assert.Equal(t, "Important", criteria.Header[1].Value)
			},
			shouldError: false,
		},
		{
			name: "Simple OR condition",
			config: SearchConfig{
				Operator: OperatorOr,
				Conditions: []ComplexSearchConfig{
					{SearchConfig: SearchConfig{From: "person1@example.com"}},
					{SearchConfig: SearchConfig{From: "person2@example.com"}},
				},
			},
			validateCriteria: func(t *testing.T, criteria *imap.SearchCriteria) {
				assert.Len(t, criteria.Or, 1)
				assert.Len(t, criteria.Or[0][0].Header, 1)
				assert.Equal(t, "From", criteria.Or[0][0].Header[0].Key)
				assert.Equal(t, "person1@example.com", criteria.Or[0][0].Header[0].Value)
				assert.Len(t, criteria.Or[0][1].Header, 1)
				assert.Equal(t, "From", criteria.Or[0][1].Header[0].Key)
				assert.Equal(t, "person2@example.com", criteria.Or[0][1].Header[0].Value)
			},
			shouldError: false,
		},
		{
			name: "Simple NOT condition",
			config: SearchConfig{
				Operator: OperatorNot,
				Conditions: []ComplexSearchConfig{
					{SearchConfig: SearchConfig{
						From: "spam@example.com",
					}},
				},
			},
			validateCriteria: func(t *testing.T, criteria *imap.SearchCriteria) {
				assert.Len(t, criteria.Not, 1)
				assert.Len(t, criteria.Not[0].Header, 1)
				assert.Equal(t, "From", criteria.Not[0].Header[0].Key)
				assert.Equal(t, "spam@example.com", criteria.Not[0].Header[0].Value)
			},
			shouldError: false,
		},
		{
			name: "Complex nested condition (AND with OR subclause)",
			config: SearchConfig{
				Operator: OperatorAnd,
				Conditions: []ComplexSearchConfig{
					{SearchConfig: SearchConfig{From: "team@company.com"}},
					{SearchConfig: SearchConfig{
						Operator: OperatorOr,
						Conditions: []ComplexSearchConfig{
							{SearchConfig: SearchConfig{Subject: "Urgent"}},
							{SearchConfig: SearchConfig{
								Flags: &FlagCriteria{
									Has: []string{"flagged"},
								},
							}},
						},
					}},
				},
			},
			validateCriteria: func(t *testing.T, criteria *imap.SearchCriteria) {
				// Expected structure:
				// criteria.Header has "From: team@company.com"
				// criteria.Or has one pair: ["Subject: Urgent", "Flag: \Flagged"]
				assert.Len(t, criteria.Header, 1)
				assert.Equal(t, "From", criteria.Header[0].Key)
				assert.Equal(t, "team@company.com", criteria.Header[0].Value)

				assert.Len(t, criteria.Or, 1)
				assert.Len(t, criteria.Or[0][0].Header, 1)
				assert.Equal(t, "Subject", criteria.Or[0][0].Header[0].Key)
				assert.Equal(t, "Urgent", criteria.Or[0][0].Header[0].Value)

				assert.Len(t, criteria.Or[0][1].Flag, 1)
				assert.Equal(t, imap.Flag("\\Flagged"), criteria.Or[0][1].Flag[0])
			},
			shouldError: false,
		},
		{
			name: "Complex nested condition (NOT with AND subclause)",
			config: SearchConfig{
				Operator: OperatorNot,
				Conditions: []ComplexSearchConfig{
					{SearchConfig: SearchConfig{
						Operator: OperatorAnd,
						Conditions: []ComplexSearchConfig{
							{SearchConfig: SearchConfig{From: "noreply@example.com"}},
							{SearchConfig: SearchConfig{Subject: "Newsletter"}},
						},
					}},
				},
			},
			validateCriteria: func(t *testing.T, criteria *imap.SearchCriteria) {
				assert.Len(t, criteria.Not, 1)
				assert.Len(t, criteria.Not[0].Header, 2)
				assert.Equal(t, "From", criteria.Not[0].Header[0].Key)
				assert.Equal(t, "noreply@example.com", criteria.Not[0].Header[0].Value)
				assert.Equal(t, "Subject", criteria.Not[0].Header[1].Key)
				assert.Equal(t, "Newsletter", criteria.Not[0].Header[1].Value)
			},
			shouldError: false,
		},
		{
			name: "Mixed simple and complex conditions",
			config: SearchConfig{
				Operator: OperatorAnd,
				Conditions: []ComplexSearchConfig{
					{SearchConfig: SearchConfig{
						On: "2024-01-15",
					}},
					{SearchConfig: SearchConfig{
						Operator: OperatorOr,
						Conditions: []ComplexSearchConfig{
							{SearchConfig: SearchConfig{From: "person1@example.com"}},
							{SearchConfig: SearchConfig{From: "person2@example.com"}},
						},
					}},
				},
			},
			validateCriteria: func(t *testing.T, criteria *imap.SearchCriteria) {
				// First condition: date range
				assert.Equal(t, startOfDay, criteria.Since)
				assert.Equal(t, endOfDay, criteria.Before)

				// Second condition: OR of two From conditions
				assert.Len(t, criteria.Or, 1)
				assert.Len(t, criteria.Or[0][0].Header, 1)
				assert.Equal(t, "From", criteria.Or[0][0].Header[0].Key)
				assert.Equal(t, "person1@example.com", criteria.Or[0][0].Header[0].Value)
				assert.Len(t, criteria.Or[0][1].Header, 1)
				assert.Equal(t, "From", criteria.Or[0][1].Header[0].Key)
				assert.Equal(t, "person2@example.com", criteria.Or[0][1].Header[0].Value)
			},
			shouldError: false,
		},
		{
			name: "Empty conditions list for AND",
			config: SearchConfig{
				Operator:   OperatorAnd,
				Conditions: []ComplexSearchConfig{},
			},
			shouldError:       true,
			expectedErrorText: "empty conditions list for AND operator",
		},
		{
			name: "Empty conditions list for OR",
			config: SearchConfig{
				Operator:   OperatorOr,
				Conditions: []ComplexSearchConfig{},
			},
			shouldError:       true,
			expectedErrorText: "empty conditions list for OR operator",
		},
		{
			name: "Empty conditions list for NOT",
			config: SearchConfig{
				Operator:   OperatorNot,
				Conditions: []ComplexSearchConfig{},
			},
			shouldError:       true,
			expectedErrorText: "NOT operator requires at least one condition",
		},
		{
			name: "Invalid operator",
			config: SearchConfig{
				Operator: "invalid",
				Conditions: []ComplexSearchConfig{
					{SearchConfig: SearchConfig{From: "person@example.com"}},
				},
			},
			shouldError:       true,
			expectedErrorText: "unsupported operator: invalid",
		},
		{
			name: "NOT with multiple conditions",
			config: SearchConfig{
				Operator: OperatorNot,
				Conditions: []ComplexSearchConfig{
					{SearchConfig: SearchConfig{From: "person1@example.com"}},
					{SearchConfig: SearchConfig{From: "person2@example.com"}},
				},
			},
			validateCriteria: nil, // No validation needed for error case
			shouldError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			criteria, _, err := BuildSearchCriteria(tt.config, nil)

			if tt.shouldError {
				assert.Error(t, err)
				if tt.expectedErrorText != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorText)
				}
				return
			}

			assert.NoError(t, err)
			if tt.validateCriteria != nil {
				tt.validateCriteria(t, criteria)
			}
		})
	}
}

func TestComplexSearchConfigValidation(t *testing.T) {
	tests := []struct {
		name              string
		config            SearchConfig
		shouldError       bool
		expectedErrorText string
	}{
		{
			name: "Valid AND condition",
			config: SearchConfig{
				Operator: OperatorAnd,
				Conditions: []ComplexSearchConfig{
					{SearchConfig: SearchConfig{From: "person@example.com"}},
					{SearchConfig: SearchConfig{Subject: "Important"}},
				},
			},
			shouldError: false,
		},
		{
			name: "Invalid operator",
			config: SearchConfig{
				Operator: "invalid",
				Conditions: []ComplexSearchConfig{
					{SearchConfig: SearchConfig{From: "person@example.com"}},
				},
			},
			shouldError:       true,
			expectedErrorText: "invalid operator: invalid (must be 'and', 'or', or 'not')",
		},
		{
			name: "No conditions provided",
			config: SearchConfig{
				Operator:   OperatorAnd,
				Conditions: []ComplexSearchConfig{},
			},
			shouldError:       true,
			expectedErrorText: "operator and specified but no conditions provided",
		},
		{
			name: "NOT with multiple conditions",
			config: SearchConfig{
				Operator: OperatorNot,
				Conditions: []ComplexSearchConfig{
					{SearchConfig: SearchConfig{From: "person1@example.com"}},
					{SearchConfig: SearchConfig{From: "person2@example.com"}},
				},
			},
			shouldError:       true,
			expectedErrorText: "operator 'not' can only have one condition",
		},
		{
			name: "Invalid nested condition",
			config: SearchConfig{
				Operator: OperatorAnd,
				Conditions: []ComplexSearchConfig{
					{SearchConfig: SearchConfig{From: "person@example.com"}},
					{SearchConfig: SearchConfig{
						Size: &SizeCriteria{
							LargerThan: "invalid",
						},
					}},
				},
			},
			shouldError:       true,
			expectedErrorText: "invalid condition at index 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.shouldError {
				assert.Error(t, err)
				if tt.expectedErrorText != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorText)
				}
				return
			}

			assert.NoError(t, err)
		})
	}
}
