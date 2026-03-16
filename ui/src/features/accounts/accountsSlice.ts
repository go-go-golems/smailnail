import { createAsyncThunk, createSlice } from "@reduxjs/toolkit";
import { api, ApiRequestError } from "../../api/client";
import type {
  Account,
  AccountListItem,
  CreateAccountInput,
  TestResult,
  UpdateAccountInput,
} from "../../api/types";

// Form draft shape
export interface AccountFormDraft {
  label: string;
  username: string;
  server: string;
  port: number;
  password: string;
  mailboxDefault: string;
  insecure: boolean;
  isDefault: boolean;
}

export const initialFormDraft: AccountFormDraft = {
  label: "",
  username: "",
  server: "",
  port: 993,
  password: "",
  mailboxDefault: "INBOX",
  insecure: false,
  isDefault: false,
};

export type SaveState = "idle" | "saving" | "saved" | "error";
export type TestState = "idle" | "running" | "success" | "warning" | "failure";
export type ViewMode = "list" | "form" | "testing" | "result";

interface UIError {
  code: string;
  message: string;
}

interface AccountsState {
  accounts: AccountListItem[];
  loadState: "idle" | "loading" | "loaded" | "error";
  formDraft: AccountFormDraft;
  editingAccountId: string | null;
  saveState: SaveState;
  testState: TestState;
  viewMode: ViewMode;
  latestAccount: Account | null;
  latestTestResult: TestResult | null;
  error: UIError | null;
}

const initialState: AccountsState = {
  accounts: [],
  loadState: "idle",
  formDraft: initialFormDraft,
  editingAccountId: null,
  saveState: "idle",
  testState: "idle",
  viewMode: "list",
  latestAccount: null,
  latestTestResult: null,
  error: null,
};

export const fetchAccounts = createAsyncThunk(
  "accounts/fetchAccounts",
  async () => {
    const res = await api.listAccounts();
    return res.data;
  },
);

export const createAccountAndTest = createAsyncThunk(
  "accounts/createAndTest",
  async (draft: AccountFormDraft, { dispatch }) => {
    const input: CreateAccountInput = {
      label: draft.label,
      server: draft.server,
      port: draft.port,
      username: draft.username,
      password: draft.password,
      mailboxDefault: draft.mailboxDefault,
      insecure: draft.insecure,
      isDefault: draft.isDefault,
    };

    const accountRes = await api.createAccount(input);
    const account = accountRes.data;

    dispatch(accountsSlice.actions.setLatestAccount(account));
    dispatch(accountsSlice.actions.setSaveState("saved"));
    dispatch(accountsSlice.actions.setTestState("running"));

    const testRes = await api.testAccount(account.id, { mode: "read_only" });
    return { account, testResult: testRes.data };
  },
);

export const updateAccountAndTest = createAsyncThunk(
  "accounts/updateAndTest",
  async (
    { id, draft }: { id: string; draft: AccountFormDraft },
    { dispatch },
  ) => {
    const input: UpdateAccountInput = {
      label: draft.label,
      server: draft.server,
      port: draft.port,
      username: draft.username,
      password: draft.password,
      mailboxDefault: draft.mailboxDefault,
      insecure: draft.insecure,
      isDefault: draft.isDefault,
    };

    const accountRes = await api.updateAccount(id, input);
    const account = accountRes.data;

    dispatch(accountsSlice.actions.setLatestAccount(account));
    dispatch(accountsSlice.actions.setSaveState("saved"));
    dispatch(accountsSlice.actions.setTestState("running"));

    const testRes = await api.testAccount(account.id, { mode: "read_only" });
    return { account, testResult: testRes.data };
  },
);

export const retestAccount = createAsyncThunk(
  "accounts/retest",
  async (id: string) => {
    const testRes = await api.testAccount(id, { mode: "read_only" });
    return testRes.data;
  },
);

export const deleteAccount = createAsyncThunk(
  "accounts/delete",
  async (id: string) => {
    await api.deleteAccount(id);
    return id;
  },
);

