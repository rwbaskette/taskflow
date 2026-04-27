package db

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) *DB {
	tmpDir := t.TempDir()
	testDBPath := filepath.Join(tmpDir, "test_task.db")
	// Set project root for schema lookup
	os.Setenv("PROJECT_ROOT", tmpDir)
	db, err := NewDB(testDBPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	return db
}

func teardownTestDB(t *testing.T, db *DB) {
	if db != nil {
		db.Close()
	}
}

func TestCreateTask(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	t.Run("valid task", func(t *testing.T) {
		task := &Task{
			ID:          "task-1",
			Milestone:   "milestone-1",
			Title:       "Test Task",
			Description: "Test Description",
			Status:      "todo",
			Actor:       "user-1",
		}

		err := db.CreateTask(task)
		if err != nil {
			t.Fatalf("CreateTask failed: %v", err)
		}

		// Verify task was created
		created, err := db.ReadTask("task-1")
		if err != nil {
			t.Fatalf("ReadTask failed: %v", err)
		}
		if created.ID != task.ID {
			t.Errorf("expected ID %q, got %q", task.ID, created.ID)
		}
		if created.Title != task.Title {
			t.Errorf("expected title %q, got %q", task.Title, created.Title)
		}
	})

	t.Run("task already exists", func(t *testing.T) {
		task := &Task{
			ID:     "task-1",
			Title:  "Duplicate Task",
			Status: "todo",
		}

		err := db.CreateTask(task)
		if err == nil {
			t.Fatal("expected error for duplicate task")
		}
		if !IsTaskAlreadyExists(err) {
			t.Errorf("expected TaskAlreadyExistsError, got %v", err)
		}
	})

	t.Run("nil task", func(t *testing.T) {
		err := db.CreateTask(nil)
		if err == nil {
			t.Fatal("expected error for nil task")
		}
		if err != ErrNilTask {
			t.Errorf("expected ErrNilTask, got %v", err)
		}
	})

	t.Run("empty ID", func(t *testing.T) {
		task := &Task{
			ID:     "",
			Title:  "Test Task",
			Status: "todo",
		}

		err := db.CreateTask(task)
		if err == nil {
			t.Fatal("expected error for empty ID")
		}
		if !IsInvalidTask(err) {
			t.Errorf("expected InvalidTaskError, got %v", err)
		}
	})

	t.Run("empty title", func(t *testing.T) {
		task := &Task{
			ID:     "task-x",
			Title:  "",
			Status: "todo",
		}

		err := db.CreateTask(task)
		if err == nil {
			t.Fatal("expected error for empty title")
		}
		if !IsInvalidTask(err) {
			t.Errorf("expected InvalidTaskError, got %v", err)
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		task := &Task{
			ID:     "task-x",
			Title:  "Test Task",
			Status: "invalid_status",
		}

		err := db.CreateTask(task)
		if err == nil {
			t.Fatal("expected error for invalid status")
		}
		if !IsInvalidTask(err) {
			t.Errorf("expected InvalidTaskError, got %v", err)
		}
	})

	t.Run("LastUpdated set automatically", func(t *testing.T) {
		task := &Task{
			ID:     "task-time",
			Title:  "Test Task",
			Status: "todo",
		}

		before := time.Now().UTC()
		err := db.CreateTask(task)
		after := time.Now().UTC().Add(time.Minute) // Add buffer for processing

		if err != nil {
			t.Fatalf("CreateTask failed: %v", err)
		}

		// Just verify it's non-zero and reasonable (within a range)
		if task.LastUpdated.IsZero() {
			t.Error("LastUpdated should not be zero")
		}
		// Check it's at or after before time
		if task.LastUpdated.Before(before) {
			t.Errorf("LastUpdated should be at or after creation time: %v < %v", task.LastUpdated, before)
		}
		// Check it's reasonably recent (not in the distant future)
		if task.LastUpdated.After(after) {
			t.Errorf("LastUpdated should be reasonable: %v > %v", task.LastUpdated, after)
		}
	})
}

func TestCreateTaskTx(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	t.Run("valid task in transaction", func(t *testing.T) {
		tx, err := db.BeginTx()
		if err != nil {
			t.Fatalf("BeginTx failed: %v", err)
		}
		defer tx.Rollback()

		task := &Task{
			ID:     "tx-task-1",
			Title:  "Tx Test Task",
			Status: "todo",
		}

		err = db.CreateTaskTx(tx, task)
		if err != nil {
			t.Fatalf("CreateTaskTx failed: %v", err)
		}

		tx.Commit()

		// Verify task was created
		created, err := db.ReadTask("tx-task-1")
		if err != nil {
			t.Fatalf("ReadTask failed: %v", err)
		}
		if created.ID != task.ID {
			t.Errorf("expected ID %q, got %q", task.ID, created.ID)
		}
	})

	t.Run("nil transaction", func(t *testing.T) {
		task := &Task{
			ID:     "task-x",
			Title:  "Test Task",
			Status: "todo",
		}

		err := db.CreateTaskTx(nil, task)
		if err == nil {
			t.Fatal("expected error for nil transaction")
		}
	})

	t.Run("task already exists in transaction", func(t *testing.T) {
		// First create a task
		db.CreateTask(&Task{
			ID:     "dup-task",
			Title:  "Original",
			Status: "todo",
		})

		tx, _ := db.BeginTx()
		defer tx.Rollback()

		err := db.CreateTaskTx(tx, &Task{
			ID:     "dup-task",
			Title:  "Duplicate",
			Status: "todo",
		})
		if err == nil {
			t.Fatal("expected error for duplicate task in tx")
		}
		if !IsTaskAlreadyExists(err) {
			t.Errorf("expected TaskAlreadyExistsError, got %v", err)
		}
	})
}

func TestReadTask(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Setup: create a task to read
	db.CreateTask(&Task{
		ID:          "read-task",
		Title:       "Read This Task",
		Description: "Test Description",
		Status:      "in_progress",
		Actor:       "user-1",
	})

	t.Run("valid read", func(t *testing.T) {
		task, err := db.ReadTask("read-task")
		if err != nil {
			t.Fatalf("ReadTask failed: %v", err)
		}
		if task.ID != "read-task" {
			t.Errorf("expected ID %q, got %q", "read-task", task.ID)
		}
		if task.Title != "Read This Task" {
			t.Errorf("expected title %q, got %q", "Read This Task", task.Title)
		}
		if task.Description != "Test Description" {
			t.Errorf("expected description %q, got %q", "Test Description", task.Description)
		}
		if task.Status != "in_progress" {
			t.Errorf("expected status %q, got %q", "in_progress", task.Status)
		}
		if task.Actor != "user-1" {
			t.Errorf("expected actor %q, got %q", "user-1", task.Actor)
		}
	})

	t.Run("task not found", func(t *testing.T) {
		_, err := db.ReadTask("nonexistent-id")
		if err == nil {
			t.Fatal("expected error for nonexistent task")
		}
		if !IsTaskNotFound(err) {
			t.Errorf("expected TaskNotFoundError, got %v", err)
		}
	})

	t.Run("empty ID", func(t *testing.T) {
		_, err := db.ReadTask("")
		if err == nil {
			t.Fatal("expected error for empty ID")
		}
		if err != ErrInvalidID {
			t.Errorf("expected ErrInvalidID, got %v", err)
		}
	})
}

func TestReadTaskTx(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Setup: create a task to read
	db.CreateTask(&Task{
		ID:     "tx-read-task",
		Title:  "Tx Read This Task",
		Status: "done",
	})

	t.Run("valid read in transaction", func(t *testing.T) {
		tx, err := db.BeginTx()
		if err != nil {
			t.Fatalf("BeginTx failed: %v", err)
		}
		defer tx.Rollback()

		task, err := db.ReadTaskTx(tx, "tx-read-task")
		if err != nil {
			t.Fatalf("ReadTaskTx failed: %v", err)
		}
		if task.ID != "tx-read-task" {
			t.Errorf("expected ID %q, got %q", "tx-read-task", task.ID)
		}
	})

	t.Run("nil transaction", func(t *testing.T) {
		_, err := db.ReadTaskTx(nil, "task-1")
		if err == nil {
			t.Fatal("expected error for nil transaction")
		}
	})

	t.Run("task not found in transaction", func(t *testing.T) {
		tx, _ := db.BeginTx()
		defer tx.Rollback()

		_, err := db.ReadTaskTx(tx, "nonexistent")
		if err == nil {
			t.Fatal("expected error for nonexistent task in tx")
		}
		if !IsTaskNotFound(err) {
			t.Errorf("expected TaskNotFoundError, got %v", err)
		}
	})
}

func TestUpdateTask(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Setup: create a task to update
	db.CreateTask(&Task{
		ID:     "update-task",
		Title:  "Original Title",
		Status: "todo",
	})

	t.Run("valid update", func(t *testing.T) {
		task := &Task{
			ID:          "update-task",
			Title:       "Updated Title",
			Description: "Updated Description",
			Status:      "done",
			Actor:       "user-2",
		}

		err := db.UpdateTask(task)
		if err != nil {
			t.Fatalf("UpdateTask failed: %v", err)
		}

		// Verify update
		updated, _ := db.ReadTask("update-task")
		if updated.Title != "Updated Title" {
			t.Errorf("expected title %q, got %q", "Updated Title", updated.Title)
		}
		if updated.Description != "Updated Description" {
			t.Errorf("expected description %q, got %q", "Updated Description", updated.Description)
		}
		if updated.Status != "done" {
			t.Errorf("expected status %q, got %q", "done", updated.Status)
		}
		if updated.Actor != "user-2" {
			t.Errorf("expected actor %q, got %q", "user-2", updated.Actor)
		}
	})

	t.Run("task not found", func(t *testing.T) {
		task := &Task{
			ID:     "nonexistent",
			Title:  "Test Task",
			Status: "todo",
		}

		err := db.UpdateTask(task)
		if err == nil {
			t.Fatal("expected error for nonexistent task")
		}
		if !IsTaskNotFound(err) {
			t.Errorf("expected TaskNotFoundError, got %v", err)
		}
	})

	t.Run("nil task", func(t *testing.T) {
		err := db.UpdateTask(nil)
		if err == nil {
			t.Fatal("expected error for nil task")
		}
		if err != ErrNilTask {
			t.Errorf("expected ErrNilTask, got %v", err)
		}
	})

	t.Run("empty ID", func(t *testing.T) {
		task := &Task{
			ID:     "",
			Title:  "Test Task",
			Status: "todo",
		}

		err := db.UpdateTask(task)
		if err == nil {
			t.Fatal("expected error for empty ID")
		}
		if !IsInvalidTask(err) {
			t.Errorf("expected InvalidTaskError, got %v", err)
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		task := &Task{
			ID:     "update-task",
			Title:  "Test Task",
			Status: "invalid",
		}

		err := db.UpdateTask(task)
		if err == nil {
			t.Fatal("expected error for invalid status")
		}
		if !IsInvalidTask(err) {
			t.Errorf("expected InvalidTaskError, got %v", err)
		}
	})

	t.Run("LastUpdated updated automatically", func(t *testing.T) {
		// Get original task
		original, _ := db.ReadTask("update-task")
		originalTitle := original.Title

		// Wait a bit to ensure time difference
		time.Sleep(10 * time.Millisecond)

		task := &Task{
			ID:     "update-task",
			Title:  "New Title",
			Status: "done",
		}

		err := db.UpdateTask(task)
		if err != nil {
			t.Fatalf("UpdateTask failed: %v", err)
		}

		updated, _ := db.ReadTask("update-task")
		if updated.Title != "New Title" {
			t.Errorf("expected title %q, got %q", "New Title", updated.Title)
		}
		if updated.Title == originalTitle {
			t.Error("LastUpdated should have changed")
		}
	})
}

func TestUpdateTaskTx(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Setup
	db.CreateTask(&Task{
		ID:     "tx-update-task",
		Title:  "Original",
		Status: "todo",
	})

	t.Run("valid update in transaction", func(t *testing.T) {
		tx, err := db.BeginTx()
		if err != nil {
			t.Fatalf("BeginTx failed: %v", err)
		}
		defer tx.Rollback()

		task := &Task{
			ID:     "tx-update-task",
			Title:  "Tx Updated",
			Status: "done",
		}

		err = db.UpdateTaskTx(tx, task)
		if err != nil {
			t.Fatalf("UpdateTaskTx failed: %v", err)
		}

		tx.Commit()

		updated, _ := db.ReadTask("tx-update-task")
		if updated.Title != "Tx Updated" {
			t.Errorf("expected title %q, got %q", "Tx Updated", updated.Title)
		}
	})

	t.Run("nil transaction", func(t *testing.T) {
		task := &Task{
			ID:     "task-x",
			Title:  "Test Task",
			Status: "todo",
		}

		err := db.UpdateTaskTx(nil, task)
		if err == nil {
			t.Fatal("expected error for nil transaction")
		}
	})

	t.Run("task not found in transaction", func(t *testing.T) {
		tx, _ := db.BeginTx()
		defer tx.Rollback()

		err := db.UpdateTaskTx(tx, &Task{
			ID:     "nonexistent",
			Title:  "Test",
			Status: "todo",
		})
		if err == nil {
			t.Fatal("expected error for nonexistent task in tx")
		}
		if !IsTaskNotFound(err) {
			t.Errorf("expected TaskNotFoundError, got %v", err)
		}
	})
}

func TestDeleteTask(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Setup: create a task to delete
	db.CreateTask(&Task{
		ID:     "delete-task",
		Title:  "Delete Me",
		Status: "todo",
	})

	t.Run("valid delete", func(t *testing.T) {
		err := db.DeleteTask("delete-task")
		if err != nil {
			t.Fatalf("DeleteTask failed: %v", err)
		}

		// Verify deletion
		_, err = db.ReadTask("delete-task")
		if err == nil {
			t.Fatal("expected error after deletion")
		}
		if !IsTaskNotFound(err) {
			t.Errorf("expected TaskNotFoundError, got %v", err)
		}
	})

	t.Run("task not found", func(t *testing.T) {
		err := db.DeleteTask("nonexistent-id")
		if err == nil {
			t.Fatal("expected error for nonexistent task")
		}
		if !IsTaskNotFound(err) {
			t.Errorf("expected TaskNotFoundError, got %v", err)
		}
	})

	t.Run("empty ID", func(t *testing.T) {
		err := db.DeleteTask("")
		if err == nil {
			t.Fatal("expected error for empty ID")
		}
		if err != ErrInvalidID {
			t.Errorf("expected ErrInvalidID, got %v", err)
		}
	})
}

func TestDeleteTaskTx(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Setup
	db.CreateTask(&Task{
		ID:     "tx-delete-task",
		Title:  "Delete Me",
		Status: "todo",
	})

	t.Run("valid delete in transaction", func(t *testing.T) {
		tx, err := db.BeginTx()
		if err != nil {
			t.Fatalf("BeginTx failed: %v", err)
		}
		defer tx.Rollback()

		err = db.DeleteTaskTx(tx, "tx-delete-task")
		if err != nil {
			t.Fatalf("DeleteTaskTx failed: %v", err)
		}

		tx.Commit()

		_, err = db.ReadTask("tx-delete-task")
		if err == nil {
			t.Fatal("expected error after deletion in tx")
		}
		if !IsTaskNotFound(err) {
			t.Errorf("expected TaskNotFoundError, got %v", err)
		}
	})

	t.Run("nil transaction", func(t *testing.T) {
		err := db.DeleteTaskTx(nil, "task-1")
		if err == nil {
			t.Fatal("expected error for nil transaction")
		}
	})

	t.Run("task not found in transaction", func(t *testing.T) {
		tx, _ := db.BeginTx()
		defer tx.Rollback()

		err := db.DeleteTaskTx(tx, "nonexistent")
		if err == nil {
			t.Fatal("expected error for nonexistent task in tx")
		}
		if !IsTaskNotFound(err) {
			t.Errorf("expected TaskNotFoundError, got %v", err)
		}
	})
}

func TestListTasks(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Setup: create multiple tasks
	tasks := []*Task{
		{ID: "task-1", Milestone: "m1", Title: "Task 1", Status: "todo", Actor: "user-1"},
		{ID: "task-2", Milestone: "m1", Title: "Task 2", Status: "in_progress", Actor: "user-1"},
		{ID: "task-3", Milestone: "m2", Title: "Task 3", Status: "done", Actor: "user-2"},
		{ID: "task-4", Milestone: "m2", Title: "Task 4", Status: "todo", Actor: "user-2"},
	}

	for _, t := range tasks {
		db.CreateTask(t)
	}

	t.Run("list all tasks", func(t *testing.T) {
		result, err := db.ListTasks(TaskFilter{})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if len(result) != 4 {
			t.Errorf("expected 4 tasks, got %d", len(result))
		}
	})

	t.Run("filter by milestone", func(t *testing.T) {
		result, err := db.ListTasks(TaskFilter{Milestone: "m1"})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 tasks, got %d", len(result))
		}
		for _, task := range result {
			if task.Milestone != "m1" {
				t.Errorf("expected milestone m1, got %q", task.Milestone)
			}
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		result, err := db.ListTasks(TaskFilter{Status: "todo"})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 tasks, got %d", len(result))
		}
		for _, task := range result {
			if task.Status != "todo" {
				t.Errorf("expected status todo, got %q", task.Status)
			}
		}
	})

	t.Run("filter by actor", func(t *testing.T) {
		result, err := db.ListTasks(TaskFilter{Actor: "user-1"})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 tasks, got %d", len(result))
		}
		for _, task := range result {
			if task.Actor != "user-1" {
				t.Errorf("expected actor user-1, got %q", task.Actor)
			}
		}
	})

	t.Run("filter by multiple criteria", func(t *testing.T) {
		result, err := db.ListTasks(TaskFilter{Milestone: "m1", Actor: "user-1"})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 tasks, got %d", len(result))
		}
	})

	t.Run("pagination with limit", func(t *testing.T) {
		result, err := db.ListTasks(TaskFilter{Limit: 2})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 tasks, got %d", len(result))
		}
	})

	t.Run("pagination with limit and offset", func(t *testing.T) {
		result, err := db.ListTasks(TaskFilter{Limit: 2, Offset: 2})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 tasks, got %d", len(result))
		}
	})

	t.Run("empty result", func(t *testing.T) {
		result, err := db.ListTasks(TaskFilter{Milestone: "nonexistent"})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected 0 tasks, got %d", len(result))
		}
	})

	t.Run("returns empty slice not nil", func(t *testing.T) {
		result, _ := db.ListTasks(TaskFilter{Milestone: "nonexistent"})
		if result == nil {
			t.Error("expected empty slice, got nil")
		}
	})
}

