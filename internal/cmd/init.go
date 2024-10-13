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
			m, err := newMigrations(cmd, false)
			if err != nil {
				return err
			}
			defer m.Close()

			return m.Init(cmd.Context())
		},
	}

	return cmd
}
