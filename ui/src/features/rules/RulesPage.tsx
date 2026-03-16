import { useEffect } from "react";
import { useAppDispatch, useAppSelector } from "../../store";
import { DryRunView } from "./DryRunView";
import { RuleDetail } from "./RuleDetail";
import { RuleForm } from "./RuleForm";
import { RuleList } from "./RuleList";
import {
  createRule,
  deleteRule,
  fetchRules,
  runDryRun,
  selectRule,
  setRuleFormDraft,
  setRulesViewMode,
  startAddRule,
  startEditRule,
  updateRule,
} from "./rulesSlice";

interface RulesPageProps {
  onBack: () => void;
}

export function RulesPage({ onBack }: RulesPageProps) {
  const dispatch = useAppDispatch();
  const {
    rules,
    loadState,
    viewMode,
    formDraft,
    editingRuleId,
    saveState,
    selectedRule,
    dryRunResult,
    dryRunState,
    error,
  } = useAppSelector((s) => s.rules);
  const accounts = useAppSelector((s) => s.accounts.accounts);

  useEffect(() => {
    if (loadState === "idle") {
      dispatch(fetchRules());
    }
  }, [dispatch, loadState]);

  // List
  if (viewMode === "list") {
    return (
      <div>
        <button
          className="btn btn-link text-decoration-none p-0 mb-3 text-body-secondary"
          onClick={onBack}
        >
          &larr; Back to accounts
        </button>
        <RuleList
          rules={rules}
          loading={loadState === "loading"}
          onAdd={() => dispatch(startAddRule())}
          onSelect={(r) => dispatch(selectRule(r))}
          onDelete={(r) => dispatch(deleteRule(r.id))}
        />
      </div>
    );
  }

  // Form
  if (viewMode === "form") {
    return (
      <RuleForm
        draft={formDraft}
        isEditing={!!editingRuleId}
        saving={saveState === "saving"}
        error={error}
        accounts={accounts}
        onFieldChange={(patch) => dispatch(setRuleFormDraft(patch))}
        onSubmit={() => {
          if (editingRuleId) {
            dispatch(updateRule({ id: editingRuleId, draft: formDraft }));
          } else {
            dispatch(createRule(formDraft));
          }
        }}
        onCancel={() => {
          dispatch(setRulesViewMode("list"));
        }}
      />
    );
  }

  // Detail
  if (viewMode === "detail" && selectedRule) {
    return (
      <RuleDetail
        rule={selectedRule}
        onEdit={() => dispatch(startEditRule(selectedRule))}
        onDryRun={() => dispatch(runDryRun({ ruleId: selectedRule.id }))}
        onBack={() => dispatch(setRulesViewMode("list"))}
      />
    );
  }

  // Dry run
  if (viewMode === "dryrun" && selectedRule) {
    return (
      <DryRunView
        rule={selectedRule}
        result={dryRunResult}
        state={dryRunState}
        error={error}
        onBack={() => dispatch(setRulesViewMode("detail"))}
        onRerun={() => dispatch(runDryRun({ ruleId: selectedRule.id }))}
      />
    );
  }

  return null;
}
