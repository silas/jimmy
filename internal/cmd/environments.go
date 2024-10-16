package cmd

import (
	"github.com/spf13/cobra"

	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

func newEnvironments() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "environments",
		Short:   "Show environments",
		Aliases: []string{"envs"},
		Args:    args(),
		RunE: func(cmd *cobra.Command, args []string) error {
			printEnums(
				cmd,
				jimmyv1.Environment_name,
				false,
				jimmyv1.Environment_ALL,
			)
			return nil
		},
	}

	return cmd
}
