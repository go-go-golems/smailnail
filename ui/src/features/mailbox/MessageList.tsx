import type { MessageView } from "../../api/types";

interface MessageListProps {
  messages: MessageView[];
  loading: boolean;
  totalCount: number;
  offset: number;
  limit: number;
  selectedUid: number | null;
  onSelect: (uid: number) => void;
  onPageChange: (offset: number) => void;
}

export function MessageList({
  messages,
  loading,
  totalCount,
  offset,
  limit,
  selectedUid,
  onSelect,
  onPageChange,
}: MessageListProps) {
  if (loading) {
    return (
      <div className="text-center py-4">
        <div className="spinner-border spinner-border-sm text-secondary" role="status" />
        <span className="ms-2 text-body-secondary small">Loading messages...</span>
      </div>
    );
  }

  if (messages.length === 0) {
    return (
      <p className="text-body-secondary small text-center py-4">
        No messages in this mailbox.
      </p>
    );
  }

  const hasPrev = offset > 0;
  const hasNext = offset + limit < totalCount;

  return (
    <div>
      <div className="list-group list-group-flush">
        {messages.map((msg) => {
          const isUnread = msg.flags && !msg.flags.includes("\\Seen");
          const fromDisplay = formatFrom(msg);
          return (
            <button
              key={msg.uid}
              className={`list-group-item list-group-item-action py-2 px-3 ${
                selectedUid === msg.uid ? "active" : ""
              }`}
              onClick={() => onSelect(msg.uid)}
            >
              <div className="d-flex justify-content-between align-items-start">
                <div className="text-truncate me-2" style={{ minWidth: 0 }}>
                  <div className={`text-truncate ${isUnread ? "fw-semibold" : ""}`}>
                    {msg.subject || "(no subject)"}
                  </div>
                  <small className="text-body-secondary text-truncate d-block">
                    {fromDisplay}
                  </small>
                </div>
                <small className="text-body-secondary text-nowrap flex-shrink-0">
                  {formatDate(msg.date)}
                </small>
              </div>
            </button>
          );
        })}
      </div>

      {/* Pagination */}
      {(hasPrev || hasNext) && (
        <div className="d-flex justify-content-between align-items-center px-3 py-2 border-top">
          <small className="text-body-secondary">
            {offset + 1}&ndash;{Math.min(offset + limit, totalCount)} of{" "}
            {totalCount}
          </small>
          <div className="btn-group btn-group-sm">
            <button
              className="btn btn-outline-secondary"
              disabled={!hasPrev}
              onClick={() => onPageChange(Math.max(0, offset - limit))}
            >
              &laquo; Prev
            </button>
            <button
              className="btn btn-outline-secondary"
              disabled={!hasNext}
              onClick={() => onPageChange(offset + limit)}
            >
              Next &raquo;
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

function formatFrom(msg: MessageView): string {
  if (!msg.from || msg.from.length === 0) return "(unknown sender)";
  const f = msg.from[0];
  if (!f) return "(unknown sender)";
  return f.name || f.address;
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return "";
  try {
    const d = new Date(dateStr);
    const now = new Date();
    if (d.toDateString() === now.toDateString()) {
      return d.toLocaleTimeString(undefined, {
        hour: "2-digit",
        minute: "2-digit",
      });
    }
    return d.toLocaleDateString(undefined, {
      month: "short",
      day: "numeric",
    });
  } catch {
    return dateStr;
  }
}
