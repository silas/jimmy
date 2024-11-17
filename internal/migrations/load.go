package migrations

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/silas/jimmy/internal/constants"
	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

func (ms *Migrations) Load(_ context.Context) error {
	err := checkFile(ms.Path, "config")
	if err != nil {
		return err
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

		ms.setMigration(m)
	}

	return nil
}
