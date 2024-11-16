package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"

	"github.com/silas/jimmy/internal/constants"
	"github.com/silas/jimmy/internal/migrations"
)

func getMigrations(cmd *cobra.Command, load bool) (*migrations.Migrations, error) {
	configPath, err := cmd.Flags().GetString(flagConfig)
	if err != nil {
		return nil, err
	}

	m := migrations.New(configPath)

	if load {
		err = m.Load(cmd.Context())
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				if flagSet(cmd, flagConfig) {
					return nil, fmt.Errorf("%q file not found", configPath)
				}
			} else {
				return nil, err
			}
		}
	}

	project, err := cmd.Flags().GetString(flagProject)
	if err != nil {
		return nil, err
	}
	if project != "" {
		m.Config.ProjectId = project
	}
	if m.Config.ProjectId == "" {
		m.Config.ProjectId = os.Getenv(constants.EnvProjectID)
	}
	if m.Config.ProjectId == "" {
		m.Config.ProjectId = os.Getenv(constants.EnvGoogleCloudProject)
	}

	instanceID, err := cmd.Flags().GetString(flagInstance)
	if err != nil {
		return nil, err
	}
	if instanceID != "" {
		m.Config.InstanceId = instanceID
	}
	if m.Config.InstanceId == "" {
		m.Config.InstanceId = os.Getenv(constants.EnvInstanceID)
	}

	databaseID, err := cmd.Flags().GetString(flagDatabase)
	if err != nil {
		return nil, err
	}
	if databaseID != "" {
		m.Config.DatabaseId = databaseID
	}
	if m.Config.DatabaseId == "" {
		m.Config.DatabaseId = os.Getenv(constants.EnvDatabaseID)
	}

	if flagSet(cmd, flagEmulator) {
		emulator, err := cmd.Flags().GetBool(flagEmulator)
		if err != nil {
			return nil, err
		}

		if emulator {
			if os.Getenv(constants.EnvEmulatorHost) == "" {
				os.Setenv(constants.EnvEmulatorHost, constants.EnvEmulatorHostDefault)
			}
		} else {
			if os.Getenv(constants.EnvEmulatorHost) != "" {
				os.Unsetenv(constants.EnvEmulatorHost)
			}
		}
	}

	if load {
		err = m.Validate()
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

func getMigration(cmd *cobra.Command, ms *migrations.Migrations) (*migrations.Migration, error) {
	var id int
	var err error

	if flagSet(cmd, flagMigration) {
		id, err = parseMigrationFlag(cmd)
		if err != nil {
			return nil, err
		}
	} else {
		id = ms.LatestID()
	}

	return ms.Get(id)
}
