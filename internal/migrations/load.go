package migrations

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"github.com/silas/jimmy/internal/constants"
	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

func (ms *Migrations) Load(_ context.Context) error {
	fileInfo, err := os.Stat(ms.Path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%q config file not found", ms.Path)
		}

		return err
	}

	mode := fileInfo.Mode()

	if mode.IsDir() {
		return fmt.Errorf("%q is a directory, expected a text file", ms.Path)
	} else if !mode.IsRegular() {
		return fmt.Errorf("%q is not a regular file", ms.Path)
	}

	err = Unmarshal(ms.Path, ms.Config)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(ms.Config.Path)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.Type().IsRegular() {
			continue
		}

		fileName := file.Name()

		if !strings.HasSuffix(fileName, constants.FileExt) {
			continue
		}

		idx := strings.Index(fileName, "_")
		if idx == -1 {
			continue
		}

		id, err := strconv.Atoi(fileName[:idx])
		if err != nil {
			continue
		}

		m, err := ms.Get(id)
		if err == nil {
			return fmt.Errorf("migration %d has conflicting migration files %q and %q",
				id, fileName, m.fileName)
		}

		m = newMigration(ms, id, fileName, &jimmyv1.Migration{})

		err = Unmarshal(m.Path(), m.data)
		if err != nil {
			return err
		}

		ms.migrations[id] = m

		ms.latestID = max(ms.latestID, id)
	}

	return nil
}
