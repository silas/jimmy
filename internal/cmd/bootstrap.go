package cmd

import (
	"github.com/spf13/cobra"
)

func newBootstrap() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bootstrap [flags]",
		Short: "Create initial migration",
		Args:  args(),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := newMigrations(cmd, true)
			if err != nil {
				return err
			}
			defer m.Close()

			id, err := m.Bootstrap(cmd.Context())
			if err != nil {
				return err
			}

			cmd.Println(m.MigrationPath(id))

			return nil
		},
	}

	return cmd
}
