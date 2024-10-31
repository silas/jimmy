package cmd

import (
	"github.com/spf13/cobra"
)

func newInit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration files",
		Args:  args(),
		RunE: func(cmd *cobra.Command, args []string) error {
			ms, err := newMigrations(cmd, false)
			if err != nil {
				return err
			}
			defer ms.Close()

			return ms.Init(cmd.Context())
		},
	}

	return cmd
}
