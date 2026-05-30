package integration

import (
	"testing"
	"time"

	"github.com/rwbaskette/taskflow/internal/service"
)

// TestAddTaskWorkflow tests the full add task workflow
func TestAddTaskWorkflow(t *testing.T) {
	cfg := setupTestDB(t)
	defer teardownTestDB(t, cfg)

	// Test adding a task with all fields
	input := &service.AddTaskInput{
		ID:          "task-001",
		Milestone:   "v1.0",
		Title:       "Implement login feature",
		Description: "Add authentication and authorization",
		Actor:       "developer",
	}

	result, err := service.AddTask(cfg.DB, input)
	if err != nil {
		t.Fatalf("AddTask failed: %v", err)
	}

	// Verify result
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

	// Verify persistence - retrieve from database
	task, err := cfg.DB.ReadTask(input.ID)
	if err != nil {
		t.Fatalf("failed to read task from database: %v", err)
	}
	if task.ID != input.ID {
		t.Errorf("persisted task ID mismatch: %s != %s", task.ID, input.ID)
	}
	if task.Title != input.Title {
		t.Errorf("persisted task title mismatch: %s != %s", task.Title, input.Title)
	}
	if task.Status != "todo" {
		t.Errorf("persisted task status mismatch: %s != todo", task.Status)
	}
}

// TestUpdateTaskWorkflow tests the full update task workflow
func TestUpdateTaskWorkflow(t *testing.T) {
	cfg := setupTestDB(t)
	defer teardownTestDB(t, cfg)

	// First, create a task
	input := &service.AddTaskInput{
		ID:          "task-002",
		Milestone:   "v1.0",
		Title:       "Original Title",
		Description: "Original Description",
		Actor:       "developer",
	}

	_, err := service.AddTask(cfg.DB, input)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Update the task
	updateInput := &service.UpdateTaskInput{
		ID:          "task-002",
		Title:       "Updated Title",
		Description: "Updated Description",
	}

	result, err := service.UpdateTask(cfg.DB, updateInput)
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	// Verify result
	if result.Title != "Updated Title" {
		t.Errorf("expected updated title, got %s", result.Title)
	}
	if result.Description != "Updated Description" {
		t.Errorf("expected updated description, got %s", result.Description)
	}

	// Verify persistence
	task, err := cfg.DB.ReadTask("task-002")
	if err != nil {
		t.Fatalf("failed to read task: %v", err)
	}
	if task.Title != "Updated Title" {
		t.Errorf("persisted title mismatch: %s", task.Title)
	}
}

// TestCompleteTaskWorkflow tests completing a task
func TestCompleteTaskWorkflow(t *testing.T) {
	cfg := setupTestDB(t)
	defer teardownTestDB(t, cfg)

	// Create a task
	input := &service.AddTaskInput{
		ID:          "task-003",
		Milestone:   "v1.0",
		Title:       "Task to complete",
		Description: "Will be completed",
		Actor:       "developer",
	}

	_, err := service.AddTask(cfg.DB, input)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Complete the task
	result, err := service.CompleteTask(cfg.DB, &service.CompleteTaskInput{ID: "task-003"})
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	// Verify status changed to done
	if result.Status != "done" {
		t.Errorf("expected status 'done', got %s", result.Status)
	}

	// Verify persistence
	task, err := cfg.DB.ReadTask("task-003")
	if err != nil {
		t.Fatalf("failed to read task: %v", err)
	}
	if task.Status != "done" {
		t.Errorf("persisted status mismatch: %s", task.Status)
	}
}

