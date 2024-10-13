package migrations

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

type CreateInput struct {
	Name     string
	SQL      string
	Env      jimmyv1.Environment
	Template jimmyv1.Template
	Type     jimmyv1.Type
}

func (m *Migrations) Create(ctx context.Context, input CreateInput) (int, error) {
	err := m.ensureEnv(ctx)
	if err != nil {
		return 0, err
	}

	statement, err := generateStatement(input.SQL, input.Env, input.Template, input.Type)
	if err != nil {
		return 0, err
	}

	name := Slugify(input.Name)
	if name == "" {
		if input.Template != jimmyv1.Template_TEMPLATE_UNSPECIFIED {
			name = strings.ToLower(jimmyv1.Template_name[int32(input.Template)])
		} else {
			name = "none"
		}
	}

	return m.createMigration(name, &jimmyv1.Migration{
		Upgrade: []*jimmyv1.Statement{statement},
	})
}

func (m *Migrations) createMigration(name string, migration *jimmyv1.Migration) (int, error) {
	m.latestId++

	fileName := fmt.Sprintf("%05d_%s.yaml", m.latestId, name)

	m.migrations[m.latestId] = fileName

	err := Marshal(filepath.Join(m.Config.Path, fileName), migration)
	if err != nil {
		return 0, err
	}

	return m.latestId, err
}
