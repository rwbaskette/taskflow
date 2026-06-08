package service

import (
	"time"

	"github.com/rwbaskette/taskflow/internal/db"
	"github.com/rwbaskette/taskflow/internal/timeutil"
)

// AddTaskInput contains the input parameters for adding a task
type AddTaskInput struct {
	ID          string
	Milestone   string
	Title       string
	Description string
	Actor       string
	BlockedBy   []string
}

// AddTaskResult contains the result of adding a task
type AddTaskResult struct {
	ID          string
	Milestone   string
	Title       string
	Description string
	Actor       string
	BlockedBy   []string
	Status      string
	LastUpdated time.Time
}

// AddTask creates a new task in the database
func AddTask(database *db.DB, input *AddTaskInput) (*AddTaskResult, error) {
	if database == nil {
		return nil, ErrNilDatabase
	}

	if input == nil {
		return nil, ErrNilInput
	}

	// Validate blocked_by dependencies before creating the task
	if err := ValidateBlockedBy(database, input.BlockedBy, input.ID); err != nil {
		return nil, err
	}

	// Create the task
	task := &db.Task{
		ID:          input.ID,
		Milestone:   input.Milestone,
		Title:       input.Title,
		Description: input.Description,
		Status:      "todo",
		Actor:       input.Actor,
		BlockedBy:   input.BlockedBy,
		LastUpdated: timeutil.Now(),
	}

	// Insert into database
	if err := database.CreateTask(task); err != nil {
		return nil, err
	}

	// Return the result
	return &AddTaskResult{
		ID:          task.ID,
		Milestone:   task.Milestone,
		Title:       task.Title,
		Description: task.Description,
		Actor:       task.Actor,
		BlockedBy:   task.BlockedBy,
		Status:      task.Status,
		LastUpdated: task.LastUpdated,
	}, nil
}
