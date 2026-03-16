import { useState } from "react";
import type { AccountListItem } from "./api/types";
import { AccountSetupPage } from "./features/accounts";
import { MailboxExplorer } from "./features/mailbox";

type AppView =
  | { kind: "accounts" }
  | { kind: "explore"; account: AccountListItem };

export function App() {
  const [view, setView] = useState<AppView>({ kind: "accounts" });

  return (
    <div data-widget="app-shell">
      <header className="mb-4 pb-3 border-bottom">
        <h1 className="h5 mb-0 text-body-secondary">smailnail</h1>
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
      </main>
    </div>
  );
}
