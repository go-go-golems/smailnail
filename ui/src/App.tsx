import { AccountSetupPage } from "./features/accounts";

export function App() {
  return (
    <div data-widget="app-shell">
      <header className="mb-4 pb-3 border-bottom">
        <h1 className="h5 mb-0 text-body-secondary">smailnail</h1>
      </header>
      <main>
        <AccountSetupPage />
      </main>
    </div>
  );
}
