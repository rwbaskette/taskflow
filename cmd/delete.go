package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	cliErrors "github.com/rwbaskette/taskflow/internal/errors"
	"github.com/rwbaskette/taskflow/internal/service"
)

var deleteJSON string

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Soft delete a task",
	Long:  "Soft delete a task by moving it to the deleted_tasks table.\n\nThe task is moved to a deleted_tasks table with a deleted_on timestamp rather than being permanently removed. Use 'task list' to find task IDs.",
	Example: `  task delete '{"id":"1"}'
  echo '{"id":"abc123"}' | task delete -
  task delete -`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		jsonArg := deleteJSON
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

		id, _ := service.GetStringFieldTrim(doc, "id")

		if err := cliErrors.ValidateID(id); err != nil {
			fatal(err)
		}

		database := openDB()
		defer database.Close()

		result, err := service.DeleteTask(database, id)
		if err != nil {
			fatal(err)
		}

		fmt.Printf("Task deleted successfully:\n")
		fmt.Printf("  ID: %s\n", result.ID)
		fmt.Printf("  Title: %s\n", result.Title)
		fmt.Printf("  Deleted On: %s\n", result.DeletedOn.Format("2006-01-02 15:04:05"))
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringVarP(&deleteJSON, "json", "j", "", "JSON document (use '-' for stdin)")
}