func TestBeginTx(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	t.Run("successful transaction start", func(t *testing.T) {
		tx, err := db.BeginTx()
		if err != nil {
			t.Fatalf("BeginTx failed: %v", err)
		}
		if tx == nil {
			t.Fatal("expected non-nil transaction")
		}
		tx.Rollback()
	})

	t.Run("transaction can execute queries", func(t *testing.T) {
		tx, _ := db.BeginTx()

		// Create task within transaction
		task := &Task{
			ID:     "tx-test",
			Title:  "Transaction Test",
			Status: "todo",
		}
		err := db.CreateTaskTx(tx, task)
		if err != nil {
			t.Fatalf("CreateTaskTx failed: %v", err)
		}

		// Read within transaction
		readTask, err := db.ReadTaskTx(tx, "tx-test")
		if err != nil {
			t.Fatalf("ReadTaskTx failed: %v", err)
		}
		if readTask.ID != "tx-test" {
			t.Errorf("expected ID %q, got %q", "tx-test", readTask.ID)
		}

		tx.Commit()

		// Verify task committed
		committed, _ := db.ReadTask("tx-test")
		if committed == nil {
			t.Error("expected task to be committed")
		}
	})

	t.Run("transaction rollback", func(t *testing.T) {
		tx, _ := db.BeginTx()

		// Create task within transaction
		task := &Task{
			ID:     "tx-rollback-test",
			Title:  "Rollback Test",
			Status: "todo",
		}
		db.CreateTaskTx(tx, task)

		// Rollback
		tx.Rollback()

		// Verify task was not committed
		_, err := db.ReadTask("tx-rollback-test")
		if err == nil {
			t.Error("expected task to be rolled back")
		}
	})
}

