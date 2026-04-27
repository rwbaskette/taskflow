package service

import (
	"testing"
	"time"

	"github.com/rwbaskette/taskflow/internal/db"
)

func TestIsTimedOut_NilTask(t *testing.T) {
	result := IsTimedOut(nil, 30)
	if result != false {
		t.Errorf("IsTimedOut(nil, 30) = %v, want false", result)
	}
}

func TestIsTimedOut_NonInProgressStatus(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		timeoutMin int
	}{
		{
			name:       "todo status not timed out",
			status:     "todo",
			timeoutMin: 30,
		},
		{
			name:       "completed status not timed out",
			status:     "done",
			timeoutMin: 30,
		},
		{
			name:       "blocked status not timed out",
			status:     "blocked",
			timeoutMin: 30,
		},
		{
			name:       "scheduled status not timed out",
			status:     "scheduled",
			timeoutMin: 30,
		},
		{
			name:       "empty status not timed out",
			status:     "",
			timeoutMin: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &db.Task{
				ID:          "task-1",
				Status:      tt.status,
				LastUpdated: time.Now().Add(-1 * time.Hour),
			}
			result := IsTimedOut(task, tt.timeoutMin)
			if result != false {
				t.Errorf("IsTimedOut(%+v) = %v, want false", task, result)
			}
		})
	}
}

func TestIsTimedOut_TimeoutBoundaries(t *testing.T) {
	// Use time.Now() as reference since IsTimedOut uses time.Since()
	now := time.Now().UTC()

	tests := []struct {
		name           string
		lastUpdated    time.Time
		timeoutMin     int
		expectTimedOut bool
	}{
		{
			name:           "well under timeout - not timed out",
			lastUpdated:    now.Add(-5 * time.Minute),
			timeoutMin:     30,
			expectTimedOut: false,
		},
		{
			name:           "just under timeout boundary - not timed out",
			lastUpdated:    now.Add(-29 * time.Minute),
			timeoutMin:     30,
			expectTimedOut: false,
		},
		{
			name:           "just over timeout boundary - timed out",
			lastUpdated:    now.Add(-31 * time.Minute),
			timeoutMin:     30,
			expectTimedOut: true,
		},
		{
			name:           "way over timeout - timed out",
			lastUpdated:    now.Add(-2 * time.Hour),
			timeoutMin:     30,
			expectTimedOut: true,
		},
		{
			name:           "very recent task - not timed out",
			lastUpdated:    now,
			timeoutMin:     30,
			expectTimedOut: false,
		},
		{
			name:           "zero timeout - times out if any elapsed time",
			lastUpdated:    now.Add(-1 * time.Hour),
			timeoutMin:     0,
			expectTimedOut: true,
		},
		{
			name:           "negative timeout - times out if any elapsed time",
			lastUpdated:    now.Add(-1 * time.Hour),
			timeoutMin:     -1,
			expectTimedOut: true,
		},
		{
			name:           "very short timeout with no elapsed time - not timed out",
			lastUpdated:    now,
			timeoutMin:     1,
			expectTimedOut: false,
		},
		{
			name:           "very long timeout with short elapsed time - not timed out",
			lastUpdated:    now.Add(-1 * time.Hour),
			timeoutMin:     24 * 60, // 24 hours
			expectTimedOut: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &db.Task{
				ID:          "task-1",
				Status:      "in_progress",
				LastUpdated: tt.lastUpdated,
			}
			result := IsTimedOut(task, tt.timeoutMin)
			if result != tt.expectTimedOut {
				t.Errorf("IsTimedOut() = %v, want %v", result, tt.expectTimedOut)
			}
		})
	}
}

func TestIsTimedOut_VariousStatuses(t *testing.T) {
	now := time.Now().Add(-1 * time.Hour)

	statuses := []string{"todo", "scheduled", "in_progress", "done", "blocked", "unknown", ""}

	for _, status := range statuses {
		t.Run("status_"+status, func(t *testing.T) {
			task := &db.Task{
				ID:          "task-1",
				Status:      status,
				LastUpdated: now,
			}

			// Only in_progress should potentially be timed out
			result := IsTimedOut(task, 30)
			if status == "in_progress" {
				if result != true {
					t.Errorf("IsTimedOut() for in_progress = %v, want true", result)
				}
			} else {
				if result != false {
					t.Errorf("IsTimedOut() for %q = %v, want false", status, result)
				}
			}
		})
	}
}

func TestGetTimedOutTasks_EmptyList(t *testing.T) {
	tasks := []db.Task{}
	result := GetTimedOutTasks(tasks, 30)
	if len(result) != 0 {
		t.Errorf("GetTimedOutTasks(empty list) = %v, want empty list", result)
	}
}

