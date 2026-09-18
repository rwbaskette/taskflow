package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/rwbaskette/taskflow/internal/clierr"
	"github.com/rwbaskette/taskflow/internal/service"
)

var unblockJSON string

var unblockCmd = &cobra.Command{
	Use:   "unblock",
	Short: "Unblock a previously blocked task",
	Long:  "Unblock a task that was previously blocked, transitioning it from 'blocked' back to 'todo' status.\n\nAn unblocked task becomes actionable again. Optionally update the description during unblocking.\nUse 'task list --status blocked' to find blocked task IDs.",
	Example: `  task unblock '{"id":"task-42"}'
  task unblock -j '{"id":"task-42","description":"Dependency resolved, ready to work"}'
  echo '{"id":"abc123"}' | task unblock -`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		database := openDB()
		defer database.Close()

		doc, err := jsonDoc(unblockJSON, args, "")
		if err != nil {
			fatal(err)
		}

		// Validate the required 'id' parameter (first validation step).
		id, err := service.GetIDField(doc)
		if err != nil {
			fatal(err)
		}

		// Validate the optional description parameter. The key must be a string
		// when present: a value that trims to empty preserves the stored
		// description, as does an absent key. A non-string value is an error.
		description := ""
		if val, exists := doc["description"]; exists {
			s, isString := val.(string)
			if !isString {
				fatal(clierr.ValidationError("description", "must be a string", "Provide the description as a text string"))
			}
			description = strings.TrimSpace(s)
		}

		result, err := service.UnblockTask(database, id, description)
		if err != nil {
			fatal(err)
		}

		printStatusResult("unblocked", result)
	},
}

func init() {
	rootCmd.AddCommand(unblockCmd)

	unblockCmd.Flags().StringVarP(&unblockJSON, "json", "j", "", "JSON document (use '-' for stdin)")
}
