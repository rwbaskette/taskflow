package db

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) *DB {
	tmpDir := t.TempDir()
	testDBPath := filepath.Join(tmpDir, "test_task.db")
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

func assertTaskNotFound(t *testing.T, err error) {
	t.Helper()
	var target *TaskNotFoundError
	if !errors.As(err, &target) {
		t.Errorf("expected TaskNotFoundError, got %v", err)
	}
}

func assertTaskAlreadyExists(t *testing.T, err error) {
	t.Helper()
	var target *TaskAlreadyExistsError
	if !errors.As(err, &target) {
		t.Errorf("expected TaskAlreadyExistsError, got %v", err)
	}
}

func assertInvalidTask(t *testing.T, err error) {
	t.Helper()
	var target *InvalidTaskError
	if !errors.As(err, &target) {
		t.Errorf("expected InvalidTaskError, got %v", err)
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
		assertTaskAlreadyExists(t, err)
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
		assertInvalidTask(t, err)
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
		assertInvalidTask(t, err)
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
		assertInvalidTask(t, err)
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
		assertTaskNotFound(t, err)
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
		assertTaskNotFound(t, err)
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
		assertInvalidTask(t, err)
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
		assertInvalidTask(t, err)
	})

	t.Run("LastUpdated updated automatically", func(t *testing.T) {
		// Get original task
		original, _ := db.ReadTask("update-task")
		originalUpdated := original.LastUpdated

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
		// Timestamps are stored at second granularity (RFC3339), so assert
		// non-decreasing rather than strictly increasing.
		if updated.LastUpdated.Before(originalUpdated) {
			t.Errorf("expected LastUpdated to not go backwards: %v < %v", updated.LastUpdated, originalUpdated)
		}
	})
}

