import { useState } from "react";
import type { AccountFormDraft } from "./accountsSlice";
import { WIDGET, parts } from "./parts";

interface AccountFormProps {
  draft: AccountFormDraft;
  isEditing: boolean;
  saving: boolean;
  error: { code: string; message: string } | null;
  onFieldChange: (patch: Partial<AccountFormDraft>) => void;
  onSubmit: () => void;
  onCancel: () => void;
}

export function AccountForm({
  draft,
  isEditing,
  saving,
  error,
  onFieldChange,
  onSubmit,
  onCancel,
}: AccountFormProps) {
  const [showAdvanced, setShowAdvanced] = useState(draft.insecure || draft.isDefault);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    onSubmit();
  }

  return (
    <form
      data-widget={WIDGET}
      data-part={parts.form}
      onSubmit={handleSubmit}
      className="mx-auto"
      style={{ maxWidth: 540 }}
    >
      <div data-part={parts.formHeader} className="mb-4">
        <h2 className="h4 mb-1">
          {isEditing ? "Edit IMAP account" : "Add IMAP account"}
        </h2>
        <p className="text-body-secondary small mb-0">
          Use the mailbox credentials you already use in your mail client.
          We store the password encrypted and run a read-only connection test.
        </p>
      </div>

      {error && (
        <div className="alert alert-danger" role="alert">
          <strong>{error.code}:</strong> {error.message}
        </div>
      )}

      <div data-part={parts.formBody}>
        <div data-part={parts.fieldGroup} className="mb-3">
          <label htmlFor="field-label" className="form-label">Label</label>
          <input
            id="field-label"
            type="text"
            className="form-control"
            placeholder="Work inbox"
            value={draft.label}
            onChange={(e) => onFieldChange({ label: e.target.value })}
          />
        </div>

        <div data-part={parts.fieldGroup} className="mb-3">
          <label htmlFor="field-username" className="form-label">Username or email</label>
          <input
            id="field-username"
            type="text"
            className="form-control"
            value={draft.username}
            onChange={(e) => onFieldChange({ username: e.target.value })}
          />
        </div>

        <div className="row mb-3">
          <div className="col-8">
            <div data-part={parts.fieldGroup}>
              <label htmlFor="field-server" className="form-label">IMAP server</label>
              <input
                id="field-server"
                type="text"
                className="form-control"
                placeholder="imap.gmail.com"
                value={draft.server}
                onChange={(e) => onFieldChange({ server: e.target.value })}
              />
              <div className="form-text">Example: imap.gmail.com or mail.example.com</div>
            </div>
          </div>
          <div className="col-4">
            <div data-part={parts.fieldGroup}>
              <label htmlFor="field-port" className="form-label">Port</label>
              <input
                id="field-port"
                type="number"
                className="form-control"
                value={draft.port}
                onChange={(e) => onFieldChange({ port: parseInt(e.target.value, 10) || 993 })}
              />
            </div>
          </div>
        </div>

        <div data-part={parts.fieldGroup} className="mb-3">
          <label htmlFor="field-password" className="form-label">Password or app password</label>
          <input
            id="field-password"
            type="password"
            className="form-control"
            value={draft.password}
            onChange={(e) => onFieldChange({ password: e.target.value })}
          />
        </div>

        <div data-part={parts.fieldGroup} className="mb-3">
          <label htmlFor="field-mailbox" className="form-label">Default mailbox</label>
          <input
            id="field-mailbox"
            type="text"
            className="form-control"
            value={draft.mailboxDefault}
            onChange={(e) => onFieldChange({ mailboxDefault: e.target.value })}
          />
        </div>

        <div data-part={parts.advancedToggle} className="mb-3">
          <button
            type="button"
            className="btn btn-link text-decoration-none p-0 text-body-secondary"
            onClick={() => setShowAdvanced((v) => !v)}
            aria-expanded={showAdvanced}
          >
            {showAdvanced ? "\u25BE" : "\u25B8"} Advanced
          </button>
        </div>

        {showAdvanced && (
          <div data-part={parts.advancedPanel} className="mb-3 ps-3 border-start">
            <div className="form-check mb-2">
              <input
                id="field-insecure"
                type="checkbox"
                className="form-check-input"
                checked={draft.insecure}
                onChange={(e) => onFieldChange({ insecure: e.target.checked })}
              />
              <label htmlFor="field-insecure" className="form-check-label">
                Skip TLS verification
              </label>
            </div>
            <div className="form-check">
              <input
                id="field-default"
                type="checkbox"
                className="form-check-input"
                checked={draft.isDefault}
                onChange={(e) => onFieldChange({ isDefault: e.target.checked })}
              />
              <label htmlFor="field-default" className="form-check-label">
                Set as default account
              </label>
            </div>
          </div>
        )}
      </div>

      <div data-part={parts.formActions} className="d-flex justify-content-between mt-4">
        <button type="button" className="btn btn-outline-secondary" onClick={onCancel}>
          Cancel
        </button>
        <button type="submit" className="btn btn-primary" disabled={saving}>
          {saving ? (
            <>
              <span className="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true" />
              Saving…
            </>
          ) : (
            "Save and test"
          )}
        </button>
      </div>
    </form>
  );
}
