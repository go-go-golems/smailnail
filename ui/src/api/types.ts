// API response envelope
export interface ApiResponse<T> {
  data: T;
  meta?: Record<string, unknown>;
}

export interface ApiError {
  error: {
    code: string;
    message: string;
    details?: Record<string, unknown>;
  };
}

// Account types matching backend JSON shapes
export interface Account {
  id: string;
  userId: string;
  label: string;
  providerHint: string;
  server: string;
  port: number;
  username: string;
  mailboxDefault: string;
  insecure: boolean;
  authKind: string;
  secretKeyId: string;
  isDefault: boolean;
  mcpEnabled: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface LatestTestSummary {
  success: boolean;
  warningCode?: string;
  errorCode?: string;
  createdAt: string;
}

export interface AccountListItem extends Account {
  latestTest?: LatestTestSummary;
}

export interface CreateAccountInput {
  label: string;
  providerHint?: string;
  server: string;
  port: number;
  username: string;
  password: string;
  mailboxDefault: string;
  insecure?: boolean;
  authKind?: string;
  isDefault?: boolean;
  mcpEnabled?: boolean;
}

export interface UpdateAccountInput {
  label?: string;
  providerHint?: string;
  server?: string;
  port?: number;
  username?: string;
  password?: string;
  mailboxDefault?: string;
  insecure?: boolean;
  authKind?: string;
  isDefault?: boolean;
  mcpEnabled?: boolean;
}

export interface TestResult {
  id: string;
  imapAccountId: string;
  testMode: string;
  success: boolean;
  tcpOk: boolean;
  loginOk: boolean;
  mailboxSelectOk: boolean;
  listOk: boolean;
  sampleFetchOk: boolean;
  writeProbeOk?: boolean;
  warningCode?: string;
  errorCode?: string;
  errorMessage?: string;
  detailsJson: string;
  createdAt: string;
}

export interface TestInput {
  mode?: string;
}

// Mailbox types
export interface MailboxInfo {
  name: string;
  path: string;
}

// Message types
export interface AddressView {
  name?: string;
  address: string;
}

export interface MimePartView {
  type?: string;
  subtype?: string;
  size?: number;
  content?: string;
  filename?: string;
  charset?: string;
}

export interface MessageView {
  uid: number;
  seqNum: number;
  subject?: string;
  from?: AddressView[];
  to?: AddressView[];
  date?: string;
  flags?: string[];
  size: number;
  mimeParts?: MimePartView[];
  totalCount?: number;
}

export interface ListMessagesParams {
  mailbox: string;
  limit?: number;
  offset?: number;
  query?: string;
  unreadOnly?: boolean;
  includeContent?: boolean;
  contentType?: string;
}
