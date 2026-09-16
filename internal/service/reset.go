package service

import (
	"github.com/rwbaskette/taskflow/internal/db"
)

// ResetTimedOut finds in-progress tasks exceeding the timeout and resets
// them to todo status. It returns the reset tasks with their true stored
// state, including the refreshed LastUpdated stamped by the database.
func ResetTimedOut(database *db.DB, timeoutMinutes int) ([]db.Task, error) {
	if timeoutMinutes <= 0 {
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
	timedOutTasks := GetTimedOutTasks(inProgressTasks, timeoutMinutes)

	// Reset each timed out task to todo status
	resetTasks := make([]db.Task, 0, len(timedOutTasks))

	for _, task := range timedOutTasks {
		task.Status = "todo"

		if err := database.UpdateTask(&task); err != nil {
			return nil, err
		}

		resetTasks = append(resetTasks, task)
	}

	return resetTasks, nil
}
