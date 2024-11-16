package cmd

import (
	"github.com/spf13/cobra"
)

func newAdd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add to an existing migration",
		Args:  args(),
	}
	cmd.CompletionOptions.DisableDefaultCmd = true

	cmd.AddCommand(newAddUpgrade())
	cmd.AddCommand(newAddProto())

	return cmd
}
