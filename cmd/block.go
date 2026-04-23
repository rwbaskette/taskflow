package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rwbaskette/taskflow/internal/db"
	"github.com/rwbaskette/taskflow/internal/service"
	cliErrors "github.com/rwbaskette/taskflow/pkg/errors"
)

var blockJSON string

var blockCmd = &cobra.Command{
	Use:     "block",
	Short:   "Block a task",
	Long:    "Block a task by providing its ID and a reason.\n\nA blocked task cannot be worked on until it is unblocked.\nUse 'task list' to find task IDs.",
	Example: `  task block '{"id":"1","reason":"Waiting for API documentation"}'
  echo '{"id":"abc123","reason":"Dependency not available"}' | task block -
  task block -`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		database, err := db.NewDB(".taskflow/tasks.db")
		if err != nil {
			cliErrors.HandleError(err)
			return
		}
		defer database.Close()

		jsonArg := blockJSON
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
		reason, _ := service.GetStringFieldTrim(doc, "reason")

		if err := cliErrors.ValidateID(id); err != nil {
			cliErrors.HandleError(err)
			return
		}

		if strings.TrimSpace(reason) == "" {
			cliErrors.HandleError(cliErrors.MissingArgumentError("reason", "reason is required in JSON document"))
			return
		}

		result, err := service.BlockTask(database, service.BlockTaskInput{
			ID:     id,
			Reason: reason,
		})
		if err != nil {
			cliErrors.HandleError(err)
			return
		}

		fmt.Printf("Task blocked successfully:\n")
		fmt.Printf("  ID: %s\n", result.ID)
		fmt.Printf("  Title: %s\n", result.Title)
		fmt.Printf("  Status: %s\n", result.Status)
	},
}

func init() {
	rootCmd.AddCommand(blockCmd)

	blockCmd.Flags().StringVarP(&blockJSON, "json", "j", "", "JSON document (use '-' for stdin)")
}