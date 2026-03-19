import { useState } from "react";
import type { RuleRecord } from "../../api/types";

interface RuleListProps {
  rules: RuleRecord[];
  loading: boolean;
  onAdd: () => void;
  onSelect: (rule: RuleRecord) => void;
  onDelete: (rule: RuleRecord) => void;
}

export function RuleList({ rules, loading, onAdd, onSelect, onDelete }: RuleListProps) {
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);

  if (loading) {
    return (
      <div className="text-center py-4">
        <div className="spinner-border spinner-border-sm text-secondary" role="status" />
      </div>
    );
  }

  return (
    <div data-widget="rules" data-part="rule-list">
      <div className="d-flex justify-content-between align-items-center mb-3">
        <h2 className="h4 mb-0">Rules</h2>
        <button className="btn btn-primary btn-sm" onClick={onAdd}>
          New rule
        </button>
      </div>

      {rules.length === 0 ? (
        <div className="text-center py-4">
          <p className="text-body-secondary mb-3">
            No rules yet. Create one to filter and act on your mailbox messages.
          </p>
          <button className="btn btn-primary" onClick={onAdd}>
            Create first rule
          </button>
        </div>
      ) : (
        <div className="list-group">
          {rules.map((rule) => (
            <div key={rule.id} className="list-group-item">
              {confirmDeleteId === rule.id ? (
                <div className="d-flex justify-content-between align-items-center">
                  <span className="text-danger small">
                    Delete <strong>{rule.name}</strong>?
                  </span>
                  <div className="btn-group btn-group-sm">
                    <button
                      className="btn btn-outline-secondary"
                      onClick={() => setConfirmDeleteId(null)}
                    >
                      Cancel
                    </button>
                    <button
                      className="btn btn-danger"
                      onClick={() => {
                        setConfirmDeleteId(null);
                        onDelete(rule);
                      }}
                    >
                      Delete
                    </button>
                  </div>
                </div>
              ) : (
                <div className="d-flex justify-content-between align-items-start">
                  <div
                    className="flex-grow-1"
                    role="button"
                    tabIndex={0}
                    onClick={() => onSelect(rule)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter") onSelect(rule);
                    }}
                  >
                    <div className="fw-medium">{rule.name}</div>
                    {rule.description && (
                      <small className="text-body-secondary d-block">
                        {rule.description}
                      </small>
                    )}
                    <small className="text-body-tertiary">
                      {rule.lastRunAt
                        ? `Last run: ${new Date(rule.lastRunAt).toLocaleString()} (${rule.lastPreviewCount} matches)`
                        : "Never run"}
                    </small>
                  </div>
                  <div className="d-flex align-items-center gap-2">
                    <StatusBadge status={rule.status} />
                    <button
                      className="btn btn-outline-danger btn-sm"
                      title="Delete"
                      onClick={(e) => {
                        e.stopPropagation();
                        setConfirmDeleteId(rule.id);
                      }}
                    >
                      &times;
                    </button>
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  if (status === "active") {
    return <span className="badge bg-success-subtle text-success-emphasis">active</span>;
  }
  if (status === "paused") {
    return <span className="badge bg-warning-subtle text-warning-emphasis">paused</span>;
  }
  return <span className="badge bg-body-secondary text-body-secondary">{status}</span>;
}
