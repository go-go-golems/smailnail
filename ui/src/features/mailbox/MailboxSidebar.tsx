import type { MailboxInfo } from "../../api/types";

interface MailboxSidebarProps {
  mailboxes: MailboxInfo[];
  selected: string | null;
  loading: boolean;
  onSelect: (mailbox: string) => void;
}

export function MailboxSidebar({
  mailboxes,
  selected,
  loading,
  onSelect,
}: MailboxSidebarProps) {
  if (loading) {
    return (
      <div className="text-center py-3">
        <div className="spinner-border spinner-border-sm text-secondary" role="status" />
      </div>
    );
  }

  if (mailboxes.length === 0) {
    return <p className="text-body-secondary small px-2">No mailboxes found.</p>;
  }

  return (
    <div className="list-group list-group-flush">
      {mailboxes.map((mb) => (
        <button
          key={mb.path}
          className={`list-group-item list-group-item-action py-2 px-3 ${
            selected === mb.name ? "active" : ""
          }`}
          onClick={() => onSelect(mb.name)}
        >
          <small>{mb.name}</small>
        </button>
      ))}
    </div>
  );
}
