package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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
	Priority    int       `json:"priority,omitempty"`
	Actor       string    `json:"actor,omitempty"`
	BlockedBy   []string  `json:"blocked_by,omitempty"`
	Created     time.Time `json:"created"`
	LastUpdated time.Time `json:"last_updated"`
}

// SortBy defines the field to sort by
type SortBy string

const (
	SortByStatus      SortBy = "status"
	SortByPriority    SortBy = "priority"
	SortByMilestone   SortBy = "milestone"
	SortByCreated     SortBy = "created"
	SortByUpdated     SortBy = "updated"
	SortByID          SortBy = "id"
	SortBySprint      SortBy = "sprint"
	SortByTitle       SortBy = "title"
	SortByDescription SortBy = "description"
	SortByActor       SortBy = "actor"
)

// ValidSortByValues returns all valid sort by values
func ValidSortByValues() []string {
	return []string{"status", "priority", "milestone", "created", "updated", "id", "sprint", "title", "description", "actor"}
}

// TaskFilter contains optional filters for listing tasks
type TaskFilter struct {
	Milestone string
	Sprint    string
	Status    string
	Actor     string
	ID        string
	SortBy    SortBy
	Limit     int
	Offset    int
}

// validateTask validates task data before creation/update
func validateTask(t *Task) error {
	if t == nil {
		return ErrNilTask
	}

	if strings.TrimSpace(t.ID) == "" {
		return NewInvalidTaskError("id", "ID cannot be empty")
	}

	if strings.TrimSpace(t.Title) == "" {
		return NewInvalidTaskError("title", "title cannot be empty")
	}

	validStatuses := map[string]bool{
		"todo":        true,
		"in_progress": true,
		"done":        true,
		"blocked":     true,
	}

	if !validStatuses[t.Status] {
		return NewInvalidTaskError("status", "status must be one of: todo, in_progress, done, blocked")
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

	// Check if task already exists
	exists, err := db.taskExists(t.ID)
	if err != nil {
		return err
	}
	if exists {
		return NewTaskAlreadyExistsError(t.ID)
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

// CreateTaskTx creates a new task within a transaction
// Deprecated: Not used in production code, only in tests
func (db *DB) CreateTaskTx(tx *sql.Tx, t *Task) error {
	if db == nil || db.conn == nil {
		return ErrNilDB
	}
	if tx == nil {
		return errors.New("nil transaction provided")
	}

	if err := validateTask(t); err != nil {
		return err
	}

	// Check if task already exists within transaction
	exists, err := db.taskExistsTx(tx, t.ID)
	if err != nil {
		return err
	}
	if exists {
		return NewTaskAlreadyExistsError(t.ID)
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

	_, err = tx.Exec(query,
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
		return fmt.Errorf("failed to create task in transaction: %w", err)
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

	var t Task
	var createdStr string
	var lastUpdatedStr string
	var blockedByStr *string

	err := db.conn.QueryRow(query, id).Scan(
		&t.ID,
		&t.Milestone,
		&t.Sprint,
		&t.Title,
		&t.Description,
		&t.Status,
		&t.Actor,
		&blockedByStr,
		&createdStr,
		&lastUpdatedStr,
	)
	if blockedByStr != nil && *blockedByStr != "" {
		json.Unmarshal([]byte(*blockedByStr), &t.BlockedBy)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewTaskNotFoundError(id)
		}
		return nil, fmt.Errorf("failed to read task: %w", err)
	}

	t.Created, err = time.Parse(time.RFC3339, createdStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created: %w", err)
	}

	t.LastUpdated, err = time.Parse(time.RFC3339, lastUpdatedStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse last_updated: %w", err)
	}

	return &t, nil
}

// ReadTaskTx retrieves a task by ID within a transaction
// Deprecated: Not used in production code, only in tests
func (db *DB) ReadTaskTx(tx *sql.Tx, id string) (*Task, error) {
	if tx == nil {
		return nil, errors.New("nil transaction provided")
	}

	if strings.TrimSpace(id) == "" {
		return nil, ErrInvalidID
	}

	query := `
		SELECT id, milestone, sprint, title, description, status, actor, blocked_by, created, last_updated
		FROM tasks
		WHERE id = ?
	`

	var t Task
	var createdStr string
	var lastUpdatedStr string
	var blockedByRaw interface{}

	err := tx.QueryRow(query, id).Scan(
		&t.ID,
		&t.Milestone,
		&t.Sprint,
		&t.Title,
		&t.Description,
		&t.Status,
		&t.Actor,
		&blockedByRaw,
		&createdStr,
		&lastUpdatedStr,
	)
	if blockedByRaw != nil {
		if s, ok := blockedByRaw.(string); ok && s != "" {
			json.Unmarshal([]byte(s), &t.BlockedBy)
		}
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewTaskNotFoundError(id)
		}
		return nil, fmt.Errorf("failed to read task in transaction: %w", err)
	}

	t.Created, err = time.Parse(time.RFC3339, createdStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created in transaction: %w", err)
	}

	t.LastUpdated, err = time.Parse(time.RFC3339, lastUpdatedStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse last_updated in transaction: %w", err)
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

	var blockedByParam interface{}
	if t.BlockedBy != nil {
		blockedByJSON, _ := json.Marshal(t.BlockedBy)
		blockedByParam = string(blockedByJSON)
	}

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
		blockedByParam,
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
		return NewTaskNotFoundError(t.ID)
	}

	return nil
}

// UpdateTaskTx updates an existing task within a transaction
// Deprecated: Not used in production code, only in tests
func (db *DB) UpdateTaskTx(tx *sql.Tx, t *Task) error {
	if tx == nil {
		return errors.New("nil transaction provided")
	}

	if err := validateTask(t); err != nil {
		return err
	}

	// Always update LastUpdated to current time
	t.LastUpdated = time.Now().UTC()

	var blockedByParam interface{}
	if t.BlockedBy != nil {
		blockedByJSON, _ := json.Marshal(t.BlockedBy)
		blockedByParam = string(blockedByJSON)
	}

	query := `
		UPDATE tasks
		SET milestone = ?, title = ?, description = ?, status = ?, actor = ?, blocked_by = ?, last_updated = ?
		WHERE id = ?
	`

	result, err := tx.Exec(query,
		t.Milestone,
		t.Title,
		t.Description,
		t.Status,
		t.Actor,
		blockedByParam,
		t.LastUpdated.Format(time.RFC3339),
		t.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task in transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return NewTaskNotFoundError(t.ID)
	}

	return nil
}

// DeleteTask deletes a task by ID
func (db *DB) DeleteTask(id string) error {
	if db == nil || db.conn == nil {
		return ErrNilDB
	}

	if strings.TrimSpace(id) == "" {
		return ErrInvalidID
	}

	query := `DELETE FROM tasks WHERE id = ?`

	result, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return NewTaskNotFoundError(id)
	}

	return nil
}

// SoftDeleteTask moves a task to the deleted_tasks table with a deleted_on timestamp
func (db *DB) SoftDeleteTask(id string) error {
	if db == nil || db.conn == nil {
		return ErrNilDB
	}

	if strings.TrimSpace(id) == "" {
		return ErrInvalidID
	}

	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var t Task
	var createdStr string
	var lastUpdatedStr string
	var blockedByRaw interface{}

	selectQuery := `
		SELECT id, milestone, sprint, title, description, status, priority, actor, blocked_by, created, last_updated
		FROM tasks WHERE id = ?
	`
	err = tx.QueryRow(selectQuery, id).Scan(
		&t.ID, &t.Milestone, &t.Sprint, &t.Title, &t.Description,
		&t.Status, &t.Priority, &t.Actor, &blockedByRaw, &createdStr, &lastUpdatedStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return NewTaskNotFoundError(id)
		}
		return fmt.Errorf("failed to read task: %w", err)
	}
	if blockedByRaw != nil {
		if s, ok := blockedByRaw.(string); ok && s != "" {
			json.Unmarshal([]byte(s), &t.BlockedBy)
		}
	}

	t.Created, _ = time.Parse(time.RFC3339, createdStr)
	t.LastUpdated, _ = time.Parse(time.RFC3339, lastUpdatedStr)
	deletedOn := time.Now().UTC()

	blockedByJSON, _ := json.Marshal(t.BlockedBy)

	insertQuery := `
		INSERT INTO deleted_tasks (id, milestone, sprint, title, description, status, priority, actor, blocked_by, created, last_updated, deleted_on)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = tx.Exec(insertQuery,
		t.ID, t.Milestone, t.Sprint, t.Title, t.Description,
		t.Status, t.Priority, t.Actor, string(blockedByJSON),
		t.Created.Format(time.RFC3339), t.LastUpdated.Format(time.RFC3339),
		deletedOn.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to insert into deleted_tasks: %w", err)
	}

	deleteQuery := `DELETE FROM tasks WHERE id = ?`
	result, err := tx.Exec(deleteQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete from tasks: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return NewTaskNotFoundError(id)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetTaskByID retrieves a task by its ID
func (db *DB) GetTaskByID(id string) (*Task, error) {
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

	var t Task
	var createdStr string
	var lastUpdatedStr string
	var blockedByStr *string

	err := db.conn.QueryRow(query, id).Scan(
		&t.ID,
		&t.Milestone,
		&t.Sprint,
		&t.Title,
		&t.Description,
		&t.Status,
		&t.Actor,
		&blockedByStr,
		&createdStr,
		&lastUpdatedStr,
	)
	if blockedByStr != nil && *blockedByStr != "" {
		json.Unmarshal([]byte(*blockedByStr), &t.BlockedBy)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewTaskNotFoundError(id)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	t.Created, err = time.Parse(time.RFC3339, createdStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created: %w", err)
	}

	t.LastUpdated, err = time.Parse(time.RFC3339, lastUpdatedStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse last_updated: %w", err)
	}

	return &t, nil
}

// DeleteTaskTx deletes a task by ID within a transaction
// Deprecated: Not used in production code, only in tests
func (db *DB) DeleteTaskTx(tx *sql.Tx, id string) error {
	if tx == nil {
		return errors.New("nil transaction provided")
	}

	if strings.TrimSpace(id) == "" {
		return ErrInvalidID
	}

	query := `DELETE FROM tasks WHERE id = ?`

	result, err := tx.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task in transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return NewTaskNotFoundError(id)
	}

	return nil
}

// ListTasks retrieves tasks with optional filters
func (db *DB) ListTasks(filter TaskFilter) ([]Task, error) {
	if db == nil || db.conn == nil {
		return nil, ErrNilDB
	}

	// Build query with filters
	query := "SELECT id, milestone, sprint, title, description, status, actor, blocked_by, created, last_updated FROM tasks WHERE 1=1"
	args := []interface{}{}

	if filter.Milestone != "" {
		query += " AND milestone = ?"
		args = append(args, filter.Milestone)
	}

	if filter.Sprint != "" {
		query += " AND sprint = ?"
		args = append(args, filter.Sprint)
	}

	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	if filter.Actor != "" {
		query += " AND actor = ?"
		args = append(args, filter.Actor)
	}

	if filter.ID != "" {
		query += " AND id = ?"
		args = append(args, filter.ID)
	}

	// Apply sorting
	orderBy := getSortOrder(filter.SortBy)
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
		var t Task
		var createdStr string
		var lastUpdatedStr string
		var blockedByStr *string

		err := rows.Scan(
			&t.ID,
			&t.Milestone,
			&t.Sprint,
			&t.Title,
			&t.Description,
			&t.Status,
			&t.Actor,
			&blockedByStr,
			&createdStr,
			&lastUpdatedStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		if blockedByStr != nil && *blockedByStr != "" {
			json.Unmarshal([]byte(*blockedByStr), &t.BlockedBy)
		}

		t.Created, err = time.Parse(time.RFC3339, createdStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created: %w", err)
		}

		t.LastUpdated, err = time.Parse(time.RFC3339, lastUpdatedStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse last_updated: %w", err)
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

// ListTasksTx retrieves tasks with optional filters within a transaction
// Deprecated: Not used in production code, only in tests
func (db *DB) ListTasksTx(tx *sql.Tx, filter TaskFilter) ([]Task, error) {
	if tx == nil {
		return nil, errors.New("nil transaction provided")
	}

	// Build query with filters
	query := "SELECT id, milestone, sprint, title, description, status, actor, blocked_by, created, last_updated FROM tasks WHERE 1=1"
	args := []interface{}{}

	if filter.Milestone != "" {
		query += " AND milestone = ?"
		args = append(args, filter.Milestone)
	}

	if filter.Sprint != "" {
		query += " AND sprint = ?"
		args = append(args, filter.Sprint)
	}

	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	if filter.Actor != "" {
		query += " AND actor = ?"
		args = append(args, filter.Actor)
	}

	if filter.ID != "" {
		query += " AND id = ?"
		args = append(args, filter.ID)
	}

	orderBy := getSortOrder(filter.SortBy)
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

	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks in transaction: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		var createdStr string
		var lastUpdatedStr string
		var blockedByStr *string

		err := rows.Scan(
			&t.ID,
			&t.Milestone,
			&t.Sprint,
			&t.Title,
			&t.Description,
			&t.Status,
			&t.Actor,
			&blockedByStr,
			&createdStr,
			&lastUpdatedStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		if blockedByStr != nil && *blockedByStr != "" {
			json.Unmarshal([]byte(*blockedByStr), &t.BlockedBy)
		}

		t.Created, err = time.Parse(time.RFC3339, createdStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created: %w", err)
		}

		t.LastUpdated, err = time.Parse(time.RFC3339, lastUpdatedStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse last_updated: %w", err)
		}

		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tasks: %w", err)
	}

	if tasks == nil {
		tasks = []Task{}
	}

	return tasks, nil
}

// BeginTx starts a new transaction
func (db *DB) BeginTx() (*sql.Tx, error) {
	if db == nil || db.conn == nil {
		return nil, ErrNilDB
	}

	tx, err := db.conn.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return tx, nil
}

// getSortOrder returns the ORDER BY clause based on the sort field
func getSortOrder(sortBy SortBy) string {
	switch sortBy {
	case SortByStatus:
		return " ORDER BY status ASC"
	case SortByPriority:
		return " ORDER BY priority ASC"
	case SortByMilestone:
		return " ORDER BY milestone ASC, last_updated DESC"
	case SortByCreated:
		return " ORDER BY created DESC"
	case SortByUpdated:
		return " ORDER BY last_updated DESC"
	case SortByID:
		return " ORDER BY id ASC"
	case SortBySprint:
		return " ORDER BY sprint ASC"
	case SortByTitle:
		return " ORDER BY title ASC"
	case SortByDescription:
		return " ORDER BY description ASC"
	case SortByActor:
		return " ORDER BY actor ASC"
	default:
		return " ORDER BY last_updated DESC"
	}
}

// taskExists checks if a task exists in the database
func (db *DB) taskExists(id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM tasks WHERE id = ? LIMIT 1)`
	var exists bool
	err := db.conn.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check task existence: %w", err)
	}
	return exists, nil
}

