package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rwbaskette/taskflow/internal/db"
)

// TestConfig holds configuration for integration tests
type TestConfig struct {
	DBPath      string
	DB          *db.DB
	ProjectRoot string
}

// setupTestDB creates a new test database for each test
func setupTestDB(t *testing.T) *TestConfig {
	// Create unique test database path
	testID := t.Name()
	dbPath := filepath.Join("/home/rwbaskette/tmp", "test_integration_"+testID+".db")

	// Ensure clean state - remove any existing test db
	if err := os.RemoveAll(dbPath); err != nil {
		t.Fatalf("failed to remove existing test db: %v", err)
	}

	// Set project root for schema lookup
	projectRoot := "/home/rwbaskette/tmp"
	os.Setenv("PROJECT_ROOT", projectRoot)

	// Create new database
	testDB, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	return &TestConfig{
		DBPath:      dbPath,
		DB:          testDB,
		ProjectRoot: projectRoot,
	}
}

// teardownTestDB closes and removes the test database
func teardownTestDB(t *testing.T, cfg *TestConfig) {
	if cfg == nil || cfg.DB == nil {
		return
	}

	// Close database connection
	if err := cfg.DB.Close(); err != nil {
		t.Logf("warning: failed to close database: %v", err)
	}

	// Remove test database file
	if err := os.RemoveAll(cfg.DBPath); err != nil {
		t.Logf("warning: failed to remove test db: %v", err)
	}

	// Clean up any WAL/SHM files
	walPath := cfg.DBPath + "-wal"
	shmPath := cfg.DBPath + "-shm"
	os.RemoveAll(walPath)
	os.RemoveAll(shmPath)
}

// setupTestDBForTest executes setup and returns cleanup function for manual control
func setupTestDBForTest(t *testing.T) func() {
	cfg := setupTestDB(t)

	// Return cleanup function
	return func() {
		teardownTestDB(t, cfg)
	}
}

// createTestTask creates a test task in the database for testing
func createTestTask(t *testing.T, cfg *TestConfig, id, title, description, milestone, actor string) error {
	task := &db.Task{
		ID:          id,
		Title:       title,
		Description: description,
		Milestone:   milestone,
		Actor:       actor,
		Status:      "todo",
	}

	return cfg.DB.CreateTask(task)
}

// getTask retrieves a task by ID from the database
func getTask(t *testing.T, cfg *TestConfig, id string) (*db.Task, error) {
	return cfg.DB.ReadTask(id)
}

// deleteTestTask deletes a test task by ID
func deleteTestTask(t *testing.T, cfg *TestConfig, id string) error {
	return cfg.DB.DeleteTask(id)
}

// listAllTasks retrieves all tasks from the database
func listAllTasks(t *testing.T, cfg *TestConfig) []db.Task {
	tasks, err := cfg.DB.ListTasks(db.TaskFilter{})
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}
	return tasks
}

// countTasks counts all tasks in the database
func countTasks(t *testing.T, cfg *TestConfig) int {
	tasks, err := cfg.DB.ListTasks(db.TaskFilter{})
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}
	return len(tasks)
}
