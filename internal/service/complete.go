package service

import (
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

// CompleteTask marks an existing task as completed. The status defaults to
// "done" when no override is given; other provided (non-empty) fields are
// applied as a partial update. The returned task reflects the true stored
// state, including the refreshed LastUpdated stamped by the database.
func CompleteTask(database *db.DB, input *CompleteTaskInput) (*db.Task, error) {
	if input.ID == "" {
		return nil, ErrInvalidID
	}

	existingTask, err := database.ReadTask(input.ID)
	if err != nil {
		return nil, err
	}

	// Apply partial updates only for provided fields
	if input.Title != "" {
		existingTask.Title = input.Title
	}

	if input.Description != "" {
		existingTask.Description = input.Description
	}

	if input.Milestone != "" {
		existingTask.Milestone = input.Milestone
	}

	if input.Actor != "" {
		existingTask.Actor = input.Actor
	}

	// Determine status: use provided status or default to "done"
	existingTask.Status = input.Status
	if existingTask.Status == "" {
		existingTask.Status = "done"
	}

	if err := database.UpdateTask(existingTask); err != nil {
		return nil, err
	}

	return existingTask, nil
}
