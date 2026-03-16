package imapjs

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	hostedapp "github.com/go-go-golems/smailnail/pkg/smailnaild"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/identity"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
)

const (
	appDBDriverFlag = "app-db-driver"
	appDBDSNFlag    = "app-db-dsn"
)

type sharedIdentityRuntime struct {
	mu              sync.RWMutex
	db              *sqlx.DB
	identityRepo    *identity.Repository
	identityService *identity.Service
}

func newSharedIdentityRuntime() *sharedIdentityRuntime {
	return &sharedIdentityRuntime{}
}

func newSharedIdentityRuntimeWithDB(db *sqlx.DB) *sharedIdentityRuntime {
	repo := identity.NewRepository(db)
	return &sharedIdentityRuntime{
		db:              db,
		identityRepo:    repo,
		identityService: identity.NewService(repo),
	}
}

func (r *sharedIdentityRuntime) commandCustomizer(cmd *cobra.Command) error {
	cmd.Flags().String(appDBDriverFlag, "sqlite3", "Driver for the shared smailnail application database")
	cmd.Flags().String(appDBDSNFlag, "", "DSN for the shared smailnail application database used for user/account ownership")
	return nil
}

func (r *sharedIdentityRuntime) startupHook(ctx context.Context) error {
	flags, _ := ctx.Value(embeddable.CommandFlagsKey).(map[string]interface{})
	dsn := strings.TrimSpace(flagString(flags, appDBDSNFlag))
	if dsn == "" {
		return nil
	}

	driver := strings.TrimSpace(flagString(flags, appDBDriverFlag))
	if driver == "" {
		driver = "sqlite3"
	}

	db, err := sqlx.Open(driver, dsn)
	if err != nil {
		return fmt.Errorf("open shared app db: %w", err)
	}
	if err := hostedapp.BootstrapApplicationDB(ctx, db); err != nil {
		_ = db.Close()
		return fmt.Errorf("bootstrap shared app db: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.db = db
	r.identityRepo = identity.NewRepository(db)
	r.identityService = identity.NewService(r.identityRepo)
	return nil
}

func (r *sharedIdentityRuntime) middleware() embeddable.ToolMiddleware {
	return func(next embeddable.ToolHandler) embeddable.ToolHandler {
		return func(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
			principal, ok := embeddable.GetAuthPrincipal(ctx)
			if !ok || strings.TrimSpace(principal.Issuer) == "" || strings.TrimSpace(principal.Subject) == "" {
				return next(ctx, args)
			}

			svc, ok := r.identityServiceFromRuntime()
			if !ok {
				return newErrorToolResult("shared app db is not configured for MCP identity resolution", nil), nil
			}

			resolved, err := svc.ResolveOrProvisionUser(ctx, identity.ExternalPrincipal{
				Issuer:            principal.Issuer,
				Subject:           principal.Subject,
				ProviderKind:      identity.ProviderKindOIDC,
				ClientID:          principal.ClientID,
				Email:             principal.Email,
				EmailVerified:     principal.EmailVerified,
				PreferredUsername: principal.PreferredUsername,
				DisplayName:       principal.DisplayName,
				AvatarURL:         principal.AvatarURL,
				Scopes:            append([]string(nil), principal.Scopes...),
				Claims:            effectiveClaims(principal),
			})
			if err != nil {
				return newErrorToolResult("failed to resolve MCP principal to local user", err), nil
			}

			return next(withResolvedIdentity(ctx, resolved), args)
		}
	}
}

func (r *sharedIdentityRuntime) identityServiceFromRuntime() (*identity.Service, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.identityService, r.identityService != nil
}

func flagString(flags map[string]interface{}, key string) string {
	if len(flags) == 0 {
		return ""
	}
	value, ok := flags[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return fmt.Sprint(typed)
	}
}

func effectiveClaims(principal embeddable.AuthPrincipal) map[string]any {
	if len(principal.Claims) > 0 {
		return principal.Claims
	}
	return map[string]any{
		"email":              principal.Email,
		"email_verified":     principal.EmailVerified,
		"preferred_username": principal.PreferredUsername,
		"name":               principal.DisplayName,
		"picture":            principal.AvatarURL,
	}
}
