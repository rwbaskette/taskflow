package service

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rwbaskette/taskflow/internal/db"
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

	// Test 1: List all tasks (no filter)
	t.Run("ListAllTasks", func(t *testing.T) {
		result, err := ListTasks(database, &ListTaskFilter{})
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
		result, err := ListTasks(database, &ListTaskFilter{Milestone: "v1.0"})
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
		result, err := ListTasks(database, &ListTaskFilter{Status: "done"})
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
		result, err := ListTasks(database, &ListTaskFilter{Actor: "alice"})
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
		result, err := ListTasks(database, &ListTaskFilter{
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
		result, err := ListTasks(database, &ListTaskFilter{Limit: 2})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if len(result.Tasks) != 2 {
			t.Errorf("Expected 2 tasks, got %d", len(result.Tasks))
		}
		// Total should reflect ALL matching tasks, not just the page returned
		if result.Total != 4 {
			t.Errorf("Expected Total=4 (full count), got %d", result.Total)
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
		result, err := ListTasks(database, &ListTaskFilter{Offset: 2})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		// With offset 2 and no limit, should get 2 remaining tasks
		if len(result.Tasks) != 2 {
			t.Errorf("Expected 2 tasks with offset 2, got %d", len(result.Tasks))
		}
		// Total should reflect ALL matching tasks, not just the page returned
		if result.Total != 4 {
			t.Errorf("Expected Total=4 (full count), got %d", result.Total)
		}
	})

	// Test 8: Nil filter (should work with defaults)
	t.Run("NilFilter", func(t *testing.T) {
		result, err := ListTasks(database, nil)
		if err != nil {
			t.Fatalf("ListTasks with nil filter failed: %v", err)
		}
		if len(result.Tasks) != 4 {
			t.Errorf("Expected 4 tasks with nil filter, got %d", len(result.Tasks))
		}
	})
}

func TestListTasks_TotalReflectsFullCount(t *testing.T) {
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

	// Insert 10 tasks so we can test pagination scenarios
	for i := 0; i < 10; i++ {
		task := db.Task{
			ID:          fmt.Sprintf("TASK-%03d", i+1),
			Milestone:   "v1.0",
			Title:       fmt.Sprintf("Task %d", i+1),
			Status:      "todo",
			Actor:       "alice",
			LastUpdated: time.Now(),
		}
		if err := database.CreateTask(&task); err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}
	}

	t.Run("Total with limit smaller than total", func(t *testing.T) {
		result, err := ListTasks(database, &ListTaskFilter{Limit: 3})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if len(result.Tasks) != 3 {
			t.Errorf("Expected 3 tasks in page, got %d", len(result.Tasks))
		}
		if result.Total != 10 {
			t.Errorf("Expected Total=10 (full count), got %d", result.Total)
		}
		if !result.HasMore {
			t.Error("Expected HasMore=true")
		}
	})

	t.Run("Total with limit equal to total", func(t *testing.T) {
		result, err := ListTasks(database, &ListTaskFilter{Limit: 10})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if len(result.Tasks) != 10 {
			t.Errorf("Expected 10 tasks in page, got %d", len(result.Tasks))
		}
		if result.Total != 10 {
			t.Errorf("Expected Total=10, got %d", result.Total)
		}
		if result.HasMore {
			t.Error("Expected HasMore=false when all results fit in one page")
		}
	})

	t.Run("Total with limit larger than total", func(t *testing.T) {
		result, err := ListTasks(database, &ListTaskFilter{Limit: 50})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if len(result.Tasks) != 10 {
			t.Errorf("Expected 10 tasks in page, got %d", len(result.Tasks))
		}
		if result.Total != 10 {
			t.Errorf("Expected Total=10, got %d", result.Total)
		}
		if result.HasMore {
			t.Error("Expected HasMore=false when limit exceeds total")
		}
	})

	t.Run("Total with offset and limit", func(t *testing.T) {
		result, err := ListTasks(database, &ListTaskFilter{Limit: 3, Offset: 5})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if len(result.Tasks) != 3 {
			t.Errorf("Expected 3 tasks in page, got %d", len(result.Tasks))
		}
		if result.Total != 10 {
			t.Errorf("Expected Total=10, got %d", result.Total)
		}
		if !result.HasMore {
			t.Error("Expected HasMore=true (offset 5 + 3 tasks < 10 total)")
		}
	})

	t.Run("Total with offset at end of results", func(t *testing.T) {
		result, err := ListTasks(database, &ListTaskFilter{Limit: 3, Offset: 9})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if len(result.Tasks) != 1 {
			t.Errorf("Expected 1 task in page, got %d", len(result.Tasks))
		}
		if result.Total != 10 {
			t.Errorf("Expected Total=10, got %d", result.Total)
		}
		if result.HasMore {
			t.Error("Expected HasMore=false (offset 9 + 1 task = 10 total)")
		}
	})

	t.Run("Total with milestone filter and limit", func(t *testing.T) {
		result, err := ListTasks(database, &ListTaskFilter{Milestone: "v1.0", Limit: 3})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if len(result.Tasks) != 3 {
			t.Errorf("Expected 3 tasks in page, got %d", len(result.Tasks))
		}
		if result.Total != 10 {
			t.Errorf("Expected Total=10 (all match milestone v1.0), got %d", result.Total)
		}
		if !result.HasMore {
			t.Error("Expected HasMore=true")
		}
	})
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

	// Test with nil filter
	result, err := ListTasks(database, nil)
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

// TestListTasks_TotalUsesFilteredCount verifies the Total field, which is
// produced by the filtered count query inlined into ListTasks, reflects the
// filtered full count ignoring limit/offset.
func TestListTasks_TotalUsesFilteredCount(t *testing.T) {
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

	// Test 1: Count all tasks
	t.Run("CountAll", func(t *testing.T) {
		result, err := ListTasks(database, &ListTaskFilter{Limit: 1})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if result.Total != 3 {
			t.Errorf("Expected count=3, got %d", result.Total)
		}
	})

	// Test 2: Count with milestone filter
	t.Run("CountWithMilestone", func(t *testing.T) {
		result, err := ListTasks(database, &ListTaskFilter{Milestone: "v1.0"})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if result.Total != 2 {
			t.Errorf("Expected count=2 for milestone v1.0, got %d", result.Total)
		}
	})

	// Test 3: Count with status filter
	t.Run("CountWithStatus", func(t *testing.T) {
		result, err := ListTasks(database, &ListTaskFilter{Status: "done"})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if result.Total != 1 {
			t.Errorf("Expected count=1 for status done, got %d", result.Total)
		}
	})

	// Test 4: Count with actor filter
	t.Run("CountWithActor", func(t *testing.T) {
		result, err := ListTasks(database, &ListTaskFilter{Actor: "alice"})
		if err != nil {
			t.Fatalf("ListTasks failed: %v", err)
		}
		if result.Total != 1 {
			t.Errorf("Expected count=1 for actor alice, got %d", result.Total)
		}
	})

	// Test 5: Count with nil filter
	t.Run("CountWithNilFilter", func(t *testing.T) {
		result, err := ListTasks(database, nil)
		if err != nil {
			t.Fatalf("ListTasks with nil filter failed: %v", err)
		}
		if result.Total != 3 {
			t.Errorf("Expected count=3 with nil filter, got %d", result.Total)
		}
	})
}

func TestGetTask_Valid(t *testing.T) {
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

	task := db.Task{
		ID:          "TASK-001",
		Milestone:   "v1.0",
		Title:       "Implement login",
		Description: "Add authentication system",
		Status:      "todo",
		Actor:       "alice",
		LastUpdated: time.Now(),
	}
	if err := database.CreateTask(&task); err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	item, err := GetTask(database, "TASK-001")
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if item.ID != "TASK-001" {
		t.Errorf("Expected ID TASK-001, got %s", item.ID)
	}
	if item.Title != "Implement login" {
		t.Errorf("Expected title 'Implement login', got %s", item.Title)
	}
	if item.Status != "todo" {
		t.Errorf("Expected status todo, got %s", item.Status)
	}
	if item.Created == "" || item.LastUpdated == "" {
		t.Error("Expected formatted timestamps to be non-empty")
	}
}

func TestGetTask_NotFound(t *testing.T) {
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

	item, err := GetTask(database, "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent task, got nil")
	}
	if item != nil {
		t.Errorf("Expected nil item, got %v", item)
	}
}

func TestGetTask_EmptyID(t *testing.T) {
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

	item, err := GetTask(database, "  ")
	if err != ErrInvalidID {
		t.Errorf("expected ErrInvalidID, got %v", err)
	}
	if item != nil {
		t.Errorf("Expected nil item, got %v", item)
	}
}
