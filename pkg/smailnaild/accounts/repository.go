package accounts

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("account not found")

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CountByUser(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, r.db.Rebind(`SELECT COUNT(*) FROM imap_accounts WHERE user_id = ?`), userID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) Create(ctx context.Context, account *Account) error {
	if account == nil {
		return fmt.Errorf("account is nil")
	}

	_, err := sqlx.NamedExecContext(ctx, r.db, `INSERT INTO imap_accounts (
		id,
		user_id,
		label,
		provider_hint,
		server,
		port,
		username,
		mailbox_default,
		insecure,
		auth_kind,
		secret_ciphertext,
		secret_nonce,
		secret_key_id,
		is_default,
		mcp_enabled
	) VALUES (
		:id,
		:user_id,
		:label,
		:provider_hint,
		:server,
		:port,
		:username,
		:mailbox_default,
		:insecure,
		:auth_kind,
		:secret_ciphertext,
		:secret_nonce,
		:secret_key_id,
		:is_default,
		:mcp_enabled
	)`, account)
	return err
}

func (r *Repository) ListByUser(ctx context.Context, userID string) ([]Account, error) {
	var ret []Account
	err := r.db.SelectContext(ctx, &ret, r.db.Rebind(`SELECT
		id,
		user_id,
		label,
		provider_hint,
		server,
		port,
		username,
		mailbox_default,
		insecure,
		auth_kind,
		secret_ciphertext,
		secret_nonce,
		secret_key_id,
		is_default,
		mcp_enabled,
		created_at,
		updated_at
	FROM imap_accounts
	WHERE user_id = ?
	ORDER BY is_default DESC, created_at ASC`), userID)
	return ret, err
}

func (r *Repository) GetByID(ctx context.Context, userID, accountID string) (*Account, error) {
	var account Account
	err := r.db.GetContext(ctx, &account, r.db.Rebind(`SELECT
		id,
		user_id,
		label,
		provider_hint,
		server,
		port,
		username,
		mailbox_default,
		insecure,
		auth_kind,
		secret_ciphertext,
		secret_nonce,
		secret_key_id,
		is_default,
		mcp_enabled,
		created_at,
		updated_at
	FROM imap_accounts
	WHERE id = ? AND user_id = ?`), accountID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *Repository) Update(ctx context.Context, account *Account) error {
	if account == nil {
		return fmt.Errorf("account is nil")
	}

	result, err := sqlx.NamedExecContext(ctx, r.db, `UPDATE imap_accounts SET
		label = :label,
		provider_hint = :provider_hint,
		server = :server,
		port = :port,
		username = :username,
		mailbox_default = :mailbox_default,
		insecure = :insecure,
		auth_kind = :auth_kind,
		secret_ciphertext = :secret_ciphertext,
		secret_nonce = :secret_nonce,
		secret_key_id = :secret_key_id,
		is_default = :is_default,
		mcp_enabled = :mcp_enabled,
		updated_at = CURRENT_TIMESTAMP
	WHERE id = :id AND user_id = :user_id`, account)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, userID, accountID string) error {
	return r.withTx(ctx, func(tx *sqlx.Tx) error {
		if _, err := tx.ExecContext(ctx, r.db.Rebind(`DELETE FROM imap_account_tests
WHERE imap_account_id IN (
	SELECT id FROM imap_accounts WHERE id = ? AND user_id = ?
)`), accountID, userID); err != nil {
			return err
		}

		result, err := tx.ExecContext(ctx, r.db.Rebind(`DELETE FROM imap_accounts WHERE id = ? AND user_id = ?`), accountID, userID)
		if err != nil {
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func (r *Repository) ClearDefaultForUser(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, r.db.Rebind(`UPDATE imap_accounts SET is_default = FALSE, updated_at = CURRENT_TIMESTAMP WHERE user_id = ?`), userID)
	return err
}

func (r *Repository) CreateAtomic(ctx context.Context, account *Account, clearExistingDefault bool) error {
	return r.withTx(ctx, func(tx *sqlx.Tx) error {
		if clearExistingDefault {
			if _, err := tx.ExecContext(ctx, r.db.Rebind(`UPDATE imap_accounts SET is_default = FALSE, updated_at = CURRENT_TIMESTAMP WHERE user_id = ?`), account.UserID); err != nil {
				return err
			}
		}
		_, err := sqlx.NamedExecContext(ctx, tx, `INSERT INTO imap_accounts (
			id,
			user_id,
			label,
			provider_hint,
			server,
			port,
			username,
			mailbox_default,
			insecure,
			auth_kind,
			secret_ciphertext,
			secret_nonce,
			secret_key_id,
			is_default,
			mcp_enabled
		) VALUES (
			:id,
			:user_id,
			:label,
			:provider_hint,
			:server,
			:port,
			:username,
			:mailbox_default,
			:insecure,
			:auth_kind,
			:secret_ciphertext,
			:secret_nonce,
			:secret_key_id,
			:is_default,
			:mcp_enabled
		)`, account)
		return err
	})
}

func (r *Repository) UpdateAtomic(ctx context.Context, account *Account, clearExistingDefault bool) error {
	return r.withTx(ctx, func(tx *sqlx.Tx) error {
		if clearExistingDefault {
			if _, err := tx.ExecContext(ctx, r.db.Rebind(`UPDATE imap_accounts
SET is_default = FALSE, updated_at = CURRENT_TIMESTAMP
WHERE user_id = ? AND id <> ?`), account.UserID, account.ID); err != nil {
				return err
			}
		}

		result, err := sqlx.NamedExecContext(ctx, tx, `UPDATE imap_accounts SET
			label = :label,
			provider_hint = :provider_hint,
			server = :server,
			port = :port,
			username = :username,
			mailbox_default = :mailbox_default,
			insecure = :insecure,
			auth_kind = :auth_kind,
			secret_ciphertext = :secret_ciphertext,
			secret_nonce = :secret_nonce,
			secret_key_id = :secret_key_id,
			is_default = :is_default,
			mcp_enabled = :mcp_enabled,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = :id AND user_id = :user_id`, account)
		if err != nil {
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func (r *Repository) withTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *Repository) CreateTest(ctx context.Context, test *AccountTest) error {
	if test == nil {
		return fmt.Errorf("account test is nil")
	}

	_, err := r.db.NamedExecContext(ctx, `INSERT INTO imap_account_tests (
		id,
		imap_account_id,
		test_mode,
		success,
		tcp_ok,
		login_ok,
		mailbox_select_ok,
		list_ok,
		sample_fetch_ok,
		write_probe_ok,
		warning_code,
		error_code,
		error_message,
		details_json
	) VALUES (
		:id,
		:imap_account_id,
		:test_mode,
		:success,
		:tcp_ok,
		:login_ok,
		:mailbox_select_ok,
		:list_ok,
		:sample_fetch_ok,
		:write_probe_ok,
		:warning_code,
		:error_code,
		:error_message,
		:details_json
	)`, test)
	return err
}

func (r *Repository) LatestTestByAccount(ctx context.Context, accountID string) (*AccountTest, error) {
	var test AccountTest
	err := r.db.GetContext(ctx, &test, r.db.Rebind(`SELECT
		id,
		imap_account_id,
		test_mode,
		success,
		tcp_ok,
		login_ok,
		mailbox_select_ok,
		list_ok,
		sample_fetch_ok,
		write_probe_ok,
		warning_code,
		error_code,
		error_message,
		details_json,
		created_at
	FROM imap_account_tests
	WHERE imap_account_id = ?
	ORDER BY created_at DESC
	LIMIT 1`), accountID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &test, nil
}
