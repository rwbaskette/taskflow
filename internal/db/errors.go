package db

import (
	"errors"
	"fmt"
)

// Sentinels for database operations
var (
	// ErrInvalidID is returned when task ID is empty or invalid
	ErrInvalidID = errors.New("invalid task ID")

	// ErrNilTask is returned when a nil task is passed to a function
	ErrNilTask = errors.New("nil task provided")

	// ErrNilDB is returned when a nil database connection is passed
	ErrNilDB = errors.New("nil database connection")
)

// TaskNotFoundError wraps task ID for detailed error messaging
type TaskNotFoundError struct {
	ID string
}

func (e *TaskNotFoundError) Error() string {
	return fmt.Sprintf("task with ID %q not found", e.ID)
}

// TaskAlreadyExistsError wraps task ID for detailed error messaging
type TaskAlreadyExistsError struct {
	ID string
}

func (e *TaskAlreadyExistsError) Error() string {
	return fmt.Sprintf("task with ID %q already exists", e.ID)
}

// InvalidTaskError wraps validation errors
type InvalidTaskError struct {
	Field   string
	Message string
}

func (e *InvalidTaskError) Error() string {
	return fmt.Sprintf("invalid task field %q: %s", e.Field, e.Message)
}

// TaskNotBlockedError is returned when an operation requires a task in
// 'blocked' status but the task exists in a different status. The Status
// field carries the task's actual current status so callers can produce a
// precise message.
type TaskNotBlockedError struct {
	ID     string
	Status string
}

func (e *TaskNotBlockedError) Error() string {
	return fmt.Sprintf("task %q is in %q status, not blocked", e.ID, e.Status)
}
