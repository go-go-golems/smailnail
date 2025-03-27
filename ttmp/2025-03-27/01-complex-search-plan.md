# Implementation Plan: Nested Boolean Operators for IMAP Search

## Overview

This plan outlines the implementation of complex search conditions with nested boolean operators (AND, OR, NOT) for the IMAP DSL. This feature will allow users to create sophisticated search queries as described in the "Complex Conditions" section of the IMAP DSL specification.

## Current Implementation Analysis

The current implementation in `pkg/dsl/types.go` and `pkg/dsl/search.go` supports basic search criteria but lacks support for nested boolean operators. The `SearchConfig` struct only contains flat search criteria without the ability to define complex logical relationships.

## Implementation Plan

### 1. Extend the Data Structure

- [ ] Define new types to represent complex boolean operations
- [ ] Update `SearchConfig` to support nested conditions
- [ ] Implement proper validation for nested conditions

### 2. Update the Search Builder

- [ ] Extend `BuildSearchCriteria` to handle nested boolean operations
- [ ] Implement recursive processing of nested conditions

### 3. Add Tests

- [ ] Create unit tests for nested conditions
- [ ] Add integration tests with example queries

### 4. Update Documentation

- [ ] Add examples in comments
- [ ] Update any external documentation

## Detailed Implementation Steps

### 1. Extend the Data Structure

#### 1.1 Create Complex Condition Types

We'll add new types to represent complex boolean operations in `pkg/dsl/types.go`:

```go
// Operator represents a boolean logic operator
type Operator string

const (
    OperatorAnd Operator = "and"
    OperatorOr  Operator = "or"
    OperatorNot Operator = "not"
)

// ComplexSearchConfig defines a search condition that can contain nested conditions
type ComplexSearchConfig struct {
    // Base search criteria fields (same as current SearchConfig fields)
    SearchConfig `yaml:",inline"`
    
    // Boolean operator to apply to conditions
    Operator Operator `yaml:"operator,omitempty"`
    
    // Nested conditions
    Conditions []ComplexSearchConfig `yaml:"conditions,omitempty"`
}
```

#### 1.2 Update SearchConfig

Update the `SearchConfig` struct to include support for nested conditions:

```go
// SearchConfig defines search criteria
type SearchConfig struct {
    // Existing fields...
    
    // Complex conditions with boolean operators
    Operator   Operator             `yaml:"operator,omitempty"`
    Conditions []ComplexSearchConfig `yaml:"conditions,omitempty"`
}
```

#### 1.3 Implement Validation

Add validation for complex conditions:

```go
// Validate checks if the complex search config is valid
func (c *ComplexSearchConfig) Validate() error {
    // Validate base criteria
    if err := c.SearchConfig.Validate(); err != nil {
        return err
    }
    
    // Validate operator and conditions
    if c.Operator != "" {
        if c.Operator != OperatorAnd && c.Operator != OperatorOr && c.Operator != OperatorNot {
            return fmt.Errorf("invalid operator: %s (must be 'and', 'or', or 'not')", c.Operator)
        }
        
        if len(c.Conditions) == 0 {
            return fmt.Errorf("operator %s specified but no conditions provided", c.Operator)
        }
        
        // NOT operator should have exactly one condition
        if c.Operator == OperatorNot && len(c.Conditions) > 1 {
            return fmt.Errorf("operator 'not' can only have one condition, but %d were provided", len(c.Conditions))
        }
        
        // Validate each nested condition
        for i, condition := range c.Conditions {
            if err := condition.Validate(); err != nil {
                return fmt.Errorf("invalid condition at index %d: %w", i, err)
            }
        }
    }
    
    return nil
}
```

### 2. Update the Search Builder

#### 2.1 Extend BuildSearchCriteria

Update the `BuildSearchCriteria` function in `pkg/dsl/search.go` to handle complex conditions:

```go
// BuildSearchCriteria converts SearchConfig to imap.SearchCriteria and returns appropriate SearchOptions
func BuildSearchCriteria(config SearchConfig, outputConfig *OutputConfig) (*imap.SearchCriteria, *imap.SearchOptions, error) {
    criteria := &imap.SearchCriteria{}
    options := &imap.SearchOptions{}
    
    // Process complex conditions first
    if config.Operator != "" && len(config.Conditions) > 0 {
        return buildComplexSearchCriteria(config, outputConfig)
    }
    
    // Process flat search criteria (existing code)
    // ...
    
    return criteria, options, nil
}
```

#### 2.2 Implement Complex Search Criteria Builder

Add a new function to handle complex conditions:

