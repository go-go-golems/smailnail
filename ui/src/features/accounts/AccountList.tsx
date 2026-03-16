import type { AccountListItem } from "../../api/types";
import { WIDGET } from "./parts";

interface AccountListProps {
  accounts: AccountListItem[];
  onAdd: () => void;
  onEdit: (account: AccountListItem) => void;
}

export function AccountList({ accounts, onAdd, onEdit }: AccountListProps) {
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
          <button
            key={account.id}
            className="list-group-item list-group-item-action d-flex justify-content-between align-items-start"
            onClick={() => onEdit(account)}
          >
            <div>
              <div className="fw-medium">{account.label || account.server}</div>
              <small className="text-body-secondary">
                {account.username}@{account.server}:{account.port}
              </small>
            </div>
            <div className="d-flex align-items-center gap-2">
              {account.isDefault && (
                <span className="badge bg-secondary-subtle text-secondary-emphasis">default</span>
              )}
              <TestBadge account={account} />
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}

function TestBadge({ account }: { account: AccountListItem }) {
  if (!account.latestTest) {
    return <span className="badge bg-body-secondary text-body-secondary">untested</span>;
  }
  if (account.latestTest.success && account.latestTest.warningCode) {
    return <span className="badge bg-warning-subtle text-warning-emphasis">warning</span>;
  }
  if (account.latestTest.success) {
    return <span className="badge bg-success-subtle text-success-emphasis">connected</span>;
  }
  return <span className="badge bg-danger-subtle text-danger-emphasis">failed</span>;
}
