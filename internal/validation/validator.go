package validation

import (
	"errors"
	"strings"
)

// Validation errors
var (
	ErrEmptyID        = errors.New("ID is required")
	ErrEmptyTitle     = errors.New("title is required")
	ErrEmptyMilestone = errors.New("milestone is required")
	ErrEmptyStatus    = errors.New("status is required")
	ErrInvalidStatus  = errors.New("invalid status value")
	ErrEmptyReason    = errors.New("reason is required when status is blocked")
	ErrWhitespaceOnly = errors.New("field cannot be only whitespace")
)

// TaskFieldValidator validates individual task fields
type TaskFieldValidator struct{}

// NewTaskFieldValidator creates a new TaskFieldValidator
func NewTaskFieldValidator() *TaskFieldValidator {
	return &TaskFieldValidator{}
}

// ValidateID validates the task ID
func (v *TaskFieldValidator) ValidateID(id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrEmptyID
	}
	return nil
}

// ValidateTitle validates the task title
func (v *TaskFieldValidator) ValidateTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return ErrEmptyTitle
	}
	return nil
}

// ValidateMilestone validates the milestone field
func (v *TaskFieldValidator) ValidateMilestone(milestone string) error {
	if strings.TrimSpace(milestone) == "" {
		return ErrEmptyMilestone
	}
	return nil
}

// ValidateStatus validates the task status
func (v *TaskFieldValidator) ValidateStatus(status string) error {
	if strings.TrimSpace(status) == "" {
		return ErrEmptyStatus
	}
	if !IsValidStatus(status) {
		return ErrInvalidStatus
	}
	return nil
}

// ValidateReason validates the reason field for blocked tasks
func (v *TaskFieldValidator) ValidateReason(reason string) error {
	if strings.TrimSpace(reason) == "" {
		return ErrEmptyReason
	}
	return nil
}

// TaskValidator validates complete task input
type TaskValidator struct {
	validateID        func(string) error
	validateTitle     func(string) error
	validateMilestone func(string) error
	validateStatus    func(string) error
	validateReason    func(string) error
}

// TaskValidatorOption is a functional option for TaskValidator
type TaskValidatorOption func(*TaskValidator)

// WithIDValidator sets a custom ID validator
func WithIDValidator(fn func(string) error) TaskValidatorOption {
	return func(tv *TaskValidator) {
		tv.validateID = fn
	}
}

// WithTitleValidator sets a custom title validator
func WithTitleValidator(fn func(string) error) TaskValidatorOption {
	return func(tv *TaskValidator) {
		tv.validateTitle = fn
	}
}

// WithMilestoneValidator sets a custom milestone validator
func WithMilestoneValidator(fn func(string) error) TaskValidatorOption {
	return func(tv *TaskValidator) {
		tv.validateMilestone = fn
	}
}

// WithStatusValidator sets a custom status validator
func WithStatusValidator(fn func(string) error) TaskValidatorOption {
	return func(tv *TaskValidator) {
		tv.validateStatus = fn
	}
}

// WithReasonValidator sets a custom reason validator
func WithReasonValidator(fn func(string) error) TaskValidatorOption {
	return func(tv *TaskValidator) {
		tv.validateReason = fn
	}
}

// NewTaskValidator creates a new TaskValidator with optional custom validators
func NewTaskValidator(opts ...TaskValidatorOption) *TaskValidator {
	tv := &TaskValidator{
		validateID:        new(TaskFieldValidator).ValidateID,
		validateTitle:     new(TaskFieldValidator).ValidateTitle,
		validateMilestone: new(TaskFieldValidator).ValidateMilestone,
		validateStatus:    new(TaskFieldValidator).ValidateStatus,
		validateReason:    new(TaskFieldValidator).ValidateReason,
	}
	for _, opt := range opts {
		opt(tv)
	}
	return tv
}

// ValidateTaskInput validates a task input for creation/update
func (tv *TaskValidator) ValidateTaskInput(id, title, milestone, status, reason string) error {
	if err := tv.validateID(id); err != nil {
		return err
	}
	if err := tv.validateTitle(title); err != nil {
		return err
	}
	if err := tv.validateMilestone(milestone); err != nil {
		return err
	}
	if err := tv.validateStatus(status); err != nil {
		return err
	}
	// Reason is required only for blocked status
	if StatusRequiresReason(status) {
		if err := tv.validateReason(reason); err != nil {
			return err
		}
	}
	return nil
}

// ValidateStatusChange validates a status change with optional reason
func (tv *TaskValidator) ValidateStatusChange(status, reason string) error {
	if err := tv.validateStatus(status); err != nil {
		return err
	}
	// Reason is required only for blocked status
	if StatusRequiresReason(status) {
		if err := tv.validateReason(reason); err != nil {
			return err
		}
	}
	return nil
}

// ValidateRequiredFields validates all required fields for a task
func (tv *TaskValidator) ValidateRequiredFields(id, title, milestone string) error {
	if err := tv.validateID(id); err != nil {
		return err
	}
	if err := tv.validateTitle(title); err != nil {
		return err
	}
	if err := tv.validateMilestone(milestone); err != nil {
		return err
	}
	return nil
}
