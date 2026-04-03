package annotate

import (
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/spf13/cobra"
)

func NewAnnotateCommand() (*cobra.Command, error) {
	annotateCmd := &cobra.Command{
		Use:   "annotate",
		Short: "Create and inspect annotations in the local mirror database",
	}

	annotationRoot, err := NewAnnotationRootCommand()
	if err != nil {
		return nil, err
	}
	annotateCmd.AddCommand(annotationRoot)

	groupRoot, err := NewGroupRootCommand()
	if err != nil {
		return nil, err
	}
	annotateCmd.AddCommand(groupRoot)

	logRoot, err := NewLogRootCommand()
	if err != nil {
		return nil, err
	}
	annotateCmd.AddCommand(logRoot)

	return annotateCmd, nil
}

func addGlazedSubcommands(root *cobra.Command, factories ...func() (cmds.Command, error)) error {
	for _, factory := range factories {
		command, err := factory()
		if err != nil {
			return err
		}
		cobraCmd, err := cli.BuildCobraCommandFromCommand(
			command,
			cli.WithParserConfig(cli.CobraParserConfig{
				AppName: "smailnail",
			}),
		)
		if err != nil {
			return fmt.Errorf("build annotate subcommand: %w", err)
		}
		root.AddCommand(cobraCmd)
	}
	return nil
}
