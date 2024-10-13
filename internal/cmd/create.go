package cmd

import (
	"github.com/spf13/cobra"

	"github.com/silas/jimmy/internal/migrations"
)

const (
	flagEnv      = "env"
	flagSql      = "sql"
	flagTemplate = "template"
	flagType     = "type"
)

func newCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [flags] name",
		Short: "Create a new migration",
		Args:  args("name"),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := newMigrations(cmd, true)
			if err != nil {
				return err
			}
			defer m.Close()

			flags, err := parseMigrationFlags(cmd)
			if err != nil {
				return err
			}

			id, err := m.Create(cmd.Context(), migrations.CreateInput{
				Name:     args[0],
				SQL:      flags.SQL,
				Env:      flags.Env,
				Template: flags.Template,
				Type:     flags.Type,
			})
			if err != nil {
				return err
			}

			cmd.Println(m.MigrationPath(id))

			return nil
		},
	}

	setupMigrationFlags(cmd)

	return cmd
}
