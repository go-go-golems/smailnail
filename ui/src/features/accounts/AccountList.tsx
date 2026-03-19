import { useState } from "react";
import type { AccountListItem } from "../../api/types";
import { WIDGET } from "./parts";

interface AccountListProps {
  accounts: AccountListItem[];
  onAdd: () => void;
  onEdit: (account: AccountListItem) => void;
  onExplore: (account: AccountListItem) => void;
  onDelete: (account: AccountListItem) => void;
}

export function AccountList({
  accounts,
  onAdd,
  onEdit,
  onExplore,
  onDelete,
}: AccountListProps) {
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);

  return (
    <div data-widget={WIDGET} data-part="account-list">
      <div className="d-flex justify-content-between align-items-center mb-3">
        <h2 className="h4 mb-0">Accounts</h2>
        <button className="btn btn-primary btn-sm" onClick={onAdd}>
          Add account
        </button>
      </div>

      <div className="list-group">
        {accounts.map((account) => (
          <div
            key={account.id}
            className="list-group-item"
          >
            {confirmDeleteId === account.id ? (
              <div className="d-flex justify-content-between align-items-center">
                <span className="text-danger small">
                  Delete <strong>{account.label || account.server}</strong>? This cannot be undone.
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
                      onDelete(account);
                    }}
                  >
                    Delete
                  </button>
                </div>
              </div>
            ) : (
              <div className="d-flex justify-content-between align-items-start">
                <div
                  className="flex-grow-1 cursor-pointer"
                  role="button"
                  tabIndex={0}
                  onClick={() => onExplore(account)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") onExplore(account);
                  }}
                >
                  <div className="fw-medium">{account.label || account.server}</div>
                  <small className="text-body-secondary">
                    {account.username}@{account.server}:{account.port}
                  </small>
                </div>
                <div className="d-flex align-items-center gap-2">
                  {account.isDefault && (
                    <span className="badge bg-secondary-subtle text-secondary-emphasis">
                      default
                    </span>
                  )}
                  <TestBadge account={account} />
                  <div className="btn-group btn-group-sm">
                    <button
                      className="btn btn-outline-secondary"
                      title="Edit"
                      onClick={(e) => {
                        e.stopPropagation();
                        onEdit(account);
                      }}
                    >
                      Edit
                    </button>
                    <button
                      className="btn btn-outline-danger"
                      title="Delete"
                      onClick={(e) => {
                        e.stopPropagation();
                        setConfirmDeleteId(account.id);
                      }}
                    >
                      &times;
                    </button>
                  </div>
                </div>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

function TestBadge({ account }: { account: AccountListItem }) {
  if (!account.latestTest) {
    return (
      <span className="badge bg-body-secondary text-body-secondary">
        untested
      </span>
    );
  }
  if (account.latestTest.success && account.latestTest.warningCode) {
    return (
      <span className="badge bg-warning-subtle text-warning-emphasis">
        warning
      </span>
    );
  }
  if (account.latestTest.success) {
    return (
      <span className="badge bg-success-subtle text-success-emphasis">
        connected
      </span>
    );
  }
  return (
    <span className="badge bg-danger-subtle text-danger-emphasis">failed</span>
  );
}
