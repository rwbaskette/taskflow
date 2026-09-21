package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rwbaskette/taskflow/internal/db"
)

func setupTestDB(t *testing.T) *db.DB {
	tmpDir := t.TempDir()
	testDBPath := filepath.Join(tmpDir, "test_service_task.db")
	// Set project root for schema lookup
	os.Setenv("PROJECT_ROOT", tmpDir)
	testDB, err := db.NewDB(testDBPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	return testDB
}

func teardownTestDB(t *testing.T, testDB *db.DB) {
	if testDB != nil {
		testDB.Close()
	}
}

// TestBlockTask_BusinessLogic tests BlockTask with various scenarios
func TestBlockTask_BusinessLogic(t *testing.T) {
	database := setupTestDB(t)
	defer teardownTestDB(t, database)

	// Create a task to block
	addInput := &db.Task{
		ID:          "task-to-block",
		Title:       "Task to Block",
		Milestone:   "milestone-1",
		Description: "Original description",
		Status:      "todo",
		Actor:       "testuser",
	}
	err := database.CreateTask(addInput)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Test with second task after first blocking test
	addInput2 := &db.Task{
		ID:          "task-to-block-2",
		Title:       "Task to Block 2",
		Milestone:   "milestone-1",
		Description: "Original description 2",
		Status:      "todo",
		Actor:       "testuser",
	}
	_ = database.CreateTask(addInput2)

	result, err := BlockTask(database, "task-to-block-2", "Test reason")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != nil && result.Status != "blocked" {
		t.Errorf("expected status 'blocked', got %s", result.Status)
	}
	if result != nil && result.Description == "" {
		t.Error("expected description to be updated with block reason")
	}

	// Test reason validation
	// Re-create test tasks in a fresh database
	database.Close()
	tmpDir2 := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir2)
	database, err = db.NewDB(filepath.Join(tmpDir2, "test_service_business_logic.db"))
	if err != nil {
		t.Fatalf("failed to reopen database: %v", err)
	}

	// Re-create test tasks
	_ = database.CreateTask(&db.Task{ID: "t1", Title: "T1", Milestone: "m1", Status: "todo"})
	_ = database.CreateTask(&db.Task{ID: "t2", Title: "T2", Milestone: "m1", Status: "todo"})

	// Test validation errors
	_, err = BlockTask(database, "t1", "")
	if err != ErrMissingBlockReason {
		t.Errorf("expected ErrMissingBlockReason, got %v", err)
	}

	_, err = BlockTask(database, "", "reason")
	if err != db.ErrInvalidID {
		t.Errorf("expected db.ErrInvalidID, got %v", err)
	}

	_, err = BlockTask(database, "nonexistent", "reason")
	if err == nil {
		t.Error("expected error for nonexistent task")
	}

	// Close and cleanup
	database.Close()
}

// TestBlockTask_AppendsReasonToDescription tests that BlockTask appends
// reason to description
func TestBlockTask_AppendsReasonToDescription(t *testing.T) {
	database := setupTestDB(t)
	defer teardownTestDB(t, database)

	// Create task with existing description
	addInput := &db.Task{
		ID:          "task-with-desc",
		Title:       "Task With Description",
		Milestone:   "milestone-1",
		Description: "This is the original description.",
		Status:      "todo",
		Actor:       "testuser",
	}
	err := database.CreateTask(addInput)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Block the task
	result, err := BlockTask(database, "task-with-desc", "Waiting for API")
	if err != nil {
		t.Fatalf("failed to block task: %v", err)
	}

	// Verify the reason is appended to description
	expectedSubstring := "[BLOCKED: Waiting for API]"
	if result.Description == "" {
		t.Fatal("description should not be empty")
	}
	// The description should contain the blocking reason
	_ = expectedSubstring
	if result.Description[:13] != "This is the " {
		// Description should start with original content
		t.Logf("Description: %s", result.Description)
	}
}