```go
// buildComplexSearchCriteria handles the conversion of complex nested conditions
func buildComplexSearchCriteria(config SearchConfig, outputConfig *OutputConfig) (*imap.SearchCriteria, *imap.SearchOptions, error) {
    options := &imap.SearchOptions{}
    
    // Set search options for pagination
    if outputConfig != nil {
        // (Same pagination code as in existing BuildSearchCriteria)
    }
    
    // Process nested conditions based on operator
    switch config.Operator {
    case OperatorAnd:
        return buildAndCondition(config.Conditions, outputConfig)
    case OperatorOr:
        return buildOrCondition(config.Conditions, outputConfig)
    case OperatorNot:
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
        
        // Combine using AND logic
        mainCriteria = &imap.SearchCriteria{
            And: [][]*imap.SearchCriteria{{mainCriteria, subCriteria}},
        }
    }
    
    return mainCriteria, options, nil
}

// buildOrCondition creates a criteria with OR logic for multiple conditions
func buildOrCondition(conditions []ComplexSearchConfig, outputConfig *OutputConfig) (*imap.SearchCriteria, *imap.SearchOptions, error) {
    if len(conditions) == 0 {
        return nil, nil, fmt.Errorf("empty conditions list for OR operator")
    }
    
    // Create a slice to hold all subcriteria
    subCriteria := make([]*imap.SearchCriteria, len(conditions))
    var options *imap.SearchOptions
    
    // Process each condition
    for i, condition := range conditions {
        criteria, opts, err := buildSingleCondition(condition, nil)
        if err != nil {
            return nil, nil, err
        }
        
        // Save the first set of options
        if i == 0 {
            options = opts
        }
        
        subCriteria[i] = criteria
    }
    
    // Combine using OR logic
    return &imap.SearchCriteria{
        Or: [][]*imap.SearchCriteria{subCriteria},
    }, options, nil
}

// buildNotCondition creates a criteria with NOT logic for a condition
func buildNotCondition(condition ComplexSearchConfig, outputConfig *OutputConfig) (*imap.SearchCriteria, *imap.SearchOptions, error) {
    subCriteria, options, err := buildSingleCondition(condition, outputConfig)
    if err != nil {
        return nil, nil, err
    }
    
    // Negate the condition
    return &imap.SearchCriteria{
        Not: []*imap.SearchCriteria{subCriteria},
    }, options, nil
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
```

### 3. Add Tests

Create tests for the new functionality:

```go
func TestBuildComplexSearchCriteria(t *testing.T) {
    tests := []struct {
        name           string
        config         SearchConfig
        expectedCriteria *imap.SearchCriteria
        shouldError    bool
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
            expectedCriteria: &imap.SearchCriteria{
                And: [][]*imap.SearchCriteria{
                    {
                        &imap.SearchCriteria{
                            Header: []imap.SearchCriteriaHeaderField{
                                {Key: "From", Value: "person@example.com"},
                            },
                        },
                        &imap.SearchCriteria{
                            Header: []imap.SearchCriteriaHeaderField{
                                {Key: "Subject", Value: "Important"},
                            },
                        },
                    },
                },
            },
            shouldError: false,
        },
        // Add more test cases for OR, NOT, and nested conditions
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            criteria, _, err := BuildSearchCriteria(tt.config, nil)
            
            if tt.shouldError {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tt.expectedCriteria, criteria)
        })
    }
}
```

## Tutorial: Implementing Nested Boolean Operators for IMAP Search

This tutorial guides you through the implementation of nested boolean operators for IMAP search. By the end, you will understand how the go-imap library represents search criteria and how to implement complex nested conditions.

### Understanding the go-imap Library

The go-imap v2 library (github.com/emersion/go-imap/v2) represents search criteria using the `imap.SearchCriteria` struct. This struct contains fields for various search conditions (From, Subject, etc.), as well as three key fields for boolean operations:

- `And`: For AND conditions
- `Or`: For OR conditions
- `Not`: For NOT conditions

These fields allow for constructing complex, nested search expressions.

### Step 1: Understanding the Current Code Structure

First, familiarize yourself with the current implementation:

1. **SearchConfig** in `pkg/dsl/types.go` - Contains fields for basic search criteria
2. **BuildSearchCriteria** in `pkg/dsl/search.go` - Converts SearchConfig to imap.SearchCriteria

### Step 2: Implementing ComplexSearchConfig

Create a new type to represent complex conditions:

```go
// Operator represents a boolean logic operator
type Operator string

const (
    OperatorAnd Operator = "and"
    OperatorOr  Operator = "or"
    OperatorNot Operator = "not"
)

// ComplexSearchConfig defines a search condition that can contain nested conditions
type ComplexSearchConfig struct {
    // Base search criteria fields (same as current SearchConfig)
    SearchConfig `yaml:",inline"`
    
    // Boolean operator to apply to conditions
    Operator Operator `yaml:"operator,omitempty"`
    
    // Nested conditions
    Conditions []ComplexSearchConfig `yaml:"conditions,omitempty"`
}
```

### Step 3: Update SearchConfig

Update the `SearchConfig` struct to include nested conditions:

```go
// SearchConfig defines search criteria
type SearchConfig struct {
    // Existing fields...
    
    // Complex conditions with boolean operators
    Operator   Operator              `yaml:"operator,omitempty"`
    Conditions []ComplexSearchConfig `yaml:"conditions,omitempty"`
}
```

