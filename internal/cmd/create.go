package cmd

import (
	"github.com/spf13/cobra"

	"github.com/silas/jimmy/internal/migrations"
)

const (
	flagEnv      = "env"
	flagSQL      = "sql"
	flagTemplate = "template"
	flagType     = "type"
	flagSquash   = "squash"
)

func newCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [flags] name",
		Short: "Create a new migration",
		Args:  args("name"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ms, err := newMigrations(cmd, true)
			if err != nil {
				return err
			}
			defer ms.Close()

			flags, err := parseMigrationFlags(cmd)
			if err != nil {
				return err
			}

			squashID, err := cmd.Flags().GetInt(flagSquash)
			if err != nil {
				return err
			}

			m, err := ms.Create(cmd.Context(), migrations.CreateInput{
				Name:     args[0],
				SQL:      flags.SQL,
				Env:      flags.Env,
				Template: flags.Template,
				Type:     flags.Type,
				SquashID: squashID,
			})
			if err != nil {
				return err
			}

			cmd.Println(m.Path())

			return nil
		},
	}

	setupMigrationFlags(cmd)

	cmd.Flags().IntP(flagSquash, "", 0, "squash ID")

	return cmd
}
