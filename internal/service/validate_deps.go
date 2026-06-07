package service

import (
	"fmt"

	"github.com/rwbaskette/taskflow/internal/db"
)

// ValidateBlockedBy validates the blocked_by dependencies for a task.
// It checks that:
// 1. All task IDs in blockedBy exist in the database
// 2. A task cannot block itself
// 3. Adding this task as blocked by the listed tasks would NOT create a cycle
//
// The function builds a directed graph where edges go from blocked task -> blocker task.
// When adding blocked_by ["A", "B"] to task "C", it checks if there exists a path
// from "A" to "C" or "B" to "C" in the graph. If such a path exists, adding the
// dependency would create a cycle.
func ValidateBlockedBy(database *db.DB, blockedBy []string, taskID string) error {
	if database == nil {
		return ErrNilDatabase
	}

	// If blockedBy is empty or nil, nothing to validate
	if len(blockedBy) == 0 {
		return nil
	}

	// Check for duplicate entries in blockedBy
	seen := make(map[string]bool)
	for _, blocker := range blockedBy {
		if seen[blocker] {
			return NewDuplicateBlockedByError(blocker)
		}
		seen[blocker] = true
	}

	// Check 1: A task cannot block itself
	for _, blocker := range blockedBy {
		if blocker == taskID {
			return NewCircularDependencyError(taskID, blockedBy)
		}
	}

	// Check 2: All task IDs in blockedBy must exist in the database
	for _, blocker := range blockedBy {
		_, err := database.ReadTask(blocker)
		if err != nil {
			if db.IsTaskNotFound(err) {
				return NewInvalidBlockedByTaskError(blocker)
			}
			return err
		}
	}

	// Check 3: Circular dependency check using DFS
	// Build adjacency list: for each existing task, map to its blockers
	// Edge direction: blocked task -> blocker task
	// So if task A is blocked by ["B", "C"], we have edges A->B and A->C
	adjList := make(map[string][]string)

	// First, get all tasks to build the graph
	allTasks, err := database.ListTasks(db.TaskFilter{})
	if err != nil {
		return err
	}

	for _, task := range allTasks {
		// Skip the task being updated/created to avoid false positives
		if task.ID == taskID {
			continue
		}
		// Add edges from this task to its blockers
		for _, blocker := range task.BlockedBy {
			// Verify the blocker exists before adding to graph
			_, err := database.ReadTask(blocker)
			if err != nil {
				if db.IsTaskNotFound(err) {
					continue // Skip invalid references
				}
				return err
			}
			adjList[task.ID] = append(adjList[task.ID], blocker)
		}
	}

	// Add the new edges that would be created
	// The new task (taskID) would have edges to each blocker in blockedBy
	// This means: taskID -> blocker (for each blocker in blockedBy)
	for _, blocker := range blockedBy {
		adjList[taskID] = append(adjList[taskID], blocker)
	}

	// Now check if adding the edges would create a cycle
	// For each blocker in blockedBy, check if there's a path from blocker to taskID
	// If blocker can reach taskID, and taskID points to blocker, that's a cycle
	for _, blocker := range blockedBy {
		if hasPathDFS(adjList, blocker, taskID) {
			return NewCircularDependencyError(taskID, blockedBy)
		}
	}

	return nil
}

// hasPathDFS checks if there exists a path from source to target in the graph
// using Depth-First Search. It detects if adding an edge would create a cycle.
func hasPathDFS(adjList map[string][]string, source, target string) bool {
	if source == target {
		return true
	}

	visited := make(map[string]bool)
	return dfsVisit(adjList, source, target, visited)
}

// dfsVisit performs DFS from node to see if we can reach target
func dfsVisit(adjList map[string][]string, current, target string, visited map[string]bool) bool {
	if current == target {
		return true
	}

	visited[current] = true

	neighbors, exists := adjList[current]
	if !exists {
		return false
	}

	for _, neighbor := range neighbors {
		if visited[neighbor] {
			continue
		}
		if dfsVisit(adjList, neighbor, target, visited) {
			return true
		}
	}

	return false
}

// ValidateBlockedByForUpdate validates blocked_by when updating an existing task.
// This is similar to ValidateBlockedBy but also considers the task's existing blockers.
func ValidateBlockedByForUpdate(database *db.DB, blockedBy []string, taskID string) error {
	if database == nil {
		return ErrNilDatabase
	}

	// If blockedBy is empty or nil, nothing to validate
	if len(blockedBy) == 0 {
		return nil
	}

	// Check for duplicate entries in blockedBy
	seen := make(map[string]bool)
	for _, blocker := range blockedBy {
		if seen[blocker] {
			return NewDuplicateBlockedByError(blocker)
		}
		seen[blocker] = true
	}

	// Check 1: A task cannot block itself
	for _, blocker := range blockedBy {
		if blocker == taskID {
			return NewCircularDependencyError(taskID, blockedBy)
		}
	}

	// Check 2: All task IDs in blockedBy must exist in the database
	for _, blocker := range blockedBy {
		_, err := database.ReadTask(blocker)
		if err != nil {
			if db.IsTaskNotFound(err) {
				return NewInvalidBlockedByTaskError(blocker)
			}
			return err
		}
	}

	// Check 3: Circular dependency check
	// Build adjacency list from all existing tasks except the one being updated
	adjList := make(map[string][]string)

	allTasks, err := database.ListTasks(db.TaskFilter{})
	if err != nil {
		return err
	}

	for _, task := range allTasks {
		// Skip the task being updated
		if task.ID == taskID {
			continue
		}
		// Add edges from this task to its blockers
		for _, blocker := range task.BlockedBy {
			// Verify the blocker exists
			_, err := database.ReadTask(blocker)
			if err != nil {
				if db.IsTaskNotFound(err) {
					continue // Skip invalid references
				}
				return err
			}
			adjList[task.ID] = append(adjList[task.ID], blocker)
		}
	}

	// Get the existing task's current blocked_by
	existingTask, err := database.ReadTask(taskID)
	if err != nil {
		return err
	}

	// Build the new graph after the update:
	// The updated task will have the new blockedBy list
	// So edges will go from taskID -> each blocker in blockedBy
	for _, blocker := range blockedBy {
		adjList[taskID] = append(adjList[taskID], blocker)
	}

	// For each potential new blocker, check if there's already a path from
	// that blocker to taskID (which would create a cycle)
	for _, blocker := range blockedBy {
		if hasPathDFS(adjList, blocker, taskID) {
			return NewCircularDependencyError(taskID, blockedBy)
		}
	}

	// Also check that no existing task that taskID blocks would create a cycle
	// If taskID currently blocks ["X", "Y"], we need to ensure X and Y don't
	// have paths to any of the new blockers
	for _, existingBlocker := range existingTask.BlockedBy {
		for _, newBlocker := range blockedBy {
			// Check if newBlocker can reach existingBlocker
			if hasPathDFS(adjList, newBlocker, existingBlocker) {
				return fmt.Errorf("circular dependency detected: %s is blocked by %s which blocks %s", taskID, newBlocker, existingBlocker)
			}
		}
	}

	return nil
}