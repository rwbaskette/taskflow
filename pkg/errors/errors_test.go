package errors

import (
	"errors"
	"testing"
)

func TestValidateID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"valid id", "1", false},
		{"valid id with letters", "abc123", false},
		{"empty id", "", true},
		{"whitespace only id", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{"valid pending", "pending", false},
		{"valid in-progress", "in-progress", false},
		{"valid completed", "completed", false},
		{"valid blocked", "blocked", false},
		{"valid timed-out", "timed-out", false},
		{"valid uppercase", "PENDING", false},
		{"valid mixed case", "Completed", false},
		{"invalid status", "invalid", true},
		{"invalid empty string", "", true},
		{"invalid random", "random-status", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStatus(tt.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMilestone(t *testing.T) {
	tests := []struct {
		name      string
		milestone string
		wantErr   bool
	}{
		{"valid milestone", "v1.0 Release", false},
		{"valid alphanumeric", "sprint-1", false},
		{"empty milestone is optional", "", false},
		{"whitespace only", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMilestone(tt.milestone)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMilestone() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateActor(t *testing.T) {
	tests := []struct {
		name    string
		actor   string
		wantErr bool
	}{
		{"valid actor", "john", false},
		{"valid actor with spaces", "John Doe", false},
		{"empty actor is optional", "", false},
		{"whitespace only", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateActor(tt.actor)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateActor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTitle(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		wantErr bool
	}{
		{"valid title", "Implement login", false},
		{"valid long title", string(make([]byte, 500)), false}, // 500 chars is valid
		{"empty title", "", true},
		{"whitespace only title", "   ", true},
		{"title too long", string(make([]byte, 501)), true}, // 501 chars is too long
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTitle(tt.title)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTitle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFlag(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		value    string
		required bool
		wantErr  bool
	}{
		{"required flag with value", "title", "task", false, false},
		{"required flag with value", "title", "task", true, false},
		{"required flag empty", "title", "", true, true},
		{"optional flag empty", "title", "", false, false},
		{"optional flag with value", "title", "task", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFlag(tt.flagName, tt.value, tt.required)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFlag() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCLIError(t *testing.T) {
	// Test error interface implementation
	t.Run("implements error interface", func(t *testing.T) {
		err := &CLIError{
			Code:    ErrInvalidArgument,
			Message: "test error",
		}
		if err.Error() != "test error" {
			t.Errorf("Error() = %v, want %v", err.Error(), "test error")
		}
	})

	// Test Unwrap
	t.Run("unwrap returns cause", func(t *testing.T) {
		cause := errors.New("cause error")
		err := &CLIError{
			Code:    ErrDatabaseError,
			Message: "test error",
			Cause:   cause,
		}
		if err.Unwrap() != cause {
			t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), cause)
		}
	})
}

func TestNewCLIError(t *testing.T) {
	cause := errors.New("original error")
	err := NewCLIError(ErrDatabaseError, "database failure", cause)

	if err.Code != ErrDatabaseError {
		t.Errorf("Code = %v, want %v", err.Code, ErrDatabaseError)
	}
	if err.Message != "database failure" {
		t.Errorf("Message = %v, want %v", err.Message, "database failure")
	}
	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
}

func TestValidationError(t *testing.T) {
	err := ValidationError("title", "cannot be empty", "Provide a title")

	if err.Code != ErrInvalidArgument {
		t.Errorf("Code = %v, want %v", err.Code, ErrInvalidArgument)
	}
	if err.Details != "title" {
		t.Errorf("Details = %v, want %v", err.Details, "title")
	}
	if err.Suggestion != "Provide a title" {
		t.Errorf("Suggestion = %v, want %v", err.Suggestion, "Provide a title")
	}
}

func TestMissingArgumentError(t *testing.T) {
	err := MissingArgumentError("title", "task add [title]")

	if err.Code != ErrMissingArgument {
		t.Errorf("Code = %v, want %v", err.Code, ErrMissingArgument)
	}
	if err.Details != "title" {
		t.Errorf("Details = %v, want %v", err.Details, "title")
	}
}

func TestResourceNotFoundError(t *testing.T) {
	err := ResourceNotFoundError("task", "123")

	if err.Code != ErrResourceNotFound {
		t.Errorf("Code = %v, want %v", err.Code, ErrResourceNotFound)
	}
	if err.Details != "123" {
		t.Errorf("Details = %v, want %v", err.Details, "123")
	}
}

func TestInvalidArgumentCountError(t *testing.T) {
	err := InvalidArgumentCountError(2, 1, "add")

	if err.Code != ErrInvalidArgument {
		t.Errorf("Code = %v, want %v", err.Code, ErrInvalidArgument)
	}
}

func TestWithDetails(t *testing.T) {
	err := &CLIError{Code: ErrInvalidArgument, Message: "test"}
	err = err.WithDetails("additional info")

	if err.Details != "additional info" {
		t.Errorf("Details = %v, want %v", err.Details, "additional info")
	}
}

func TestWithSuggestion(t *testing.T) {
	err := &CLIError{Code: ErrInvalidArgument, Message: "test"}
	err = err.WithSuggestion("try again")

	if err.Suggestion != "try again" {
		t.Errorf("Suggestion = %v, want %v", err.Suggestion, "try again")
	}
}
