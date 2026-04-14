package service

import (
	"time"

	"github.com/user/project/internal/db"
)

// CompleteTaskInput contains the input parameters for completing a task
type CompleteTaskInput struct {
	ID string
}

// CompleteTaskResult contains the result of completing a task
type CompleteTaskResult struct {
	ID          string
	Milestone   string
	Title       string
	Description string
	Actor       string
	Status      string
	LastUpdated time.Time
}

// CompleteTask marks an existing task as completed
func CompleteTask(database *db.DB, taskID string) (*CompleteTaskResult, error) {
	if database == nil {
		return nil, ErrNilDatabase
	}

	// Validate ID is provided
	if taskID == "" {
		return nil, ErrInvalidID
	}

	// Validate task exists
	existingTask, err := database.ReadTask(taskID)
	if err != nil {
		return nil, err
	}

	// Update task status to completed and set timestamp
	updatedTask := &db.Task{
		ID:          existingTask.ID,
		Milestone:   existingTask.Milestone,
		Title:       existingTask.Title,
		Description: existingTask.Description,
		Actor:       existingTask.Actor,
		Status:      "done",
		LastUpdated: time.Now().UTC(),
	}

	// Update in database
	if err := database.UpdateTask(updatedTask); err != nil {
		return nil, err
	}

	// Return the result
	return &CompleteTaskResult{
		ID:          updatedTask.ID,
		Milestone:   updatedTask.Milestone,
		Title:       updatedTask.Title,
		Description: updatedTask.Description,
		Actor:       updatedTask.Actor,
		Status:      updatedTask.Status,
		LastUpdated: updatedTask.LastUpdated,
	}, nil
}
