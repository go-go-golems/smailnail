import type {
  Account,
  AccountListItem,
  ApiResponse,
  CreateAccountInput,
  TestInput,
  TestResult,
  UpdateAccountInput,
} from "./types";

class ApiClient {
  private baseUrl = "/api";

  private async request<T>(
    path: string,
    options?: RequestInit,
  ): Promise<ApiResponse<T>> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      headers: { "Content-Type": "application/json" },
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
