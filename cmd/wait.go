package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/rwbaskette/taskflow/internal/db"
	"github.com/rwbaskette/taskflow/internal/service"
	cliErrors "github.com/rwbaskette/taskflow/pkg/errors"
)

var waitTimeout int

var waitCmd = &cobra.Command{
	Use:     "wait",
	Short:   "Wait for one or more tasks to complete",
	Long:    "Wait for one or more tasks to complete. Blocks until all specified tasks reach 'done' status or the timeout is reached.",
	Example: `  task wait "task-1" "task-2"
  task wait "task-1" --timeout 60000
  task wait "task-1" "task-2" --timeout 30000`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		database, err := db.NewDB(db.DefaultDBPath())
		if err != nil {
			cliErrors.HandleError(err)
			return
		}
		defer database.Close()

		// Validate task IDs
		for _, id := range args {
			if err := cliErrors.ValidateID(id); err != nil {
				cliErrors.HandleError(err)
				return
			}
		}

		input := &service.WaitTaskInput{
			TaskIDs: args,
			Timeout: waitTimeout,
		}

		results, err := service.WaitTask(database, input)
		if err != nil {
			cliErrors.HandleError(err)
			return
		}

		// Output results
		outputResults(results, cmd)
	},
}

func outputResults(results []service.WaitTaskResult, cmd *cobra.Command) {
	// Check if JSON output is requested
	jsonFormat, _ := cmd.Flags().GetBool("json")

	if jsonFormat {
		// JSON output
		jsonBytes, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			cliErrors.HandleError(err)
			return
		}
		fmt.Println(string(jsonBytes))
	} else {
		// Human-readable output
		allCompleted := true
		for _, result := range results {
			if result.State != "completed" {
				allCompleted = false
				break
			}
		}

		if allCompleted {
			fmt.Println("All tasks completed:")
		} else {
			fmt.Println("Wait completed with the following results:")
		}

		for _, result := range results {
			state := result.State
			if state == "completed" {
				state = "✓ " + state
			} else if state == "timed_out" {
				state = "⏱ " + state
			}
			fmt.Printf("  [%s] %s (ID: %s)\n", state, result.Title, result.TaskID)
		}

		// Set exit code based on results
		if !allCompleted {
			os.Exit(1)
		}
	}
}

func init() {
	rootCmd.AddCommand(waitCmd)

	waitCmd.Flags().IntVarP(&waitTimeout, "timeout", "t", 0, "Timeout in milliseconds (0 = wait forever)")
	waitCmd.Flags().Bool("json", false, "Output results as JSON")
}