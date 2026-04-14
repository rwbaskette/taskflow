package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/user/project/internal/db"
	"github.com/user/project/internal/service"
	cliErrors "github.com/user/project/pkg/errors"
)

var completeID string

var completeCmd = &cobra.Command{
	Use:   "complete",
	Short: "Mark a task as completed",
	Long: `Mark a task as completed by providing its ID.

This will update the status of the task to completed.
Use 'task list' to find task IDs.`,
	Example: `  task complete --id "1"
  task complete --id "abc123"`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate required ID
		if err := cliErrors.ValidateID(completeID); err != nil {
			cliErrors.HandleError(err)
			return
		}

		// Initialize database
		database, err := db.NewDB("data/tasks.db")
		if err != nil {
			cliErrors.HandleError(err)
			return
		}
		defer database.Close()

		// Complete task
		result, err := service.CompleteTask(database, completeID)
		if err != nil {
			cliErrors.HandleError(err)
			return
		}

		fmt.Printf("Task completed successfully:\n")
		fmt.Printf("  ID: %s\n", result.ID)
		fmt.Printf("  Title: %s\n", result.Title)
		fmt.Printf("  Status: %s\n", result.Status)
	},
}

func init() {
	rootCmd.AddCommand(completeCmd)

	completeCmd.Flags().StringVarP(&completeID, "id", "i", "", "Task ID (required)")

	// Mark ID as required
	if err := completeCmd.MarkFlagRequired("id"); err != nil {
		// Log but don't fail - MarkFlagRequired can fail if flag doesn't exist
		_ = err
	}
}
