import type { AccountListItem } from "../../api/types";
import type { RuleFormDraft } from "./rulesSlice";

interface RuleFormProps {
  draft: RuleFormDraft;
  isEditing: boolean;
  saving: boolean;
  error: string | null;
  accounts: AccountListItem[];
  onFieldChange: (patch: Partial<RuleFormDraft>) => void;
  onSubmit: () => void;
  onCancel: () => void;
}

export function RuleForm({
  draft,
  isEditing,
  saving,
  error,
  accounts,
  onFieldChange,
  onSubmit,
  onCancel,
}: RuleFormProps) {
  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    onSubmit();
  }

  return (
    <form
      data-widget="rules"
      data-part="rule-form"
      onSubmit={handleSubmit}
      className="mx-auto"
      style={{ maxWidth: 640 }}
    >
      <h2 className="h4 mb-3">{isEditing ? "Edit rule" : "New rule"}</h2>

      {error && (
        <div className="alert alert-danger" role="alert">
          {error}
        </div>
      )}

      <div className="mb-3">
        <label htmlFor="rule-name" className="form-label">Name</label>
        <input
          id="rule-name"
          type="text"
          className="form-control"
          value={draft.name}
          onChange={(e) => onFieldChange({ name: e.target.value })}
        />
      </div>

      <div className="mb-3">
        <label htmlFor="rule-desc" className="form-label">Description</label>
        <input
          id="rule-desc"
          type="text"
          className="form-control"
          placeholder="Optional"
          value={draft.description}
          onChange={(e) => onFieldChange({ description: e.target.value })}
        />
      </div>

      <div className="row mb-3">
        <div className="col-8">
          <label htmlFor="rule-account" className="form-label">IMAP account</label>
          <select
            id="rule-account"
            className="form-select"
            value={draft.imapAccountId}
            onChange={(e) => onFieldChange({ imapAccountId: e.target.value })}
          >
            <option value="">Select account...</option>
            {accounts.map((a) => (
              <option key={a.id} value={a.id}>
                {a.label || a.server}
              </option>
            ))}
          </select>
        </div>
        <div className="col-4">
          <label htmlFor="rule-status" className="form-label">Status</label>
          <select
            id="rule-status"
            className="form-select"
            value={draft.status}
            onChange={(e) => onFieldChange({ status: e.target.value })}
          >
            <option value="active">Active</option>
            <option value="paused">Paused</option>
            <option value="draft">Draft</option>
          </select>
        </div>
      </div>

      <div className="mb-3">
        <label htmlFor="rule-yaml" className="form-label">Rule YAML</label>
        <textarea
          id="rule-yaml"
          className="form-control font-monospace"
          rows={12}
          value={draft.ruleYaml}
          onChange={(e) => onFieldChange({ ruleYaml: e.target.value })}
          placeholder={YAML_PLACEHOLDER}
        />
        <div className="form-text">
          Define query, output, and actions using the smailnail DSL.
        </div>
      </div>

      <div className="d-flex justify-content-between mt-4">
        <button type="button" className="btn btn-outline-secondary" onClick={onCancel}>
          Cancel
        </button>
        <button type="submit" className="btn btn-primary" disabled={saving}>
          {saving ? (
            <>
              <span className="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true" />
              Saving...
            </>
          ) : (
            "Save rule"
          )}
        </button>
      </div>
    </form>
  );
}

const YAML_PLACEHOLDER = `query:
  mailbox: INBOX
  criteria:
    from: notifications@example.com

output:
  fields:
    - name: uid
    - name: subject
    - name: from`;