// TestBlockTaskWorkflow tests blocking a task
func TestBlockTaskWorkflow(t *testing.T) {
	cfg := setupTestDB(t)
	defer teardownTestDB(t, cfg)

	// Create a task
	input := &service.AddTaskInput{
		ID:          "task-004",
		Milestone:   "v1.0",
		Title:       "Task to block",
		Description: "Will be blocked",
		Actor:       "developer",
	}

	_, err := service.AddTask(cfg.DB, input)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Block the task
	result, err := service.BlockTask(cfg.DB, service.BlockTaskInput{
		ID:     "task-004",
		Reason: "Waiting for dependency",
	})
	if err != nil {
		t.Fatalf("BlockTask failed: %v", err)
	}

	// Verify status changed to blocked
	if result.Status != "blocked" {
		t.Errorf("expected status 'blocked', got %s", result.Status)
	}

	// Verify persistence
	task, err := cfg.DB.ReadTask("task-004")
	if err != nil {
		t.Fatalf("failed to read task: %v", err)
	}
	if task.Status != "blocked" {
		t.Errorf("persisted status mismatch: %s", task.Status)
	}
}

// TestListTasksWorkflow tests listing tasks with filters
func TestListTasksWorkflow(t *testing.T) {
	cfg := setupTestDB(t)
	defer teardownTestDB(t, cfg)

	// Create multiple tasks
	tasks := []struct {
		id        string
		title     string
		milestone string
		actor     string
		status    string
	}{
		{"task-101", "Task 1", "v1.0", "dev1", "todo"},
		{"task-102", "Task 2", "v1.0", "dev2", "in_progress"},
		{"task-103", "Task 3", "v2.0", "dev1", "done"},
		{"task-104", "Task 4", "v2.0", "dev2", "blocked"},
	}

	for _, task := range tasks {
		input := &service.AddTaskInput{
			ID:          task.id,
			Milestone:   task.milestone,
			Title:       task.title,
			Description: "Description for " + task.title,
			Actor:       task.actor,
		}
		_, err := service.AddTask(cfg.DB, input)
		if err != nil {
			t.Fatalf("failed to create task %s: %v", task.id, err)
		}

		// Update status if needed
		if task.status != "todo" {
			updateInput := &service.UpdateTaskInput{
				ID:     task.id,
				Status: task.status,
			}
			_, err = service.UpdateTask(cfg.DB, updateInput)
			if err != nil {
				t.Fatalf("failed to update task %s: %v", task.id, err)
			}
		}
	}

	// Test list all tasks
	listSvc := service.NewListService(cfg.DB)
	result, err := listSvc.ListTasks(&service.ListTaskFilter{})
	if err != nil {
		t.Fatalf("ListTasks failed: %v", err)
	}
	if len(result.Tasks) != 4 {
		t.Errorf("expected 4 tasks, got %d", len(result.Tasks))
	}

	// Test list with milestone filter
	result, err = listSvc.ListTasks(&service.ListTaskFilter{
		Milestone: "v1.0",
	})
	if err != nil {
		t.Fatalf("ListTasks with filter failed: %v", err)
	}
	if len(result.Tasks) != 2 {
		t.Errorf("expected 2 tasks for milestone v1.0, got %d", len(result.Tasks))
	}

	// Test list with status filter
	result, err = listSvc.ListTasks(&service.ListTaskFilter{
		Status: "done",
	})
	if err != nil {
		t.Fatalf("ListTasks with status filter failed: %v", err)
	}
	if len(result.Tasks) != 1 {
		t.Errorf("expected 1 done task, got %d", len(result.Tasks))
	}

	// Test list with actor filter
	result, err = listSvc.ListTasks(&service.ListTaskFilter{
		Actor: "dev1",
	})
	if err != nil {
		t.Fatalf("ListTasks with actor filter failed: %v", err)
	}
	if len(result.Tasks) != 2 {
		t.Errorf("expected 2 tasks for actor dev1, got %d", len(result.Tasks))
	}
}

