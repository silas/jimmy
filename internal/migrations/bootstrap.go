package migrations

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"

	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

func (ms *Migrations) Bootstrap(ctx context.Context) (*Migration, error) {
	err := ms.ensureAll(ctx)
	if err != nil {
		return nil, err
	}

	var upgrade []*jimmyv1.Statement

	dbAdmin, err := ms.DatabaseAdmin(ctx)
	if err != nil {
		return nil, err
	}

	ddl, err := dbAdmin.GetDatabaseDdl(ctx, &databasepb.GetDatabaseDdlRequest{
		Database: ms.DatabaseName(),
	})
	if err != nil {
		return nil, err
	}

	if len(ddl.Statements) == 0 {
		return nil, errors.New("no statements")
	}

	migrationTableDDL := fmt.Sprintf("CREATE TABLE %s (", ms.Config.Table)

	for _, sql := range ddl.Statements {
		if strings.Contains(sql, migrationTableDDL) {
			continue
		}

		statement, err := newStatement(
			sql,
			jimmyv1.Environment_ALL,
			jimmyv1.Template_TEMPLATE_UNSPECIFIED,
			jimmyv1.Type_DDL,
		)
		if err != nil {
			return nil, err
		}

		upgrade = append(upgrade, statement)
	}

	squashID := int32(ms.latestID)

	return ms.create("init", &jimmyv1.Migration{
		Upgrade:  upgrade,
		SquashId: &squashID,
	})
}