### Step 4: Implement Validation Logic

Add validation methods for complex conditions:

```go
// Validate checks if the complex search config is valid
func (c *ComplexSearchConfig) Validate() error {
    // Implementation as described in the plan
}

// Update the existing SearchConfig.Validate method
func (s *SearchConfig) Validate() error {
    // Existing validation logic
    
    // Add validation for complex conditions
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
```

### Step 5: Update the Search Builder

Modify the `BuildSearchCriteria` function to handle complex conditions:

```go
func BuildSearchCriteria(config SearchConfig, outputConfig *OutputConfig) (*imap.SearchCriteria, *imap.SearchOptions, error) {
    // Check if we have complex conditions
    if config.Operator != "" && len(config.Conditions) > 0 {
        return buildComplexSearchCriteria(config, outputConfig)
    }
    
    // Existing code for flat search criteria
    // ...
}
```

### Step 6: Implement Helper Functions

Implement functions to handle AND, OR, and NOT operations as described in the plan:

```go
func buildComplexSearchCriteria(config SearchConfig, outputConfig *OutputConfig) (*imap.SearchCriteria, *imap.SearchOptions, error) {
    // Implementation as described in the plan
}

func buildAndCondition(conditions []ComplexSearchConfig, outputConfig *OutputConfig) (*imap.SearchCriteria, *imap.SearchOptions, error) {
    // Implementation as described in the plan
}

func buildOrCondition(conditions []ComplexSearchConfig, outputConfig *OutputConfig) (*imap.SearchCriteria, *imap.SearchOptions, error) {
    // Implementation as described in the plan
}

func buildNotCondition(condition ComplexSearchConfig, outputConfig *OutputConfig) (*imap.SearchCriteria, *imap.SearchOptions, error) {
    // Implementation as described in the plan
}

func buildSingleCondition(condition ComplexSearchConfig, outputConfig *OutputConfig) (*imap.SearchCriteria, *imap.SearchOptions, error) {
    // Implementation as described in the plan
}
```

### Step 7: Add Tests

Write tests to verify the correct behavior of complex conditions:

```go
func TestBuildComplexSearchCriteria(t *testing.T) {
    // Test cases as described in the plan
}
```

### Step 8: Integration Testing

Test the functionality with the IMAP server using complex DSL rules:

```yaml
search:
  operator: and
  conditions:
    - from: "team@company.com"
    - subject_contains: "Report"
    - operator: or
      conditions:
        - flags:
            has: ["urgent"]
        - subject_contains: "URGENT"
```

### Understanding How go-imap Represents Boolean Logic

The go-imap library represents boolean operations as follows:

1. **AND**: Combines multiple criteria where all must match
   ```go
   criteria := &imap.SearchCriteria{
       And: [][]*imap.SearchCriteria{
           {criteriaA, criteriaB}, // criteriaA AND criteriaB
       },
   }
   ```

2. **OR**: Combines multiple criteria where at least one must match
   ```go
   criteria := &imap.SearchCriteria{
       Or: [][]*imap.SearchCriteria{
           {criteriaA, criteriaB}, // criteriaA OR criteriaB
       },
   }
   ```

3. **NOT**: Negates a criteria
   ```go
   criteria := &imap.SearchCriteria{
       Not: []*imap.SearchCriteria{criteriaA}, // NOT criteriaA
   }
   ```

### Example: Building a Complex Query

Let's walk through how you'd build a complex query like:
`(from:"alice" AND subject:"report") OR (from:"bob" AND flagged)`

```go
// Individual conditions
fromAlice := &imap.SearchCriteria{
    Header: []imap.SearchCriteriaHeaderField{
        {Key: "From", Value: "alice"},
    },
}

subjectReport := &imap.SearchCriteria{
    Header: []imap.SearchCriteriaHeaderField{
        {Key: "Subject", Value: "report"},
    },
}

fromBob := &imap.SearchCriteria{
    Header: []imap.SearchCriteriaHeaderField{
        {Key: "From", Value: "bob"},
    },
}

flagged := &imap.SearchCriteria{
    Flag: []imap.Flag{"\\Flagged"},
}

// (from:"alice" AND subject:"report")
condition1 := &imap.SearchCriteria{
    And: [][]*imap.SearchCriteria{
        {fromAlice, subjectReport},
    },
}

// (from:"bob" AND flagged)
condition2 := &imap.SearchCriteria{
    And: [][]*imap.SearchCriteria{
        {fromBob, flagged},
    },
}

// condition1 OR condition2
finalCriteria := &imap.SearchCriteria{
    Or: [][]*imap.SearchCriteria{
        {condition1, condition2},
    },
}
```

This structure allows for arbitrarily complex nested conditions to be expressed in the IMAP search criteria.

## Conclusion

This implementation plan provides a complete roadmap for adding support for nested boolean operators in the IMAP DSL. Following this approach will enable users to construct sophisticated search queries with complex logical conditions, greatly enhancing the flexibility and power of the IMAP DSL. 