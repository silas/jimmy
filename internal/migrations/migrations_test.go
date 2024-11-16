package migrations_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path"
	"slices"
	"strings"
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
		require.Equal(t, 0, h.Migrations.LatestID())
	}

	// validate
	{
		require.NoError(t, h.Migrations.Validate())
	}

	// dml migration
	{
		m, err := h.Migrations.Create(h.Ctx, migrations.CreateInput{
			Name:       "test-init",
			TemplateID: "create-table",
		})
		require.NoError(t, err)
		require.Equal(t, 1, m.ID())
		require.Equal(t, m.ID(), h.Migrations.LatestID())
		require.Equal(t, "00001_test_init.yaml", m.FileName())
		require.True(t, strings.HasSuffix(m.Path(), "/"+m.FileName()))

		stat, err := os.Stat(m.Path())
		require.NoError(t, err)
		require.True(t, stat.Mode().IsRegular())

		require.Equal(t, "test init", m.Name())

		_, err = h.records()
		require.Error(t, err)

		var started, completed bool
		var batchCount int

		err = h.Migrations.Upgrade(
			h.Ctx,
			migrations.UpgradeOnStart(func(m *migrations.Migration) {
				require.False(t, started)
				require.Equal(t, 1, m.ID())
				require.Equal(t, "test init", m.Name())
				started = true
			}),
			migrations.UpgradeOnBatch(func(m *migrations.Migration, batch []*jimmyv1.Statement) {
				require.Equal(t, 1, m.ID())
				require.Equal(t, "test init", m.Name())
				require.Len(t, batch, 1)
				batchCount++
			}),
			migrations.UpgradeOnComplete(func(m *migrations.Migration) {
				require.False(t, completed)
				require.Equal(t, 1, m.ID())
				require.Equal(t, "test init", m.Name())
				completed = true
			}),
		)
		require.NoError(t, err)
		require.True(t, started)
		require.Equal(t, 1, batchCount)
		require.True(t, completed)

		records, err := h.records()
		require.NoError(t, err)
		require.Len(t, records, 1)

		record := records[0]
		require.Equal(t, 1, record.ID)

		m, err = h.Migrations.Get(record.ID)
		require.NoError(t, err)

		statements := slices.Collect(m.Upgrade())
		require.Len(t, statements, 1)

		statement := statements[0]
		require.Contains(t, statement.Sql, "CREATE TABLE")
		require.Equal(t, jimmyv1.Environment_ALL.String(), statement.Env.String())
		require.Equal(t, jimmyv1.Type_DDL.String(), statement.Type.String())
	}

	// no-op when latest
	{
		err := h.Migrations.Upgrade(h.Ctx)
		require.NoError(t, err)

		records, err := h.records()
		require.NoError(t, err)
		require.Len(t, records, 1)
	}

	// add migration
	{
		m, err := h.Migrations.Create(h.Ctx, migrations.CreateInput{
			Name: "add-slug",
			SQL:  "ALTER TABLE test ADD COLUMN slug STRING(MAX)",
		})
		require.NoError(t, err)
		require.NotNil(t, m)
		require.Equal(t, 2, m.ID())
		require.Equal(t, 2, h.Migrations.LatestID())

		err = h.Migrations.Add(h.Ctx, migrations.AddInput{
			ID:  2,
			SQL: "ALTER TABLE test ALTER COLUMN id SET DEFAULT (GENERATE_UUID())",
		})
		require.NoError(t, err)

		err = h.Migrations.Upgrade(h.Ctx)
		require.NoError(t, err)

		records, err := h.records()
		require.NoError(t, err)
		require.Len(t, records, 2)

		record := records[1]
		require.Equal(t, 2, record.ID)
		require.False(t, record.StartTime.IsZero())
		require.True(t, record.CompleteTime.After(record.StartTime))

		m, err = h.Migrations.Get(record.ID)
		require.NoError(t, err)

		statements := slices.Collect(m.Upgrade())
		require.Len(t, statements, 2)

		statement := statements[0]
		require.Contains(t, statement.Sql, "ADD COLUMN slug")
		require.Equal(t, jimmyv1.Environment_ALL.String(), statement.Env.String())
		require.Equal(t, jimmyv1.Type_DDL.String(), statement.Type.String())

		statement = statements[1]
		require.Contains(t, statement.Sql, "ALTER COLUMN id")
		require.Equal(t, jimmyv1.Environment_ALL.String(), statement.Env.String())
		require.Equal(t, jimmyv1.Type_DDL.String(), statement.Type.String())
	}

	// insert table
	{
		m, err := h.Migrations.Create(h.Ctx, migrations.CreateInput{
			Name: "insert",
			SQL:  `INSERT INTO test (name, update_time) VALUES ("one", CURRENT_TIMESTAMP)`,
		})
		require.NoError(t, err)
		require.NotNil(t, m)
		require.Equal(t, 3, m.ID())

		err = h.Migrations.Upgrade(h.Ctx)
		require.NoError(t, err)

		records, err := h.records()
		require.NoError(t, err)
		require.Len(t, records, 3)

		record := records[2]
		require.Equal(t, 3, record.ID)

		m, err = h.Migrations.Get(record.ID)
		require.NoError(t, err)

		statements := slices.Collect(m.Upgrade())
		require.Len(t, statements, 1)

		statement := statements[0]
		require.Contains(t, statement.Sql, "INSERT INTO")
		require.Equal(t, jimmyv1.Environment_ALL.String(), statement.Env.String())
		require.Equal(t, jimmyv1.Type_DML.String(), statement.Type.String())

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
		m, err := h.Migrations.Create(h.Ctx, migrations.CreateInput{
			Name: "insert-no-emulator",
			SQL:  `INSERT INTO test (name, update_time) VALUES ("two", CURRENT_TIMESTAMP)`,
			Env:  jimmyv1.Environment_GOOGLE_CLOUD,
		})
		require.NoError(t, err)
		require.NotNil(t, m)
		require.Equal(t, 4, m.ID())

		err = h.Migrations.Upgrade(h.Ctx)
		require.NoError(t, err)

		records, err := h.records()
		require.NoError(t, err)
		require.Len(t, records, 4)

		record := records[3]
		require.Equal(t, 4, record.ID)

		m, err = h.Migrations.Get(record.ID)
		require.NoError(t, err)

		statements := slices.Collect(m.Upgrade())
		require.Len(t, statements, 1)

		statement := statements[0]
		require.Contains(t, statement.Sql, "INSERT INTO")
		require.Equal(t, jimmyv1.Environment_GOOGLE_CLOUD.String(), statement.Env.String())
		require.Equal(t, jimmyv1.Type_DML.String(), statement.Type.String())

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
		m, err := h.Migrations.Create(h.Ctx, migrations.CreateInput{
			Name: "invalid",
			SQL:  `UPDATE test SET update_time = CURRENT_TIMESTAMP WHERE 1=1`,
			Type: jimmyv1.Type_PARTITIONED_DML,
		})
		require.NoError(t, err)
		require.Equal(t, 5, m.ID())

		err = h.Migrations.Upgrade(h.Ctx)
		require.NoError(t, err)

		records, err := h.records()
		require.NoError(t, err)
		require.Len(t, records, 5)

		record := records[4]
		require.Equal(t, 5, record.ID)

		m, err = h.Migrations.Get(record.ID)
		require.NoError(t, err)

		statements := slices.Collect(m.Upgrade())
		require.Len(t, statements, 1)

		statement := statements[0]
		require.Contains(t, statement.Sql, "UPDATE test SET update_time")
		require.Equal(t, jimmyv1.Environment_ALL.String(), statement.Env.String())
		require.Equal(t, jimmyv1.Type_PARTITIONED_DML.String(), statement.Type.String())
	}

	// squash
	{
		m, err := h.Migrations.Create(h.Ctx, migrations.CreateInput{
			Name: "skip",
			SQL:  `INSERT INTO test (name, update_time) VALUES ("three", CURRENT_TIMESTAMP)`,
		})
		require.NoError(t, err)

		_, err = h.Migrations.Create(h.Ctx, migrations.CreateInput{
			Name: "skip2",
			SQL:  `INSERT INTO test (name, update_time) VALUES ("four", CURRENT_TIMESTAMP)`,
		})
		require.NoError(t, err)

		_, err = h.Migrations.Create(h.Ctx, migrations.CreateInput{
			Name:     "squash",
			SQL:      `INSERT INTO test (name, update_time) VALUES ("five", CURRENT_TIMESTAMP)`,
			SquashID: m.ID(),
		})
		require.NoError(t, err)

		err = h.Migrations.Upgrade(h.Ctx)
		require.NoError(t, err)

		records, err := h.records()
		require.NoError(t, err)
		require.Len(t, records, 6)
	}

	// squash (skip)
	{
		_, err = h.Migrations.Create(h.Ctx, migrations.CreateInput{
			Name:     "squash-skip",
			SQL:      `INSERT INTO test (name, update_time) VALUES ("seven", CURRENT_TIMESTAMP)`,
			SquashID: 5,
		})
		require.NoError(t, err)

		err = h.Migrations.Upgrade(h.Ctx)
		require.NoError(t, err)

		records, err := h.records()
		require.NoError(t, err)
		require.Len(t, records, 6)
	}

	// failure
	{
		m, err := h.Migrations.Create(h.Ctx, migrations.CreateInput{
			Name: "invalid",
			SQL:  `CREATE failure`,
		})
		require.NoError(t, err)
		require.Equal(t, 10, m.ID())

		err = h.Migrations.Upgrade(h.Ctx)
		require.Error(t, err)

		records, err := h.records()
		require.NoError(t, err)
		require.Len(t, records, 7)

		record := records[6]
		require.Equal(t, 10, record.ID)
		require.True(t, record.CompleteTime.IsZero())
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
	h.Migrations.Config.ProjectId = "demo-project"
	h.Migrations.Config.InstanceId = "test"
	h.Migrations.Config.DatabaseId = fmt.Sprintf("test%d", rand.Int63n(999999))
	h.Migrations.Config.Templates = map[string]*jimmyv1.Template{
		"hello-world": {
			Sql: `
CREATE TABLE hello_world (
  id STRING(MAX) NOT NULL,
  name STRING(MAX) NOT NULL
) PRIMARY KEY (id)`,
		},
	}

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

func (h *helperData) records() ([]*Record, error) {
	var records []*Record

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
		var record Record

		var id int64
		var completeTime spanner.NullTime

		err = r.Columns(&id, &record.StartTime, &completeTime)
		if err != nil {
			return err
		}

		record.ID = int(id)
		record.CompleteTime = completeTime.Time

		records = append(records, &record)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return records, nil
}

type Record struct {
	ID           int
	StartTime    time.Time
	CompleteTime time.Time
}
