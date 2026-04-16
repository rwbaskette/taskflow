package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/user/project/internal/db"
	"github.com/user/project/internal/service"
	cliErrors "github.com/user/project/pkg/errors"
)

var (
	updateID          string
	updateTitle       string
	updateDescription string
	updateStatus      string
	updateActor       string
	updateMilestone   string
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing task",
	Long: `Update an existing task by its ID.

You can update the title, description, status, milestone, and/or actor.
At least one update field must be provided.`,
	Example: `  task update --id "1" --title "New title"
  task update --id "abc" --description "Updated description" --milestone "v2.0"
  task update --id "1" --status "in_progress" --actor "new-owner"
  task update --id "1" --milestone "v2.0" --description "New description"`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate required ID
		if err := cliErrors.ValidateID(updateID); err != nil {
			cliErrors.HandleError(err)
			return
		}

		// Validate optional fields (only if provided)
		if updateTitle != "" {
			if err := cliErrors.ValidateTitle(updateTitle); err != nil {
				cliErrors.HandleError(err)
				return
			}
		}

		if updateMilestone != "" {
			if err := cliErrors.ValidateMilestone(updateMilestone); err != nil {
				cliErrors.HandleError(err)
				return
			}
		}

		if updateActor != "" {
			if err := cliErrors.ValidateActor(updateActor); err != nil {
				cliErrors.HandleError(err)
				return
			}
		}

		if updateStatus != "" {
			if err := cliErrors.ValidateStatus(updateStatus); err != nil {
				cliErrors.HandleError(err)
				return
			}
		}

		// Check that at least one update field is provided
		if updateTitle == "" && updateDescription == "" && updateStatus == "" && updateMilestone == "" && updateActor == "" {
			cliErrors.HandleError(cliErrors.MissingArgumentError("update field", "at least one of --title, --description, --status, --milestone, or --actor is required"))
			return
		}

		// Initialize database
		database, err := db.NewDB(".taskflow/tasks.db")
		if err != nil {
			cliErrors.HandleError(err)
			return
		}
		defer database.Close()

		// Create input
		input := &service.UpdateTaskInput{
			ID:          updateID,
			Title:       updateTitle,
			Description: updateDescription,
			Milestone:   updateMilestone,
			Status:      updateStatus,
			Actor:       updateActor,
		}

		// Update task
		result, err := service.UpdateTask(database, input)
		if err != nil {
			cliErrors.HandleError(err)
			return
		}

		fmt.Printf("Task updated successfully:\n")
		fmt.Printf("  ID: %s\n", result.ID)
		fmt.Printf("  Title: %s\n", result.Title)
		fmt.Printf("  Description: %s\n", result.Description)
		fmt.Printf("  Milestone: %s\n", result.Milestone)
		if result.Actor != "" {
			fmt.Printf("  Actor: %s\n", result.Actor)
		}
		fmt.Printf("  Status: %s\n", result.Status)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringVarP(&updateID, "id", "i", "", "Task ID (required)")
	updateCmd.Flags().StringVarP(&updateTitle, "title", "t", "", "New task title")
	updateCmd.Flags().StringVarP(&updateDescription, "description", "d", "", "New task description")
	updateCmd.Flags().StringVarP(&updateStatus, "status", "s", "", "New task status (todo, in_progress, done, blocked)")
	updateCmd.Flags().StringVarP(&updateMilestone, "milestone", "m", "", "New milestone for the task")
	updateCmd.Flags().StringVarP(&updateActor, "actor", "a", "", "New actor assigned to the task")

	// Mark ID as required
	if err := updateCmd.MarkFlagRequired("id"); err != nil {
		// Log but don't fail - MarkFlagRequired can fail if flag doesn't exist
		_ = err
	}
}
