// Package errors provides error handling and validation for the CLI.
package errors

import (
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
	ErrDatabaseError ErrorCode = "DATABASE_ERROR"
	ErrFileNotFound  ErrorCode = "FILE_NOT_FOUND"
	ErrConfiguration ErrorCode = "CONFIGURATION_ERROR"
	ErrUnexpected    ErrorCode = "UNEXPECTED_ERROR"
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
		Message:    fmt.Sprintf("%s not found: %s", resourceType, resourceID),
		Details:    resourceID,
		Suggestion: "Use 'task list' to see available tasks",
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

// PrintError prints a formatted error message to stderr.
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

// HandleError handles an error and exits if it's fatal.
func HandleError(err error) {
	if err == nil {
		return
	}

	PrintError(err)

	if cliErr, ok := err.(*CLIError); ok {
		if cliErr.Code == ErrUnexpected || cliErr.Code == ErrDatabaseError {
			os.Exit(1)
		}
	}

	os.Exit(1)
}

// ValidateID checks if a task ID is valid.
func ValidateID(id string) error {
	if strings.TrimSpace(id) == "" {
		return ValidationError("task-id", "cannot be empty", "Provide a valid task ID")
	}
	return nil
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