// TestResetTimedOutWorkflow tests resetting timed out tasks
func TestResetTimedOutWorkflow(t *testing.T) {
	cfg := setupTestDB(t)
	defer teardownTestDB(t, cfg)

	// Create and start a task
	input := &service.AddTaskInput{
		ID:          "task-005",
		Milestone:   "v1.0",
		Title:       "Task to timeout",
		Description: "Will timeout and be reset",
		Actor:       "developer",
	}

	_, err := service.AddTask(cfg.DB, input)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Set task to in_progress
	updateInput := &service.UpdateTaskInput{
		ID:     "task-005",
		Status: "in_progress",
	}
	_, err = service.UpdateTask(cfg.DB, updateInput)
	if err != nil {
		t.Fatalf("failed to set task to in_progress: %v", err)
	}

	// Reset tasks with invalid timeout (0 should fail validation)
	_, err = service.ResetTimedOut(cfg.DB, service.ResetTimedOutInput{
		TimeoutMinutes: 0,
	})
	if err == nil {
		t.Fatal("expected error for zero timeout")
	}

	// Reset tasks with large timeout (no tasks should exceed it)
	result, err := service.ResetTimedOut(cfg.DB, service.ResetTimedOutInput{
		TimeoutMinutes: 1000,
	})
	if err != nil {
		t.Fatalf("ResetTimedOut failed: %v", err)
	}

	// With large timeout, no tasks should be reset (no time has actually passed)
	if len(result.ResetTasks) != 0 {
		t.Errorf("expected 0 tasks reset with large timeout, got %d", len(result.ResetTasks))
	}

	// Verify status is still in_progress
	task, err := cfg.DB.ReadTask("task-005")
	if err != nil {
		t.Fatalf("failed to read task: %v", err)
	}
	if task.Status != "in_progress" {
		t.Errorf("expected status 'in_progress', got %s", task.Status)
	}
}

// TestMultiCommandSequence tests a sequence of commands: add -> update -> complete
func TestMultiCommandSequence(t *testing.T) {
	cfg := setupTestDB(t)
	defer teardownTestDB(t, cfg)

	// Step 1: Add a task
	addInput := &service.AddTaskInput{
		ID:          "seq-001",
		Milestone:   "v1.0",
		Title:       "Sequential Task",
		Description: "Initial description",
		Actor:       "developer",
	}

	addResult, err := service.AddTask(cfg.DB, addInput)
	if err != nil {
		t.Fatalf("AddTask failed: %v", err)
	}
	initialTime := addResult.LastUpdated

	// Small delay to ensure timestamp changes
	time.Sleep(time.Millisecond)
	// Step 2: Update the task
	updateInput := &service.UpdateTaskInput{
		ID:          "seq-001",
		Title:       "Updated Sequential Task",
		Description: "Updated description",
		Milestone:   "v2.0",
	}

	updateResult, err := service.UpdateTask(cfg.DB, updateInput)
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}
	updatedTime := updateResult.LastUpdated

	// Verify timestamp was updated
	if !updatedTime.After(initialTime) {
		t.Errorf("expected updated timestamp after initial time")
	}

	// Small delay to ensure timestamp changes before complete
	time.Sleep(time.Millisecond)

	// Step 3: Complete the task
	completeResult, err := service.CompleteTask(cfg.DB, &service.CompleteTaskInput{ID: "seq-001"})
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	// Step 4: Verify final state in database
	task, err := cfg.DB.ReadTask("seq-001")
	if err != nil {
		t.Fatalf("failed to read task: %v", err)
	}

	// Verify all fields are correct
	if task.Title != "Updated Sequential Task" {
		t.Errorf("expected title 'Updated Sequential Task', got %s", task.Title)
	}
	if task.Description != "Updated description" {
		t.Errorf("expected description 'Updated description', got %s", task.Description)
	}
	if task.Milestone != "v2.0" {
		t.Errorf("expected milestone 'v2.0', got %s", task.Milestone)
	}
	if task.Status != "done" {
		t.Errorf("expected status 'done', got %s", task.Status)
	}

	// Verify LastUpdated was updated on complete
	finalTime := completeResult.LastUpdated
	if !finalTime.After(updatedTime) {
		t.Errorf("expected final timestamp after updated time")
	}
}

