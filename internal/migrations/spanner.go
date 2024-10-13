package migrations

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	"cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/silas/jimmy/internal/constants"
)

func (m *Migrations) InstanceAdmin(ctx context.Context) (*instance.InstanceAdminClient, error) {
	if m.instanceAdmin == nil {
		var err error

		m.instanceAdmin, err = instance.NewInstanceAdminClient(ctx)
		if err != nil {
			return nil, err
		}
	}

	return m.instanceAdmin, nil
}

func (m *Migrations) DatabaseAdmin(ctx context.Context) (*database.DatabaseAdminClient, error) {
	if m.databaseAdmin == nil {
		var err error

		m.databaseAdmin, err = database.NewDatabaseAdminClient(ctx)
		if err != nil {
			return nil, err
		}
	}

	return m.databaseAdmin, nil
}

func (m *Migrations) Database(ctx context.Context) (*spanner.Client, error) {
	if m.database == nil {
		var err error

		m.database, err = spanner.NewClient(ctx, m.DatabaseName())
		if err != nil {
			return nil, err
		}
	}

	return m.database, nil
}

func (m *Migrations) ensureAll(ctx context.Context) error {
	err := m.ensureEnv(ctx)
	if err != nil {
		return err
	}

	err = m.ensureInstance(ctx)
	if err != nil {
		return fmt.Errorf("failed to ensure instance: %w", err)
	}

	err = m.ensureDatabase(ctx)
	if err != nil {
		return fmt.Errorf("failed to ensure database: %w", err)
	}

	return nil
}

func (m *Migrations) ensureEnv(_ context.Context) error {
	err := os.MkdirAll(m.Config.Path, 0755)
	if err != nil {
		return err
	}

	return nil
}

func (m *Migrations) ensureInstance(ctx context.Context) error {
	if !m.emulator || m.instanceEnsured {
		return nil
	}
	m.instanceEnsured = true

	instanceAdmin, err := m.InstanceAdmin(ctx)
	if err != nil {
		return err
	}

	inst, err := instanceAdmin.GetInstance(ctx, &instancepb.GetInstanceRequest{
		Name: m.InstanceName(),
	})
	if err != nil && status.Code(err) != codes.NotFound {
		return err
	}

	if inst == nil {
		op, err := instanceAdmin.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
			Parent:     m.InstancesName(),
			InstanceId: m.Config.InstanceId,
		})
		if err != nil {
			return err
		}

		inst, err = op.Wait(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrations) ensureDatabase(ctx context.Context) error {
	if !m.emulator || m.databaseEnsured {
		return nil
	}
	m.databaseEnsured = true

	dbAdmin, err := m.DatabaseAdmin(ctx)
	if err != nil {
		return err
	}

	db, err := dbAdmin.GetDatabase(ctx, &databasepb.GetDatabaseRequest{
		Name: m.DatabaseName(),
	})
	if err != nil && status.Code(err) != codes.NotFound {
		return err
	}

	if db == nil {
		op, err := dbAdmin.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
			Parent:          m.InstanceName(),
			CreateStatement: fmt.Sprintf("CREATE DATABASE %s", m.Config.DatabaseId),
		})
		if err != nil {
			return err
		}

		db, err = op.Wait(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrations) ensureTable(
	ctx context.Context,
	dbAdmin *database.DatabaseAdminClient,
	db *spanner.Client,
) error {
	var exists bool

	err := db.Single().Query(ctx, spanner.Statement{
		SQL: constants.SelectMigrationsTable,
		Params: map[string]any{
			"tableSchema": "",
			"tableName":   m.Config.Table,
		},
	}).Do(func(r *spanner.Row) error {
		exists = true
		return nil
	})
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	op, err := dbAdmin.UpdateDatabaseDdl(ctx, &databasepb.UpdateDatabaseDdlRequest{
		Database: m.DatabaseName(),
		Statements: []string{
			fmt.Sprintf(constants.CreateMigrationTable, m.Config.Table),
		},
	})
	if err != nil {
		return err
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	return nil
}
