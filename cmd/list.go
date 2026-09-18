package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rwbaskette/taskflow/internal/db"
	cliErrors "github.com/rwbaskette/taskflow/internal/errors"
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
		doc, err := jsonDoc(listJSON, args, "{}")
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
			if canonical := canonicalStatus(listFilterStatus); canonical != "" {
				listStatusFilter = canonical
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
			fatal(errors.New("limit cannot be negative"))
		}
		if listOffset < 0 {
			fatal(errors.New("offset cannot be negative"))
		}

		database := openDB()
		defer database.Close()

		if listFilterID != "" {
			task, err := database.ReadTask(listFilterID)
			if err != nil {
				fatal(err)
			}
			renderListResult(&service.ListTaskResult{
				Tasks:   []service.TaskItem{service.TaskToItem(*task)},
				Total:   1,
				Limit:   1,
				Offset:  0,
				HasMore: false,
			})
			return
		}

		dbFilter := db.TaskFilter{
			Milestone: listFilterMilestone,
			Status:    listStatusFilter,
			Actor:     listFilterActor,
			SortBy:    db.SortBy(listSortBy),
			Limit:     listLimit,
			Offset:    listOffset,
		}

		tasks, err := database.ListTasks(dbFilter)
		if err != nil {
			fatal(err)
		}

		items := make([]service.TaskItem, 0, len(tasks))
		for _, task := range tasks {
			items = append(items, service.TaskToItem(task))
		}

		// Get the true total count of all matching records (ignoring
		// limit/offset) so that pagination metadata is accurate. Limit and
		// Offset are intentionally omitted — CountTasks ignores them, but we
		// leave them zero for clarity.
		total, err := database.CountTasks(db.TaskFilter{
			Milestone: listFilterMilestone,
			Status:    listStatusFilter,
			Actor:     listFilterActor,
		})
		if err != nil {
			fatal(err)
		}

		// Determine whether there are more results beyond the current page.
		// There are more if the current page doesn't reach the end of the
		// full set.
		hasMore := listOffset+len(items) < total

		renderListResult(&service.ListTaskResult{
			Tasks:   items,
			Total:   total,
			Limit:   listLimit,
			Offset:  listOffset,
			HasMore: hasMore,
		})
	},
}

// renderListResult prints the list result as two-space indented JSON.
func renderListResult(result *service.ListTaskResult) {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Println("[]")
		return
	}

	fmt.Println(string(jsonData))
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVarP(&listJSON, "json", "j", "", "JSON document (use '-' for stdin)")
}
