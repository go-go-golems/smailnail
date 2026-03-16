import type {
  Account,
  AccountListItem,
  ApiResponse,
  CurrentUser,
  CreateAccountInput,
  CreateRuleInput,
  DryRunInput,
  DryRunResult,
  ListMessagesParams,
  MailboxInfo,
  MessageView,
  RuleRecord,
  TestInput,
  TestResult,
  UpdateAccountInput,
  UpdateRuleInput,
} from "./types";

class ApiClient {
  private baseUrl = "/api";

  private async request<T>(
    path: string,
    options?: RequestInit,
  ): Promise<ApiResponse<T>> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      ...options,
    });

    if (response.status === 204) {
      return { data: undefined as T };
    }

    const body = await response.json();

    if (!response.ok) {
      const errorCode = body?.error?.code ?? "unknown";
      const errorMessage = body?.error?.message ?? response.statusText;
      throw new ApiRequestError(errorCode, errorMessage, response.status);
    }

    return body as ApiResponse<T>;
  }

  async listAccounts(): Promise<ApiResponse<AccountListItem[]>> {
    return this.request<AccountListItem[]>("/accounts");
  }

  async getCurrentUser(): Promise<ApiResponse<CurrentUser>> {
    return this.request<CurrentUser>("/me");
  }

  async createAccount(
    input: CreateAccountInput,
  ): Promise<ApiResponse<Account>> {
    return this.request<Account>("/accounts", {
      method: "POST",
      body: JSON.stringify(input),
    });
  }

  async getAccount(id: string): Promise<ApiResponse<Account>> {
    return this.request<Account>(`/accounts/${encodeURIComponent(id)}`);
  }

  async updateAccount(
    id: string,
    input: UpdateAccountInput,
  ): Promise<ApiResponse<Account>> {
    return this.request<Account>(`/accounts/${encodeURIComponent(id)}`, {
      method: "PATCH",
      body: JSON.stringify(input),
    });
  }

  async deleteAccount(id: string): Promise<void> {
    await this.request(`/accounts/${encodeURIComponent(id)}`, {
      method: "DELETE",
    });
  }

  async testAccount(
    id: string,
    input?: TestInput,
  ): Promise<ApiResponse<TestResult>> {
    return this.request<TestResult>(
      `/accounts/${encodeURIComponent(id)}/test`,
      {
        method: "POST",
        body: input ? JSON.stringify(input) : undefined,
      },
    );
  }

  async listMailboxes(
    accountId: string,
  ): Promise<ApiResponse<MailboxInfo[]>> {
    return this.request<MailboxInfo[]>(
      `/accounts/${encodeURIComponent(accountId)}/mailboxes`,
    );
  }

  async listMessages(
    accountId: string,
    params: ListMessagesParams,
  ): Promise<ApiResponse<MessageView[]>> {
    const qs = new URLSearchParams();
    qs.set("mailbox", params.mailbox);
    if (params.limit !== undefined) qs.set("limit", String(params.limit));
    if (params.offset !== undefined) qs.set("offset", String(params.offset));
    if (params.query) qs.set("query", params.query);
    if (params.unreadOnly) qs.set("unread_only", "true");
    if (params.includeContent) qs.set("include_content", "true");
    if (params.contentType) qs.set("content_type", params.contentType);
    return this.request<MessageView[]>(
      `/accounts/${encodeURIComponent(accountId)}/messages?${qs.toString()}`,
    );
  }

  async getMessage(
    accountId: string,
    mailbox: string,
    uid: number,
  ): Promise<ApiResponse<MessageView>> {
    const qs = new URLSearchParams({ mailbox });
    return this.request<MessageView>(
      `/accounts/${encodeURIComponent(accountId)}/messages/${uid}?${qs.toString()}`,
    );
  }

  // Rules
  async listRules(): Promise<ApiResponse<RuleRecord[]>> {
    return this.request<RuleRecord[]>("/rules");
  }

  async createRule(input: CreateRuleInput): Promise<ApiResponse<RuleRecord>> {
    return this.request<RuleRecord>("/rules", {
      method: "POST",
      body: JSON.stringify(input),
    });
  }

  async getRule(id: string): Promise<ApiResponse<RuleRecord>> {
    return this.request<RuleRecord>(`/rules/${encodeURIComponent(id)}`);
  }

  async updateRule(
    id: string,
    input: UpdateRuleInput,
  ): Promise<ApiResponse<RuleRecord>> {
    return this.request<RuleRecord>(`/rules/${encodeURIComponent(id)}`, {
      method: "PATCH",
      body: JSON.stringify(input),
    });
  }

  async deleteRule(id: string): Promise<void> {
    await this.request(`/rules/${encodeURIComponent(id)}`, {
      method: "DELETE",
    });
  }

  async dryRunRule(
    id: string,
    input?: DryRunInput,
  ): Promise<ApiResponse<DryRunResult>> {
    return this.request<DryRunResult>(
      `/rules/${encodeURIComponent(id)}/dry-run`,
      {
        method: "POST",
        body: input ? JSON.stringify(input) : undefined,
      },
    );
  }
}

export class ApiRequestError extends Error {
  constructor(
    public code: string,
    message: string,
    public status: number,
  ) {
    super(message);
    this.name = "ApiRequestError";
  }
}

export const api = new ApiClient();
