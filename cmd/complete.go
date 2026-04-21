package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/user/project/internal/db"
	"github.com/user/project/internal/service"
	cliErrors "github.com/user/project/pkg/errors"
)

var completeJSON string

var completeCmd = &cobra.Command{
	Use:     "complete",
	Short:   "Mark a task as completed",
	Long:    "Mark a task as completed by providing its ID.\n\nThe completion can be specified as a JSON document via argument or stdin.\nFields: id (required), title, description, status, milestone, actor.",
	Example: `  task complete '{"id":"1"}'
  echo '{"id":"1","actor":"new-owner"}' | task complete -
  task complete -`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		database, err := db.NewDB(".taskflow/tasks.db")
		if err != nil {
			cliErrors.HandleError(err)
			return
		}
		defer database.Close()

		jsonArg := completeJSON
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
		title, hasTitle := service.GetStringField(doc, "title")
		description, _ := service.GetStringField(doc, "description")
		status, hasStatus := service.GetStringField(doc, "status")
		milestone, hasMilestone := service.GetStringField(doc, "milestone")
		actor, hasActor := service.GetStringField(doc, "actor")

		if err := cliErrors.ValidateID(id); err != nil {
			cliErrors.HandleError(err)
			return
		}

		if hasTitle {
			if err := cliErrors.ValidateTitle(title); err != nil {
				cliErrors.HandleError(err)
				return
			}
		}
		if hasMilestone {
			if err := cliErrors.ValidateMilestone(milestone); err != nil {
				cliErrors.HandleError(err)
				return
			}
		}
		if hasActor {
			if err := cliErrors.ValidateActor(actor); err != nil {
				cliErrors.HandleError(err)
				return
			}
		}
		if hasStatus {
			if err := cliErrors.ValidateStatus(status); err != nil {
				cliErrors.HandleError(err)
				return
			}
		}

		input := &service.CompleteTaskInput{
			ID:          id,
			Title:       title,
			Description: description,
			Milestone:   milestone,
			Status:      status,
			Actor:       actor,
		}

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

	completeCmd.Flags().StringVarP(&completeJSON, "json", "j", "", "JSON document (use '-' for stdin)")
}