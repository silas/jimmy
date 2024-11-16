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
		Short: "Show template options",
		Args:  args(),
		RunE: func(cmd *cobra.Command, args []string) error {
			var templates iter.Seq2[string, *jimmyv1.Template]

			ms, err := newMigrations(cmd, true)
			if err != nil {
				templates = migrations.BuiltinTemplates()
			} else {
				templates = ms.Templates()
				ms.Close()
			}

			for templateID, template := range templates {
				if template.GetDefault() {
					cmd.Println(templateID, "(default)")
				} else {
					cmd.Println(templateID)
				}
			}

			return nil
		},
	}

	return cmd
}
