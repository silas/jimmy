package migrations

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/silas/jimmy/internal/constants"
)

func (m *Migrations) Init(ctx context.Context) error {
	_, err := os.Stat(m.Path)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	} else {
		return fmt.Errorf("%q already exists", m.Path)
	}

	if m.Config.Path == "" {
		m.Config.Path = constants.MigrationsPath
	}

	if m.Config.Table == "" {
		m.Config.Table = constants.MigrationsTable
	}

	err = Marshal(m.Path, m.Config)
	if err != nil {
		return err
	}

	err = m.ensureEnv(ctx)
	if err != nil {
		return err
	}

	return nil
}
