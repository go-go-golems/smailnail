import type { TestResult } from "../../api/types";
import type { TestState } from "./accountsSlice";
import { WIDGET, parts } from "./parts";
import { errorMessages, testStages, warningMessages } from "./types";

interface TestResultViewProps {
  testResult: TestResult;
  testState: TestState;
  onEdit: () => void;
  onRetest: () => void;
  onDone: () => void;
}

export function TestResultView({
  testResult,
  testState,
  onEdit,
  onRetest,
  onDone,
}: TestResultViewProps) {
  const headline =
    testState === "success"
      ? "Connection looks good"
      : testState === "warning"
        ? "Connection works, but needs attention"
        : "Connection failed";

  const headlineClass =
    testState === "failure"
      ? "text-danger"
      : testState === "warning"
        ? "text-warning-emphasis"
        : "text-success";

  const borderClass =
    testState === "failure"
      ? "border-danger"
      : testState === "warning"
        ? "border-warning"
        : "border-success";

  const errorInfo = testResult.errorCode
    ? errorMessages[testResult.errorCode]
    : null;

  const warningInfo = testResult.warningCode
    ? warningMessages[testResult.warningCode]
    : null;

  // Parse details for sample info
  let details: Record<string, unknown> = {};
  if (testResult.detailsJson) {
    try {
      details = JSON.parse(testResult.detailsJson);
    } catch {
      // ignore
    }
  }

  return (
    <div
      data-widget={WIDGET}
      data-part={parts.resultPanel}
      data-state={testState}
      className={`mx-auto py-4 border-start ps-4 ${borderClass}`}
      style={{ maxWidth: 540, borderWidth: 3 }}
    >
      <h2 data-part={parts.resultHeadline} className={`h4 mb-2 ${headlineClass}`}>
        {headline}
      </h2>

      {testState !== "failure" && (
        <p className="text-body-secondary small mb-3">
          smailnail connected to this mailbox and completed the initial read-only checks.
        </p>
      )}

      {/* Checklist */}
      <ul data-part={parts.resultChecklist} className="list-unstyled mb-3">
        {testStages.map((stage) => {
          const ok = Boolean(testResult[stage.resultField]);
          return (
            <li key={stage.key} className="d-flex align-items-center mb-1">
              {ok ? (
                <span className="text-success me-2" aria-label="passed">&#10003;</span>
              ) : (
                <span className="text-danger me-2" aria-label="failed">&#10007;</span>
              )}
              <span className={ok ? "" : "text-danger"}>
                {stageResultLabel(stage.key, ok)}
              </span>
            </li>
          );
        })}
      </ul>

      {/* Sample info on success */}
      {testState === "success" && typeof details["sampleSubject"] === "string" && (
        <div data-part={parts.sampleInfo} className="bg-body-tertiary rounded p-3 mb-3 small">
          <div><strong>Sample mailbox:</strong> {String(details["sampleMailbox"] ?? "INBOX")}</div>
          <div><strong>Sample subject:</strong> {details["sampleSubject"]}</div>
        </div>
      )}

      {/* Warning banner */}
      {warningInfo && (
        <div data-part={parts.resultWarning} className="alert alert-warning" role="alert">
          <strong>{warningInfo.headline}</strong>
          <p className="mb-0 small">{warningInfo.body}</p>
        </div>
      )}

      {/* Error details */}
      {testState === "failure" && (
        <div data-part={parts.resultError}>
          {errorInfo && (
            <>
              <p className="mb-2">{errorInfo.headline}</p>
              <p className="text-body-secondary small mb-1">Try checking:</p>
              <ul className="small text-body-secondary mb-3">
                {errorInfo.hints.map((hint) => (
                  <li key={hint}>{hint}</li>
                ))}
              </ul>
            </>
          )}
          {testResult.errorMessage && (
            <details className="small mb-3">
              <summary className="text-body-tertiary">Technical details</summary>
              <code className="d-block mt-1 text-break">
                {testResult.errorCode}: {testResult.errorMessage}
              </code>
            </details>
          )}
        </div>
      )}

      {/* Actions */}
      <div data-part={parts.resultActions} className="d-flex justify-content-between mt-4">
        <button className="btn btn-outline-secondary" onClick={onEdit}>
          Edit account
        </button>
        <div className="d-flex gap-2">
          <button className="btn btn-outline-secondary" onClick={onRetest}>
            Test again
          </button>
          {testState !== "failure" && (
            <button className="btn btn-primary" onClick={onDone}>
              Done
            </button>
          )}
        </div>
      </div>
    </div>
  );
}

function stageResultLabel(key: string, ok: boolean): string {
  const labels: Record<string, [string, string]> = {
    tcp: ["Connected over TLS", "TLS connection failed"],
    login: ["Logged in", "Login failed"],
    mailbox: ["Opened mailbox", "Mailbox open failed"],
    list: ["Listed mailboxes", "Mailbox listing failed"],
    fetch: ["Fetched a sample message", "Sample fetch failed"],
  };
  const pair = labels[key];
  if (!pair) return key;
  return ok ? pair[0] : pair[1];
}
