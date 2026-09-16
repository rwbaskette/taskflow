package service

import (
	"time"

	"github.com/rwbaskette/taskflow/internal/db"
)

// DeleteTaskResult contains the result of deleting a task
type DeleteTaskResult struct {
	ID        string
	Title     string
	DeletedOn time.Time
}

// DeleteTask soft-deletes a task by moving it to the deleted_tasks table
func DeleteTask(database *db.DB, id string) (*DeleteTaskResult, error) {
	if id == "" {
		return nil, ErrInvalidID
	}

	existingTask, err := database.ReadTask(id)
	if err != nil {
		return nil, err
	}

	deletedOn, err := database.SoftDeleteTask(id)
	if err != nil {
		return nil, err
	}

	return &DeleteTaskResult{
		ID:        existingTask.ID,
		Title:     existingTask.Title,
		DeletedOn: deletedOn,
	}, nil
}
