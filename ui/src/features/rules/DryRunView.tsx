import type { DryRunResult, RuleRecord } from "../../api/types";

interface DryRunViewProps {
  rule: RuleRecord;
  result: DryRunResult | null;
  state: "idle" | "running" | "done" | "error";
  error: string | null;
  onBack: () => void;
  onRerun: () => void;
}

export function DryRunView({
  rule,
  result,
  state,
  error,
  onBack,
  onRerun,
}: DryRunViewProps) {
  return (
    <div data-widget="rules" data-part="dryrun-view" className="mx-auto" style={{ maxWidth: 640 }}>
      <button
        className="btn btn-link text-decoration-none p-0 mb-3 text-body-secondary"
        onClick={onBack}
      >
        &larr; Back to {rule.name}
      </button>

      <h2 className="h4 mb-3">Dry run: {rule.name}</h2>

      {state === "running" && (
        <div className="text-center py-4">
          <div className="spinner-border text-primary" role="status" />
          <p className="text-body-secondary small mt-2">
            Running dry run against the IMAP account...
          </p>
        </div>
      )}

      {state === "error" && (
        <div className="alert alert-danger" role="alert">
          {error || "Dry run failed."}
          <div className="mt-2">
            <button className="btn btn-outline-danger btn-sm" onClick={onRerun}>
              Try again
            </button>
          </div>
        </div>
      )}

      {state === "done" && result && (
        <>
          <div className="alert alert-info" role="alert">
            <strong>{result.matchedCount}</strong> message{result.matchedCount !== 1 ? "s" : ""} matched.
          </div>

          {/* Action plan */}
          {result.actionPlan && Object.keys(result.actionPlan).length > 0 && (
            <div className="mb-3">
              <strong className="small">Action plan:</strong>
              <pre className="bg-body-tertiary rounded p-3 mt-1 small font-monospace" style={{ whiteSpace: "pre-wrap" }}>
                {JSON.stringify(result.actionPlan, null, 2)}
              </pre>
            </div>
          )}

          {/* Sample rows */}
          {result.sampleRows && result.sampleRows.length > 0 && (
            <div className="mb-3">
              <strong className="small">Sample matches ({result.sampleRows.length}):</strong>
              <div className="list-group mt-1">
                {result.sampleRows.map((row) => (
                  <div key={row.uid} className="list-group-item py-2">
                    <div className="fw-medium small">{row.subject || "(no subject)"}</div>
                    <div className="text-body-secondary" style={{ fontSize: "0.75rem" }}>
                      UID {row.uid}
                      {row.from && row.from.length > 0 && (
                        <> &mdash; {row.from[0]?.name || row.from[0]?.address}</>
                      )}
                      {row.date && <> &mdash; {new Date(row.date).toLocaleString()}</>}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          <div className="d-flex justify-content-end mt-3">
            <button className="btn btn-outline-secondary btn-sm" onClick={onRerun}>
              Run again
            </button>
          </div>
        </>
      )}
    </div>
  );
}
