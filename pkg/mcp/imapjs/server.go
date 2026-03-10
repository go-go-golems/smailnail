package imapjs

import (
	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/spf13/cobra"
)

func AddMCPCommand(rootCmd *cobra.Command) error {
	return embeddable.AddMCPCommand(rootCmd,
		embeddable.WithName("smailnail IMAP JS MCP"),
		embeddable.WithVersion("0.1.0"),
		embeddable.WithServerDescription("Execute smailnail JavaScript snippets and query their API documentation"),
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
	)
}
