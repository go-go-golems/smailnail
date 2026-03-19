package rules

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/go-go-golems/smailnail/pkg/dsl"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/accounts"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

var ErrValidation = errors.New("invalid rule input")

const (
	StatusDraft = "draft"
	ModeDryRun  = "dry_run"
)

type AccountResolver interface {
	ResolveConnection(ctx context.Context, userID, accountID string) (*accounts.ConnectionDetails, error)
}

type AccountMailboxOpener interface {
	AccountResolver
	OpenSelectedMailbox(ctx context.Context, userID, accountID, mailbox string) (*imapclient.Client, string, error)
}

type Service struct {
	repo     *Repository
	accounts AccountMailboxOpener
	now      func() time.Time
	newID    func() string
}

type CreateInput struct {
	IMAPAccountID string `json:"imapAccountId"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Status        string `json:"status"`
	SourceKind    string `json:"sourceKind"`
	RuleYAML      string `json:"ruleYaml"`
}

type UpdateInput struct {
	IMAPAccountID *string `json:"imapAccountId"`
	Name          *string `json:"name"`
	Description   *string `json:"description"`
	Status        *string `json:"status"`
	SourceKind    *string `json:"sourceKind"`
	RuleYAML      *string `json:"ruleYaml"`
}

type DryRunInput struct {
	IMAPAccountID string `json:"imapAccountId"`
}

type DryRunResult struct {
	RuleID        string                 `json:"ruleId"`
	IMAPAccountID string                 `json:"imapAccountId"`
	MatchedCount  int                    `json:"matchedCount"`
	ActionPlan    map[string]any         `json:"actionPlan"`
	SampleRows    []accounts.MessageView `json:"sampleRows"`
	CreatedAt     time.Time              `json:"createdAt"`
}

func NewService(repo *Repository, accountResolver AccountMailboxOpener) *Service {
	return &Service{
		repo:     repo,
		accounts: accountResolver,
		now:      func() time.Time { return time.Now().UTC() },
		newID:    uuid.NewString,
	}
}

func (s *Service) List(ctx context.Context, userID string) ([]RuleRecord, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Get(ctx context.Context, userID, ruleID string) (*RuleRecord, error) {
	return s.repo.GetByID(ctx, userID, ruleID)
}

func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (*RuleRecord, error) {
	if strings.TrimSpace(input.IMAPAccountID) == "" {
		return nil, fmt.Errorf("%w: imapAccountId is required", ErrValidation)
	}
	normalized, parsedRule, err := normalizeRuleInput(input.Name, input.Description, input.RuleYAML)
	if err != nil {
		return nil, err
	}

	record := &RuleRecord{
		ID:            s.newID(),
		UserID:        strings.TrimSpace(userID),
		IMAPAccountID: strings.TrimSpace(input.IMAPAccountID),
		Name:          parsedRule.Name,
		Description:   parsedRule.Description,
		Status:        normalizeStatus(input.Status),
		SourceKind:    normalizeSourceKind(input.SourceKind),
		RuleYAML:      normalized,
	}
	if err := s.repo.Create(ctx, record); err != nil {
		return nil, err
	}
	return s.Get(ctx, userID, record.ID)
}

func (s *Service) Update(ctx context.Context, userID, ruleID string, input UpdateInput) (*RuleRecord, error) {
	record, err := s.repo.GetByID(ctx, userID, ruleID)
	if err != nil {
		return nil, err
	}

	if input.IMAPAccountID != nil && strings.TrimSpace(*input.IMAPAccountID) != "" {
		record.IMAPAccountID = strings.TrimSpace(*input.IMAPAccountID)
	}
	if input.Status != nil {
		record.Status = normalizeStatus(*input.Status)
	}
	if input.SourceKind != nil {
		record.SourceKind = normalizeSourceKind(*input.SourceKind)
	}

	ruleYAML := record.RuleYAML
	if input.RuleYAML != nil && strings.TrimSpace(*input.RuleYAML) != "" {
		ruleYAML = *input.RuleYAML
	}

	name := record.Name
	if input.Name != nil {
		name = strings.TrimSpace(*input.Name)
	}
	description := record.Description
	if input.Description != nil {
		description = strings.TrimSpace(*input.Description)
	}

	normalized, parsedRule, err := normalizeRuleInput(name, description, ruleYAML)
	if err != nil {
		return nil, err
	}

	record.Name = parsedRule.Name
	record.Description = parsedRule.Description
	record.RuleYAML = normalized

	if err := s.repo.Update(ctx, record); err != nil {
		return nil, err
	}
	return s.Get(ctx, userID, ruleID)
}

func (s *Service) Delete(ctx context.Context, userID, ruleID string) error {
	return s.repo.Delete(ctx, userID, ruleID)
}

func (s *Service) DryRun(ctx context.Context, userID, ruleID string, input DryRunInput) (*DryRunResult, error) {
	record, err := s.repo.GetByID(ctx, userID, ruleID)
	if err != nil {
		return nil, err
	}

	accountID := strings.TrimSpace(input.IMAPAccountID)
	if accountID == "" {
		accountID = record.IMAPAccountID
	}
	if accountID == "" {
		return nil, fmt.Errorf("%w: imapAccountId is required", ErrValidation)
	}

	rule, err := dsl.ParseRuleString(record.RuleYAML)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidation, err)
	}

	if rule.Output.Format == "" {
		rule.Output.Format = "json"
	}
	switch {
	case rule.Output.Limit <= 0:
		rule.Output.Limit = 20
	case rule.Output.Limit > 25:
		rule.Output.Limit = 25
	}

	connection, err := s.accounts.ResolveConnection(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}

	imapClient, _, err := s.accounts.OpenSelectedMailbox(ctx, userID, accountID, connection.Mailbox)
	if err != nil {
		return nil, err
	}
	defer func() { _ = imapClient.Close() }()

	messages, err := rule.FetchMessages(imapClient)
	if err != nil {
		return nil, fmt.Errorf("%w: dry-run fetch failed: %v", accounts.ErrIMAP, err)
	}

	sampleRows := make([]accounts.MessageView, 0, len(messages))
	for _, message := range messages {
		sampleRows = append(sampleRows, accountMessageToView(message))
	}

	actionPlan := summarizeActions(&rule.Actions)
	createdAt := s.now()
	actionSummaryJSON, err := json.Marshal(actionPlan)
	if err != nil {
		return nil, err
	}
	sampleResultsJSON, err := json.Marshal(sampleRows)
	if err != nil {
		return nil, err
	}

	run := &RuleRun{
		ID:                s.newID(),
		RuleID:            record.ID,
		UserID:            userID,
		IMAPAccountID:     accountID,
		Mode:              ModeDryRun,
		MatchedCount:      len(messages),
		ActionSummaryJSON: string(actionSummaryJSON),
		SampleResultsJSON: string(sampleResultsJSON),
	}
	if err := s.repo.CreateRun(ctx, run); err != nil {
		return nil, err
	}

	record.LastPreviewCount = len(messages)
	record.LastRunAt = &createdAt
	if err := s.repo.Update(ctx, record); err != nil {
		return nil, err
	}

	return &DryRunResult{
		RuleID:        record.ID,
		IMAPAccountID: accountID,
		MatchedCount:  len(messages),
		ActionPlan:    actionPlan,
		SampleRows:    sampleRows,
		CreatedAt:     createdAt,
	}, nil
}

func normalizeRuleInput(name, description, ruleYAML string) (string, *dsl.Rule, error) {
	if strings.TrimSpace(ruleYAML) == "" {
		return "", nil, fmt.Errorf("%w: ruleYaml is required", ErrValidation)
	}

	rule, err := dsl.ParseRuleString(ruleYAML)
	if err != nil {
		return "", nil, fmt.Errorf("%w: %v", ErrValidation, err)
	}

	if strings.TrimSpace(name) != "" {
		rule.Name = strings.TrimSpace(name)
	}
	if strings.TrimSpace(description) != "" {
		rule.Description = strings.TrimSpace(description)
	}

	normalized, err := yaml.Marshal(rule)
	if err != nil {
		return "", nil, err
	}
	return string(normalized), rule, nil
}

func normalizeStatus(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return StatusDraft
	}
	return status
}

func normalizeSourceKind(sourceKind string) string {
	sourceKind = strings.TrimSpace(sourceKind)
	if sourceKind == "" {
		return "ui"
	}
	return sourceKind
}

func summarizeActions(actions *dsl.ActionConfig) map[string]any {
	ret := map[string]any{}
	if actions == nil {
		return ret
	}

	if actions.Flags != nil {
		if len(actions.Flags.Add) > 0 {
			ret["addFlags"] = append([]string{}, actions.Flags.Add...)
		}
		if len(actions.Flags.Remove) > 0 {
			ret["removeFlags"] = append([]string{}, actions.Flags.Remove...)
		}
	}
	if actions.MoveTo != "" {
		ret["moveTo"] = actions.MoveTo
	}
	if actions.CopyTo != "" {
		ret["copyTo"] = actions.CopyTo
	}
	if actions.Delete != nil {
		ret["delete"] = actions.Delete
	}
	if actions.Export != nil {
		ret["export"] = actions.Export
	}
	return ret
}

func accountMessageToView(message *dsl.EmailMessage) accounts.MessageView {
	ret := accounts.MessageView{
		UID:        message.UID,
		SeqNum:     message.SeqNum,
		Flags:      append([]string{}, message.Flags...),
		Size:       message.Size,
		TotalCount: message.TotalCount,
	}

	if message.Envelope != nil {
		ret.Subject = message.Envelope.Subject
		if !message.Envelope.Date.IsZero() {
			ret.Date = message.Envelope.Date.Format(time.RFC3339)
		}
		for _, addr := range message.Envelope.From {
			ret.From = append(ret.From, accounts.AddressView{Name: addr.Name, Address: addr.Address})
		}
		for _, addr := range message.Envelope.To {
			ret.To = append(ret.To, accounts.AddressView{Name: addr.Name, Address: addr.Address})
		}
	}
	for _, part := range message.MimeParts {
		ret.MimeParts = append(ret.MimeParts, accounts.MimePartView{
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
