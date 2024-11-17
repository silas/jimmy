package cmd

import (
	"github.com/spf13/cobra"

	"github.com/silas/jimmy/internal/migrations"
)

func newInit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize migrations",
		Args:  args(),
		RunE: func(cmd *cobra.Command, args []string) error {
			ms, err := getMigrations(cmd, false)
			if err != nil {
				return err
			}
			defer ms.Close()

			ctx := cmd.Context()

			err = ms.Init(ctx)
			if err != nil {
				return err
			}

			bootstrap, err := cmd.Flags().GetBool(flagBootstrap)
			if err != nil {
				return err
			}

			if !bootstrap {
				return nil
			}

			err = ms.Load(ctx)
			if err != nil {
				return err
			}

			m, err := ms.Bootstrap(cmd.Context(), migrations.BootstrapInput{})
			if err != nil {
				return err
			}

			cmd.Println(m.Path())

			return nil
		},
	}

	cmd.Flags().BoolP(flagBootstrap, "", false, "create initial migration from current schema")

	return cmd
}
