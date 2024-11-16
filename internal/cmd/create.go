package cmd

import (
	"github.com/spf13/cobra"

	"github.com/silas/jimmy/internal/migrations"
)

func newCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [flags] name",
		Short: "Create a new migration",
		Args:  args("name"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ms, err := getMigrations(cmd, true)
			if err != nil {
				return err
			}
			defer ms.Close()

			var m *migrations.Migration

			bootstrap, err := cmd.Flags().GetBool(flagBootstrap)
			if err != nil {
				return err
			}

			if bootstrap {
				m, err = ms.Bootstrap(cmd.Context(), migrations.BootstrapInput{
					Name: args[0],
				})
				if err != nil {
					return err
				}
			} else {
				flags, err := parseStatementFlags(cmd)
				if err != nil {
					return err
				}

				squashID, err := cmd.Flags().GetInt(flagSquash)
				if err != nil {
					return err
				}

				m, err = ms.Create(cmd.Context(), migrations.CreateInput{
					Name:       args[0],
					SQL:        flags.SQL,
					Env:        flags.Env,
					TemplateID: flags.Template,
					Type:       flags.Type,
					SquashID:   squashID,
				})
				if err != nil {
					return err
				}
			}

			cmd.Println(m.Path())

			return nil
		},
	}

	setupStatementFlags(cmd)

	cmd.Flags().BoolP(flagBootstrap, "", false, "populate from current schema")
	cmd.Flags().IntP(flagSquash, "", 0, "squash ID")

	return cmd
}
