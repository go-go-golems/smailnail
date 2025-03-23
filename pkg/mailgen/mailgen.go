package mailgen

import (
	"bytes"
	"context"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/go-go-golems/smailnail/pkg/types"
	"github.com/pkg/errors"
)

// MailGenerator is the main email generator
type MailGenerator struct {
	config *types.TemplateConfig
	funcs  template.FuncMap
}

// NewMailGenerator creates a new MailGenerator
func NewMailGenerator(config *types.TemplateConfig) *MailGenerator {
	funcs := sprig.TxtFuncMap()

	// Add any additional custom functions here if needed
	// funcs["customFunc"] = customFunc

	return &MailGenerator{
		config: config,
		funcs:  funcs,
	}
}

// Generate generates emails based on the configuration
func (g *MailGenerator) Generate(ctx context.Context) ([]*types.Email, error) {
	if err := g.config.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid configuration")
	}

	var emails []*types.Email

	// Process each generate entry
	for _, genConfig := range g.config.Generate {
		rule := g.config.Rules[genConfig.Rule]
		tmpl := g.config.Templates[rule.Template]

		// Generate specified number of emails for this rule
		for i := 0; i < genConfig.Count; i++ {
			// Choose a variation for this email
			variationIndex := i % len(rule.Variations)
			variation := rule.Variations[variationIndex]

			// First process all variation values as templates
			processedVariation := make(map[string]string)
			for key, value := range variation {
				// Create context for variation template processing
				varCtx := map[string]interface{}{
					"variables": g.config.Variables,
					"index":     i,
					"template":  tmpl,
					"rule":      rule,
				}

				// Process the variation value as a template
				processed, err := g.processTemplate(key, value, varCtx)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to process variation value for key '%s'", key)
				}
				processedVariation[key] = processed
			}

			// Create context for email template processing
			ctx := map[string]interface{}{
				"variables": g.config.Variables,
				"index":     i,
				"template":  tmpl,
				"rule":      rule,
			}

			// Add all processed variation values to context root
			for k, v := range processedVariation {
				ctx[k] = v
			}

			// Process the email template
			email, err := g.processEmailTemplate(tmpl, ctx)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to process template for rule '%s', index %d", genConfig.Rule, i)
			}

			emails = append(emails, email)
		}
	}

	return emails, nil
}

// processEmailTemplate processes a single email template with the given context
func (g *MailGenerator) processEmailTemplate(emailTemplate types.EmailTemplate, ctx map[string]interface{}) (*types.Email, error) {
	// Create a new email
	email := &types.Email{}

	// Process subject
	subject, err := g.processTemplate("subject", emailTemplate.Subject, ctx)
	if err != nil {
		return nil, err
	}
	email.Subject = subject

	// Process from
	from, err := g.processTemplate("from", emailTemplate.From, ctx)
	if err != nil {
		return nil, err
	}
	email.From = from

	// Process optional fields if present
	if emailTemplate.To != "" {
		to, err := g.processTemplate("to", emailTemplate.To, ctx)
		if err != nil {
			return nil, err
		}
		email.To = to
	}

	if emailTemplate.Cc != "" {
		cc, err := g.processTemplate("cc", emailTemplate.Cc, ctx)
		if err != nil {
			return nil, err
		}
		email.Cc = cc
	}

	if emailTemplate.Bcc != "" {
		bcc, err := g.processTemplate("bcc", emailTemplate.Bcc, ctx)
		if err != nil {
			return nil, err
		}
		email.Bcc = bcc
	}

	if emailTemplate.ReplyTo != "" {
		replyTo, err := g.processTemplate("reply_to", emailTemplate.ReplyTo, ctx)
		if err != nil {
			return nil, err
		}
		email.ReplyTo = replyTo
	}

	// Process body
	body, err := g.processTemplate("body", emailTemplate.Body, ctx)
	if err != nil {
		return nil, err
	}
	email.Body = body

	return email, nil
}

// processTemplate processes a template string with the given context
func (g *MailGenerator) processTemplate(name, tmpl string, ctx map[string]interface{}) (string, error) {
	// Parse the template
	t, err := template.New(name).Funcs(g.funcs).Parse(tmpl)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse template '%s'", name)
	}

	// Execute the template
	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		return "", errors.Wrapf(err, "failed to execute template '%s'", name)
	}

	return buf.String(), nil
}
