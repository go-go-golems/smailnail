package accounts

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/go-go-golems/smailnail/pkg/dsl"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/secrets"
	"github.com/google/uuid"
)

var (
	ErrValidation = errors.New("invalid account input")
	ErrIMAP       = errors.New("imap operation failed")
)

const (
	AuthKindPassword = "password"
	TestModeReadOnly = "read_only"
)

type Service struct {
	repo             *Repository
	secrets          *secrets.Config
	now              func() time.Time
	newID            func() string
	runReadOnlyProbe func(connection *ConnectionDetails) (*readOnlyProbeResult, string, error)
}

type CreateInput struct {
	Label          string `json:"label"`
	ProviderHint   string `json:"providerHint"`
	Server         string `json:"server"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	MailboxDefault string `json:"mailboxDefault"`
	Insecure       bool   `json:"insecure"`
	AuthKind       string `json:"authKind"`
	IsDefault      bool   `json:"isDefault"`
	MCPEnabled     bool   `json:"mcpEnabled"`
}

type UpdateInput struct {
	Label          *string `json:"label"`
	ProviderHint   *string `json:"providerHint"`
	Server         *string `json:"server"`
	Port           *int    `json:"port"`
	Username       *string `json:"username"`
	Password       *string `json:"password"`
	MailboxDefault *string `json:"mailboxDefault"`
	Insecure       *bool   `json:"insecure"`
	AuthKind       *string `json:"authKind"`
	IsDefault      *bool   `json:"isDefault"`
	MCPEnabled     *bool   `json:"mcpEnabled"`
}

type TestInput struct {
	Mode string `json:"mode"`
}

type ListMessagesInput struct {
	Mailbox        string
	Limit          int
	Offset         int
	Query          string
	UnreadOnly     bool
	IncludeContent bool
	ContentType    string
	ContentMaxLen  int
}

type LatestTestSummary struct {
	Success     bool      `json:"success"`
	WarningCode string    `json:"warningCode,omitempty"`
	ErrorCode   string    `json:"errorCode,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

type AccountListItem struct {
	Account
	LatestTest *LatestTestSummary `json:"latestTest,omitempty"`
}

type TestResult struct {
	ID              string         `json:"id"`
	IMAPAccountID   string         `json:"imapAccountId"`
	TestMode        string         `json:"testMode"`
	Success         bool           `json:"success"`
	TCPOK           bool           `json:"tcpOk"`
	LoginOK         bool           `json:"loginOk"`
	MailboxSelectOK bool           `json:"mailboxSelectOk"`
	ListOK          bool           `json:"listOk"`
	SampleFetchOK   bool           `json:"sampleFetchOk"`
	WriteProbeOK    *bool          `json:"writeProbeOk,omitempty"`
	WarningCode     string         `json:"warningCode,omitempty"`
	ErrorCode       string         `json:"errorCode,omitempty"`
	ErrorMessage    string         `json:"errorMessage,omitempty"`
	Details         map[string]any `json:"details,omitempty"`
	CreatedAt       time.Time      `json:"createdAt"`
}

type MailboxInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
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

type ConnectionDetails struct {
	Account  *Account
	Password string
	Mailbox  string
}

type readOnlyProbeResult struct {
	TCPOK           bool
	LoginOK         bool
	MailboxSelectOK bool
	ListOK          bool
	SampleFetchOK   bool
	Details         map[string]any
}

func NewService(repo *Repository, secretConfig *secrets.Config) *Service {
	return &Service{
		repo:             repo,
		secrets:          secretConfig,
		now:              func() time.Time { return time.Now().UTC() },
		newID:            uuid.NewString,
		runReadOnlyProbe: runReadOnlyProbe,
	}
}

func (s *Service) List(ctx context.Context, userID string) ([]AccountListItem, error) {
	accounts, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	ret := make([]AccountListItem, 0, len(accounts))
	for _, account := range accounts {
		item := AccountListItem{Account: account}
		item.SecretCiphertext = ""
		item.SecretNonce = ""

		test, err := s.repo.LatestTestByAccount(ctx, account.ID)
		if err != nil {
			return nil, err
		}
		if test != nil {
			item.LatestTest = &LatestTestSummary{
				Success:     test.Success,
				WarningCode: test.WarningCode,
				ErrorCode:   test.ErrorCode,
				CreatedAt:   test.CreatedAt,
			}
		}

		ret = append(ret, item)
	}
	return ret, nil
}

func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (*Account, error) {
	if err := s.validateCreateInput(input); err != nil {
		return nil, err
	}

	envelope, err := secrets.EncryptString(s.secrets, input.Password)
	if err != nil {
		return nil, err
	}

	account := &Account{
		ID:               s.newID(),
		UserID:           strings.TrimSpace(userID),
		Label:            strings.TrimSpace(input.Label),
		ProviderHint:     strings.TrimSpace(input.ProviderHint),
		Server:           strings.TrimSpace(input.Server),
		Port:             normalizePort(input.Port),
		Username:         strings.TrimSpace(input.Username),
		MailboxDefault:   normalizeMailbox(input.MailboxDefault),
		Insecure:         input.Insecure,
		AuthKind:         normalizeAuthKind(input.AuthKind),
		SecretCiphertext: envelope.Ciphertext,
		SecretNonce:      envelope.Nonce,
		SecretKeyID:      envelope.KeyID,
		IsDefault:        input.IsDefault,
		MCPEnabled:       input.MCPEnabled,
	}

	count, err := s.repo.CountByUser(ctx, account.UserID)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		account.IsDefault = true
	}
	if account.IsDefault {
		if err := s.repo.ClearDefaultForUser(ctx, account.UserID); err != nil {
			return nil, err
		}
	}

	if err := s.repo.Create(ctx, account); err != nil {
		return nil, err
	}
	return s.Get(ctx, account.UserID, account.ID)
}

