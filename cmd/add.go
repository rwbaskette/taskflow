package cmd

import (
	"github.com/spf13/cobra"

	"github.com/rwbaskette/taskflow/internal/db"
	cliErrors "github.com/rwbaskette/taskflow/internal/errors"
	"github.com/rwbaskette/taskflow/internal/service"
)

var addJSON string

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new task",
	Long:  "Add a new task to the task list.\n\nThe task can be specified as a JSON document via argument or stdin.\nFields: id, milestone, title, description, actor (all fields except description are required).",
	Example: `  task add '{"id":"1","title":"Implement login","milestone":"v1","description":"Add login"}'
  echo '{"id":"2","title":"Fix bug","milestone":"v1","description":"Fix memory leak"}' | task add -
  task add -`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		database := openDB()
		defer database.Close()

		doc, err := jsonDoc(addJSON, args, "")
		if err != nil {
			fatal(err)
		}

		id, _ := service.GetStringFieldTrim(doc, "id")
		milestone, _ := service.GetStringFieldTrim(doc, "milestone")
		title, _ := service.GetStringFieldTrim(doc, "title")
		description, _ := service.GetStringFieldTrim(doc, "description")
		actor, _ := service.GetStringFieldTrim(doc, "actor")

		if err := cliErrors.ValidateID(id); err != nil {
			fatal(err)
		}
		if err := cliErrors.ValidateMilestone(milestone); err != nil {
			fatal(err)
		}
		if err := cliErrors.ValidateTitle(title); err != nil {
			fatal(err)
		}
		if description == "" {
			fatal(cliErrors.MissingArgumentError("description", "description is required in JSON document"))
		}

		task := &db.Task{
			ID:          id,
			Milestone:   milestone,
			Title:       title,
			Description: description,
			Status:      "todo",
			Actor:       actor,
		}

		if err := database.CreateTask(task); err != nil {
			fatal(err)
		}

		printTaskResult("added", task)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVarP(&addJSON, "json", "j", "", "JSON document (use '-' for stdin)")
}
