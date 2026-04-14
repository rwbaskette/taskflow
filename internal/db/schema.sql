-- Tasks table schema
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    milestone TEXT,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'todo',
    actor TEXT,
    last_updated TEXT NOT NULL
);

-- Index for faster milestone lookups
CREATE INDEX IF NOT EXISTS idx_tasks_milestone ON tasks(milestone);

-- Index for faster status lookups
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);