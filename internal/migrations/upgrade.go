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

type OnMigration func(id int, name string)

type upgradeOptions struct {
	onInit     func(toRun int)
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

func (m *Migrations) Upgrade(ctx context.Context, opts ...UpgradeOption) error {
	o := &upgradeOptions{}

	for _, opt := range opts {
		opt(o)
	}

	err := m.ensureAll(ctx)
	if err != nil {
		return err
	}

	dbAdmin, err := m.DatabaseAdmin(ctx)
	if err != nil {
		return err
	}

	db, err := m.Database(ctx)
	if err != nil {
		return err
	}

	err = m.ensureTable(ctx, dbAdmin, db)
	if err != nil {
		return fmt.Errorf("failed to ensure migration table: %w", err)
	}

	currentID, err := m.getCurrentID(ctx, db)
	if err != nil {
		return err
	}

	for id := currentID + 1; id <= m.latestId; id++ {
		migration, err := m.LoadMigration(id)
		if err != nil {
			return fmt.Errorf("failed to load migration %d: %w", id, err)
		}

		name := m.MigrationName(id)

		if o.onStart != nil {
			o.onStart(id, name)
		}

		err = m.startMigration(ctx, db, id)
		if err != nil {
			return err
		}

		var sqlStatements []string
		var sqlType jimmyv1.Type

		for pos, statement := range migration.Upgrade {
			switch statement.Env {
			case jimmyv1.Environment_ALL:
				// ok
			case jimmyv1.Environment_GOOGLE_CLOUD:
				if m.emulator {
					continue
				}
			case jimmyv1.Environment_EMULATOR:
				if !m.emulator {
					continue
				}
			default:
				return fmt.Errorf("unhandled environment %d upgrade[%d]: %s", id, pos, statement.Env.String())
			}

			if statement.Type == jimmyv1.Type_AUTOMATIC {
				statement.Type = detectType(statement.Sql)
			}

			if sqlType != statement.Type {
				err = m.runStatements(ctx, dbAdmin, db, sqlType, sqlStatements)
				if err != nil {
					return err
				}
				sqlStatements = nil
			}

			sqlStatements = append(sqlStatements, statement.Sql)
			sqlType = statement.Type
		}

		err = m.runStatements(ctx, dbAdmin, db, sqlType, sqlStatements)
		if err != nil {
			return err
		}

		err = m.completeMigration(ctx, db, id)
		if err != nil {
			return err
		}

		if o.onComplete != nil {
			o.onComplete(id, name)
		}
	}

	return nil
}

func (m *Migrations) getCurrentID(ctx context.Context, db *spanner.Client) (int, error) {
	var currentID int64
	var complete bool

	err := db.Single().Query(ctx, spanner.Statement{
		SQL: fmt.Sprintf(constants.SelectMigration, m.Config.Table),
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

func (m *Migrations) startMigration(ctx context.Context, db *spanner.Client, id int) error {
	_, err := db.Apply(ctx, []*spanner.Mutation{
		spanner.Insert(
			m.Config.Table,
			[]string{"id", "start_time"},
			[]any{int64(id), spanner.CommitTimestamp},
		),
	})
	return err
}

func (m *Migrations) runStatements(
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
			Database:   m.DatabaseName(),
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

func (m *Migrations) completeMigration(ctx context.Context, db *spanner.Client, id int) error {
	_, err := db.Apply(ctx, []*spanner.Mutation{
		spanner.Update(
			m.Config.Table,
			[]string{"id", "complete_time"},
			[]any{int64(id), spanner.CommitTimestamp},
		),
	})
	return err
}
