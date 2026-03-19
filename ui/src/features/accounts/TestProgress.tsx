import { WIDGET, parts } from "./parts";
import { testStages } from "./types";

export function TestProgress() {
  return (
    <div
      data-widget={WIDGET}
      data-part={parts.testProgress}
      className="mx-auto py-4"
      style={{ maxWidth: 480 }}
    >
      <h2 className="h4 mb-2">Testing connection</h2>
      <p className="text-body-secondary small mb-4">
        We are checking this account in read-only mode.
      </p>

      <ul className="list-unstyled">
        {testStages.map((stage, i) => (
          <li
            key={stage.key}
            data-part={parts.testStage}
            data-state={i === 0 ? "running" : "pending"}
            className="d-flex align-items-center mb-2"
          >
            {i === 0 ? (
              <span className="spinner-border spinner-border-sm text-primary me-2" role="status" />
            ) : (
              <span className="text-body-tertiary me-2" style={{ width: 16, textAlign: "center" }}>&bull;</span>
            )}
            <span className={i === 0 ? "" : "text-body-tertiary"}>
              {stage.label}
            </span>
          </li>
        ))}
      </ul>
    </div>
  );
}
