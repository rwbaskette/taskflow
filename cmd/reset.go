package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rwbaskette/taskflow/internal/service"
)

var resetJSON string

var resetCmd = &cobra.Command{
	Use:   "reset-timedout",
	Short: "Reset timed out tasks to todo status",
	Long:  "Find in-progress tasks that have exceeded the specified timeout duration and reset them to todo status.\n\nThis command scans all tasks currently in 'in_progress' status and resets any that have been in that state longer than the specified timeout.",
	Example: `  task reset-timedout '{"minutes":30}'
  echo '{"minutes":60}' | task reset-timedout -
  task reset-timedout -`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		doc := jsonDoc(resetJSON, args, "{}")

		resetTimeoutMinutes := 30
		if v, ok := service.GetNumberField(doc, "minutes"); ok {
			resetTimeoutMinutes = int(v)
		}

		if resetTimeoutMinutes <= 0 {
			fatal(errors.New("timeout minutes must be a positive integer"))
		}

		database := openDB()
		defer database.Close()

		resetTasks, err := service.ResetTimedOut(database, resetTimeoutMinutes)
		if err != nil {
			fatal(err)
		}

		if len(resetTasks) == 0 {
			fmt.Println("No tasks were timed out.")
			return
		}

		fmt.Printf("Reset %d timed out task(s) to todo status:\n", len(resetTasks))
		for _, task := range resetTasks {
			fmt.Printf("  - %s: %s (was in progress since %s)\n",
				task.ID,
				task.Title,
				task.LastUpdated.Format("2006-01-02 15:04:05"))
		}
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)

	resetCmd.Flags().StringVarP(&resetJSON, "json", "j", "", "JSON document (use '-' for stdin)")
}
