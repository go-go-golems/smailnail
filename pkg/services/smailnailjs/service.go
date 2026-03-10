package smailnailjs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/go-go-golems/smailnail/pkg/dsl"
	imaplayer "github.com/go-go-golems/smailnail/pkg/imap"
	"gopkg.in/yaml.v3"
)

type BuildRuleOptions struct {
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Since            string   `json:"since"`
	Before           string   `json:"before"`
	WithinDays       int      `json:"withinDays"`
	From             string   `json:"from"`
	To               string   `json:"to"`
	Subject          string   `json:"subject"`
	SubjectContains  string   `json:"subjectContains"`
	BodyContains     string   `json:"bodyContains"`
	HasFlags         []string `json:"hasFlags"`
	NotHasFlags      []string `json:"notHasFlags"`
	LargerThan       string   `json:"largerThan"`
	SmallerThan      string   `json:"smallerThan"`
	Limit            int      `json:"limit"`
	Offset           int      `json:"offset"`
	AfterUID         uint32   `json:"afterUid"`
	BeforeUID        uint32   `json:"beforeUid"`
	Format           string   `json:"format"`
	IncludeContent   bool     `json:"includeContent"`
	ContentType      string   `json:"contentType"`
	ContentMaxLength int      `json:"contentMaxLength"`
}

