package migrations

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"

	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

func (m *Migrations) Load(_ context.Context) error {
	fileInfo, err := os.Stat(m.Path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%q config file not found", m.Path)
		}

		return err
	}

	mode := fileInfo.Mode()

	if mode.IsDir() {
		return fmt.Errorf("%q is a directory, expected a text file", m.Path)
	} else if !mode.IsRegular() {
		return fmt.Errorf("%q is not a regular file", m.Path)
	}

	err = Unmarshal(m.Path, m.Config)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(m.Config.Path)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.Type().IsRegular() {
			continue
		}

		name := file.Name()

		idx := strings.Index(file.Name(), "_")
		if idx == -1 {
			continue
		}

		id, err := strconv.Atoi(name[:idx])
		if err != nil {
			continue
		}

		if conflictingName, conflicts := m.migrations[id]; conflicts {
			return fmt.Errorf("migration %d has conflicting migration files %q and %q", id, name, conflictingName)
		}

		m.migrations[id] = name

		m.latestId = max(m.latestId, id)
	}

	return nil
}

func (m *Migrations) LoadMigration(id int) (*jimmyv1.Migration, error) {
	migration := &jimmyv1.Migration{}

	err := Unmarshal(m.MigrationPath(id), migration)
	if err != nil {
		return nil, err
	}

	return migration, nil
}
