package migrations

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/silas/jimmy/internal/constants"
)

func (ms *Migrations) Init(ctx context.Context) error {
	_, err := os.Stat(ms.Path)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	} else {
		return fmt.Errorf("%q already exists", ms.Path)
	}

	if ms.Config.Path == "" {
		ms.Config.Path = constants.MigrationsPath
	}

	if ms.Config.Table == "" {
		ms.Config.Table = constants.MigrationsTable
	}

	err = Marshal(ms.Path, ms.Config)
	if err != nil {
		return err
	}

	err = ms.ensureEnv(ctx)
	if err != nil {
		return err
	}

	return nil
}
