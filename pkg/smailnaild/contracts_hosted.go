package smailnaild

import (
	"time"

	appv1 "github.com/go-go-golems/smailnail/pkg/gen/smailnail/app/v1"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/accounts"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/identity"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/rules"
	"google.golang.org/protobuf/types/known/structpb"
)

func formatHostedProtoTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339Nano)
}

func databaseInfoToProto(info DatabaseInfo) *appv1.DatabaseInfo {
	return &appv1.DatabaseInfo{
		Driver: info.Driver,
		Target: info.Target,
		Mode:   info.Mode,
	}
}

func currentUserToProto(user *identity.User) *appv1.CurrentUser {
	if user == nil {
		return nil
	}
	primaryEmail := user.PrimaryEmail
	displayName := user.DisplayName
	avatarURL := user.AvatarURL
	createdAt := formatHostedProtoTime(user.CreatedAt)
	updatedAt := formatHostedProtoTime(user.UpdatedAt)
	return &appv1.CurrentUser{
		Id:           user.ID,
		PrimaryEmail: &primaryEmail,
		DisplayName:  &displayName,
		AvatarUrl:    &avatarURL,
		CreatedAt:    &createdAt,
		UpdatedAt:    &updatedAt,
	}
}

func accountToProto(account *accounts.Account) *appv1.Account {
	if account == nil {
		return nil
	}
	return &appv1.Account{
		Id:             account.ID,
		UserId:         account.UserID,
		Label:          account.Label,
		ProviderHint:   account.ProviderHint,
		Server:         account.Server,
		Port:           int32(account.Port),
		Username:       account.Username,
		MailboxDefault: account.MailboxDefault,
		Insecure:       account.Insecure,
		AuthKind:       account.AuthKind,
		SecretKeyId:    account.SecretKeyID,
		IsDefault:      account.IsDefault,
		McpEnabled:     account.MCPEnabled,
		CreatedAt:      formatHostedProtoTime(account.CreatedAt),
		UpdatedAt:      formatHostedProtoTime(account.UpdatedAt),
	}
}

func accountListItemToProto(item accounts.AccountListItem) *appv1.AccountListItem {
	ret := &appv1.AccountListItem{
		Id:             item.ID,
		UserId:         item.UserID,
		Label:          item.Label,
		ProviderHint:   item.ProviderHint,
		Server:         item.Server,
		Port:           int32(item.Port),
		Username:       item.Username,
		MailboxDefault: item.MailboxDefault,
		Insecure:       item.Insecure,
		AuthKind:       item.AuthKind,
		SecretKeyId:    item.SecretKeyID,
		IsDefault:      item.IsDefault,
		McpEnabled:     item.MCPEnabled,
		CreatedAt:      formatHostedProtoTime(item.CreatedAt),
		UpdatedAt:      formatHostedProtoTime(item.UpdatedAt),
	}
	if item.LatestTest != nil {
		warningCode := item.LatestTest.WarningCode
		errorCode := item.LatestTest.ErrorCode
		ret.LatestTest = &appv1.LatestTestSummary{
			Success:     item.LatestTest.Success,
			WarningCode: optionalStringPointer(warningCode),
			ErrorCode:   optionalStringPointer(errorCode),
			CreatedAt:   formatHostedProtoTime(item.LatestTest.CreatedAt),
		}
	}
	return ret
}

func accountsListToProto(items []accounts.AccountListItem) []*appv1.AccountListItem {
	ret := make([]*appv1.AccountListItem, 0, len(items))
	for _, item := range items {
		ret = append(ret, accountListItemToProto(item))
	}
	return ret
}

func createAccountRequestToDomain(req *appv1.CreateAccountRequest) accounts.CreateInput {
	if req == nil {
		return accounts.CreateInput{}
	}
	return accounts.CreateInput{
		Label:          req.GetLabel(),
		ProviderHint:   req.GetProviderHint(),
		Server:         req.GetServer(),
		Port:           int(req.GetPort()),
		Username:       req.GetUsername(),
		Password:       req.GetPassword(),
		MailboxDefault: req.GetMailboxDefault(),
		Insecure:       req.GetInsecure(),
		AuthKind:       req.GetAuthKind(),
		IsDefault:      req.GetIsDefault(),
		MCPEnabled:     req.GetMcpEnabled(),
	}
}

