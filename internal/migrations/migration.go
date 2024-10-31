package migrations

import (
	"iter"
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

func (m *Migration) Name() string {
	return strings.ReplaceAll(m.Slug(), "_", " ")
}

func (m *Migration) SquashID() (int, bool) {
	if m != nil && m.data != nil {
		return int(m.data.GetSquashId()), m.data.SquashId != nil
	}
	return 0, false
}

func (m *Migration) Upgrade() iter.Seq[*jimmyv1.Statement] {
	return func(yield func(*jimmyv1.Statement) bool) {
		if m != nil {
			for _, s := range m.data.GetUpgrade() {
				if !yield(proto.Clone(s).(*jimmyv1.Statement)) {
					return
				}
			}
		}
	}
}
