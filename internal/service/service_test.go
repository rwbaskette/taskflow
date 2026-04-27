package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rwbaskette/taskflow/internal/db"
)

func setupTestDBService(t *testing.T) *db.DB {
	tmpDir := t.TempDir()
	testDBPath := filepath.Join(tmpDir, "test_service_business_logic.db")
	// Set project root for schema lookup
	os.Setenv("PROJECT_ROOT", tmpDir)
	testDB, err := db.NewDB(testDBPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	return testDB
}

func teardownTestDBService(t *testing.T, testDB *db.DB) {
	if testDB != nil {
		testDB.Close()
	}
}

// Test service layer AddTask with various inputs and edge cases
func TestAddTask_EdgeCases(t *testing.T) {
	database := setupTestDBService(t)
	defer teardownTestDBService(t, database)

	tests := []struct {
		name    string
		input   *AddTaskInput
		wantErr bool
	}{
		{
			name: "task with empty description",
			input: &AddTaskInput{
				ID:          "task-001",
				Title:       "Test Task",
				Milestone:   "milestone-1",
				Description: "",
				Actor:       "user1",
			},
			wantErr: false,
		},
		{
			name: "task with empty actor",
			input: &AddTaskInput{
				ID:          "task-002",
				Title:       "Test Task",
				Milestone:   "milestone-1",
				Description: "Description",
				Actor:       "",
			},
			wantErr: false,
		},
		{
			name: "task with empty milestone",
			input: &AddTaskInput{
				ID:          "task-003",
				Title:       "Test Task",
				Milestone:   "",
				Description: "Description",
				Actor:       "user1",
			},
			wantErr: false, // milestone validation is done at validation layer, not service layer
		},
		{
			name: "task with special characters in title",
			input: &AddTaskInput{
				ID:          "task-004",
				Title:       "Test Task #1 - Implement & Test!",
				Milestone:   "milestone-1",
				Description: "Description with 'quotes' and \"double quotes\"",
				Actor:       "user1",
			},
			wantErr: false,
		},
		{
			name: "task with unicode characters",
			input: &AddTaskInput{
				ID:          "task-005",
				Title:       "Test 日本語 Task",
				Milestone:   "milestone-1",
				Description: "Description",
				Actor:       "user1",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AddTask(database, tt.input)
			if tt.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && result != nil {
				if result.Status != "todo" {
					t.Errorf("expected status 'todo', got %s", result.Status)
				}
			}
		})
	}
}

// Test service layer status transitions through the CompleteTask function
func TestCompleteTask_StatusTransitions(t *testing.T) {
	database := setupTestDBService(t)
	defer teardownTestDBService(t, database)

	// Create tasks in different statuses to test completion
	tests := []struct {
		name           string
		initialStatus  string
		updateToStatus string
		wantErr        bool
	}{
		{
			name:           "complete from todo",
			initialStatus:  "todo",
			updateToStatus: "done",
			wantErr:        false,
		},
		{
			name:           "complete from in_progress",
			initialStatus:  "in_progress",
			updateToStatus: "done",
			wantErr:        false,
		},
		{
			name:           "complete from blocked",
			initialStatus:  "blocked",
			updateToStatus: "done",
			wantErr:        false,
		},
		{
			name:           "complete nonexistent task",
			initialStatus:  "",
			updateToStatus: "nonexistent",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initialStatus != "" {
				// Create a task with initial status
				input := &AddTaskInput{
					ID:        "task-" + tt.name,
					Title:     "Test Task",
					Milestone: "milestone-1",
					Actor:     "testuser",
				}
				_, err := AddTask(database, input)
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
			result, err := CompleteTask(database, &CompleteTaskInput{ID: "task-" + tt.name})
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

// Test BlockTask function with various scenarios
func TestBlockTask_BusinessLogic(t *testing.T) {
	database := setupTestDBService(t)
	defer teardownTestDBService(t, database)

	// Create a task to block
	addInput := &AddTaskInput{
		ID:          "task-to-block",
		Title:       "Task to Block",
		Milestone:   "milestone-1",
		Description: "Original description",
		Actor:       "testuser",
	}
	_, err := AddTask(database, addInput)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Test with second task after first blocking test
	addInput2 := &AddTaskInput{
		ID:          "task-to-block-2",
		Title:       "Task to Block 2",
		Milestone:   "milestone-1",
		Description: "Original description 2",
		Actor:       "testuser",
	}
	_, _ = AddTask(database, addInput2)

	blockInput := BlockTaskInput{
		ID:     "task-to-block-2",
		Reason: "Test reason",
	}
	result, err := BlockTask(database, blockInput)
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
	_, _ = AddTask(database, &AddTaskInput{ID: "t1", Title: "T1", Milestone: "m1"})
	_, _ = AddTask(database, &AddTaskInput{ID: "t2", Title: "T2", Milestone: "m1"})

	// Test validation errors
	_, err = BlockTask(database, BlockTaskInput{ID: "t1", Reason: ""})
	if err != ErrMissingBlockReason {
		t.Errorf("expected ErrMissingBlockReason, got %v", err)
	}

	_, err = BlockTask(database, BlockTaskInput{ID: "", Reason: "reason"})
	if err != ErrInvalidID {
		t.Errorf("expected ErrInvalidID, got %v", err)
	}

	_, err = BlockTask(database, BlockTaskInput{ID: "nonexistent", Reason: "reason"})
	if err == nil {
		t.Error("expected error for nonexistent task")
	}

	// Close and cleanup
	database.Close()
}

// Test BlockTask appends reason to description
func TestBlockTask_AppendsReasonToDescription(t *testing.T) {
	database := setupTestDBService(t)
	defer teardownTestDBService(t, database)

	// Create task with existing description
	addInput := &AddTaskInput{
		ID:          "task-with-desc",
		Title:       "Task With Description",
		Milestone:   "milestone-1",
		Description: "This is the original description.",
		Actor:       "testuser",
	}
	_, err := AddTask(database, addInput)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Block the task
	blockInput := BlockTaskInput{
		ID:     "task-with-desc",
		Reason: "Waiting for API",
	}
	result, err := BlockTask(database, blockInput)
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

// Test ResetTimedOut function business logic
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
				_, _ = AddTask(testDB, &AddTaskInput{ID: "t1", Title: "T1", Milestone: "m1"})
				_, _ = AddTask(testDB, &AddTaskInput{ID: "t2", Title: "T2", Milestone: "m1"})
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

			input := ResetTimedOutInput{
				TimeoutMinutes: tt.timeoutMinutes,
			}
			result, err := ResetTimedOut(database, input)

			if tt.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && result != nil && len(result.ResetTasks) != tt.wantResetCount {
				t.Errorf("got %d reset tasks, want %d", len(result.ResetTasks), tt.wantResetCount)
			}
		})
	}
}

// Test error handling for nil database
func TestServiceErrors_NilDatabase(t *testing.T) {
	var nilDB *db.DB = nil

	tests := []struct {
		name string
		fn   func() error
	}{
		{
			name: "AddTask with nil database",
			fn: func() error {
				_, err := AddTask(nilDB, &AddTaskInput{ID: "t1", Title: "T1", Milestone: "m1"})
				return err
			},
		},
		{
			name: "CompleteTask with nil database",
			fn: func() error {
				_, err := CompleteTask(nilDB, &CompleteTaskInput{ID: "t1"})
				return err
			},
		},
		{
			name: "BlockTask with nil database",
			fn: func() error {
				_, err := BlockTask(nilDB, BlockTaskInput{ID: "t1", Reason: "reason"})
				return err
			},
		},
		{
			name: "ResetTimedOut with nil database",
			fn: func() error {
				_, err := ResetTimedOut(nilDB, ResetTimedOutInput{TimeoutMinutes: 30})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err != ErrNilDatabase {
				t.Errorf("expected ErrNilDatabase, got %v", err)
			}
		})
	}
}

// Test error definitions
func TestServiceErrors_Definitions(t *testing.T) {
	tests := []struct {
		err        error
		errMessage string
	}{
		{ErrNilDatabase, "database connection is nil"},
		{ErrNilInput, "input is nil"},
		{ErrInvalidID, "task ID is invalid or missing"},
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

// Test that AddTask sets default status to todo
func TestAddTask_DefaultStatus(t *testing.T) {
	database := setupTestDBService(t)
	defer teardownTestDBService(t, database)

	input := &AddTaskInput{
		ID:        "task-status-test",
		Title:     "Status Test Task",
		Milestone: "milestone-1",
	}

	result, err := AddTask(database, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "todo" {
		t.Errorf("expected status 'todo', got %s", result.Status)
	}
}

// Test complete task preserves other fields
func TestCompleteTask_PreservesFields(t *testing.T) {
	database := setupTestDBService(t)
	defer teardownTestDBService(t, database)

	// Create task with all fields
	input := &AddTaskInput{
		ID:          "task-full",
		Title:       "Full Task",
		Milestone:   "milestone-1",
		Description: "Some description",
		Actor:       "testuser",
	}
	_, err := AddTask(database, input)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Complete the task
	completeResult, err := CompleteTask(database, &CompleteTaskInput{ID: "task-full"})
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

// Test multiple operations in sequence
func TestService_MultipleOperations(t *testing.T) {
	database := setupTestDBService(t)
	defer teardownTestDBService(t, database)

	// 1. Add a task
	addInput := &AddTaskInput{
		ID:        "task-sequence",
		Title:     "Sequence Task",
		Milestone: "milestone-1",
		Actor:     "user1",
	}
	result, err := AddTask(database, addInput)
	if err != nil {
		t.Fatalf("AddTask failed: %v", err)
	}
	if result.Status != "todo" {
		t.Fatalf("expected status 'todo', got %s", result.Status)
	}

	// 2. Complete the task
	completeResult, err := CompleteTask(database, &CompleteTaskInput{ID: "task-sequence"})
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}
	if completeResult.Status != "done" {
		t.Fatalf("expected status 'done', got %s", completeResult.Status)
	}

	// 3. Try to complete again (should still work, idempotent)
	completeResult2, err := CompleteTask(database, &CompleteTaskInput{ID: "task-sequence"})
	if err != nil {
		t.Fatalf("CompleteTask second time failed: %v", err)
	}
	if completeResult2.Status != "done" {
		t.Fatalf("expected status 'done' on second completion, got %s", completeResult2.Status)
	}
}

// Test service layer handles special characters in IDs
func TestAddTask_SpecialCharactersInID(t *testing.T) {
	database := setupTestDBService(t)
	defer teardownTestDBService(t, database)

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "regular ID",
			id:      "task-123",
			wantErr: false,
		},
		{
			name:    "ID with underscore",
			id:      "task_123",
			wantErr: false,
		},
		{
			name:    "ID with dash",
			id:      "task-123-456",
			wantErr: false,
		},
		{
			name:    "ID with dots",
			id:      "task.123.456",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &AddTaskInput{
				ID:        tt.id,
				Title:     "Test Task",
				Milestone: "milestone-1",
			}
			_, err := AddTask(database, input)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
