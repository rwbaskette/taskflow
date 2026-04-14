package service

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/user/project/internal/db"
)

func TestListTasks_WithValidFilters(t *testing.T) {
	// Create a temporary database file for testing
	tmpFile, err := os.CreateTemp("", "test-db-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	database, err := db.NewDB(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer database.Close()

	// Insert test data
	testTasks := []db.Task{
		{
			ID:          "TASK-001",
			Milestone:   "v1.0",
			Title:       "Implement login",
			Description: "Add authentication system",
			Status:      "done",
			Actor:       "alice",
			LastUpdated: time.Now(),
		},
		{
			ID:          "TASK-002",
			Milestone:   "v1.0",
			Title:       "Fix bugs",
			Description: "Fix critical bugs",
			Status:      "in_progress",
			Actor:       "bob",
			LastUpdated: time.Now(),
		},
		{
			ID:          "TASK-003",
			Milestone:   "v2.0",
			Title:       "Add new feature",
			Description: "New feature implementation",
			Status:      "todo",
			Actor:       "alice",
			LastUpdated: time.Now(),
		},
		{
			ID:          "TASK-004",
			Milestone:   "v1.0",
			Title:       "Code review",
			Description: "Review pull requests",
			Status:      "blocked",
			Actor:       "charlie",
			LastUpdated: time.Now(),
		},
	}

	for _, task := range testTasks {
		if err := database.CreateTask(&task); err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}
	}

	service := NewListService(database)

	// Test 1: List all tasks (no filter)
	t.Run("ListAllTasks", func(t *testing.T) {
		result, err := service.ListTasks(&ListTaskFilter{})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if len(result.Tasks) != 4 {
			t.Errorf("Expected 4 tasks, got %d", len(result.Tasks))
		}
		if result.Total != 4 {
			t.Errorf("Expected Total=4, got %d", result.Total)
		}
	})

	// Test 2: Filter by milestone
	t.Run("FilterByMilestone", func(t *testing.T) {
		result, err := service.ListTasks(&ListTaskFilter{Milestone: "v1.0"})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if len(result.Tasks) != 3 {
			t.Errorf("Expected 3 tasks for milestone v1.0, got %d", len(result.Tasks))
		}
		for _, task := range result.Tasks {
			if task.Milestone != "v1.0" {
				t.Errorf("Expected milestone v1.0, got %s", task.Milestone)
			}
		}
	})

	// Test 3: Filter by status
	t.Run("FilterByStatus", func(t *testing.T) {
		result, err := service.ListTasks(&ListTaskFilter{Status: "done"})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if len(result.Tasks) != 1 {
			t.Errorf("Expected 1 task with status done, got %d", len(result.Tasks))
		}
		if result.Tasks[0].Status != "done" {
			t.Errorf("Expected status done, got %s", result.Tasks[0].Status)
		}
	})

	// Test 4: Filter by actor
	t.Run("FilterByActor", func(t *testing.T) {
		result, err := service.ListTasks(&ListTaskFilter{Actor: "alice"})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if len(result.Tasks) != 2 {
			t.Errorf("Expected 2 tasks for actor alice, got %d", len(result.Tasks))
		}
		for _, task := range result.Tasks {
			if task.Actor != "alice" {
				t.Errorf("Expected actor alice, got %s", task.Actor)
			}
		}
	})

	// Test 5: Combined filters
	t.Run("CombinedFilters", func(t *testing.T) {
		result, err := service.ListTasks(&ListTaskFilter{
			Milestone: "v1.0",
			Status:    "done",
		})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if len(result.Tasks) != 1 {
			t.Errorf("Expected 1 task, got %d", len(result.Tasks))
		}
		if result.Tasks[0].ID != "TASK-001" {
			t.Errorf("Expected TASK-001, got %s", result.Tasks[0].ID)
		}
	})

	// Test 6: Pagination - limit
	t.Run("PaginationLimit", func(t *testing.T) {
		result, err := service.ListTasks(&ListTaskFilter{Limit: 2})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if len(result.Tasks) != 2 {
			t.Errorf("Expected 2 tasks, got %d", len(result.Tasks))
		}
		if !result.HasMore {
			t.Error("Expected HasMore=true when more tasks exist")
		}
		if result.Limit != 2 {
			t.Errorf("Expected Limit=2, got %d", result.Limit)
		}
	})

	// Test 7: Pagination - offset
	t.Run("PaginationOffset", func(t *testing.T) {
		result, err := service.ListTasks(&ListTaskFilter{Offset: 2})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		// With offset 2 and no limit, should get 2 remaining tasks
		if len(result.Tasks) != 2 {
			t.Errorf("Expected 2 tasks with offset 2, got %d", len(result.Tasks))
		}
	})

	// Test 8: Nil filter (should work with defaults)
	t.Run("NilFilter", func(t *testing.T) {
		result, err := service.ListTasks(nil)
		if err != nil {
			t.Fatalf("ListTasks with nil filter failed: %v", err)
		}
		if len(result.Tasks) != 4 {
			t.Errorf("Expected 4 tasks with nil filter, got %d", len(result.Tasks))
		}
	})
}

