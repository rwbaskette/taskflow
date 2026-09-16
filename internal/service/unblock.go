package service

import (
	"errors"

	"github.com/rwbaskette/taskflow/internal/db"
	cliErrors "github.com/rwbaskette/taskflow/internal/errors"
)

// UnblockTask unblocks a previously blocked task, transitioning it from
// 'blocked' back to 'todo' status. It clears the blocked_by field and
// optionally overwrites the description (when a non-empty description is
// given). The database-level UPDATE includes a WHERE status = 'blocked'
// guard as defense-in-depth against race conditions, and the returned task
// reflects the true stored state.
func UnblockTask(database *db.DB, id, description string) (*db.Task, error) {
	if id == "" {
		return nil, ErrInvalidID
	}

	var newDescription *string
	if description != "" {
		newDescription = &description
	}

	task, err := database.UnblockTask(id, newDescription)
	if err != nil {
		var notFound *db.TaskNotFoundError
		if errors.As(err, &notFound) {
			return nil, cliErrors.ResourceNotFoundError("task", id)
		}
		var notBlocked *db.TaskNotBlockedError
		if errors.As(err, &notBlocked) {
			return nil, cliErrors.InvalidStatusTransitionError(id, notBlocked.Status)
		}
		return nil, err
	}

	return task, nil
}
