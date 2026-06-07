package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rwbaskette/taskflow/internal/db"
)

func setupTestDBForDeps(t *testing.T) *db.DB {
	tmpDir := t.TempDir()
	testDBPath := filepath.Join(tmpDir, "test_validate_deps.db")
	// Set project root for schema lookup
	os.Setenv("PROJECT_ROOT", tmpDir)
	testDB, err := db.NewDB(testDBPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	return testDB
}

func teardownTestDBForDeps(t *testing.T, testDB *db.DB) {
	if testDB != nil {
		testDB.Close()
	}
}

func TestValidateBlockedBy_EmptyBlockedBy(t *testing.T) {
	database := setupTestDBForDeps(t)
	defer teardownTestDBForDeps(t, database)

	// Empty blockedBy should be valid
	err := ValidateBlockedBy(database, []string{}, "task-1")
	if err != nil {
		t.Errorf("expected nil error for empty blockedBy, got %v", err)
	}

	// Nil blockedBy should also be valid
	err = ValidateBlockedBy(database, nil, "task-1")
	if err != nil {
		t.Errorf("expected nil error for nil blockedBy, got %v", err)
	}
}

func TestValidateBlockedBy_TaskBlocksItself(t *testing.T) {
	database := setupTestDBForDeps(t)
	defer teardownTestDBForDeps(t, database)

	// Create two tasks
	task1 := &db.Task{
		ID:        "task-1",
		Title:     "Task 1",
		Status:    "todo",
		Created:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	err := database.CreateTask(task1)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Try to add task-1 as blocked by ["task-1"] - should fail
	err = ValidateBlockedBy(database, []string{"task-1"}, "task-1")
	if err == nil {
		t.Error("expected error when task blocks itself, got nil")
	}
	if !IsCircularDependency(err) {
		t.Errorf("expected ErrCircularDependency, got %v", err)
	}
}

func TestValidateBlockedBy_NonExistentBlocker(t *testing.T) {
	database := setupTestDBForDeps(t)
	defer teardownTestDBForDeps(t, database)

	// Create a task
	task1 := &db.Task{
		ID:        "task-1",
		Title:     "Task 1",
		Status:    "todo",
		Created:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	err := database.CreateTask(task1)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Try to add blocked_by with non-existent task
	err = ValidateBlockedBy(database, []string{"non-existent"}, "task-1")
	if err == nil {
		t.Error("expected error for non-existent blocker, got nil")
	}
	if !IsInvalidBlockedByTask(err) {
		t.Errorf("expected ErrInvalidBlockedByTask, got %v", err)
	}
}

func TestValidateBlockedBy_ValidBlockedBy(t *testing.T) {
	database := setupTestDBForDeps(t)
	defer teardownTestDBForDeps(t, database)

	// Create tasks
	task1 := &db.Task{
		ID:        "task-1",
		Title:     "Task 1",
		Status:    "todo",
		Created:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	err := database.CreateTask(task1)
	if err != nil {
		t.Fatalf("failed to create task1: %v", err)
	}

	task2 := &db.Task{
		ID:        "task-2",
		Title:     "Task 2",
		Status:    "todo",
		Created:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	err = database.CreateTask(task2)
	if err != nil {
		t.Fatalf("failed to create task2: %v", err)
	}

	// task-3 blocked by task-1 and task-2 should be valid
	err = ValidateBlockedBy(database, []string{"task-1", "task-2"}, "task-3")
	if err != nil {
		t.Errorf("expected nil error for valid blockedBy, got %v", err)
	}
}

func TestValidateBlockedBy_CircularDependency(t *testing.T) {
	database := setupTestDBForDeps(t)
	defer teardownTestDBForDeps(t, database)

	// Create A, B, C where A blocks B
	taskA := &db.Task{
		ID:        "A",
		Title:     "Task A",
		Status:    "todo",
		Created:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	err := database.CreateTask(taskA)
	if err != nil {
		t.Fatalf("failed to create task A: %v", err)
	}

	taskB := &db.Task{
		ID:        "B",
		Title:     "Task B",
		Status:    "todo",
		BlockedBy: []string{"A"}, // B is blocked by A
		Created:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	err = database.CreateTask(taskB)
	if err != nil {
		t.Fatalf("failed to create task B: %v", err)
	}

	// Now try to make A blocked by B - this would create A -> B -> A cycle
	err = ValidateBlockedBy(database, []string{"B"}, "A")
	if err == nil {
		t.Error("expected circular dependency error, got nil")
	}
	if !IsCircularDependency(err) {
		t.Errorf("expected ErrCircularDependency, got %v", err)
	}
}

func TestValidateBlockedBy_TransitiveCircularDependency(t *testing.T) {
	database := setupTestDBForDeps(t)
	defer teardownTestDBForDeps(t, database)

	// Create A blocks B, B blocks C
	taskA := &db.Task{
		ID:        "A",
		Title:     "Task A",
		Status:    "todo",
		Created:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	err := database.CreateTask(taskA)
	if err != nil {
		t.Fatalf("failed to create task A: %v", err)
	}

	taskB := &db.Task{
		ID:        "B",
		Title:     "Task B",
		Status:    "todo",
		BlockedBy: []string{"A"}, // B is blocked by A
		Created:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	err = database.CreateTask(taskB)
	if err != nil {
		t.Fatalf("failed to create task B: %v", err)
	}

	taskC := &db.Task{
		ID:        "C",
		Title:     "Task C",
		Status:    "todo",
		BlockedBy: []string{"B"}, // C is blocked by B
		Created:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	err = database.CreateTask(taskC)
	if err != nil {
		t.Fatalf("failed to create task C: %v", err)
	}

	// Now try to make A blocked by C - this would create A -> B -> C -> A cycle
	err = ValidateBlockedBy(database, []string{"C"}, "A")
	if err == nil {
		t.Error("expected circular dependency error, got nil")
	}
	if !IsCircularDependency(err) {
		t.Errorf("expected ErrCircularDependency, got %v", err)
	}
}

func TestValidateBlockedBy_NilDatabase(t *testing.T) {
	var database *db.DB = nil

	err := ValidateBlockedBy(database, []string{"task-1"}, "task-2")
	if err != ErrNilDatabase {
		t.Errorf("expected ErrNilDatabase, got %v", err)
	}
}

func TestAddTask_WithBlockedBy(t *testing.T) {
	database := setupTestDBForDeps(t)
	defer teardownTestDBForDeps(t, database)

	// Create blocking tasks
	task1 := &db.Task{
		ID:        "blocker-1",
		Title:     "Blocker 1",
		Status:    "todo",
		Created:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	err := database.CreateTask(task1)
	if err != nil {
		t.Fatalf("failed to create task1: %v", err)
	}

	task2 := &db.Task{
		ID:        "blocker-2",
		Title:     "Blocker 2",
		Status:    "todo",
		Created:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	err = database.CreateTask(task2)
	if err != nil {
		t.Fatalf("failed to create task2: %v", err)
	}

	// Add task with blocked_by
	input := &AddTaskInput{
		ID:        "task-new",
		Title:     "New Task",
		Milestone: "m1",
		BlockedBy: []string{"blocker-1", "blocker-2"},
	}

	result, err := AddTask(database, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.BlockedBy) != 2 {
		t.Errorf("expected 2 blockers, got %d", len(result.BlockedBy))
	}
}

func TestAddTask_InvalidBlockedBy(t *testing.T) {
	database := setupTestDBForDeps(t)
	defer teardownTestDBForDeps(t, database)

	// Add task with non-existent blocker
	input := &AddTaskInput{
		ID:        "task-new",
		Title:     "New Task",
		Milestone: "m1",
		BlockedBy: []string{"non-existent"},
	}

	result, err := AddTask(database, input)
	if err == nil {
		t.Fatal("expected error for invalid blocker")
	}
	if !IsInvalidBlockedByTask(err) {
		t.Errorf("expected ErrInvalidBlockedByTask, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestUpdateTask_UpdateBlockedBy(t *testing.T) {
	database := setupTestDBForDeps(t)
	defer teardownTestDBForDeps(t, database)

	// Create tasks
	task1 := &db.Task{
		ID:        "task-1",
		Title:     "Task 1",
		Status:    "todo",
		Created:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	err := database.CreateTask(task1)
	if err != nil {
		t.Fatalf("failed to create task1: %v", err)
	}

	task2 := &db.Task{
		ID:        "task-2",
		Title:     "Task 2",
		Status:    "todo",
		Created:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	err = database.CreateTask(task2)
	if err != nil {
		t.Fatalf("failed to create task2: %v", err)
	}

	// Update task-1 to be blocked by task-2
	input := &UpdateTaskInput{
		ID:        "task-1",
		BlockedBy: []string{"task-2"},
	}

	result, err := UpdateTask(database, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.BlockedBy) != 1 || result.BlockedBy[0] != "task-2" {
		t.Errorf("expected blockedBy ['task-2'], got %v", result.BlockedBy)
	}
}

func TestUpdateTask_InvalidBlockedByUpdate(t *testing.T) {
	database := setupTestDBForDeps(t)
	defer teardownTestDBForDeps(t, database)

	// Create task
	task1 := &db.Task{
		ID:        "task-1",
		Title:     "Task 1",
		Status:    "todo",
		Created:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	err := database.CreateTask(task1)
	if err != nil {
		t.Fatalf("failed to create task1: %v", err)
	}

	// Try to update with non-existent blocker
	input := &UpdateTaskInput{
		ID:        "task-1",
		BlockedBy: []string{"non-existent"},
	}

	result, err := UpdateTask(database, input)
	if err == nil {
		t.Fatal("expected error for invalid blocker")
	}
	if !IsInvalidBlockedByTask(err) {
		t.Errorf("expected ErrInvalidBlockedByTask, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestUpdateTask_CircularDependencyUpdate(t *testing.T) {
	database := setupTestDBForDeps(t)
	defer teardownTestDBForDeps(t, database)

	// Create A and B where A blocks B
	taskA := &db.Task{
		ID:        "A",
		Title:     "Task A",
		Status:    "todo",
		Created:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	err := database.CreateTask(taskA)
	if err != nil {
		t.Fatalf("failed to create task A: %v", err)
	}

	taskB := &db.Task{
		ID:        "B",
		Title:     "Task B",
		Status:    "todo",
		BlockedBy: []string{"A"},
		Created:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	err = database.CreateTask(taskB)
	if err != nil {
		t.Fatalf("failed to create task B: %v", err)
	}

	// Try to make A blocked by B - should fail (circular)
	input := &UpdateTaskInput{
		ID:        "A",
		BlockedBy: []string{"B"},
	}

	result, err := UpdateTask(database, input)
	if err == nil {
		t.Fatal("expected circular dependency error")
	}
	if !IsCircularDependency(err) {
		t.Errorf("expected ErrCircularDependency, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestIsCircularDependency(t *testing.T) {
	err := NewCircularDependencyError("task-1", []string{"task-2"})
	if !IsCircularDependency(err) {
		t.Error("expected true for CircularDependencyError")
	}
	if IsCircularDependency(nil) {
		t.Error("expected false for nil")
	}
}

func TestIsInvalidBlockedByTask(t *testing.T) {
	err := NewInvalidBlockedByTaskError("task-1")
	if !IsInvalidBlockedByTask(err) {
		t.Error("expected true for InvalidBlockedByTaskError")
	}
	if IsInvalidBlockedByTask(nil) {
		t.Error("expected false for nil")
	}
}

func TestValidateBlockedBy_DuplicateBlockers(t *testing.T) {
	database := setupTestDBForDeps(t)
	defer teardownTestDBForDeps(t, database)

	// Create a task
	task1 := &db.Task{
		ID:        "task-1",
		Title:     "Task 1",
		Status:    "todo",
		Created:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	err := database.CreateTask(task1)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	task2 := &db.Task{
		ID:        "task-2",
		Title:     "Task 2",
		Status:    "todo",
		Created:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	err = database.CreateTask(task2)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Try to add task-2 as blocked by ["task-1", "task-1"] - should fail
	err = ValidateBlockedBy(database, []string{"task-1", "task-1"}, "task-2")
	if err == nil {
		t.Error("expected error for duplicate blockers, got nil")
	}
	if !IsDuplicateBlockedBy(err) {
		t.Errorf("expected ErrDuplicateBlockedBy, got %v", err)
	}
}

func TestUpdateTask_ClearBlockedBy(t *testing.T) {
	database := setupTestDBForDeps(t)
	defer teardownTestDBForDeps(t, database)

	// Create task-1 and task-2
	task1 := &db.Task{
		ID:        "task-1",
		Title:     "Task 1",
		Status:    "todo",
		Created:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	err := database.CreateTask(task1)
	if err != nil {
		t.Fatalf("failed to create task1: %v", err)
	}

	task2 := &db.Task{
		ID:        "task-2",
		Title:     "Task 2",
		Status:    "todo",
		Created:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	err = database.CreateTask(task2)
	if err != nil {
		t.Fatalf("failed to create task2: %v", err)
	}

	// First, set task-1 to be blocked by task-2
	input := &UpdateTaskInput{
		ID:        "task-1",
		BlockedBy: []string{"task-2"},
	}

	result, err := UpdateTask(database, input)
	if err != nil {
		t.Fatalf("unexpected error setting blockedBy: %v", err)
	}
	if len(result.BlockedBy) != 1 || result.BlockedBy[0] != "task-2" {
		t.Errorf("expected blockedBy ['task-2'], got %v", result.BlockedBy)
	}

	// Now clear blockedBy by updating with empty array
	clearInput := &UpdateTaskInput{
		ID:        "task-1",
		BlockedBy: []string{},
	}

	clearResult, err := UpdateTask(database, clearInput)
	if err != nil {
		t.Fatalf("unexpected error clearing blockedBy: %v", err)
	}
	if len(clearResult.BlockedBy) != 0 {
		t.Errorf("expected empty blockedBy, got %v", clearResult.BlockedBy)
	}
}

func TestValidateBlockedBy_LargeDependencyChain(t *testing.T) {
	database := setupTestDBForDeps(t)
	defer teardownTestDBForDeps(t, database)

	// Create a chain: A blocks B, B blocks C, C blocks D, D blocks E
	// This means: A is blocked by B, B is blocked by C, C is blocked by D, D is blocked by E
	// Graph direction: task -> blocker (so A->B, B->C, C->D, D->E means A blocked by B, B blocked by C, etc.)
	tasks := []string{"A", "B", "C", "D", "E"}

	for i, id := range tasks {
		var blockedBy []string
		if i > 0 {
			blockedBy = []string{tasks[i]}
		}
		// A is blocked by B, B is blocked by C, C is blocked by D, D is blocked by E
		if id == "A" {
			blockedBy = []string{"B"}
		} else if id == "B" {
			blockedBy = []string{"C"}
		} else if id == "C" {
			blockedBy = []string{"D"}
		} else if id == "D" {
			blockedBy = []string{"E"}
		}

		taskObj := &db.Task{
			ID:        id,
			Title:     "Task " + id,
			Status:    "todo",
			BlockedBy: blockedBy,
			Created:   time.Now().UTC(),
			LastUpdated: time.Now().UTC(),
		}
		err := database.CreateTask(taskObj)
		if err != nil {
			t.Fatalf("failed to create task %s: %v", id, err)
		}
	}

	// Try to make E blocked by A - this would create a cycle:
	// A is blocked by B, B by C, C by D, D by E (existing chain)
	// If E is also blocked by A, then E can reach A via D->C->B->A, and A would point to E = cycle
	err := ValidateBlockedBy(database, []string{"A"}, "E")
	if err == nil {
		t.Error("expected circular dependency error for large chain, got nil")
	}
	if !IsCircularDependency(err) {
		t.Errorf("expected ErrCircularDependency, got %v", err)
	}

	// Validate that adding F blocked by E is valid (no cycle)
	taskF := &db.Task{
		ID:        "F",
		Title:     "Task F",
		Status:    "todo",
		Created:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	err = database.CreateTask(taskF)
	if err != nil {
		t.Fatalf("failed to create task F: %v", err)
	}

	// F blocked by E should be valid - F->E doesn't create a cycle
	err = ValidateBlockedBy(database, []string{"E"}, "F")
	if err != nil {
		t.Errorf("expected nil error for valid large chain dependency, got %v", err)
	}
}

func TestIsDuplicateBlockedBy(t *testing.T) {
	err := NewDuplicateBlockedByError("task-1")
	if !IsDuplicateBlockedBy(err) {
		t.Error("expected true for DuplicateBlockedByError")
	}
	if IsDuplicateBlockedBy(nil) {
		t.Error("expected false for nil")
	}
}