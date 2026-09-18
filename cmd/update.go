package cmd

import (
	"github.com/spf13/cobra"

	cliErrors "github.com/rwbaskette/taskflow/internal/errors"
	"github.com/rwbaskette/taskflow/internal/service"
)

var updateJSON string

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing task",
	Long:  "Update an existing task by its ID.\n\nThe update can be specified as a JSON document via argument or stdin.\nFields: id (required), title, description, status, milestone, actor.",
	Example: `  task update '{"id":"1","title":"New title"}'
  echo '{"id":"1","status":"in_progress"}' | task update -
  task update -`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		database := openDB()
		defer database.Close()

		doc, err := jsonDoc(updateJSON, args, "")
		if err != nil {
			fatal(err)
		}

		id, _ := service.GetStringFieldTrim(doc, "id")
		title, hasTitle := service.GetStringFieldTrim(doc, "title")
		description, hasDesc := service.GetStringFieldTrim(doc, "description")
		status, hasStatus := service.GetStringFieldTrim(doc, "status")
		milestone, hasMilestone := service.GetStringFieldTrim(doc, "milestone")
		actor, hasActor := service.GetStringFieldTrim(doc, "actor")

		if err := cliErrors.ValidateID(id); err != nil {
			fatal(err)
		}

		validateOptionalTaskFields(title, milestone, actor)

		if hasStatus {
			status = normalizeStatus(status)
		}

		if !hasTitle && !hasDesc && !hasStatus && !hasMilestone && !hasActor {
			fatal(cliErrors.MissingArgumentError("update field", "at least one of title, description, status, milestone, or actor is required in JSON"))
		}

		input := &service.UpdateTaskInput{
			ID:          id,
			Title:       title,
			Description: description,
			Milestone:   milestone,
			Status:      status,
			Actor:       actor,
		}

		result, err := service.UpdateTask(database, input)
		if err != nil {
			fatal(err)
		}

		printTaskResult("updated", result)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringVarP(&updateJSON, "json", "j", "", "JSON document (use '-' for stdin)")
}
