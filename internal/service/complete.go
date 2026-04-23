package service

import (
	"time"

	"github.com/rwbaskette/taskflow/internal/db"
)

// CompleteTaskInput contains the input parameters for completing a task
type CompleteTaskInput struct {
	ID          string
	Title       string
	Description string
	Milestone   string
	Status      string
	Actor       string
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
func CompleteTask(database *db.DB, input *CompleteTaskInput) (*CompleteTaskResult, error) {
	if database == nil {
		return nil, ErrNilDatabase
	}

	// Validate ID is provided
	if input.ID == "" {
		return nil, ErrInvalidID
	}

	// Validate task exists
	existingTask, err := database.ReadTask(input.ID)
	if err != nil {
		return nil, err
	}

	// Determine status: use provided status or default to "done"
	status := input.Status
	if status == "" {
		status = "done"
	}

	// Update task with provided fields or existing values
	updatedTask := &db.Task{
		ID:          existingTask.ID,
		Milestone:   input.Milestone,
		Title:       input.Title,
		Description: input.Description,
		Actor:       input.Actor,
		Status:      status,
		LastUpdated: time.Now().UTC(),
	}

	// Use existing values for empty fields
	if updatedTask.Title == "" {
		updatedTask.Title = existingTask.Title
	}
	if updatedTask.Description == "" {
		updatedTask.Description = existingTask.Description
	}
	if updatedTask.Milestone == "" {
		updatedTask.Milestone = existingTask.Milestone
	}
	if updatedTask.Actor == "" {
		updatedTask.Actor = existingTask.Actor
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
