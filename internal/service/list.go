package service

import (
	"fmt"
	"strings"

	"github.com/rwbaskette/taskflow/internal/db"
)

// ListTaskFilter contains filters for listing tasks
type ListTaskFilter struct {
	Milestone string
	Status    string
	Actor     string
	ID        string
	SortBy    string
	Limit     int
	Offset    int
}

// ListTaskResult contains the result of listing tasks
type ListTaskResult struct {
	Tasks   []TaskItem
	Total   int
	Limit   int
	Offset  int
	HasMore bool
}

// TaskItem represents a task in list output
type TaskItem struct {
	ID          string
	Milestone   string
	Sprint      string
	Title       string
	Description string
	Status      string
	Actor       string
	BlockedBy   []string
	Created     string
	LastUpdated string
}

// taskToItem maps a database task to a list output item.
func taskToItem(task db.Task) TaskItem {
	return TaskItem{
		ID:          task.ID,
		Milestone:   task.Milestone,
		Sprint:      task.Sprint,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Actor:       task.Actor,
		BlockedBy:   task.BlockedBy,
		Created:     task.Created.Format("2006-01-02 15:04:05"),
		LastUpdated: task.LastUpdated.Format("2006-01-02 15:04:05"),
	}
}

// ListTasks retrieves tasks based on filters with pagination
func ListTasks(database *db.DB, filter *ListTaskFilter) (*ListTaskResult, error) {
	if filter == nil {
		filter = &ListTaskFilter{}
	}

	// Build database filter
	dbFilter := db.TaskFilter{
		Milestone: filter.Milestone,
		Status:    filter.Status,
		Actor:     filter.Actor,
		ID:        filter.ID,
		SortBy:    db.SortBy(filter.SortBy),
		Limit:     filter.Limit,
		Offset:    filter.Offset,
	}

	// Fetch tasks from database
	tasks, err := database.ListTasks(dbFilter)
	if err != nil {
		return nil, err
	}

	// Convert to service items
	items := make([]TaskItem, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, taskToItem(task))
	}

	// Get the true total count of all matching records (ignoring limit/offset)
	// so that pagination metadata is accurate. Limit and Offset are
	// intentionally omitted — CountTasks ignores them, but we leave them zero
	// for clarity.
	total, err := database.CountTasks(db.TaskFilter{
		Milestone: filter.Milestone,
		Status:    filter.Status,
		Actor:     filter.Actor,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Determine whether there are more results beyond the current page.
	// There are more if the current page doesn't reach the end of the full set.
	hasMore := filter.Offset+len(items) < total

	return &ListTaskResult{
		Tasks:   items,
		Total:   total,
		Limit:   filter.Limit,
		Offset:  filter.Offset,
		HasMore: hasMore,
	}, nil
}

// GetTask retrieves a task by ID
func GetTask(database *db.DB, id string) (*TaskItem, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrInvalidID
	}

	task, err := database.ReadTask(id)
	if err != nil {
		return nil, err
	}

	item := taskToItem(*task)
	return &item, nil
}
