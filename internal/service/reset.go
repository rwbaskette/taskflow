package service

import (
	"time"

	"github.com/rwbaskette/taskflow/internal/db"
)

// ResetTimedOutInput contains the input parameters for resetting timed out tasks
type ResetTimedOutInput struct {
	TimeoutMinutes int
}

// ResetTimedOutResult contains the result of resetting timed out tasks
type ResetTimedOutResult struct {
	ResetTasks []ResetTaskResult
}

// ResetTaskResult contains the result of resetting a single task
type ResetTaskResult struct {
	ID          string
	Milestone   string
	Title       string
	Description string
	Actor       string
	Status      string
	LastUpdated time.Time
}

// ResetTimedOut finds in-progress tasks exceeding timeout and resets them to todo
func ResetTimedOut(database *db.DB, input ResetTimedOutInput) (*ResetTimedOutResult, error) {
	if database == nil {
		return nil, ErrNilDatabase
	}

	// Validate timeout minutes
	if input.TimeoutMinutes <= 0 {
		return nil, ErrInvalidTimeout
	}

	// Find all in-progress tasks
	filter := db.TaskFilter{
		Status: "in_progress",
	}

	inProgressTasks, err := database.ListTasks(filter)
	if err != nil {
		return nil, err
	}

	// Find tasks that have exceeded the timeout
	timedOutTasks := GetTimedOutTasks(inProgressTasks, input.TimeoutMinutes)

	// Reset each timed out task to todo status
	var resetTasks []ResetTaskResult

	for _, task := range timedOutTasks {
		updatedTask := &db.Task{
			ID:          task.ID,
			Milestone:   task.Milestone,
			Title:       task.Title,
			Description: task.Description,
			Actor:       task.Actor,
			Status:      "todo",
			LastUpdated: time.Now().UTC(),
		}

		if err := database.UpdateTask(updatedTask); err != nil {
			return nil, err
		}

		resetTasks = append(resetTasks, ResetTaskResult{
			ID:          updatedTask.ID,
			Milestone:   updatedTask.Milestone,
			Title:       updatedTask.Title,
			Description: updatedTask.Description,
			Actor:       updatedTask.Actor,
			Status:      updatedTask.Status,
			LastUpdated: updatedTask.LastUpdated,
		})
	}

	return &ResetTimedOutResult{
		ResetTasks: resetTasks,
	}, nil
}
