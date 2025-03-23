# Email Generation DSL Specification

## Overview
This document defines a YAML-based Domain Specific Language (DSL) for generating email content using Go templates and Sprig functions. The DSL follows a rule-based approach, allowing a single template to generate multiple email variations through defined rules and variables.

## Core Structure

The DSL consists of four primary sections, each with a specific purpose and schema:

```yaml
variables:    # Global variable definitions (strings only)
templates:    # Email template definitions
rules:       # Generation rule definitions
generate:    # Generation execution configuration
```

## Variables Section

### Purpose
Defines reusable string values accessible throughout all templates and rules.

### Schema
```yaml
variables:
  # Simple key-value pairs (all values must be strings)
  key: "string value"
  
  # Lists of strings
  list_name:
    - "string1"
    - "string2"
  
  # Categorized strings
  category_name:
    subcategory: 
      - "string1"
      - "string2"
```

### Example
```yaml
variables:
  sender_name: "John Doe"
  sender_email: "john@example.com"
  subjects:
    - "Hello there!"
    - "Important update"
  greetings:
    formal:
      - "Dear"
      - "Hello"
    informal:
      - "Hi"
      - "Hey"
```

## Templates Section

### Purpose
Defines the structure of different email types with placeholders for dynamic content.

### Schema
```yaml
templates:
  template_id:                       # Unique template identifier
    subject: "template_string"       # Email subject line template
    from: "template_string"          # Sender information template
    to: "template_string"            # Optional recipient template
    cc: "template_string"            # Optional CC template
    bcc: "template_string"           # Optional BCC template
    reply_to: "template_string"      # Optional reply-to template
    body: |                          # Email body template (multi-line)
      template_content
```

### Example
```yaml
templates:
  personal_email:
    subject: "{{ .subject }}"
    from: "{{ .sender_name }} <{{ .sender_email }}>"
    body: |
      {{ .greeting }} {{ .recipient_name }},
      
      {{ .content }}
      
      {{ .closing }},
      {{ .sender_name }}
```

## Rules Section

### Purpose
Defines how to generate variations of emails by applying string values to templates.

### Schema
```yaml
rules:
  rule_id:                           # Unique rule identifier
    template: template_id            # References a template defined above
    variations:                      # Array of variation definitions
      - property1: "string_value"    # Properties must be strings or template strings
        property2: "string_value"    # Each property is processed as a template
```

### Example
```yaml
rules:
  personal_emails:
    template: personal_email      
    variations:                   
      - sender_name: "{{ .variables.sender_name }}"
        sender_email: "{{ .variables.sender_email }}"
        recipient_name: "User"
        greeting: "{{ pickRandom .variables.greetings.formal }}"
        subject: "{{ pickRandom .variables.subjects }}"
        content: "Welcome to our service!"
        closing: "Best regards"
```

## Generate Section

### Purpose
Configures which rules to execute and how many variations to generate.

### Schema
```yaml
generate:
  - rule: rule_id                # Which rule to use
    count: number                # How many variations to generate
    output: "path_template"      # Optional: output file path template
```

### Example
```yaml
generate:
  - rule: personal_emails    
    count: 5                
    output: "emails/personal_{{ now | date \"2006-01-02\" }}/email_{{ .index }}.txt"
```

## Template Context

During template processing, the following context is available:

### Root Context Properties
- `.variables` - All defined variables (all strings)
- `.index` - Current generation index (in generate section)
- `.template` - Current template being processed
- `.rule` - Current rule being processed

### Variation Properties
All properties defined in the current variation are available at the root level:
- `.property_name` - Direct access to any property defined in the variation (all strings)

## Common Sprig Functions

- `pickRandom` - Randomly select from a list - `{{ pickRandom .variables.some_list }}`
- `randInt` - Generate random integer - `{{ randInt 1 100 }}`
- `lower`/`upper` - Case conversion - `{{ lower .some_string }}`
- `title` - Title case - `{{ title .some_string }}`
- `list` - Create a list - `{{ list "item1" "item2" }}`
- `now` - Current time - `{{ now }}`
- `date` - Format date - `{{ now | date "2006-01-02" }}`
- `replace` - Replace string - `{{ replace "old" "new" .string }}`
- `contains` - Check if string contains substring - `{{ contains "needle" .haystack }}`
- `default` - Default value - `{{ default "default" .variable }}`
- `quote` - Quote a string - `{{ quote .string }}`