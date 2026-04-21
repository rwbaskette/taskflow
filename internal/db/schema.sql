-- Tasks table schema
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

-- Index for faster milestone lookups
CREATE INDEX IF NOT EXISTS idx_tasks_milestone ON tasks(milestone);

-- Index for faster status lookups
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);

-- Index for faster sprint lookups
CREATE INDEX IF NOT EXISTS idx_tasks_sprint ON tasks(sprint);

-- Index for faster priority lookups
CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);

-- Soft-deleted tasks table
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

-- Index for faster deleted_on lookups
CREATE INDEX IF NOT EXISTS idx_deleted_tasks_deleted_on ON deleted_tasks(deleted_on);