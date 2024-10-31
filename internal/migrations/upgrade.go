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

type OnMigrationBatch func(m *Migration, batch []*jimmyv1.Statement)

type upgradeOptions struct {
	onStart    OnMigration
	onBatch    OnMigrationBatch
	onComplete OnMigration
}

type UpgradeOption func(o *upgradeOptions)

func UpgradeOnStart(onStart OnMigration) UpgradeOption {
	return func(o *upgradeOptions) {
		o.onStart = onStart
	}
}

func UpgradeOnBatch(onBatch OnMigrationBatch) UpgradeOption {
	return func(o *upgradeOptions) {
		o.onBatch = onBatch
	}
}

func UpgradeOnComplete(onComplete OnMigration) UpgradeOption {
	return func(o *upgradeOptions) {
		o.onComplete = onComplete
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

	id := currentID

	for id < ms.latestID {
		id++
		startID := id

		if skipID := ms.squash[id]; skipID > id {
			id = skipID
		}

		m, err := ms.Get(id)
		if err != nil {
			return err
		}

		squashID, found := m.SquashID()
		if found && squashID != startID {
			continue
		}

		if o.onStart != nil {
			o.onStart(m)
		}

		err = ms.startMigration(ctx, db, id)
		if err != nil {
			return err
		}

		var batch []*jimmyv1.Statement

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

			if len(batch) > 0 && batch[0].Type != statement.Type {
				err = ms.batch(ctx, dbAdmin, db, m, batch, o.onBatch)
				if err != nil {
					return err
				}

				batch = nil
			}

			batch = append(batch, statement)
		}

		err = ms.batch(ctx, dbAdmin, db, m, batch, o.onBatch)
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

func (ms *Migrations) batch(
	ctx context.Context,
	dbAdmin *database.DatabaseAdminClient,
	db *spanner.Client,
	m *Migration,
	batch []*jimmyv1.Statement,
	onRun OnMigrationBatch,
) error {
	if len(batch) == 0 {
		return nil
	}

	if onRun != nil {
		onRun(m, batch)
	}

	switch batch[0].Type {
	case jimmyv1.Type_DDL:
		var statements []string

		for _, b := range batch {
			statements = append(statements, b.Sql)
		}

		op, err := dbAdmin.UpdateDatabaseDdl(ctx, &databasepb.UpdateDatabaseDdlRequest{
			Database:   ms.DatabaseName(),
			Statements: statements,
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

		for _, b := range batch {
			statements = append(statements, spanner.Statement{SQL: b.Sql})
		}

		_, err := db.ReadWriteTransaction(ctx, func(ctx context.Context, tx *spanner.ReadWriteTransaction) error {
			_, err := tx.BatchUpdate(ctx, statements)
			return err
		})
		if err != nil {
			return err
		}
	case jimmyv1.Type_PARTITIONED_DML:
		for _, b := range batch {
			_, err := db.PartitionedUpdate(ctx, spanner.Statement{
				SQL: b.Sql,
			})
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unhandled type %s", batch[0].String())
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
