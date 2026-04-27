package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	version = "0.1.0"
	commit  = ""
	date    = ""
)

var rootCmd = &cobra.Command{
	Use:   "taskflow",
	Short: "A task management CLI tool",
	Long: `Task is a CLI tool for managing tasks with support for
adding, updating, completing, blocking, listing, and resetting tasks.

For more information, visit the project documentation.

Environment Variables:
  TASKFLOW_DIR  Custom directory for the database (default: .taskflow)
                  When set, the database will be created at $TASKFLOW_DIR/tasks.db
                  Example: TASKFLOW_DIR=/tmp/work/taskflow taskflow list`,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize()

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	// Set version - Cobra handles --version flag automatically
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("Task CLI version: {{.Version}}\n")
}
