interface LoggedOutShellProps {
  onRetry: () => void;
  error?: string | null;
}

export function LoggedOutShell({ onRetry, error }: LoggedOutShellProps) {
  return (
    <section
      data-widget="auth-shell"
      data-part="logged-out"
      className="rounded-4 border p-4 p-md-5"
    >
      <div className="small text-uppercase text-body-secondary mb-3">
        Hosted account setup
      </div>
      <h1 className="display-6 mb-3">Sign in to manage your IMAP accounts.</h1>
      <p className="text-body-secondary mb-4" style={{ maxWidth: 560 }}>
        smailnail stores mailbox credentials against your authenticated identity.
        Start with one IMAP account and a lightweight read-only connection test.
      </p>
      {error && (
        <div className="alert alert-danger" role="alert">
          {error}
        </div>
      )}
      <div className="d-flex flex-wrap gap-2">
        <a className="btn btn-primary" href="/auth/login">
          Sign in with your identity provider
        </a>
        <button className="btn btn-outline-secondary" onClick={onRetry}>
          Retry
        </button>
      </div>
      <div className="mt-4 small text-body-secondary">
        If you expected to already be signed in, retry first. If the session has
        expired, the login button will start the server-side OIDC flow again.
      </div>
    </section>
  );
}
