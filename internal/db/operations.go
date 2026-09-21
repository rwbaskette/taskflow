package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

// Task represents a task in the database
type Task struct {
	ID          string    `json:"id"`
	Milestone   string    `json:"milestone,omitempty"`
	Sprint      string    `json:"sprint,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	Actor       string    `json:"actor,omitempty"`
	BlockedBy   []string  `json:"blocked_by,omitempty"`
	Created     time.Time `json:"created"`
	LastUpdated time.Time `json:"last_updated"`
}

// ValidSortKeys are the canonical sort keys accepted by TaskFilter.SortBy.
var ValidSortKeys = []string{
	"status",
	"milestone",
	"created",
	"updated",
	"id",
	"title",
	"description",
	"actor",
}

// sortClauses maps each valid sort key to its ORDER BY clause. Keys not in
// the map, including the empty string, fall back to the default clause.
var sortClauses = map[string]string{
	"status":      " ORDER BY status ASC",
	"milestone":   " ORDER BY milestone ASC, last_updated DESC",
	"created":     " ORDER BY created DESC",
	"updated":     " ORDER BY last_updated DESC",
	"id":          " ORDER BY id ASC",
	"title":       " ORDER BY title ASC",
	"description": " ORDER BY description ASC",
	"actor":       " ORDER BY actor ASC",
}

// sortOrder returns the ORDER BY clause for the given raw sort key.
// Unknown keys (including "") fall back to the default ordering.
func sortOrder(sortBy string) string {
	if clause, ok := sortClauses[sortBy]; ok {
		return clause
	}
	return " ORDER BY last_updated DESC"
}

// TaskFilter contains optional filters for listing tasks
type TaskFilter struct {
	Milestone string
	Status    string
	Actor     string
	ID        string
	SortBy    string
	Limit     int
	Offset    int
}

// scanTask scans one row of the standard task SELECT (id, milestone, sprint,
// title, description, status, actor, blocked_by, created, last_updated) into a
// Task. It unmarshals blocked_by (SQL NULL or empty means no blockers) and
// parses created/last_updated as RFC3339. The DB always writes those
// timestamps in RFC3339, so a parse failure means corrupt data and is
// propagated.
func scanTask(scan interface{ Scan(dest ...any) error }) (Task, error) {
	var t Task
	var createdStr string
	var lastUpdatedStr string
	var blockedByStr *string

	if err := scan.Scan(
		&t.ID, &t.Milestone, &t.Sprint, &t.Title, &t.Description,
		&t.Status, &t.Actor, &blockedByStr, &createdStr, &lastUpdatedStr,
	); err != nil {
		return Task{}, err
	}

	if blockedByStr != nil && *blockedByStr != "" {
		if err := json.Unmarshal([]byte(*blockedByStr), &t.BlockedBy); err != nil {
			return Task{}, fmt.Errorf("parse blocked_by: %w", err)
		}
	}

	var err error
	t.Created, err = time.Parse(time.RFC3339, createdStr)
	if err != nil {
		return Task{}, fmt.Errorf("parse created: %w", err)
	}

	t.LastUpdated, err = time.Parse(time.RFC3339, lastUpdatedStr)
	if err != nil {
		return Task{}, fmt.Errorf("parse last_updated: %w", err)
	}

	return t, nil
}

// ValidStatuses are the canonical status values stored in the database.
var ValidStatuses = []string{"todo", "in_progress", "done", "blocked"}

// validateTask validates task data before creation/update
func validateTask(t *Task) error {
	if t == nil {
		return ErrNilTask
	}

	if strings.TrimSpace(t.ID) == "" {
		return &InvalidTaskError{Field: "id", Message: "ID cannot be empty"}
	}

	if strings.TrimSpace(t.Title) == "" {
		return &InvalidTaskError{Field: "title", Message: "title cannot be empty"}
	}

	if !slices.Contains(ValidStatuses, t.Status) {
		return &InvalidTaskError{Field: "status", Message: "status must be one of: " + strings.Join(ValidStatuses, ", ")}
	}

	return nil
}

// CreateTask creates a new task in the database
func (db *DB) CreateTask(t *Task) error {
	if db == nil || db.conn == nil {
		return ErrNilDB
	}

	if err := validateTask(t); err != nil {
		return err
	}

	// Check if the task already exists
	var exists bool
	err := db.conn.QueryRow(`SELECT EXISTS(SELECT 1 FROM tasks WHERE id = ?)`, t.ID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check task existence: %w", err)
	}
	if exists {
		return &TaskAlreadyExistsError{ID: t.ID}
	}

	// Set Created to now if not set
	if t.Created.IsZero() {
		t.Created = time.Now().UTC()
	}
	// Set LastUpdated to now if not set
	if t.LastUpdated.IsZero() {
		t.LastUpdated = time.Now().UTC()
	}

	blockedByJSON, _ := json.Marshal(t.BlockedBy)

	query := `
		INSERT INTO tasks (id, milestone, sprint, title, description, status, actor, blocked_by, created, last_updated)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = db.conn.Exec(query,
		t.ID,
		t.Milestone,
		t.Sprint,
		t.Title,
		t.Description,
		t.Status,
		t.Actor,
		string(blockedByJSON),
		t.Created.Format(time.RFC3339),
		t.LastUpdated.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	return nil
}

// ReadTask retrieves a task by ID
func (db *DB) ReadTask(id string) (*Task, error) {
	if db == nil || db.conn == nil {
		return nil, ErrNilDB
	}

	if strings.TrimSpace(id) == "" {
		return nil, ErrInvalidID
	}

	query := `
		SELECT id, milestone, sprint, title, description, status, actor, blocked_by, created, last_updated
		FROM tasks
		WHERE id = ?
	`

	t, err := scanTask(db.conn.QueryRow(query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &TaskNotFoundError{ID: id}
		}
		return nil, fmt.Errorf("failed to read task: %w", err)
	}

	return &t, nil
}

// UpdateTask updates an existing task
func (db *DB) UpdateTask(t *Task) error {
	if db == nil || db.conn == nil {
		return ErrNilDB
	}

	if err := validateTask(t); err != nil {
		return err
	}

	// Always update LastUpdated to current time
	t.LastUpdated = time.Now().UTC()

	blockedByJSON, _ := json.Marshal(t.BlockedBy)

	query := `
		UPDATE tasks
		SET milestone = ?, title = ?, description = ?, status = ?, actor = ?, blocked_by = ?, last_updated = ?
		WHERE id = ?
	`

	result, err := db.conn.Exec(query,
		t.Milestone,
		t.Title,
		t.Description,
		t.Status,
		t.Actor,
		string(blockedByJSON),
		t.LastUpdated.Format(time.RFC3339),
		t.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return &TaskNotFoundError{ID: t.ID}
	}

	return nil
}

// SoftDeleteTask moves a task to the deleted_tasks table with a deleted_on
// timestamp. A single INSERT..SELECT copies the row verbatim (including
// blocked_by) and stamps deleted_on, so no pre-read scan or parse is needed.
// It returns the deleted_on value it stored, so callers display exactly the
// timestamp that was persisted instead of guessing their own.
func (db *DB) SoftDeleteTask(id string) (time.Time, error) {
	if db == nil || db.conn == nil {
		return time.Time{}, ErrNilDB
	}

	if strings.TrimSpace(id) == "" {
		return time.Time{}, ErrInvalidID
	}

	tx, err := db.conn.Begin()
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	deletedOn := time.Now().UTC()
	insertQuery := `
		INSERT INTO deleted_tasks (id, milestone, sprint, title, description, status, actor, blocked_by, created, last_updated, deleted_on)
		SELECT id, milestone, sprint, title, description, status, actor, blocked_by, created, last_updated, ?
		FROM tasks WHERE id = ?
	`
	result, err := tx.Exec(insertQuery, deletedOn.Format(time.RFC3339), id)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to insert into deleted_tasks: %w", err)
	}

	// RowsAffected == 0 means the SELECT matched no row, i.e. the task
	// does not exist.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return time.Time{}, &TaskNotFoundError{ID: id}
	}

	deleteQuery := `DELETE FROM tasks WHERE id = ?`
	if _, err := tx.Exec(deleteQuery, id); err != nil {
		return time.Time{}, fmt.Errorf("failed to delete from tasks: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return time.Time{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return deletedOn, nil
}

// buildTaskWhere builds the " AND ..." WHERE suffix and query args shared by
// ListTasks and CountTasks so the filter conditions cannot drift apart.
func buildTaskWhere(filter TaskFilter) (string, []interface{}) {
	where := ""
	args := []interface{}{}

	if filter.Milestone != "" {
		where += " AND milestone = ?"
		args = append(args, filter.Milestone)
	}

	if filter.Status != "" {
		where += " AND status = ?"
		args = append(args, filter.Status)
	}

	if filter.Actor != "" {
		where += " AND actor = ?"
		args = append(args, filter.Actor)
	}

	if filter.ID != "" {
		where += " AND id = ?"
		args = append(args, filter.ID)
	}

	return where, args
}

// ListTasks retrieves tasks with optional filters
func (db *DB) ListTasks(filter TaskFilter) ([]Task, error) {
	if db == nil || db.conn == nil {
		return nil, ErrNilDB
	}

	// Build query with filters
	query := "SELECT id, milestone, sprint, title, description, status, actor, blocked_by, created, last_updated FROM tasks WHERE 1=1"
	where, args := buildTaskWhere(filter)
	query += where

	// Apply sorting
	orderBy := sortOrder(filter.SortBy)
	query += orderBy

	// Apply pagination
	// SQLite requires LIMIT when using OFFSET
	if filter.Offset > 0 {
		if filter.Limit > 0 {
			query += " LIMIT ?"
			args = append(args, filter.Limit)
		} else {
			// Use a large default limit when only offset is specified
			query += " LIMIT ?"
			args = append(args, 10000)
		}
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	} else if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tasks: %w", err)
	}

	// Return empty slice instead of nil for consistency
	if tasks == nil {
		tasks = []Task{}
	}

	return tasks, nil
}

