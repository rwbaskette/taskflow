package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/rwbaskette/taskflow/internal/db"
)

// WaitTaskInput contains the input parameters for waiting on tasks
type WaitTaskInput struct {
	TaskIDs []string
	Timeout int // Timeout in milliseconds; 0 means no timeout (wait forever)
}

// WaitTaskResult contains the result of waiting on tasks
type WaitTaskResult struct {
	TaskID string
	Status string
	State  string // "completed" or "timed_out"
	Title  string
}

// ErrTaskNotFound is returned when a task does not exist
var ErrTaskNotFound = fmt.Errorf("task not found")

// ErrInvalidTaskID is returned when a task ID is invalid
var ErrInvalidTaskID = fmt.Errorf("invalid task ID")

// WaitTask waits for one or more tasks to complete.
// It returns when all tasks are done or when the timeout is reached.
// If timeout is 0, it waits indefinitely.
// If a task is already completed, it returns immediately for that task.
func WaitTask(database *db.DB, input *WaitTaskInput) ([]WaitTaskResult, error) {
	if database == nil {
		return nil, ErrNilDatabase
	}

	if input == nil {
		return nil, ErrNilInput
	}

	if len(input.TaskIDs) == 0 {
		return nil, ErrInvalidID
	}

	// Validate all task IDs
	for _, id := range input.TaskIDs {
		if strings.TrimSpace(id) == "" {
			return nil, ErrInvalidID
		}
	}

	results := make([]WaitTaskResult, 0, len(input.TaskIDs))
	pendingIDs := make([]string, 0, len(input.TaskIDs))

	// Initial check: see which tasks are already done
	for _, taskID := range input.TaskIDs {
		task, err := database.ReadTask(taskID)
		if err != nil {
			// Check if it's a "not found" error
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "TaskNotFoundError") {
				// Return error for invalid task ID per spec
				return nil, fmt.Errorf("%w: %s", ErrTaskNotFound, taskID)
			}
			return nil, err
		}

		result := WaitTaskResult{
			TaskID: taskID,
			Status: task.Status,
			Title:  task.Title,
		}

		if task.Status == "done" {
			result.State = "completed"
			results = append(results, result)
		} else {
			result.State = "pending"
			pendingIDs = append(pendingIDs, taskID)
		}
	}

	// If all tasks are already done, return immediately
	if len(pendingIDs) == 0 {
		return results, nil
	}

	// Set up timeout
	var deadline time.Time
	var hasTimeout bool
	if input.Timeout > 0 {
		deadline = time.Now().Add(time.Duration(input.Timeout) * time.Millisecond)
		hasTimeout = true
	}

	// Poll interval
	pollInterval := 500 * time.Millisecond

	// Wait for pending tasks
	for len(pendingIDs) > 0 {
		// Check timeout
		if hasTimeout && time.Now().After(deadline) {
			// Mark remaining tasks as timed out
			for _, taskID := range pendingIDs {
				task, err := database.ReadTask(taskID)
				var title string
				if err == nil {
					title = task.Title
				}
				results = append(results, WaitTaskResult{
					TaskID: taskID,
					Status: "timed_out",
					State:  "timed_out",
					Title:  title,
				})
			}
			return results, nil
		}

		// Wait before polling
		time.Sleep(pollInterval)

		// Check each pending task
		stillPending := make([]string, 0, len(pendingIDs))
		for _, taskID := range pendingIDs {
			task, err := database.ReadTask(taskID)
			if err != nil {
				// If error reading task, check if it's "not found"
				if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "TaskNotFoundError") {
					// Task was deleted or something went wrong
					results = append(results, WaitTaskResult{
						TaskID: taskID,
						Status: "error",
						State:  "error",
						Title:  "",
					})
					continue
				}
				return nil, err
			}

			if task.Status == "done" {
				results = append(results, WaitTaskResult{
					TaskID: taskID,
					Status: task.Status,
					State:  "completed",
					Title:  task.Title,
				})
			} else {
				stillPending = append(stillPending, taskID)
			}
		}
		pendingIDs = stillPending
	}

	return results, nil
}