// TestMultiCommandSequenceWithBlockAndReset tests add -> block -> complete
func TestMultiCommandSequenceWithBlockAndReset(t *testing.T) {
	cfg := setupTestDB(t)
	defer teardownTestDB(t, cfg)

	// Step 1: Add task
	addInput := &service.AddTaskInput{
		ID:          "seq-002",
		Milestone:   "v1.0",
		Title:       "Complex Sequence Task",
		Description: "Will go through multiple states",
		Actor:       "developer",
	}

	_, err := service.AddTask(cfg.DB, addInput)
	if err != nil {
		t.Fatalf("AddTask failed: %v", err)
	}

	// Step 2: Block the task
	_, err = service.BlockTask(cfg.DB, service.BlockTaskInput{
		ID:     "seq-002",
		Reason: "Waiting for API",
	})
	if err != nil {
		t.Fatalf("BlockTask failed: %v", err)
	}

	// Verify blocked status
	task, err := cfg.DB.ReadTask("seq-002")
	if err != nil {
		t.Fatalf("failed to read task: %v", err)
	}
	if task.Status != "blocked" {
		t.Errorf("expected status 'blocked', got %s", task.Status)
	}

	// Step 3: Complete the blocked task (should work)
	_, err = service.CompleteTask(cfg.DB, &service.CompleteTaskInput{ID: "seq-002"})
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	// Verify final status
	task, err = cfg.DB.ReadTask("seq-002")
	if err != nil {
		t.Fatalf("failed to read task: %v", err)
	}
	if task.Status != "done" {
		t.Errorf("expected final status 'done', got %s", task.Status)
	}
}

// TestDatabasePersistence tests that data persists correctly across operations
func TestDatabasePersistence(t *testing.T) {
	cfg := setupTestDB(t)
	defer teardownTestDB(t, cfg)

	// Create task with all fields
	input := &service.AddTaskInput{
		ID:          "persist-001",
		Milestone:   "v1.0",
		Title:       "Persistent Task",
		Description: "This task should persist",
		Actor:       "testuser",
	}

	_, err := service.AddTask(cfg.DB, input)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Verify task exists
	_, err = cfg.DB.ReadTask("persist-001")
	if err != nil {
		t.Fatalf("task should exist in database")
	}

	// Retrieve and verify all fields
	task, err := cfg.DB.ReadTask("persist-001")
	if err != nil {
		t.Fatalf("failed to read task: %v", err)
	}

	// Check all fields
	if task.ID != input.ID {
		t.Errorf("ID mismatch: %s != %s", task.ID, input.ID)
	}
	if task.Title != input.Title {
		t.Errorf("Title mismatch: %s != %s", task.Title, input.Title)
	}
	if task.Description != input.Description {
		t.Errorf("Description mismatch: %s != %s", task.Description, input.Description)
	}
	if task.Milestone != input.Milestone {
		t.Errorf("Milestone mismatch: %s != %s", task.Milestone, input.Milestone)
	}
	if task.Actor != input.Actor {
		t.Errorf("Actor mismatch: %s != %s", task.Actor, input.Actor)
	}
	if task.Status != "todo" {
		t.Errorf("Status mismatch: %s != todo", task.Status)
	}
	if task.LastUpdated.IsZero() {
		t.Error("LastUpdated should not be zero")
	}

	// Perform multiple updates and verify persistence
	for i := 0; i < 3; i++ {
		updateInput := &service.UpdateTaskInput{
			ID:          "persist-001",
			Description: "Updated description " + string(rune('0'+i)),
		}
		_, err = service.UpdateTask(cfg.DB, updateInput)
		if err != nil {
			t.Fatalf("update %d failed: %v", i, err)
		}
	}

	// Final verification
	task, err = cfg.DB.ReadTask("persist-001")
	if err != nil {
		t.Fatalf("failed to read task after updates: %v", err)
	}
	if task.Description != "Updated description 2" {
		t.Errorf("final description mismatch: %s", task.Description)
	}
}

