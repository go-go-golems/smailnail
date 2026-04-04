package sqlite

import (
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/spf13/cobra"
)

func NewSQLiteCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "sqlite",
		Short: "Serve and inspect the local mirror sqlite database",
	}

	serveCommand, err := NewServeCommand()
	if err != nil {
		return nil, err
	}
	cobraCommand, err := cli.BuildCobraCommandFromCommand(
		serveCommand,
		cli.WithParserConfig(cli.CobraParserConfig{
			AppName: "smailnail",
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("build sqlite serve command: %w", err)
	}
	root.AddCommand(cobraCommand)

	return root, nil
}
