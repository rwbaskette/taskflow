-- Schema versions table for tracking applied migrations
-- Each migration inserts a row here when successfully applied
CREATE TABLE IF NOT EXISTS schema_versions (
    id INTEGER PRIMARY KEY,
    version TEXT NOT NULL UNIQUE,
    applied_at TEXT NOT NULL,
    description TEXT
);

-- Index for faster version lookups
CREATE INDEX IF NOT EXISTS idx_schema_versions_version ON schema_versions(version);