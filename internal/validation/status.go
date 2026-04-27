package validation

// Valid task statuses
const (
	StatusTodo       = "todo"
	StatusScheduled  = "scheduled"
	StatusInProgress = "in-progress"
	StatusCompleted  = "completed"
	StatusBlocked    = "blocked"
)

// ValidStatuses contains all valid task status values
var ValidStatuses = []string{
	StatusTodo,
	StatusScheduled,
	StatusInProgress,
	StatusCompleted,
	StatusBlocked,
}

// IsValidStatus checks if the given status is valid
func IsValidStatus(status string) bool {
	for _, valid := range ValidStatuses {
		if status == valid {
			return true
		}
	}
	return false
}

// StatusRequiresReason returns true if the status requires a reason/blocking explanation
func StatusRequiresReason(status string) bool {
	return status == StatusBlocked
}
