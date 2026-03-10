package smailnailjs

import "github.com/go-go-golems/smailnail/pkg/dsl"

type RuleView struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Search      SearchView             `json:"search"`
	Output      OutputView             `json:"output"`
	Actions     map[string]interface{} `json:"actions,omitempty"`
}

type SearchView struct {
	Since           string   `json:"since,omitempty"`
	Before          string   `json:"before,omitempty"`
	On              string   `json:"on,omitempty"`
	WithinDays      int      `json:"withinDays,omitempty"`
	From            string   `json:"from,omitempty"`
	To              string   `json:"to,omitempty"`
	Cc              string   `json:"cc,omitempty"`
	Bcc             string   `json:"bcc,omitempty"`
	Subject         string   `json:"subject,omitempty"`
	SubjectContains string   `json:"subjectContains,omitempty"`
	BodyContains    string   `json:"bodyContains,omitempty"`
	Text            string   `json:"text,omitempty"`
	HasFlags        []string `json:"hasFlags,omitempty"`
	NotHasFlags     []string `json:"notHasFlags,omitempty"`
	LargerThan      string   `json:"largerThan,omitempty"`
	SmallerThan     string   `json:"smallerThan,omitempty"`
}

type OutputView struct {
	Format    string      `json:"format"`
	Limit     int         `json:"limit"`
	Offset    int         `json:"offset"`
	AfterUID  uint32      `json:"afterUid,omitempty"`
	BeforeUID uint32      `json:"beforeUid,omitempty"`
	Fields    []FieldView `json:"fields"`
}

type FieldView struct {
	Name    string       `json:"name"`
	Content *ContentView `json:"content,omitempty"`
}

type ContentView struct {
	Type        string   `json:"type,omitempty"`
	MaxLength   int      `json:"maxLength,omitempty"`
	MinLength   int      `json:"minLength,omitempty"`
	Mode        string   `json:"mode,omitempty"`
	Types       []string `json:"types,omitempty"`
	ShowTypes   bool     `json:"showTypes,omitempty"`
	ShowContent bool     `json:"showContent,omitempty"`
}

type MessageView struct {
	UID        uint32         `json:"uid"`
	SeqNum     uint32         `json:"seqNum"`
	Subject    string         `json:"subject,omitempty"`
	From       []AddressView  `json:"from,omitempty"`
	To         []AddressView  `json:"to,omitempty"`
	Date       string         `json:"date,omitempty"`
	Flags      []string       `json:"flags,omitempty"`
	Size       uint32         `json:"size"`
	MimeParts  []MimePartView `json:"mimeParts,omitempty"`
	TotalCount uint32         `json:"totalCount,omitempty"`
}

type AddressView struct {
	Name    string `json:"name,omitempty"`
	Address string `json:"address"`
}

type MimePartView struct {
	Type     string `json:"type,omitempty"`
	Subtype  string `json:"subtype,omitempty"`
	Size     uint32 `json:"size,omitempty"`
	Content  string `json:"content,omitempty"`
	Filename string `json:"filename,omitempty"`
	Charset  string `json:"charset,omitempty"`
}

func ruleToView(rule *dsl.Rule) RuleView {
	ret := RuleView{
		Name:        rule.Name,
		Description: rule.Description,
		Search: SearchView{
			Since:           rule.Search.Since,
			Before:          rule.Search.Before,
			On:              rule.Search.On,
			WithinDays:      rule.Search.WithinDays,
			From:            rule.Search.From,
			To:              rule.Search.To,
			Cc:              rule.Search.Cc,
			Bcc:             rule.Search.Bcc,
			Subject:         rule.Search.Subject,
			SubjectContains: rule.Search.SubjectContains,
			BodyContains:    rule.Search.BodyContains,
			Text:            rule.Search.Text,
		},
		Output: OutputView{
			Format:    rule.Output.Format,
			Limit:     rule.Output.Limit,
			Offset:    rule.Output.Offset,
			AfterUID:  rule.Output.AfterUID,
			BeforeUID: rule.Output.BeforeUID,
			Fields:    fieldsToViews(rule.Output.Fields),
		},
	}

	if rule.Search.Flags != nil {
		ret.Search.HasFlags = append([]string{}, rule.Search.Flags.Has...)
		ret.Search.NotHasFlags = append([]string{}, rule.Search.Flags.NotHas...)
	}
	if rule.Search.Size != nil {
		ret.Search.LargerThan = rule.Search.Size.LargerThan
		ret.Search.SmallerThan = rule.Search.Size.SmallerThan
	}

	if actions, err := yamlTaggedMap(rule.Actions); err == nil && len(actions) > 0 {
		ret.Actions = actions
	}

	return ret
}

func fieldsToViews(fields []interface{}) []FieldView {
	ret := make([]FieldView, 0, len(fields))
	for _, fieldInterface := range fields {
		field, ok := fieldInterface.(dsl.Field)
		if !ok {
			continue
		}

		view := FieldView{Name: field.Name}
		if field.Content != nil {
			view.Content = &ContentView{
				Type:        field.Content.Type,
				MaxLength:   field.Content.MaxLength,
				MinLength:   field.Content.MinLength,
				Mode:        field.Content.Mode,
				Types:       append([]string{}, field.Content.Types...),
				ShowTypes:   field.Content.ShowTypes,
				ShowContent: field.Content.ShowContent,
			}
		}
		ret = append(ret, view)
	}
	return ret
}

func messageToView(msg *dsl.EmailMessage) MessageView {
	ret := MessageView{
		UID:        msg.UID,
		SeqNum:     msg.SeqNum,
		Flags:      append([]string{}, msg.Flags...),
		Size:       msg.Size,
		TotalCount: msg.TotalCount,
	}

	if msg.Envelope != nil {
		ret.Subject = msg.Envelope.Subject
		if !msg.Envelope.Date.IsZero() {
			ret.Date = msg.Envelope.Date.Format("2006-01-02T15:04:05Z07:00")
		}

		ret.From = make([]AddressView, 0, len(msg.Envelope.From))
		for _, addr := range msg.Envelope.From {
			ret.From = append(ret.From, AddressView{
				Name:    addr.Name,
				Address: addr.Address,
			})
		}

		ret.To = make([]AddressView, 0, len(msg.Envelope.To))
		for _, addr := range msg.Envelope.To {
			ret.To = append(ret.To, AddressView{
				Name:    addr.Name,
				Address: addr.Address,
			})
		}
	}

	ret.MimeParts = make([]MimePartView, 0, len(msg.MimeParts))
	for _, part := range msg.MimeParts {
		ret.MimeParts = append(ret.MimeParts, MimePartView{
			Type:     part.Type,
			Subtype:  part.Subtype,
			Size:     part.Size,
			Content:  part.Content,
			Filename: part.Filename,
			Charset:  part.Charset,
		})
	}

	return ret
}
