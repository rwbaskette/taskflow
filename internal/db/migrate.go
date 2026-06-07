package db

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// LatestVersion is the current schema version.
// Update this when adding new migrations.
const LatestVersion = "1.0.0"

// GetSchemaVersion returns the current schema version from the database.
// Returns "0.0.0" if no migrations have been applied (empty database).
// Returns an error if the database is corrupted.
func (db *DB) GetSchemaVersion() (string, error) {
	if db == nil || db.conn == nil {
		return "", ErrNilDB
	}

	// Check if schema_versions table exists
	var tableExists int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM pragma_table_info('schema_versions')").Scan(&tableExists)
	if err != nil || tableExists == 0 {
		// Table doesn't exist or is empty - this is a pre-migration database
		// Check if tasks table exists to determine if this is an old database
		var tasksCount int
		db.conn.QueryRow("SELECT COUNT(*) FROM pragma_table_info('tasks')").Scan(&tasksCount)
		if tasksCount > 0 {
			// Old database before migrations - return special sentinel
			return "pre-migration", nil
		}
		return "0.0.0", nil
	}

	// Get the latest applied version
	query := `SELECT version FROM schema_versions ORDER BY id DESC LIMIT 1`
	var version string
	err = db.conn.QueryRow(query).Scan(&version)
	if err != nil {
		// No versions applied yet
		return "0.0.0", nil
	}

	return version, nil
}

// GetAppliedMigrations returns all applied migration versions in order.
func (db *DB) GetAppliedMigrations() ([]string, error) {
	if db == nil || db.conn == nil {
		return nil, ErrNilDB
	}

	query := `SELECT version FROM schema_versions ORDER BY id ASC`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		versions = append(versions, v)
	}

	return versions, nil
}

// GetPendingMigrations returns versions that need to be applied to reach LatestVersion.
func (db *DB) GetPendingMigrations() ([]string, error) {
	applied, err := db.GetAppliedMigrations()
	if err != nil {
		return nil, err
	}

	// Get all available migration versions
	available := GetAvailableMigrations()

	// Find pending (available but not applied)
	appliedMap := make(map[string]bool)
	for _, v := range applied {
		appliedMap[v] = true
	}

	var pending []string
	for _, v := range available {
		if !appliedMap[v] {
			pending = append(pending, v)
		}
	}

	// Sort by version
	sort.Slice(pending, func(i, j int) bool {
		return CompareVersions(pending[i], pending[j]) < 0
	})

	return pending, nil
}

// GetAvailableMigrations returns all migration versions in order.
func GetAvailableMigrations() []string {
	// This should read from the migrations directory
	// For now, we return the known versions
	// In production, this would scan the migrations directory
	versions := []string{"1.0.0"}

	// Sort versions
	sort.Slice(versions, func(i, j int) bool {
		return CompareVersions(versions[i], versions[j]) < 0
	})

	return versions
}

// CompareVersions compares two semver versions.
// Returns -1 if v1 < v2, 0 if equal, 1 if v1 > v2.
func CompareVersions(v1, v2 string) int {
	pv1 := parseVersion(v1)
	pv2 := parseVersion(v2)

	if pv1.major < pv2.major {
		return -1
	}
	if pv1.major > pv2.major {
		return 1
	}
	if pv1.minor < pv2.minor {
		return -1
	}
	if pv1.minor > pv2.minor {
		return 1
	}
	if pv1.patch < pv2.patch {
		return -1
	}
	if pv1.patch > pv2.patch {
		return 1
	}
	return 0
}

// MajorVersion returns the major version number.
func MajorVersion(v string) int {
	return parseVersion(v).major
}

type versionParts struct {
	major, minor, patch int
}

func parseVersion(v string) versionParts {
	// Handle special cases
	if v == "pre-migration" {
		return versionParts{0, 0, 0}
	}

	parts := strings.Split(v, ".")
	if len(parts) < 3 {
		parts = append(parts, "0")
	}
	if len(parts) < 3 {
		parts = append(parts, "0")
	}

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	return versionParts{major: major, minor: minor, patch: patch}
}

// RunMigration applies a migration to the database.
// If the migration has already been applied, this is a no-op.
// Returns an error if migration fails.
func (db *DB) RunMigration(targetVersion string) error {
	if db == nil || db.conn == nil {
		return ErrNilDB
	}

	current, err := db.GetSchemaVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Check if already at or past target
	if CompareVersions(current, targetVersion) >= 0 {
		return nil // Already at or past target version
	}

	// Check for major version mismatch
	// Allow migration from "pre-migration" (legacy) and "0.0.0" (fresh install)
	if MajorVersion(current) != MajorVersion(targetVersion) && current != "pre-migration" && current != "0.0.0" {
		return fmt.Errorf("cannot migrate across major versions (%s -> %s): use RunMigrationForce to override", current, targetVersion)
	}

	// Apply migration
	if err := db.applyMigration(targetVersion); err != nil {
		return fmt.Errorf("failed to apply migration %s: %w", targetVersion, err)
	}

	return nil
}