func (s *Service) Get(ctx context.Context, userID, accountID string) (*Account, error) {
	account, err := s.repo.GetByID(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}
	account.SecretCiphertext = ""
	account.SecretNonce = ""
	return account, nil
}

func (s *Service) Update(ctx context.Context, userID, accountID string, input UpdateInput) (*Account, error) {
	account, err := s.repo.GetByID(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}

	if input.Label != nil {
		account.Label = strings.TrimSpace(*input.Label)
	}
	if input.ProviderHint != nil {
		account.ProviderHint = strings.TrimSpace(*input.ProviderHint)
	}
	if input.Server != nil {
		account.Server = strings.TrimSpace(*input.Server)
	}
	if input.Port != nil {
		account.Port = normalizePort(*input.Port)
	}
	if input.Username != nil {
		account.Username = strings.TrimSpace(*input.Username)
	}
	if input.MailboxDefault != nil {
		account.MailboxDefault = normalizeMailbox(*input.MailboxDefault)
	}
	if input.Insecure != nil {
		account.Insecure = *input.Insecure
	}
	if input.AuthKind != nil {
		account.AuthKind = normalizeAuthKind(*input.AuthKind)
	}
	if input.IsDefault != nil {
		account.IsDefault = *input.IsDefault
	}
	if input.MCPEnabled != nil {
		account.MCPEnabled = *input.MCPEnabled
	}
	if input.Password != nil && strings.TrimSpace(*input.Password) != "" {
		envelope, err := secrets.EncryptString(s.secrets, *input.Password)
		if err != nil {
			return nil, err
		}
		account.SecretCiphertext = envelope.Ciphertext
		account.SecretNonce = envelope.Nonce
		account.SecretKeyID = envelope.KeyID
	}

	if err := s.validateStoredAccount(account); err != nil {
		return nil, err
	}

	if account.IsDefault {
		if err := s.repo.ClearDefaultForUser(ctx, account.UserID); err != nil {
			return nil, err
		}
	}

	if err := s.repo.Update(ctx, account); err != nil {
		return nil, err
	}
	return s.Get(ctx, userID, accountID)
}

