package cmd

import (
	"fmt"

	"github.com/rwbaskette/taskflow/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the taskflow version",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "taskflow version %s\n", version.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
