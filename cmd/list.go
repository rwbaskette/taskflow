package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	cliErrors "github.com/rwbaskette/taskflow/internal/errors"
	"github.com/rwbaskette/taskflow/internal/output"
	"github.com/rwbaskette/taskflow/internal/service"
)

var listJSON string

// validSortBy are the sort keys list accepts. "priority" was removed: the
// db-side priority ordering was deleted, and the key silently fell
// through to default ordering.
var validSortBy = []string{
	"status",
	"milestone",
	"created",
	"updated",
	"id",
	"title",
	"description",
	"actor",
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Long:  "List all tasks with optional filters.\n\nYou can filter by milestone, status, or actor to find specific tasks.",
	Example: `  task list '{}'
  task list '{"milestone":"sprint-1"}'
  task list '{"status":"todo","actor":"john"}'
  task list '{"sort_by":"status"}'
  task list '{"limit":10,"offset":0}'
  task list '{"id":"task-123"}'`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		jsonArg := listJSON
		if jsonArg == "" && len(args) > 0 {
			jsonArg = args[0]
		}
		if jsonArg == "" {
			jsonArg = "{}"
		}

		doc, err := service.ParseJSONFromArg(jsonArg)
		if err != nil {
			fatal(err)
		}

		listFilterMilestone, _ := service.GetStringFieldTrim(doc, "milestone")
		listFilterStatus, _ := service.GetStringFieldTrim(doc, "status")
		listFilterActor, _ := service.GetStringFieldTrim(doc, "actor")
		listFilterID, _ := service.GetStringFieldTrim(doc, "id")
		listSortBy, _ := service.GetStringFieldTrim(doc, "sort_by")

		// listFilterStatus is trimmed by GetStringFieldTrim; compare the
		// lowercased value so any case of "all" means every status.
		lowerStatus := strings.ToLower(listFilterStatus)
		listStatusFilter := listFilterStatus
		if lowerStatus == "all" {
			listStatusFilter = ""
		}

		if listFilterStatus != "" && lowerStatus != "all" {
			if err := cliErrors.ValidateStatus(listFilterStatus); err != nil {
				fatal(err)
			}
		}
		if err := cliErrors.ValidateMilestone(listFilterMilestone); err != nil {
			fatal(err)
		}
		if err := cliErrors.ValidateActor(listFilterActor); err != nil {
			fatal(err)
		}

		if listSortBy != "" {
			if !slices.Contains(validSortBy, listSortBy) {
				fatal(cliErrors.ValidationError("sort-by",
					fmt.Sprintf("'%s' is not valid", listSortBy),
					fmt.Sprintf("Valid sort values: %s", strings.Join(validSortBy, ", "))))
			}
		}

		if listFilterID != "" {
			if err := cliErrors.ValidateID(listFilterID); err != nil {
				fatal(err)
			}
		}

		listLimit := 20
		listOffset := 0

		if v, ok := service.GetNumberField(doc, "limit"); ok {
			listLimit = int(v)
		}
		if v, ok := service.GetNumberField(doc, "offset"); ok {
			listOffset = int(v)
		}

		if listLimit < 0 {
			fatal(fmt.Errorf("limit cannot be negative"))
		}
		if listOffset < 0 {
			fatal(fmt.Errorf("offset cannot be negative"))
		}

		database := openDB()
		defer database.Close()

		if listFilterID != "" {
			task, err := service.GetTask(database, listFilterID)
			if err != nil {
				fatal(err)
			}
			renderer := output.NewTaskTableRenderer()
			renderer.Render(&service.ListTaskResult{
				Tasks:   []service.TaskItem{*task},
				Total:   1,
				Limit:   1,
				Offset:  0,
				HasMore: false,
			})
			return
		}

		filter := &service.ListTaskFilter{
			Milestone: listFilterMilestone,
			Status:    listStatusFilter,
			Actor:     listFilterActor,
			ID:        listFilterID,
			SortBy:    listSortBy,
			Limit:     listLimit,
			Offset:    listOffset,
		}

		result, err := service.ListTasks(database, filter)
		if err != nil {
			fatal(err)
		}

		renderer := output.NewTaskTableRenderer()
		renderer.Render(result)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVarP(&listJSON, "json", "j", "", "JSON document (use '-' for stdin)")
}
