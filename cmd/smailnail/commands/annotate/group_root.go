package annotate

import (
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/spf13/cobra"
)

func NewGroupRootCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "group",
		Short: "Manage target groups",
	}
	if err := addGlazedSubcommands(
		root,
		func() (cmds.Command, error) { return NewGroupCreateCommand() },
		func() (cmds.Command, error) { return NewGroupListCommand() },
		func() (cmds.Command, error) { return NewGroupAddTargetCommand() },
		func() (cmds.Command, error) { return NewGroupMembersCommand() },
	); err != nil {
		return nil, err
	}
	return root, nil
}