func (s *Service) Delete(ctx context.Context, userID, accountID string) error {
	return s.repo.Delete(ctx, userID, accountID)
}

func (s *Service) ResolveConnection(ctx context.Context, userID, accountID string) (*ConnectionDetails, error) {
	account, err := s.repo.GetByID(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}

	password, err := secrets.DecryptString(s.secrets, &secrets.Envelope{
		Ciphertext: account.SecretCiphertext,
		Nonce:      account.SecretNonce,
		KeyID:      account.SecretKeyID,
	})
	if err != nil {
		return nil, err
	}

	return &ConnectionDetails{
		Account:  account,
		Password: password,
		Mailbox:  normalizeMailbox(account.MailboxDefault),
	}, nil
}

func (s *Service) RunTest(ctx context.Context, userID, accountID string, input TestInput) (*TestResult, error) {
	connection, err := s.ResolveConnection(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}

	mode := strings.TrimSpace(input.Mode)
	if mode == "" {
		mode = TestModeReadOnly
	}
	if mode != TestModeReadOnly {
		return nil, fmt.Errorf("%w: unsupported test mode %q", ErrValidation, mode)
	}

	result := &TestResult{
		ID:            s.newID(),
		IMAPAccountID: connection.Account.ID,
		TestMode:      mode,
		Details: map[string]any{
			"mailbox": connection.Mailbox,
			"server":  connection.Account.Server,
			"port":    connection.Account.Port,
		},
		CreatedAt: s.now(),
	}

	probe, stage, stageErr := s.runReadOnlyProbe(connection)
	if shouldRetryReadOnlyProbe(stageErr) {
		probe, stage, stageErr = s.runReadOnlyProbe(connection)
	}
	if probe != nil {
		applyReadOnlyProbeResult(result, probe)
	}
	if stageErr != nil {
		result.ErrorCode, result.WarningCode, result.ErrorMessage = classifyFailure(connection.Account, stage, stageErr)
		if result.WarningCode == "" && connection.Account.Insecure {
			result.WarningCode = "tls-verification-disabled"
		}
		if err := s.persistTestResult(ctx, result); err != nil {
			return nil, err
		}
		return result, nil
	}

	result.Success = result.TCPOK && result.LoginOK && result.MailboxSelectOK && result.ListOK && result.SampleFetchOK
	if connection.Account.Insecure {
		result.WarningCode = "tls-verification-disabled"
	}

	if err := s.persistTestResult(ctx, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) ListMailboxes(ctx context.Context, userID, accountID string) ([]MailboxInfo, error) {
	client, _, err := s.openSelectedMailbox(ctx, userID, accountID, "")
	if err != nil {
		return nil, err
	}
	defer func() { _ = client.Close() }()

	mailboxes, err := client.List("", "*", nil).Collect()
	if err != nil {
		return nil, fmt.Errorf("%w: list mailboxes: %v", ErrIMAP, err)
	}

	ret := make([]MailboxInfo, 0, len(mailboxes))
	for _, mailbox := range mailboxes {
		ret = append(ret, MailboxInfo{
			Name: mailbox.Mailbox,
			Path: mailbox.Mailbox,
		})
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Path < ret[j].Path
	})
	return ret, nil
}

func (s *Service) ListMessages(ctx context.Context, userID, accountID string, input ListMessagesInput) ([]MessageView, string, error) {
	client, mailbox, err := s.openSelectedMailbox(ctx, userID, accountID, input.Mailbox)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = client.Close() }()

	rule := buildPreviewRule(input)
	messages, err := rule.FetchMessages(client)
	if err != nil {
		return nil, "", fmt.Errorf("%w: fetch preview messages: %v", ErrIMAP, err)
	}

	ret := make([]MessageView, 0, len(messages))
	for _, msg := range messages {
		ret = append(ret, messageToView(msg))
	}
	return ret, mailbox, nil
}

