package migrations

import (
	"context"
	"errors"

	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"

	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

func (m *Migrations) Bootstrap(ctx context.Context) (int, error) {
	err := m.ensureAll(ctx)
	if err != nil {
		return 0, err
	}

	if m.latestId != 0 {
		return 0, errors.New("bootstrap requires no existing migrations")
	}

	var upgrade []*jimmyv1.Statement

	dbAdmin, err := m.DatabaseAdmin(ctx)
	if err != nil {
		return 0, err
	}

	ddl, err := dbAdmin.GetDatabaseDdl(ctx, &databasepb.GetDatabaseDdlRequest{
		Database: m.DatabaseName(),
	})
	if err != nil {
		return 0, err
	}

	if len(ddl.Statements) == 0 {
		return 0, errors.New("no statements")
	}

	for _, sql := range ddl.Statements {
		statement, err := generateStatement(
			sql,
			jimmyv1.Environment_ALL,
			jimmyv1.Template_TEMPLATE_UNSPECIFIED,
			jimmyv1.Type_DDL,
		)
		if err != nil {
			return 0, err
		}

		upgrade = append(upgrade, statement)
	}

	return m.createMigration("init", &jimmyv1.Migration{
		Upgrade: upgrade,
	})
}
