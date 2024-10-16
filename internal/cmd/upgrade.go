package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/silas/jimmy/internal/migrations"
)

func newUpgrade() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "upgrade",
		Short:   "Run all schema upgrades",
		Aliases: []string{"up"},
		Args:    args(),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := newMigrations(cmd, true)
			if err != nil {
				return err
			}
			defer m.Close()

			err = m.Upgrade(
				cmd.Context(),
				migrations.UpgradeOnStart(func(id int, name string) {
					cmd.Println(fmt.Sprintf("migration[%d]: Running %q", id, name))
				}),
				migrations.UpgradeOnComplete(func(id int, name string) {
					cmd.Println(fmt.Sprintf("migration[%d]: Completed", id))
				}),
			)
			if err != nil {
				return err
			}

			cmd.Println("Done")

			return nil
		},
	}

	return cmd
}
