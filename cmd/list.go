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

var (
	listFilterMilestone string
	listFilterSprint    string
	listFilterStatus    string
	listFilterActor     string
	listFilterID        string
	listSortBy          string
	listFormat          string
	listLimit           int
	listOffset          int
	listAll             bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Long: `List all tasks with optional filters.

You can filter by milestone, status, or actor to find specific tasks.
Use the --all flag to include completed tasks in the listing.
Use --format to choose output format (table, markdown, or xml).
Use --sort-by to sort by status, priority, milestone, created, updated, id, sprint, title, description, or actor.
Use --id to get a specific task by its ID.`,
	Example: `  task list
  task list -m "sprint-1"
  task list -s pending -a john
  task list --format markdown
  task list --limit 10 --offset 0
  task list --all
  task list --sort-by status
  task list --sort-by created
  task list --sort-by title
  task list --sort-by actor
  task list --id task-123`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Handle "all" status filter - clear status filter when "all" is passed
		listStatusFilter := listFilterStatus
		if strings.ToLower(listFilterStatus) == "all" {
			listStatusFilter = ""
		}

		// Validate filter values
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

		// Validate sort by
		if listSortBy != "" {
			validSortBy := map[string]bool{
				"status":      true,
				"priority":    true,
				"milestone":   true,
				"created":     true,
				"updated":     true,
				"id":          true,
				"sprint":      true,
				"title":       true,
				"description": true,
				"actor":       true,
			}
			if !validSortBy[listSortBy] {
				cliErrors.HandleError(cliErrors.ValidationError("sort-by",
					fmt.Sprintf("'%s' is not valid", listSortBy),
					fmt.Sprintf("Valid sort values: status, priority, milestone, created, updated, id, sprint, title, description, actor")))
				return
			}
		}

		// Validate ID
		if listFilterID != "" {
			if err := cliErrors.ValidateID(listFilterID); err != nil {
				cliErrors.HandleError(err)
				return
			}
		}

		// Validate format
		format, err := output.ParseOutputFormat(listFormat)
		if err != nil {
			cliErrors.HandleError(err)
			return
		}

		// Validate pagination parameters
		if listLimit < 0 {
			cliErrors.HandleError(fmt.Errorf("limit cannot be negative"))
			return
		}
		if listOffset < 0 {
			cliErrors.HandleError(fmt.Errorf("offset cannot be negative"))
			return
		}

		// Initialize database
		database, err := db.NewDB(".taskflow/tasks.db")
		if err != nil {
			cliErrors.HandleError(err)
			return
		}
		defer database.Close()

		// Create list service
		listService := service.NewListService(database)

		// If ID is specified, get a single task
		if listFilterID != "" {
			task, err := listService.GetTask(listFilterID)
			if err != nil {
				cliErrors.HandleError(err)
				return
			}
			// Render single task
			renderer := output.NewTaskTableRenderer(format)
			renderer.Render(&service.ListTaskResult{
				Tasks:   []service.TaskItem{*task},
				Total:   1,
				Limit:   1,
				Offset:  0,
				HasMore: false,
			})
			return
		}

		// Build filter
		filter := &service.ListTaskFilter{
			Milestone: listFilterMilestone,
			Sprint:    listFilterSprint,
			Status:    listStatusFilter,
			Actor:     listFilterActor,
			ID:        listFilterID,
			SortBy:    listSortBy,
			Limit:     listLimit,
			Offset:    listOffset,
			ShowAll:   listAll,
		}

		// Execute list operation
		result, err := listService.ListTasks(filter)
		if err != nil {
			cliErrors.HandleError(err)
			return
		}

		// Render output
		renderer := output.NewTaskTableRenderer(format)
		renderer.Render(result)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVarP(&listAll, "all", "a", false, "Show all tasks including completed")
	listCmd.Flags().StringVarP(&listFilterMilestone, "milestone", "m", "", "Filter by milestone")
	listCmd.Flags().StringVarP(&listFilterSprint, "sprint", "r", "", "Filter by sprint")
	listCmd.Flags().StringVarP(&listFilterStatus, "status", "s", "", "Filter by status")
	listCmd.Flags().StringVarP(&listFilterActor, "actor", "", "", "Filter by actor")
	listCmd.Flags().StringVarP(&listFilterID, "id", "", "", "Get task by ID")
	listCmd.Flags().StringVarP(&listSortBy, "sort-by", "", "", "Sort by field (status, priority, milestone, created, updated)")
	listCmd.Flags().StringVarP(&listFormat, "format", "f", "table", "Output format (table|markdown|xml)")
	listCmd.Flags().IntVarP(&listLimit, "limit", "l", 20, "Maximum number of tasks to display")
	listCmd.Flags().IntVarP(&listOffset, "offset", "o", 0, "Number of tasks to skip")

	// Register custom completers for sort-by
	listCmd.RegisterFlagCompletionFunc("sort-by", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"status", "priority", "milestone", "created", "updated"}, cobra.ShellCompDirectiveNoFileComp
	})

	// Register custom completers for status
	listCmd.RegisterFlagCompletionFunc("status", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"todo", "in_progress", "done", "blocked"}, cobra.ShellCompDirectiveNoFileComp
	})

	// Register custom completers for format
	listCmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return output.GetValidFormats(), cobra.ShellCompDirectiveNoFileComp
	})
}