func TestListTasks_WithNilDatabase(t *testing.T) {
	service := NewListService(nil)

	result, err := service.ListTasks(&ListTaskFilter{})
	if err == nil {
		t.Error("Expected error with nil database, got nil")
	}
	if !errors.Is(err, ErrNilDatabase) {
		t.Errorf("Expected ErrNilDatabase, got %v", err)
	}
	if result != nil {
		t.Error("Expected nil result with nil database")
	}
}

func TestListTasks_WithNilFilter(t *testing.T) {
	// Create a temporary database file for testing
	tmpFile, err := os.CreateTemp("", "test-db-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	database, err := db.NewDB(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer database.Close()

	// Insert a test task
	task := db.Task{
		ID:          "TASK-001",
		Title:       "Test task",
		Status:      "todo",
		LastUpdated: time.Now(),
	}
	if err := database.CreateTask(&task); err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	service := NewListService(database)

	// Test with nil filter
	result, err := service.ListTasks(nil)
	if err != nil {
		t.Fatalf("ListTasks with nil filter failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if len(result.Tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(result.Tasks))
	}
}

func TestGetFilteredCount_Basic(t *testing.T) {
	// Create a temporary database file for testing
	tmpFile, err := os.CreateTemp("", "test-db-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	database, err := db.NewDB(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer database.Close()

	// Insert test tasks
	testTasks := []db.Task{
		{
			ID:          "TASK-001",
			Milestone:   "v1.0",
			Title:       "Task 1",
			Status:      "done",
			Actor:       "alice",
			LastUpdated: time.Now(),
		},
		{
			ID:          "TASK-002",
			Milestone:   "v1.0",
			Title:       "Task 2",
			Status:      "in_progress",
			Actor:       "bob",
			LastUpdated: time.Now(),
		},
		{
			ID:          "TASK-003",
			Milestone:   "v2.0",
			Title:       "Task 3",
			Status:      "todo",
			LastUpdated: time.Now(),
		},
	}

	for _, task := range testTasks {
		if err := database.CreateTask(&task); err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}
	}

	service := NewListService(database)

	// Test 1: Count all tasks
	t.Run("CountAll", func(t *testing.T) {
		count, err := service.GetFilteredCount(&ListTaskFilter{})
		if err != nil {
			t.Fatalf("GetFilteredCount failed: %v", err)
		}
		if count != 3 {
			t.Errorf("Expected count=3, got %d", count)
		}
	})

	// Test 2: Count with milestone filter
	t.Run("CountWithMilestone", func(t *testing.T) {
		count, err := service.GetFilteredCount(&ListTaskFilter{Milestone: "v1.0"})
		if err != nil {
			t.Fatalf("GetFilteredCount failed: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count=2 for milestone v1.0, got %d", count)
		}
	})

	// Test 3: Count with status filter
	t.Run("CountWithStatus", func(t *testing.T) {
		count, err := service.GetFilteredCount(&ListTaskFilter{Status: "done"})
		if err != nil {
			t.Fatalf("GetFilteredCount failed: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected count=1 for status done, got %d", count)
		}
	})

	// Test 4: Count with actor filter
	t.Run("CountWithActor", func(t *testing.T) {
		count, err := service.GetFilteredCount(&ListTaskFilter{Actor: "alice"})
		if err != nil {
			t.Fatalf("GetFilteredCount failed: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected count=1 for actor alice, got %d", count)
		}
	})

	// Test 5: Count with nil filter
	t.Run("CountWithNilFilter", func(t *testing.T) {
		count, err := service.GetFilteredCount(nil)
		if err != nil {
			t.Fatalf("GetFilteredCount with nil filter failed: %v", err)
		}
		if count != 3 {
			t.Errorf("Expected count=3 with nil filter, got %d", count)
		}
	})
}

func TestGetFilteredCount_WithNilDatabase(t *testing.T) {
	service := NewListService(nil)

	count, err := service.GetFilteredCount(&ListTaskFilter{})
	if err == nil {
		t.Error("Expected error with nil database, got nil")
	}
	if !errors.Is(err, ErrNilDatabase) {
		t.Errorf("Expected ErrNilDatabase, got %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count=0 with nil database, got %d", count)
	}
}
