import type { RuleRecord } from "../../api/types";

interface RuleDetailProps {
  rule: RuleRecord;
  onEdit: () => void;
  onDryRun: () => void;
  onBack: () => void;
}

export function RuleDetail({ rule, onEdit, onDryRun, onBack }: RuleDetailProps) {
  return (
    <div data-widget="rules" data-part="rule-detail" className="mx-auto" style={{ maxWidth: 640 }}>
      <button
        className="btn btn-link text-decoration-none p-0 mb-3 text-body-secondary"
        onClick={onBack}
      >
        &larr; Back to rules
      </button>

      <div className="d-flex justify-content-between align-items-start mb-3">
        <div>
          <h2 className="h4 mb-1">{rule.name}</h2>
          {rule.description && (
            <p className="text-body-secondary small mb-1">{rule.description}</p>
          )}
          <div className="d-flex gap-2 align-items-center">
            <StatusBadge status={rule.status} />
            <small className="text-body-tertiary">
              {rule.lastRunAt
                ? `Last run ${new Date(rule.lastRunAt).toLocaleString()} — ${rule.lastPreviewCount} matches`
                : "Never run"}
            </small>
          </div>
        </div>
        <div className="btn-group btn-group-sm">
          <button className="btn btn-outline-secondary" onClick={onEdit}>
            Edit
          </button>
          <button className="btn btn-primary" onClick={onDryRun}>
            Dry run
          </button>
        </div>
      </div>

      <div className="mb-3">
        <strong className="small">Rule YAML:</strong>
        <pre className="bg-body-tertiary rounded p-3 mt-1 small font-monospace" style={{ whiteSpace: "pre-wrap" }}>
          {rule.ruleYaml}
        </pre>
      </div>

      <div className="row small text-body-secondary">
        <div className="col-6">
          <strong>Created:</strong> {new Date(rule.createdAt).toLocaleString()}
        </div>
        <div className="col-6">
          <strong>Updated:</strong> {new Date(rule.updatedAt).toLocaleString()}
        </div>
      </div>
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
