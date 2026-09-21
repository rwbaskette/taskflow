-- Tasks table schema
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    milestone TEXT,
    sprint TEXT,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'todo',
    actor TEXT,
    blocked_by TEXT,
    created TEXT NOT NULL,
    last_updated TEXT NOT NULL
);

-- Index for faster milestone lookups
CREATE INDEX IF NOT EXISTS idx_tasks_milestone ON tasks(milestone);

-- Index for faster status lookups
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);

-- Index for faster sprint lookups
CREATE INDEX IF NOT EXISTS idx_tasks_sprint ON tasks(sprint);

-- Soft-deleted tasks table
CREATE TABLE IF NOT EXISTS deleted_tasks (
    id TEXT PRIMARY KEY,
    milestone TEXT,
    sprint TEXT,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL,
    actor TEXT,
    blocked_by TEXT,
    created TEXT NOT NULL,
    last_updated TEXT NOT NULL,
    deleted_on TEXT NOT NULL
);

-- Index for faster deleted_on lookups
CREATE INDEX IF NOT EXISTS idx_deleted_tasks_deleted_on ON deleted_tasks(deleted_on);