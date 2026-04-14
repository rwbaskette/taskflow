package db

import (
	"errors"
	"fmt"
)

// Custom error types for database operations
var (
	// ErrTaskNotFound is returned when a task is not found in the database
	ErrTaskNotFound = errors.New("task not found")

	// ErrTaskAlreadyExists is returned when attempting to create a task with an existing ID
	ErrTaskAlreadyExists = errors.New("task already exists")

	// ErrInvalidTask is returned when task data is invalid
	ErrInvalidTask = errors.New("invalid task data")

	// ErrInvalidID is returned when task ID is empty or invalid
	ErrInvalidID = errors.New("invalid task ID")

	// ErrTransactionFailed is returned when a transaction fails
	ErrTransactionFailed = errors.New("transaction failed")

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

// NewTaskNotFoundError creates a new TaskNotFoundError
func NewTaskNotFoundError(id string) *TaskNotFoundError {
	return &TaskNotFoundError{ID: id}
}

// TaskAlreadyExistsError wraps task ID for detailed error messaging
type TaskAlreadyExistsError struct {
	ID string
}

func (e *TaskAlreadyExistsError) Error() string {
	return fmt.Sprintf("task with ID %q already exists", e.ID)
}

// NewTaskAlreadyExistsError creates a new TaskAlreadyExistsError
func NewTaskAlreadyExistsError(id string) *TaskAlreadyExistsError {
	return &TaskAlreadyExistsError{ID: id}
}

// InvalidTaskError wraps validation errors
type InvalidTaskError struct {
	Field   string
	Message string
}

func (e *InvalidTaskError) Error() string {
	return fmt.Sprintf("invalid task field %q: %s", e.Field, e.Message)
}

// NewInvalidTaskError creates a new InvalidTaskError
func NewInvalidTaskError(field, message string) *InvalidTaskError {
	return &InvalidTaskError{Field: field, Message: message}
}

// IsTaskNotFound checks if an error is a TaskNotFoundError
func IsTaskNotFound(err error) bool {
	if err == nil {
		return false
	}
	var taskErr *TaskNotFoundError
	return errors.As(err, &taskErr)
}

// IsTaskAlreadyExists checks if an error is a TaskAlreadyExistsError
func IsTaskAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	var taskErr *TaskAlreadyExistsError
	return errors.As(err, &taskErr)
}

// IsInvalidTask checks if an error is an InvalidTaskError
func IsInvalidTask(err error) bool {
	if err == nil {
		return false
	}
	var taskErr *InvalidTaskError
	return errors.As(err, &taskErr)
}
