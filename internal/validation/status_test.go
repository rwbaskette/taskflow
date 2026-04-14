package validation

import "testing"

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "valid status todo",
			status:   StatusTodo,
			expected: true,
		},
		{
			name:     "valid status scheduled",
			status:   StatusScheduled,
			expected: true,
		},
		{
			name:     "valid status in-progress",
			status:   StatusInProgress,
			expected: true,
		},
		{
			name:     "valid status completed",
			status:   StatusCompleted,
			expected: true,
		},
		{
			name:     "valid status blocked",
			status:   StatusBlocked,
			expected: true,
		},
		{
			name:     "invalid status empty",
			status:   "",
			expected: false,
		},
		{
			name:     "invalid status unknown",
			status:   "unknown",
			expected: false,
		},
		{
			name:     "invalid status invalid",
			status:   "invalid",
			expected: false,
		},
		{
			name:     "invalid status todo uppercase",
			status:   "TODO",
			expected: false,
		},
		{
			name:     "invalid status completed with space",
			status:   " completed ",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidStatus(tt.status)
			if result != tt.expected {
				t.Errorf("IsValidStatus(%q) = %v, expected %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestStatusRequiresReason(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "blocked status requires reason",
			status:   StatusBlocked,
			expected: true,
		},
		{
			name:     "todo status does not require reason",
			status:   StatusTodo,
			expected: false,
		},
		{
			name:     "scheduled status does not require reason",
			status:   StatusScheduled,
			expected: false,
		},
		{
			name:     "in-progress status does not require reason",
			status:   StatusInProgress,
			expected: false,
		},
		{
			name:     "completed status does not require reason",
			status:   StatusCompleted,
			expected: false,
		},
		{
			name:     "empty status does not require reason",
			status:   "",
			expected: false,
		},
		{
			name:     "invalid status does not require reason",
			status:   "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StatusRequiresReason(tt.status)
			if result != tt.expected {
				t.Errorf("StatusRequiresReason(%q) = %v, expected %v", tt.status, result, tt.expected)
			}
		})
	}
}
