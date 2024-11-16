package migrations

import (
	"context"
	"fmt"

	"cloud.google.com/go/spanner"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/silas/jimmy/internal/constants"
	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

type OnMigration func(m *Migration)

type OnMigrationBatch func(m *Migration, batch *Batch)

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

	err = ms.ensureTable(ctx)
	if err != nil {
		return fmt.Errorf("failed to ensure migration table: %w", err)
	}

	currentID, err := ms.getCurrentID(ctx)
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

		err = ms.startMigration(ctx, id)
		if err != nil {
			return err
		}

		batch := &Batch{}

		for pos, s := range m.data.Upgrade {
			switch s.Env {
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
				return fmt.Errorf("unhandled environment %d upgrade[%d]: %s",
					id, pos, s.Env.String())
			}

			if s.Type == jimmyv1.Type_AUTOMATIC {
				s.Type = detectType(s.Sql)
			}

			if batch.flush(s) {
				err = ms.runBatch(ctx, m, batch, o.onBatch)
				if err != nil {
					return err
				}

				batch.reset()
			}

			batch.add(s)
		}

		err = ms.runBatch(ctx, m, batch, o.onBatch)
		if err != nil {
			return err
		}

		err = ms.completeMigration(ctx, id)
		if err != nil {
			return err
		}

		if o.onComplete != nil {
			o.onComplete(m)
		}
	}

	return nil
}

func (ms *Migrations) getCurrentID(ctx context.Context) (int, error) {
	db, err := ms.Database(ctx)
	if err != nil {
		return 0, err
	}

	var currentID int64
	var complete bool

	err = db.Single().Query(ctx, spanner.Statement{
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

func (ms *Migrations) startMigration(ctx context.Context, id int) error {
	db, err := ms.Database(ctx)
	if err != nil {
		return err
	}

	_, err = db.Apply(ctx, []*spanner.Mutation{
		spanner.Insert(
			ms.Config.Table,
			[]string{"id", "start_time"},
			[]any{int64(id), spanner.CommitTimestamp},
		),
	})
	return err
}

type Batch struct {
	Statements        []*jimmyv1.Statement
	FileDescriptorSet string
}

func (b *Batch) flush(s *jimmyv1.Statement) bool {
	if b == nil || s == nil || len(b.Statements) == 0 {
		return false
	}

	if b.Statements[0].Type != s.Type {
		return true
	}

	set := s.GetFileDescriptorSet()

	if set != "" && set != b.FileDescriptorSet {
		return true
	}

	return false
}

func (b *Batch) add(s *jimmyv1.Statement) {
	b.Statements = append(b.Statements, s)

	if s.GetFileDescriptorSet() != "" {
		b.FileDescriptorSet = s.GetFileDescriptorSet()
	}
}

func (b *Batch) reset() {
	b.Statements = nil
	b.FileDescriptorSet = ""
}

func (ms *Migrations) runBatch(
	ctx context.Context,
	m *Migration,
	batch *Batch,
	onRun OnMigrationBatch,
) error {
	if batch == nil || len(batch.Statements) == 0 {
		return nil
	}

	if onRun != nil {
		onRun(m, batch)
	}

	switch batch.Statements[0].Type {
	case jimmyv1.Type_DDL:
		var statements []string

		for _, s := range batch.Statements {
			statements = append(statements, s.Sql)
		}

		req := &databasepb.UpdateDatabaseDdlRequest{
			Database:   ms.DatabaseName(),
			Statements: statements,
		}

		// attach proto descriptors
		if batch.FileDescriptorSet != "" {
			id := batch.FileDescriptorSet

			var fileDescriptorSet *descriptorpb.FileDescriptorSet

			if len(m.data.FileDescriptorSets) > 0 {
				fileDescriptorSet = m.data.FileDescriptorSets[id]
			}

			if fileDescriptorSet == nil {
				return fmt.Errorf("file descriptor set %q not found", id)
			}

			b, err := proto.Marshal(fileDescriptorSet)
			if err != nil {
				return fmt.Errorf("failed to marshal %q file descriptor set", id)
			}

			req.ProtoDescriptors = b
		}

		dbAdmin, err := ms.DatabaseAdmin(ctx)
		if err != nil {
			return err
		}

		op, err := dbAdmin.UpdateDatabaseDdl(ctx, req)
		if err != nil {
			return err
		}

		err = op.Wait(ctx)
		if err != nil {
			return err
		}
	case jimmyv1.Type_DML:
		var statements []spanner.Statement

		for _, s := range batch.Statements {
			statements = append(statements, spanner.Statement{SQL: s.Sql})
		}

		db, err := ms.Database(ctx)
		if err != nil {
			return err
		}

		_, err = db.ReadWriteTransaction(ctx, func(ctx context.Context, tx *spanner.ReadWriteTransaction) error {
			_, err := tx.BatchUpdate(ctx, statements)
			return err
		})
		if err != nil {
			return err
		}
	case jimmyv1.Type_PARTITIONED_DML:
		db, err := ms.Database(ctx)
		if err != nil {
			return err
		}

		for _, s := range batch.Statements {
			_, err := db.PartitionedUpdate(ctx, spanner.Statement{
				SQL: s.Sql,
			})
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unhandled type %s", batch.Statements[0].String())
	}

	return nil
}

func (ms *Migrations) completeMigration(ctx context.Context, id int) error {
	db, err := ms.Database(ctx)
	if err != nil {
		return err
	}

	_, err = db.Apply(ctx, []*spanner.Mutation{
		spanner.Update(
			ms.Config.Table,
			[]string{"id", "complete_time"},
			[]any{int64(id), spanner.CommitTimestamp},
		),
	})
	return err
}