func (s *Service) GetMessage(ctx context.Context, userID, accountID string, mailbox string, uid uint32) (*MessageView, string, error) {
	if uid == 0 {
		return nil, "", fmt.Errorf("%w: uid must be greater than zero", ErrValidation)
	}

	client, resolvedMailbox, err := s.openSelectedMailbox(ctx, userID, accountID, mailbox)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = client.Close() }()

	rule := buildDetailRule(uid)
	messages, err := rule.FetchMessages(client)
	if err != nil {
		return nil, "", fmt.Errorf("%w: fetch message detail: %v", ErrIMAP, err)
	}
	if len(messages) == 0 {
		return nil, resolvedMailbox, ErrNotFound
	}

	view := messageToView(messages[0])
	return &view, resolvedMailbox, nil
}

func (s *Service) OpenSelectedMailbox(ctx context.Context, userID, accountID, mailbox string) (*imapclient.Client, string, error) {
	return s.openSelectedMailbox(ctx, userID, accountID, mailbox)
}

func (s *Service) validateCreateInput(input CreateInput) error {
	if strings.TrimSpace(input.Label) == "" {
		return fmt.Errorf("%w: label is required", ErrValidation)
	}
	if strings.TrimSpace(input.Server) == "" {
		return fmt.Errorf("%w: server is required", ErrValidation)
	}
	if strings.TrimSpace(input.Username) == "" {
		return fmt.Errorf("%w: username is required", ErrValidation)
	}
	if strings.TrimSpace(input.Password) == "" {
		return fmt.Errorf("%w: password is required", ErrValidation)
	}
	if normalizePort(input.Port) <= 0 {
		return fmt.Errorf("%w: port must be greater than zero", ErrValidation)
	}
	return nil
}

func (s *Service) validateStoredAccount(account *Account) error {
	if account == nil {
		return fmt.Errorf("%w: account is nil", ErrValidation)
	}
	if strings.TrimSpace(account.Label) == "" {
		return fmt.Errorf("%w: label is required", ErrValidation)
	}
	if strings.TrimSpace(account.Server) == "" {
		return fmt.Errorf("%w: server is required", ErrValidation)
	}
	if strings.TrimSpace(account.Username) == "" {
		return fmt.Errorf("%w: username is required", ErrValidation)
	}
	if account.Port <= 0 {
		return fmt.Errorf("%w: port must be greater than zero", ErrValidation)
	}
	if strings.TrimSpace(account.SecretCiphertext) == "" || strings.TrimSpace(account.SecretNonce) == "" {
		return fmt.Errorf("%w: encrypted password is required", ErrValidation)
	}
	account.MailboxDefault = normalizeMailbox(account.MailboxDefault)
	account.AuthKind = normalizeAuthKind(account.AuthKind)
	return nil
}

func (s *Service) persistTestResult(ctx context.Context, result *TestResult) error {
	detailsJSON, err := json.Marshal(result.Details)
	if err != nil {
		return err
	}

	record := &AccountTest{
		ID:              result.ID,
		IMAPAccountID:   result.IMAPAccountID,
		TestMode:        result.TestMode,
		Success:         result.Success,
		TCPOK:           result.TCPOK,
		LoginOK:         result.LoginOK,
		MailboxSelectOK: result.MailboxSelectOK,
		ListOK:          result.ListOK,
		SampleFetchOK:   result.SampleFetchOK,
		WriteProbeOK:    result.WriteProbeOK,
		WarningCode:     result.WarningCode,
		ErrorCode:       result.ErrorCode,
		ErrorMessage:    result.ErrorMessage,
		DetailsJSON:     string(detailsJSON),
	}
	return s.repo.CreateTest(ctx, record)
}