// TestCleanupBetweenTests ensures test isolation
func TestCleanupBetweenTests(t *testing.T) {
	cfg := setupTestDB(t)
	defer teardownTestDB(t, cfg)

	// Create a task
	input := &service.AddTaskInput{
		ID:          "cleanup-test",
		Milestone:   "v1.0",
		Title:       "Cleanup Test Task",
		Description: "Testing cleanup",
		Actor:       "tester",
	}

	_, err := service.AddTask(cfg.DB, input)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Delete the task
	err = cfg.DB.DeleteTask("cleanup-test")
	if err != nil {
		t.Fatalf("failed to delete task: %v", err)
	}

	// Verify task no longer exists
	_, err = cfg.DB.ReadTask("cleanup-test")
	if err == nil {
		t.Error("task should not exist after deletion")
	}
}

// TestPartialUpdatePersistence tests that partial updates preserve other fields
func TestPartialUpdatePersistence(t *testing.T) {
	cfg := setupTestDB(t)
	defer teardownTestDB(t, cfg)

	// Create a task with all fields
	input := &service.AddTaskInput{
		ID:          "partial-001",
		Milestone:   "v1.0",
		Title:       "Original Title",
		Description: "Original Description",
		Actor:       "developer",
	}

	_, err := service.AddTask(cfg.DB, input)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Update only the title (partial update)
	updateInput := &service.UpdateTaskInput{
		ID:    "partial-001",
		Title: "New Title Only",
	}

	_, err = service.UpdateTask(cfg.DB, updateInput)
	if err != nil {
		t.Fatalf("partial update failed: %v", err)
	}

	// Verify other fields are preserved
	task, err := cfg.DB.ReadTask("partial-001")
	if err != nil {
		t.Fatalf("failed to read task: %v", err)
	}

	if task.Title != "New Title Only" {
		t.Errorf("title not updated: %s", task.Title)
	}
	if task.Description != "Original Description" {
		t.Errorf("description was changed: %s", task.Description)
	}
	if task.Milestone != "v1.0" {
		t.Errorf("milestone was changed: %s", task.Milestone)
	}
	if task.Actor != "developer" {
		t.Errorf("actor was changed: %s", task.Actor)
	}
}

