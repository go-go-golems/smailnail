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
			embeddable.WithDescription("Documentation query tool placeholder; implementation added in a later slice"),
		),
	)
}
