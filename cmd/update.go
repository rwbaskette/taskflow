package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rwbaskette/taskflow/internal/db"
	"github.com/rwbaskette/taskflow/internal/service"
	cliErrors "github.com/rwbaskette/taskflow/pkg/errors"
)

var updateJSON string

var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update an existing task",
	Long:    "Update an existing task by its ID.\n\nThe update can be specified as a JSON document via argument or stdin.\nFields: id (required), title, description, status, milestone, actor.",
	Example: `  task update '{"id":"1","title":"New title"}'
  echo '{"id":"1","status":"in_progress"}' | task update -
  task update -`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		database, err := db.NewDB(db.DefaultDBPath())
		if err != nil {
			cliErrors.HandleError(err)
			return
		}
		defer database.Close()

		jsonArg := updateJSON
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
		description, hasDesc := service.GetStringField(doc, "description")
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

		if !hasTitle && !hasDesc && !hasStatus && !hasMilestone && !hasActor {
			cliErrors.HandleError(cliErrors.MissingArgumentError("update field", "at least one of title, description, status, milestone, or actor is required in JSON"))
			return
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

	updateCmd.Flags().StringVarP(&updateJSON, "json", "j", "", "JSON document (use '-' for stdin)")
}