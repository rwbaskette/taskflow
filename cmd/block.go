package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/user/project/internal/db"
	"github.com/user/project/internal/service"
	cliErrors "github.com/user/project/pkg/errors"
)

var blockID string
var blockReason string

var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "Block a task",
	Long: `Block a task by providing its ID and a reason.

A blocked task cannot be worked on until it is unblocked.
Use 'task list' to find task IDs.`,
	Example: `  task block --id "1" --reason "Waiting for API documentation"
  task block --id "abc123" -r "Dependency not available"`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate required ID
		if err := cliErrors.ValidateID(blockID); err != nil {
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

		// Block task
		result, err := service.BlockTask(database, service.BlockTaskInput{
			ID:     blockID,
			Reason: blockReason,
		})
		if err != nil {
			cliErrors.HandleError(err)
			return
		}

		fmt.Printf("Task blocked successfully:\n")
		fmt.Printf("  ID: %s\n", result.ID)
		fmt.Printf("  Title: %s\n", result.Title)
		fmt.Printf("  Status: %s\n", result.Status)
	},
}

func init() {
	rootCmd.AddCommand(blockCmd)

	blockCmd.Flags().StringVarP(&blockID, "id", "i", "", "Task ID (required)")
	blockCmd.Flags().StringVarP(&blockReason, "reason", "r", "", "Reason for blocking the task (required)")

	// Mark ID and reason as required
	if err := blockCmd.MarkFlagRequired("id"); err != nil {
		_ = err
	}
	if err := blockCmd.MarkFlagRequired("reason"); err != nil {
		_ = err
	}
}
