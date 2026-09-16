package service

import (
	"github.com/rwbaskette/taskflow/internal/db"
)

// UpdateTaskInput contains the input parameters for updating a task
type UpdateTaskInput struct {
	ID          string
	Title       string
	Description string
	Milestone   string
	Status      string
	Actor       string
}

// UpdateTask updates an existing task with partial update support: only
// fields that are provided (non-empty) are applied. The returned task
// reflects the true stored state, including the refreshed LastUpdated
// stamped by the database.
func UpdateTask(database *db.DB, input *UpdateTaskInput) (*db.Task, error) {
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

	if input.Status != "" {
		existingTask.Status = input.Status
	}

	if input.Actor != "" {
		existingTask.Actor = input.Actor
	}

	if err := database.UpdateTask(existingTask); err != nil {
		return nil, err
	}

	return existingTask, nil
}
