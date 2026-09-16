package service

import (
	"github.com/rwbaskette/taskflow/internal/db"
)

// AddTaskInput contains the input parameters for adding a task
type AddTaskInput struct {
	ID          string
	Milestone   string
	Title       string
	Description string
	Actor       string
}

// AddTask creates a new task in the database and returns the stored task.
// The database stamps Created and LastUpdated when they are zero.
func AddTask(database *db.DB, input *AddTaskInput) (*db.Task, error) {
	task := &db.Task{
		ID:          input.ID,
		Milestone:   input.Milestone,
		Title:       input.Title,
		Description: input.Description,
		Status:      "todo",
		Actor:       input.Actor,
	}

	if err := database.CreateTask(task); err != nil {
		return nil, err
	}

	return task, nil
}
