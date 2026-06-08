package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rwbaskette/taskflow/internal/db"
)

func setupTestDBUnblock(t *testing.T) *db.DB {
	tmpDir := t.TempDir()
	testDBPath := filepath.Join(tmpDir, "test_unblock_task.db")
	// Set project root for schema lookup
	os.Setenv("PROJECT_ROOT", tmpDir)
	testDB, err := db.NewDB(testDBPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	return testDB
}

func teardownTestDBUnblock(t *testing.T, testDB *db.DB) {
	if testDB != nil {
		testDB.Close()
	}
}

// TestUnblockTask_RefreshesLastUpdated verifies that the last_updated timestamp
// is refreshed when unblocking a task. It captures the timestamp before calling
// unblock, then verifies the new timestamp is strictly greater than the original
// (with a small tolerance for clock drift).
func TestUnblockTask_RefreshesLastUpdated(t *testing.T) {
	database := setupTestDBUnblock(t)
	defer teardownTestDBUnblock(t, database)

	// Step 1: Create a task
	addInput := &AddTaskInput{
		ID:          "task-unblock-ts",
		Milestone:   "milestone-1",
		Title:       "Unblock Timestamp Test",
		Description: "This task tests timestamp refresh on unblock",
		Actor:       "testuser",
	}
	_, err := AddTask(database, addInput)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Step 2: Block the task so it is in "blocked" status
	blockInput := BlockTaskInput{
		ID:     "task-unblock-ts",
		Reason: "Waiting on dependency",
	}
	_, err = BlockTask(database, blockInput)
	if err != nil {
		t.Fatalf("failed to block task: %v", err)
	}

	// Step 3: Read the task and capture the last_updated timestamp before unblocking
	preUnblockTask, err := database.ReadTask("task-unblock-ts")
	if err != nil {
		t.Fatalf("failed to read task before unblock: %v", err)
	}
	originalLastUpdated := preUnblockTask.LastUpdated

	// Step 4: Wait briefly to ensure a measurable time difference
	time.Sleep(100 * time.Millisecond)

	// Step 5: Call the unblock handler with a valid task ID
	result, err := UnblockTask(database, UnblockTaskInput{ID: "task-unblock-ts"})
	if err != nil {
		t.Fatalf("unexpected error from UnblockTask: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result from UnblockTask")
	}

	// Step 6: Read the updated task from the database to get the persisted timestamp
	postUnblockTask, err := database.ReadTask("task-unblock-ts")
	if err != nil {
		t.Fatalf("failed to read task after unblock: %v", err)
	}
	newLastUpdated := postUnblockTask.LastUpdated

	// Step 7: Verify the task status changed to "todo"
	if result.Status != "todo" {
		t.Errorf("expected status 'todo', got %s", result.Status)
	}

	// Step 8: Verify the blocked_by field was cleared
	if result.BlockedBy != nil && len(result.BlockedBy) > 0 {
		t.Errorf("expected blocked_by to be cleared, got %v", result.BlockedBy)
	}

	// Step 9: Verify new timestamp is strictly greater than original
	// Allow a small tolerance (1 second) for clock drift
	tolerance := 1 * time.Second
	if newLastUpdated.Add(tolerance).Before(originalLastUpdated) {
		t.Errorf("expected last_updated to be strictly greater than original; "+
			"original=%v, new=%v, tolerance=%v",
			originalLastUpdated, newLastUpdated, tolerance)
	}

	// Step 10: Verify the returned LastUpdated also reflects the update
	if result.LastUpdated.IsZero() {
		t.Error("expected LastUpdated in result to be non-zero")
	}
	if result.LastUpdated.Add(tolerance).Before(originalLastUpdated) {
		t.Errorf("expected result.LastUpdated to be strictly greater than original; "+
			"original=%v, result.LastUpdated=%v",
			originalLastUpdated, result.LastUpdated)
	}

	t.Logf("original last_updated: %v", originalLastUpdated)
	t.Logf("new last_updated:      %v", newLastUpdated)
	t.Logf("delta:                 %v", newLastUpdated.Sub(originalLastUpdated))
}

// TestUnblockTask_StatusTransition verifies unblocking transitions from blocked to todo
func TestUnblockTask_StatusTransition(t *testing.T) {
	database := setupTestDBUnblock(t)
	defer teardownTestDBUnblock(t, database)

	// Create and block a task
	_, err := AddTask(database, &AddTaskInput{
		ID:        "task-status",
		Title:     "Status Transition Test",
		Milestone: "milestone-1",
		Actor:     "testuser",
	})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	_, err = BlockTask(database, BlockTaskInput{
		ID:     "task-status",
		Reason: "Test reason",
	})
	if err != nil {
		t.Fatalf("failed to block task: %v", err)
	}

	// Unblock
	result, err := UnblockTask(database, UnblockTaskInput{ID: "task-status"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "todo" {
		t.Errorf("expected status 'todo', got %s", result.Status)
	}
}

// TestUnblockTask_OnlyBlockedTasksCanBeUnblocked verifies unblocking a non-blocked task fails
func TestUnblockTask_OnlyBlockedTasksCanBeUnblocked(t *testing.T) {
	database := setupTestDBUnblock(t)
	defer teardownTestDBUnblock(t, database)

	// Create a task (status is "todo", not "blocked")
	_, err := AddTask(database, &AddTaskInput{
		ID:        "task-not-blocked",
		Title:     "Not Blocked",
		Milestone: "milestone-1",
		Actor:     "testuser",
	})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Attempt to unblock a non-blocked task
	_, err = UnblockTask(database, UnblockTaskInput{ID: "task-not-blocked"})
	if err == nil {
		t.Fatal("expected error when unblocking a non-blocked task, got nil")
	}
}

// TestUnblockTask_ClearsBlockedByField verifies that the blocked_by field is
// cleared (set to nil/empty) when calling UnblockTask, both in the returned
// result and in the persisted database record.
func TestUnblockTask_ClearsBlockedByField(t *testing.T) {
	database := setupTestDBUnblock(t)
	defer teardownTestDBUnblock(t, database)

	// Step 1: Create a task
	_, err := AddTask(database, &AddTaskInput{
		ID:          "task-clear-blockedby",
		Title:       "Clear BlockedBy Field Test",
		Milestone:   "milestone-1",
		Description: "This task tests blocked_by clearing on unblock",
		Actor:       "testuser",
	})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Step 2: Block the task so the blocked_by field is set to a non-nil value
	blockInput := BlockTaskInput{
		ID:     "task-clear-blockedby",
		Reason: "Waiting on external dependency",
	}
	blockResult, err := BlockTask(database, blockInput)
	if err != nil {
		t.Fatalf("failed to block task: %v", err)
	}

	// Verify the task is now blocked and blocked_by is non-empty
	if blockResult.Status != "blocked" {
		t.Fatalf("expected status 'blocked' after blocking, got %s", blockResult.Status)
	}
	if blockResult.BlockedBy == nil || len(blockResult.BlockedBy) == 0 {
		t.Fatal("expected blocked_by to be set after blocking, got nil or empty")
	}
	t.Logf("blocked_by before unblock: %v", blockResult.BlockedBy)

	// Step 3: Call UnblockTask with the task ID
	result, err := UnblockTask(database, UnblockTaskInput{ID: "task-clear-blockedby"})
	if err != nil {
		t.Fatalf("unexpected error from UnblockTask: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result from UnblockTask")
	}

	// Step 4: Assert that the returned result's BlockedBy field is nil or empty
	if result.BlockedBy != nil && len(result.BlockedBy) > 0 {
		t.Errorf("expected result.BlockedBy to be cleared, got %v", result.BlockedBy)
	}

	// Step 5: Read the task from the database and verify the persisted blocked_by is also nil or empty
	persistedTask, err := database.ReadTask("task-clear-blockedby")
	if err != nil {
		t.Fatalf("failed to read task from database: %v", err)
	}
	if persistedTask.BlockedBy != nil && len(persistedTask.BlockedBy) > 0 {
		t.Errorf("expected persisted blocked_by to be nil or empty, got %v", persistedTask.BlockedBy)
	}

	t.Logf("blocked_by after unblock (result):    %v", result.BlockedBy)
	t.Logf("blocked_by after unblock (persisted): %v", persistedTask.BlockedBy)
}

// TestUnblockTask_OverwritesDescription verifies that providing a description
// parameter to UnblockTask overwrites the task's description with the new value.
func TestUnblockTask_OverwritesDescription(t *testing.T) {
	database := setupTestDBUnblock(t)
	defer teardownTestDBUnblock(t, database)

	const originalDescription = "This is the original description before unblock"
	const newDescription = "This is the new description after unblock"

	// Step 1: Create a task with an original description
	_, err := AddTask(database, &AddTaskInput{
		ID:          "task-desc-overwrite",
		Title:       "Description Overwrite Test",
		Milestone:   "milestone-1",
		Description: originalDescription,
		Actor:       "testuser",
	})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Step 2: Block the task so it is in "blocked" status.
	// Note: BlockTask appends "[BLOCKED: reason]" to the description.
	_, err = BlockTask(database, BlockTaskInput{
		ID:     "task-desc-overwrite",
		Reason: "Waiting on dependency",
	})
	if err != nil {
		t.Fatalf("failed to block task: %v", err)
	}

	// Step 3: Verify the task is blocked and description has the block suffix
	preUnblockTask, err := database.ReadTask("task-desc-overwrite")
	if err != nil {
		t.Fatalf("failed to read task before unblock: %v", err)
	}
	expectedBlockedDesc := originalDescription + "\n[BLOCKED: Waiting on dependency]"
	if preUnblockTask.Description != expectedBlockedDesc {
		t.Fatalf("expected blocked description %q, got %q", expectedBlockedDesc, preUnblockTask.Description)
	}

	// Step 4: Call UnblockTask with a new description
	result, err := UnblockTask(database, UnblockTaskInput{
		ID:          "task-desc-overwrite",
		Description: newDescription,
	})
	if err != nil {
		t.Fatalf("unexpected error from UnblockTask: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result from UnblockTask")
	}

	// Step 5: Verify the returned result has the new description
	if result.Description != newDescription {
		t.Errorf("expected result description %q, got %q", newDescription, result.Description)
	}

	// Step 6: Verify the description was persisted in the database
	postUnblockTask, err := database.ReadTask("task-desc-overwrite")
	if err != nil {
		t.Fatalf("failed to read task after unblock: %v", err)
	}
	if postUnblockTask.Description != newDescription {
		t.Errorf("expected persisted description %q, got %q", newDescription, postUnblockTask.Description)
	}

	// Step 7: Verify the original description (with block suffix) is no longer present
	if postUnblockTask.Description == expectedBlockedDesc {
		t.Error("expected original blocked description to be overwritten, but it was preserved")
	}

	t.Logf("original description:      %q", originalDescription)
	t.Logf("blocked description:       %q", expectedBlockedDesc)
	t.Logf("new description (result):  %q", result.Description)
	t.Logf("new description (persist): %q", postUnblockTask.Description)
}

// TestUnblockTask_PreservesDescription verifies that when no description
// parameter is provided, the original description (including any block suffix)
// is preserved unchanged.
func TestUnblockTask_PreservesDescription(t *testing.T) {
	database := setupTestDBUnblock(t)
	defer teardownTestDBUnblock(t, database)

	const originalDescription = "This is the original description that must be preserved"

	// Step 1: Create and block a task with a description
	_, err := AddTask(database, &AddTaskInput{
		ID:          "task-desc",
		Title:       "Description Preservation Test",
		Milestone:   "milestone-1",
		Description: originalDescription,
		Actor:       "testuser",
	})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Note: BlockTask appends "[BLOCKED: reason]" to the description.
	_, err = BlockTask(database, BlockTaskInput{
		ID:     "task-desc",
		Reason: "Test reason",
	})
	if err != nil {
		t.Fatalf("failed to block task: %v", err)
	}

	// Step 2: Verify the blocked description is stored in the database
	preUnblockTask, err := database.ReadTask("task-desc")
	if err != nil {
		t.Fatalf("failed to read task before unblock: %v", err)
	}
	expectedBlockedDesc := originalDescription + "\n[BLOCKED: Test reason]"
	if preUnblockTask.Description != expectedBlockedDesc {
		t.Fatalf("expected blocked description %q, got %q", expectedBlockedDesc, preUnblockTask.Description)
	}

	// Step 3: Unblock without providing a description parameter
	result, err := UnblockTask(database, UnblockTaskInput{ID: "task-desc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Step 4: Verify the returned result has the full blocked description preserved
	if result.Description != expectedBlockedDesc {
		t.Errorf("expected preserved description %q, got %q", expectedBlockedDesc, result.Description)
	}

	// Step 5: Verify the original description (with block suffix) was persisted unchanged
	postUnblockTask, err := database.ReadTask("task-desc")
	if err != nil {
		t.Fatalf("failed to read task after unblock: %v", err)
	}
	if postUnblockTask.Description != expectedBlockedDesc {
		t.Errorf("expected persisted description to remain %q, got %q", expectedBlockedDesc, postUnblockTask.Description)
	}

	t.Logf("original description:  %q", originalDescription)
	t.Logf("blocked description:   %q", expectedBlockedDesc)
	t.Logf("preserved description: %q", postUnblockTask.Description)
}

// TestUnblockTask_StatusUpdatedFromBlockedToTodo verifies that the status field
// is correctly updated from 'blocked' to 'todo' after calling UnblockTask.
// It asserts the change both in the returned result and in the persisted
// database record.
func TestUnblockTask_StatusUpdatedFromBlockedToTodo(t *testing.T) {
	database := setupTestDBUnblock(t)
	defer teardownTestDBUnblock(t, database)

	// Step 1: Create a task in 'todo' status
	_, err := AddTask(database, &AddTaskInput{
		ID:          "task-status-update",
		Title:       "Status Update Test",
		Milestone:   "milestone-1",
		Description: "This task tests the status update on unblock",
		Actor:       "testuser",
	})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Step 2: Block the task so it is in 'blocked' status
	_, err = BlockTask(database, BlockTaskInput{
		ID:     "task-status-update",
		Reason: "Waiting on dependency",
	})
	if err != nil {
		t.Fatalf("failed to block task: %v", err)
	}

	// Step 3: Verify the task is currently in 'blocked' status in the database
	preUnblockTask, err := database.ReadTask("task-status-update")
	if err != nil {
		t.Fatalf("failed to read task before unblock: %v", err)
	}
	if preUnblockTask.Status != "blocked" {
		t.Fatalf("expected task status to be 'blocked' before unblock, got %s", preUnblockTask.Status)
	}
	t.Logf("task status before unblock: %s", preUnblockTask.Status)

	// Step 4: Call the unblock handler directly with a valid task ID
	result, err := UnblockTask(database, UnblockTaskInput{ID: "task-status-update"})
	if err != nil {
		t.Fatalf("unexpected error from UnblockTask: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result from UnblockTask")
	}

	// Step 5: Assert that the returned result has status 'todo'
	if result.Status != "todo" {
		t.Errorf("expected result status 'todo', got %s", result.Status)
	}
	t.Logf("task status in result: %s", result.Status)

	// Step 6: Assert that the persisted database record also has status 'todo'
	postUnblockTask, err := database.ReadTask("task-status-update")
	if err != nil {
		t.Fatalf("failed to read task from database after unblock: %v", err)
	}
	if postUnblockTask.Status != "todo" {
		t.Errorf("expected persisted task status 'todo', got %s", postUnblockTask.Status)
	}
	t.Logf("task status persisted in database: %s", postUnblockTask.Status)
}
