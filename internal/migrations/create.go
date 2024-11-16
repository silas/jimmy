package migrations

import (
	"context"
	"fmt"

	"github.com/silas/jimmy/internal/constants"
	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

type CreateInput struct {
	Name       string
	SQL        string
	Env        jimmyv1.Environment
	TemplateID string
	Type       jimmyv1.Type
	SquashID   int
}

func (ms *Migrations) Create(ctx context.Context, input CreateInput) (*Migration, error) {
	err := ms.ensureEnv(ctx)
	if err != nil {
		return nil, err
	}

	statement, err := ms.newStatement(input.SQL, input.Env, input.TemplateID, input.Type)
	if err != nil {
		return nil, err
	}

	slug := Slugify(input.Name)
	if slug == "" {
		slug = Slugify(input.TemplateID)
	}
	if slug == "" {
		slug = "none"
	}

	m := &jimmyv1.Migration{
		Upgrade: []*jimmyv1.Statement{statement},
	}
	if input.SquashID > 0 {
		sm, err := ms.Get(input.SquashID)
		if err != nil {
			return nil, err
		}

		if squashID, found := sm.SquashID(); found {
			return nil, fmt.Errorf(
				"new squash migration can't reference existing squash migration %d",
				squashID,
			)
		}

		squashID := int32(sm.ID())

		m.SquashId = &squashID
	}

	return ms.create(slug, m)
}

func (ms *Migrations) create(slug string, data *jimmyv1.Migration) (*Migration, error) {
	id := ms.latestID + 1

	m := newMigration(
		ms,
		id,
		fmt.Sprintf("%05d_%s%s", id, slug, constants.FileExt),
		data,
	)

	ms.setMigration(m)

	err := Marshal(m.Path(), data)
	if err != nil {
		return nil, err
	}

	return m, err
}

func (ms *Migrations) setMigration(m *Migration) {
	ms.migrations[m.id] = m
	ms.latestID = max(ms.latestID, m.id)

	squashID, found := m.SquashID()
	if found {
		ms.squash[squashID] = max(m.id, ms.squash[squashID])
	}
}