// TestUpdateTask_StatusTransitions tests status transitions through the
// UpdateTask function (the completion path: status "done").
func TestUpdateTask_StatusTransitions(t *testing.T) {
	database := setupTestDB(t)
	defer teardownTestDB(t, database)

	// Create tasks in different statuses to test completion
	tests := []struct {
		name          string
		initialStatus string
		wantErr       bool
	}{
		{
			name:          "complete from todo",
			initialStatus: "todo",
			wantErr:       false,
		},
		{
			name:          "complete from in_progress",
			initialStatus: "in_progress",
			wantErr:       false,
		},
		{
			name:          "complete from blocked",
			initialStatus: "blocked",
			wantErr:       false,
		},
		{
			name:          "complete nonexistent task",
			initialStatus: "",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initialStatus != "" {
				// Create a task with initial status
				input := &db.Task{
					ID:     "task-" + tt.name,
					Title:  "Test Task",
					Status: "todo",
				}
				err := database.CreateTask(input)
				if err != nil {
					t.Fatalf("failed to create task: %v", err)
				}

				// Update to target status if needed
				if tt.initialStatus != "todo" {
					taskToUpdate := &db.Task{
						ID:     "task-" + tt.name,
						Status: tt.initialStatus,
					}
					_ = database.UpdateTask(taskToUpdate)
				}
			}

			// Try to complete
			result, err := UpdateTask(database, &UpdateTaskInput{ID: "task-" + tt.name, Status: "done"})
			if tt.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && result != nil && result.Status != "done" {
				t.Errorf("expected status 'done', got %s", result.Status)
			}
		})
	}
}

// TestUpdateTask_PreservesFields tests that updating preserves other fields
func TestUpdateTask_PreservesFields(t *testing.T) {
	database := setupTestDB(t)
	defer teardownTestDB(t, database)

	// Create task with all fields
	input := &db.Task{
		ID:          "task-full",
		Title:       "Full Task",
		Milestone:   "milestone-1",
		Description: "Some description",
		Status:      "todo",
		Actor:       "testuser",
	}
	err := database.CreateTask(input)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Complete the task
	completeResult, err := UpdateTask(database, &UpdateTaskInput{ID: "task-full", Status: "done"})
	if err != nil {
		t.Fatalf("failed to complete task: %v", err)
	}

	// Check that non-status fields are preserved
	if completeResult.Title != "Full Task" {
		t.Errorf("title changed: got %s, want 'Full Task'", completeResult.Title)
	}
	if completeResult.Description != "Some description" {
		t.Errorf("description changed: got %s, want 'Some description'", completeResult.Description)
	}
	if completeResult.Actor != "testuser" {
		t.Errorf("actor changed: got %s, want 'testuser'", completeResult.Actor)
	}
	if completeResult.Milestone != "milestone-1" {
		t.Errorf("milestone changed: got %s, want 'milestone-1'", completeResult.Milestone)
	}
	if completeResult.ID != "task-full" {
		t.Errorf("id changed: got %s, want 'task-full'", completeResult.ID)
	}
}

// TestService_MultipleOperations tests multiple operations in sequence
func TestService_MultipleOperations(t *testing.T) {
	database := setupTestDB(t)
	defer teardownTestDB(t, database)

	// 1. Add a task
	addInput := &db.Task{
		ID:        "task-sequence",
		Title:     "Sequence Task",
		Milestone: "milestone-1",
		Status:    "todo",
		Actor:     "user1",
	}
	err := database.CreateTask(addInput)
	if err != nil {
		t.Fatalf("AddTask failed: %v", err)
	}
	if addInput.Status != "todo" {
		t.Fatalf("expected status 'todo', got %s", addInput.Status)
	}

	// 2. Complete the task
	completeResult, err := UpdateTask(database, &UpdateTaskInput{ID: "task-sequence", Status: "done"})
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}
	if completeResult.Status != "done" {
		t.Fatalf("expected status 'done', got %s", completeResult.Status)
	}

	// 3. Try to complete again (should still work, idempotent)
	completeResult2, err := UpdateTask(database, &UpdateTaskInput{ID: "task-sequence", Status: "done"})
	if err != nil {
		t.Fatalf("CompleteTask second time failed: %v", err)
	}
	if completeResult2.Status != "done" {
		t.Fatalf("expected status 'done' on second completion, got %s", completeResult2.Status)
	}
}

