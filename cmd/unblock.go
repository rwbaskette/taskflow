package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rwbaskette/taskflow/internal/db"
	"github.com/rwbaskette/taskflow/internal/service"
	cliErrors "github.com/rwbaskette/taskflow/pkg/errors"
)

var unblockJSON string

var unblockCmd = &cobra.Command{
	Use:     "unblock",
	Short:   "Unblock a previously blocked task",
	Long:    "Unblock a task that was previously blocked, transitioning it from 'blocked' back to 'todo' status.\n\nAn unblocked task becomes actionable again. Optionally update the description during unblocking.\nUse 'task list --status blocked' to find blocked task IDs.",
	Example: `  task unblock '{"id":"task-42"}'
  task unblock -j '{"id":"task-42","description":"Dependency resolved, ready to work"}'
  echo '{"id":"abc123"}' | task unblock -`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		database, err := db.NewDB(db.DefaultDBPath())
		if err != nil {
			cliErrors.HandleError(err)
			return
		}
		defer database.Close()

		jsonArg := unblockJSON
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

		// Validate the required 'id' parameter (first validation step)
		if _, exists := doc["id"]; !exists {
			// id parameter is missing entirely
			cliErrors.HandleError(cliErrors.MissingIDError())
			return
		}

		idStr, isString := doc["id"].(string)
		if !isString {
			// id is present but not a string type
			cliErrors.HandleError(cliErrors.NonStringIDError(doc["id"]))
			return
		}

		id := strings.TrimSpace(idStr)
		if id == "" {
			// id is present but empty string
			cliErrors.HandleError(cliErrors.EmptyIDError())
			return
		}

		description, descriptionOK := service.GetStringFieldTrim(doc, "description")

		// Validate description is a string if provided (optional parameter)
		if !descriptionOK {
			if _, exists := doc["description"]; exists {
				// Field exists but is not a string - that's an error
				cliErrors.HandleError(cliErrors.ValidationError("description", "must be a string", "Provide the description as a text string"))
				return
			}
			// Field doesn't exist - that's fine, description is optional
			description = ""
		}

		result, err := service.UnblockTask(database, service.UnblockTaskInput{
			ID:          id,
			Description: description,
		})
		if err != nil {
			cliErrors.HandleError(err)
			return
		}

		fmt.Printf("Task unblocked successfully:\n")
		fmt.Printf("  ID: %s\n", result.ID)
		fmt.Printf("  Title: %s\n", result.Title)
		fmt.Printf("  Status: %s\n", result.Status)
	},
}

func init() {
	rootCmd.AddCommand(unblockCmd)

	unblockCmd.Flags().StringVarP(&unblockJSON, "json", "j", "", "JSON document (use '-' for stdin)")
}
