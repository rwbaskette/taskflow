package service

import (
	"os"
	"testing"

	"github.com/user/project/internal/db"
)

const testDBPathComplete = "/home/rwbaskette/tmp/test_complete_task.db"

func setupTestDBComplete(t *testing.T) *db.DB {
	if err := os.RemoveAll(testDBPathComplete); err != nil {
		t.Fatalf("failed to remove test db: %v", err)
	}
	os.Setenv("PROJECT_ROOT", "/home/rwbaskette/tmp")
	testDB, err := db.NewDB(testDBPathComplete)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	return testDB
}

func teardownTestDBComplete(t *testing.T, testDB *db.DB) {
	if testDB != nil {
		testDB.Close()
	}
	os.RemoveAll(testDBPathComplete)
}

func TestCompleteTask_ValidTaskCompletion(t *testing.T) {
	database := setupTestDBComplete(t)
	defer teardownTestDBComplete(t, database)

	addInput := &AddTaskInput{
		ID:          "task-001",
		Milestone:   "milestone-1",
		Title:       "Test Task",
		Description: "This is a test task",
		Actor:       "testuser",
	}

	_, err := AddTask(database, addInput)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	result, err := CompleteTask(database, "task-001")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ID != "task-001" {
		t.Errorf("expected ID task-001, got %s", result.ID)
	}
	if result.Status != "done" {
		t.Errorf("expected status 'done', got %s", result.Status)
	}
	if result.Title != "Test Task" {
		t.Errorf("expected Title 'Test Task', got %s", result.Title)
	}
}

func TestCompleteTask_NilDatabase(t *testing.T) {
	var database *db.DB = nil

	result, err := CompleteTask(database, "task-001")

	if err != ErrNilDatabase {
		t.Errorf("expected ErrNilDatabase, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestCompleteTask_EmptyID(t *testing.T) {
	database := setupTestDBComplete(t)
	defer teardownTestDBComplete(t, database)

	result, err := CompleteTask(database, "")

	if err != ErrInvalidID {
		t.Errorf("expected ErrInvalidID, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestCompleteTask_TaskNotFound(t *testing.T) {
	database := setupTestDBComplete(t)
	defer teardownTestDBComplete(t, database)

	result, err := CompleteTask(database, "nonexistent-task")

	if err == nil {
		t.Error("expected error for nonexistent task, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestCompleteTask_UpdatesStatusFromTodo(t *testing.T) {
	database := setupTestDBComplete(t)
	defer teardownTestDBComplete(t, database)

	addInput := &AddTaskInput{
		ID:          "task-002",
		Milestone:   "milestone-1",
		Title:       "Task to Complete",
		Description: "This task should be completed",
		Actor:       "testuser",
	}

	addResult, err := AddTask(database, addInput)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}
	if addResult == nil {
		t.Fatal("expected non-nil add result")
	}
	if addResult.Status != "todo" {
		t.Errorf("expected initial status 'todo', got %s", addResult.Status)
	}

	result, err := CompleteTask(database, "task-002")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Status != "done" {
		t.Errorf("expected status 'done', got %s", result.Status)
	}
}

func TestCompleteTask_UpdatesStatusFromInProgress(t *testing.T) {
	database := setupTestDBComplete(t)
	defer teardownTestDBComplete(t, database)

	addInput := &AddTaskInput{
		ID:          "task-003",
		Milestone:   "milestone-1",
		Title:       "In Progress Task",
		Description: "This task is in progress",
		Actor:       "testuser",
	}

	_, err := AddTask(database, addInput)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	updateInput := &UpdateTaskInput{
		ID:     "task-003",
		Status: "in_progress",
	}
	_, err = UpdateTask(database, updateInput)
	if err != nil {
		t.Fatalf("failed to update task status: %v", err)
	}

	result, err := CompleteTask(database, "task-003")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Status != "done" {
		t.Errorf("expected status 'done', got %s", result.Status)
	}
}
