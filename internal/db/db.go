package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rwbaskette/taskflow/internal/anchor"

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
	conn, err := sql.Open("sqlite", "file:"+absPath+"?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)&_pragma=cache_size(10000)&_pragma=time_format(sqlite)")
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

// migrate runs schema migrations, creating tables if they don't exist. The
// schema is compiled into the binary via //go:embed schema.sql (a missing
// file is a build error), so no filesystem lookup is needed at runtime.
func (db *DB) migrate() error {
	if _, err := db.conn.Exec(string(embeddedSchema)); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}
	return nil
}

// DefaultDBPath resolves the database path (taskflow-init-anchor.md sections
// 3 and 9):
//
//  1. TASKFLOW_DIR set and non-empty: abs($TASKFLOW_DIR/tasks.db). The walk is
//     skipped entirely; env always wins; contract unchanged (the DB is
//     auto-created there, as before).
//  2. Otherwise the anchor walk (internal/anchor.Resolve from the
//     symlink-resolved cwd). A directory anchor or a valid pointer anchor
//     yields its DBPath. No DB is created here.
//
// On failure it returns ("", err). A dangling pointer maps to an *AnchorError
// with Reason "dangling" and the pointer file path, so runtime commands fail
// with the section 6 pointer-error contract; it never silently auto-creates a
// dangling target (design section 2); the *AnchorError is built from the
// Reason constant, matching the package's own conventions. Not-found,
// malformed, and wrong-type errors are returned as Resolve produced them.
//
// Walk branch only: after a successful Resolve the resolved DBPath is
// stat-checked. When the DB file is missing (deleted by hand, or a pointer to
// an existing anchor home whose tasks.db is absent), it returns an
// *AnchorError with Reason "missing-db" and the DB path instead of the path,
// so runtime commands never silently auto-create an empty DB over a deleted
// one (design section 3: "No DB is created. Only taskflow init creates or
// repairs."). The TASKFLOW_DIR env branch above is exempt by design.
func DefaultDBPath() (string, error) {
	if dir := os.Getenv("TASKFLOW_DIR"); dir != "" {
		absPath, err := filepath.Abs(dir)
		if err != nil {
			return filepath.Join(dir, "tasks.db"), nil
		}
		return filepath.Join(absPath, "tasks.db"), nil
	}

	a, aerr := anchor.Resolve("")
	if aerr != nil {
		return "", aerr
	}
	if a.PointerStatus == anchor.StatusDangling {
		return "", &anchor.AnchorError{
			Chain:       a.Chain,
			PointerPath: filepath.Join(a.AnchorPath, ".taskflow"),
			Reason:      anchor.ReasonDangling,
		}
	}
	// The anchor resolved; the DB file itself may still be gone. NewDB would
	// silently auto-create an empty DB here, which section 3 forbids.
	if _, serr := os.Stat(a.DBPath); serr != nil && os.IsNotExist(serr) {
		return "", &anchor.AnchorError{
			Chain:       a.Chain,
			PointerPath: a.DBPath,
			Reason:      anchor.ReasonMissingDB,
		}
	}
	return a.DBPath, nil
}
