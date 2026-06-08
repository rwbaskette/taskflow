// Package errors provides error handling and validation for the CLI.
package errors

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Colors for terminal output
const (
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
)

// ErrorCode represents categorized error types.
type ErrorCode string

const (
	// Validation errors
	ErrInvalidArgument  ErrorCode = "INVALID_ARGUMENT"
	ErrMissingArgument  ErrorCode = "MISSING_ARGUMENT"
	ErrInvalidFormat    ErrorCode = "INVALID_FORMAT"
	ErrResourceNotFound ErrorCode = "RESOURCE_NOT_FOUND"
	ErrResourceExists   ErrorCode = "RESOURCE_EXISTS"
	ErrPermissionDenied ErrorCode = "PERMISSION_DENIED"

	// System errors
	ErrDatabaseError           ErrorCode = "DATABASE_ERROR"
	ErrFileNotFound            ErrorCode = "FILE_NOT_FOUND"
	ErrConfiguration           ErrorCode = "CONFIGURATION_ERROR"
	ErrUnexpected              ErrorCode = "UNEXPECTED_ERROR"
	ErrInvalidStatusTransition ErrorCode = "INVALID_STATUS_TRANSITION"
)

// CLIError represents a structured CLI error with context.
type CLIError struct {
	Code       ErrorCode
	Message    string
	Details    string
	Suggestion string
	Cause      error
}

// Error implements the error interface.
func (e *CLIError) Error() string {
	return e.Message
}

// Unwrap returns the underlying cause of the error.
func (e *CLIError) Unwrap() error {
	return e.Cause
}

// NewCLIError creates a new CLI error with the given parameters.
func NewCLIError(code ErrorCode, message string, cause error) *CLIError {
	return &CLIError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// WithDetails adds detailed context to an existing error.
func (e *CLIError) WithDetails(details string) *CLIError {
	e.Details = details
	return e
}

// WithSuggestion adds a helpful suggestion to an existing error.
func (e *CLIError) WithSuggestion(suggestion string) *CLIError {
	e.Suggestion = suggestion
	return e
}

// ValidationError creates a validation error with optional suggestion.
func ValidationError(field, message, suggestion string) *CLIError {
	suggestionMsg := ""
	if suggestion != "" {
		suggestionMsg = fmt.Sprintf("\nSuggestion: %s", suggestion)
	}
	return &CLIError{
		Code:       ErrInvalidArgument,
		Message:    fmt.Sprintf("Invalid value for %s: %s%s", field, message, suggestionMsg),
		Details:    field,
		Suggestion: suggestion,
	}
}

// MissingArgumentError creates an error for missing required arguments.
func MissingArgumentError(argName, usage string) *CLIError {
	return &CLIError{
		Code:       ErrMissingArgument,
		Message:    fmt.Sprintf("Missing required argument: %s\nUsage: %s", argName, usage),
		Details:    argName,
		Suggestion: fmt.Sprintf("Run 'task <command> --help' for usage information"),
	}
}

// ResourceNotFoundError creates an error for missing resources.
func ResourceNotFoundError(resourceType, resourceID string) *CLIError {
	return &CLIError{
		Code:       ErrResourceNotFound,
		Message:    fmt.Sprintf("No %s found with id '%s'", resourceType, resourceID),
		Details:    resourceID,
		Suggestion: "Use 'task list' to see available tasks",
	}
}

// InvalidStatusTransitionError creates an error for invalid task status transitions.
// The Details field stores the task context as a JSON object for proper JSON output.
func InvalidStatusTransitionError(taskID, currentStatus string) *CLIError {
	taskDetails, _ := json.Marshal(map[string]interface{}{
		"id":             taskID,
		"current_status": currentStatus,
	})
	return &CLIError{
		Code:       ErrInvalidStatusTransition,
		Message:    fmt.Sprintf("Task %s is in '%s' status and cannot be unblocked. Only tasks in 'blocked' status can be unblocked.", taskID, currentStatus),
		Details:    string(taskDetails),
		Suggestion: "Use 'task list --status blocked' to find blocked tasks",
	}
}

// InvalidArgumentCountError creates an error for wrong argument count.
func InvalidArgumentCountError(expected, actual int, command string) *CLIError {
	return &CLIError{
		Code:       ErrInvalidArgument,
		Message:    fmt.Sprintf("Invalid argument count: expected %d, got %d", expected, actual),
		Suggestion: fmt.Sprintf("Run 'task %s --help' for correct usage", command),
	}
}

// PrintError prints a formatted error message to stderr (human-readable).
func PrintError(err error) {
	if err == nil {
		return
	}

	var cliErr *CLIError
	if cli, ok := err.(*CLIError); ok {
		cliErr = cli
	} else {
		cliErr = &CLIError{
			Code:    ErrUnexpected,
			Message: err.Error(),
		}
	}

	if cliErr == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "\n%sError: %s%s\n", ColorRed, cliErr.Message, ColorReset)

	if cliErr.Details != "" {
		fmt.Fprintf(os.Stderr, "  Details: %s\n", cliErr.Details)
	}

	if cliErr.Suggestion != "" {
		fmt.Fprintf(os.Stderr, "%s  %s%s\n", ColorBlue, cliErr.Suggestion, ColorReset)
	}

	fmt.Fprintf(os.Stderr, "  Code: %s\n\n", cliErr.Code)
}