func (s *Service) openSelectedMailbox(ctx context.Context, userID, accountID, mailbox string) (*imapclient.Client, string, error) {
	connection, err := s.ResolveConnection(ctx, userID, accountID)
	if err != nil {
		return nil, "", err
	}

	client, stage, stageErr := dialAndLogin(connection)
	if stageErr != nil {
		return nil, "", fmt.Errorf("%w: %s: %v", ErrIMAP, stage, stageErr)
	}

	resolvedMailbox := normalizeMailbox(mailbox)
	if resolvedMailbox == "" {
		resolvedMailbox = connection.Mailbox
	}
	if _, err := client.Select(resolvedMailbox, nil).Wait(); err != nil {
		_ = client.Close()
		return nil, "", fmt.Errorf("%w: select mailbox %q: %v", ErrIMAP, resolvedMailbox, err)
	}

	return client, resolvedMailbox, nil
}

func dialAndLogin(connection *ConnectionDetails) (*imapclient.Client, string, error) {
	options := &imapclient.Options{
		TLSConfig: &tls.Config{
			// #nosec G402 -- explicit user-controlled flag for local/self-signed IMAP targets.
			InsecureSkipVerify: connection.Account.Insecure,
		},
	}

	serverAddr := fmt.Sprintf("%s:%d", connection.Account.Server, connection.Account.Port)
	client, err := imapclient.DialTLS(serverAddr, options)
	if err != nil {
		return nil, "connect", err
	}
	if err := client.Login(connection.Account.Username, connection.Password).Wait(); err != nil {
		_ = client.Close()
		return nil, "login", err
	}
	return client, "", nil
}

func runReadOnlyProbe(connection *ConnectionDetails) (*readOnlyProbeResult, string, error) {
	result := &readOnlyProbeResult{
		Details: map[string]any{
			"mailbox": connection.Mailbox,
			"server":  connection.Account.Server,
			"port":    connection.Account.Port,
		},
	}

	client, stage, stageErr := dialAndLogin(connection)
	if stageErr != nil {
		return result, stage, stageErr
	}
	defer func() { _ = client.Close() }()

	result.TCPOK = true
	result.LoginOK = true

	selected, err := client.Select(connection.Mailbox, nil).Wait()
	if err != nil {
		return result, "select", err
	}
	result.MailboxSelectOK = true
	result.Details["selectedMailbox"] = connection.Mailbox
	result.Details["numMessages"] = selected.NumMessages

	mailboxes, err := client.List("", "*", nil).Collect()
	if err != nil {
		return result, "list", err
	}
	result.ListOK = true

	sampleMailboxes := make([]string, 0, len(mailboxes))
	for _, mailbox := range mailboxes {
		sampleMailboxes = append(sampleMailboxes, mailbox.Mailbox)
	}
	sort.Strings(sampleMailboxes)
	result.Details["sampleMailboxes"] = sampleMailboxes

	result.SampleFetchOK = true
	if selected.NumMessages > 0 {
		seqSet := imap.SeqSetNum(selected.NumMessages)
		messages, err := client.Fetch(seqSet, &imap.FetchOptions{
			UID:        true,
			Envelope:   true,
			Flags:      true,
			RFC822Size: true,
		}).Collect()
		if err != nil {
			result.SampleFetchOK = false
			return result, "fetch", err
		}
		if len(messages) > 0 && messages[0].Envelope != nil {
			result.Details["sampleSubject"] = messages[0].Envelope.Subject
		}
	}

	return result, "", nil
}

func shouldRetryReadOnlyProbe(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, net.ErrClosed) {
		return true
	}
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	for _, fragment := range []string{
		"use of closed network connection",
		"broken pipe",
		"connection reset by peer",
		"unexpected eof",
	} {
		if strings.Contains(message, fragment) {
			return true
		}
	}
	return false
}

func applyReadOnlyProbeResult(target *TestResult, probe *readOnlyProbeResult) {
	if target == nil || probe == nil {
		return
	}
	target.TCPOK = probe.TCPOK
	target.LoginOK = probe.LoginOK
	target.MailboxSelectOK = probe.MailboxSelectOK
	target.ListOK = probe.ListOK
	target.SampleFetchOK = probe.SampleFetchOK
	if len(probe.Details) == 0 {
		return
	}
	if target.Details == nil {
		target.Details = map[string]any{}
	}
	for key, value := range probe.Details {
		target.Details[key] = value
	}
}

