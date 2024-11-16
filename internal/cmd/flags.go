package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/silas/jimmy/internal/migrations"
	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

const (
	flagBootstrap = "bootstrap"
	flagEnv       = "env"
	flagMigration = "migration"
	flagSQL       = "sql"
	flagSquash    = "squash"
	flagTemplate  = "template"
	flagType      = "type"
)

func setupMigrationFlag(cmd *cobra.Command) {
	cmd.Flags().IntP(flagMigration, "m", 0, "migration ID")
}

func parseMigrationFlag(cmd *cobra.Command) (int, error) {
	return cmd.Flags().GetInt(flagMigration)
}

func setupStatementFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(flagSQL, "s", "", "migration SQL")
	cmd.Flags().StringP(flagEnv, "e", "", "execution environment (GOOGLE_CLOUD, EMULATOR)")
	cmd.Flags().StringP(flagTemplate, "t", "", "SQL template")
	cmd.Flags().StringP(flagType, "", "", "type of statement (DDL, DML, PARTITIONED_DML)")
}

type statementFlags struct {
	SQL      string
	Template string
	Env      jimmyv1.Environment
	Type     jimmyv1.Type
}

func parseStatementFlags(cmd *cobra.Command) (flags statementFlags, err error) {
	flags.SQL, err = cmd.Flags().GetString(flagSQL)
	if err != nil {
		return
	}

	flags.Template, err = cmd.Flags().GetString(flagTemplate)
	if err != nil {
		return
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
