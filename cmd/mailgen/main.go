package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/smailnail/cmd/mailgen/cmds"

	clay "github.com/go-go-golems/clay/pkg"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "mailgen",
		Short: "Generate test emails from YAML templates",
		Long:  "A command-line tool for generating test emails from YAML templates using the Go template engine and Sprig functions.",
	}

	// Initialize help system
	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	err := clay.InitViper("smailnail", rootCmd)
	cobra.CheckErr(err)

	// Create and register the generate command
	generateCommand, err := cmds.NewGenerateCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building generate command: %v\n", err)
		os.Exit(1)
	}
	generateCmd, err := cli.BuildCobraCommandFromCommand(generateCommand,
		cli.WithCobraMiddlewaresFunc(GetCobraCommandSmailnailMiddlewares),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building generate command: %v\n", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(generateCmd)

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func GetCobraCommandSmailnailMiddlewares(
	parsedCommandLayers *layers.ParsedLayers,
	cmd *cobra.Command,
	args []string,
) ([]middlewares.Middleware, error) {
	commandSettings := &cli.CommandSettings{}
	err := parsedCommandLayers.InitializeStruct(cli.CommandSettingsSlug, commandSettings)
	if err != nil {
		return nil, err
	}

	middlewares_ := []middlewares.Middleware{
		middlewares.ParseFromCobraCommand(cmd,
			parameters.WithParseStepSource("cobra"),
		),
		middlewares.GatherArguments(args,
			parameters.WithParseStepSource("arguments"),
		),
	}

	middlewares_ = append(middlewares_,
		// viper and default
		middlewares.GatherFlagsFromViper(parameters.WithParseStepSource("viper")),
		middlewares.SetFromDefaults(parameters.WithParseStepSource("defaults")),
	)

	return middlewares_, nil
}
