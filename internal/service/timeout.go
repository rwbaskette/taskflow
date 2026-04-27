package service

import (
	"time"

	"github.com/rwbaskette/taskflow/internal/db"
)

// IsTimedOut checks if a task has exceeded the specified timeout duration
func IsTimedOut(task *db.Task, timeoutMinutes int) bool {
	if task == nil {
		return false
	}

	if task.Status != "in_progress" {
		return false
	}

	timeoutDuration := time.Duration(timeoutMinutes) * time.Minute
	elapsed := time.Since(task.LastUpdated)

	return elapsed > timeoutDuration
}

// GetTimedOutTasks filters tasks to find those that have exceeded the timeout
func GetTimedOutTasks(tasks []db.Task, timeoutMinutes int) []db.Task {
	var timedOutTasks []db.Task

	for _, task := range tasks {
		if IsTimedOut(&task, timeoutMinutes) {
			timedOutTasks = append(timedOutTasks, task)
		}
	}

	return timedOutTasks
}
