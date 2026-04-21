package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/user/project/internal/db"
	"github.com/user/project/internal/service"
	cliErrors "github.com/user/project/pkg/errors"
)

var addJSON string

var addCmd = &cobra.Command{
	Use:     "add",
	Short:   "Add a new task",
	Long:    "Add a new task to the task list.\n\nThe task can be specified as a JSON document via argument or stdin.\nFields: id, milestone, title, description, actor (all fields except description are required).",
	Example: `  task add '{"id":"1","title":"Implement login","milestone":"v1","description":"Add login"}'
  echo '{"id":"2","title":"Fix bug","milestone":"v1","description":"Fix memory leak"}' | task add -
  task add -`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		database, err := db.NewDB(".taskflow/tasks.db")
		if err != nil {
			cliErrors.HandleError(err)
			return
		}
		defer database.Close()

		jsonArg := addJSON
		if jsonArg == "" && len(args) > 0 {
			jsonArg = args[0]
		}
		if jsonArg == "" {
			cliErrors.HandleError(cliErrors.MissingArgumentError("json", "provide JSON document via argument or stdin"))
			return
		}

		doc, err := service.ParseJSONFromArg(jsonArg)
		if err != nil {
			cliErrors.HandleError(err)
			return
		}

		id, _ := service.GetStringFieldTrim(doc, "id")
		milestone, _ := service.GetStringFieldTrim(doc, "milestone")
		title, _ := service.GetStringFieldTrim(doc, "title")
		description, _ := service.GetStringFieldTrim(doc, "description")
		actor, _ := service.GetStringFieldTrim(doc, "actor")

		if err := cliErrors.ValidateID(id); err != nil {
			cliErrors.HandleError(err)
			return
		}
		if err := cliErrors.ValidateMilestone(milestone); err != nil {
			cliErrors.HandleError(err)
			return
		}
		if err := cliErrors.ValidateTitle(title); err != nil {
			cliErrors.HandleError(err)
			return
		}
		if description == "" {
			cliErrors.HandleError(cliErrors.MissingArgumentError("description", "description is required in JSON document"))
			return
		}

		input := &service.AddTaskInput{
			ID:          id,
			Milestone:   milestone,
			Title:       title,
			Description: description,
			Actor:       actor,
		}

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

	if os.Getenv("TF_TEST_STDIN") == "1" {
		addCmd.Flags().StringVarP(&addJSON, "json", "j", "", "JSON document (use '-' for stdin)")
	} else {
		addCmd.Flags().StringVarP(&addJSON, "json", "j", "", "JSON document (use '-' for stdin)")
	}
}