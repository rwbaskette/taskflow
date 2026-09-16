package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	cliErrors "github.com/rwbaskette/taskflow/internal/errors"
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

		jsonArg := unblockJSON
		if jsonArg == "" && len(args) > 0 {
			jsonArg = args[0]
		}
		if jsonArg == "" {
			fatal(cliErrors.MissingArgumentError("json", "provide JSON document via argument or stdin"))
		}

		doc, err := service.ParseJSONFromArg(jsonArg)
		if err != nil {
			fatal(err)
		}

		// Validate the required 'id' parameter (first validation step).
		id, err := service.GetIDField(doc)
		if err != nil {
			fatal(err)
		}

		// Validate the optional description parameter. The key must be a
		// string when present: an empty or whitespace-only string preserves
		// the stored description, as does an absent key. A non-string value
		// is an error.
		description := ""
		if val, exists := doc["description"]; exists {
			s, isString := val.(string)
			if !isString {
				fatal(cliErrors.ValidationError("description", "must be a string", "Provide the description as a text string"))
			}
			if strings.TrimSpace(s) != "" {
				description = s
			}
		}

		result, err := service.UnblockTask(database, id, description)
		if err != nil {
			fatal(err)
		}

		fmt.Println("Task unblocked successfully:")
		fmt.Printf("  ID: %s\n", result.ID)
		fmt.Printf("  Title: %s\n", result.Title)
		fmt.Printf("  Status: %s\n", result.Status)
	},
}

func init() {
	rootCmd.AddCommand(unblockCmd)

	unblockCmd.Flags().StringVarP(&unblockJSON, "json", "j", "", "JSON document (use '-' for stdin)")
}
