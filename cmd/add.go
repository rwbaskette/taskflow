package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/user/project/internal/db"
	"github.com/user/project/internal/service"
	cliErrors "github.com/user/project/pkg/errors"
)

var (
	addTitle       string
	addDescription string
	addMilestone   string
	addActor       string
	addID          string
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new task",
	Long: `Add a new task to the task list.

The id, milestone, title and description are required.
After adding a task, use 'task list' to see all tasks.`,
	Example: `  task add --id "1" --title "Implement login feature" --milestone "v1.0" --description "Add login functionality"
  task add --id "2" --title "Fix memory leak" --description "Memory leak in data processing" --milestone "v2.0"
  task add --id "3" --title "Deploy to production" --milestone "v2.0 Release" --actor "devops"`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate required fields
		if err := cliErrors.ValidateID(addID); err != nil {
			cliErrors.HandleError(err)
			return
		}
		if err := cliErrors.ValidateMilestone(addMilestone); err != nil {
			cliErrors.HandleError(err)
			return
		}
		if err := cliErrors.ValidateTitle(addTitle); err != nil {
			cliErrors.HandleError(err)
			return
		}

		// Description is required
		if addDescription == "" {
			cliErrors.HandleError(cliErrors.MissingArgumentError("description", "--description is required"))
			return
		}

		// Initialize database
		database, err := db.NewDB("data/tasks.db")
		if err != nil {
			cliErrors.HandleError(err)
			return
		}
		defer database.Close()

		// Create input
		input := &service.AddTaskInput{
			ID:          addID,
			Milestone:   addMilestone,
			Title:       addTitle,
			Description: addDescription,
			Actor:       addActor,
		}

		// Add task
		result, err := service.AddTask(database, input)
		if err != nil {
			cliErrors.HandleError(err)
			return
		}

		fmt.Printf("Task added successfully:\n")
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
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVarP(&addID, "id", "i", "", "Task ID (required)")
	addCmd.Flags().StringVarP(&addTitle, "title", "t", "", "Task title (required)")
	addCmd.Flags().StringVarP(&addDescription, "description", "d", "", "Task description (required)")
	addCmd.Flags().StringVarP(&addMilestone, "milestone", "m", "", "Milestone for the task (required)")
	addCmd.Flags().StringVarP(&addActor, "actor", "a", "", "Actor assigned to the task")
}
