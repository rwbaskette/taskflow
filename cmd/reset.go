package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rwbaskette/taskflow/internal/db"
	"github.com/rwbaskette/taskflow/internal/service"
	cliErrors "github.com/rwbaskette/taskflow/pkg/errors"
)

var resetJSON string

var resetCmd = &cobra.Command{
	Use:     "reset-timedout",
	Short:   "Reset timed out tasks to todo status",
	Long:    "Find in-progress tasks that have exceeded the specified timeout duration and reset them to todo status.\n\nThis command scans all tasks currently in 'in_progress' status and resets any that have been in that state longer than the specified timeout.",
	Example: `  task reset-timedout '{"minutes":30}'
  echo '{"minutes":60}' | task reset-timedout -
  task reset-timedout -`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		jsonArg := resetJSON
		if jsonArg == "" && len(args) > 0 {
			jsonArg = args[0]
		}
		if jsonArg == "" {
			jsonArg = "{}"
		}

		doc, err := service.ParseJSONFromArg(jsonArg)
		if err != nil {
			cliErrors.HandleError(err)
			return
		}

		resetTimeoutMinutes := 30
		if v, ok := service.GetNumberField(doc, "minutes"); ok {
			resetTimeoutMinutes = int(v)
		}

		if resetTimeoutMinutes <= 0 {
			cliErrors.HandleError(fmt.Errorf("timeout minutes must be a positive integer"))
			return
		}

		database, err := db.NewDB(db.DefaultDBPath())
		if err != nil {
			cliErrors.HandleError(err)
			return
		}
		defer database.Close()

		input := service.ResetTimedOutInput{
			TimeoutMinutes: resetTimeoutMinutes,
		}

		result, err := service.ResetTimedOut(database, input)
		if err != nil {
			cliErrors.HandleError(err)
			return
		}

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

	resetCmd.Flags().StringVarP(&resetJSON, "json", "j", "", "JSON document (use '-' for stdin)")
}
