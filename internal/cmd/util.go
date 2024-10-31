package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/silas/jimmy/internal/constants"
	"github.com/silas/jimmy/internal/migrations"
	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

func args(checkArgs ...string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > len(checkArgs) {
			return fmt.Errorf(
				"unrecognized arguments: %s",
				strings.Join(args[len(checkArgs):], " "),
			)
		}
		if len(checkArgs) > len(args) {
			return fmt.Errorf("%s required", checkArgs[len(args)])
		}
		return nil
	}
}

func flagSet(cmd *cobra.Command, name string) bool {
	flag := cmd.Flag(name)
	return flag != nil && flag.Changed
}

func newMigrations(cmd *cobra.Command, load bool) (*migrations.Migrations, error) {
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

type migrationFlags struct {
	SQL      string
	Template jimmyv1.Template
	Env      jimmyv1.Environment
	Type     jimmyv1.Type
}

func parseMigrationFlags(cmd *cobra.Command) (flags migrationFlags, err error) {
	flags.SQL, err = cmd.Flags().GetString(flagSQL)
	if err != nil {
		return
	}

	templateValue, err := cmd.Flags().GetString(flagTemplate)
	if err != nil {
		return
	}

	if templateValue != "" {
		templateValue = strings.ToUpper(migrations.Slugify(templateValue))

		templateInt, found := jimmyv1.Template_value[templateValue]
		if !found {
			return flags, fmt.Errorf("%q is not a valid template", templateValue)
		}

		flags.Template = jimmyv1.Template(templateInt)
	}

	envValue, err := cmd.Flags().GetString(flagEnv)
	if err != nil {
		return flags, err
	}

	if envValue != "" {
		envValue = strings.ToUpper(migrations.Slugify(envValue))

		envInt, found := jimmyv1.Environment_value[envValue]
		if !found {
			return flags, fmt.Errorf("%q is not a valid env", envValue)
		}

		flags.Env = jimmyv1.Environment(envInt)
	}

	typeValue, err := cmd.Flags().GetString(flagType)
	if err != nil {
		return flags, err
	}

	if typeValue != "" {
		typeInt, found := jimmyv1.Type_value[typeValue]
		if !found {
			return flags, fmt.Errorf("%q is not a valid type", envValue)
		}

		flags.Type = jimmyv1.Type(typeInt)
	}

	return flags, nil
}

func setupMigrationFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(flagSQL, "s", "", "migration SQL")
	cmd.Flags().StringP(flagEnv, "e", "", "execution environment (GOOGLE_CLOUD, EMULATOR)")
	cmd.Flags().StringP(flagTemplate, "t", "", "SQL template")
	cmd.Flags().StringP(flagType, "", "", "type of statement (DDL, DML, PARTITIONED_DML)")
}

func outputEnums[d ~int32](
	cmd *cobra.Command,
	enum map[int32]string,
	skipFirst bool,
	defaultValue d,
) {
	for i := range int32(len(enum)) {
		if i == 0 && skipFirst {
			continue
		}

		name := enum[i]
		name = strings.ToLower(name)
		name = strings.ReplaceAll(name, "_", "-")

		if i == int32(defaultValue) {
			cmd.Println(name, "(default)")
		} else {
			cmd.Println(name)
		}
	}
}
