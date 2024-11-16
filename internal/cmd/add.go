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
			ms, err := newMigrations(cmd, true)
			if err != nil {
				return err
			}
			defer ms.Close()

			var id int

			argID := args[0]
			if argID == "latest" {
				id = ms.LatestID()
			} else {
				id, err = strconv.Atoi(argID)
				if err != nil {
					return fmt.Errorf("migration %q not found", argID)
				}
			}

			m, err := ms.Get(id)
			if err != nil {
				return err
			}

			flags, err := parseMigrationFlags(cmd)
			if err != nil {
				return err
			}

			err = ms.Add(cmd.Context(), migrations.AddInput{
				ID:         id,
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

	setupMigrationFlags(cmd)

	return cmd
}
