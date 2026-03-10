package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/smailnail/pkg/mcp/imapjs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	rootCmd := &cobra.Command{
		Use:   "smailnail-imap-mcp",
		Short: "MCP server for the smailnail JavaScript IMAP runtime",
	}

	if err := imapjs.AddMCPCommand(rootCmd); err != nil {
		fmt.Fprintf(os.Stderr, "Error adding MCP command: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}
