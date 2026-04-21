package service

import (
	"strings"

	"github.com/user/project/internal/db"
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

	// Calculate pagination info
	hasMore := false
	total := len(items)

	if filter.Limit > 0 && len(items) == filter.Limit {
		hasMore = true
	}

	// If limit is 0, show all, hasMore is false
	if filter.Limit == 0 {
		hasMore = false
	}

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

	// Use a COUNT query to get the total matching tasks
	dbFilter := db.TaskFilter{
		Milestone: filter.Milestone,
		Status:    filter.Status,
		Actor:     filter.Actor,
		Limit:     0, // No limit for count
		Offset:    0, // No offset for count
	}

	tasks, err := s.database.ListTasks(dbFilter)
	if err != nil {
		return 0, err
	}

	return len(tasks), nil
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