type ConnectOptions struct {
	Server   string `json:"server"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Mailbox  string `json:"mailbox"`
	Insecure bool   `json:"insecure"`
}

type Session interface {
	Mailbox() string
	Close()
}

type Dialer interface {
	Dial(ctx context.Context, opts ConnectOptions) (Session, error)
}

type Option func(*Service)

type Service struct {
	dialer Dialer
}

func New(opts ...Option) *Service {
	ret := &Service{
		dialer: RealDialer{},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(ret)
		}
	}
	return ret
}

func WithDialer(dialer Dialer) Option {
	return func(s *Service) {
		if dialer != nil {
			s.dialer = dialer
		}
	}
}

func DecodeBuildRuleOptions(input map[string]interface{}) (BuildRuleOptions, error) {
	return decodeJSONMap[BuildRuleOptions](input)
}

func DecodeConnectOptions(input map[string]interface{}) (ConnectOptions, error) {
	return decodeJSONMap[ConnectOptions](input)
}

func (s *Service) ParseRuleString(yamlString string) (*dsl.Rule, error) {
	return dsl.ParseRuleString(yamlString)
}

func (s *Service) ParseRuleMap(yamlString string) (map[string]interface{}, error) {
	rule, err := s.ParseRuleString(yamlString)
	if err != nil {
		return nil, err
	}
	return RuleToMap(rule)
}

func BuildDSLRule(opts BuildRuleOptions) (*dsl.Rule, error) {
	searchConfig := dsl.SearchConfig{
		Since:           opts.Since,
		Before:          opts.Before,
		WithinDays:      opts.WithinDays,
		From:            opts.From,
		To:              opts.To,
		Subject:         opts.Subject,
		SubjectContains: opts.SubjectContains,
		BodyContains:    opts.BodyContains,
	}

	if len(opts.HasFlags) > 0 || len(opts.NotHasFlags) > 0 {
		searchConfig.Flags = &dsl.FlagCriteria{
			Has:    opts.HasFlags,
			NotHas: opts.NotHasFlags,
		}
	}

	if opts.LargerThan != "" || opts.SmallerThan != "" {
		searchConfig.Size = &dsl.SizeCriteria{
			LargerThan:  opts.LargerThan,
			SmallerThan: opts.SmallerThan,
		}
	}

	fields := []interface{}{
		dsl.Field{Name: "uid"},
		dsl.Field{Name: "subject"},
		dsl.Field{Name: "from"},
		dsl.Field{Name: "to"},
		dsl.Field{Name: "date"},
		dsl.Field{Name: "flags"},
		dsl.Field{Name: "size"},
	}

	if opts.IncludeContent {
		contentField := &dsl.ContentField{
			ShowContent: true,
			MaxLength:   opts.ContentMaxLength,
		}
		if opts.ContentType != "" {
			contentField.Mode = "filter"
			contentField.Types = []string{opts.ContentType}
		}
		fields = append(fields, dsl.Field{
			Name:    "mime_parts",
			Content: contentField,
		})
	}

	name := opts.Name
	if name == "" {
		name = "js-rule"
	}

	description := opts.Description
	if description == "" {
		description = "Rule generated from JavaScript API"
	}

	format := opts.Format
	if format == "" {
		format = "text"
	}

	rule := &dsl.Rule{
		Name:        name,
		Description: description,
		Search:      searchConfig,
		Output: dsl.OutputConfig{
			Format:    format,
			Limit:     opts.Limit,
			Offset:    opts.Offset,
			AfterUID:  opts.AfterUID,
			BeforeUID: opts.BeforeUID,
			Fields:    fields,
		},
	}

	if err := rule.Validate(); err != nil {
		return nil, fmt.Errorf("invalid rule: %w", err)
	}

	return rule, nil
}

func (s *Service) BuildRuleMap(opts BuildRuleOptions) (map[string]interface{}, error) {
	rule, err := BuildDSLRule(opts)
	if err != nil {
		return nil, err
	}
	return RuleToMap(rule)
}

func (s *Service) ShapeMessageMap(msg *dsl.EmailMessage) (map[string]interface{}, error) {
	return toJSONMap(messageToView(msg))
}

func (s *Service) Connect(ctx context.Context, opts ConnectOptions) (Session, error) {
	if s == nil || s.dialer == nil {
		return nil, fmt.Errorf("service dialer is not configured")
	}
	return s.dialer.Dial(ctx, opts)
}

type RealDialer struct{}

type realSession struct {
	client  *imapclient.Client
	mailbox string
}

func (d RealDialer) Dial(_ context.Context, opts ConnectOptions) (Session, error) {
	normalized := normalizeConnectOptions(opts)
	if normalized.Server == "" {
		return nil, fmt.Errorf("server is required")
	}
	if normalized.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if normalized.Password == "" {
		return nil, fmt.Errorf("password is required")
	}

	settings := imaplayer.IMAPSettings{
		Server:   normalized.Server,
		Port:     normalized.Port,
		Username: normalized.Username,
		Password: normalized.Password,
		Mailbox:  normalized.Mailbox,
		Insecure: normalized.Insecure,
	}
	client, err := settings.ConnectToIMAPServer()
	if err != nil {
		return nil, err
	}
	if _, err := client.Select(settings.Mailbox, nil).Wait(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to select mailbox %q: %w", settings.Mailbox, err)
	}
	return &realSession{
		client:  client,
		mailbox: settings.Mailbox,
	}, nil
}

func (s *realSession) Mailbox() string {
	return s.mailbox
}

func (s *realSession) Close() {
	if s != nil && s.client != nil {
		_ = s.client.Close()
	}
}

func normalizeConnectOptions(opts ConnectOptions) ConnectOptions {
	if opts.Port == 0 {
		opts.Port = 993
	}
	if opts.Mailbox == "" {
		opts.Mailbox = "INBOX"
	}
	return opts
}

func RuleToMap(rule *dsl.Rule) (map[string]interface{}, error) {
	view := ruleToView(rule)
	return toJSONMap(view)
}

func decodeJSONMap[T any](input map[string]interface{}) (T, error) {
	var ret T
	raw, err := json.Marshal(input)
	if err != nil {
		return ret, fmt.Errorf("marshal input map: %w", err)
	}
	if err := json.Unmarshal(raw, &ret); err != nil {
		return ret, fmt.Errorf("decode input map: %w", err)
	}
	return ret, nil
}

func toJSONMap(v interface{}) (map[string]interface{}, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal value: %w", err)
	}

	var ret map[string]interface{}
	if err := json.Unmarshal(raw, &ret); err != nil {
		return nil, fmt.Errorf("decode value map: %w", err)
	}
	return ret, nil
}

func yamlTaggedMap(v interface{}) (map[string]interface{}, error) {
	raw, err := yaml.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal yaml value: %w", err)
	}
	var ret map[string]interface{}
	if err := yaml.Unmarshal(raw, &ret); err != nil {
		return nil, fmt.Errorf("decode yaml value: %w", err)
	}
	return ret, nil
}
