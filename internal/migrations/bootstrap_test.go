package migrations_test

import (
	"testing"

	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	"github.com/stretchr/testify/require"

	"github.com/silas/jimmy/internal/constants"
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

	data := m.Data()
	require.NotNil(t, data)
	require.Len(t, data.Upgrade, 1)
	require.Contains(t, data.Upgrade[0].Sql, "CREATE TABLE test")
	require.Equal(t, jimmyv1.Type_DDL, data.Upgrade[0].Type)
}
