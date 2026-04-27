package validation

import "testing"

func TestTaskFieldValidator_ValidateID(t *testing.T) {
	v := NewTaskFieldValidator()

	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{
			name:    "valid ID",
			id:      "task-123",
			wantErr: nil,
		},
		{
			name:    "valid ID with numbers",
			id:      "12345",
			wantErr: nil,
		},
		{
			name:    "empty ID",
			id:      "",
			wantErr: ErrEmptyID,
		},
		{
			name:    "whitespace only ID",
			id:      "   ",
			wantErr: ErrEmptyID,
		},
		{
			name:    "tabs only ID",
			id:      "\t\t",
			wantErr: ErrEmptyID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateID(tt.id)
			if err != tt.wantErr {
				t.Errorf("ValidateID(%q) = %v, want %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

func TestTaskFieldValidator_ValidateTitle(t *testing.T) {
	v := NewTaskFieldValidator()

	tests := []struct {
		name    string
		title   string
		wantErr error
	}{
		{
			name:    "valid title",
			title:   "Implement authentication",
			wantErr: nil,
		},
		{
			name:    "valid title with special chars",
			title:   "Task #1 - Complete!",
			wantErr: nil,
		},
		{
			name:    "empty title",
			title:   "",
			wantErr: ErrEmptyTitle,
		},
		{
			name:    "whitespace only title",
			title:   "    ",
			wantErr: ErrEmptyTitle,
		},
		{
			name:    "newlines only title",
			title:   "\n\n",
			wantErr: ErrEmptyTitle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateTitle(tt.title)
			if err != tt.wantErr {
				t.Errorf("ValidateTitle(%q) = %v, want %v", tt.title, err, tt.wantErr)
			}
		})
	}
}

func TestTaskFieldValidator_ValidateMilestone(t *testing.T) {
	v := NewTaskFieldValidator()

	tests := []struct {
		name      string
		milestone string
		wantErr   error
	}{
		{
			name:      "valid milestone",
			milestone: "milestone-1",
			wantErr:   nil,
		},
		{
			name:      "valid milestone with spaces",
			milestone: "Auth Feature",
			wantErr:   nil,
		},
		{
			name:      "empty milestone",
			milestone: "",
			wantErr:   ErrEmptyMilestone,
		},
		{
			name:      "whitespace only milestone",
			milestone: "   ",
			wantErr:   ErrEmptyMilestone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateMilestone(tt.milestone)
			if err != tt.wantErr {
				t.Errorf("ValidateMilestone(%q) = %v, want %v", tt.milestone, err, tt.wantErr)
			}
		})
	}
}

func TestTaskFieldValidator_ValidateStatus(t *testing.T) {
	v := NewTaskFieldValidator()

	tests := []struct {
		name    string
		status  string
		wantErr error
	}{
		{
			name:    "valid status todo",
			status:  StatusTodo,
			wantErr: nil,
		},
		{
			name:    "valid status scheduled",
			status:  StatusScheduled,
			wantErr: nil,
		},
		{
			name:    "valid status in-progress",
			status:  StatusInProgress,
			wantErr: nil,
		},
		{
			name:    "valid status completed",
			status:  StatusCompleted,
			wantErr: nil,
		},
		{
			name:    "valid status blocked",
			status:  StatusBlocked,
			wantErr: nil,
		},
		{
			name:    "empty status",
			status:  "",
			wantErr: ErrEmptyStatus,
		},
		{
			name:    "whitespace only status",
			status:  "   ",
			wantErr: ErrEmptyStatus,
		},
		{
			name:    "invalid status unknown",
			status:  "unknown",
			wantErr: ErrInvalidStatus,
		},
		{
			name:    "invalid status uppercase",
			status:  "TODO",
			wantErr: ErrInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateStatus(tt.status)
			if err != tt.wantErr {
				t.Errorf("ValidateStatus(%q) = %v, want %v", tt.status, err, tt.wantErr)
			}
		})
	}
}

func TestTaskFieldValidator_ValidateReason(t *testing.T) {
	v := NewTaskFieldValidator()

	tests := []struct {
		name    string
		reason  string
		wantErr error
	}{
		{
			name:    "valid reason",
			reason:  "Waiting on API",
			wantErr: nil,
		},
		{
			name:    "valid reason with special chars",
			reason:  "Blocked by: dependency #123",
			wantErr: nil,
		},
		{
			name:    "empty reason",
			reason:  "",
			wantErr: ErrEmptyReason,
		},
		{
			name:    "whitespace only reason",
			reason:  "   ",
			wantErr: ErrEmptyReason,
		},
		{
			name:    "newlines only reason",
			reason:  "\n\n",
			wantErr: ErrEmptyReason,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateReason(tt.reason)
			if err != tt.wantErr {
				t.Errorf("ValidateReason(%q) = %v, want %v", tt.reason, err, tt.wantErr)
			}
		})
	}
}

