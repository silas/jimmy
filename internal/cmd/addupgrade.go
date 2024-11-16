package cmd

import (
	"github.com/spf13/cobra"

	"github.com/silas/jimmy/internal/migrations"
)

func newAddUpgrade() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Add an upgrade statement",
		Args:  args(),
		RunE: func(cmd *cobra.Command, args []string) error {
			ms, err := getMigrations(cmd, true)
			if err != nil {
				return err
			}
			defer ms.Close()

			m, err := getMigration(cmd, ms)
			if err != nil {
				return err
			}

			flags, err := parseStatementFlags(cmd)
			if err != nil {
				return err
			}

			err = ms.AddUpgrade(cmd.Context(), migrations.AddUpgradeInput{
				ID:         m.ID(),
				SQL:        flags.SQL,
				TemplateID: flags.Template,
				Env:        flags.Env,
				Type:       flags.Type,
			})
			if err != nil {
				return err
			}

			cmd.Println(m.Path())

			return nil
		},
	}

	setupMigrationFlag(cmd)
	setupStatementFlags(cmd)

	return cmd
}
