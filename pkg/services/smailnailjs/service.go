package smailnailjs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/go-go-golems/smailnail/pkg/dsl"
	"github.com/go-go-golems/smailnail/pkg/mailruntime"
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
	AccountID string `json:"accountId"`
	Server    string `json:"server"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Mailbox   string `json:"mailbox"`
	Insecure  bool   `json:"insecure"`
}

type SieveConnectOptions struct {
	AccountID string `json:"accountId"`
	Server    string `json:"server"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type MailboxInfo = mailruntime.MailboxInfo
type MailboxStatus = mailruntime.MailboxStatus
type SearchCriteria = mailruntime.SearchCriteria
type FetchField = mailruntime.FetchField
type FetchedMessage = mailruntime.FetchedMessage
type SieveCapabilities = mailruntime.SieveCapabilities
type ScriptInfo = mailruntime.ScriptInfo

const (
	FetchUID           FetchField = mailruntime.FetchUID
	FetchFlags         FetchField = mailruntime.FetchFlags
	FetchInternalDate  FetchField = mailruntime.FetchInternalDate
	FetchSize          FetchField = mailruntime.FetchSize
	FetchEnvelope      FetchField = mailruntime.FetchEnvelope
	FetchHeaders       FetchField = mailruntime.FetchHeaders
	FetchBodyText      FetchField = mailruntime.FetchBodyText
	FetchBodyHTML      FetchField = mailruntime.FetchBodyHTML
	FetchBodyRaw       FetchField = mailruntime.FetchBodyRaw
	FetchAttachments   FetchField = mailruntime.FetchAttachments
	FetchBodyStructure FetchField = mailruntime.FetchBodyStructure
)

type MailboxSelection struct {
	Name        string `json:"name"`
	ReadOnly    bool   `json:"readOnly"`
	UIDValidity uint32 `json:"uidValidity"`
	UIDNext     uint32 `json:"uidNext"`
	NumMessages uint32 `json:"numMessages"`
}

type Session interface {
	Mailbox() string
	Capabilities() map[string]bool
	List(pattern string) ([]MailboxInfo, error)
	Status(name string) (*MailboxStatus, error)
	SelectMailbox(name string, readOnly bool) (*MailboxSelection, error)
	Search(criteria *SearchCriteria) ([]uint32, error)
	Fetch(uids []uint32, fields []FetchField) ([]*FetchedMessage, error)
	AddFlags(uids []uint32, flags []string, silent bool) error
	RemoveFlags(uids []uint32, flags []string, silent bool) error
	SetFlags(uids []uint32, flags []string, silent bool) error
	Move(uids []uint32, dest string) error
	Copy(uids []uint32, dest string) error
	Delete(uids []uint32, expunge bool) error
	Expunge(uids []uint32) error
	Append(mailbox string, message []byte, flags []string, date *time.Time) (uint32, error)
	Close()
}

type SieveSession interface {
	Capabilities() SieveCapabilities
	ListScripts() ([]ScriptInfo, error)
	GetScript(name string) (string, error)
	PutScript(name, content string, activate bool) error
	Activate(name string) error
	Deactivate() error
	DeleteScript(name string) error
	RenameScript(oldName, newName string) error
	CheckScript(content string) error
	HaveSpace(name string, sizeBytes int) (bool, error)
	Close()
}

type Dialer interface {
	Dial(ctx context.Context, opts ConnectOptions) (Session, error)
}

type SieveDialer interface {
	DialSieve(ctx context.Context, opts SieveConnectOptions) (SieveSession, error)
}

type StoredAccountResolver interface {
	ResolveConnectOptions(ctx context.Context, accountID string) (ConnectOptions, error)
}

type Option func(*Service)

type Service struct {
	dialer                Dialer
	sieveDialer           SieveDialer
	storedAccountResolver StoredAccountResolver
}

