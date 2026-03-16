package identity

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("identity not found")

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, user *User) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	_, err := r.db.NamedExecContext(ctx, `INSERT INTO users (
		id,
		primary_email,
		display_name,
		avatar_url
	) VALUES (
		:id,
		:primary_email,
		:display_name,
		:avatar_url
	)`, user)
	return err
}

func (r *Repository) GetUserByID(ctx context.Context, userID string) (*User, error) {
	var user User
	err := r.db.GetContext(ctx, &user, r.db.Rebind(`SELECT
		id,
		primary_email,
		display_name,
		avatar_url,
		created_at,
		updated_at
	FROM users
	WHERE id = ?`), userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) UpdateUserProfile(ctx context.Context, user *User) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	result, err := r.db.NamedExecContext(ctx, `UPDATE users SET
		primary_email = :primary_email,
		display_name = :display_name,
		avatar_url = :avatar_url,
		updated_at = CURRENT_TIMESTAMP
	WHERE id = :id`, user)
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

func (r *Repository) CreateExternalIdentity(ctx context.Context, identity *ExternalIdentity) error {
	if identity == nil {
		return fmt.Errorf("external identity is nil")
	}

	_, err := r.db.NamedExecContext(ctx, `INSERT INTO user_external_identities (
		id,
		user_id,
		issuer,
		subject,
		provider_kind,
		email,
		email_verified,
		preferred_username,
		raw_claims_json
	) VALUES (
		:id,
		:user_id,
		:issuer,
		:subject,
		:provider_kind,
		:email,
		:email_verified,
		:preferred_username,
		:raw_claims_json
	)`, identity)
	return err
}

func (r *Repository) UpdateExternalIdentity(ctx context.Context, identity *ExternalIdentity) error {
	if identity == nil {
		return fmt.Errorf("external identity is nil")
	}

	result, err := r.db.NamedExecContext(ctx, `UPDATE user_external_identities SET
		email = :email,
		email_verified = :email_verified,
		preferred_username = :preferred_username,
		raw_claims_json = :raw_claims_json,
		updated_at = CURRENT_TIMESTAMP
	WHERE id = :id`, identity)
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

func (r *Repository) GetResolvedByIssuerSubject(ctx context.Context, issuer, subject string) (*ResolvedIdentity, error) {
	type row struct {
		UserID                    string       `db:"user_id"`
		UserPrimaryEmail          string       `db:"user_primary_email"`
		UserDisplayName           string       `db:"user_display_name"`
		UserAvatarURL             string       `db:"user_avatar_url"`
		UserCreatedAt             sql.NullTime `db:"user_created_at"`
		UserUpdatedAt             sql.NullTime `db:"user_updated_at"`
		IdentityID                string       `db:"identity_id"`
		IdentityProviderKind      string       `db:"identity_provider_kind"`
		IdentityEmail             string       `db:"identity_email"`
		IdentityEmailVerified     bool         `db:"identity_email_verified"`
		IdentityPreferredUsername string       `db:"identity_preferred_username"`
		IdentityRawClaimsJSON     string       `db:"identity_raw_claims_json"`
		IdentityCreatedAt         sql.NullTime `db:"identity_created_at"`
		IdentityUpdatedAt         sql.NullTime `db:"identity_updated_at"`
	}

	var result row
	err := r.db.GetContext(ctx, &result, r.db.Rebind(`SELECT
		u.id AS user_id,
		u.primary_email AS user_primary_email,
		u.display_name AS user_display_name,
		u.avatar_url AS user_avatar_url,
		u.created_at AS user_created_at,
		u.updated_at AS user_updated_at,
		e.id AS identity_id,
		e.provider_kind AS identity_provider_kind,
		e.email AS identity_email,
		e.email_verified AS identity_email_verified,
		e.preferred_username AS identity_preferred_username,
		e.raw_claims_json AS identity_raw_claims_json,
		e.created_at AS identity_created_at,
		e.updated_at AS identity_updated_at
	FROM user_external_identities e
	JOIN users u ON u.id = e.user_id
	WHERE e.issuer = ? AND e.subject = ?`), issuer, subject)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &ResolvedIdentity{
		User: &User{
			ID:           result.UserID,
			PrimaryEmail: result.UserPrimaryEmail,
			DisplayName:  result.UserDisplayName,
			AvatarURL:    result.UserAvatarURL,
			CreatedAt:    nullTime(result.UserCreatedAt),
			UpdatedAt:    nullTime(result.UserUpdatedAt),
		},
		ExternalIdentity: &ExternalIdentity{
			ID:                result.IdentityID,
			UserID:            result.UserID,
			Issuer:            issuer,
			Subject:           subject,
			ProviderKind:      result.IdentityProviderKind,
			Email:             result.IdentityEmail,
			EmailVerified:     result.IdentityEmailVerified,
			PreferredUsername: result.IdentityPreferredUsername,
			RawClaimsJSON:     result.IdentityRawClaimsJSON,
			CreatedAt:         nullTime(result.IdentityCreatedAt),
			UpdatedAt:         nullTime(result.IdentityUpdatedAt),
		},
	}, nil
}

func (r *Repository) CreateSession(ctx context.Context, session *WebSession) error {
	if session == nil {
		return fmt.Errorf("session is nil")
	}

	_, err := r.db.NamedExecContext(ctx, `INSERT INTO web_sessions (
		id,
		user_id,
		issuer,
		subject,
		expires_at,
		created_at,
		last_seen_at
	) VALUES (
		:id,
		:user_id,
		:issuer,
		:subject,
		:expires_at,
		:created_at,
		:last_seen_at
	)`, session)
	return err
}

func (r *Repository) GetSessionByID(ctx context.Context, sessionID string) (*WebSession, error) {
	var session WebSession
	err := r.db.GetContext(ctx, &session, r.db.Rebind(`SELECT
		id,
		user_id,
		issuer,
		subject,
		expires_at,
		created_at,
		last_seen_at
	FROM web_sessions
	WHERE id = ?`), sessionID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *Repository) DeleteSession(ctx context.Context, sessionID string) error {
	result, err := r.db.ExecContext(ctx, r.db.Rebind(`DELETE FROM web_sessions WHERE id = ?`), sessionID)
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

func nullTime(value sql.NullTime) time.Time {
	if value.Valid {
		return value.Time
	}
	return time.Time{}
}
