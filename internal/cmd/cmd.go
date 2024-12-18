package cmd

import (
	"github.com/spf13/cobra"

	"github.com/silas/jimmy/internal/constants"
)

const (
	flagConfig   = "config"
	flagEmulator = "emulator"
	flagProject  = "project"
	flagInstance = "instance"
	flagDatabase = "database"
)

func New() *cobra.Command {
	cobra.EnableCommandSorting = false

	cmd := &cobra.Command{
		Use:           constants.AppName,
		Version:       constants.Version,
		SilenceUsage:  false,
		SilenceErrors: true,
	}
	cmd.CompletionOptions.DisableDefaultCmd = true

	cmd.PersistentFlags().StringP(flagConfig, "c", constants.ConfigFile, "configuration file")
	cmd.PersistentFlags().BoolP(flagEmulator, "", false, "set whether to enable emulator mode (default automatically detected)")
	cmd.PersistentFlags().StringP(flagProject, "p", "", "set Google project ID")
	cmd.PersistentFlags().StringP(flagInstance, "i", "", "set Spanner instance ID")
	cmd.PersistentFlags().StringP(flagDatabase, "d", "", "set Spanner database ID")

	cmd.AddCommand(newInit())
	cmd.AddCommand(newCreate())
	cmd.AddCommand(newAdd())
	cmd.AddCommand(newUpgrade())
	cmd.AddCommand(newTemplates())

	return cmd
}