func TestErrorPredicates(t *testing.T) {
	t.Run("IsTaskNotFound with direct error", func(t *testing.T) {
		err := NewTaskNotFoundError("test-id")
		if !IsTaskNotFound(err) {
			t.Error("IsTaskNotFound should return true for TaskNotFoundError")
		}
	})

	t.Run("IsTaskNotFound with wrapped error", func(t *testing.T) {
		err := NewTaskNotFoundError("test-id")
		wrapped := fmt.Errorf("wrapped: %w", err)
		if !IsTaskNotFound(wrapped) {
			t.Error("IsTaskNotFound should return true for wrapped TaskNotFoundError")
		}
	})

	t.Run("IsTaskNotFound with nil", func(t *testing.T) {
		if IsTaskNotFound(nil) {
			t.Error("IsTaskNotFound should return false for nil")
		}
	})

	t.Run("IsTaskAlreadyExists with direct error", func(t *testing.T) {
		err := NewTaskAlreadyExistsError("test-id")
		if !IsTaskAlreadyExists(err) {
			t.Error("IsTaskAlreadyExists should return true for TaskAlreadyExistsError")
		}
	})

	t.Run("IsTaskAlreadyExists with wrapped error", func(t *testing.T) {
		err := NewTaskAlreadyExistsError("test-id")
		wrapped := fmt.Errorf("wrapped: %w", err)
		if !IsTaskAlreadyExists(wrapped) {
			t.Error("IsTaskAlreadyExists should return true for wrapped TaskAlreadyExistsError")
		}
	})

	t.Run("IsInvalidTask with direct error", func(t *testing.T) {
		err := NewInvalidTaskError("field", "message")
		if !IsInvalidTask(err) {
			t.Error("IsInvalidTask should return true for InvalidTaskError")
		}
	})

	t.Run("IsInvalidTask with wrapped error", func(t *testing.T) {
		err := NewInvalidTaskError("field", "message")
		wrapped := fmt.Errorf("wrapped: %w", err)
		if !IsInvalidTask(wrapped) {
			t.Error("IsInvalidTask should return true for wrapped InvalidTaskError")
		}
	})
}
