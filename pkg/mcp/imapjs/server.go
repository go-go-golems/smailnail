package imapjs

import (
	"net/http"

	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/accounts"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/identity"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
)

func AddMCPCommand(rootCmd *cobra.Command) error {
	identityRuntime := newSharedIdentityRuntime()
	options := append(baseServerOptions(identityRuntime),
		embeddable.WithCommandCustomizer(identityRuntime.commandCustomizer),
		embeddable.WithHooks(&embeddable.Hooks{
			OnServerStart: identityRuntime.startupHook,
		}),
	)

	return embeddable.AddMCPCommand(rootCmd, options...)
}

type MountedOptions struct {
	Transport       string
	Auth            embeddable.AuthOptions
	DB              *sqlx.DB
	IdentityService *identity.Service
	AccountService  *accounts.Service
}

func MountHTTPHandlers(mux *http.ServeMux, options MountedOptions) error {
	identityRuntime := newSharedIdentityRuntimeWithServices(options.DB, options.IdentityService, options.AccountService)
	config := embeddable.NewServerConfig()
	serverOptions := baseServerOptions(identityRuntime)
	if options.Transport != "" {
		serverOptions = append(serverOptions, embeddable.WithDefaultTransport(options.Transport))
	}
	serverOptions = append(serverOptions, embeddable.WithAuth(options.Auth))
	for _, opt := range serverOptions {
		if err := opt(config); err != nil {
			return err
		}
	}
	return embeddable.MountHTTPHandlers(mux, config)
}

func baseServerOptions(identityRuntime *sharedIdentityRuntime) []embeddable.ServerOption {
	return []embeddable.ServerOption{
		embeddable.WithName("smailnail IMAP JS MCP"),
		embeddable.WithVersion("0.1.0"),
		embeddable.WithServerDescription("Execute smailnail JavaScript snippets and query their API documentation"),
		embeddable.WithDefaultTransport("streamable_http"),
		embeddable.WithDefaultPort(3201),
		embeddable.WithMiddleware(identityRuntime.middleware()),
		embeddable.WithTool("executeIMAPJS", executeIMAPJSHandler,
			embeddable.WithDescription("Execute JavaScript against the smailnail go-go-goja module and return a JSON result"),
			embeddable.WithStringArg("code", "JavaScript source to execute", true),
		),
		embeddable.WithTool("getIMAPJSDocumentation", getIMAPJSDocumentationHandler,
			embeddable.WithDescription("Query the structured smailnail JavaScript API documentation by mode, symbol, concept, example, or render output"),
			embeddable.WithStringArg("mode", "Documentation query mode: overview, package, symbol, example, concept, search, or render", false),
			embeddable.WithStringArg("package", "Package name for package or overview mode", false),
			embeddable.WithStringArg("symbol", "Symbol name for symbol mode", false),
			embeddable.WithStringArg("example", "Example id for example mode", false),
			embeddable.WithStringArg("concept", "Concept name for concept mode", false),
			embeddable.WithStringArg("query", "Free-text query for search mode", false),
			embeddable.WithIntArg("limit", "Maximum number of symbols or examples to return", false),
			embeddable.WithBoolArg("includeBody", "Include example source bodies in results", false),
		),
	}
}