// RunMigrationForce applies a migration even across major version boundaries.
// Use with caution - this can introduce breaking schema changes.
func (db *DB) RunMigrationForce(targetVersion string) error {
	if db == nil || db.conn == nil {
		return ErrNilDB
	}

	// Apply migration without major version check
	if err := db.applyMigration(targetVersion); err != nil {
		return fmt.Errorf("failed to apply migration %s: %w", targetVersion, err)
	}

	return nil
}

// applyMigration performs the actual migration to target version.
func (db *DB) applyMigration(targetVersion string) error {
	// Run the migration file for this version
	migrationSQL, err := GetMigrationSQL(targetVersion)
	if err != nil {
		return fmt.Errorf("migration file not found: %w", err)
	}

	// Begin transaction
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(migrationSQL); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record migration in schema_versions
	appliedAt := time.Now().UTC().Format(RFC3339Milli)
	desc := fmt.Sprintf("Migration to version %s", targetVersion)
	_, err = tx.Exec(
		"INSERT INTO schema_versions (version, applied_at, description) VALUES (?, ?, ?)",
		targetVersion, appliedAt, desc,
	)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

// GetMigrationSQL returns the SQL content for a migration version.
// This reads from embedded migration files or the migrations directory.
func GetMigrationSQL(version string) (string, error) {
	// Get the migrations directory path
	possiblePaths := getMigrationPaths()

	var migrationData []byte
	var lastErr error

	for _, migPath := range possiblePaths {
		// Try to find the migration file
		migrationFile := filepath.Join(migPath, fmt.Sprintf("001_initial--%s.sql", version))
		migrationData, lastErr = readFileOrEmbed(migrationFile, getEmbedPrefix(version))
		if lastErr == nil {
			break
		}

		// Try alternative naming pattern
		migrationFile = filepath.Join(migPath, fmt.Sprintf("migration--%s.sql", version))
		migrationData, lastErr = readFileOrEmbed(migrationFile, getEmbedPrefix(version))
		if lastErr == nil {
			break
		}
	}

	if len(migrationData) == 0 && lastErr != nil {
		return "", fmt.Errorf("migration file for %s not found: %w", version, lastErr)
	}

	return string(migrationData), nil
}

func getMigrationPaths() []string {
	paths := []string{}

	// From PROJECT_ROOT environment variable
	if projectRoot := getEnv("PROJECT_ROOT"); projectRoot != "" {
		paths = append(paths, filepath.Join(projectRoot, "internal", "db", "migrations"))
	}

	// From current working directory
	if cwd := getCWD(); cwd != "" {
		paths = append(paths, filepath.Join(cwd, "internal", "db", "migrations"))
	}

	return paths
}

func getEmbedPrefix(version string) string {
	return fmt.Sprintf("//go:embed migrations/001_initial--%s.sql", version)
}

// readFileOrEmbed reads a file, falling back to embedded content.
func readFileOrEmbed(path, embedKey string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		return data, nil
	}

	// Fall back to embedded schema if this is the initial migration
	if strings.Contains(embedKey, "1.0.0") {
		return []byte(GetInitialMigrationSQL()), nil
	}

	return nil, err
}

// GetInitialMigrationSQLFallback returns the SQL for the initial 1.0.0 schema.
func GetInitialMigrationSQL() string {
	// Fallback implementation when embedded file is not available
	return `CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    milestone TEXT,
    sprint TEXT,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'todo',
    priority INTEGER DEFAULT 0,
    actor TEXT,
    blocked_by TEXT,
    created TEXT NOT NULL,
    last_updated TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tasks_milestone ON tasks(milestone);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_sprint ON tasks(sprint);
CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);

CREATE TABLE IF NOT EXISTS deleted_tasks (
    id TEXT PRIMARY KEY,
    milestone TEXT,
    sprint TEXT,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL,
    priority INTEGER DEFAULT 0,
    actor TEXT,
    blocked_by TEXT,
    created TEXT NOT NULL,
    last_updated TEXT NOT NULL,
    deleted_on TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_deleted_tasks_deleted_on ON deleted_tasks(deleted_on);

CREATE TABLE IF NOT EXISTS schema_versions (
    id INTEGER PRIMARY KEY,
    version TEXT NOT NULL UNIQUE,
    applied_at TEXT NOT NULL,
    description TEXT
);

CREATE INDEX IF NOT EXISTS idx_schema_versions_version ON schema_versions(version);
`
}

// Helper functions for testing and environment access
func getEnv(key string) string {
	return os.Getenv(key)
}

func getCWD() string {
	cwd, _ := os.Getwd()
	return cwd
}