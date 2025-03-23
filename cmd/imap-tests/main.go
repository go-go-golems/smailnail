package main

import (
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/smailnail/cmd/imap-tests/commands"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func main() {
	// Configure logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	rootCmd := &cobra.Command{
		Use:   "imap-tests",
		Short: "Test commands for IMAP operations",
		Long:  "A collection of commands to test IMAP operations using the go-imap and go-message libraries",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Set debug level if verbose flag is set
			if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
		},
	}

	// Add global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable debug logging")

	// Initialize help system
	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	// Initialize commands
	createMailboxCmd, err := commands.NewCreateMailboxCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create createMailbox command")
	}

	storeTextMessageCmd, err := commands.NewStoreTextMessageCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create storeTextMessage command")
	}

	storeHTMLMessageCmd, err := commands.NewStoreHTMLMessageCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create storeHTMLMessage command")
	}

	storeAttachmentCmd, err := commands.NewStoreAttachmentCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create storeAttachment command")
	}

	// Convert glazed commands to cobra commands
	createMailboxCobraCmd, err := cli.BuildCobraCommandFromGlazeCommand(createMailboxCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build createMailbox cobra command")
	}

	storeTextMessageCobraCmd, err := cli.BuildCobraCommandFromGlazeCommand(storeTextMessageCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build storeTextMessage cobra command")
	}

	storeHTMLMessageCobraCmd, err := cli.BuildCobraCommandFromGlazeCommand(storeHTMLMessageCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build storeHTMLMessage cobra command")
	}

	storeAttachmentCobraCmd, err := cli.BuildCobraCommandFromGlazeCommand(storeAttachmentCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build storeAttachment cobra command")
	}

	// Add commands to root
	rootCmd.AddCommand(
		createMailboxCobraCmd,
		storeTextMessageCobraCmd,
		storeHTMLMessageCobraCmd,
		storeAttachmentCobraCmd,
	)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute root command")
	}
}
