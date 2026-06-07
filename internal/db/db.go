package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	_ "embed"
)

//go:embed schema.sql
var embeddedSchema string

// DB represents a database connection with connection pooling
type DB struct {
	conn *sql.DB
	path string
}

// NewDB creates a new database connection with connection pooling
func NewDB(dbPath string) (*DB, error) {
	// Resolve the database path to an absolute path
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve database path: %w", err)
	}

	// Ensure the directory exists
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection with SQLite
	conn, err := sql.Open("sqlite", "file:"+absPath+"?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=cache_size(10000)&_pragma=time_format(sqlite)")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool settings
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	// Verify the connection is valid
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create the DB instance
	database := &DB{
		conn: conn,
		path: absPath,
	}

	// Run schema migrations
	if err := database.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return database, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// Path returns the absolute path to the database file
func (db *DB) Path() string {
	return db.path
}

// migrate runs schema migrations, ensuring the database is at the latest version.
// It uses the migration framework to apply any pending migrations.
func (db *DB) migrate() error {
	// First, apply the initial schema to ensure base tables exist
	// This handles both new databases and legacy databases without schema_versions
	if err := db.applyInitialSchema(); err != nil {
		return fmt.Errorf("failed to apply initial schema: %w", err)
	}

	// Now check if we need to run any pending migrations
	currentVersion, err := db.GetSchemaVersion()
	if err != nil {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	// If already at latest version, nothing to do
	if currentVersion == LatestVersion {
		return nil
	}

	// If pre-migration database, run migration to get to latest
	if currentVersion == "pre-migration" || CompareVersions(currentVersion, LatestVersion) < 0 {
		// Get pending migrations
		pending, err := db.GetPendingMigrations()
		if err != nil {
			return fmt.Errorf("failed to get pending migrations: %w", err)
		}

		// Apply each pending migration
		for _, version := range pending {
			// Check major version mismatch (only for non-pre-migration databases)
			// Also allow migration from "0.0.0" (no version = fresh install)
			if currentVersion != "pre-migration" && currentVersion != "0.0.0" && MajorVersion(currentVersion) != MajorVersion(version) {
				// Skip major version upgrades on auto-migrate (caller should use --force)
				return fmt.Errorf("cannot auto-migrate across major versions (current: %s, target: %s): use --force flag to override", currentVersion, version)
			}

			if err := db.RunMigration(version); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", version, err)
			}
			currentVersion = version
		}
	}

	return nil
}

// applyInitialSchema creates the base schema if it doesn't exist.
// This is called before the migration system to ensure basic tables are present.
func (db *DB) applyInitialSchema() error {
	// Try multiple locations for the schema file
	// Priority: PROJECT_ROOT env > db.path derivation > cwd

	possiblePaths := []string{}

	// From PROJECT_ROOT environment variable (used in tests)
	if projectRoot := os.Getenv("PROJECT_ROOT"); projectRoot != "" {
		possiblePaths = append(possiblePaths, filepath.Join(projectRoot, "internal", "db", "schema.sql"))
	}

	// From absolute db path: /home/rwbaskette/tmp/... -> /home/rwbaskette/tmp/internal/db/schema.sql
	if filepath.IsAbs(db.path) {
		projectRoot := filepath.Dir(db.path)
		possiblePaths = append(possiblePaths, filepath.Join(projectRoot, "internal", "db", "schema.sql"))
	}

	// From current working directory
	cwd, _ := os.Getwd()
	possiblePaths = append(possiblePaths, filepath.Join(cwd, "internal", "db", "schema.sql"))

	// Try to find the schema file
	var schemaData []byte
	var lastErr error
	for _, schemaFile := range possiblePaths {
		schemaData, lastErr = os.ReadFile(schemaFile)
		if lastErr == nil {
			break
		}
	}

	// Fallback to embedded schema if file not found
	if len(schemaData) == 0 && embeddedSchema != "" {
		schemaData = []byte(embeddedSchema)
		lastErr = nil
	}

	if len(schemaData) == 0 {
		return fmt.Errorf("failed to read schema file: %w", lastErr)
	}

	schema := string(schemaData)

	// Execute the schema SQL
	_, err := db.conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to execute initial schema: %w", err)
	}

	return nil
}

// DefaultDBPath returns the database path based on the TASKFLOW_DIR environment variable.
// If TASKFLOW_DIR is set and non-empty, returns "$TASKFLOW_DIR/tasks.db" (resolved to absolute path).
// Otherwise, returns the default ".taskflow/tasks.db".
func DefaultDBPath() string {
	if dir := os.Getenv("TASKFLOW_DIR"); dir != "" {
		absPath, err := filepath.Abs(dir)
		if err != nil {
			return filepath.Join(dir, "tasks.db")
		}
		return filepath.Join(absPath, "tasks.db")
	}
	return ".taskflow/tasks.db"
}

// DB returns the underlying sql.DB for direct queries if needed
func (db *DB) DB() *sql.DB {
	return db.conn
}
