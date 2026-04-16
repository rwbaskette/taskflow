package service

import (
	"time"

	"github.com/user/project/internal/db"
)

// DeleteTaskInput contains the input parameters for deleting a task
type DeleteTaskInput struct {
	ID string
}

// DeleteTaskResult contains the result of deleting a task
type DeleteTaskResult struct {
	ID        string
	Title     string
	DeletedOn time.Time
}

// DeleteTask soft-deletes a task by moving it to the deleted_tasks table
func DeleteTask(database *db.DB, input *DeleteTaskInput) (*DeleteTaskResult, error) {
	if database == nil {
		return nil, ErrNilDatabase
	}

	if input.ID == "" {
		return nil, ErrInvalidID
	}

	existingTask, err := database.ReadTask(input.ID)
	if err != nil {
		return nil, err
	}

	if err := database.SoftDeleteTask(input.ID); err != nil {
		return nil, err
	}

	return &DeleteTaskResult{
		ID:        existingTask.ID,
		Title:     existingTask.Title,
		DeletedOn: time.Now().UTC(),
	}, nil
}
