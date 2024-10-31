package migrations

import (
	"context"
	"fmt"

	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"

	"github.com/silas/jimmy/internal/constants"
	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

type OnMigration func(m *Migration)

type upgradeOptions struct {
	onStart    OnMigration
	onComplete OnMigration
}

type UpgradeOption func(o *upgradeOptions)

func UpgradeOnStart(m OnMigration) UpgradeOption {
	return func(o *upgradeOptions) {
		o.onStart = m
	}
}

func UpgradeOnComplete(m OnMigration) UpgradeOption {
	return func(o *upgradeOptions) {
		o.onComplete = m
	}
}

func (ms *Migrations) Upgrade(ctx context.Context, opts ...UpgradeOption) error {
	o := &upgradeOptions{}

	for _, opt := range opts {
		opt(o)
	}

	err := ms.ensureAll(ctx)
	if err != nil {
		return err
	}

	dbAdmin, err := ms.DatabaseAdmin(ctx)
	if err != nil {
		return err
	}

	db, err := ms.Database(ctx)
	if err != nil {
		return err
	}

	err = ms.ensureTable(ctx, dbAdmin, db)
	if err != nil {
		return fmt.Errorf("failed to ensure migration table: %w", err)
	}

	currentID, err := ms.getCurrentID(ctx, db)
	if err != nil {
		return err
	}

	for id := currentID + 1; id <= ms.latestID; id++ {
		m, err := ms.Get(id)
		if err != nil {
			return err
		}

		if o.onStart != nil {
			o.onStart(m)
		}

		err = ms.startMigration(ctx, db, id)
		if err != nil {
			return err
		}

		var sqlStatements []string
		var sqlType jimmyv1.Type

		for pos, statement := range m.data.Upgrade {
			switch statement.Env {
			case jimmyv1.Environment_ALL:
				// ok
			case jimmyv1.Environment_GOOGLE_CLOUD:
				if ms.emulator {
					continue
				}
			case jimmyv1.Environment_EMULATOR:
				if !ms.emulator {
					continue
				}
			default:
				return fmt.Errorf("unhandled environment %d upgrade[%d]: %s", id, pos, statement.Env.String())
			}

			if statement.Type == jimmyv1.Type_AUTOMATIC {
				statement.Type = detectType(statement.Sql)
			}

			if sqlType != statement.Type {
				err = ms.runStatements(ctx, dbAdmin, db, sqlType, sqlStatements)
				if err != nil {
					return err
				}
				sqlStatements = nil
			}

			sqlStatements = append(sqlStatements, statement.Sql)
			sqlType = statement.Type
		}

		err = ms.runStatements(ctx, dbAdmin, db, sqlType, sqlStatements)
		if err != nil {
			return err
		}

		err = ms.completeMigration(ctx, db, id)
		if err != nil {
			return err
		}

		if o.onComplete != nil {
			o.onComplete(m)
		}
	}

	return nil
}

func (ms *Migrations) getCurrentID(ctx context.Context, db *spanner.Client) (int, error) {
	var currentID int64
	var complete bool

	err := db.Single().Query(ctx, spanner.Statement{
		SQL: fmt.Sprintf(constants.SelectMigration, ms.Config.Table),
	}).Do(func(r *spanner.Row) error {
		return r.Columns(&currentID, &complete)
	})
	if err != nil {
		return 0, err
	}

	if currentID > 0 && !complete {
		return 0, fmt.Errorf("migration %d is incomplete", currentID)
	}

	return int(currentID), nil
}

func (ms *Migrations) startMigration(ctx context.Context, db *spanner.Client, id int) error {
	_, err := db.Apply(ctx, []*spanner.Mutation{
		spanner.Insert(
			ms.Config.Table,
			[]string{"id", "start_time"},
			[]any{int64(id), spanner.CommitTimestamp},
		),
	})
	return err
}

func (ms *Migrations) runStatements(
	ctx context.Context,
	dbAdmin *database.DatabaseAdminClient,
	db *spanner.Client,
	sqlType jimmyv1.Type,
	sqlStatements []string,
) error {
	if len(sqlStatements) == 0 {
		return nil
	}

	switch sqlType {
	case jimmyv1.Type_DDL:
		op, err := dbAdmin.UpdateDatabaseDdl(ctx, &databasepb.UpdateDatabaseDdlRequest{
			Database:   ms.DatabaseName(),
			Statements: sqlStatements,
		})
		if err != nil {
			return err
		}

		err = op.Wait(ctx)
		if err != nil {
			return err
		}
	case jimmyv1.Type_DML:
		var statements []spanner.Statement

		for _, sql := range sqlStatements {
			statements = append(statements, spanner.Statement{
				SQL: sql,
			})
		}

		_, err := db.ReadWriteTransaction(ctx, func(ctx context.Context, tx *spanner.ReadWriteTransaction) error {
			_, err := tx.BatchUpdate(ctx, statements)
			return err
		})
		if err != nil {
			return err
		}
	case jimmyv1.Type_PARTITIONED_DML:
		for _, sql := range sqlStatements {
			_, err := db.PartitionedUpdate(ctx, spanner.Statement{
				SQL: sql,
			})
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unhandled type %s", sqlType.String())
	}

	return nil
}

func (ms *Migrations) completeMigration(ctx context.Context, db *spanner.Client, id int) error {
	_, err := db.Apply(ctx, []*spanner.Mutation{
		spanner.Update(
			ms.Config.Table,
			[]string{"id", "complete_time"},
			[]any{int64(id), spanner.CommitTimestamp},
		),
	})
	return err
}
