package service

import (
	"time"

	"github.com/rwbaskette/taskflow/internal/db"
)

// GetTimedOutTasks filters tasks to those whose LastUpdated is older than
// the timeout duration. Callers must pre-filter by status (ResetTimedOut
// queries only 'in_progress' tasks), so status is intentionally not checked
// here. The input slice is not modified.
func GetTimedOutTasks(tasks []db.Task, timeoutMinutes int) []db.Task {
	var timedOutTasks []db.Task

	timeoutDuration := time.Duration(timeoutMinutes) * time.Minute

	for _, task := range tasks {
		elapsed := time.Since(task.LastUpdated)
		if elapsed > timeoutDuration {
			timedOutTasks = append(timedOutTasks, task)
		}
	}

	return timedOutTasks
}
