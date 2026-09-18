package cmd

import (
	"github.com/spf13/cobra"

	"github.com/rwbaskette/taskflow/internal/clierr"
	"github.com/rwbaskette/taskflow/internal/service"
)

var completeJSON string

var completeCmd = &cobra.Command{
	Use:   "complete",
	Short: "Mark a task as completed",
	Long:  "Mark a task as completed by providing its ID.\n\nThe completion can be specified as a JSON document via argument or stdin.\nFields: id (required), title, description, status, milestone, actor.",
	Example: `  task complete '{"id":"1"}'
  echo '{"id":"1","actor":"new-owner"}' | task complete -
  task complete -`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		database := openDB()
		defer database.Close()

		doc, err := jsonDoc(completeJSON, args, "")
		if err != nil {
			fatal(err)
		}

		id, _ := service.GetStringFieldTrim(doc, "id")
		title, _ := service.GetStringFieldTrim(doc, "title")
		description, _ := service.GetStringFieldTrim(doc, "description")
		statusInput, hasStatus := service.GetStringFieldTrim(doc, "status")
		milestone, _ := service.GetStringFieldTrim(doc, "milestone")
		actor, _ := service.GetStringFieldTrim(doc, "actor")

		if err := clierr.ValidateID(id); err != nil {
			fatal(err)
		}

		validateOptionalTaskFields(title, milestone, actor)

		// The completion status defaults to "done" when no override is given.
		status := "done"
		if hasStatus {
			status = normalizeStatus(statusInput)
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

		printStatusResult("completed", result)
	},
}

func init() {
	rootCmd.AddCommand(completeCmd)

	completeCmd.Flags().StringVarP(&completeJSON, "json", "j", "", "JSON document (use '-' for stdin)")
}