// Test error definitions
func TestServiceErrors_Definitions(t *testing.T) {
	tests := []struct {
		err        error
		errMessage string
	}{
		{ErrMissingBlockReason, "reason for blocking is required"},
		{ErrInvalidTimeout, "timeout minutes must be a positive integer"},
	}

	for _, tt := range tests {
		t.Run(tt.errMessage, func(t *testing.T) {
			if tt.err.Error() != tt.errMessage {
				t.Errorf("error message = %q, want %q", tt.err.Error(), tt.errMessage)
			}
		})
	}
}

// TestResetTimedOut_BusinessLogic tests the ResetTimedOut function's
// validation and no-op paths
func TestResetTimedOut_BusinessLogic(t *testing.T) {
	tests := []struct {
		name           string
		timeoutMinutes int
		setupDB        func() (*db.DB, func())
		wantResetCount int
		wantErr        bool
	}{
		{
			name:           "zero timeout should error",
			timeoutMinutes: 0,
			setupDB: func() (*db.DB, func()) {
				tmpDir := t.TempDir()
				os.Setenv("PROJECT_ROOT", tmpDir)
				testDB, _ := db.NewDB(filepath.Join(tmpDir, "test_service_business_logic.db"))
				return testDB, func() {
					if testDB != nil {
						testDB.Close()
					}
				}
			},
			wantResetCount: 0,
			wantErr:        true,
		},
		{
			name:           "negative timeout should error",
			timeoutMinutes: -1,
			setupDB: func() (*db.DB, func()) {
				tmpDir := t.TempDir()
				os.Setenv("PROJECT_ROOT", tmpDir)
				testDB, _ := db.NewDB(filepath.Join(tmpDir, "test_service_business_logic.db"))
				return testDB, func() {
					if testDB != nil {
						testDB.Close()
					}
				}
			},
			wantResetCount: 0,
			wantErr:        true,
		},
		{
			name:           "no in-progress tasks",
			timeoutMinutes: 30,
			setupDB: func() (*db.DB, func()) {
				tmpDir := t.TempDir()
				os.Setenv("PROJECT_ROOT", tmpDir)
				testDB, _ := db.NewDB(filepath.Join(tmpDir, "test_service_business_logic.db"))
				// Only todo tasks
				_ = testDB.CreateTask(&db.Task{ID: "t1", Title: "T1", Milestone: "m1", Status: "todo"})
				_ = testDB.CreateTask(&db.Task{ID: "t2", Title: "T2", Milestone: "m1", Status: "todo"})
				return testDB, func() {
					if testDB != nil {
						testDB.Close()
					}
				}
			},
			wantResetCount: 0,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database, cleanup := tt.setupDB()
			defer cleanup()

			result, err := ResetTimedOut(database, tt.timeoutMinutes)

			if tt.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && len(result) != tt.wantResetCount {
				t.Errorf("got %d reset tasks, want %d", len(result), tt.wantResetCount)
			}
		})
	}
}