func updateAccountRequestToDomain(req *appv1.UpdateAccountRequest) accounts.UpdateInput {
	if req == nil {
		return accounts.UpdateInput{}
	}
	input := accounts.UpdateInput{}
	if req.Label != nil {
		input.Label = req.Label
	}
	if req.ProviderHint != nil {
		input.ProviderHint = req.ProviderHint
	}
	if req.Server != nil {
		input.Server = req.Server
	}
	if req.Port != nil {
		value := int(*req.Port)
		input.Port = &value
	}
	if req.Username != nil {
		input.Username = req.Username
	}
	if req.Password != nil {
		input.Password = req.Password
	}
	if req.MailboxDefault != nil {
		input.MailboxDefault = req.MailboxDefault
	}
	if req.Insecure != nil {
		input.Insecure = req.Insecure
	}
	if req.AuthKind != nil {
		input.AuthKind = req.AuthKind
	}
	if req.IsDefault != nil {
		input.IsDefault = req.IsDefault
	}
	if req.McpEnabled != nil {
		input.MCPEnabled = req.McpEnabled
	}
	return input
}

func testAccountRequestToDomain(req *appv1.TestAccountRequest) accounts.TestInput {
	if req == nil {
		return accounts.TestInput{}
	}
	return accounts.TestInput{Mode: req.GetMode()}
}

func testResultToProto(result *accounts.TestResult) (*appv1.TestResult, error) {
	if result == nil {
		return nil, nil
	}
	ret := &appv1.TestResult{
		Id:              result.ID,
		ImapAccountId:   result.IMAPAccountID,
		TestMode:        result.TestMode,
		Success:         result.Success,
		TcpOk:           result.TCPOK,
		LoginOk:         result.LoginOK,
		MailboxSelectOk: result.MailboxSelectOK,
		ListOk:          result.ListOK,
		SampleFetchOk:   result.SampleFetchOK,
		WriteProbeOk:    result.WriteProbeOK,
		WarningCode:     optionalStringPointer(result.WarningCode),
		ErrorCode:       optionalStringPointer(result.ErrorCode),
		ErrorMessage:    optionalStringPointer(result.ErrorMessage),
		CreatedAt:       formatHostedProtoTime(result.CreatedAt),
	}
	if len(result.Details) > 0 {
		structured, err := structpb.NewStruct(result.Details)
		if err != nil {
			return nil, err
		}
		ret.Details = structured
	}
	return ret, nil
}

func mailboxInfoToProto(mailbox accounts.MailboxInfo) *appv1.MailboxInfo {
	return &appv1.MailboxInfo{Name: mailbox.Name, Path: mailbox.Path}
}

func mailboxesToProto(mailboxes []accounts.MailboxInfo) []*appv1.MailboxInfo {
	ret := make([]*appv1.MailboxInfo, 0, len(mailboxes))
	for _, mailbox := range mailboxes {
		ret = append(ret, mailboxInfoToProto(mailbox))
	}
	return ret
}

func addressViewToProto(address accounts.AddressView) *appv1.AddressView {
	return &appv1.AddressView{
		Name:    optionalStringPointer(address.Name),
		Address: address.Address,
	}
}

func addressesToProto(addresses []accounts.AddressView) []*appv1.AddressView {
	ret := make([]*appv1.AddressView, 0, len(addresses))
	for _, address := range addresses {
		ret = append(ret, addressViewToProto(address))
	}
	return ret
}

func mimePartViewToProto(part accounts.MimePartView) *appv1.MimePartView {
	ret := &appv1.MimePartView{}
	if value := optionalStringPointer(part.Type); value != nil {
		ret.Type = value
	}
	if value := optionalStringPointer(part.Subtype); value != nil {
		ret.Subtype = value
	}
	if part.Size != 0 {
		value := part.Size
		ret.Size = &value
	}
	if value := optionalStringPointer(part.Content); value != nil {
		ret.Content = value
	}
	if value := optionalStringPointer(part.Filename); value != nil {
		ret.Filename = value
	}
	if value := optionalStringPointer(part.Charset); value != nil {
		ret.Charset = value
	}
	return ret
}

func mimePartsToProto(parts []accounts.MimePartView) []*appv1.MimePartView {
	ret := make([]*appv1.MimePartView, 0, len(parts))
	for _, part := range parts {
		ret = append(ret, mimePartViewToProto(part))
	}
	return ret
}

