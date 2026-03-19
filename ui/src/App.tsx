import { useEffect, useState } from "react";
import type { AccountListItem } from "./api/types";
import { AccountSetupPage } from "./features/accounts";
import { fetchCurrentUser, LoggedOutShell } from "./features/auth";
import { MailboxExplorer } from "./features/mailbox";
import { RulesPage } from "./features/rules";
import { useAppDispatch, useAppSelector } from "./store";

type AppView =
  | { kind: "accounts" }
  | { kind: "explore"; account: AccountListItem }
  | { kind: "rules" };

export function App() {
  const dispatch = useAppDispatch();
  const { status, user, error } = useAppSelector((state) => state.auth);
  const [view, setView] = useState<AppView>({ kind: "accounts" });

  useEffect(() => {
    if (status === "idle") {
      void dispatch(fetchCurrentUser());
    }
  }, [dispatch, status]);

  if (status === "idle" || status === "loading") {
    return (
      <div
        data-widget="auth-shell"
        data-part="loading"
        className="text-center py-5"
      >
        <div className="spinner-border text-secondary" role="status">
          <span className="visually-hidden">Loading...</span>
        </div>
      </div>
    );
  }

  if (status === "unauthenticated") {
    return <LoggedOutShell onRetry={() => void dispatch(fetchCurrentUser())} />;
  }

  if (status === "error") {
    return (
      <LoggedOutShell
        onRetry={() => void dispatch(fetchCurrentUser())}
        error={error?.message ?? "Failed to load the current user."}
      />
    );
  }

  const displayName = user?.displayName || user?.primaryEmail || user?.id;

  return (
    <div data-widget="app-shell">
      <header className="d-flex justify-content-between align-items-center mb-4 pb-3 border-bottom">
        <div className="d-flex align-items-center gap-3">
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
        </div>
        <div
          data-widget="auth-shell"
          data-part="user-badge"
          className="d-flex align-items-center gap-2 text-end"
        >
          {user?.avatarUrl ? (
            <img
              src={user.avatarUrl}
              alt=""
              className="rounded-circle border"
              width={32}
              height={32}
            />
          ) : (
            <div
              className="rounded-circle border d-flex align-items-center justify-content-center small text-body-secondary"
              style={{ width: 32, height: 32 }}
            >
              {(displayName || "U").slice(0, 1).toUpperCase()}
            </div>
          )}
          <div>
            <div className="small fw-semibold">{displayName}</div>
            {user?.primaryEmail && user.primaryEmail !== displayName && (
              <div className="small text-body-secondary">{user.primaryEmail}</div>
            )}
          </div>
          <a className="btn btn-outline-secondary btn-sm" href="/auth/logout">
            Log out
          </a>
        </div>
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
