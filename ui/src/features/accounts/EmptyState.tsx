import { WIDGET, parts } from "./parts";

interface EmptyStateProps {
  onAddAccount: () => void;
}

export function EmptyState({ onAddAccount }: EmptyStateProps) {
  return (
    <div data-widget={WIDGET} data-part={parts.emptyState} className="text-center py-5">
      <h2 data-part={parts.emptyHeadline} className="h4 mb-3">
        Connect your first mailbox
      </h2>
      <p data-part={parts.emptyBody} className="text-body-secondary mb-4" style={{ maxWidth: 480, margin: "0 auto" }}>
        Add an IMAP account so smailnail can test the connection and prepare
        mailbox previews. The first test is read-only.
      </p>
      <button className="btn btn-primary" onClick={onAddAccount}>
        Add account
      </button>
    </div>
  );
}