func classifyFailure(account *Account, stage string, err error) (string, string, string) {
	errorCode := "account-test-failed"
	switch stage {
	case "connect":
		errorCode = "account-test-connect-failed"
	case "login":
		errorCode = "account-test-login-failed"
	case "select":
		errorCode = "account-test-mailbox-select-failed"
	case "list":
		errorCode = "account-test-mailbox-list-failed"
	case "fetch":
		errorCode = "account-test-sample-fetch-failed"
	}

	warningCode := ""
	if stage == "login" {
		switch strings.ToLower(strings.TrimSpace(account.ProviderHint)) {
		case "gmail", "google", "icloud", "yahoo":
			warningCode = "provider-app-password-recommended"
		}
	}
	if warningCode == "" && account.Insecure {
		warningCode = "tls-verification-disabled"
	}

	return errorCode, warningCode, err.Error()
}

func buildPreviewRule(input ListMessagesInput) *dsl.Rule {
	limit := input.Limit
	switch {
	case limit <= 0:
		limit = 20
	case limit > 100:
		limit = 100
	}

	search := dsl.SearchConfig{}
	if query := strings.TrimSpace(input.Query); query != "" {
		search.SubjectContains = query
	}
	if input.UnreadOnly {
		search.Flags = &dsl.FlagCriteria{
			NotHas: []string{"seen"},
		}
	}

	return &dsl.Rule{
		Name:        "hosted-preview",
		Description: "Hosted preview query",
		Search:      search,
		Output: dsl.OutputConfig{
			Format: "json",
			Limit:  limit,
			Offset: max(input.Offset, 0),
			Fields: []interface{}{
				dsl.Field{Name: "uid"},
				dsl.Field{Name: "subject"},
				dsl.Field{Name: "from"},
				dsl.Field{Name: "to"},
				dsl.Field{Name: "date"},
				dsl.Field{Name: "flags"},
				dsl.Field{Name: "size"},
			},
		},
	}
}

func buildDetailRule(uid uint32) *dsl.Rule {
	contentType := "text/*"
	return &dsl.Rule{
		Name:        "hosted-message-detail",
		Description: "Hosted message detail preview",
		Search:      dsl.SearchConfig{},
		Output: dsl.OutputConfig{
			Format: "json",
			Limit:  1,
			AfterUID: func() uint32 {
				if uid > 1 {
					return uid - 1
				}
				return 0
			}(),
			BeforeUID: uid + 1,
			Fields: []interface{}{
				dsl.Field{Name: "uid"},
				dsl.Field{Name: "subject"},
				dsl.Field{Name: "from"},
				dsl.Field{Name: "to"},
				dsl.Field{Name: "date"},
				dsl.Field{Name: "flags"},
				dsl.Field{Name: "size"},
				dsl.Field{
					Name: "mime_parts",
					Content: &dsl.ContentField{
						ShowContent: true,
						MaxLength:   4096,
						Mode:        "filter",
						Types:       []string{contentType},
					},
				},
			},
		},
	}
}

func normalizeMailbox(mailbox string) string {
	if strings.TrimSpace(mailbox) == "" {
		return "INBOX"
	}
	return strings.TrimSpace(mailbox)
}

func normalizePort(port int) int {
	if port <= 0 {
		return 993
	}
	return port
}

func normalizeAuthKind(authKind string) string {
	if strings.TrimSpace(authKind) == "" {
		return AuthKindPassword
	}
	return strings.TrimSpace(authKind)
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
			ret.Date = msg.Envelope.Date.Format(time.RFC3339)
		}

		for _, addr := range msg.Envelope.From {
			ret.From = append(ret.From, AddressView{Name: addr.Name, Address: addr.Address})
		}
		for _, addr := range msg.Envelope.To {
			ret.To = append(ret.To, AddressView{Name: addr.Name, Address: addr.Address})
		}
	}

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
