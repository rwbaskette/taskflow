package service

import (
	"os"
	"testing"

	"github.com/user/project/internal/db"
)

const testDBPath = "/home/rwbaskette/tmp/test_service_task.db"

func setupTestDB(t *testing.T) *db.DB {
	if err := os.RemoveAll(testDBPath); err != nil {
		t.Fatalf("failed to remove test db: %v", err)
	}
	// Set project root for schema lookup
	os.Setenv("PROJECT_ROOT", "/home/rwbaskette/tmp")
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
	os.RemoveAll(testDBPath)
}

func TestAddTask_ValidTaskCreation(t *testing.T) {
	// Setup: Create a test database
	database := setupTestDB(t)
	defer teardownTestDB(t, database)

	input := &AddTaskInput{
		ID:          "task-001",
		Milestone:   "milestone-1",
		Title:       "Test Task",
		Description: "This is a test task",
		Actor:       "testuser",
	}

	// Execute: Call AddTask with valid input
	result, err := AddTask(database, input)

	// Verify: Should succeed and return result
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ID != input.ID {
		t.Errorf("expected ID %s, got %s", input.ID, result.ID)
	}
	if result.Title != input.Title {
		t.Errorf("expected Title %s, got %s", input.Title, result.Title)
	}
	if result.Status != "todo" {
		t.Errorf("expected status 'todo', got %s", result.Status)
	}
	if result.Actor != input.Actor {
		t.Errorf("expected Actor %s, got %s", input.Actor, result.Actor)
	}
}

func TestAddTask_NilDatabase(t *testing.T) {
	// Setup: Nil database
	var database *db.DB = nil

	input := &AddTaskInput{
		ID:          "task-001",
		Milestone:   "milestone-1",
		Title:       "Test Task",
		Description: "Test description",
		Actor:       "testuser",
	}

	// Execute: Call AddTask with nil database
	result, err := AddTask(database, input)

	// Verify: Should return error for nil database
	if err != ErrNilDatabase {
		t.Errorf("expected ErrNilDatabase, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestAddTask_NilInput(t *testing.T) {
	// Setup: Create a test database
	database := setupTestDB(t)
	defer teardownTestDB(t, database)

	// Execute: Call AddTask with nil input
	result, err := AddTask(database, nil)

	// Verify: Should return error for nil input
	if err != ErrNilInput {
		t.Errorf("expected ErrNilInput, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestAddTask_DuplicateID(t *testing.T) {
	// Setup: Create a test database
	database := setupTestDB(t)
	defer teardownTestDB(t, database)

	// First, create a task with a specific ID
	input1 := &AddTaskInput{
		ID:          "duplicate-task-id",
		Milestone:   "milestone-1",
		Title:       "First Task",
		Description: "First task description",
		Actor:       "testuser",
	}

	result1, err := AddTask(database, input1)
	if err != nil {
		t.Fatalf("failed to create first task: %v", err)
	}
	if result1 == nil {
		t.Fatal("expected non-nil result for first task creation")
	}
	if result1.ID != "duplicate-task-id" {
		t.Errorf("expected ID 'duplicate-task-id', got %s", result1.ID)
	}

	// Execute: Try to create a task with the same ID (duplicate)
	input2 := &AddTaskInput{
		ID:          "duplicate-task-id", // Same ID as above
		Milestone:   "milestone-1",
		Title:       "Second Task",
		Description: "Second task description",
		Actor:       "testuser2",
	}

	result2, err := AddTask(database, input2)

	// Verify: Should return error for duplicate ID
	if err == nil {
		t.Error("expected error for duplicate ID, got nil")
	}

	// Check if it's the expected TaskAlreadyExists error
	if !db.IsTaskAlreadyExists(err) {
		t.Errorf("expected TaskAlreadyExists error, got %v", err)
	}
	if result2 != nil {
		t.Errorf("expected nil result for duplicate, got %v", result2)
	}
}
