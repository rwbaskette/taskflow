package service

import (
	"fmt"
	"strings"

	"github.com/rwbaskette/taskflow/internal/db"
)

// ListTaskFilter contains filters for listing tasks
type ListTaskFilter struct {
	Milestone string
	Sprint    string
	Status    string
	Actor     string
	ID        string
	SortBy    string
	Limit     int
	Offset    int
	ShowAll   bool
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

// ListService handles the list operation with filtering
type ListService struct {
	database *db.DB
}

// NewListService creates a new ListService
func NewListService(database *db.DB) *ListService {
	return &ListService{
		database: database,
	}
}

// ListTasks retrieves tasks based on filters with pagination
func (s *ListService) ListTasks(filter *ListTaskFilter) (*ListTaskResult, error) {
	if s.database == nil {
		return nil, ErrNilDatabase
	}

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
	tasks, err := s.database.ListTasks(dbFilter)
	if err != nil {
		return nil, err
	}

	// Convert to service items
	items := make([]TaskItem, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, TaskItem{
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
		})
	}

	// Get the true total count of all matching records (ignoring limit/offset)
	// so that pagination metadata is accurate.
	total, err := s.GetFilteredCount(filter)
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

// GetFilteredCount returns the count of tasks matching the filter (excluding limit/offset)
func (s *ListService) GetFilteredCount(filter *ListTaskFilter) (int, error) {
	if s.database == nil {
		return 0, ErrNilDatabase
	}

	if filter == nil {
		filter = &ListTaskFilter{}
	}

	// Use a COUNT query to get the total matching tasks without fetching all rows
	dbFilter := db.TaskFilter{
		Milestone: filter.Milestone,
		Status:    filter.Status,
		Actor:     filter.Actor,
		// Limit and Offset are intentionally omitted — CountTasks ignores them,
		// but we set them to zero for clarity.
	}

	return s.database.CountTasks(dbFilter)
}

// GetTask retrieves a task by ID
func (s *ListService) GetTask(id string) (*TaskItem, error) {
	if s.database == nil {
		return nil, ErrNilDatabase
	}

	if strings.TrimSpace(id) == "" {
		return nil, ErrInvalidID
	}

	task, err := s.database.GetTaskByID(id)
	if err != nil {
		return nil, err
	}

	return &TaskItem{
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
	}, nil
}
