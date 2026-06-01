package service

import (
	"fmt"
	"log"
	"time"

	"github.com/rwbaskette/taskflow/internal/db"
	cliErrors "github.com/rwbaskette/taskflow/pkg/errors"
)

// UnblockTaskInput contains the input parameters for unblocking a task
type UnblockTaskInput struct {
	ID          string
	Description string
}

// UnblockTaskResult contains the result of unblocking a task
type UnblockTaskResult struct {
	ID          string
	Milestone   string
	Title       string
	Description string
	Actor       string
	Status      string
	BlockedBy   []string
	LastUpdated time.Time
}

// UnblockTask unblocks a previously blocked task, transitioning it from
// 'blocked' back to 'todo' status. It clears the blocked_by field and
// optionally overwrites the description. The database-level validation
// includes a WHERE status = 'blocked' guard to prevent unauthorized
// status transitions. All changes are performed within a single database
// transaction to ensure consistency.
func UnblockTask(database *db.DB, input UnblockTaskInput) (*UnblockTaskResult, error) {
	if database == nil {
		return nil, ErrNilDatabase
	}

	// Validate ID is provided
	if input.ID == "" {
		return nil, ErrInvalidID
	}

	// Read the existing task to get its current state for the result and
	// to provide a descriptive error message if the task is not blocked.
	existingTask, err := database.ReadTask(input.ID)
	if err != nil {
		if db.IsTaskNotFound(err) {
			return nil, cliErrors.ResourceNotFoundError("task", input.ID)
		}
		return nil, err
	}

	// Verify the task is in blocked status. This application-level check
	// provides a clear error message with the current status. The database-
	// level guard (WHERE status = 'blocked') in the UPDATE statement serves
	// as a defense-in-depth safety net against race conditions.
	if existingTask.Status != "blocked" {
		log.Printf("[WARN] task_unblock rejected: task %s is in '%s' status, not 'blocked'. Unblock operation aborted.",
			input.ID, existingTask.Status)
		return nil, cliErrors.InvalidStatusTransitionError(input.ID, existingTask.Status)
	}

	// Determine whether to overwrite the description.
	var newDescription *string
	if input.Description != "" {
		newDescription = &input.Description
	}

	// Begin a transaction to ensure the status change is performed atomically.
	tx, err := database.BeginTx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Use the database-level unblock method which includes the WHERE
	// status = 'blocked' guard. This is the authoritative check that
	// prevents unauthorized status transitions at the database level.
	now := time.Now().UTC()
	if err := database.UnblockTaskTx(tx, input.ID, newDescription, now); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Build the result. Use the existing task's description if it was
	// preserved (no new description provided), otherwise use the new one.
	resultDescription := existingTask.Description
	if newDescription != nil {
		resultDescription = *newDescription
	}

	// Return the result
	return &UnblockTaskResult{
		ID:          existingTask.ID,
		Milestone:   existingTask.Milestone,
		Title:       existingTask.Title,
		Description: resultDescription,
		Actor:       existingTask.Actor,
		Status:      "todo",
		BlockedBy:   nil,
		LastUpdated: now,
	}, nil
}