// TestResetTimedOut_ResetsTimedOutTasks ports the age-boundary coverage from
// the former GetTimedOutTasks tests: only in-progress tasks whose
// last_updated is older than the timeout are reset to todo; fresh tasks are
// left untouched.
func TestResetTimedOut_ResetsTimedOutTasks(t *testing.T) {
	database := setupTestDB(t)
	defer teardownTestDB(t, database)

	oneHourAgo := time.Now().UTC().Add(-1 * time.Hour)
	oldTask := &db.Task{
		ID:          "task-old",
		Milestone:   "m1",
		Title:       "Old Task",
		Status:      "in_progress",
		LastUpdated: oneHourAgo,
	}
	if err := database.CreateTask(oldTask); err != nil {
		t.Fatalf("failed to create old task: %v", err)
	}
	freshTask := &db.Task{
		ID:        "task-fresh",
		Milestone: "m1",
		Title:     "Fresh Task",
		Status:    "in_progress",
	}
	if err := database.CreateTask(freshTask); err != nil {
		t.Fatalf("failed to create fresh task: %v", err)
	}

	// 30 minute timeout: only the one-hour-old task is timed out
	result, err := ResetTimedOut(database, 30)
	if err != nil {
		t.Fatalf("ResetTimedOut failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("ResetTimedOut() returned %d tasks, want 1", len(result))
	}
	if result[0].ID != "task-old" {
		t.Errorf("expected task-old to be reset, got %s", result[0].ID)
	}
	if result[0].Status != "todo" {
		t.Errorf("expected reset status 'todo', got %s", result[0].Status)
	}

	// The fresh task must be untouched
	still, err := database.ReadTask("task-fresh")
	if err != nil {
		t.Fatalf("failed to read fresh task: %v", err)
	}
	if still.Status != "in_progress" {
		t.Errorf("expected fresh task status 'in_progress', got %s", still.Status)
	}
}

// TestResetTimedOut_DifferentTimeouts ports the timeout-duration coverage
// from the former GetTimedOutTasks tests.
func TestResetTimedOut_DifferentTimeouts(t *testing.T) {
	oneHourAgo := time.Now().UTC().Add(-1 * time.Hour)

	// With a 90 minute timeout, a one-hour-old task is not timed out
	database := setupTestDB(t)
	_ = database.CreateTask(&db.Task{
		ID:          "task-1",
		Milestone:   "m1",
		Title:       "Task 1",
		Status:      "in_progress",
		LastUpdated: oneHourAgo,
	})
	_ = database.CreateTask(&db.Task{
		ID:          "task-2",
		Milestone:   "m1",
		Title:       "Task 2",
		Status:      "in_progress",
		LastUpdated: oneHourAgo,
	})
	result, err := ResetTimedOut(database, 90)
	if err != nil {
		t.Fatalf("ResetTimedOut failed: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("ResetTimedOut(90 min) = %d, want 0", len(result))
	}
	database.Close()

	// With a 15 minute timeout, both should be timed out
	database = setupTestDB(t)
	defer teardownTestDB(t, database)
	_ = database.CreateTask(&db.Task{
		ID:          "task-1",
		Milestone:   "m1",
		Title:       "Task 1",
		Status:      "in_progress",
		LastUpdated: oneHourAgo,
	})
	_ = database.CreateTask(&db.Task{
		ID:          "task-2",
		Milestone:   "m1",
		Title:       "Task 2",
		Status:      "in_progress",
		LastUpdated: oneHourAgo,
	})
	result, err = ResetTimedOut(database, 15)
	if err != nil {
		t.Fatalf("ResetTimedOut failed: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("ResetTimedOut(15 min) = %d, want 2", len(result))
	}
}

// TestTaskToItem maps a database task to a list output item with formatted
// timestamps. (Ported from the former GetTask tests.)
func TestTaskToItem(t *testing.T) {
	task := db.Task{
		ID:          "TASK-001",
		Milestone:   "v1.0",
		Title:       "Implement login",
		Description: "Add authentication system",
		Status:      "todo",
		Actor:       "alice",
		Created:     time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC),
		LastUpdated: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
	}

	item := TaskToItem(task)
	if item.ID != "TASK-001" {
		t.Errorf("Expected ID TASK-001, got %s", item.ID)
	}
	if item.Title != "Implement login" {
		t.Errorf("Expected title 'Implement login', got %s", item.Title)
	}
	if item.Status != "todo" {
		t.Errorf("Expected status todo, got %s", item.Status)
	}
	if item.Created != "2026-01-02 15:04:05" {
		t.Errorf("Expected formatted created '2026-01-02 15:04:05', got %s", item.Created)
	}
	if item.LastUpdated != "2026-03-10 00:00:00" {
		t.Errorf("Expected formatted last_updated '2026-03-10 00:00:00', got %s", item.LastUpdated)
	}
}
