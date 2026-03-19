package imapjs

import (
	"context"
	"testing"

	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	hostedapp "github.com/go-go-golems/smailnail/pkg/smailnaild"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/identity"
)

func TestIdentityMiddlewareResolvesAuthPrincipalIntoContext(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := hostedapp.BootstrapApplicationDB(context.Background(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	runtime := newSharedIdentityRuntimeWithDB(db)
	middleware := runtime.middleware()

	ctx := embeddable.WithAuthPrincipal(context.Background(), embeddable.AuthPrincipal{
		Issuer:            "https://auth.example.com/realms/smailnail",
		Subject:           "subject-1",
		ClientID:          "smailnail-mcp",
		Email:             "intern@example.com",
		EmailVerified:     true,
		PreferredUsername: "intern",
		DisplayName:       "Intern Example",
		AvatarURL:         "https://example.com/avatar.png",
		Scopes:            []string{"openid", "profile"},
	})

	result, err := middleware(func(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
		resolved, ok := ResolvedIdentityFromContext(ctx)
		if !ok {
			t.Fatalf("expected resolved identity in context")
		}
		if resolved.User.PrimaryEmail != "intern@example.com" {
			t.Fatalf("unexpected resolved user: %+v", resolved.User)
		}
		return newJSONToolResult(map[string]any{"userID": resolved.User.ID})
	})(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("middleware handler error = %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success result, got %#v", result)
	}
}

func TestIdentityMiddlewareMatchesExistingWebProvisionedUser(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := hostedapp.BootstrapApplicationDB(context.Background(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	repo := identity.NewRepository(db)
	svc := identity.NewService(repo)
	existing, err := svc.ResolveOrProvisionUser(context.Background(), identity.ExternalPrincipal{
		Issuer:            "https://auth.example.com/realms/smailnail",
		Subject:           "subject-1",
		ProviderKind:      identity.ProviderKindOIDC,
		ClientID:          "smailnail-web",
		Email:             "intern@example.com",
		EmailVerified:     true,
		PreferredUsername: "intern",
		DisplayName:       "Intern Example",
	})
	if err != nil {
		t.Fatalf("ResolveOrProvisionUser() error = %v", err)
	}

	runtime := newSharedIdentityRuntimeWithDB(db)
	middleware := runtime.middleware()
	ctx := embeddable.WithAuthPrincipal(context.Background(), embeddable.AuthPrincipal{
		Issuer:            "https://auth.example.com/realms/smailnail",
		Subject:           "subject-1",
		ClientID:          "smailnail-mcp",
		Email:             "intern@example.com",
		EmailVerified:     true,
		PreferredUsername: "intern",
		DisplayName:       "Intern Example",
	})

	_, err = middleware(func(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
		resolved, ok := ResolvedIdentityFromContext(ctx)
		if !ok {
			t.Fatalf("expected resolved identity in context")
		}
		if resolved.User.ID != existing.User.ID {
			t.Fatalf("resolved user id = %q, want %q", resolved.User.ID, existing.User.ID)
		}
		return newJSONToolResult(map[string]any{"userID": resolved.User.ID})
	})(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("middleware handler error = %v", err)
	}
}
