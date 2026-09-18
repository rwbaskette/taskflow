// Package errors provides error handling and validation for the CLI.
package errors

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ErrorCode represents categorized error types.
type ErrorCode string

const (
	// Validation errors
	ErrInvalidArgument  ErrorCode = "INVALID_ARGUMENT"
	ErrMissingArgument  ErrorCode = "MISSING_ARGUMENT"
	ErrResourceNotFound ErrorCode = "RESOURCE_NOT_FOUND"

	// System errors
	ErrUnexpected              ErrorCode = "UNEXPECTED_ERROR"
	ErrInvalidStatusTransition ErrorCode = "INVALID_STATUS_TRANSITION"
)

// CLIError represents a structured CLI error with context.
type CLIError struct {
	Code    ErrorCode
	Message string

	// Task carries structured context for INVALID_STATUS_TRANSITION errors.
	// Recognized keys: "id", "current_status".
	Task map[string]string
	// MissingParams carries the missing parameter names for
	// MISSING_ARGUMENT errors.
	MissingParams []string
	// InvalidParams carries the offending parameter values for
	// INVALID_ARGUMENT errors.
	InvalidParams map[string]interface{}
}

// Error implements the error interface.
func (e *CLIError) Error() string {
	return e.Message
}

// ValidationError creates a validation error with optional suggestion.
func ValidationError(field, message, suggestion string) *CLIError {
	suggestionMsg := ""
	if suggestion != "" {
		suggestionMsg = fmt.Sprintf("\nSuggestion: %s", suggestion)
	}
	return &CLIError{
		Code:    ErrInvalidArgument,
		Message: fmt.Sprintf("Invalid value for %s: %s%s", field, message, suggestionMsg),
	}
}

// MissingArgumentError creates an error for missing required arguments.
func MissingArgumentError(argName, usage string) *CLIError {
	return &CLIError{
		Code:    ErrMissingArgument,
		Message: fmt.Sprintf("Missing required argument: %s\nUsage: %s", argName, usage),
	}
}

// ResourceNotFoundError creates an error for missing resources.
func ResourceNotFoundError(resourceType, resourceID string) *CLIError {
	return &CLIError{
		Code:    ErrResourceNotFound,
		Message: fmt.Sprintf("No %s found with id '%s'", resourceType, resourceID),
	}
}

// InvalidStatusTransitionError creates an error for invalid task status transitions.
// The Task field carries the task context for proper JSON output.
func InvalidStatusTransitionError(taskID, currentStatus string) *CLIError {
	return &CLIError{
		Code:    ErrInvalidStatusTransition,
		Message: fmt.Sprintf("Task %s is in '%s' status and cannot be unblocked. Only tasks in 'blocked' status can be unblocked.", taskID, currentStatus),
		Task: map[string]string{
			"id":             taskID,
			"current_status": currentStatus,
		},
	}
}

// formatCLIErrorAsJSON writes a JSON error response for the given CLIError to
// stderr and exits with code 1. The body shape per error code matches the
// spec-defined error response structures (e.g. for task_unblock).
func formatCLIErrorAsJSON(cliErr *CLIError) {
	body := map[string]interface{}{
		"status":     "error",
		"error_code": string(cliErr.Code),
		"message":    cliErr.Message,
	}

	switch cliErr.Code {
	case ErrInvalidStatusTransition:
		// {"status":"error","error_code":"INVALID_STATUS_TRANSITION","message":"...","task":{"id":"...","current_status":"..."}}
		task := map[string]string{"id": "", "current_status": ""}
		if cliErr.Task != nil {
			task["id"] = cliErr.Task["id"]
			task["current_status"] = cliErr.Task["current_status"]
		}
		body["task"] = task

	case ErrMissingArgument:
		// {"status":"error","error_code":"MISSING_ARGUMENT","message":"...","missing_parameters":["id"]}
		if len(cliErr.MissingParams) > 0 {
			body["missing_parameters"] = cliErr.MissingParams
		}

	case ErrInvalidArgument:
		// {"status":"error","error_code":"INVALID_ARGUMENT","message":"...","invalid_parameters":{"id":"..."}}
		if len(cliErr.InvalidParams) > 0 {
			body["invalid_parameters"] = cliErr.InvalidParams
		}
	}

	// Marshaling a string-keyed map of strings/slices/maps cannot fail; the
	// branch is a safety net.
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		fmt.Fprintln(os.Stderr, cliErr.Message)
	} else {
		fmt.Fprintln(os.Stderr, string(jsonBytes))
	}
	os.Exit(1)
}

