-- Migration: Initial Schema (version 1.0.0)
-- ID: 001_initial
-- Date: 2026-06-06
-- Status: Initial schema for taskflow database
--
-- This migration creates the initial schema including:
-- - tasks table: main task storage
-- - deleted_tasks table: soft-deleted tasks
-- - schema_versions table: migration tracking
--
-- The migration is idempotent - running it multiple times is safe.
-- Each existing table uses "IF NOT EXISTS" clause.
--
-- Note: This migration does NOT automatically record itself in schema_versions.
-- The calling code (migrate.go) is responsible for recording the version.

-- Tasks table: main task storage
CREATE TABLE IF NOT EXISTS tasks (
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

-- Indexes for tasks table
CREATE INDEX IF NOT EXISTS idx_tasks_milestone ON tasks(milestone);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_sprint ON tasks(sprint);
CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);

-- Deleted tasks table: soft-deleted tasks
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

-- Index for deleted_tasks table
CREATE INDEX IF NOT EXISTS idx_deleted_tasks_deleted_on ON deleted_tasks(deleted_on);

-- Schema versions table: migration tracking
CREATE TABLE IF NOT EXISTS schema_versions (
    id INTEGER PRIMARY KEY,
    version TEXT NOT NULL UNIQUE,
    applied_at TEXT NOT NULL,
    description TEXT
);

-- Index for schema_versions table
CREATE INDEX IF NOT EXISTS idx_schema_versions_version ON schema_versions(version);

-- Return success indicator
SELECT '1.0.0 schema migration complete' AS status;