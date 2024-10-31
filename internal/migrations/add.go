package migrations

import (
	"context"

	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

type AddInput struct {
	ID       int
	SQL      string
	Env      jimmyv1.Environment
	Template jimmyv1.Template
	Type     jimmyv1.Type
}

func (ms *Migrations) Add(_ context.Context, input AddInput) error {
	m, err := ms.Get(input.ID)
	if err != nil {
		return err
	}

	statement, err := newStatement(input.SQL, input.Env, input.Template, input.Type)
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
