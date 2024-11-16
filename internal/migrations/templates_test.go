package migrations_test

import (
	"iter"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/silas/jimmy/internal/migrations"
	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

func TestMigrations_Templates(t *testing.T) {
	h := helper(t)

	testCases := []struct {
		Name      string
		Templates iter.Seq2[string, *jimmyv1.Template]
		Contains  []string
	}{
		{
			Name:      "Builtin templates",
			Templates: migrations.BuiltinTemplates(),
			Contains:  []string{"create-table"},
		},
		{
			Name:      "Config templates",
			Templates: h.Migrations.Templates(),
			Contains:  []string{"create-table", "hello-world"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var prevTemplateID, defaultTemplateID string
			var templateIDs []string

			for templateID, template := range h.Migrations.Templates() {
				if prevTemplateID != "" {
					require.Greater(t, templateID, prevTemplateID)
				}

				if template.GetDefault() {
					require.Empty(t, defaultTemplateID)

					defaultTemplateID = templateID
				}

				templateIDs = append(templateIDs, templateID)
				prevTemplateID = templateID
			}

			require.Equal(t, "create-table", defaultTemplateID)

			for _, templateID := range tc.Contains {
				require.Contains(t, templateIDs, templateID)
			}
		})
	}
}
