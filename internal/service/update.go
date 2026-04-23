package service

import (
	"time"

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

// UpdateTaskResult contains the result of updating a task
type UpdateTaskResult struct {
	ID          string
	Milestone   string
	Title       string
	Description string
	Actor       string
	Status      string
	LastUpdated time.Time
}

// UpdateTask updates an existing task with partial update support
func UpdateTask(database *db.DB, input *UpdateTaskInput) (*UpdateTaskResult, error) {
	if database == nil {
		return nil, ErrNilDatabase
	}

	if input == nil {
		return nil, ErrNilInput
	}

	// Validate ID is provided
	if input.ID == "" {
		return nil, ErrInvalidID
	}

	// Check if task exists
	existingTask, err := database.ReadTask(input.ID)
	if err != nil {
		return nil, err
	}

	// Build updated task with partial update support
	// Only update fields that are provided (non-empty)
	updatedTask := &db.Task{
		ID:          existingTask.ID,
		Milestone:   existingTask.Milestone,
		Title:       existingTask.Title,
		Description: existingTask.Description,
		Status:      existingTask.Status,
		Actor:       existingTask.Actor,
		LastUpdated: time.Now().UTC(),
	}

	// Apply partial updates only for provided fields
	if input.Title != "" {
		updatedTask.Title = input.Title
	}

	if input.Description != "" {
		updatedTask.Description = input.Description
	}

	if input.Milestone != "" {
		updatedTask.Milestone = input.Milestone
	}

	if input.Status != "" {
		updatedTask.Status = input.Status
	}

	if input.Actor != "" {
		updatedTask.Actor = input.Actor
	}

	// Update in database
	if err := database.UpdateTask(updatedTask); err != nil {
		return nil, err
	}

	// Return the result
	return &UpdateTaskResult{
		ID:          updatedTask.ID,
		Milestone:   updatedTask.Milestone,
		Title:       updatedTask.Title,
		Description: updatedTask.Description,
		Actor:       updatedTask.Actor,
		Status:      updatedTask.Status,
		LastUpdated: updatedTask.LastUpdated,
	}, nil
}
