package service

import (
	"fmt"
	"strings"

	"github.com/rwbaskette/taskflow/internal/db"
)

// BlockTaskInput contains the input parameters for blocking a task
type BlockTaskInput struct {
	ID     string
	Reason string
}

// BlockTask blocks an existing task with a reason. The reason is appended to
// the task description and recorded in the blocked_by list. The returned
// task reflects the true stored state, including the refreshed LastUpdated
// stamped by the database.
func BlockTask(database *db.DB, input BlockTaskInput) (*db.Task, error) {
	if input.ID == "" {
		return nil, ErrInvalidID
	}

	if strings.TrimSpace(input.Reason) == "" {
		return nil, ErrMissingBlockReason
	}

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

	existingTask.Status = "blocked"
	existingTask.Description = newDescription
	existingTask.BlockedBy = append(existingTask.BlockedBy, input.Reason)

	if err := database.UpdateTask(existingTask); err != nil {
		return nil, err
	}

	return existingTask, nil
}
