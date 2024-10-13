package migrations_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path"
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	"github.com/stretchr/testify/require"

	"github.com/silas/jimmy/internal/constants"
	"github.com/silas/jimmy/internal/migrations"
	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

func TestMigrations(t *testing.T) {
	h := helper(t)

	db, err := h.Migrations.Database(h.Ctx)
	require.NoError(t, err)

	// initialize
	{
		err := h.Migrations.Init(h.Ctx)
		require.NoError(t, err)
		require.Equal(t, 0, h.Migrations.LatestId())
	}

	// dml migration
	{
		id, err := h.Migrations.Create(h.Ctx, migrations.CreateInput{
			Name:     "test",
			Template: jimmyv1.Template_CREATE_TABLE,
		})
		require.NoError(t, err)
		require.Equal(t, 1, id)
		require.Equal(t, id, h.Migrations.LatestId())

		_, err = h.list()
		require.Error(t, err)

		err = h.Migrations.Upgrade(h.Ctx)
		require.NoError(t, err)

		data, err := h.list()
		require.NoError(t, err)
		require.Len(t, data, 1)

		latest := data[0]
		require.Equal(t, 1, latest.ID)

		migration, err := h.Migrations.LoadMigration(latest.ID)
		require.NoError(t, err)
		require.Len(t, migration.Upgrade, 1)

		upgrade := migration.Upgrade[0]
		require.Contains(t, upgrade.Sql, "CREATE TABLE")
		require.Equal(t, jimmyv1.Environment_ALL.String(), upgrade.Env.String())
		require.Equal(t, jimmyv1.Type_DDL.String(), upgrade.Type.String())
	}

	// no-op when latest
	{
		err := h.Migrations.Upgrade(h.Ctx)
		require.NoError(t, err)

		data, err := h.list()
		require.NoError(t, err)
		require.Len(t, data, 1)
	}

	// add migration
	{
		id, err := h.Migrations.Create(h.Ctx, migrations.CreateInput{
			Name: "add-slug",
			SQL:  "ALTER TABLE test ADD COLUMN slug STRING(MAX)",
		})
		require.NoError(t, err)
		require.Equal(t, 2, id)
		require.Equal(t, 2, h.Migrations.LatestId())

		err = h.Migrations.Add(h.Ctx, migrations.AddInput{
			ID:  2,
			SQL: "ALTER TABLE test ALTER COLUMN id SET DEFAULT (GENERATE_UUID())",
		})
		require.NoError(t, err)

		err = h.Migrations.Upgrade(h.Ctx)
		require.NoError(t, err)

		data, err := h.list()
		require.NoError(t, err)
		require.Len(t, data, 2)

		latest := data[1]
		require.Equal(t, 2, latest.ID)
		require.False(t, latest.StartTime.IsZero())
		require.True(t, latest.CompleteTime.After(latest.StartTime))

		migration, err := h.Migrations.LoadMigration(latest.ID)
		require.NoError(t, err)
		require.Len(t, migration.Upgrade, 2)

		upgrade := migration.Upgrade[0]
		require.Contains(t, upgrade.Sql, "ADD COLUMN slug")
		require.Equal(t, jimmyv1.Environment_ALL.String(), upgrade.Env.String())
		require.Equal(t, jimmyv1.Type_DDL.String(), upgrade.Type.String())

		upgrade = migration.Upgrade[1]
		require.Contains(t, upgrade.Sql, "ALTER COLUMN id")
		require.Equal(t, jimmyv1.Environment_ALL.String(), upgrade.Env.String())
		require.Equal(t, jimmyv1.Type_DDL.String(), upgrade.Type.String())
	}

	// insert table
	{
		id, err := h.Migrations.Create(h.Ctx, migrations.CreateInput{
			Name: "insert",
			SQL:  `INSERT INTO test (name, update_time) VALUES ("one", CURRENT_TIMESTAMP)`,
		})
		require.NoError(t, err)
		require.Equal(t, 3, id)

		err = h.Migrations.Upgrade(h.Ctx)
		require.NoError(t, err)

		data, err := h.list()
		require.NoError(t, err)
		require.Len(t, data, 3)

		latest := data[2]
		require.Equal(t, 3, latest.ID)

		migration, err := h.Migrations.LoadMigration(latest.ID)
		require.NoError(t, err)
		require.Len(t, migration.Upgrade, 1)

		upgrade := migration.Upgrade[0]
		require.Contains(t, upgrade.Sql, "INSERT INTO")
		require.Equal(t, jimmyv1.Environment_ALL.String(), upgrade.Env.String())
		require.Equal(t, jimmyv1.Type_DML.String(), migration.Upgrade[0].Type.String())

		var count int

		err = db.Single().Read(
			h.Ctx,
			"test",
			spanner.AllKeys(),
			[]string{"id", "name", "update_time"},
		).Do(func(r *spanner.Row) error {
			count++

			var id string
			var name spanner.NullString
			var updateTime time.Time

			err = r.Columns(&id, &name, &updateTime)
			require.NoError(t, err)
			require.NotEmpty(t, id)
			require.Equal(t, "one", name.StringVal)
			require.False(t, updateTime.IsZero())

			return nil
		})
		require.NoError(t, err)
		require.Equal(t, 1, count)
	}

	// emulator only
	{
		id, err := h.Migrations.Create(h.Ctx, migrations.CreateInput{
			Name: "insert-no-emulator",
			SQL:  `INSERT INTO test (name, update_time) VALUES ("two", CURRENT_TIMESTAMP)`,
			Env:  jimmyv1.Environment_GOOGLE_CLOUD,
		})
		require.NoError(t, err)
		require.Equal(t, 4, id)

		err = h.Migrations.Upgrade(h.Ctx)
		require.NoError(t, err)

		data, err := h.list()
		require.NoError(t, err)
		require.Len(t, data, 4)

		latest := data[3]
		require.Equal(t, 4, latest.ID)

		migration, err := h.Migrations.LoadMigration(latest.ID)
		require.NoError(t, err)
		require.Len(t, migration.Upgrade, 1)

		upgrade := migration.Upgrade[0]
		require.Contains(t, upgrade.Sql, "INSERT INTO")
		require.Equal(t, jimmyv1.Environment_GOOGLE_CLOUD.String(), upgrade.Env.String())
		require.Equal(t, jimmyv1.Type_DML.String(), migration.Upgrade[0].Type.String())

		var count int

		err = db.Single().Read(
			h.Ctx,
			"test",
			spanner.AllKeys(),
			[]string{"id", "name", "update_time"},
		).Do(func(r *spanner.Row) error {
			count++
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, 1, count)
	}

	// update partitioned DML
	{
		id, err := h.Migrations.Create(h.Ctx, migrations.CreateInput{
			Name: "invalid",
			SQL:  `UPDATE test SET update_time = CURRENT_TIMESTAMP WHERE 1=1`,
			Type: jimmyv1.Type_PARTITIONED_DML,
		})
		require.NoError(t, err)
		require.Equal(t, 5, id)

		err = h.Migrations.Upgrade(h.Ctx)
		require.NoError(t, err)

		data, err := h.list()
		require.NoError(t, err)
		require.Len(t, data, 5)

		latest := data[4]
		require.Equal(t, 5, latest.ID)

		migration, err := h.Migrations.LoadMigration(latest.ID)
		require.NoError(t, err)
		require.Len(t, migration.Upgrade, 1)

		upgrade := migration.Upgrade[0]
		require.Contains(t, upgrade.Sql, "UPDATE test SET update_time")
		require.Equal(t, jimmyv1.Environment_ALL.String(), upgrade.Env.String())
		require.Equal(t, jimmyv1.Type_PARTITIONED_DML.String(), migration.Upgrade[0].Type.String())
	}

	// failure
	{
		id, err := h.Migrations.Create(h.Ctx, migrations.CreateInput{
			Name: "invalid",
			SQL:  `CREATE failure`,
		})
		require.NoError(t, err)
		require.Equal(t, 6, id)

		err = h.Migrations.Upgrade(h.Ctx)
		require.Error(t, err)

		data, err := h.list()
		require.NoError(t, err)
		require.Len(t, data, 6)

		latest := data[5]
		require.Equal(t, 6, latest.ID)
		require.True(t, latest.CompleteTime.IsZero())
	}
}

func init() {
	if os.Getenv(constants.EnvEmulatorHost) == "" {
		os.Setenv(constants.EnvEmulatorHost, constants.EnvEmulatorHostDefault)
	}
}

func helper(t *testing.T) *helperData {
	tmpDir, err := os.MkdirTemp("", "jimmy")
	require.NoError(t, err)

	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	h := &helperData{
		Ctx:  context.Background(),
		Path: tmpDir,
	}

	h.Migrations = migrations.New(
		path.Join(tmpDir, constants.ConfigFile),
	)
	t.Cleanup(h.Migrations.Close)

	h.Migrations.Config.Path = path.Join(tmpDir, constants.MigrationsPath)
	h.Migrations.Config.Project = "demo-project"
	h.Migrations.Config.InstanceId = "test"
	h.Migrations.Config.DatabaseId = fmt.Sprintf("test%d", rand.Int63n(999999))

	t.Cleanup(func() {
		dbAdmin, err := h.Migrations.DatabaseAdmin(h.Ctx)
		require.NoError(t, err)

		err = dbAdmin.DropDatabase(h.Ctx, &databasepb.DropDatabaseRequest{
			Database: h.Migrations.DatabaseName(),
		})
		require.NoError(t, err)
	})

	return h
}

type helperData struct {
	Ctx        context.Context
	Path       string
	Migrations *migrations.Migrations
}

func (h *helperData) list() ([]*Migration, error) {
	var data []*Migration

	db, err := h.Migrations.Database(h.Ctx)
	if err != nil {
		return nil, err
	}

	err = db.Single().Read(
		h.Ctx,
		h.Migrations.Config.Table,
		spanner.AllKeys(),
		[]string{"id", "start_time", "complete_time"},
	).Do(func(r *spanner.Row) error {
		var m Migration

		var id int64
		var completeTime spanner.NullTime

		err = r.Columns(&id, &m.StartTime, &completeTime)
		if err != nil {
			return err
		}

		m.ID = int(id)
		m.CompleteTime = completeTime.Time

		data = append(data, &m)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

type Migration struct {
	ID           int
	StartTime    time.Time
	CompleteTime time.Time
}
