package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/user/project/internal/db"
	"github.com/user/project/internal/service"
	cliErrors "github.com/user/project/pkg/errors"
)

var (
	completeID          string
	completeTitle       string
	completeDescription string
	completeMilestone   string
	completeStatus      string
	completeActor       string
)

var completeCmd = &cobra.Command{
	Use:   "complete",
	Short: "Mark a task as completed",
	Long: `Mark a task as completed by providing its ID.

This will update the status of the task to completed.
Use 'task list' to find task IDs.`,
	Example: `  task complete --id "1"
  task complete --id "abc123"
  task complete --id "abc123" --title "New title" --description "New description"
  task complete --id "abc123" --actor "new-actor" --milestone "M3"`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate required ID
		if err := cliErrors.ValidateID(completeID); err != nil {
			cliErrors.HandleError(err)
			return
		}

		// Validate optional fields
		if completeTitle != "" {
			if err := cliErrors.ValidateTitle(completeTitle); err != nil {
				cliErrors.HandleError(err)
				return
			}
		}

		if completeMilestone != "" {
			if err := cliErrors.ValidateMilestone(completeMilestone); err != nil {
				cliErrors.HandleError(err)
				return
			}
		}

		if completeActor != "" {
			if err := cliErrors.ValidateActor(completeActor); err != nil {
				cliErrors.HandleError(err)
				return
			}
		}

		if completeStatus != "" {
			if err := cliErrors.ValidateStatus(completeStatus); err != nil {
				cliErrors.HandleError(err)
				return
			}
		}

		// Initialize database
		database, err := db.NewDB("data/tasks.db")
		if err != nil {
			cliErrors.HandleError(err)
			return
		}
		defer database.Close()

		// Create input
		input := &service.CompleteTaskInput{
			ID:          completeID,
			Title:       completeTitle,
			Description: completeDescription,
			Milestone:   completeMilestone,
			Status:      completeStatus,
			Actor:       completeActor,
		}

		// Complete task
		result, err := service.CompleteTask(database, input)
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
	completeCmd.Flags().StringVarP(&completeTitle, "title", "t", "", "Task title")
	completeCmd.Flags().StringVarP(&completeDescription, "description", "d", "", "Task description")
	completeCmd.Flags().StringVarP(&completeStatus, "status", "s", "", "Task status (todo, in_progress, done, blocked)")
	completeCmd.Flags().StringVarP(&completeMilestone, "milestone", "m", "", "Milestone for the task")
	completeCmd.Flags().StringVarP(&completeActor, "actor", "a", "", "Actor assigned to the task")

	// Mark ID as required
	if err := completeCmd.MarkFlagRequired("id"); err != nil {
		// Log but don't fail - MarkFlagRequired can fail if flag doesn't exist
		_ = err
	}
}
