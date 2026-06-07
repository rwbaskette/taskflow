package service

import "errors"

// Service-specific errors
var (
	ErrNilDatabase          = errors.New("database connection is nil")
	ErrNilInput             = errors.New("input is nil")
	ErrInvalidID            = errors.New("task ID is invalid or missing")
	ErrMissingBlockReason   = errors.New("reason for blocking is required")
	ErrInvalidTimeout       = errors.New("timeout minutes must be a positive integer")
	ErrCircularDependency   = errors.New("circular dependency detected")
	ErrInvalidBlockedByTask = errors.New("blocked_by contains a non-existent task ID")
	ErrDuplicateBlockedBy   = errors.New("blocked_by contains duplicate task IDs")
)

// NewCircularDependencyError creates a new error for circular dependency detection
func NewCircularDependencyError(taskID string, blockedBy []string) error {
	return &CircularDependencyError{
		TaskID:    taskID,
		BlockedBy: blockedBy,
	}
}

// CircularDependencyError wraps details about which task would create a cycle
type CircularDependencyError struct {
	TaskID    string
	BlockedBy []string
}

func (e *CircularDependencyError) Error() string {
	return "circular dependency detected: task cannot block itself or create a cycle"
}

// NewInvalidBlockedByTaskError creates a new error for invalid blocked_by task
func NewInvalidBlockedByTaskError(taskID string) error {
	return &InvalidBlockedByTaskError{
		TaskID: taskID,
	}
}

// InvalidBlockedByTaskError wraps the invalid task ID
type InvalidBlockedByTaskError struct {
	TaskID string
}

func (e *InvalidBlockedByTaskError) Error() string {
	return "blocked_by contains a non-existent task ID"
}

// IsCircularDependency checks if an error is a CircularDependencyError
func IsCircularDependency(err error) bool {
	if err == nil {
		return false
	}
	var circErr *CircularDependencyError
	return errors.As(err, &circErr)
}

// IsInvalidBlockedByTask checks if an error is an InvalidBlockedByTaskError
func IsInvalidBlockedByTask(err error) bool {
	if err == nil {
		return false
	}
	var invErr *InvalidBlockedByTaskError
	return errors.As(err, &invErr)
}

// NewDuplicateBlockedByError creates a new error for duplicate blocked_by entries
func NewDuplicateBlockedByError(taskID string) error {
	return &DuplicateBlockedByError{
		TaskID: taskID,
	}
}

// DuplicateBlockedByError wraps the task ID that appears multiple times
type DuplicateBlockedByError struct {
	TaskID string
}

func (e *DuplicateBlockedByError) Error() string {
	return "blocked_by contains duplicate task IDs"
}

// IsDuplicateBlockedBy checks if an error is a DuplicateBlockedByError
func IsDuplicateBlockedBy(err error) bool {
	if err == nil {
		return false
	}
	var dupErr *DuplicateBlockedByError
	return errors.As(err, &dupErr)
}
