package cmd

import (
	"iter"

	"github.com/spf13/cobra"

	"github.com/silas/jimmy/internal/migrations"
	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

func newTemplates() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "templates",
		Short: "Show templates",
		Args:  args(),
		RunE: func(cmd *cobra.Command, args []string) error {
			var templates iter.Seq2[string, *jimmyv1.Template]

			ms, err := getMigrations(cmd, true)
			if err != nil {
				templates = migrations.BuiltinTemplates()
			} else {
				templates = ms.Templates()
				ms.Close()
			}

			for templateID := range templates {
				cmd.Println(templateID)
			}

			return nil
		},
	}

	return cmd
}
