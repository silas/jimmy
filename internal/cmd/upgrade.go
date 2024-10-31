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
			ms, err := newMigrations(cmd, true)
			if err != nil {
				return err
			}
			defer ms.Close()

			err = ms.Upgrade(
				cmd.Context(),
				migrations.UpgradeOnStart(func(m *migrations.Migration) {
					cmd.Println(fmt.Sprintf("migration[%d]: Running %q", m.ID(), m.Summary()))
				}),
				migrations.UpgradeOnComplete(func(m *migrations.Migration) {
					cmd.Println(fmt.Sprintf("migration[%d]: Completed", m.ID()))
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
