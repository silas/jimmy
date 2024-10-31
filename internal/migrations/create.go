package migrations

import (
	"context"
	"fmt"
	"strings"

	"github.com/silas/jimmy/internal/constants"
	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

type CreateInput struct {
	Name     string
	SQL      string
	Env      jimmyv1.Environment
	Template jimmyv1.Template
	Type     jimmyv1.Type
}

func (ms *Migrations) Create(ctx context.Context, input CreateInput) (*Migration, error) {
	err := ms.ensureEnv(ctx)
	if err != nil {
		return nil, err
	}

	statement, err := newStatement(input.SQL, input.Env, input.Template, input.Type)
	if err != nil {
		return nil, err
	}

	slug := Slugify(input.Name)
	if slug == "" {
		if input.Template != jimmyv1.Template_TEMPLATE_UNSPECIFIED {
			slug = strings.ToLower(jimmyv1.Template_name[int32(input.Template)])
		} else {
			slug = "none"
		}
	}

	return ms.create(slug, &jimmyv1.Migration{
		Upgrade: []*jimmyv1.Statement{statement},
	})
}

func (ms *Migrations) create(slug string, data *jimmyv1.Migration) (*Migration, error) {
	ms.latestID++

	m := newMigration(
		ms,
		ms.latestID,
		fmt.Sprintf("%05d_%s%s", ms.latestID, slug, constants.FileExt),
		data,
	)

	ms.migrations[ms.latestID] = m

	err := Marshal(m.Path(), data)
	if err != nil {
		return nil, err
	}

	return m, err
}
