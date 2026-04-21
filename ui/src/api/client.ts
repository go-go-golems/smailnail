import type {
  AccountsResponse,
  CreateAccountInput,
  CreateAccountResponse,
  CreateRuleInput,
  CreateRuleResponse,
  DryRunInput,
  GetCurrentUserResponse,
  DryRunResponse,
  GetAccountResponse,
  GetRuleResponse,
  MailboxesResponse,
  ListMessagesParams,
  MessageResponse,
  MessagesResponse,
  RulesResponse,
  TestInput,
  TestAccountResultResponse,
  UpdateAccountInput,
  UpdateAccountResponse,
  UpdateRuleInput,
  UpdateRuleResponse,
} from "./types";

class ApiClient {
  private baseUrl = "/api";

  private async request<T>(path: string, options?: RequestInit): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      ...options,
    });

    if (response.status === 204) {
      return {} as T;
    }

    const body = await response.json();

    if (!response.ok) {
      const errorCode = body?.error?.code ?? "unknown";
      const errorMessage = body?.error?.message ?? response.statusText;
      throw new ApiRequestError(errorCode, errorMessage, response.status);
    }

    return body as T;
  }

  async listAccounts(): Promise<AccountsResponse> {
    return this.request<AccountsResponse>("/accounts");
  }

  async getCurrentUser(): Promise<GetCurrentUserResponse> {
    return this.request<GetCurrentUserResponse>("/me");
  }

  async createAccount(input: CreateAccountInput): Promise<CreateAccountResponse> {
    return this.request<CreateAccountResponse>("/accounts", {
      method: "POST",
      body: JSON.stringify({
        providerHint: "",
        insecure: false,
        authKind: "",
        isDefault: false,
        mcpEnabled: false,
        ...input,
      }),
    });
  }

  async getAccount(id: string): Promise<GetAccountResponse> {
    return this.request<GetAccountResponse>(`/accounts/${encodeURIComponent(id)}`);
  }

  async updateAccount(
    id: string,
    input: UpdateAccountInput,
  ): Promise<UpdateAccountResponse> {
    return this.request<UpdateAccountResponse>(`/accounts/${encodeURIComponent(id)}`, {
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
  ): Promise<TestAccountResultResponse> {
    return this.request<TestAccountResultResponse>(
      `/accounts/${encodeURIComponent(id)}/test`,
      {
        method: "POST",
        body: input ? JSON.stringify(input) : undefined,
      },
    );
  }

  async listMailboxes(accountId: string): Promise<MailboxesResponse> {
    return this.request<MailboxesResponse>(`/accounts/${encodeURIComponent(accountId)}/mailboxes`);
  }

  async listMessages(
    accountId: string,
    params: ListMessagesParams,
  ): Promise<MessagesResponse> {
    const qs = new URLSearchParams();
    qs.set("mailbox", params.mailbox);
    if (params.limit !== undefined) qs.set("limit", String(params.limit));
    if (params.offset !== undefined) qs.set("offset", String(params.offset));
    if (params.query) qs.set("query", params.query);
    if (params.unreadOnly) qs.set("unread_only", "true");
    if (params.includeContent) qs.set("include_content", "true");
    if (params.contentType) qs.set("content_type", params.contentType);
    return this.request<MessagesResponse>(
      `/accounts/${encodeURIComponent(accountId)}/messages?${qs.toString()}`,
    );
  }

  async getMessage(
    accountId: string,
    mailbox: string,
    uid: number,
  ): Promise<MessageResponse> {
    const qs = new URLSearchParams({ mailbox });
    return this.request<MessageResponse>(
      `/accounts/${encodeURIComponent(accountId)}/messages/${uid}?${qs.toString()}`,
    );
  }

  async listRules(): Promise<RulesResponse> {
    return this.request<RulesResponse>("/rules");
  }

  async createRule(input: CreateRuleInput): Promise<CreateRuleResponse> {
    return this.request<CreateRuleResponse>("/rules", {
      method: "POST",
      body: JSON.stringify(input),
    });
  }

  async getRule(id: string): Promise<GetRuleResponse> {
    return this.request<GetRuleResponse>(`/rules/${encodeURIComponent(id)}`);
  }

  async updateRule(id: string, input: UpdateRuleInput): Promise<UpdateRuleResponse> {
    return this.request<UpdateRuleResponse>(`/rules/${encodeURIComponent(id)}`, {
      method: "PATCH",
      body: JSON.stringify(input),
    });
  }

  async deleteRule(id: string): Promise<void> {
    await this.request(`/rules/${encodeURIComponent(id)}`, {
      method: "DELETE",
    });
  }

  async dryRunRule(id: string, input?: DryRunInput): Promise<DryRunResponse> {
    return this.request<DryRunResponse>(`/rules/${encodeURIComponent(id)}/dry-run`, {
      method: "POST",
      body: input ? JSON.stringify(input) : undefined,
    });
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
