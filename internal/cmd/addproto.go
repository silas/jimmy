package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/silas/jimmy/internal/migrations"
)

func newAddProto() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proto [flags] name file",
		Short: "Add a protobuf-serialized file descriptor set",
		Args:  args("name", "file"),
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

			if args[0] == "" {
				return errors.New("name is required")
			} else if args[1] == "" {
				return errors.New("path is required")
			}

			err = ms.AddProto(cmd.Context(), migrations.AddProtoInput{
				ID:   m.ID(),
				Name: args[0],
				Path: args[1],
			})
			if err != nil {
				return err
			}

			cmd.Println(m.Path())

			return nil
		},
	}

	setupMigrationFlag(cmd)

	return cmd
}