// HandleError handles an error and exits with a non-zero status.
// For CLI errors, it outputs a JSON-formatted error response to stderr.
// For other errors, it outputs a human-readable error message.
func HandleError(err error) {
	if err == nil {
		return
	}

	if cliErr, ok := err.(*CLIError); ok {
		formatCLIErrorAsJSON(cliErr)
	} else {
		// Non-CLIError: format as unexpected error and output JSON
		cliErr := &CLIError{
			Code:    ErrUnexpected,
			Message: err.Error(),
		}
		formatCLIErrorAsJSON(cliErr)
	}
}

// ValidateID checks if a task ID is valid.
func ValidateID(id string) error {
	if strings.TrimSpace(id) == "" {
		return ValidationError("task-id", "cannot be empty", "Provide a valid task ID")
	}
	return nil
}

// MissingIDError creates an error for a missing 'id' parameter.
func MissingIDError() *CLIError {
	return &CLIError{
		Code:          ErrMissingArgument,
		Message:       "The required parameter 'id' is missing. Please provide a valid task identifier.",
		MissingParams: []string{"id"},
	}
}

// EmptyIDError creates an error for an empty 'id' parameter.
func EmptyIDError() *CLIError {
	return &CLIError{
		Code:          ErrInvalidArgument,
		Message:       "The 'id' parameter must be a non-empty string.",
		InvalidParams: map[string]interface{}{"id": ""},
	}
}

// NonStringIDError creates an error when the 'id' parameter is not a string type.
func NonStringIDError(actualValue interface{}) *CLIError {
	return &CLIError{
		Code:          ErrInvalidArgument,
		Message:       "The 'id' parameter must be a string.",
		InvalidParams: map[string]interface{}{"id": actualValue},
	}
}

// validStatusAliases are the status values accepted by ValidateStatus:
// canonical statuses plus their aliases, keyed lowercase.
var validStatusAliases = map[string]bool{
	"todo":        true,
	"in_progress": true,
	"done":        true,
	"blocked":     true,
	"pending":     true,
	"in-progress": true,
	"inprogress":  true,
	"completed":   true,
}

// ValidateStatus checks if a status value is valid.
func ValidateStatus(status string) error {
	// Normalize once: both the "all" special case and the alias lookup use the
	// trimmed, lowercased value. The error message keeps the original input.
	trimmed := strings.ToLower(strings.TrimSpace(status))

	// "all" is a special value meaning show all statuses
	if trimmed == "all" {
		return nil
	}

	if validStatusAliases[trimmed] {
		return nil
	}

	return ValidationError(
		"status",
		fmt.Sprintf("'%s' is not valid", status),
		fmt.Sprintf("Valid statuses: %s (or aliases: pending, in-progress, completed)",
			strings.Join([]string{"todo", "in_progress", "done", "blocked", "all"}, ", ")),
	)
}

// validateOptionalText validates an optional text field: empty means absent
// (valid), whitespace-only is invalid.
func validateOptionalText(field, value string) error {
	if value == "" {
		return nil // field is optional
	}
	if strings.TrimSpace(value) == "" {
		return ValidationError(field, "cannot be empty or whitespace", "Provide a valid "+field+" name")
	}
	return nil
}

// ValidateMilestone checks if a milestone is valid.
func ValidateMilestone(milestone string) error {
	return validateOptionalText("milestone", milestone)
}

// ValidateActor checks if an actor is valid.
func ValidateActor(actor string) error {
	return validateOptionalText("actor", actor)
}

// ValidateTitle checks if a title is valid.
func ValidateTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return ValidationError("title", "cannot be empty", "Provide a valid task title")
	}
	if len(title) > 500 {
		return ValidationError("title", "exceeds maximum length of 500 characters", "Shorten the title")
	}
	return nil
}
