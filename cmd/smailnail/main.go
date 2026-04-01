package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/go-go-golems/smailnail/cmd/smailnail/commands"
	smailnaildocs "github.com/go-go-golems/smailnail/cmd/smailnail/docs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func main() {
	// Setup logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	rootCmd := &cobra.Command{
		Use:   "smailnail",
		Short: "Process mail rules on an IMAP server",
	}

	helpSystem := help.NewHelpSystem()
	if err := smailnaildocs.AddDocToHelpSystem(helpSystem); err != nil {
		fmt.Printf("Error loading help docs: %v\n", err)
		os.Exit(1)
	}
	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

	log.Debug().Msg("Starting smailnail")

	// Create and add the mail-rules command
	mailRulesCmd, err := commands.NewMailRulesCommand()
	if err != nil {
		fmt.Printf("Error creating mail rules command: %v\n", err)
		os.Exit(1)
	}

	cobraMailRulesCmd, err := cli.BuildCobraCommandFromCommand(mailRulesCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			AppName: "smailnail",
		}),
	)
	if err != nil {
		fmt.Printf("Error building Cobra command: %v\n", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(cobraMailRulesCmd)

	// Create and add the fetch-mail command
	fetchMailCmd, err := commands.NewFetchMailCommand()
	if err != nil {
		fmt.Printf("Error creating fetch mail command: %v\n", err)
		os.Exit(1)
	}

	cobraFetchMailCmd, err := cli.BuildCobraCommandFromCommand(fetchMailCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			AppName: "smailnail",
		}),
	)
	if err != nil {
		fmt.Printf("Error building Cobra command: %v\n", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(cobraFetchMailCmd)

	mirrorCmd, err := commands.NewMirrorCommand()
	if err != nil {
		fmt.Printf("Error creating mirror command: %v\n", err)
		os.Exit(1)
	}

	cobraMirrorCmd, err := cli.BuildCobraCommandFromCommand(mirrorCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			AppName: "smailnail",
		}),
	)
	if err != nil {
		fmt.Printf("Error building Cobra command: %v\n", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(cobraMirrorCmd)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Setup context with cancellation
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}
