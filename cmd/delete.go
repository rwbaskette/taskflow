package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/user/project/internal/db"
	"github.com/user/project/internal/service"
	cliErrors "github.com/user/project/pkg/errors"
)

var deleteID string

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Soft delete a task",
	Long: `Soft delete a task by moving it to the deleted_tasks table.

The task is moved to a deleted_tasks table with a deleted_on timestamp
rather than being permanently removed. Use 'task list' to find task IDs.`,
	Example: `  task delete --id "1"
  task delete --id "abc123"`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cliErrors.ValidateID(deleteID); err != nil {
			cliErrors.HandleError(err)
			return
		}

		database, err := db.NewDB(".taskflow/tasks.db")
		if err != nil {
			cliErrors.HandleError(err)
			return
		}
		defer database.Close()

		input := &service.DeleteTaskInput{
			ID: deleteID,
		}

		result, err := service.DeleteTask(database, input)
		if err != nil {
			cliErrors.HandleError(err)
			return
		}

		fmt.Printf("Task deleted successfully:\n")
		fmt.Printf("  ID: %s\n", result.ID)
		fmt.Printf("  Title: %s\n", result.Title)
		fmt.Printf("  Deleted On: %s\n", result.DeletedOn.Format("2006-01-02 15:04:05"))
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringVarP(&deleteID, "id", "i", "", "Task ID (required)")

	if err := deleteCmd.MarkFlagRequired("id"); err != nil {
		_ = err
	}
}
