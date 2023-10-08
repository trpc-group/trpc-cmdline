// Package version provides version command.
package version

import (
	"fmt"

	"github.com/spf13/cobra"

	"trpc.group/trpc-go/trpc-cmdline/config"
)

// CMD returns the version command.
func CMD() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show the version of trpc command (commit hash)",
		Long:  "Show the version of trpc command (commit hash).",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println("trpc-group/trpc-cmdline version:", config.TRPCCliVersion)
		},
	}
	return versionCmd
}
