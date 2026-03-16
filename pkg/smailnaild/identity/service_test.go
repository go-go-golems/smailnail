package identity

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	hostedapp "github.com/go-go-golems/smailnail/pkg/smailnaild"
)

func TestResolveOrProvisionUserCreatesUserAndIdentity(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := hostedapp.BootstrapApplicationDB(context.Background(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	repo := NewRepository(db)
	service := NewService(repo)
	service.newID = sequenceIDs("user-1", "identity-1")

	resolved, err := service.ResolveOrProvisionUser(context.Background(), ExternalPrincipal{
		Issuer:            "https://auth.example.com/realms/smailnail",
		Subject:           "abc123",
		Email:             "intern@example.com",
		EmailVerified:     true,
		PreferredUsername: "intern",
		DisplayName:       "New Intern",
		Claims: map[string]any{
			"email": "intern@example.com",
		},
	})
	if err != nil {
		t.Fatalf("ResolveOrProvisionUser() error = %v", err)
	}

	if resolved.User.ID != "user-1" {
		t.Fatalf("user ID = %q", resolved.User.ID)
	}
	if resolved.ExternalIdentity.ID != "identity-1" {
		t.Fatalf("identity ID = %q", resolved.ExternalIdentity.ID)
	}
	if resolved.ExternalIdentity.UserID != resolved.User.ID {
		t.Fatalf("identity user ID = %q, want %q", resolved.ExternalIdentity.UserID, resolved.User.ID)
	}
}

func TestResolveOrProvisionUserIsIdempotentAndRefreshesProfile(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := hostedapp.BootstrapApplicationDB(context.Background(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	repo := NewRepository(db)
	service := NewService(repo)
	service.newID = sequenceIDs("user-1", "identity-1", "unused-user", "unused-identity")

	first, err := service.ResolveOrProvisionUser(context.Background(), ExternalPrincipal{
		Issuer:            "https://auth.example.com/realms/smailnail",
		Subject:           "abc123",
		Email:             "intern@example.com",
		EmailVerified:     true,
		PreferredUsername: "intern",
		DisplayName:       "Original Name",
	})
	if err != nil {
		t.Fatalf("first ResolveOrProvisionUser() error = %v", err)
	}

	second, err := service.ResolveOrProvisionUser(context.Background(), ExternalPrincipal{
		Issuer:            "https://auth.example.com/realms/smailnail",
		Subject:           "abc123",
		Email:             "intern+updated@example.com",
		EmailVerified:     true,
		PreferredUsername: "intern-updated",
		DisplayName:       "Updated Name",
		AvatarURL:         "https://example.com/avatar.png",
		Claims: map[string]any{
			"department": "engineering",
		},
	})
	if err != nil {
		t.Fatalf("second ResolveOrProvisionUser() error = %v", err)
	}

	if first.User.ID != second.User.ID {
		t.Fatalf("user IDs differ: %q vs %q", first.User.ID, second.User.ID)
	}
	if second.User.DisplayName != "Updated Name" {
		t.Fatalf("display name = %q", second.User.DisplayName)
	}
	if second.User.PrimaryEmail != "intern+updated@example.com" {
		t.Fatalf("primary email = %q", second.User.PrimaryEmail)
	}
	if second.ExternalIdentity.PreferredUsername != "intern-updated" {
		t.Fatalf("preferred username = %q", second.ExternalIdentity.PreferredUsername)
	}
}

func TestRepositorySessionRoundTrip(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := hostedapp.BootstrapApplicationDB(context.Background(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	repo := NewRepository(db)
	session := &WebSession{
		ID:         "session-1",
		UserID:     "user-1",
		Issuer:     "https://auth.example.com/realms/smailnail",
		Subject:    "abc123",
		ExpiresAt:  time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC),
		CreatedAt:  time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC),
		LastSeenAt: time.Date(2026, 3, 16, 12, 30, 0, 0, time.UTC),
	}

	if err := repo.CreateSession(context.Background(), session); err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	got, err := repo.GetSessionByID(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("GetSessionByID() error = %v", err)
	}
	if got.UserID != session.UserID {
		t.Fatalf("session user ID = %q", got.UserID)
	}

	if err := repo.DeleteSession(context.Background(), session.ID); err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}
	if _, err := repo.GetSessionByID(context.Background(), session.ID); err == nil {
		t.Fatal("expected GetSessionByID() to fail after delete")
	}
}

func sequenceIDs(values ...string) func() string {
	index := 0
	return func() string {
		if index >= len(values) {
			return "exhausted-id"
		}
		value := values[index]
		index++
		return value
	}
}
