import type {
  Account as GeneratedAccount,
  AccountListItem as GeneratedAccountListItem,
  AccountResponse,
  CreateAccountRequest,
  CreateRuleRequest,
  CurrentUser as GeneratedCurrentUser,
  CurrentUserResponse,
  DryRunResult as GeneratedDryRunResult,
  DryRunRuleRequest,
  DryRunRuleResponse,
  ErrorResponse,
  GetMessageResponse,
  InfoResponse,
  ListAccountsMeta,
  ListAccountsResponse,
  ListMailboxesMeta,
  ListMailboxesResponse,
  ListMessagesMeta,
  ListMessagesRequest,
  ListMessagesResponse,
  ListRulesMeta,
  ListRulesResponse,
  MailboxInfo as GeneratedMailboxInfo,
  MessageView as GeneratedMessageView,
  RuleRecord as GeneratedRuleRecord,
  RuleResponse,
  TestAccountRequest,
  TestAccountResponse,
  TestResult as GeneratedTestResult,
  UpdateAccountRequest,
  UpdateRuleRequest,
} from "../gen/smailnail/app/v1/hosted";

export type ApiError = ErrorResponse;
export type AppInfo = InfoResponse;

export type CurrentUser = GeneratedCurrentUser;
export type GetCurrentUserResponse = Omit<CurrentUserResponse, "data"> & { data: CurrentUser };

export type Account = GeneratedAccount;
export type LatestTestSummary = NonNullable<GeneratedAccountListItem["latestTest"]>;
export type AccountListItem = GeneratedAccountListItem;
export type AccountsMeta = ListAccountsMeta;
export type AccountsResponse = ListAccountsResponse;
export type GetAccountResponse = Omit<AccountResponse, "data"> & { data: Account };
export type CreateAccountResponse = Omit<AccountResponse, "data"> & { data: Account };
export type UpdateAccountResponse = Omit<AccountResponse, "data"> & { data: Account };

export type CreateAccountInput = Omit<
  CreateAccountRequest,
  "providerHint" | "insecure" | "authKind" | "isDefault" | "mcpEnabled"
> & {
  providerHint?: string;
  insecure?: boolean;
  authKind?: string;
  isDefault?: boolean;
  mcpEnabled?: boolean;
};

export type UpdateAccountInput = UpdateAccountRequest;

export type TestResult = Omit<GeneratedTestResult, "details"> & {
  details?: Record<string, unknown>;
};
export type TestInput = TestAccountRequest;
export type TestAccountResultResponse = Omit<TestAccountResponse, "data"> & { data: TestResult };

export type MailboxInfo = GeneratedMailboxInfo;
export type MailboxesMeta = ListMailboxesMeta;
export type MailboxesResponse = ListMailboxesResponse;

export type MessageView = Omit<GeneratedMessageView, "from" | "to" | "mimeParts"> & {
  from?: GeneratedMessageView["from"];
  to?: GeneratedMessageView["to"];
  mimeParts?: GeneratedMessageView["mimeParts"];
};

export type ListMessagesParams = Omit<
  ListMessagesRequest,
  "limit" | "offset" | "query" | "unreadOnly" | "includeContent" | "contentType"
> & {
  limit?: number;
  offset?: number;
  query?: string;
  unreadOnly?: boolean;
  includeContent?: boolean;
  contentType?: string;
};

export type MessagesMeta = ListMessagesMeta;
export type MessagesResponse = ListMessagesResponse;
export type MessageResponse = Omit<GetMessageResponse, "data"> & { data: MessageView };

export type RuleRecord = GeneratedRuleRecord;
export type RulesMeta = ListRulesMeta;
export type RulesResponse = ListRulesResponse;
export type GetRuleResponse = Omit<RuleResponse, "data"> & { data: RuleRecord };
export type CreateRuleResponse = Omit<RuleResponse, "data"> & { data: RuleRecord };
export type UpdateRuleResponse = Omit<RuleResponse, "data"> & { data: RuleRecord };

export type CreateRuleInput = CreateRuleRequest;
export type UpdateRuleInput = UpdateRuleRequest;
export type DryRunInput = DryRunRuleRequest;
export type DryRunResult = Omit<GeneratedDryRunResult, "actionPlan" | "sampleRows"> & {
  actionPlan?: Record<string, unknown>;
  sampleRows: MessageView[];
};
export type DryRunResponse = Omit<DryRunRuleResponse, "data"> & { data: DryRunResult };
