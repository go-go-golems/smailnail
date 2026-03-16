import { useEffect } from "react";
import { useAppDispatch, useAppSelector } from "../../store";
import { AccountForm } from "./AccountForm";
import { AccountList } from "./AccountList";
import {
  createAccountAndTest,
  fetchAccounts,
  retestAccount,
  setFormDraft,
  setViewMode,
  startAddAccount,
  startEditAccount,
  updateAccountAndTest,
} from "./accountsSlice";
import { EmptyState } from "./EmptyState";
import { TestProgress } from "./TestProgress";
import { TestResultView } from "./TestResultView";

export function AccountSetupPage() {
  const dispatch = useAppDispatch();
  const {
    accounts,
    loadState,
    formDraft,
    editingAccountId,
    saveState,
    testState,
    viewMode,
    latestAccount,
    latestTestResult,
    error,
  } = useAppSelector((s) => s.accounts);

  useEffect(() => {
    if (loadState === "idle") {
      dispatch(fetchAccounts());
    }
  }, [dispatch, loadState]);

  if (loadState === "loading") {
    return (
      <div className="text-center py-5">
        <div className="spinner-border text-secondary" role="status">
          <span className="visually-hidden">Loading...</span>
        </div>
      </div>
    );
  }

  // Empty state
  if (viewMode === "list" && accounts.length === 0 && loadState === "loaded") {
    return <EmptyState onAddAccount={() => dispatch(startAddAccount())} />;
  }

  // Account list
  if (viewMode === "list" && accounts.length > 0) {
    return (
      <AccountList
        accounts={accounts}
        onAdd={() => dispatch(startAddAccount())}
        onEdit={(a) => dispatch(startEditAccount(a))}
      />
    );
  }

  // Add/edit form
  if (viewMode === "form") {
    return (
      <AccountForm
        draft={formDraft}
        isEditing={!!editingAccountId}
        saving={saveState === "saving"}
        error={error}
        onFieldChange={(patch) => dispatch(setFormDraft(patch))}
        onSubmit={() => {
          if (editingAccountId) {
            dispatch(updateAccountAndTest({ id: editingAccountId, draft: formDraft }));
          } else {
            dispatch(createAccountAndTest(formDraft));
          }
        }}
        onCancel={() => {
          dispatch(setViewMode("list"));
          dispatch(fetchAccounts());
        }}
      />
    );
  }

  // Testing in progress
  if (viewMode === "testing") {
    return <TestProgress />;
  }

  // Test result
  if (viewMode === "result" && latestTestResult) {
    return (
      <TestResultView
        testResult={latestTestResult}
        testState={testState}
        onEdit={() => dispatch(setViewMode("form"))}
        onRetest={() => {
          const id = latestAccount?.id ?? editingAccountId;
          if (id) dispatch(retestAccount(id));
        }}
        onDone={() => {
          dispatch(setViewMode("list"));
          dispatch(fetchAccounts());
        }}
      />
    );
  }

  return null;
}
