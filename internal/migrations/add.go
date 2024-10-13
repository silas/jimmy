package migrations

import (
	"context"
	"fmt"

	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

type AddInput struct {
	ID       int
	SQL      string
	Env      jimmyv1.Environment
	Template jimmyv1.Template
	Type     jimmyv1.Type
}

func (m *Migrations) Add(_ context.Context, input AddInput) error {
	migration, err := m.LoadMigration(input.ID)
	if err != nil {
		return fmt.Errorf("failed to load migration %d: %w", input.ID, err)
	}

	statement, err := generateStatement(input.SQL, input.Env, input.Template, input.Type)
	if err != nil {
		return err
	}

	migration.Upgrade = append(migration.Upgrade, statement)

	err = Marshal(m.MigrationPath(input.ID), migration)
	if err != nil {
		return err
	}

	return nil
}
