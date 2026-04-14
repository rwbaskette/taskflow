package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

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
	conn, err := sql.Open("sqlite3", absPath+"?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000")
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

// migrate runs schema migrations, creating tables if they don't exist
func (db *DB) migrate() error {
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

	if len(schemaData) == 0 {
		return fmt.Errorf("failed to read schema file: %w", lastErr)
	}

	schema := string(schemaData)

	// Execute the schema SQL
	_, err := db.conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

// DB returns the underlying sql.DB for direct queries if needed
func (db *DB) DB() *sql.DB {
	return db.conn
}
