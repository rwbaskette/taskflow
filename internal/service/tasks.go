package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rwbaskette/taskflow/internal/db"
	cliErrors "github.com/rwbaskette/taskflow/internal/errors"
)

// Task-operation errors (the remaining service-level sentinels).
var (
	ErrMissingBlockReason = errors.New("reason for blocking is required")
	ErrInvalidTimeout     = errors.New("timeout minutes must be a positive integer")
)

// ListTaskResult contains the result of listing tasks
type ListTaskResult struct {
	Tasks   []TaskItem `json:"tasks"`
	Total   int        `json:"total"`
	Limit   int        `json:"limit"`
	Offset  int        `json:"offset"`
	HasMore bool       `json:"hasMore"`
}

// TaskItem represents a task in list output
type TaskItem struct {
	ID          string
	Milestone   string
	Sprint      string
	Title       string
	Description string
	Status      string
	Actor       string
	BlockedBy   []string
	Created     string
	LastUpdated string
}

// TaskToItem maps a database task to a list output item.
func TaskToItem(task db.Task) TaskItem {
	return TaskItem{
		ID:          task.ID,
		Milestone:   task.Milestone,
		Sprint:      task.Sprint,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Actor:       task.Actor,
		BlockedBy:   task.BlockedBy,
		Created:     task.Created.Format("2006-01-02 15:04:05"),
		LastUpdated: task.LastUpdated.Format("2006-01-02 15:04:05"),
	}
}

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

// BlockTaskInput contains the input parameters for blocking a task
type BlockTaskInput struct {
	ID     string
	Reason string
}

// BlockTask blocks an existing task with a reason. The reason is appended to
// the task description and recorded in the blocked_by list. The returned
// task reflects the true stored state, including the refreshed LastUpdated
// stamped by the database.
func BlockTask(database *db.DB, input BlockTaskInput) (*db.Task, error) {
	if strings.TrimSpace(input.Reason) == "" {
		return nil, ErrMissingBlockReason
	}

	existingTask, err := database.ReadTask(input.ID)
	if err != nil {
		return nil, err
	}

	// Append reason to existing description
	newDescription := existingTask.Description
	if newDescription != "" {
		newDescription += "\n"
	}
	newDescription += fmt.Sprintf("[BLOCKED: %s]", input.Reason)

	existingTask.Status = "blocked"
	existingTask.Description = newDescription
	existingTask.BlockedBy = append(existingTask.BlockedBy, input.Reason)

	if err := database.UpdateTask(existingTask); err != nil {
		return nil, err
	}

	return existingTask, nil
}

// UnblockTask unblocks a previously blocked task, transitioning it from
// 'blocked' back to 'todo' status. It clears the blocked_by field and
// optionally overwrites the description (when a non-empty description is
// given). The database-level UPDATE includes a WHERE status = 'blocked'
// guard as defense-in-depth against race conditions, and the returned task
// reflects the true stored state.
func UnblockTask(database *db.DB, id, description string) (*db.Task, error) {
	var newDescription *string
	if description != "" {
		newDescription = &description
	}

	task, err := database.UnblockTask(id, newDescription)
	if err != nil {
		var notFound *db.TaskNotFoundError
		if errors.As(err, &notFound) {
			return nil, cliErrors.ResourceNotFoundError("task", id)
		}
		var notBlocked *db.TaskNotBlockedError
		if errors.As(err, &notBlocked) {
			return nil, cliErrors.InvalidStatusTransitionError(id, notBlocked.Status)
		}
		return nil, err
	}

	return task, nil
}

// DeleteTaskResult contains the result of deleting a task
type DeleteTaskResult struct {
	ID        string
	Title     string
	DeletedOn time.Time
}

// DeleteTask soft-deletes a task by moving it to the deleted_tasks table
func DeleteTask(database *db.DB, id string) (*DeleteTaskResult, error) {
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

	// Find tasks that have exceeded the timeout: elapsed time since the
	// last update must be greater than the timeout duration.
	timeoutDuration := time.Duration(timeoutMinutes) * time.Minute
	var timedOutTasks []db.Task
	for _, task := range inProgressTasks {
		elapsed := time.Since(task.LastUpdated)
		if elapsed > timeoutDuration {
			timedOutTasks = append(timedOutTasks, task)
		}
	}

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