// TestListTasksPagination tests pagination of list operation
func TestListTasksPagination(t *testing.T) {
	cfg := setupTestDB(t)
	defer teardownTestDB(t, cfg)

	// Create 10 tasks
	for i := 0; i < 10; i++ {
		id := string(rune('0' + i%10))
		if i >= 10 {
			id = "page-task-" + string(rune('0'+(i-10)))
		} else {
			id = "page-task-" + id
		}
		input := &service.AddTaskInput{
			ID:          id,
			Milestone:   "v1.0",
			Title:       "Task " + id,
			Description: "Task description " + id,
			Actor:       "developer",
		}
		_, err := service.AddTask(cfg.DB, input)
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
	}

	listSvc := service.NewListService(cfg.DB)

	// Test limit
	result, err := listSvc.ListTasks(&service.ListTaskFilter{
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("ListTasks with limit failed: %v", err)
	}
	if len(result.Tasks) != 5 {
		t.Errorf("expected 5 tasks with limit, got %d", len(result.Tasks))
	}
	if !result.HasMore {
		t.Error("expected HasMore to be true when more tasks exist")
	}

	// Test offset
	result, err = listSvc.ListTasks(&service.ListTaskFilter{
		Limit:  5,
		Offset: 5,
	})
	if err != nil {
		t.Fatalf("ListTasks with offset failed: %v", err)
	}
	if len(result.Tasks) != 5 {
		t.Errorf("expected 5 tasks with offset 5, got %d", len(result.Tasks))
	}
}

// TestErrorHandling tests error handling for non-existent tasks
func TestErrorHandling(t *testing.T) {
	cfg := setupTestDB(t)
	defer teardownTestDB(t, cfg)

	// Test reading non-existent task
	_, err := cfg.DB.ReadTask("non-existent")
	if err == nil {
		t.Error("expected error for non-existent task")
	}

	// Test updating non-existent task
	updateInput := &service.UpdateTaskInput{
		ID:    "non-existent",
		Title: "Won't work",
	}
	_, err = service.UpdateTask(cfg.DB, updateInput)
	if err == nil {
		t.Error("expected error for updating non-existent task")
	}

	// Test completing non-existent task
	_, err = service.CompleteTask(cfg.DB, &service.CompleteTaskInput{ID: "non-existent"})
	if err == nil {
		t.Error("expected error for completing non-existent task")
	}

	// Test blocking non-existent task
	_, err = service.BlockTask(cfg.DB, service.BlockTaskInput{
		ID:     "non-existent",
		Reason: "Won't work",
	})
	if err == nil {
		t.Error("expected error for blocking non-existent task")
	}

	// Test deleting non-existent task
	err = cfg.DB.DeleteTask("non-existent")
	if err == nil {
		t.Error("expected error for deleting non-existent task")
	}
}

// TestDuplicateID tests that duplicate IDs are properly rejected
func TestDuplicateID(t *testing.T) {
	cfg := setupTestDB(t)
	defer teardownTestDB(t, cfg)

	// Create first task
	input1 := &service.AddTaskInput{
		ID:          "duplicate-test",
		Milestone:   "v1.0",
		Title:       "First Task",
		Description: "First description",
		Actor:       "developer",
	}

	_, err := service.AddTask(cfg.DB, input1)
	if err != nil {
		t.Fatalf("failed to create first task: %v", err)
	}

	// Try to create duplicate
	input2 := &service.AddTaskInput{
		ID:          "duplicate-test",
		Milestone:   "v1.0",
		Title:       "Second Task",
		Description: "Second description",
		Actor:       "developer2",
	}

	_, err = service.AddTask(cfg.DB, input2)
	if err == nil {
		t.Error("expected error for duplicate ID")
	}

	// Verify only one task exists
	count := countTasks(t, cfg)
	if count != 1 {
		t.Errorf("expected 1 task, got %d", count)
	}
}

// TestStatusTransitions tests valid status transitions
func TestStatusTransitions(t *testing.T) {
	cfg := setupTestDB(t)
	defer teardownTestDB(t, cfg)

	// Create task (starts as todo)
	input := &service.AddTaskInput{
		ID:          "status-001",
		Milestone:   "v1.0",
		Title:       "Status Test Task",
		Description: "Testing status transitions",
		Actor:       "developer",
	}

	_, err := service.AddTask(cfg.DB, input)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Verify initial status
	task, err := cfg.DB.ReadTask("status-001")
	if err != nil {
		t.Fatalf("failed to read task: %v", err)
	}
	if task.Status != "todo" {
		t.Errorf("initial status should be 'todo', got %s", task.Status)
	}

	// Transition: todo -> in_progress
	updateInput := &service.UpdateTaskInput{
		ID:     "status-001",
		Status: "in_progress",
	}
	_, err = service.UpdateTask(cfg.DB, updateInput)
	if err != nil {
		t.Fatalf("failed to update status: %v", err)
	}

	task, err = cfg.DB.ReadTask("status-001")
	if err != nil {
		t.Fatalf("failed to read task: %v", err)
	}
	if task.Status != "in_progress" {
		t.Errorf("status should be 'in_progress', got %s", task.Status)
	}

	// Transition: in_progress -> done
	_, err = service.CompleteTask(cfg.DB, &service.CompleteTaskInput{ID: "status-001"})
	if err != nil {
		t.Fatalf("failed to complete task: %v", err)
	}

	task, err = cfg.DB.ReadTask("status-001")
	if err != nil {
		t.Fatalf("failed to read task: %v", err)
	}
	if task.Status != "done" {
		t.Errorf("status should be 'done', got %s", task.Status)
	}
}

// TestBlockAppendsReasonToDescription tests that blocking appends reason to description
func TestBlockAppendsReasonToDescription(t *testing.T) {
	cfg := setupTestDB(t)
	defer teardownTestDB(t, cfg)

	// Create a task with a description
	input := &service.AddTaskInput{
		ID:          "block-desc-001",
		Milestone:   "v1.0",
		Title:       "Task to block",
		Description: "Original description",
		Actor:       "developer",
	}

	_, err := service.AddTask(cfg.DB, input)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Block the task with a reason
	blockReason := "Waiting for dependency"
	_, err = service.BlockTask(cfg.DB, service.BlockTaskInput{
		ID:     "block-desc-001",
		Reason: blockReason,
	})
	if err != nil {
		t.Fatalf("BlockTask failed: %v", err)
	}

	// Verify description was updated with block reason
	task, err := cfg.DB.ReadTask("block-desc-001")
	if err != nil {
		t.Fatalf("failed to read task: %v", err)
	}

	expectedDesc := "Original description\n[BLOCKED: " + blockReason + "]"
	if task.Description != expectedDesc {
		t.Errorf("expected description '%s', got '%s'", expectedDesc, task.Description)
	}
}

// TestMultipleTasksInSequence tests adding, updating, and completing multiple tasks
func TestMultipleTasksInSequence(t *testing.T) {
	cfg := setupTestDB(t)
	defer teardownTestDB(t, cfg)

	// Create 5 tasks
	for i := 1; i <= 5; i++ {
		id := string(rune('0' + i))
		input := &service.AddTaskInput{
			ID:          "multi-" + id,
			Milestone:   "v1.0",
			Title:       "Task " + id,
			Description: "Description " + id,
			Actor:       "developer",
		}

		_, err := service.AddTask(cfg.DB, input)
		if err != nil {
			t.Fatalf("failed to create task %d: %v", i, err)
		}
	}

	// Complete tasks 1, 3, 5
	for _, i := range []int{1, 3, 5} {
		id := string(rune('0' + i))
		_, err := service.CompleteTask(cfg.DB, &service.CompleteTaskInput{ID: "multi-" + id})
		if err != nil {
			t.Fatalf("failed to complete task %d: %v", i, err)
		}
	}

	// Verify counts
	listSvc := service.NewListService(cfg.DB)

	allResult, err := listSvc.ListTasks(&service.ListTaskFilter{})
	if err != nil {
		t.Fatalf("ListTasks failed: %v", err)
	}
	if len(allResult.Tasks) != 5 {
		t.Errorf("expected 5 total tasks, got %d", len(allResult.Tasks))
	}

	todoResult, err := listSvc.ListTasks(&service.ListTaskFilter{Status: "todo"})
	if err != nil {
		t.Fatalf("ListTasks with status filter failed: %v", err)
	}
	if len(todoResult.Tasks) != 2 {
		t.Errorf("expected 2 todo tasks, got %d", len(todoResult.Tasks))
	}

	doneResult, err := listSvc.ListTasks(&service.ListTaskFilter{Status: "done"})
	if err != nil {
		t.Fatalf("ListTasks with status filter failed: %v", err)
	}
	if len(doneResult.Tasks) != 3 {
		t.Errorf("expected 3 done tasks, got %d", len(doneResult.Tasks))
	}
}

// TestListTasksSortBy tests sorting of tasks by various fields
func TestListTasksSortBy(t *testing.T) {
	cfg := setupTestDB(t)
	defer teardownTestDB(t, cfg)

	tasks := []struct {
		id          string
		title       string
		actor       string
		milestone   string
		status      string
		description string
	}{
		{"sort-001", "Alpha Task", "alice", "v1.0", "todo", "Description A"},
		{"sort-002", "Beta Task", "bob", "v2.0", "in_progress", "Description B"},
		{"sort-003", "Gamma Task", "alice", "v1.0", "done", "Description C"},
		{"sort-004", "Delta Task", "charlie", "v2.0", "blocked", "Description D"},
	}

	for _, task := range tasks {
		input := &service.AddTaskInput{
			ID:          task.id,
			Title:       task.title,
			Actor:       task.actor,
			Milestone:   task.milestone,
			Description: task.description,
		}
		_, err := service.AddTask(cfg.DB, input)
		if err != nil {
			t.Fatalf("failed to create task %s: %v", task.id, err)
		}

		if task.status != "todo" {
			updateInput := &service.UpdateTaskInput{
				ID:     task.id,
				Status: task.status,
			}
			_, err = service.UpdateTask(cfg.DB, updateInput)
			if err != nil {
				t.Fatalf("failed to update task %s status: %v", task.id, err)
			}
		}
	}

	listSvc := service.NewListService(cfg.DB)

	// Test sort by title
	result, err := listSvc.ListTasks(&service.ListTaskFilter{SortBy: "title"})
	if err != nil {
		t.Fatalf("ListTasks sort by title failed: %v", err)
	}
	if len(result.Tasks) != 4 {
		t.Fatalf("expected 4 tasks, got %d", len(result.Tasks))
	}
	if result.Tasks[0].Title != "Alpha Task" {
		t.Errorf("expected first task title 'Alpha Task', got '%s'", result.Tasks[0].Title)
	}
	if result.Tasks[3].Title != "Gamma Task" {
		t.Errorf("expected last task title 'Gamma Task', got '%s'", result.Tasks[3].Title)
	}

	// Test sort by actor
	result, err = listSvc.ListTasks(&service.ListTaskFilter{SortBy: "actor"})
	if err != nil {
		t.Fatalf("ListTasks sort by actor failed: %v", err)
	}
	if result.Tasks[0].Actor != "alice" {
		t.Errorf("expected first task actor 'alice', got '%s'", result.Tasks[0].Actor)
	}

	// Test sort by status
	result, err = listSvc.ListTasks(&service.ListTaskFilter{SortBy: "status"})
	if err != nil {
		t.Fatalf("ListTasks sort by status failed: %v", err)
	}
	if len(result.Tasks) != 4 {
		t.Fatalf("expected 4 tasks, got %d", len(result.Tasks))
	}

	// Test sort by milestone
	result, err = listSvc.ListTasks(&service.ListTaskFilter{SortBy: "milestone"})
	if err != nil {
		t.Fatalf("ListTasks sort by milestone failed: %v", err)
	}
	if len(result.Tasks) != 4 {
		t.Fatalf("expected 4 tasks, got %d", len(result.Tasks))
	}

	// Test sort by id
	result, err = listSvc.ListTasks(&service.ListTaskFilter{SortBy: "id"})
	if err != nil {
		t.Fatalf("ListTasks sort by id failed: %v", err)
	}
	if result.Tasks[0].ID != "sort-001" {
		t.Errorf("expected first task id 'sort-001', got '%s'", result.Tasks[0].ID)
	}
	if result.Tasks[3].ID != "sort-004" {
		t.Errorf("expected last task id 'sort-004', got '%s'", result.Tasks[3].ID)
	}

	// Test sort by description
	result, err = listSvc.ListTasks(&service.ListTaskFilter{SortBy: "description"})
	if err != nil {
		t.Fatalf("ListTasks sort by description failed: %v", err)
	}
	if len(result.Tasks) != 4 {
		t.Fatalf("expected 4 tasks, got %d", len(result.Tasks))
	}
}
