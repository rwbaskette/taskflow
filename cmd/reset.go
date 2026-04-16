package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/user/project/internal/db"
	"github.com/user/project/internal/service"
	cliErrors "github.com/user/project/pkg/errors"
)

var resetTimeoutMinutes int

var resetCmd = &cobra.Command{
	Use:   "reset-timedout",
	Short: "Reset timed out tasks to todo status",
	Long: `Find in-progress tasks that have exceeded the specified timeout duration
and reset them to todo status.

This command scans all tasks currently in "in_progress" status and resets
any that have been in that state longer than the specified timeout.

Use the --minutes flag to specify the timeout duration in minutes.`,
	Example: `  task reset-timedout --minutes 30
  task reset-timedout --minutes 60
  task reset-timedout -m 45`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate timeout minutes
		if resetTimeoutMinutes <= 0 {
			cliErrors.HandleError(fmt.Errorf("timeout minutes must be a positive integer"))
			return
		}

		// Initialize database
		database, err := db.NewDB(".taskflow/tasks.db")
		if err != nil {
			cliErrors.HandleError(err)
			return
		}
		defer database.Close()

		// Create reset service input
		input := service.ResetTimedOutInput{
			TimeoutMinutes: resetTimeoutMinutes,
		}

		// Execute reset timed out operation
		result, err := service.ResetTimedOut(database, input)
		if err != nil {
			cliErrors.HandleError(err)
			return
		}

		// Log affected tasks
		if len(result.ResetTasks) == 0 {
			fmt.Println("No tasks were timed out.")
			return
		}

		fmt.Printf("Reset %d timed out task(s) to todo status:\n", len(result.ResetTasks))
		for _, task := range result.ResetTasks {
			fmt.Printf("  - %s: %s (was in progress since %s)\n",
				task.ID,
				task.Title,
				task.LastUpdated.Format("2006-01-02 15:04:05"))
		}
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)

	resetCmd.Flags().IntVarP(&resetTimeoutMinutes, "minutes", "m", 30, "Timeout duration in minutes")
}
