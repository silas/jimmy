package migrations

import (
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/silas/jimmy/internal/constants"
	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

type Migration struct {
	ms       *Migrations
	id       int
	fileName string
	data     *jimmyv1.Migration
}

func newMigration(
	ms *Migrations,
	id int,
	fileName string,
	data *jimmyv1.Migration,
) *Migration {
	return &Migration{
		ms:       ms,
		id:       id,
		fileName: fileName,
		data:     data,
	}
}

func (m *Migration) ID() int {
	if m != nil {
		return m.id
	}
	return 0
}

func (m *Migration) FileName() string {
	if m != nil {
		return m.fileName
	}
	return ""
}

func (m *Migration) Path() string {
	if m != nil {
		return filepath.Join(m.ms.Config.Path, m.FileName())
	}
	return ""
}

func (m *Migration) Slug() string {
	fileName := m.FileName()

	idx := strings.Index(fileName, "_")
	if idx == -1 {
		return ""
	}

	slug := fileName[idx+1:]
	slug, _ = strings.CutSuffix(slug, constants.FileExt)
	return slug
}

func (m *Migration) Summary() string {
	return strings.ReplaceAll(m.Slug(), "_", " ")
}

func (m *Migration) Data() *jimmyv1.Migration {
	if m != nil && m.data != nil {
		return proto.Clone(m.data).(*jimmyv1.Migration)
	}
	return nil
}
