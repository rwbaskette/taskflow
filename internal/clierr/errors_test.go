package clierr

import (
	"reflect"
	"testing"
)

func TestValidateID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
		wantMsg string // expected err.Error() when wantErr
	}{
		{"valid id", "1", false, ""},
		{"valid id with letters", "abc123", false, ""},
		{"valid id with surrounding whitespace", "  id-7  ", false, ""},
		{"empty id", "", true, "Invalid value for task-id: cannot be empty\nSuggestion: Provide a valid task ID"},
		{"whitespace only id", "   ", true, "Invalid value for task-id: cannot be empty\nSuggestion: Provide a valid task ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != tt.wantMsg {
				t.Errorf("ValidateID() message = %q, want %q", err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestValidateStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		wantErr bool
		wantMsg string // expected err.Error() when wantErr
	}{
		{"valid pending", "pending", false, ""},
		{"valid in-progress", "in-progress", false, ""},
		{"valid completed", "completed", false, ""},
		{"valid blocked", "blocked", false, ""},
		{"timed-out is invalid", "timed-out", true,
			"Invalid value for status: 'timed-out' is not valid\nSuggestion: Valid statuses: todo, in_progress, done, blocked, all (or aliases: pending, in-progress, completed)"},
		{"valid uppercase", "PENDING", false, ""},
		{"valid mixed case", "Completed", false, ""},
		{"valid padded", "  done  ", false, ""},
		{"special value all", "all", false, ""},
		{"special value ALL", "ALL", false, ""},
		// "all" is normalized once (trim + lowercase) before both the special
		// case check and the alias lookup, so padded/cased "all" is accepted
		// exactly like the canonical statuses.
		{"padded lowercase all", " all ", false, ""},
		{"padded uppercase all", " ALL ", false, ""},
		{"invalid status", "invalid", true,
			"Invalid value for status: 'invalid' is not valid\nSuggestion: Valid statuses: todo, in_progress, done, blocked, all (or aliases: pending, in-progress, completed)"},
		{"invalid empty string", "", true,
			"Invalid value for status: '' is not valid\nSuggestion: Valid statuses: todo, in_progress, done, blocked, all (or aliases: pending, in-progress, completed)"},
		{"invalid random", "random-status", true,
			"Invalid value for status: 'random-status' is not valid\nSuggestion: Valid statuses: todo, in_progress, done, blocked, all (or aliases: pending, in-progress, completed)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStatus(tt.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != tt.wantMsg {
				t.Errorf("ValidateStatus() message = %q, want %q", err.Error(), tt.wantMsg)
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
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name       string
		field      string
		message    string
		suggestion string
		wantMsg    string
	}{
		{
			name:       "with suggestion appends Suggestion line",
			field:      "title",
			message:    "cannot be empty",
			suggestion: "Provide a title",
			wantMsg:    "Invalid value for title: cannot be empty\nSuggestion: Provide a title",
		},
		{
			name:       "without suggestion has no Suggestion line",
			field:      "actor",
			message:    "is not valid",
			suggestion: "",
			wantMsg:    "Invalid value for actor: is not valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidationError(tt.field, tt.message, tt.suggestion)

			if err.Code != ErrInvalidArgument {
				t.Errorf("Code = %v, want %v", err.Code, ErrInvalidArgument)
			}
			if err.Error() != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestMissingArgumentError(t *testing.T) {
	err := MissingArgumentError("title", "task add [title]")

	if err.Code != ErrMissingArgument {
		t.Errorf("Code = %v, want %v", err.Code, ErrMissingArgument)
	}
	wantMsg := "Missing required argument: title\nUsage: task add [title]"
	if err.Error() != wantMsg {
		t.Errorf("Error() = %q, want %q", err.Error(), wantMsg)
	}
}

func TestResourceNotFoundError(t *testing.T) {
	err := ResourceNotFoundError("task", "123")

	if err.Code != ErrResourceNotFound {
		t.Errorf("Code = %v, want %v", err.Code, ErrResourceNotFound)
	}
	wantMsg := "No task found with id '123'"
	if err.Error() != wantMsg {
		t.Errorf("Error() = %q, want %q", err.Error(), wantMsg)
	}
}

func TestInvalidStatusTransitionError(t *testing.T) {
	tests := []struct {
		name          string
		taskID        string
		currentStatus string
	}{
		{"blocked task", "task-1", "blocked"},
		{"done task", "abc-123", "done"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InvalidStatusTransitionError(tt.taskID, tt.currentStatus)

			if err.Code != ErrInvalidStatusTransition {
				t.Errorf("Code = %v, want %v", err.Code, ErrInvalidStatusTransition)
			}
			wantMsg := "Task " + tt.taskID + " is in '" + tt.currentStatus +
				"' status and cannot be unblocked. Only tasks in 'blocked' status can be unblocked."
			if err.Error() != wantMsg {
				t.Errorf("Error() = %q, want %q", err.Error(), wantMsg)
			}
			wantTask := map[string]string{"id": tt.taskID, "current_status": tt.currentStatus}
			if !reflect.DeepEqual(err.Task, wantTask) {
				t.Errorf("Task = %v, want %v", err.Task, wantTask)
			}
		})
	}
}

func TestIDErrorConstructors(t *testing.T) {
	t.Run("MissingIDError", func(t *testing.T) {
		err := MissingIDError()
		if err.Code != ErrMissingArgument {
			t.Errorf("Code = %v, want %v", err.Code, ErrMissingArgument)
		}
		wantMsg := "The required parameter 'id' is missing. Please provide a valid task identifier."
		if err.Error() != wantMsg {
			t.Errorf("Error() = %q, want %q", err.Error(), wantMsg)
		}
		if !reflect.DeepEqual(err.MissingParams, []string{"id"}) {
			t.Errorf("MissingParams = %v, want [id]", err.MissingParams)
		}
	})

	t.Run("EmptyIDError", func(t *testing.T) {
		err := EmptyIDError()
		if err.Code != ErrInvalidArgument {
			t.Errorf("Code = %v, want %v", err.Code, ErrInvalidArgument)
		}
		wantMsg := "The 'id' parameter must be a non-empty string."
		if err.Error() != wantMsg {
			t.Errorf("Error() = %q, want %q", err.Error(), wantMsg)
		}
		wantParams := map[string]interface{}{"id": ""}
		if !reflect.DeepEqual(err.InvalidParams, wantParams) {
			t.Errorf("InvalidParams = %v, want %v", err.InvalidParams, wantParams)
		}
	})

	t.Run("NonStringIDError carries the actual value", func(t *testing.T) {
		tests := []struct {
			name  string
			value interface{}
		}{
			{"number", 42},
			{"bool", true},
			{"object", map[string]interface{}{"nested": "value"}},
			{"nil", nil},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := NonStringIDError(tt.value)
				if err.Code != ErrInvalidArgument {
					t.Errorf("Code = %v, want %v", err.Code, ErrInvalidArgument)
				}
				wantMsg := "The 'id' parameter must be a string."
				if err.Error() != wantMsg {
					t.Errorf("Error() = %q, want %q", err.Error(), wantMsg)
				}
				wantParams := map[string]interface{}{"id": tt.value}
				if !reflect.DeepEqual(err.InvalidParams, wantParams) {
					t.Errorf("InvalidParams = %v, want %v", err.InvalidParams, wantParams)
				}
			})
		}
	})
}

// TestHandleErrorNil verifies HandleError returns normally (without exiting
// the process) when given a nil error. Non-nil errors are not testable
// in-process: HandleError calls formatCLIErrorAsJSON, which calls os.Exit(1).
func TestHandleErrorNil(t *testing.T) {
	HandleError(nil) // must return without exiting
}
