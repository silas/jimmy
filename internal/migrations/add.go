package migrations

import (
	"context"

	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

type AddInput struct {
	ID         int
	SQL        string
	Env        jimmyv1.Environment
	TemplateID string
	Type       jimmyv1.Type
}

func (ms *Migrations) Add(_ context.Context, input AddInput) error {
	m, err := ms.Get(input.ID)
	if err != nil {
		return err
	}

	statement, err := ms.newStatement(input.SQL, input.Env, input.TemplateID, input.Type)
	if err != nil {
		return err
	}

	m.data.Upgrade = append(m.data.Upgrade, statement)

	err = Marshal(m.Path(), m.data)
	if err != nil {
		return err
	}

	return nil
}
