package cmd

import (
	"github.com/spf13/cobra"

	cliErrors "github.com/rwbaskette/taskflow/internal/errors"
	"github.com/rwbaskette/taskflow/internal/service"
)

var blockJSON string

var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "Block a task",
	Long:  "Block a task by providing its ID and a reason.\n\nA blocked task cannot be worked on until it is unblocked.\nUse 'task list' to find task IDs.",
	Example: `  task block '{"id":"1","reason":"Waiting for API documentation"}'
  echo '{"id":"abc123","reason":"Dependency not available"}' | task block -
  task block -`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		database := openDB()
		defer database.Close()

		doc, err := jsonDoc(blockJSON, args, "")
		if err != nil {
			fatal(err)
		}

		id, _ := service.GetStringFieldTrim(doc, "id")
		reason, _ := service.GetStringFieldTrim(doc, "reason")

		if err := cliErrors.ValidateID(id); err != nil {
			fatal(err)
		}

		// GetStringFieldTrim already trims and reports an empty reason as
		// absent, so an empty value here means the reason is missing.
		if reason == "" {
			fatal(cliErrors.MissingArgumentError("reason", "reason is required in JSON document"))
		}

		result, err := service.BlockTask(database, service.BlockTaskInput{
			ID:     id,
			Reason: reason,
		})
		if err != nil {
			fatal(err)
		}

		printStatusResult("blocked", result)
	},
}

func init() {
	rootCmd.AddCommand(blockCmd)

	blockCmd.Flags().StringVarP(&blockJSON, "json", "j", "", "JSON document (use '-' for stdin)")
}