// GetListFilterMilestone returns the milestone filter value
func GetListFilterMilestone() string {
	return listFilterMilestone
}

// GetListFilterSprint returns the sprint filter value
func GetListFilterSprint() string {
	return listFilterSprint
}

// GetListFilterStatus returns the status filter value
func GetListFilterStatus() string {
	return listFilterStatus
}

// GetListFilterActor returns the actor filter value
func GetListFilterActor() string {
	return listFilterActor
}

// GetListFormat returns the format value
func GetListFormat() string {
	return listFormat
}

// GetListLimit returns the limit value
func GetListLimit() int {
	return listLimit
}

// GetListOffset returns the offset value
func GetListOffset() int {
	return listOffset
}

// GetListAll returns the all flag value
func GetListAll() bool {
	return listAll
}

// GetListFilterID returns the ID filter value
func GetListFilterID() string {
	return listFilterID
}

// GetListSortBy returns the sort by value
func GetListSortBy() string {
	return listSortBy
}

// SetListFilterMilestone sets the milestone filter value (for testing)
func SetListFilterMilestone(val string) {
	listFilterMilestone = val
}

// SetListFilterSprint sets the sprint filter value (for testing)
func SetListFilterSprint(val string) {
	listFilterSprint = val
}

// SetListFilterStatus sets the status filter value (for testing)
func SetListFilterStatus(val string) {
	listFilterStatus = val
}

// SetListFilterActor sets the actor filter value (for testing)
func SetListFilterActor(val string) {
	listFilterActor = val
}

// SetListFormat sets the format value (for testing)
func SetListFormat(val string) {
	listFormat = val
}

// SetListLimit sets the limit value (for testing)
func SetListLimit(val int) {
	listLimit = val
}

// SetListOffset sets the offset value (for testing)
func SetListOffset(val int) {
	listOffset = val
}

// SetListAll sets the all flag value (for testing)
func SetListAll(val bool) {
	listAll = val
}

// SetListFilterID sets the ID filter value (for testing)
func SetListFilterID(val string) {
	listFilterID = val
}

// SetListSortBy sets the sort by value (for testing)
func SetListSortBy(val string) {
	listSortBy = val
}

// ResetListFlags resets all list flags to default values (for testing)
func ResetListFlags() {
	listFilterMilestone = ""
	listFilterStatus = ""
	listFilterActor = ""
	listFilterID = ""
	listSortBy = ""
	listFormat = "table"
	listLimit = 20
	listOffset = 0
	listAll = false
}

// ParseLimit parses limit from string (for testing)
func ParseLimit(s string) (int, error) {
	if s == "" {
		return 20, nil
	}
	return strconv.Atoi(s)
}

// ParseOffset parses offset from string (for testing)
func ParseOffset(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}