// CountTasks returns the total number of tasks matching the given filter,
// ignoring Limit and Offset. This is used for pagination metadata so that
// callers can know the full result-set size without fetching all rows.
func (db *DB) CountTasks(filter TaskFilter) (int, error) {
	if db == nil || db.conn == nil {
		return 0, ErrNilDB
	}

	// Build the COUNT query using the same WHERE conditions as ListTasks
	// but without ORDER BY, LIMIT, or OFFSET.
	query := "SELECT COUNT(*) FROM tasks WHERE 1=1"
	where, args := buildTaskWhere(filter)
	query += where

	var count int
	err := db.conn.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	return count, nil
}

// UnblockTask transitions a task from 'blocked' to 'todo' status in a single
// atomic database operation. The WHERE clause includes a status = 'blocked'
// guard to prevent unauthorized status transitions. The blocked_by field is
// set to SQL NULL and last_updated is stamped with the current UTC time. If a
// non-empty newDescription is provided, it overwrites the existing
// description; otherwise the description is preserved unchanged.
//
// On success it returns the freshly read task, so callers see the true stored
// state (including the real last_updated timestamp). If the task does not
// exist it returns *TaskNotFoundError; if it exists but is not in 'blocked'
// status it returns *TaskNotBlockedError.
func (db *DB) UnblockTask(id string, newDescription string) (*Task, error) {
	if db == nil || db.conn == nil {
		return nil, ErrNilDB
	}

	if strings.TrimSpace(id) == "" {
		return nil, ErrInvalidID
	}

	// Build the SET clause: include description only when a new value is given.
	set := "last_updated = ?"
	args := []interface{}{time.Now().UTC().Format(time.RFC3339)}
	if newDescription != "" {
		set = "description = ?, " + set
		args = append([]interface{}{newDescription}, args...)
	}

	query := `
		UPDATE tasks
		SET status = 'todo',
		    blocked_by = NULL,
		    ` + set + `
		WHERE id = ? AND status = 'blocked'
	`
	args = append(args, id)

	result, err := db.conn.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to unblock task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		// Either the task doesn't exist or it is not in 'blocked' status.
		// ReadTask distinguishes the two cases.
		task, err := db.ReadTask(id)
		if err != nil {
			return nil, err // *TaskNotFoundError
		}
		return nil, &TaskNotBlockedError{ID: id, Status: task.Status}
	}

	return db.ReadTask(id)
}
