package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/go-go-golems/smailnail/cmd/smailnail/commands"
	annotatecommands "github.com/go-go-golems/smailnail/cmd/smailnail/commands/annotate"
	enrichcommands "github.com/go-go-golems/smailnail/cmd/smailnail/commands/enrich"
	smailnaildocs "github.com/go-go-golems/smailnail/cmd/smailnail/docs"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "smailnail",
		Short: "Process mail rules on an IMAP server",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return logging.InitLoggerFromCobra(cmd)
		},
	}
	if err := logging.AddLoggingSectionToRootCommand(rootCmd, "smailnail"); err != nil {
		fmt.Printf("Error adding logging flags: %v\n", err)
		os.Exit(1)
	}

	helpSystem := help.NewHelpSystem()
	if err := smailnaildocs.AddDocToHelpSystem(helpSystem); err != nil {
		fmt.Printf("Error loading help docs: %v\n", err)
		os.Exit(1)
	}
	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

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

	mergeMirrorCmd, err := commands.NewMergeMirrorShardsCommand()
	if err != nil {
		fmt.Printf("Error creating merge mirror shards command: %v\n", err)
		os.Exit(1)
	}

	cobraMergeMirrorCmd, err := cli.BuildCobraCommandFromCommand(mergeMirrorCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			AppName: "smailnail",
		}),
	)
	if err != nil {
		fmt.Printf("Error building merge mirror shards Cobra command: %v\n", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(cobraMergeMirrorCmd)

	enrichCmd, err := enrichcommands.NewEnrichCommand()
	if err != nil {
		fmt.Printf("Error creating enrich command group: %v\n", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(enrichCmd)

	annotateCmd, err := annotatecommands.NewAnnotateCommand()
	if err != nil {
		fmt.Printf("Error creating annotate command group: %v\n", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(annotateCmd)

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
