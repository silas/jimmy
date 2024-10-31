package cmd

import (
	"github.com/spf13/cobra"
)

func newShow() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show options",
		Args:  args(),
	}
	cmd.CompletionOptions.DisableDefaultCmd = true

	cmd.AddCommand(newEnvironments())
	cmd.AddCommand(newTemplates())

	return cmd
}
