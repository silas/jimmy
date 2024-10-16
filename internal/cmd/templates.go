package cmd

import (
	"github.com/spf13/cobra"

	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

func newTemplates() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "templates",
		Short: "Show templates",
		Args:  args(),
		RunE: func(cmd *cobra.Command, args []string) error {
			printEnums(
				cmd,
				jimmyv1.Template_name,
				true,
				jimmyv1.Template_CREATE_TABLE,
			)
			return nil
		},
	}

	return cmd
}
