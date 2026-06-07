package service

import (
	"testing"
	"time"

	"github.com/rwbaskette/taskflow/internal/db"
)

func TestWaitTask_AlreadyCompleted(t *testing.T) {
	// Setup: Create a test database with a completed task
	database := setupTestDB(t)
	defer teardownTestDB(t, database)

	// Create a task in "done" status
	task := &db.Task{
		ID:          "completed-task",
		Milestone:   "milestone-1",
		Title:       "Completed Task",
		Description: "Task that is already done",
		Status:      "done",
		Actor:       "testuser",
		LastUpdated: time.Now().UTC(),
	}
	if err := database.CreateTask(task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Execute: Wait on the already completed task
	input := &WaitTaskInput{
		TaskIDs: []string{"completed-task"},
		Timeout: 0, // No timeout
	}

	results, err := WaitTask(database, input)

	// Verify: Should return immediately with completed status
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].State != "completed" {
		t.Errorf("expected state 'completed', got %s", results[0].State)
	}
	if results[0].TaskID != "completed-task" {
		t.Errorf("expected task ID 'completed-task', got %s", results[0].TaskID)
	}
}

func TestWaitTask_NilDatabase(t *testing.T) {
	// Setup: Nil database
	var database *db.DB = nil

	input := &WaitTaskInput{
		TaskIDs: []string{"task-1"},
		Timeout: 0,
	}

	// Execute: Call WaitTask with nil database
	result, err := WaitTask(database, input)

	// Verify: Should return error for nil database
	if err != ErrNilDatabase {
		t.Errorf("expected ErrNilDatabase, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestWaitTask_NilInput(t *testing.T) {
	// Setup: Create a test database
	database := setupTestDB(t)
	defer teardownTestDB(t, database)

	// Execute: Call WaitTask with nil input
	result, err := WaitTask(database, nil)

	// Verify: Should return error for nil input
	if err != ErrNilInput {
		t.Errorf("expected ErrNilInput, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestWaitTask_EmptyTaskIDs(t *testing.T) {
	// Setup: Create a test database
	database := setupTestDB(t)
	defer teardownTestDB(t, database)

	input := &WaitTaskInput{
		TaskIDs: []string{}, // Empty task IDs
		Timeout: 0,
	}

	// Execute: Call WaitTask with empty task IDs
	result, err := WaitTask(database, input)

	// Verify: Should return error for empty task IDs
	if err != ErrInvalidID {
		t.Errorf("expected ErrInvalidID, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestWaitTask_TaskNotFound(t *testing.T) {
	// Setup: Create a test database (empty)
	database := setupTestDB(t)
	defer teardownTestDB(t, database)

	input := &WaitTaskInput{
		TaskIDs: []string{"nonexistent-task"},
		Timeout: 1000,
	}

	// Execute: Wait on a non-existent task
	result, err := WaitTask(database, input)

	// Verify: Should return error for non-existent task
	if err == nil {
		t.Error("expected error for non-existent task, got nil")
	}
	// The error should contain "task not found"
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestWaitTask_MultipleTasks(t *testing.T) {
	// Setup: Create a test database with multiple tasks
	database := setupTestDB(t)
	defer teardownTestDB(t, database)

	// Create a completed task
	task1 := &db.Task{
		ID:          "task-1",
		Milestone:   "milestone-1",
		Title:       "Task 1",
		Description: "First task",
		Status:      "done",
		Actor:       "testuser",
		LastUpdated: time.Now().UTC(),
	}
	if err := database.CreateTask(task1); err != nil {
		t.Fatalf("failed to create task1: %v", err)
	}

	// Create an in-progress task
	task2 := &db.Task{
		ID:          "task-2",
		Milestone:   "milestone-1",
		Title:       "Task 2",
		Description: "Second task",
		Status:      "in_progress",
		Actor:       "testuser",
		LastUpdated: time.Now().UTC(),
	}
	if err := database.CreateTask(task2); err != nil {
		t.Fatalf("failed to create task2: %v", err)
	}

	// Execute: Wait on both tasks with a short timeout
	input := &WaitTaskInput{
		TaskIDs: []string{"task-1", "task-2"},
		Timeout: 500, // Short timeout - task-2 won't complete
	}

	results, err := WaitTask(database, input)

	// Verify: Should return results for both tasks
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Find each result
	var task1Result, task2Result *WaitTaskResult
	for i := range results {
		if results[i].TaskID == "task-1" {
			task1Result = &results[i]
		} else if results[i].TaskID == "task-2" {
			task2Result = &results[i]
		}
	}

	// Task 1 should be completed (it was already done)
	if task1Result == nil || task1Result.State != "completed" {
		t.Errorf("expected task-1 to be completed, got %v", task1Result)
	}

	// Task 2 should be timed out (it was in_progress and timeout hit)
	if task2Result == nil || task2Result.State != "timed_out" {
		t.Errorf("expected task-2 to be timed_out, got %v", task2Result)
	}
}

func TestWaitTask_TimeoutReached(t *testing.T) {
	// Setup: Create a test database with an in-progress task
	database := setupTestDB(t)
	defer teardownTestDB(t, database)

	// Create an in-progress task
	task := &db.Task{
		ID:          "pending-task",
		Milestone:   "milestone-1",
		Title:       "Pending Task",
		Description: "Task that will not complete",
		Status:      "in_progress",
		Actor:       "testuser",
		LastUpdated: time.Now().UTC(),
	}
	if err := database.CreateTask(task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Execute: Wait on the task with a very short timeout
	input := &WaitTaskInput{
		TaskIDs: []string{"pending-task"},
		Timeout: 100, // 100ms timeout - should trigger immediately since task won't complete
	}

	// Use a channel to detect when the function returns
	done := make(chan struct{})
	var results []WaitTaskResult
	var waitErr error

	go func() {
		results, waitErr = WaitTask(database, input)
		close(done)
	}()

	// Wait for up to 5 seconds for the function to complete
	select {
	case <-done:
		// Verify: Should return with timed_out status
		if waitErr != nil {
			t.Errorf("unexpected error: %v", waitErr)
		}
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}
		if results[0].State != "timed_out" {
			t.Errorf("expected state 'timed_out', got %s", results[0].State)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("WaitTask did not return within timeout")
	}
}