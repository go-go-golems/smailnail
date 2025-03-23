# Mailgen

A command-line tool for generating test emails from YAML templates using the Go template engine and Sprig functions.

## Installation

```bash
go install github.com/wesen/corporate-headquarters/smailnail/cmd/mailgen@latest
```

## Usage

```bash
mailgen generate --config example.yaml [--output-dir ./output] [--write-files]
```

### Options

- `--config`: Path to YAML config file (required)
- `--output-dir`: Directory to output generated emails (default: ./output)
- `--write-files`: Write emails to files (default: false)
- `--format`: Output format (default: table, options: table, json, yaml, csv)

## YAML Configuration

The YAML configuration consists of four main sections:

1. `variables`: Global variable definitions
2. `templates`: Email template definitions
3. `rules`: Generation rule definitions
4. `generate`: Generation execution configuration

### Example

```yaml
variables:
  senders:
    - name: "John Doe"
      email: "john@example.com"
  subjects:
    - "Hello there!"

templates:
  basic:
    subject: "{{ .subject }}"
    from: "{{ .sender.name }} <{{ .sender.email }}>"
    to: "{{ .recipient }}"
    body: |
      Hello {{ .recipient_name }},
      
      {{ .content }}
      
      Best regards,
      {{ .sender.name }}

rules:
  welcome:
    template: basic
    variations:
      - sender: "{{ index .variables.senders 0 }}"
        subject: "{{ index .variables.subjects 0 }}"
        recipient: "user@example.com"
        recipient_name: "User"
        content: "Welcome to our service!"

generate:
  - rule: welcome
    count: 5
```

## Template Context

During template processing, the following context is available:

- `.variables`: All defined variables
- `.index`: Current generation index (in generate section)
- `.template`: Current template being processed
- `.rule`: Current rule being processed

Additionally, all properties defined in the current variation are available at the root level.
```

```
 _______  _______    _______  _______ 
|       ||       |  |       ||       |
|    ___||   _   |  |    ___||   _   |
|   | __ |  | |  |  |   | __ |  | |  |
|   ||  ||  |_|  |  |   ||  ||  |_|  |
|   |_| ||       |  |   |_| ||       |
|_______||_______|  |_______||_______|
 _______  _______  __   __  _______  ___      _______  _______  _______ 
|       ||       ||  |_|  ||       ||   |    |   _   ||       ||       |
|_     _||    ___||       ||    _  ||   |    |  |_|  ||_     _||    ___|
  |   |  |   |___ |       ||   |_| ||   |    |       |  |   |  |   |___ 
  |   |  |    ___||       ||    ___||   |___ |       |  |   |  |    ___|
  |   |  |   |___ | ||_|| ||   |    |       ||   _   |  |   |  |   |___ 
  |___|  |_______||_|   |_||___|    |_______||__| |__|  |___|  |_______|
```

---

```
 _______  _______  ___      _______  __   __  _______ 
|       ||       ||   |    |       ||  |_|  ||       |
|    ___||   _   ||   |    |    ___||       ||  _____|
|   | __ |  | |  ||   |    |   |___ |       || |_____ 
|   ||  ||  |_|  ||   |___ |    ___||       ||_____  |
|   |_| ||       ||       ||   |___ | ||_|| | _____| |
|_______||_______||_______||_______||_|   |_||_______|
 __   __  _______  ___   _  _______    __   __  _______  ______   _______ 
|  |_|  ||   _   ||   | | ||       |  |  |_|  ||       ||    _ | |       |
|       ||  |_|  ||   |_| ||    ___|  |       ||   _   ||   | || |    ___|
|       ||       ||      _||   |___   |       ||  | |  ||   |_|| |   |___ 
|       ||       ||     |_ |    ___|  |       ||  |_|  ||    __ ||    ___|
| ||_|| ||   _   ||    _  ||   |___   | ||_|| ||       ||   |  |||   |___ 
|_|   |_||__| |__||___| |_||_______|  |_|   |_||_______||___|  |||_______|
 _______  _______    _______  _______ 
|       ||       |  |       ||       |
|    ___||   _   |  |    ___||   _   |
|   | __ |  | |  |  |   | __ |  | |  |
|   ||  ||  |_|  |  |   ||  ||  |_|  |
|   |_| ||       |  |   |_| ||       |
|_______||_______|  |_______||_______|
 _______  _______  ___      _______  __   __  _______ 
|       ||       ||   |    |       ||  |_|  ||       |
|    ___||   _   ||   |    |    ___||       ||  _____|
|   | __ |  | |  ||   |    |   |___ |       || |_____ 
|   ||  ||  |_|  ||   |___ |    ___||       ||_____  |
|   |_| ||       ||       ||   |___ | ||_|| | _____| |
|_______||_______||_______||_______||_|   |_||_______|
```
