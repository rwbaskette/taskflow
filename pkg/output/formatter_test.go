package output

import (
	"testing"
)

func TestColorConstants(t *testing.T) {
	tests := []struct {
		name    string
		colorFn func() string
		want    string
	}{
		{"ColorRed", func() string { return ColorRed }, "\033[31m"},
		{"ColorGreen", func() string { return ColorGreen }, "\033[32m"},
		{"ColorYellow", func() string { return ColorYellow }, "\033[33m"},
		{"ColorBlue", func() string { return ColorBlue }, "\033[34m"},
		{"ColorReset", func() string { return ColorReset }, "\033[0m"},
		{"ColorBold", func() string { return ColorBold }, "\033[1m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.colorFn(); got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestPrintSuccess(t *testing.T) {
	// Just verify it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintSuccess panicked: %v", r)
		}
	}()

	PrintSuccess("test message")
}

func TestPrintInfo(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintInfo panicked: %v", r)
		}
	}()

	PrintInfo("test message")
}

func TestPrintWarning(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintWarning panicked: %v", r)
		}
	}()

	PrintWarning("test message")
}

func TestPrintVersion(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintVersion panicked: %v", r)
		}
	}()

	PrintVersion("task", "1.0.0", "abc123", "2024-01-01")
}

func TestPrintVersionEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintVersion panicked: %v", r)
		}
	}()

	PrintVersion("task", "1.0.0", "", "")
}

func TestFormatTaskList(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("FormatTaskList panicked: %v", r)
		}
	}()

	// Test with empty list
	FormatTaskList([]map[string]interface{}{})

	// Test with tasks
	tasks := []map[string]interface{}{
		{
			"id":        "1",
			"title":     "Test Task",
			"status":    "pending",
			"milestone": "v1.0",
			"actor":     "john",
		},
	}
	FormatTaskList(tasks)
}

func TestFormatTaskDetails(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("FormatTaskDetails panicked: %v", r)
		}
	}()

	task := map[string]interface{}{
		"id":          "1",
		"title":       "Test Task",
		"description": "A test task",
		"status":      "pending",
		"milestone":   "v1.0",
		"actor":       "john",
		"created_at":  "2024-01-01",
		"updated_at":  "2024-01-02",
	}
	FormatTaskDetails(task)
}

func TestPrintUsageError(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintUsageError panicked: %v", r)
		}
	}()

	PrintUsageError("add", "title is required")
}

func TestConfirm(t *testing.T) {
	// This is interactive, just verify function exists
	t.Log("Confirm is interactive, skipping automated test")
}

func TestPrintError(t *testing.T) {
	// Test with nil error
	PrintError(nil)

	// Test with regular error
	PrintError(&testCLIError{
		code:       "TEST_CODE",
		message:    "test error",
		details:    "test details",
		suggestion: "test suggestion",
	})
}

type testCLIError struct {
	code       string
	message    string
	details    string
	suggestion string
}

func (e *testCLIError) Error() string {
	return e.message
}
