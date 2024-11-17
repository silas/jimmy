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
			ms, err := getMigrations(cmd, true)
			if err != nil {
				return err
			}
			defer ms.Close()

			err = ms.Upgrade(
				cmd.Context(),
				migrations.UpgradeOnStart(func(m *migrations.Migration) {
					cmd.Println(fmt.Sprintf("migration[%d]: Started %q", m.ID(), m.Name()))
				}),
				migrations.UpgradeOnBatch(func(m *migrations.Migration, batch *migrations.Batch) {
					var suffix string

					if len(batch.Statements) != 1 {
						suffix = "s"
					}

					if batch.FileDescriptorSet != "" {
						suffix += fmt.Sprintf(" with file descriptor set %q",
							batch.FileDescriptorSet)
					}

					cmd.Println(fmt.Sprintf(
						"migration[%d]: Running %d %s statement%s",
						m.ID(),
						len(batch.Statements),
						batch.Statements[0].Type.String(),
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
