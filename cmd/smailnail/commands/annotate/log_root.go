package annotate

import (
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/spf13/cobra"
)

func NewLogRootCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "log",
		Short: "Manage annotation logs",
	}
	if err := addGlazedSubcommands(
		root,
		func() (cmds.Command, error) { return NewLogAddCommand() },
		func() (cmds.Command, error) { return NewLogListCommand() },
		func() (cmds.Command, error) { return NewLogLinkTargetCommand() },
		func() (cmds.Command, error) { return NewLogTargetsCommand() },
	); err != nil {
		return nil, err
	}
	return root, nil
}
