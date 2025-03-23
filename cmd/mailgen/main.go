package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/smailnail/cmd/mailgen/cmds"
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

	// Create and register the generate command
	generateCommand, err := cmds.NewGenerateCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building generate command: %v\n", err)
		os.Exit(1)
	}
	generateCmd, err := cli.BuildCobraCommandFromCommand(generateCommand)
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
