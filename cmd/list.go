package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/user/project/internal/db"
	"github.com/user/project/internal/service"
	cliErrors "github.com/user/project/pkg/errors"
	"github.com/user/project/pkg/output"
)

var listJSON string

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all tasks",
	Long:    "List all tasks with optional filters.\n\nYou can filter by milestone, status, or actor to find specific tasks. Use --all to include completed tasks. Use --format to choose output format (table, markdown, or xml).",
	Example: `  task list '{}'
  task list '{"milestone":"sprint-1"}'
  task list '{"status":"todo","actor":"john"}'
  task list '{"format":"markdown"}'
  task list '{"sort_by":"status"}'
  task list '{"limit":10,"offset":0}'
  task list '{"all":true}'
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
			cliErrors.HandleError(err)
			return
		}

		listFilterMilestone, _ := service.GetStringFieldTrim(doc, "milestone")
		listFilterStatus, _ := service.GetStringFieldTrim(doc, "status")
		listFilterActor, _ := service.GetStringFieldTrim(doc, "actor")
		listFilterID, _ := service.GetStringFieldTrim(doc, "id")
		listSortBy, _ := service.GetStringFieldTrim(doc, "sort_by")
		listFormat, _ := service.GetStringFieldTrim(doc, "format")

		listStatusFilter := listFilterStatus
		if strings.ToLower(listFilterStatus) == "all" {
			listStatusFilter = ""
		}

		if listFilterStatus != "" && strings.ToLower(listFilterStatus) != "all" {
			if err := cliErrors.ValidateStatus(listFilterStatus); err != nil {
				cliErrors.HandleError(err)
				return
			}
		}
		if err := cliErrors.ValidateMilestone(listFilterMilestone); err != nil {
			cliErrors.HandleError(err)
			return
		}
		if err := cliErrors.ValidateActor(listFilterActor); err != nil {
			cliErrors.HandleError(err)
			return
		}

		if listSortBy != "" {
			validSortBy := map[string]bool{
				"status":      true,
				"priority":    true,
				"milestone":   true,
				"created":     true,
				"updated":     true,
				"id":          true,
				"title":       true,
				"description": true,
				"actor":       true,
			}
			if !validSortBy[listSortBy] {
				cliErrors.HandleError(cliErrors.ValidationError("sort-by",
					fmt.Sprintf("'%s' is not valid", listSortBy),
					fmt.Sprintf("Valid sort values: status, priority, milestone, created, updated, id, title, description, actor")))
				return
			}
		}

		if listFilterID != "" {
			if err := cliErrors.ValidateID(listFilterID); err != nil {
				cliErrors.HandleError(err)
				return
			}
		}

		listLimit := 20
		listOffset := 0
		listAll := false

		if v, ok := service.GetNumberField(doc, "limit"); ok {
			listLimit = int(v)
		}
		if v, ok := service.GetNumberField(doc, "offset"); ok {
			listOffset = int(v)
		}
		if v, ok := service.GetBooleanField(doc, "all"); ok {
			listAll = v
		}

		if listLimit < 0 {
			cliErrors.HandleError(fmt.Errorf("limit cannot be negative"))
			return
		}
		if listOffset < 0 {
			cliErrors.HandleError(fmt.Errorf("offset cannot be negative"))
			return
		}

		if listFormat != "" {
			validFormats := map[string]bool{"table": true, "markdown": true, "xml": true}
			if !validFormats[listFormat] {
				cliErrors.HandleError(cliErrors.ValidationError("format",
					fmt.Sprintf("'%s' is not valid", listFormat),
					"Valid formats: table, markdown, xml"))
				return
			}
		}

		database, err := db.NewDB(".taskflow/tasks.db")
		if err != nil {
			cliErrors.HandleError(err)
			return
		}
		defer database.Close()

		listService := service.NewListService(database)

		if listFilterID != "" {
			task, err := listService.GetTask(listFilterID)
			if err != nil {
				cliErrors.HandleError(err)
				return
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
			ShowAll:   listAll,
		}

		result, err := listService.ListTasks(filter)
		if err != nil {
			cliErrors.HandleError(err)
			return
		}

		renderer := output.NewTaskTableRenderer()
		renderer.Render(result)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVarP(&listJSON, "json", "j", "", "JSON document (use '-' for stdin)")
}

func ParseLimit(s string) (int, error) {
	if s == "" {
		return 20, nil
	}
	return strconv.Atoi(s)
}

func ParseOffset(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}