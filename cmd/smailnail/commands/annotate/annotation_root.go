package annotate

import (
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/spf13/cobra"
)

func NewAnnotationRootCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "annotation",
		Short: "Manage target annotations",
	}
	if err := addGlazedSubcommands(
		root,
		func() (cmds.Command, error) { return NewAnnotationAddCommand() },
		func() (cmds.Command, error) { return NewAnnotationListCommand() },
		func() (cmds.Command, error) { return NewAnnotationReviewCommand() },
	); err != nil {
		return nil, err
	}
	return root, nil
}
