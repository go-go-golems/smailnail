import { createAsyncThunk, createSlice } from "@reduxjs/toolkit";
import { api } from "../../api/client";
import type {
  CreateRuleInput,
  DryRunResult,
  RuleRecord,
  UpdateRuleInput,
} from "../../api/types";

export interface RuleFormDraft {
  name: string;
  description: string;
  imapAccountId: string;
  status: string;
  ruleYaml: string;
}

export const initialRuleDraft: RuleFormDraft = {
  name: "",
  description: "",
  imapAccountId: "",
  status: "active",
  ruleYaml: "",
};

export type RulesViewMode = "list" | "form" | "detail" | "dryrun";

interface RulesState {
  rules: RuleRecord[];
  loadState: "idle" | "loading" | "loaded" | "error";
  viewMode: RulesViewMode;
  formDraft: RuleFormDraft;
  editingRuleId: string | null;
  saveState: "idle" | "saving" | "saved" | "error";
  selectedRule: RuleRecord | null;
  dryRunResult: DryRunResult | null;
  dryRunState: "idle" | "running" | "done" | "error";
  error: string | null;
}

const initialState: RulesState = {
  rules: [],
  loadState: "idle",
  viewMode: "list",
  formDraft: initialRuleDraft,
  editingRuleId: null,
  saveState: "idle",
  selectedRule: null,
  dryRunResult: null,
  dryRunState: "idle",
  error: null,
};

export const fetchRules = createAsyncThunk("rules/fetchRules", async () => {
  const res = await api.listRules();
  return res.data;
});

export const createRule = createAsyncThunk(
  "rules/create",
  async (draft: RuleFormDraft) => {
    const input: CreateRuleInput = {
      imapAccountId: draft.imapAccountId,
      name: draft.name,
      description: draft.description,
      status: draft.status,
      sourceKind: "yaml",
      ruleYaml: draft.ruleYaml,
    };
    const res = await api.createRule(input);
    return res.data;
  },
);

export const updateRule = createAsyncThunk(
  "rules/update",
  async ({ id, draft }: { id: string; draft: RuleFormDraft }) => {
    const input: UpdateRuleInput = {
      imapAccountId: draft.imapAccountId,
      name: draft.name,
      description: draft.description,
      status: draft.status,
      ruleYaml: draft.ruleYaml,
    };
    const res = await api.updateRule(id, input);
    return res.data;
  },
);

export const deleteRule = createAsyncThunk(
  "rules/delete",
  async (id: string) => {
    await api.deleteRule(id);
    return id;
  },
);

export const runDryRun = createAsyncThunk(
  "rules/dryRun",
  async ({ ruleId, imapAccountId }: { ruleId: string; imapAccountId?: string }) => {
    const res = await api.dryRunRule(
      ruleId,
      imapAccountId ? { imapAccountId } : undefined,
    );
    return res.data;
  },
);

const rulesSlice = createSlice({
  name: "rules",
  initialState,
  reducers: {
    setRuleFormDraft(state, action: { payload: Partial<RuleFormDraft> }) {
      Object.assign(state.formDraft, action.payload);
    },
    startAddRule(state) {
      state.formDraft = initialRuleDraft;
      state.editingRuleId = null;
      state.saveState = "idle";
      state.error = null;
      state.viewMode = "form";
    },
    startEditRule(state, action: { payload: RuleRecord }) {
      const r = action.payload;
      state.formDraft = {
        name: r.name,
        description: r.description,
        imapAccountId: r.imapAccountId,
        status: r.status,
        ruleYaml: r.ruleYaml,
      };
      state.editingRuleId = r.id;
      state.saveState = "idle";
      state.error = null;
      state.viewMode = "form";
    },
    selectRule(state, action: { payload: RuleRecord }) {
      state.selectedRule = action.payload;
      state.dryRunResult = null;
      state.dryRunState = "idle";
      state.viewMode = "detail";
    },
    setRulesViewMode(state, action: { payload: RulesViewMode }) {
      state.viewMode = action.payload;
      if (action.payload === "list") {
        state.selectedRule = null;
        state.dryRunResult = null;
        state.dryRunState = "idle";
        state.error = null;
      }
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchRules.pending, (state) => {
        state.loadState = "loading";
      })
      .addCase(fetchRules.fulfilled, (state, action) => {
        state.loadState = "loaded";
        state.rules = action.payload;
      })
      .addCase(fetchRules.rejected, (state, action) => {
        state.loadState = "error";
        state.error = action.error.message ?? "Failed to load rules";
      })

      .addCase(createRule.pending, (state) => {
        state.saveState = "saving";
        state.error = null;
      })
      .addCase(createRule.fulfilled, (state, action) => {
        state.saveState = "saved";
        state.rules.push(action.payload);
        state.selectedRule = action.payload;
        state.viewMode = "detail";
      })
      .addCase(createRule.rejected, (state, action) => {
        state.saveState = "error";
        state.error = action.error.message ?? "Failed to create rule";
      })

      .addCase(updateRule.pending, (state) => {
        state.saveState = "saving";
        state.error = null;
      })
      .addCase(updateRule.fulfilled, (state, action) => {
        state.saveState = "saved";
        const idx = state.rules.findIndex((r) => r.id === action.payload.id);
        if (idx >= 0) {
          state.rules[idx] = action.payload;
        }
        state.selectedRule = action.payload;
        state.viewMode = "detail";
      })
      .addCase(updateRule.rejected, (state, action) => {
        state.saveState = "error";
        state.error = action.error.message ?? "Failed to update rule";
      })

      .addCase(deleteRule.fulfilled, (state, action) => {
        state.rules = state.rules.filter((r) => r.id !== action.payload);
        if (state.selectedRule?.id === action.payload) {
          state.selectedRule = null;
        }
        state.viewMode = "list";
      })

      .addCase(runDryRun.pending, (state) => {
        state.dryRunState = "running";
        state.dryRunResult = null;
        state.error = null;
        state.viewMode = "dryrun";
      })
      .addCase(runDryRun.fulfilled, (state, action) => {
        state.dryRunState = "done";
        state.dryRunResult = action.payload;
      })
      .addCase(runDryRun.rejected, (state, action) => {
        state.dryRunState = "error";
        state.error = action.error.message ?? "Dry run failed";
      });
  },
});

export const {
  setRuleFormDraft,
  startAddRule,
  startEditRule,
  selectRule,
  setRulesViewMode,
} = rulesSlice.actions;

export const rulesReducer = rulesSlice.reducer;
