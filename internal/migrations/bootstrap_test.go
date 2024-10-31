package migrations_test

import (
	"slices"
	"testing"

	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	"github.com/stretchr/testify/require"

	"github.com/silas/jimmy/internal/constants"
	"github.com/silas/jimmy/internal/migrations"
	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

func TestMigrations_Bootstrap(t *testing.T) {
	h := helper(t)

	err := h.Migrations.Init(h.Ctx)
	require.NoError(t, err)
	require.Equal(t, 0, h.Migrations.LatestID())

	_, err = h.Migrations.Bootstrap(h.Ctx)
	require.EqualError(t, err, "no statements")

	dbAdmin, err := h.Migrations.DatabaseAdmin(h.Ctx)
	require.NoError(t, err)

	op, err := dbAdmin.UpdateDatabaseDdl(h.Ctx, &databasepb.UpdateDatabaseDdlRequest{
		Database:   h.Migrations.DatabaseName(),
		Statements: []string{constants.CreateTable},
	})
	require.NoError(t, err)
	require.NoError(t, op.Wait(h.Ctx))

	m, err := h.Migrations.Bootstrap(h.Ctx)
	require.NoError(t, err)
	require.Equal(t, 1, m.ID())
	require.Equal(t, m.ID(), h.Migrations.LatestID())

	statements := slices.Collect(m.Upgrade())
	require.Len(t, statements, 1)
	require.Contains(t, statements[0].Sql, "CREATE TABLE test")
	require.Equal(t, jimmyv1.Type_DDL, statements[0].Type)

	_, err = h.Migrations.Create(h.Ctx, migrations.CreateInput{
		Name: "insert",
		SQL:  `INSERT INTO test (id, update_time) VALUES ("one", CURRENT_TIMESTAMP)`,
	})
	require.NoError(t, err)

	err = h.Migrations.Upgrade(h.Ctx)
	require.NoError(t, err)

	records, err := h.records()
	require.NoError(t, err)
	require.Len(t, records, 1)
	require.Equal(t, 2, records[0].ID)
}
