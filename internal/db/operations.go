package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Task represents a task in the database
type Task struct {
	ID          string    `json:"id"`
	Milestone   string    `json:"milestone,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	Actor       string    `json:"actor,omitempty"`
	LastUpdated time.Time `json:"last_updated"`
}

// TaskFilter contains optional filters for listing tasks
type TaskFilter struct {
	Milestone string
	Status    string
	Actor     string
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

	// Set LastUpdated to now if not set
	if t.LastUpdated.IsZero() {
		t.LastUpdated = time.Now().UTC()
	}

	query := `
		INSERT INTO tasks (id, milestone, title, description, status, actor, last_updated)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = db.conn.Exec(query,
		t.ID,
		t.Milestone,
		t.Title,
		t.Description,
		t.Status,
		t.Actor,
		t.LastUpdated.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	return nil
}

// CreateTaskTx creates a new task within a transaction
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

	// Set LastUpdated to now if not set
	if t.LastUpdated.IsZero() {
		t.LastUpdated = time.Now().UTC()
	}

	query := `
		INSERT INTO tasks (id, milestone, title, description, status, actor, last_updated)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = tx.Exec(query,
		t.ID,
		t.Milestone,
		t.Title,
		t.Description,
		t.Status,
		t.Actor,
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
		SELECT id, milestone, title, description, status, actor, last_updated
		FROM tasks
		WHERE id = ?
	`

	var t Task
	var lastUpdatedStr string

	err := db.conn.QueryRow(query, id).Scan(
		&t.ID,
		&t.Milestone,
		&t.Title,
		&t.Description,
		&t.Status,
		&t.Actor,
		&lastUpdatedStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewTaskNotFoundError(id)
		}
		return nil, fmt.Errorf("failed to read task: %w", err)
	}

	t.LastUpdated, err = time.Parse(time.RFC3339, lastUpdatedStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse last_updated: %w", err)
	}

	return &t, nil
}

// ReadTaskTx retrieves a task by ID within a transaction
func (db *DB) ReadTaskTx(tx *sql.Tx, id string) (*Task, error) {
	if tx == nil {
		return nil, errors.New("nil transaction provided")
	}

	if strings.TrimSpace(id) == "" {
		return nil, ErrInvalidID
	}

	query := `
		SELECT id, milestone, title, description, status, actor, last_updated
		FROM tasks
		WHERE id = ?
	`

	var t Task
	var lastUpdatedStr string

	err := tx.QueryRow(query, id).Scan(
		&t.ID,
		&t.Milestone,
		&t.Title,
		&t.Description,
		&t.Status,
		&t.Actor,
		&lastUpdatedStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewTaskNotFoundError(id)
		}
		return nil, fmt.Errorf("failed to read task in transaction: %w", err)
	}

	t.LastUpdated, err = time.Parse(time.RFC3339, lastUpdatedStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse last_updated: %w", err)
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

	query := `
		UPDATE tasks
		SET milestone = ?, title = ?, description = ?, status = ?, actor = ?, last_updated = ?
		WHERE id = ?
	`

	result, err := db.conn.Exec(query,
		t.Milestone,
		t.Title,
		t.Description,
		t.Status,
		t.Actor,
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
func (db *DB) UpdateTaskTx(tx *sql.Tx, t *Task) error {
	if tx == nil {
		return errors.New("nil transaction provided")
	}

	if err := validateTask(t); err != nil {
		return err
	}

	// Always update LastUpdated to current time
	t.LastUpdated = time.Now().UTC()

	query := `
		UPDATE tasks
		SET milestone = ?, title = ?, description = ?, status = ?, actor = ?, last_updated = ?
		WHERE id = ?
	`

	result, err := tx.Exec(query,
		t.Milestone,
		t.Title,
		t.Description,
		t.Status,
		t.Actor,
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

// DeleteTaskTx deletes a task by ID within a transaction
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
	query := "SELECT id, milestone, title, description, status, actor, last_updated FROM tasks WHERE 1=1"
	args := []interface{}{}

	if filter.Milestone != "" {
		query += " AND milestone = ?"
		args = append(args, filter.Milestone)
	}

	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	if filter.Actor != "" {
		query += " AND actor = ?"
		args = append(args, filter.Actor)
	}

	// Order by last_updated descending (most recent first)
	query += " ORDER BY last_updated DESC"

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
		var lastUpdatedStr string

		err := rows.Scan(
			&t.ID,
			&t.Milestone,
			&t.Title,
			&t.Description,
			&t.Status,
			&t.Actor,
			&lastUpdatedStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
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
func (db *DB) ListTasksTx(tx *sql.Tx, filter TaskFilter) ([]Task, error) {
	if tx == nil {
		return nil, errors.New("nil transaction provided")
	}

	// Build query with filters
	query := "SELECT id, milestone, title, description, status, actor, last_updated FROM tasks WHERE 1=1"
	args := []interface{}{}

	if filter.Milestone != "" {
		query += " AND milestone = ?"
		args = append(args, filter.Milestone)
	}

	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	if filter.Actor != "" {
		query += " AND actor = ?"
		args = append(args, filter.Actor)
	}

	query += " ORDER BY last_updated DESC"

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
		var lastUpdatedStr string

		err := rows.Scan(
			&t.ID,
			&t.Milestone,
			&t.Title,
			&t.Description,
			&t.Status,
			&t.Actor,
			&lastUpdatedStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
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
