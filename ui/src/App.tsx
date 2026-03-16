import { useState } from "react";
import type { AccountListItem } from "./api/types";
import { AccountSetupPage } from "./features/accounts";
import { MailboxExplorer } from "./features/mailbox";
import { RulesPage } from "./features/rules";

type AppView =
  | { kind: "accounts" }
  | { kind: "explore"; account: AccountListItem }
  | { kind: "rules" };

export function App() {
  const [view, setView] = useState<AppView>({ kind: "accounts" });

  return (
    <div data-widget="app-shell">
      <header className="d-flex justify-content-between align-items-center mb-4 pb-3 border-bottom">
        <h1
          className="h5 mb-0 text-body-secondary"
          role="button"
          style={{ cursor: "pointer" }}
          onClick={() => setView({ kind: "accounts" })}
        >
          smailnail
        </h1>
        {view.kind === "accounts" && (
          <button
            className="btn btn-outline-secondary btn-sm"
            onClick={() => setView({ kind: "rules" })}
          >
            Rules
          </button>
        )}
      </header>
      <main>
        {view.kind === "accounts" && (
          <AccountSetupPage
            onExploreAccount={(account) =>
              setView({ kind: "explore", account })
            }
          />
        )}
        {view.kind === "explore" && (
          <MailboxExplorer
            accountId={view.account.id}
            accountLabel={view.account.label || view.account.server}
            onBack={() => setView({ kind: "accounts" })}
          />
        )}
        {view.kind === "rules" && (
          <RulesPage onBack={() => setView({ kind: "accounts" })} />
        )}
      </main>
    </div>
  );
}
