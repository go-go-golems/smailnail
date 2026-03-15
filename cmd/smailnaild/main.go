package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"

	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	helpcmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/go-go-golems/smailnail/cmd/smailnaild/commands"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "smailnaild",
		Short: "Hosted smailnail application skeleton",
		Long:  "smailnaild is the hosted application entrypoint for smailnail.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return logging.InitLoggerFromCobra(cmd)
		},
	}

	helpSystem := help.NewHelpSystem()
	helpcmd.SetupCobraRootCommand(helpSystem, rootCmd)

	if err := clay.InitGlazed("smailnaild", rootCmd); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize root command")
	}

	serveCmd, err := commands.NewServeCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create serve command")
	}

	cobraServeCmd, err := cli.BuildCobraCommandFromCommand(
		serveCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			AppName: "smailnaild",
		}),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build serve command")
	}
	rootCmd.AddCommand(cobraServeCmd)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := rootCmd.ExecuteContext(ctx); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Failed to execute root command")
	}
}