// taskExistsTx checks if a task exists within a transaction
func (db *DB) taskExistsTx(tx *sql.Tx, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM tasks WHERE id = ? LIMIT 1)`
	var exists bool
	err := tx.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check task existence in transaction: %w", err)
	}
	return exists, nil
}

// UnblockTask transitions a task from 'blocked' to 'todo' status in a single
// atomic database operation. The WHERE clause includes a status = 'blocked'
// guard to prevent unauthorized status transitions. The blocked_by field is
// set to SQL NULL and the last_updated field is refreshed to the current UTC
// timestamp. If a new description is provided, it overwrites the existing
// description; otherwise the description is preserved unchanged.
//
// Returns NewTaskNotFoundError if the task does not exist or is not in
// 'blocked' status (0 rows affected).
func (db *DB) UnblockTask(id string, newDescription *string, now time.Time) error {
	if db == nil || db.conn == nil {
		return ErrNilDB
	}

	if strings.TrimSpace(id) == "" {
		return ErrInvalidID
	}

	// Determine the description value to use.
	// If newDescription is nil, we use a placeholder that preserves the existing value
	// via a CASE expression in the UPDATE.
	var descriptionClause string

	if newDescription != nil && *newDescription != "" {
		// Overwrite with the new description
		descriptionClause = "description = ?, "
	} else {
		// Preserve the existing description - use a CASE expression
		// that sets description to itself (no-op) when no new value is provided.
		// We use a placeholder with a special marker, but since we can't use
		// raw SQL expressions with parameterized queries, we'll use a different
		// approach: construct the SQL dynamically.
		descriptionClause = ""
	}

	// Build the UPDATE query dynamically based on whether description is being updated.
	var query string
	var args []interface{}

	if descriptionClause != "" {
		query = `
			UPDATE tasks
			SET status = 'todo',
			    blocked_by = NULL,
			    ` + descriptionClause + `last_updated = ?
			WHERE id = ? AND status = 'blocked'
		`
		args = []interface{}{
			*newDescription,
			now.Format(time.RFC3339),
			id,
		}
	} else {
		query = `
			UPDATE tasks
			SET status = 'todo',
			    blocked_by = NULL,
			    last_updated = ?
			WHERE id = ? AND status = 'blocked'
		`
		args = []interface{}{
			now.Format(time.RFC3339),
			id,
		}
	}

	result, err := db.conn.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to unblock task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		// Either the task doesn't exist or it's not in 'blocked' status.
		// We return TaskNotFoundError for consistency with other operations.
		return NewTaskNotFoundError(id)
	}

	return nil
}

// UnblockTaskTx transitions a task from 'blocked' to 'todo' status within a
// transaction. This is the transactional variant of UnblockTask.
//
// Deprecated: Not used in production code, only in tests.
func (db *DB) UnblockTaskTx(tx *sql.Tx, id string, newDescription *string, now time.Time) error {
	if db == nil || db.conn == nil {
		return ErrNilDB
	}

	if tx == nil {
		return errors.New("nil transaction provided")
	}

	if strings.TrimSpace(id) == "" {
		return ErrInvalidID
	}

	// Determine the description value to use.
	var descriptionClause string

	if newDescription != nil && *newDescription != "" {
		descriptionClause = "description = ?, "
	} else {
		descriptionClause = ""
	}

	// Build the UPDATE query dynamically based on whether description is being updated.
	var query string
	var args []interface{}

	if descriptionClause != "" {
		query = `
			UPDATE tasks
			SET status = 'todo',
			    blocked_by = NULL,
			    ` + descriptionClause + `last_updated = ?
			WHERE id = ? AND status = 'blocked'
		`
		args = []interface{}{
			*newDescription,
			now.Format(time.RFC3339),
			id,
		}
	} else {
		query = `
			UPDATE tasks
			SET status = 'todo',
			    blocked_by = NULL,
			    last_updated = ?
			WHERE id = ? AND status = 'blocked'
		`
		args = []interface{}{
			now.Format(time.RFC3339),
			id,
		}
	}

	result, err := tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to unblock task in transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return NewTaskNotFoundError(id)
	}

	return nil
}