func New(opts ...Option) *Service {
	ret := &Service{
		dialer:      RealDialer{},
		sieveDialer: RealSieveDialer{},
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

func WithSieveDialer(dialer SieveDialer) Option {
	return func(s *Service) {
		if dialer != nil {
			s.sieveDialer = dialer
		}
	}
}

func WithStoredAccountResolver(resolver StoredAccountResolver) Option {
	return func(s *Service) {
		if resolver != nil {
			s.storedAccountResolver = resolver
		}
	}
}

func DecodeBuildRuleOptions(input map[string]interface{}) (BuildRuleOptions, error) {
	return decodeJSONMap[BuildRuleOptions](input)
}

func DecodeConnectOptions(input map[string]interface{}) (ConnectOptions, error) {
	return decodeJSONMap[ConnectOptions](input)
}

func DecodeSieveConnectOptions(input map[string]interface{}) (SieveConnectOptions, error) {
	return decodeJSONMap[SieveConnectOptions](input)
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
	if opts.AccountID != "" {
		if s.storedAccountResolver == nil {
			return nil, fmt.Errorf("stored account resolution is not configured")
		}
		resolved, err := s.storedAccountResolver.ResolveConnectOptions(ctx, opts.AccountID)
		if err != nil {
			return nil, err
		}
		if opts.Mailbox != "" {
			resolved.Mailbox = opts.Mailbox
		}
		opts = resolved
	}
	return s.dialer.Dial(ctx, opts)
}

func (s *Service) ConnectSieve(ctx context.Context, opts SieveConnectOptions) (SieveSession, error) {
	if s == nil || s.sieveDialer == nil {
		return nil, fmt.Errorf("service sieve dialer is not configured")
	}
	if opts.AccountID != "" {
		if s.storedAccountResolver == nil {
			return nil, fmt.Errorf("stored account resolution is not configured")
		}
		resolved, err := s.storedAccountResolver.ResolveConnectOptions(ctx, opts.AccountID)
		if err != nil {
			return nil, err
		}
		if opts.Server == "" {
			opts.Server = resolved.Server
		}
		if opts.Username == "" {
			opts.Username = resolved.Username
		}
		if opts.Password == "" {
			opts.Password = resolved.Password
		}
	}
	return s.sieveDialer.DialSieve(ctx, normalizeSieveConnectOptions(opts))
}

type RealDialer struct{}

type RealSieveDialer struct{}

type realSession struct {
	client  *mailruntime.IMAPClient
	mailbox string
}

type realSieveSession struct {
	client *mailruntime.SieveClient
}

func (d RealDialer) Dial(ctx context.Context, opts ConnectOptions) (Session, error) {
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

	client, err := mailruntime.Connect(ctx, mailruntime.IMAPOptions{
		Host:     normalized.Server,
		Port:     normalized.Port,
		TLS:      true,
		Insecure: normalized.Insecure,
		Username: normalized.Username,
		Password: normalized.Password,
	})
	if err != nil {
		return nil, err
	}
	if _, err := client.SelectMailbox(normalized.Mailbox, false); err != nil {
		_ = client.Logout()
		return nil, fmt.Errorf("failed to select mailbox %q: %w", normalized.Mailbox, err)
	}
	return &realSession{
		client:  client,
		mailbox: normalized.Mailbox,
	}, nil
}

func (d RealSieveDialer) DialSieve(_ context.Context, opts SieveConnectOptions) (SieveSession, error) {
	normalized := normalizeSieveConnectOptions(opts)
	if normalized.Server == "" {
		return nil, fmt.Errorf("server is required")
	}
	if normalized.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if normalized.Password == "" {
		return nil, fmt.Errorf("password is required")
	}

	client, err := mailruntime.ConnectSieve(mailruntime.SieveOptions{
		Host:     normalized.Server,
		Port:     normalized.Port,
		Username: normalized.Username,
		Password: normalized.Password,
	})
	if err != nil {
		return nil, err
	}
	return &realSieveSession{client: client}, nil
}

func (s *realSession) Mailbox() string {
	return s.mailbox
}

func (s *realSession) Capabilities() map[string]bool {
	return s.client.Capabilities()
}

func (s *realSession) List(pattern string) ([]MailboxInfo, error) {
	return s.client.List(pattern)
}

func (s *realSession) Status(name string) (*MailboxStatus, error) {
	return s.client.Status(name)
}

func (s *realSession) SelectMailbox(name string, readOnly bool) (*MailboxSelection, error) {
	if s.client.SelectedMailbox() != "" && s.client.SelectedMailbox() != name {
		if err := s.client.UnselectMailbox(); err != nil {
			return nil, err
		}
	}
	data, err := s.client.SelectMailbox(name, readOnly)
	if err != nil {
		return nil, err
	}
	s.mailbox = name
	ret := &MailboxSelection{
		Name:        name,
		ReadOnly:    readOnly,
		UIDValidity: data.UIDValidity,
		NumMessages: data.NumMessages,
	}
	if data.UIDNext != 0 {
		ret.UIDNext = uint32(data.UIDNext)
	}
	return ret, nil
}

func (s *realSession) Search(criteria *SearchCriteria) ([]uint32, error) {
	uids, err := s.client.Search(criteria)
	if err != nil {
		return nil, err
	}
	return imapUIDsToUint32(uids), nil
}

func (s *realSession) Fetch(uids []uint32, fields []FetchField) ([]*FetchedMessage, error) {
	return s.client.Fetch(uint32ToIMAPUIDs(uids), fields)
}

func (s *realSession) AddFlags(uids []uint32, flags []string, silent bool) error {
	return s.client.StoreFlags(uint32ToIMAPUIDs(uids), imap.StoreFlagsAdd, stringsToFlags(flags), silent)
}

func (s *realSession) RemoveFlags(uids []uint32, flags []string, silent bool) error {
	return s.client.StoreFlags(uint32ToIMAPUIDs(uids), imap.StoreFlagsDel, stringsToFlags(flags), silent)
}

func (s *realSession) SetFlags(uids []uint32, flags []string, silent bool) error {
	return s.client.StoreFlags(uint32ToIMAPUIDs(uids), imap.StoreFlagsSet, stringsToFlags(flags), silent)
}

func (s *realSession) Move(uids []uint32, dest string) error {
	return s.client.MoveUIDs(uint32ToIMAPUIDs(uids), dest)
}

func (s *realSession) Copy(uids []uint32, dest string) error {
	return s.client.CopyUIDs(uint32ToIMAPUIDs(uids), dest)
}

func (s *realSession) Delete(uids []uint32, expunge bool) error {
	return s.client.DeleteUIDs(uint32ToIMAPUIDs(uids), expunge)
}

func (s *realSession) Expunge(uids []uint32) error {
	return s.client.Expunge(uint32ToIMAPUIDs(uids))
}

func (s *realSession) Append(mailbox string, message []byte, flags []string, date *time.Time) (uint32, error) {
	if mailbox == "" {
		mailbox = s.mailbox
	}
	uid, err := s.client.Append(mailbox, message, stringsToFlags(flags), date)
	if err != nil {
		return 0, err
	}
	return uint32(uid), nil
}

func (s *realSession) Close() {
	if s != nil && s.client != nil {
		_ = s.client.Logout()
	}
}

func (s *realSieveSession) Capabilities() SieveCapabilities {
	return s.client.Capabilities()
}

func (s *realSieveSession) ListScripts() ([]ScriptInfo, error) {
	return s.client.ListScripts()
}

func (s *realSieveSession) GetScript(name string) (string, error) {
	return s.client.GetScript(name)
}

func (s *realSieveSession) PutScript(name, content string, activate bool) error {
	return s.client.PutScript(name, content, activate)
}

func (s *realSieveSession) Activate(name string) error {
	return s.client.Activate(name)
}

func (s *realSieveSession) Deactivate() error {
	return s.client.Deactivate()
}

func (s *realSieveSession) DeleteScript(name string) error {
	return s.client.DeleteScript(name)
}

func (s *realSieveSession) RenameScript(oldName, newName string) error {
	return s.client.RenameScript(oldName, newName)
}

func (s *realSieveSession) CheckScript(content string) error {
	return s.client.CheckScript(content)
}

func (s *realSieveSession) HaveSpace(name string, sizeBytes int) (bool, error) {
	return s.client.HaveSpace(name, sizeBytes)
}

func (s *realSieveSession) Close() {
	if s != nil && s.client != nil {
		_ = s.client.Logout()
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

func normalizeSieveConnectOptions(opts SieveConnectOptions) SieveConnectOptions {
	if opts.Port == 0 {
		opts.Port = 4190
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

func uint32ToIMAPUIDs(uids []uint32) []imap.UID {
	ret := make([]imap.UID, 0, len(uids))
	for _, uid := range uids {
		ret = append(ret, imap.UID(uid))
	}
	return ret
}

func imapUIDsToUint32(uids []imap.UID) []uint32 {
	ret := make([]uint32, 0, len(uids))
	for _, uid := range uids {
		ret = append(ret, uint32(uid))
	}
	return ret
}

func stringsToFlags(flags []string) []imap.Flag {
	ret := make([]imap.Flag, 0, len(flags))
	for _, flag := range flags {
		ret = append(ret, imap.Flag(flag))
	}
	return ret
}