// FormatCLIErrorAsJSON formats a CLIError into a JSON response matching the
// spec-defined error response structures for task_unblock. It outputs the JSON
// to stderr and exits with code 1.
func FormatCLIErrorAsJSON(cliErr *CLIError) {
	var body map[string]interface{}

	switch cliErr.Code {
	case ErrResourceNotFound:
		// RESOURCE_NOT_FOUND: {"status":"error","error_code":"RESOURCE_NOT_FOUND","message":"..."}
		body = map[string]interface{}{
			"status":     "error",
			"error_code": string(cliErr.Code),
			"message":    cliErr.Message,
		}

	case ErrInvalidStatusTransition:
		// INVALID_STATUS_TRANSITION: {"status":"error","error_code":"INVALID_STATUS_TRANSITION","message":"...","task":{"id":"...","current_status":"..."}}
		task := map[string]interface{}{
			"id":             cliErr.Details,
			"current_status": cliErr.Details,
		}
		// Parse the task details from the Details field
		var taskDetails map[string]interface{}
		if err := json.Unmarshal([]byte(cliErr.Details), &taskDetails); err == nil {
			if id, ok := taskDetails["id"]; ok {
				task["id"] = id
			}
			if status, ok := taskDetails["current_status"]; ok {
				task["current_status"] = status
			}
		}
		body = map[string]interface{}{
			"status":     "error",
			"error_code": string(cliErr.Code),
			"message":    cliErr.Message,
			"task":       task,
		}

	case ErrMissingArgument:
		// MISSING_ARGUMENT: {"status":"error","error_code":"MISSING_ARGUMENT","message":"...","missing_parameters":["id"]}
		var missingParams []string
		if err := json.Unmarshal([]byte(cliErr.Details), &missingParams); err == nil {
			body = map[string]interface{}{
				"status":             "error",
				"error_code":         string(cliErr.Code),
				"message":            cliErr.Message,
				"missing_parameters": missingParams,
			}
		} else {
			body = map[string]interface{}{
				"status":     "error",
				"error_code": string(cliErr.Code),
				"message":    cliErr.Message,
			}
		}

	case ErrInvalidArgument:
		// INVALID_ARGUMENT: {"status":"error","error_code":"INVALID_ARGUMENT","message":"...","invalid_parameters":{"id":"..."}}
		var invalidParams map[string]interface{}
		if err := json.Unmarshal([]byte(cliErr.Details), &invalidParams); err == nil {
			body = map[string]interface{}{
				"status":             "error",
				"error_code":         string(cliErr.Code),
				"message":            cliErr.Message,
				"invalid_parameters": invalidParams,
			}
		} else {
			body = map[string]interface{}{
				"status":     "error",
				"error_code": string(cliErr.Code),
				"message":    cliErr.Message,
			}
		}

	default:
		// For any other error, output a generic error JSON
		body = map[string]interface{}{
			"status":     "error",
			"error_code": string(cliErr.Code),
			"message":    cliErr.Message,
		}
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		// Fallback to human-readable output if JSON marshaling fails
		PrintError(cliErr)
		return
	}

	fmt.Fprintln(os.Stderr, string(jsonBytes))
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
		FormatCLIErrorAsJSON(cliErr)
	} else {
		// Non-CLIError: format as unexpected error and output JSON
		cliErr := &CLIError{
			Code:    ErrUnexpected,
			Message: err.Error(),
		}
		FormatCLIErrorAsJSON(cliErr)
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
		Code:       ErrMissingArgument,
		Message:    "The required parameter 'id' is missing. Please provide a valid task identifier.",
		Details:    `["id"]`,
		Suggestion: "Provide a valid task ID in the JSON document",
	}
}

