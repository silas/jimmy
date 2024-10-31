package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/silas/jimmy/internal/migrations"
	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
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
					cmd.Println(fmt.Sprintf("migration[%d]: Started %q", m.ID(), m.Name()))
				}),
				migrations.UpgradeOnBatch(func(m *migrations.Migration, batch []*jimmyv1.Statement) {
					var suffix string

					if len(batch) != 1 {
						suffix = "s"
					}

					cmd.Println(fmt.Sprintf(
						"migration[%d]: Running %d %s statement%s",
						m.ID(),
						len(batch),
						batch[0].Type.String(),
						suffix,
					))
				}),
				migrations.UpgradeOnComplete(func(m *migrations.Migration) {
					cmd.Println(fmt.Sprintf("migration[%d]: Completed", m.ID()))
				}),
			)
			if err != nil {
				return err
			}

			cmd.Println(fmt.Sprintf("Done at migration %d", ms.LatestID()))

			return nil
		},
	}

	return cmd
}
