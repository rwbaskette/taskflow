package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetSchemaVersion_newDB(t *testing.T) {
	// Create a fresh database with no schema_versions
	tmpDir := t.TempDir()
	testDBPath := filepath.Join(tmpDir, "test_new_migrate.db")

	// Create database
	database, err := NewDB(testDBPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer database.Close()

	// Get version - should be 1.0.0 after migration is run
	version, err := database.GetSchemaVersion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After NewDB with migrate(), should be at latest version
	if version != LatestVersion {
		t.Errorf("expected version %s, got %s", LatestVersion, version)
	}
}

func TestGetSchemaVersion_existing(t *testing.T) {
	// Create database with some tasks
	tmpDir := t.TempDir()
	testDBPath := filepath.Join(tmpDir, "test_existing_migrate.db")

	database, err := NewDB(testDBPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer database.Close()

	// Add a task
	task := &Task{
		ID:          "test-task-1",
		Milestone:   "milestone-1",
		Title:       "Test Task",
		Description: "Test description",
		Status:      "todo",
		Actor:       "testuser",
		Created:     time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	if err := database.CreateTask(task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Get version
	version, err := database.GetSchemaVersion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if version != LatestVersion {
		t.Errorf("expected version %s, got %s", LatestVersion, version)
	}
}

func TestBackupAndRestore(t *testing.T) {
	// Create database with data
	tmpDir := t.TempDir()
	testDBPath := filepath.Join(tmpDir, "test_backup.db")

	database, err := NewDB(testDBPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer database.Close()

	// Add a task
	task := &Task{
		ID:          "backup-test-task",
		Milestone:   "milestone-1",
		Title:       "Backup Test",
		Description: "Testing backup and restore",
		Status:      "todo",
		Actor:       "testuser",
		Created:     time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}
	if err := database.CreateTask(task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Create backup
	err = database.BackupDB()
	if err != nil {
		t.Fatalf("failed to create backup: %v", err)
	}

	// Verify backup exists
	if !database.BackupExists() {
		t.Fatal("backup should exist after BackupDB()")
	}

	// Get backup path
	backupPath, err := database.FullBackupPath()
	if err != nil {
		t.Fatalf("failed to get backup path: %v", err)
	}

	// Verify backup file exists on disk
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatal("backup file should exist on disk")
	}

	// Restore to a new database
	restorePath := filepath.Join(tmpDir, "test_restore.db")
	err = RestoreDB(backupPath, restorePath)
	if err != nil {
		t.Fatalf("failed to restore database: %v", err)
	}

	// Open restored database and verify data
	restoreDB, err := NewDB(restorePath)
	if err != nil {
		t.Fatalf("failed to open restored database: %v", err)
	}
	defer restoreDB.Close()

	restoredTask, err := restoreDB.ReadTask("backup-test-task")
	if err != nil {
		t.Fatalf("failed to read restored task: %v", err)
	}

	if restoredTask.Title != "Backup Test" {
		t.Errorf("expected title 'Backup Test', got '%s'", restoredTask.Title)
	}
}

func TestMigrationPreservesData(t *testing.T) {
	// This test would verify that migration preserves data
	// For now, we just verify the backup mechanism works
	tmpDir := t.TempDir()
	testDBPath := filepath.Join(tmpDir, "test_preserve.db")

	database, err := NewDB(testDBPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer database.Close()

	// Add tasks
	for i := 0; i < 5; i++ {
		task := &Task{
			ID:          "preserve-task-" + string(rune('0'+i)),
			Milestone:   "milestone-1",
			Title:       "Preserve Task",
			Description: "Testing data preservation",
			Status:      "todo",
			Actor:       "testuser",
			Created:     time.Now().UTC(),
			LastUpdated: time.Now().UTC(),
		}
		if err := database.CreateTask(task); err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
	}

	// Backup should succeed
	err = database.BackupDB()
	if err != nil {
		t.Fatalf("backup should succeed: %v", err)
	}

	// Should not be able to backup again
	err = database.BackupDB()
	if err == nil {
		t.Error("second backup should fail")
	}
}

func TestMigrationIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	testDBPath := filepath.Join(tmpDir, "test_idempotent.db")

	database, err := NewDB(testDBPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer database.Close()

	// Get initial version
	initialVersion, err := database.GetSchemaVersion()
	if err != nil {
		t.Fatalf("failed to get initial version: %v", err)
	}

	// Try to run same migration again
	err = database.RunMigration(initialVersion)
	if err != nil {
		t.Fatalf("running migration to same version should be no-op: %v", err)
	}

	// Version should be unchanged
	newVersion, err := database.GetSchemaVersion()
	if err != nil {
		t.Fatalf("failed to get new version: %v", err)
	}

	if initialVersion != newVersion {
		t.Errorf("version should be unchanged: got %s, want %s", newVersion, initialVersion)
	}
}

func TestGetAppliedMigrations(t *testing.T) {
	tmpDir := t.TempDir()
	testDBPath := filepath.Join(tmpDir, "test_applied.db")

	database, err := NewDB(testDBPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer database.Close()

	migrations, err := database.GetAppliedMigrations()
	if err != nil {
		t.Fatalf("failed to get applied migrations: %v", err)
	}

	// Should have at least one migration applied (the initial schema)
	if len(migrations) == 0 {
		t.Error("expected at least one migration to be applied")
	}
}

func TestGetPendingMigrations(t *testing.T) {
	tmpDir := t.TempDir()
	testDBPath := filepath.Join(tmpDir, "test_pending.db")

	database, err := NewDB(testDBPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer database.Close()

	pending, err := database.GetPendingMigrations()
	if err != nil {
		t.Fatalf("failed to get pending migrations: %v", err)
	}

	// Should have no pending migrations (already at latest)
	if len(pending) != 0 {
		t.Errorf("expected 0 pending migrations, got %d: %v", len(pending), pending)
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.9.0", "1.10.0", -1},
		{"1.10.0", "1.9.0", 1},
	}

	for _, tt := range tests {
		result := CompareVersions(tt.v1, tt.v2)
		if result != tt.expected {
			t.Errorf("CompareVersions(%s, %s) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
		}
	}
}

func TestMajorVersion(t *testing.T) {
	tests := []struct {
		version   string
		major     int
	}{
		{"1.0.0", 1},
		{"2.5.3", 2},
		{"10.0.0", 10},
		{"0.9.9", 0},
	}

	for _, tt := range tests {
		result := MajorVersion(tt.version)
		if result != tt.major {
			t.Errorf("MajorVersion(%s) = %d, want %d", tt.version, result, tt.major)
		}
	}
}

func TestBackupPreventsOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	testDBPath := filepath.Join(tmpDir, "test_prevent_overwrite.db")

	database, err := NewDB(testDBPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer database.Close()

	// Create backup
	err = database.BackupDB()
	if err != nil {
		t.Fatalf("failed to create backup: %v", err)
	}

	// Second backup should fail
	err = database.BackupDB()
	if err == nil {
		t.Error("expected error when backup already exists")
	}

	if err != nil && !contains(err.Error(), "backup already exists") {
		t.Errorf("expected 'backup already exists' error, got: %v", err)
	}
}

func TestSetSchemaVersion(t *testing.T) {
	// Test explicitly setting the schema version via RunMigration
	tmpDir := t.TempDir()
	testDBPath := filepath.Join(tmpDir, "test_set_version.db")

	database, err := NewDB(testDBPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer database.Close()

	// Verify the current schema version is "1.0.0" (latest)
	version, err := database.GetSchemaVersion()
	if err != nil {
		t.Fatalf("unexpected error getting version: %v", err)
	}
	if version != LatestVersion {
		t.Errorf("expected version %s, got %s", LatestVersion, version)
	}

	// Test edge case: nil DB
	var nilDB *DB
	err = nilDB.RunMigration("1.0.0")
	if err != ErrNilDB {
		t.Errorf("expected ErrNilDB for nil DB, got: %v", err)
	}

	// Test edge case: DB without proper connection (simulated by closing DB)
	// Note: After Close(), db.conn is NOT nil - it's a non-nil pointer to a closed database.
	// RunMigration checks `if db == nil || db.conn == nil` and returns ErrNilDB only when db.conn
	// is literally nil. Since db.conn remains non-nil after Close(), we expect a transaction error
	// rather than ErrNilDB.
	database.Close()
	err = database.RunMigration("1.0.0")
	if err == nil {
		t.Error("expected error for closed DB, got nil")
	}
}

func TestRunMigrations_newDB(t *testing.T) {
	// Test that running migrations on a new database works correctly
	tmpDir := t.TempDir()
	testDBPath := filepath.Join(tmpDir, "test_run_migrations_new.db")

	// Create fresh database
	database, err := NewDB(testDBPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer database.Close()

	// Run migration to 1.0.0
	err = database.RunMigration("1.0.0")
	if err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Verify the migration completed without error
	version, err := database.GetSchemaVersion()
	if err != nil {
		t.Fatalf("failed to get schema version: %v", err)
	}
	if version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", version)
	}

	// Verify the schema_versions table is properly populated
	migrations, err := database.GetAppliedMigrations()
	if err != nil {
		t.Fatalf("failed to get applied migrations: %v", err)
	}

	if len(migrations) == 0 {
		t.Error("expected at least one migration to be applied")
	}

	// Verify the applied migration is recorded
	found := false
	for _, m := range migrations {
		if m == "1.0.0" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 1.0.0 to be in applied migrations")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}