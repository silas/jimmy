package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/silas/jimmy/internal/migrations"
)

func newAdd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [flags] id",
		Short: "Add a statement to an existing migration",
		Args:  args("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := newMigrations(cmd, true)
			if err != nil {
				return err
			}
			defer m.Close()

			var id int

			argId := args[0]
			if argId == "latest" {
				id = m.LatestId()
			} else {
				id, err = strconv.Atoi(argId)
				if err != nil {
					return err
				}
			}

			migrationPath := m.MigrationPath(id)

			if migrationPath == "" {
				return fmt.Errorf("migration %d not found", id)
			}

			flags, err := parseMigrationFlags(cmd)
			if err != nil {
				return err
			}

			err = m.Add(cmd.Context(), migrations.AddInput{
				ID:       id,
				SQL:      flags.SQL,
				Template: flags.Template,
				Env:      flags.Env,
				Type:     flags.Type,
			})
			if err != nil {
				return err
			}

			cmd.Println(migrationPath)

			return nil
		},
	}

	setupMigrationFlags(cmd)

	return cmd
}
