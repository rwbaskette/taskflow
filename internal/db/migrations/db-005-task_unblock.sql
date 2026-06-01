-- Migration: task_unblock feature support
-- ID: db-005
-- Date: 2026-05-31
-- Status: No schema changes required
--
-- Summary:
--   This migration confirms the task_unblock feature can operate on the
--   existing schema without modifications.
--
-- Schema review findings:
--   1. status column (TEXT NOT NULL DEFAULT 'todo') already supports:
--      - 'todo', 'blocked', 'in_progress', 'done' values
--      - No constraint changes needed
--
--   2. blocked_by column (TEXT, nullable) already supports:
--      - NULL values indicating no blocking relationship
--      - Text values referencing blocking task IDs
--      - No NOT NULL constraint to remove
--
--   3. last_updated column (TEXT NOT NULL) already supports:
--      - ISO 8601 UTC timestamp storage
--      - No type or constraint changes needed
--
-- Rollback:
--   No changes were made, so no rollback is required.
--   This migration is inherently idempotent and reversible.

-- Verify task_unblock prerequisites exist (no-op if already present)
-- These checks confirm the schema is ready for task_unblock operations

SELECT 'task_unblock schema check passed' AS status
WHERE EXISTS (
    SELECT 1 FROM pragma_table_info('tasks')
    WHERE name = 'status' AND dflt_value = 'todo'
)
AND EXISTS (
    SELECT 1 FROM pragma_table_info('tasks')
    WHERE name = 'blocked_by'
)
AND EXISTS (
    SELECT 1 FROM pragma_table_info('tasks')
    WHERE name = 'last_updated'
);