// EmptyIDError creates an error for an empty 'id' parameter.
func EmptyIDError() *CLIError {
	invalidParams, _ := json.Marshal(map[string]interface{}{
		"id": "",
	})
	return &CLIError{
		Code:       ErrInvalidArgument,
		Message:    "The 'id' parameter must be a non-empty string.",
		Details:    string(invalidParams),
		Suggestion: "Provide a non-empty task ID in the JSON document",
	}
}

// NonStringIDError creates an error when the 'id' parameter is not a string type.
func NonStringIDError(actualValue interface{}) *CLIError {
	invalidParams, _ := json.Marshal(map[string]interface{}{
		"id": actualValue,
	})
	return &CLIError{
		Code:       ErrInvalidArgument,
		Message:    "The 'id' parameter must be a string.",
		Details:    string(invalidParams),
		Suggestion: "Provide a string value for the 'id' parameter",
	}
}

// ValidateStatus checks if a status value is valid.
func ValidateStatus(status string) error {
	// "all" is a special value meaning show all statuses
	if strings.ToLower(status) == "all" {
		return nil
	}

	// Map aliases to canonical status values
	aliasMap := map[string]string{
		"pending":     "todo",
		"in-progress": "in_progress",
		"inprogress":  "in_progress",
		"completed":   "done",
		"timed-out":   "blocked",
		"timedout":    "blocked",
		"todo":        "todo",
		"in_progress": "in_progress",
		"done":        "done",
		"blocked":     "blocked",
	}

	// Normalize the status (trim whitespace)
	statusNormalized := strings.TrimSpace(status)
	statusLower := strings.ToLower(statusNormalized)

	// Check if it's a valid alias or direct value
	if canonical, ok := aliasMap[statusLower]; ok {
		_ = canonical // Alias mapping successful
		return nil
	}

	validStatuses := []string{"todo", "in_progress", "done", "blocked", "all"}
	return ValidationError(
		"status",
		fmt.Sprintf("'%s' is not valid", status),
		fmt.Sprintf("Valid statuses: %s (or aliases: pending, in-progress, completed, timed-out)",
			strings.Join(validStatuses, ", ")),
	)
}

// ValidateMilestone checks if a milestone is valid.
func ValidateMilestone(milestone string) error {
	if milestone == "" {
		return nil // milestone is optional
	}
	if strings.TrimSpace(milestone) == "" {
		return ValidationError("milestone", "cannot be empty or whitespace", "Provide a valid milestone name")
	}
	return nil
}

// ValidateActor checks if an actor is valid.
func ValidateActor(actor string) error {
	if actor == "" {
		return nil // actor is optional
	}
	if strings.TrimSpace(actor) == "" {
		return ValidationError("actor", "cannot be empty or whitespace", "Provide a valid actor name")
	}
	return nil
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

// ValidateFlag checks if a flag value is valid.
func ValidateFlag(flagName, value string, required bool) error {
	if required && strings.TrimSpace(value) == "" {
		return MissingArgumentError(flagName, fmt.Sprintf("--%s is required", flagName))
	}
	return nil
}


