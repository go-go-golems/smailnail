import type { MessageView } from "../../api/types";

interface MessageDetailProps {
  message: MessageView;
  loading: boolean;
  onBack: () => void;
}

export function MessageDetail({ message, loading, onBack }: MessageDetailProps) {
  if (loading) {
    return (
      <div className="text-center py-4">
        <div className="spinner-border spinner-border-sm text-secondary" role="status" />
      </div>
    );
  }

  const textPart = message.mimeParts?.find(
    (p) => p.type === "text" && p.subtype === "plain" && p.content,
  );
  const htmlPart = message.mimeParts?.find(
    (p) => p.type === "text" && p.subtype === "html" && p.content,
  );
  const attachments = message.mimeParts?.filter(
    (p) => p.filename,
  );

  return (
    <div data-widget="mailbox-explorer" data-part="message-detail">
      <button
        className="btn btn-link text-decoration-none p-0 mb-3 text-body-secondary"
        onClick={onBack}
      >
        &larr; Back to list
      </button>

      <div className="mb-3">
        <h3 className="h5 mb-1">{message.subject || "(no subject)"}</h3>
        <div className="text-body-secondary small">
          <div>
            <strong>From:</strong>{" "}
            {formatAddresses(message.from)}
          </div>
          <div>
            <strong>To:</strong>{" "}
            {formatAddresses(message.to)}
          </div>
          {message.date && (
            <div>
              <strong>Date:</strong>{" "}
              {new Date(message.date).toLocaleString()}
            </div>
          )}
          {message.flags && message.flags.length > 0 && (
            <div>
              <strong>Flags:</strong>{" "}
              {message.flags.map((f) => (
                <span key={f} className="badge bg-body-secondary text-body-secondary me-1">
                  {f.replace("\\", "")}
                </span>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Body */}
      {textPart?.content ? (
        <pre className="bg-body-tertiary rounded p-3 small" style={{ whiteSpace: "pre-wrap", wordBreak: "break-word" }}>
          {textPart.content}
        </pre>
      ) : htmlPart?.content ? (
        <div className="bg-body-tertiary rounded p-3 small">
          <em className="text-body-secondary">HTML content (preview not available)</em>
        </div>
      ) : (
        <p className="text-body-secondary small">No text content available.</p>
      )}

      {/* Attachments */}
      {attachments && attachments.length > 0 && (
        <div className="mt-3">
          <strong className="small">Attachments:</strong>
          <ul className="list-unstyled mt-1">
            {attachments.map((a, i) => (
              <li key={i} className="small text-body-secondary">
                {a.filename} ({a.type}/{a.subtype}, {formatSize(a.size ?? 0)})
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* MIME structure */}
      {message.mimeParts && message.mimeParts.length > 0 && (
        <details className="mt-3">
          <summary className="small text-body-tertiary">
            MIME structure ({message.mimeParts.length} parts, {formatSize(message.size)})
          </summary>
          <ul className="list-unstyled mt-1 ms-2">
            {message.mimeParts.map((p, i) => (
              <li key={i} className="small text-body-tertiary">
                {p.type}/{p.subtype}
                {p.filename ? ` (${p.filename})` : ""}
                {p.charset ? ` [${p.charset}]` : ""}
                {" "}&mdash; {formatSize(p.size ?? 0)}
              </li>
            ))}
          </ul>
        </details>
      )}
    </div>
  );
}

function formatAddresses(
  addresses?: { name?: string; address: string }[],
): string {
  if (!addresses || addresses.length === 0) return "(none)";
  return addresses
    .map((a) => (a.name ? `${a.name} <${a.address}>` : a.address))
    .join(", ");
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
