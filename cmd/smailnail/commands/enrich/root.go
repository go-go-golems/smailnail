package enrich

import (
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/spf13/cobra"
)

func NewEnrichCommand() (*cobra.Command, error) {
	enrichCmd := &cobra.Command{
		Use:   "enrich",
		Short: "Run enrichment passes over the local mirror database",
	}

	factories := []func() (cmds.Command, error){
		func() (cmds.Command, error) { return NewSendersCommand() },
		func() (cmds.Command, error) { return NewThreadsCommand() },
		func() (cmds.Command, error) { return NewUnsubscribeCommand() },
		func() (cmds.Command, error) { return NewAllCommand() },
	}

	for _, factory := range factories {
		command, err := factory()
		if err != nil {
			return nil, err
		}
		cobraCmd, err := cli.BuildCobraCommandFromCommand(
			command,
			cli.WithParserConfig(cli.CobraParserConfig{
				AppName: "smailnail",
			}),
		)
		if err != nil {
			return nil, fmt.Errorf("build enrich subcommand: %w", err)
		}
		enrichCmd.AddCommand(cobraCmd)
	}

	return enrichCmd, nil
}