func messageViewToProto(message accounts.MessageView) *appv1.MessageView {
	ret := &appv1.MessageView{
		Uid:       message.UID,
		SeqNum:    message.SeqNum,
		From:      addressesToProto(message.From),
		To:        addressesToProto(message.To),
		Flags:     append([]string(nil), message.Flags...),
		Size:      message.Size,
		MimeParts: mimePartsToProto(message.MimeParts),
	}
	if value := optionalStringPointer(message.Subject); value != nil {
		ret.Subject = value
	}
	if value := optionalStringPointer(message.Date); value != nil {
		ret.Date = value
	}
	if message.TotalCount != 0 {
		value := message.TotalCount
		ret.TotalCount = &value
	}
	return ret
}

func messagesToProto(messages []accounts.MessageView) []*appv1.MessageView {
	ret := make([]*appv1.MessageView, 0, len(messages))
	for _, message := range messages {
		ret = append(ret, messageViewToProto(message))
	}
	return ret
}

func listMessagesRequestToDomain(req *appv1.ListMessagesRequest) accounts.ListMessagesInput {
	if req == nil {
		return accounts.ListMessagesInput{}
	}
	input := accounts.ListMessagesInput{
		Mailbox:     req.GetMailbox(),
		Limit:       int(req.GetLimit()),
		Offset:      int(req.GetOffset()),
		Query:       req.GetQuery(),
		UnreadOnly:  req.GetUnreadOnly(),
		ContentType: req.GetContentType(),
	}
	if req.IncludeContent != nil {
		input.IncludeContent = req.GetIncludeContent()
	}
	return input
}

func ruleRecordToProto(record *rules.RuleRecord) *appv1.RuleRecord {
	if record == nil {
		return nil
	}
	ret := &appv1.RuleRecord{
		Id:               record.ID,
		UserId:           record.UserID,
		ImapAccountId:    record.IMAPAccountID,
		Name:             record.Name,
		Description:      record.Description,
		Status:           record.Status,
		SourceKind:       record.SourceKind,
		RuleYaml:         record.RuleYAML,
		LastPreviewCount: int32(record.LastPreviewCount),
		CreatedAt:        formatHostedProtoTime(record.CreatedAt),
		UpdatedAt:        formatHostedProtoTime(record.UpdatedAt),
	}
	if record.LastRunAt != nil && !record.LastRunAt.IsZero() {
		value := formatHostedProtoTime(*record.LastRunAt)
		ret.LastRunAt = &value
	}
	return ret
}

func rulesToProto(records []rules.RuleRecord) []*appv1.RuleRecord {
	ret := make([]*appv1.RuleRecord, 0, len(records))
	for i := range records {
		ret = append(ret, ruleRecordToProto(&records[i]))
	}
	return ret
}

func createRuleRequestToDomain(req *appv1.CreateRuleRequest) rules.CreateInput {
	if req == nil {
		return rules.CreateInput{}
	}
	return rules.CreateInput{
		IMAPAccountID: req.GetImapAccountId(),
		Name:          req.GetName(),
		Description:   req.GetDescription(),
		Status:        req.GetStatus(),
		SourceKind:    req.GetSourceKind(),
		RuleYAML:      req.GetRuleYaml(),
	}
}

func updateRuleRequestToDomain(req *appv1.UpdateRuleRequest) rules.UpdateInput {
	if req == nil {
		return rules.UpdateInput{}
	}
	return rules.UpdateInput{
		IMAPAccountID: req.ImapAccountId,
		Name:          req.Name,
		Description:   req.Description,
		Status:        req.Status,
		SourceKind:    req.SourceKind,
		RuleYAML:      req.RuleYaml,
	}
}

func dryRunRuleRequestToDomain(req *appv1.DryRunRuleRequest) rules.DryRunInput {
	if req == nil {
		return rules.DryRunInput{}
	}
	return rules.DryRunInput{IMAPAccountID: req.GetImapAccountId()}
}

func dryRunResultToProto(result *rules.DryRunResult) (*appv1.DryRunResult, error) {
	if result == nil {
		return nil, nil
	}
	ret := &appv1.DryRunResult{
		RuleId:        result.RuleID,
		ImapAccountId: result.IMAPAccountID,
		MatchedCount:  int32(result.MatchedCount),
		SampleRows:    messagesToProto(result.SampleRows),
		CreatedAt:     formatHostedProtoTime(result.CreatedAt),
	}
	if len(result.ActionPlan) > 0 {
		structured, err := structpb.NewStruct(result.ActionPlan)
		if err != nil {
			return nil, err
		}
		ret.ActionPlan = structured
	}
	return ret, nil
}

func optionalStringPointer(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