const accountsSlice = createSlice({
  name: "accounts",
  initialState,
  reducers: {
    setFormDraft(state, action: { payload: Partial<AccountFormDraft> }) {
      Object.assign(state.formDraft, action.payload);
    },
    resetForm(state) {
      state.formDraft = initialFormDraft;
      state.editingAccountId = null;
      state.saveState = "idle";
      state.testState = "idle";
      state.latestAccount = null;
      state.latestTestResult = null;
      state.error = null;
    },
    startAddAccount(state) {
      state.formDraft = initialFormDraft;
      state.editingAccountId = null;
      state.saveState = "idle";
      state.testState = "idle";
      state.latestTestResult = null;
      state.error = null;
      state.viewMode = "form";
    },
    startEditAccount(state, action: { payload: AccountListItem }) {
      const a = action.payload;
      state.formDraft = {
        label: a.label,
        username: a.username,
        server: a.server,
        port: a.port,
        password: "",
        mailboxDefault: a.mailboxDefault,
        insecure: a.insecure,
        isDefault: a.isDefault,
      };
      state.editingAccountId = a.id;
      state.saveState = "idle";
      state.testState = "idle";
      state.latestTestResult = null;
      state.error = null;
      state.viewMode = "form";
    },
    setViewMode(state, action: { payload: ViewMode }) {
      state.viewMode = action.payload;
    },
    setSaveState(state, action: { payload: SaveState }) {
      state.saveState = action.payload;
    },
    setTestState(state, action: { payload: TestState }) {
      state.testState = action.payload;
    },
    setLatestAccount(state, action: { payload: Account }) {
      state.latestAccount = action.payload;
    },
  },
  extraReducers: (builder) => {
    builder
      // fetchAccounts
      .addCase(fetchAccounts.pending, (state) => {
        state.loadState = "loading";
      })
      .addCase(fetchAccounts.fulfilled, (state, action) => {
        state.loadState = "loaded";
        state.accounts = action.payload;
        state.viewMode =
          action.payload.length === 0 ? "list" : state.viewMode;
      })
      .addCase(fetchAccounts.rejected, (state, action) => {
        state.loadState = "error";
        state.error = {
          code: "fetch-failed",
          message: action.error.message ?? "Failed to load accounts",
        };
      })

      // createAccountAndTest
      .addCase(createAccountAndTest.pending, (state) => {
        state.saveState = "saving";
        state.testState = "idle";
        state.error = null;
        state.viewMode = "testing";
      })
      .addCase(createAccountAndTest.fulfilled, (state, action) => {
        const { testResult } = action.payload;
        state.latestTestResult = testResult;
        state.viewMode = "result";
        if (testResult.success && testResult.warningCode) {
          state.testState = "warning";
        } else if (testResult.success) {
          state.testState = "success";
        } else {
          state.testState = "failure";
        }
      })
      .addCase(createAccountAndTest.rejected, (state, action) => {
        state.saveState = "error";
        state.viewMode = "form";
        const err = action.error;
        if (action.payload instanceof ApiRequestError) {
          state.error = {
            code: action.payload.code,
            message: action.payload.message,
          };
        } else {
          state.error = {
            code: "save-failed",
            message: err.message ?? "Failed to save account",
          };
        }
      })

      // updateAccountAndTest
      .addCase(updateAccountAndTest.pending, (state) => {
        state.saveState = "saving";
        state.testState = "idle";
        state.error = null;
        state.viewMode = "testing";
      })
      .addCase(updateAccountAndTest.fulfilled, (state, action) => {
        const { testResult } = action.payload;
        state.latestTestResult = testResult;
        state.viewMode = "result";
        if (testResult.success && testResult.warningCode) {
          state.testState = "warning";
        } else if (testResult.success) {
          state.testState = "success";
        } else {
          state.testState = "failure";
        }
      })
      .addCase(updateAccountAndTest.rejected, (state, action) => {
        state.saveState = "error";
        state.viewMode = "form";
        state.error = {
          code: "update-failed",
          message: action.error.message ?? "Failed to update account",
        };
      })

      // retestAccount
      .addCase(retestAccount.pending, (state) => {
        state.testState = "running";
        state.viewMode = "testing";
        state.error = null;
      })
      .addCase(retestAccount.fulfilled, (state, action) => {
        const testResult = action.payload;
        state.latestTestResult = testResult;
        state.viewMode = "result";
        if (testResult.success && testResult.warningCode) {
          state.testState = "warning";
        } else if (testResult.success) {
          state.testState = "success";
        } else {
          state.testState = "failure";
        }
      })
      .addCase(retestAccount.rejected, (state, action) => {
        state.testState = "failure";
        state.viewMode = "result";
        state.error = {
          code: "test-failed",
          message: action.error.message ?? "Connection test failed",
        };
      })

      // deleteAccount
      .addCase(deleteAccount.fulfilled, (state, action) => {
        state.accounts = state.accounts.filter(
          (a) => a.id !== action.payload,
        );
        if (state.latestAccount?.id === action.payload) {
          state.latestAccount = null;
          state.latestTestResult = null;
        }
        state.viewMode = "list";
        state.editingAccountId = null;
      });
  },
});

export const {
  setFormDraft,
  resetForm,
  startAddAccount,
  startEditAccount,
  setViewMode,
} = accountsSlice.actions;

export const accountsReducer = accountsSlice.reducer;
