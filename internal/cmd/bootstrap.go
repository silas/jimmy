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
			ms, err := newMigrations(cmd, true)
			if err != nil {
				return err
			}
			defer ms.Close()

			m, err := ms.Bootstrap(cmd.Context())
			if err != nil {
				return err
			}

			cmd.Println(m.Path())

			return nil
		},
	}

	return cmd
}