func TestGetTimedOutTasks_NilList(t *testing.T) {
	var tasks []db.Task = nil
	result := GetTimedOutTasks(tasks, 30)
	if len(result) != 0 {
		t.Errorf("GetTimedOutTasks(nil) = %v, want empty list", result)
	}
}

func TestGetTimedOutTasks_MixedTasks(t *testing.T) {
	now := time.Now().UTC()
	oneHourAgo := now.Add(-1 * time.Hour)
	tenMinutesAgo := now.Add(-10 * time.Minute)

	tasks := []db.Task{
		{
			ID:          "task-1",
			Status:      "in_progress",
			LastUpdated: oneHourAgo,
		},
		{
			ID:          "task-2",
			Status:      "todo",
			LastUpdated: oneHourAgo,
		},
		{
			ID:          "task-3",
			Status:      "in_progress",
			LastUpdated: tenMinutesAgo,
		},
		{
			ID:          "task-4",
			Status:      "done",
			LastUpdated: oneHourAgo,
		},
		{
			ID:          "task-5",
			Status:      "blocked",
			LastUpdated: oneHourAgo,
		},
	}

	// 30 minute timeout
	result := GetTimedOutTasks(tasks, 30)

	// Only task-1 should be timed out (in_progress and 1 hour old)
	if len(result) != 1 {
		t.Errorf("GetTimedOutTasks() returned %d tasks, want 1", len(result))
	}
	if len(result) > 0 && result[0].ID != "task-1" {
		t.Errorf("Expected task-1 to be timed out, got %s", result[0].ID)
	}
}

func TestGetTimedOutTasks_AllTimedOut(t *testing.T) {
	now := time.Now().UTC()
	oneHourAgo := now.Add(-1 * time.Hour)

	tasks := []db.Task{
		{
			ID:          "task-1",
			Status:      "in_progress",
			LastUpdated: oneHourAgo,
		},
		{
			ID:          "task-2",
			Status:      "in_progress",
			LastUpdated: oneHourAgo,
		},
		{
			ID:          "task-3",
			Status:      "in_progress",
			LastUpdated: oneHourAgo,
		},
	}

	result := GetTimedOutTasks(tasks, 30)

	if len(result) != 3 {
		t.Errorf("GetTimedOutTasks() returned %d tasks, want 3", len(result))
	}
}

func TestGetTimedOutTasks_NoneTimedOut(t *testing.T) {
	now := time.Now().UTC()

	tasks := []db.Task{
		{
			ID:          "task-1",
			Status:      "in_progress",
			LastUpdated: now,
		},
		{
			ID:          "task-2",
			Status:      "in_progress",
			LastUpdated: now,
		},
		{
			ID:          "task-3",
			Status:      "in_progress",
			LastUpdated: now,
		},
	}

	result := GetTimedOutTasks(tasks, 30)

	if len(result) != 0 {
		t.Errorf("GetTimedOutTasks() returned %d tasks, want 0", len(result))
	}
}

func TestGetTimedOutTasks_DifferentTimeouts(t *testing.T) {
	now := time.Now().UTC()
	oneHourAgo := now.Add(-1 * time.Hour)

	tasks := []db.Task{
		{
			ID:          "task-1",
			Status:      "in_progress",
			LastUpdated: oneHourAgo,
		},
		{
			ID:          "task-2",
			Status:      "in_progress",
			LastUpdated: oneHourAgo,
		},
	}

	// With 90 minute timeout, neither should be timed out
	result := GetTimedOutTasks(tasks, 90)
	if len(result) != 0 {
		t.Errorf("GetTimedOutTasks(90 min) = %d, want 0", len(result))
	}

	// With 15 minute timeout, both should be timed out
	result = GetTimedOutTasks(tasks, 15)
	if len(result) != 2 {
		t.Errorf("GetTimedOutTasks(15 min) = %d, want 2", len(result))
	}
}

func TestGetTimedOutTasks_PreservesOriginalTasks(t *testing.T) {
	now := time.Now().UTC()
	oneHourAgo := now.Add(-1 * time.Hour)

	tasks := []db.Task{
		{
			ID:          "task-1",
			Status:      "in_progress",
			LastUpdated: oneHourAgo,
			Title:       "Original Title",
		},
	}

	_ = GetTimedOutTasks(tasks, 30)

	// Original tasks slice should not be modified
	if len(tasks) != 1 {
		t.Errorf("Original tasks slice length = %d, want 1", len(tasks))
	}
	if tasks[0].ID != "task-1" {
		t.Errorf("Original task ID = %s, want task-1", tasks[0].ID)
	}
}
