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

func (ms *Migrations) InstanceAdmin(ctx context.Context) (*instance.InstanceAdminClient, error) {
	if ms.instanceAdmin == nil {
		var err error

		ms.instanceAdmin, err = instance.NewInstanceAdminClient(ctx)
		if err != nil {
			return nil, err
		}
	}

	return ms.instanceAdmin, nil
}

func (ms *Migrations) DatabaseAdmin(ctx context.Context) (*database.DatabaseAdminClient, error) {
	if ms.databaseAdmin == nil {
		var err error

		ms.databaseAdmin, err = database.NewDatabaseAdminClient(ctx)
		if err != nil {
			return nil, err
		}
	}

	return ms.databaseAdmin, nil
}

func (ms *Migrations) Database(ctx context.Context) (*spanner.Client, error) {
	if ms.database == nil {
		var err error

		ms.database, err = spanner.NewClient(ctx, ms.DatabaseName())
		if err != nil {
			return nil, err
		}
	}

	return ms.database, nil
}

func (ms *Migrations) ensureAll(ctx context.Context) error {
	err := ms.ensureEnv(ctx)
	if err != nil {
		return err
	}

	err = ms.ensureInstance(ctx)
	if err != nil {
		return fmt.Errorf("failed to ensure instance: %w", err)
	}

	err = ms.ensureDatabase(ctx)
	if err != nil {
		return fmt.Errorf("failed to ensure database: %w", err)
	}

	return nil
}

func (ms *Migrations) ensureEnv(_ context.Context) error {
	err := os.MkdirAll(ms.Config.Path, 0755)
	if err != nil {
		return err
	}

	return nil
}

func (ms *Migrations) ensureInstance(ctx context.Context) error {
	if !ms.emulator || ms.instanceEnsured {
		return nil
	}
	ms.instanceEnsured = true

	instanceAdmin, err := ms.InstanceAdmin(ctx)
	if err != nil {
		return err
	}

	inst, err := instanceAdmin.GetInstance(ctx, &instancepb.GetInstanceRequest{
		Name: ms.InstanceName(),
	})
	if err != nil && status.Code(err) != codes.NotFound {
		return err
	}

	if inst == nil {
		op, err := instanceAdmin.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
			Parent:     ms.InstancesName(),
			InstanceId: ms.Config.InstanceId,
		})
		if err != nil {
			return err
		}

		_, err = op.Wait(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ms *Migrations) ensureDatabase(ctx context.Context) error {
	if !ms.emulator || ms.databaseEnsured {
		return nil
	}
	ms.databaseEnsured = true

	dbAdmin, err := ms.DatabaseAdmin(ctx)
	if err != nil {
		return err
	}

	db, err := dbAdmin.GetDatabase(ctx, &databasepb.GetDatabaseRequest{
		Name: ms.DatabaseName(),
	})
	if err != nil && status.Code(err) != codes.NotFound {
		return err
	}

	if db == nil {
		op, err := dbAdmin.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
			Parent:          ms.InstanceName(),
			CreateStatement: fmt.Sprintf("CREATE DATABASE %s", ms.Config.DatabaseId),
		})
		if err != nil {
			return err
		}

		_, err = op.Wait(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ms *Migrations) ensureTable(
	ctx context.Context,
	dbAdmin *database.DatabaseAdminClient,
	db *spanner.Client,
) error {
	var exists bool

	err := db.Single().Query(ctx, spanner.Statement{
		SQL: constants.SelectMigrationsTable,
		Params: map[string]any{
			"tableSchema": "",
			"tableName":   ms.Config.Table,
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
		Database: ms.DatabaseName(),
		Statements: []string{
			fmt.Sprintf(constants.CreateMigrationTable, ms.Config.Table),
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
