package migrations

import (
	"fmt"
	"os"
	"path/filepath"

	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"

	"github.com/silas/jimmy/internal/constants"
	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

type Migrations struct {
	Path   string
	Config *jimmyv1.Config

	emulator   bool
	migrations map[int]string
	latestId   int

	instanceAdmin *instance.InstanceAdminClient
	databaseAdmin *database.DatabaseAdminClient
	database      *spanner.Client

	instanceEnsured bool
	databaseEnsured bool
}

func New(path string) *Migrations {
	return &Migrations{
		Path:   path,
		Config: &jimmyv1.Config{},

		emulator:   os.Getenv(constants.EnvEmulatorHost) != "",
		migrations: map[int]string{},
	}
}

func (m *Migrations) Close() {
	if m.instanceAdmin != nil {
		instanceAdmin := m.instanceAdmin
		m.instanceAdmin = nil
		defer instanceAdmin.Close()
	}

	if m.databaseAdmin != nil {
		databaseAdmin := m.databaseAdmin
		m.databaseAdmin = nil
		defer databaseAdmin.Close()
	}

	if m.database != nil {
		db := m.database
		m.database = nil
		defer db.Close()
	}
}

func (m *Migrations) LatestId() int {
	return m.latestId
}

func (m *Migrations) MigrationPath(id int) string {
	name := m.migrations[id]
	if name == "" {
		return ""
	}

	return filepath.Join(m.Config.Path, name)
}

func (m *Migrations) InstancesName() string {
	return fmt.Sprintf("projects/%s/instances", m.Config.Project)
}

func (m *Migrations) InstanceName() string {
	return fmt.Sprintf("%s/%s", m.InstancesName(), m.Config.InstanceId)
}

func (m *Migrations) DatabasesName() string {
	return fmt.Sprintf("%s/databases", m.InstanceName())
}

func (m *Migrations) DatabaseName() string {
	return fmt.Sprintf("%s/%s", m.DatabasesName(), m.Config.DatabaseId)
}
