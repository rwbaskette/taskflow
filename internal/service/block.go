package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/rwbaskette/taskflow/internal/db"
)

// BlockTaskInput contains the input parameters for blocking a task
type BlockTaskInput struct {
	ID     string
	Reason string
}

// BlockTaskResult contains the result of blocking a task
type BlockTaskResult struct {
	ID          string
	Milestone   string
	Title       string
	Description string
	Actor       string
	Status      string
	BlockedBy   []string
	LastUpdated time.Time
}

// BlockTask blocks an existing task with a reason
func BlockTask(database *db.DB, input BlockTaskInput) (*BlockTaskResult, error) {
	if database == nil {
		return nil, ErrNilDatabase
	}

	// Validate ID is provided
	if input.ID == "" {
		return nil, ErrInvalidID
	}

	// Validate reason is provided
	if strings.TrimSpace(input.Reason) == "" {
		return nil, ErrMissingBlockReason
	}

	// Validate task exists
	existingTask, err := database.ReadTask(input.ID)
	if err != nil {
		return nil, err
	}

	// Append reason to existing description
	newDescription := existingTask.Description
	if newDescription != "" {
		newDescription += "\n"
	}
	newDescription += fmt.Sprintf("[BLOCKED: %s]", input.Reason)

	// Build blocked_by list - add the reason as a blocker entry
	blockedBy := existingTask.BlockedBy
	if blockedBy == nil {
		blockedBy = []string{}
	}
	blockedBy = append(blockedBy, input.Reason)

	// Update task status to blocked and append reason to description
	updatedTask := &db.Task{
		ID:          existingTask.ID,
		Milestone:   existingTask.Milestone,
		Sprint:      existingTask.Sprint,
		Title:       existingTask.Title,
		Description: newDescription,
		Actor:       existingTask.Actor,
		Status:      "blocked",
		BlockedBy:   blockedBy,
		LastUpdated: time.Now().UTC(),
	}

	// Update in database
	if err := database.UpdateTask(updatedTask); err != nil {
		return nil, err
	}

	// Return the result
	return &BlockTaskResult{
		ID:          updatedTask.ID,
		Milestone:   updatedTask.Milestone,
		Title:       updatedTask.Title,
		Description: updatedTask.Description,
		Actor:       updatedTask.Actor,
		Status:      updatedTask.Status,
		BlockedBy:   updatedTask.BlockedBy,
		LastUpdated: updatedTask.LastUpdated,
	}, nil
}