func TestTaskValidator_ValidateTaskInput(t *testing.T) {
	v := NewTaskValidator()

	tests := []struct {
		name      string
		id        string
		title     string
		milestone string
		status    string
		reason    string
		wantErr   error
	}{
		{
			name:      "valid input with todo status",
			id:        "task-1",
			title:     "Implement feature",
			milestone: "milestone-1",
			status:    StatusTodo,
			reason:    "",
			wantErr:   nil,
		},
		{
			name:      "valid input with blocked status",
			id:        "task-2",
			title:     "Implement feature",
			milestone: "milestone-1",
			status:    StatusBlocked,
			reason:    "Waiting on dependency",
			wantErr:   nil,
		},
		{
			name:      "blocked status without reason",
			id:        "task-3",
			title:     "Implement feature",
			milestone: "milestone-1",
			status:    StatusBlocked,
			reason:    "",
			wantErr:   ErrEmptyReason,
		},
		{
			name:      "empty ID",
			id:        "",
			title:     "Implement feature",
			milestone: "milestone-1",
			status:    StatusTodo,
			reason:    "",
			wantErr:   ErrEmptyID,
		},
		{
			name:      "empty title",
			id:        "task-1",
			title:     "",
			milestone: "milestone-1",
			status:    StatusTodo,
			reason:    "",
			wantErr:   ErrEmptyTitle,
		},
		{
			name:      "empty milestone",
			id:        "task-1",
			title:     "Implement feature",
			milestone: "",
			status:    StatusTodo,
			reason:    "",
			wantErr:   ErrEmptyMilestone,
		},
		{
			name:      "invalid status",
			id:        "task-1",
			title:     "Implement feature",
			milestone: "milestone-1",
			status:    "invalid",
			reason:    "",
			wantErr:   ErrInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateTaskInput(tt.id, tt.title, tt.milestone, tt.status, tt.reason)
			if err != tt.wantErr {
				t.Errorf("ValidateTaskInput() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaskValidator_ValidateStatusChange(t *testing.T) {
	v := NewTaskValidator()

	tests := []struct {
		name    string
		status  string
		reason  string
		wantErr error
	}{
		{
			name:    "valid status change to todo",
			status:  StatusTodo,
			reason:  "",
			wantErr: nil,
		},
		{
			name:    "valid status change to scheduled",
			status:  StatusScheduled,
			reason:  "",
			wantErr: nil,
		},
		{
			name:    "valid status change to in-progress",
			status:  StatusInProgress,
			reason:  "",
			wantErr: nil,
		},
		{
			name:    "valid status change to completed",
			status:  StatusCompleted,
			reason:  "",
			wantErr: nil,
		},
		{
			name:    "valid status change to blocked with reason",
			status:  StatusBlocked,
			reason:  "Waiting on external API",
			wantErr: nil,
		},
		{
			name:    "blocked status without reason",
			status:  StatusBlocked,
			reason:  "",
			wantErr: ErrEmptyReason,
		},
		{
			name:    "invalid status",
			status:  "invalid",
			reason:  "",
			wantErr: ErrInvalidStatus,
		},
		{
			name:    "empty status",
			status:  "",
			reason:  "",
			wantErr: ErrEmptyStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateStatusChange(tt.status, tt.reason)
			if err != tt.wantErr {
				t.Errorf("ValidateStatusChange(%q, %q) = %v, want %v", tt.status, tt.reason, err, tt.wantErr)
			}
		})
	}
}

func TestTaskValidator_ValidateRequiredFields(t *testing.T) {
	v := NewTaskValidator()

	tests := []struct {
		name      string
		id        string
		title     string
		milestone string
		wantErr   error
	}{
		{
			name:      "all fields valid",
			id:        "task-1",
			title:     "Implement feature",
			milestone: "milestone-1",
			wantErr:   nil,
		},
		{
			name:      "empty ID",
			id:        "",
			title:     "Implement feature",
			milestone: "milestone-1",
			wantErr:   ErrEmptyID,
		},
		{
			name:      "empty title",
			id:        "task-1",
			title:     "",
			milestone: "milestone-1",
			wantErr:   ErrEmptyTitle,
		},
		{
			name:      "empty milestone",
			id:        "task-1",
			title:     "Implement feature",
			milestone: "",
			wantErr:   ErrEmptyMilestone,
		},
		{
			name:      "all fields empty",
			id:        "",
			title:     "",
			milestone: "",
			wantErr:   ErrEmptyID,
		},
		{
			name:      "whitespace ID",
			id:        "   ",
			title:     "Valid title",
			milestone: "Valid milestone",
			wantErr:   ErrEmptyID,
		},
		{
			name:      "whitespace title",
			id:        "valid-id",
			title:     "   ",
			milestone: "Valid milestone",
			wantErr:   ErrEmptyTitle,
		},
		{
			name:      "whitespace milestone",
			id:        "valid-id",
			title:     "Valid title",
			milestone: "   ",
			wantErr:   ErrEmptyMilestone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateRequiredFields(tt.id, tt.title, tt.milestone)
			if err != tt.wantErr {
				t.Errorf("ValidateRequiredFields(%q, %q, %q) = %v, want %v", tt.id, tt.title, tt.milestone, err, tt.wantErr)
			}
		})
	}
}

func TestNewTaskValidator_WithOptions(t *testing.T) {
	customValidatorCalled := false
	customValidator := func(id string) error {
		if id == "custom-id" {
			customValidatorCalled = true
			return nil
		}
		return ErrEmptyID
	}

	v := NewTaskValidator(WithIDValidator(customValidator))

	err := v.ValidateRequiredFields("custom-id", "Valid Title", "Valid Milestone")
	if err != nil {
		t.Errorf("Custom validator should not fail for custom-id: %v", err)
	}

	err = v.ValidateRequiredFields("other-id", "Valid Title", "Valid Milestone")
	if err != ErrEmptyID {
		t.Errorf("Custom validator should fail for other-id: %v", err)
	}

	if !customValidatorCalled {
		t.Error("Custom validator was not called")
	}
}