func TestSoftDeleteTask(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	t.Run("soft delete returns stored timestamp and removes task", func(t *testing.T) {
		db.CreateTask(&Task{
			ID:     "softdel-task",
			Title:  "Delete Me",
			Status: "todo",
		})

		before := time.Now().UTC()
		deletedOn, err := db.SoftDeleteTask("softdel-task")
		if err != nil {
			t.Fatalf("SoftDeleteTask failed: %v", err)
		}
		after := time.Now().UTC().Add(time.Minute)

		// The returned timestamp must match what was stored (non-zero,
		// within the call window).
		if deletedOn.IsZero() {
			t.Error("expected non-zero deleted_on")
		}
		if deletedOn.Before(before) || deletedOn.After(after) {
			t.Errorf("expected deleted_on within call window, got %v", deletedOn)
		}

		// Verify the task no longer exists in the active table
		_, err = db.ReadTask("softdel-task")
		if err == nil {
			t.Fatal("expected error after soft deletion")
		}
		assertTaskNotFound(t, err)
	})

	t.Run("soft delete task not found", func(t *testing.T) {
		_, err := db.SoftDeleteTask("nonexistent-id")
		if err == nil {
			t.Fatal("expected error for nonexistent task")
		}
		assertTaskNotFound(t, err)
	})

	t.Run("soft delete empty ID", func(t *testing.T) {
		_, err := db.SoftDeleteTask("")
		if err == nil {
			t.Fatal("expected error for empty ID")
		}
		if err != ErrInvalidID {
			t.Errorf("expected ErrInvalidID, got %v", err)
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

func TestCountTasks(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Setup: create multiple tasks
	tasks := []*Task{
		{ID: "task-1", Milestone: "m1", Sprint: "s1", Title: "Task 1", Status: "todo", Actor: "user-1"},
		{ID: "task-2", Milestone: "m1", Sprint: "s1", Title: "Task 2", Status: "in_progress", Actor: "user-1"},
		{ID: "task-3", Milestone: "m2", Sprint: "s2", Title: "Task 3", Status: "done", Actor: "user-2"},
		{ID: "task-4", Milestone: "m2", Sprint: "s2", Title: "Task 4", Status: "todo", Actor: "user-2"},
		{ID: "task-5", Milestone: "m1", Sprint: "s3", Title: "Task 5", Status: "blocked", Actor: "user-1"},
	}

	for _, task := range tasks {
		if err := db.CreateTask(task); err != nil {
			t.Fatalf("CreateTask failed: %v", err)
		}
	}

	t.Run("count all tasks", func(t *testing.T) {
		count, err := db.CountTasks(TaskFilter{})
		if err != nil {
			t.Fatalf("CountTasks failed: %v", err)
		}
		if count != 5 {
			t.Errorf("expected count=5, got %d", count)
		}
	})

	t.Run("count ignores limit and offset", func(t *testing.T) {
		// Even with limit=2 and offset=1, count should return the full total
		count, err := db.CountTasks(TaskFilter{Limit: 2, Offset: 1})
		if err != nil {
			t.Fatalf("CountTasks failed: %v", err)
		}
		if count != 5 {
			t.Errorf("expected count=5 (ignoring limit/offset), got %d", count)
		}
	})

	t.Run("count with milestone filter", func(t *testing.T) {
		count, err := db.CountTasks(TaskFilter{Milestone: "m1"})
		if err != nil {
			t.Fatalf("CountTasks failed: %v", err)
		}
		if count != 3 {
			t.Errorf("expected count=3 for milestone m1, got %d", count)
		}
	})

	t.Run("count with status filter", func(t *testing.T) {
		count, err := db.CountTasks(TaskFilter{Status: "todo"})
		if err != nil {
			t.Fatalf("CountTasks failed: %v", err)
		}
		if count != 2 {
			t.Errorf("expected count=2 for status todo, got %d", count)
		}
	})

	t.Run("count with actor filter", func(t *testing.T) {
		count, err := db.CountTasks(TaskFilter{Actor: "user-1"})
		if err != nil {
			t.Fatalf("CountTasks failed: %v", err)
		}
		if count != 3 {
			t.Errorf("expected count=3 for actor user-1, got %d", count)
		}
	})

	t.Run("count with id filter", func(t *testing.T) {
		count, err := db.CountTasks(TaskFilter{ID: "task-3"})
		if err != nil {
			t.Fatalf("CountTasks failed: %v", err)
		}
		if count != 1 {
			t.Errorf("expected count=1 for id task-3, got %d", count)
		}
	})

	t.Run("count with combined filters", func(t *testing.T) {
		count, err := db.CountTasks(TaskFilter{Milestone: "m1", Status: "todo", Actor: "user-1"})
		if err != nil {
			t.Fatalf("CountTasks failed: %v", err)
		}
		if count != 1 {
			t.Errorf("expected count=1 for combined filters, got %d", count)
		}
	})

	t.Run("count with no matching results", func(t *testing.T) {
		count, err := db.CountTasks(TaskFilter{Milestone: "nonexistent"})
		if err != nil {
			t.Fatalf("CountTasks failed: %v", err)
		}
		if count != 0 {
			t.Errorf("expected count=0 for no matches, got %d", count)
		}
	})
}

// ===== Database-Level Unblock Validation Tests =====

func TestUnblockTask(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	t.Run("unblock blocked task succeeds and returns stored state", func(t *testing.T) {
		// Create a task and block it
		db.CreateTask(&Task{
			ID:     "unblock-success",
			Title:  "Blocked Task",
			Status: "blocked",
		})

		// Unblock without a new description
		task, err := db.UnblockTask("unblock-success", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// The returned task must be the freshly stored state
		if task.Status != "todo" {
			t.Errorf("expected status 'todo', got %s", task.Status)
		}
		if task.ID != "unblock-success" {
			t.Errorf("expected ID 'unblock-success', got %s", task.ID)
		}

		// Verify against the persisted record; LastUpdated must match exactly
		// because UnblockTask returns the freshly read task.
		updated, err := db.ReadTask("unblock-success")
		if err != nil {
			t.Fatalf("failed to read task: %v", err)
		}
		if updated.Status != "todo" {
			t.Errorf("expected status 'todo', got %s", updated.Status)
		}
		if !task.LastUpdated.Equal(updated.LastUpdated) {
			t.Errorf("expected returned LastUpdated %v to equal stored %v", task.LastUpdated, updated.LastUpdated)
		}
	})

	t.Run("unblock with new description overwrites description", func(t *testing.T) {
		db.CreateTask(&Task{
			ID:          "unblock-desc",
			Title:       "Task With Description",
			Status:      "blocked",
			Description: "Original description",
		})

		newDesc := "New description after unblock"
		task, err := db.UnblockTask("unblock-desc", newDesc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if task.Status != "todo" {
			t.Errorf("expected status 'todo', got %s", task.Status)
		}
		if task.Description != newDesc {
			t.Errorf("expected description %q, got %q", newDesc, task.Description)
		}

		updated, err := db.ReadTask("unblock-desc")
		if err != nil {
			t.Fatalf("failed to read task: %v", err)
		}
		if updated.Description != newDesc {
			t.Errorf("expected persisted description %q, got %q", newDesc, updated.Description)
		}
	})

	t.Run("unblock with empty-string description preserves description", func(t *testing.T) {
		db.CreateTask(&Task{
			ID:          "unblock-empty-desc",
			Title:       "Empty Description Test",
			Status:      "blocked",
			Description: "Preserve me",
		})

		task, err := db.UnblockTask("unblock-empty-desc", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if task.Description != "Preserve me" {
			t.Errorf("expected description to be preserved, got %q", task.Description)
		}
	})

	t.Run("unblock non-blocked task returns TaskNotBlockedError", func(t *testing.T) {
		// Create a task in todo status
		db.CreateTask(&Task{
			ID:     "unblock-not-blocked",
			Title:  "Not Blocked Task",
			Status: "todo",
		})

		_, err := db.UnblockTask("unblock-not-blocked", "")
		if err == nil {
			t.Fatal("expected error when unblocking a non-blocked task")
		}
		var notBlocked *TaskNotBlockedError
		if !errors.As(err, &notBlocked) {
			t.Errorf("expected TaskNotBlockedError, got %v", err)
		} else {
			if notBlocked.ID != "unblock-not-blocked" {
				t.Errorf("expected ID 'unblock-not-blocked', got %q", notBlocked.ID)
			}
			if notBlocked.Status != "todo" {
				t.Errorf("expected Status 'todo', got %q", notBlocked.Status)
			}
			wantMsg := `task "unblock-not-blocked" is in "todo" status, not blocked`
			if notBlocked.Error() != wantMsg {
				t.Errorf("expected message %q, got %q", wantMsg, notBlocked.Error())
			}
		}

		// Verify the task status was not changed
		updated, err := db.ReadTask("unblock-not-blocked")
		if err != nil {
			t.Fatalf("failed to read task: %v", err)
		}
		if updated.Status != "todo" {
			t.Errorf("expected status to remain 'todo', got %s", updated.Status)
		}
	})

	t.Run("unblock done task fails", func(t *testing.T) {
		db.CreateTask(&Task{
			ID:     "unblock-done",
			Title:  "Done Task",
			Status: "done",
		})

		_, err := db.UnblockTask("unblock-done", "")
		if err == nil {
			t.Fatal("expected error when unblocking a done task")
		}
		var notBlocked *TaskNotBlockedError
		if !errors.As(err, &notBlocked) {
			t.Errorf("expected TaskNotBlockedError, got %v", err)
		} else if notBlocked.Status != "done" {
			t.Errorf("expected Status 'done', got %q", notBlocked.Status)
		}

		updated, err := db.ReadTask("unblock-done")
		if err != nil {
			t.Fatalf("failed to read task: %v", err)
		}
		if updated.Status != "done" {
			t.Errorf("expected status to remain 'done', got %s", updated.Status)
		}
	})

	t.Run("unblock in_progress task fails", func(t *testing.T) {
		db.CreateTask(&Task{
			ID:     "unblock-inprogress",
			Title:  "In Progress Task",
			Status: "in_progress",
		})

		_, err := db.UnblockTask("unblock-inprogress", "")
		if err == nil {
			t.Fatal("expected error when unblocking an in_progress task")
		}
		var notBlocked *TaskNotBlockedError
		if !errors.As(err, &notBlocked) {
			t.Errorf("expected TaskNotBlockedError, got %v", err)
		} else if notBlocked.Status != "in_progress" {
			t.Errorf("expected Status 'in_progress', got %q", notBlocked.Status)
		}

		updated, err := db.ReadTask("unblock-inprogress")
		if err != nil {
			t.Fatalf("failed to read task: %v", err)
		}
		if updated.Status != "in_progress" {
			t.Errorf("expected status to remain 'in_progress', got %s", updated.Status)
		}
	})

	t.Run("unblock non-existent task returns TaskNotFoundError", func(t *testing.T) {
		_, err := db.UnblockTask("nonexistent-task", "")
		if err == nil {
			t.Fatal("expected error when unblocking a non-existent task")
		}
		assertTaskNotFound(t, err)
	})

	t.Run("unblock clears blocked_by to NULL", func(t *testing.T) {
		// Create a task with blocked_by set
		db.CreateTask(&Task{
			ID:        "unblock-clear-blockedby",
			Title:     "Blocked By Task",
			Status:    "blocked",
			BlockedBy: []string{"dep-1"},
		})

		if _, err := db.UnblockTask("unblock-clear-blockedby", ""); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Read the raw database record to verify blocked_by is NULL
		var blockedByStr *string
		err := db.conn.QueryRow("SELECT blocked_by FROM tasks WHERE id = ?", "unblock-clear-blockedby").Scan(&blockedByStr)
		if err != nil {
			t.Fatalf("failed to read raw blocked_by: %v", err)
		}
		if blockedByStr != nil {
			t.Errorf("expected blocked_by to be NULL, got %q", *blockedByStr)
		}
	})

	t.Run("unblock updates last_updated timestamp", func(t *testing.T) {
		db.CreateTask(&Task{
			ID:     "unblock-timestamp",
			Title:  "Timestamp Test",
			Status: "blocked",
		})

		// Read the original last_updated
		original, _ := db.ReadTask("unblock-timestamp")
		originalUpdated := original.LastUpdated

		// Wait to ensure time difference
		time.Sleep(10 * time.Millisecond)

		task, err := db.UnblockTask("unblock-timestamp", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Timestamps are stored at second granularity (RFC3339), so assert
		// non-decreasing rather than strictly increasing.
		if task.LastUpdated.Before(originalUpdated) {
			t.Errorf("expected last_updated to not go backwards: %v < %v", task.LastUpdated, originalUpdated)
		}
	})

	t.Run("unblock idempotency - second unblock fails", func(t *testing.T) {
		db.CreateTask(&Task{
			ID:     "unblock-idempotent",
			Title:  "Idempotent Test",
			Status: "blocked",
		})

		// First unblock should succeed
		if _, err := db.UnblockTask("unblock-idempotent", ""); err != nil {
			t.Fatalf("first unblock failed: %v", err)
		}

		// Second unblock should fail (task is now in todo status)
		_, err := db.UnblockTask("unblock-idempotent", "")
		if err == nil {
			t.Fatal("expected error on second unblock")
		}
		var notBlocked *TaskNotBlockedError
		if !errors.As(err, &notBlocked) {
			t.Errorf("expected TaskNotBlockedError, got %v", err)
		}

		// Verify status is still todo
		updated, _ := db.ReadTask("unblock-idempotent")
		if updated.Status != "todo" {
			t.Errorf("expected status to remain 'todo', got %s", updated.Status)
		}
	})

	t.Run("unblock with empty id fails", func(t *testing.T) {
		_, err := db.UnblockTask("", "")
		if err == nil {
			t.Fatal("expected error for empty id")
		}
		if err != ErrInvalidID {
			t.Errorf("expected ErrInvalidID, got %v", err)
		}
	})

	t.Run("unblock nil db fails", func(t *testing.T) {
		var nilDB *DB
		_, err := nilDB.UnblockTask("test", "")
		if err == nil {
			t.Fatal("expected error for nil db")
		}
		if err != ErrNilDB {
			t.Errorf("expected ErrNilDB, got %v", err)
		}
	})
}

// TestCorruptTimestampTask pins the scanTask parse-error behavior: a row
// whose created/last_updated columns are not RFC3339 must surface as an
// error from ReadTask, never as a zero-time task. SoftDeleteTask copies the
// row verbatim via INSERT..SELECT without a pre-read parse, so it succeeds
// and removes the row. The corrupt rows fill every nullable column so the
// row-level Scan succeeds and the failure lands in the timestamp/blocked_by
// parse, not the scan.
func TestCorruptTimestampTask(t *testing.T) {
	database := setupTestDB(t)
	defer teardownTestDB(t, database)

	_, err := database.conn.Exec(
		`INSERT INTO tasks (id, milestone, sprint, title, description, status, actor, created, last_updated)
		 VALUES ('corrupt-1', 'm', 's', 'Corrupt', 'd', 'todo', 'a', 'not-a-time', 'not-a-time')`)
	if err != nil {
		t.Fatalf("failed to insert corrupt row: %v", err)
	}

	if _, err := database.ReadTask("corrupt-1"); !strings.Contains(err.Error(), "parse created") {
		t.Errorf("ReadTask: expected a created-timestamp parse error, got %v", err)
	}
	if _, err := database.SoftDeleteTask("corrupt-1"); err != nil {
		t.Errorf("SoftDeleteTask: expected success (verbatim copy, no parse), got %v", err)
	}

	// Corrupt blocked_by with valid timestamps must also surface as an
	// error: only SQL NULL or the empty string mean "no blockers".
	_, err = database.conn.Exec(
		`INSERT INTO tasks (id, milestone, sprint, title, description, status, actor, blocked_by, created, last_updated)
		 VALUES ('corrupt-2', 'm', 's', 'Corrupt blocked_by', 'd', 'todo', 'a', 'not-json', ?, ?)`,
		time.Now().UTC().Format(time.RFC3339),
		time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("failed to insert corrupt blocked_by row: %v", err)
	}

	if _, err := database.ReadTask("corrupt-2"); !strings.Contains(err.Error(), "parse blocked_by") {
		t.Errorf("ReadTask: expected a blocked_by parse error, got %v", err)
	}
}
