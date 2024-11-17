package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
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

func displayDuration(t time.Time) string {
	return fmt.Sprintf("in %s", time.Since(t).Round(time.Millisecond))
}
