package migrations

import (
	"fmt"
	"os"

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
	migrations map[int]*Migration
	squash     map[int]int
	latestID   int

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
		migrations: map[int]*Migration{},
		squash:     map[int]int{},
	}
}

func (ms *Migrations) Close() {
	if ms.instanceAdmin != nil {
		instanceAdmin := ms.instanceAdmin
		ms.instanceAdmin = nil
		defer instanceAdmin.Close()
	}

	if ms.databaseAdmin != nil {
		databaseAdmin := ms.databaseAdmin
		ms.databaseAdmin = nil
		defer databaseAdmin.Close()
	}

	if ms.database != nil {
		db := ms.database
		ms.database = nil
		defer db.Close()
	}
}

func (ms *Migrations) LatestID() int {
	return ms.latestID
}

func (ms *Migrations) Get(id int) (*Migration, error) {
	m := ms.migrations[id]
	if m == nil {
		return nil, fmt.Errorf("migration %d not found", id)
	}

	return m, nil
}

func (ms *Migrations) InstancesName() string {
	return fmt.Sprintf("projects/%s/instances", ms.Config.ProjectId)
}

func (ms *Migrations) InstanceName() string {
	return fmt.Sprintf("%s/%s", ms.InstancesName(), ms.Config.InstanceId)
}

func (ms *Migrations) DatabasesName() string {
	return fmt.Sprintf("%s/databases", ms.InstanceName())
}

func (ms *Migrations) DatabaseName() string {
	return fmt.Sprintf("%s/%s", ms.DatabasesName(), ms.Config.DatabaseId)
}
