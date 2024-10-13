package cmd

import (
	"github.com/spf13/cobra"
)

func newUpgrade() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Run all schema upgrades",
		Args:  args(),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := newMigrations(cmd, true)
			if err != nil {
				return err
			}
			defer m.Close()

			return m.Upgrade(cmd.Context())
		},
	}

	return cmd
}
