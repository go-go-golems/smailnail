import type { TestResult } from "../../api/types";

// Test stage for the progress checklist
export interface TestStage {
  key: string;
  label: string;
  resultField: keyof Pick<
    TestResult,
    "tcpOk" | "loginOk" | "mailboxSelectOk" | "listOk" | "sampleFetchOk"
  >;
}

export const testStages: TestStage[] = [
  { key: "tcp", label: "Connecting securely", resultField: "tcpOk" },
  { key: "login", label: "Signing in", resultField: "loginOk" },
  {
    key: "mailbox",
    label: "Opening mailbox",
    resultField: "mailboxSelectOk",
  },
  { key: "list", label: "Listing mailboxes", resultField: "listOk" },
  {
    key: "fetch",
    label: "Fetching sample message",
    resultField: "sampleFetchOk",
  },
];

// Maps backend error codes to user-facing messages
export const errorMessages: Record<string, { headline: string; hints: string[] }> = {
  "account-test-connect-failed": {
    headline: "Could not reach the IMAP server.",
    hints: [
      "IMAP server hostname or IP",
      "Port number (usually 993)",
      "Network or firewall settings",
    ],
  },
  "account-test-login-failed": {
    headline: "The server was reachable, but sign-in failed.",
    hints: [
      "Username or email",
      "Password or app password",
      "Provider-specific IMAP access settings",
    ],
  },
  "account-test-mailbox-select-failed": {
    headline: "The account worked, but the default mailbox could not be opened.",
    hints: [
      "Default mailbox name (usually INBOX)",
      "Mailbox permissions",
    ],
  },
  "account-test-mailbox-list-failed": {
    headline: "Sign-in worked, but mailbox listing failed.",
    hints: [
      "IMAP server permissions",
      "Account access level",
    ],
  },
  "account-test-sample-fetch-failed": {
    headline: "Connection worked, but fetching a sample message failed.",
    hints: [
      "The mailbox might be empty",
      "Message access permissions",
    ],
  },
};

// Maps warning codes to user-facing copy
export const warningMessages: Record<string, { headline: string; body: string }> = {
  "tls-verification-disabled": {
    headline: "TLS verification is disabled",
    body: "This is acceptable for local testing, but should not stay this way for a production mailbox.",
  },
  "provider-app-password-recommended": {
    headline: "App password recommended",
    body: "This provider works best with an app-specific password. Using your main password may stop working.",
  },
};
