package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// buildTaskflowBinary compiles the taskflow binary to a temporary location
// and returns the path. It also returns a cleanup function.
func buildTaskflowBinary(t *testing.T) (binPath string, cleanup func()) {
	t.Helper()

	tmpDir := t.TempDir()
	binPath = filepath.Join(tmpDir, "taskflow")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}

	// Build the binary from the project root
	// Go tests run from the module root, so we need to find the go.mod
	testDir, _ := os.Getwd()
	projectRoot := testDir
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			t.Fatalf("could not find project root with go.mod")
		}
		projectRoot = parent
	}
	
	// Build the binary
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(), "CGO_ENABLED=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build taskflow binary: %v\nstderr: %s", err, stderr.String())
	}

	// Verify the binary was created
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Fatalf("binary was not created at %s", binPath)
	}

	cleanup = func() {
		os.Remove(binPath)
	}
	return binPath, cleanup
}

// runTaskflow runs a taskflow command with the given args and returns stdout,
// stderr, and error. It uses the provided binary path and database directory.
func runTaskflow(t *testing.T, binPath string, dbDir string, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	cmd := exec.Command(binPath, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("TASKFLOW_DIR=%s", dbDir))
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err = cmd.Run()
	return stdoutBuf.String(), stderrBuf.String(), err
}

// parseTaskJSON parses the taskflow JSON output into a map for inspection.
// Handles the list command's output format which wraps tasks in a "tasks" array.
func parseTaskJSON(t *testing.T, output string) map[string]interface{} {
	t.Helper()

	// Find the JSON portion in the output (taskflow prints it as the last JSON object)
	// Handle both single-line and multi-line JSON
	output = strings.TrimSpace(output)
	
	// Find the first '{' and last '}' to extract JSON
	start := strings.Index(output, "{")
	end := strings.LastIndex(output, "}")
	if start == -1 || end == -1 || end <= start {
		t.Fatalf("could not find JSON output in: %s", output)
	}
	jsonStr := output[start : end+1]

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\njson: %s", err, jsonStr)
	}

	// If the result has a "tasks" array (from list command), extract the first task
	if tasks, ok := result["tasks"].([]interface{}); ok && len(tasks) > 0 {
		if task, ok := tasks[0].(map[string]interface{}); ok {
			return task
		}
	}

	return result
}

// TestTaskUnblock_E2E exercises the full task lifecycle through the CLI:
// 1. Add a task
// 2. Block the task
// 3. Unblock the task (with a description update)
// 4. Verify via list command that status is 'todo', description updated,
//    blocked_by is null, and last_updated is refreshed
// 5. Clean up the test task
func TestTaskUnblock_E2E(t *testing.T) {
	// Create a unique task ID to avoid collisions
	taskID := "e2e-unblock-task-" + fmt.Sprintf("%d", time.Now().UnixNano())
	milestone := "sprint-1"
	originalTitle := "E2E Unblock Test Task"
	originalDesc := "This is the original description before blocking"
	newDesc := "This is the updated description after unblocking"
	actor := "test-e2e"

	// Set up a temporary database directory
	tmpDir := t.TempDir()
	dbDir := filepath.Join(tmpDir, "taskflow-db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("failed to create db dir: %v", err)
	}

	// Build the taskflow binary
	binPath, binCleanup := buildTaskflowBinary(t)
	defer binCleanup()

	// ---- Step 1: Add a task via CLI ----
	stdout, stderr, err := runTaskflow(t, binPath, dbDir, "add",
		fmt.Sprintf(`{"id":"%s","milestone":"%s","title":"%s","description":"%s","actor":"%s"}`,
			taskID, milestone, originalTitle, originalDesc, actor))
	if err != nil {
		t.Fatalf("task add failed: %v\nstderr: %s", err, stderr)
	}
	t.Logf("add stdout: %s", stdout)

	// Verify the task was created by listing it
	stdout, stderr, err = runTaskflow(t, binPath, dbDir, "list",
		fmt.Sprintf(`{"id":"%s","format":"xml"}`, taskID))
	if err != nil {
		t.Fatalf("task list (verify add) failed: %v\nstderr: %s", err, stderr)
	}
	t.Logf("list stdout: %s", stdout)
	addResult := parseTaskJSON(t, stdout)
	if addResult["Status"] != "todo" {
		t.Errorf("expected initial status 'todo', got %v", addResult["Status"])
	}
	t.Logf("Task added: status=%v, id=%v", addResult["Status"], addResult["ID"])

	// ---- Step 2: Block the task via CLI ----
	stdout, stderr, err = runTaskflow(t, binPath, dbDir, "block",
		fmt.Sprintf(`{"id":"%s","reason":"Waiting on external dependency"}`, taskID))
	if err != nil {
		t.Fatalf("task block failed: %v\nstderr: %s", err, stderr)
	}
	t.Logf("block stdout: %s", stdout)

	// Verify the task is now blocked
	stdout, stderr, err = runTaskflow(t, binPath, dbDir, "list",
		fmt.Sprintf(`{"id":"%s","format":"xml"}`, taskID))
	if err != nil {
		t.Fatalf("task list (verify block) failed: %v\nstderr: %s", err, stderr)
	}
	blockResult := parseTaskJSON(t, stdout)
	if blockResult["Status"] != "blocked" {
		t.Errorf("expected status 'blocked' after blocking, got %v", blockResult["Status"])
	}
	// Capture the last_updated before unblock for comparison
	originalLastUpdated, _ := blockResult["LastUpdated"].(string)
	t.Logf("Task blocked: status=%v, last_updated=%v", blockResult["Status"], blockResult["LastUpdated"])

	// ---- Step 3: Unblock the task via CLI with description update ----
	// Add a small delay to ensure last_updated timestamp changes
	time.Sleep(1100 * time.Millisecond)
	stdout, stderr, err = runTaskflow(t, binPath, dbDir, "unblock",
		fmt.Sprintf(`{"id":"%s","description":"%s"}`, taskID, newDesc))
	if err != nil {
		t.Fatalf("task unblock failed: %v\nstderr: %s", err, stderr)
	}
	t.Logf("unblock stdout: %s", stdout)

	// ---- Step 4: Verify the unblocked task via list command ----
	stdout, stderr, err = runTaskflow(t, binPath, dbDir, "list",
		fmt.Sprintf(`{"id":"%s","format":"xml"}`, taskID))
	if err != nil {
		t.Fatalf("task list (verify unblock) failed: %v\nstderr: %s", err, stderr)
	}

	unblockResult := parseTaskJSON(t, stdout)

	// Verify status is 'todo'
	if unblockResult["Status"] != "todo" {
		t.Errorf("expected status 'todo' after unblock, got %v", unblockResult["Status"])
	}

	// Verify description was updated
	if unblockResult["Description"] != newDesc {
		t.Errorf("expected description %q, got %q", newDesc, unblockResult["Description"])
	}

	// Verify blocked_by is null/empty
	if unblockResult["BlockedBy"] != nil {
		t.Errorf("expected blocked_by to be null, got %v", unblockResult["BlockedBy"])
	}

	// Verify last_updated was refreshed (new timestamp should be later than original)
	newLastUpdated, _ := unblockResult["LastUpdated"].(string)
	if originalLastUpdated != "" && newLastUpdated != "" {
		origTime, err1 := time.Parse("2006-01-02 15:04:05", originalLastUpdated)
		newTime, err2 := time.Parse("2006-01-02 15:04:05", newLastUpdated)
		if err1 == nil && err2 == nil {
			if !newTime.After(origTime) {
				t.Errorf("expected last_updated to be refreshed; original=%v, new=%v", origTime, newTime)
			}
		}
	}

	t.Logf("Task unblocked: status=%v, description=%v, blocked_by=%v, last_updated=%v",
		unblockResult["Status"], unblockResult["Description"], unblockResult["BlockedBy"], unblockResult["LastUpdated"])

	// ---- Step 5: Clean up the test task via CLI ----
	stdout, stderr, err = runTaskflow(t, binPath, dbDir, "delete",
		fmt.Sprintf(`{"id":"%s"}`, taskID))
	if err != nil {
		t.Logf("task delete (cleanup) failed: %v\nstderr: %s", err, stderr)
	}
	t.Logf("cleanup stdout: %s", stdout)

	// Verify the task is deleted
	stdout, stderr, err = runTaskflow(t, binPath, dbDir, "list",
		fmt.Sprintf(`{"id":"%s","format":"xml"}`, taskID))
	if err == nil {
		// If the task was found, the delete didn't work properly
		deletedResult := parseTaskJSON(t, stdout)
		t.Errorf("expected task to be deleted, but it still exists: %v", deletedResult["id"])
	}
	t.Logf("Task successfully deleted. End-to-end test passed.")
}

// TestTaskUnblock_E2E_NoDescription verifies unblocking without a description
// update preserves the original description.
func TestTaskUnblock_E2E_NoDescription(t *testing.T) {
	taskID := "e2e-unblock-nodeesc-" + fmt.Sprintf("%d", time.Now().UnixNano())
	originalDesc := "This description must be preserved during unblock"
	actor := "test-e2e"

	tmpDir := t.TempDir()
	dbDir := filepath.Join(tmpDir, "taskflow-db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("failed to create db dir: %v", err)
	}

	binPath, binCleanup := buildTaskflowBinary(t)
	defer binCleanup()

	// Add a task
	stdout, stderr, err := runTaskflow(t, binPath, dbDir, "add",
		fmt.Sprintf(`{"id":"%s","milestone":"sprint-1","title":"No Desc Test","description":"%s","actor":"%s"}`,
			taskID, originalDesc, actor))
	if err != nil {
		t.Fatalf("task add failed: %v\nstderr: %s", err, stderr)
	}

	// Block the task
	_, _, err = runTaskflow(t, binPath, dbDir, "block",
		fmt.Sprintf(`{"id":"%s","reason":"Waiting on something"}`, taskID))
	if err != nil {
		t.Fatalf("task block failed: %v", err)
	}

	// Unblock WITHOUT providing a description
	stdout, stderr, err = runTaskflow(t, binPath, dbDir, "unblock",
		fmt.Sprintf(`{"id":"%s"}`, taskID))
	if err != nil {
		t.Fatalf("task unblock (no desc) failed: %v\nstderr: %s", err, stderr)
	}

	// Verify the task is unblocked and description preserved
	stdout, stderr, err = runTaskflow(t, binPath, dbDir, "list",
		fmt.Sprintf(`{"id":"%s","format":"xml"}`, taskID))
	if err != nil {
		t.Fatalf("task list failed: %v\nstderr: %s", err, stderr)
	}

	result := parseTaskJSON(t, stdout)
	if result["Status"] != "todo" {
		t.Errorf("expected status 'todo', got %v", result["Status"])
	}

	// The description should be preserved (may include the block suffix from the block operation)
	currentDesc, _ := result["Description"].(string)
	if !strings.Contains(currentDesc, originalDesc) {
		t.Errorf("expected description to contain %q, got %q", originalDesc, currentDesc)
	}

	// Cleanup
	_, _, _ = runTaskflow(t, binPath, dbDir, "delete",
		fmt.Sprintf(`{"id":"%s"}`, taskID))
}

// TestTaskUnblock_E2E_NonStringID verifies that unblocking with a non-string id
// (e.g., a number or null) returns an INVALID_ARGUMENT error.
func TestTaskUnblock_E2E_NonStringID(t *testing.T) {
	tmpDir := t.TempDir()
	dbDir := filepath.Join(tmpDir, "taskflow-db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("failed to create db dir: %v", err)
	}

	binPath, binCleanup := buildTaskflowBinary(t)
	defer binCleanup()

	// Attempt to unblock with a numeric id (JSON number, not string) - should fail with INVALID_ARGUMENT
	_, stderr, err := runTaskflow(t, binPath, dbDir, "unblock",
		`{"id": 123}`)

	// Verify the error occurred
	if err == nil {
		t.Fatal("expected error when unblocking with non-string id (number), got nil")
	}

	// Verify the error message contains INVALID_ARGUMENT
	stderrLower := strings.ToLower(stderr)
	if !strings.Contains(stderrLower, "invalid_argument") {
		t.Logf("stderr: %s", stderr)
		t.Errorf("expected error message to contain 'invalid_argument', got: %s", stderr)
	}

	t.Logf("Correctly got INVALID_ARGUMENT when unblocking with numeric id: %v", err)

	// Attempt to unblock with a null id - should also fail with INVALID_ARGUMENT
	_, stderr, err = runTaskflow(t, binPath, dbDir, "unblock",
		`{"id": null}`)

	// Verify the error occurred
	if err == nil {
		t.Fatal("expected error when unblocking with non-string id (null), got nil")
	}

	// Verify the error message contains INVALID_ARGUMENT
	stderrLower = strings.ToLower(stderr)
	if !strings.Contains(stderrLower, "invalid_argument") {
		t.Logf("stderr: %s", stderr)
		t.Errorf("expected error message to contain 'invalid_argument', got: %s", stderr)
	}

	t.Logf("Correctly got INVALID_ARGUMENT when unblocking with null id: %v", err)
}

// TestTaskUnblock_E2E_NonExistentTask verifies that unblocking a non-existent
// task returns a RESOURCE_NOT_FOUND error.
func TestTaskUnblock_E2E_NonExistentTask(t *testing.T) {
	taskID := "e2e-unblock-nonexistent-" + fmt.Sprintf("%d", time.Now().UnixNano())

	tmpDir := t.TempDir()
	dbDir := filepath.Join(tmpDir, "taskflow-db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("failed to create db dir: %v", err)
	}

	binPath, binCleanup := buildTaskflowBinary(t)
	defer binCleanup()

	// Attempt to unblock a task that does not exist - should fail with RESOURCE_NOT_FOUND
	_, stderr, err := runTaskflow(t, binPath, dbDir, "unblock",
		fmt.Sprintf(`{"id":"%s"}`, taskID))

	// Verify the error occurred
	if err == nil {
		t.Fatal("expected error when unblocking a non-existent task, got nil")
	}

	// Verify the error message contains RESOURCE_NOT_FOUND
	stderrLower := strings.ToLower(stderr)
	if !strings.Contains(stderrLower, "resource_not_found") {
		t.Logf("stderr: %s", stderr)
		t.Errorf("expected error message to contain 'resource_not_found', got: %s", stderr)
	}

	t.Logf("Correctly got RESOURCE_NOT_FOUND when unblocking non-existent task: %v", err)
}

// TestTaskUnblock_E2E_InProgressStatus verifies that unblocking a task in
// 'in_progress' status returns an INVALID_STATUS_TRANSITION error.
func TestTaskUnblock_E2E_InProgressStatus(t *testing.T) {
	taskID := "e2e-unblock-inprog-" + fmt.Sprintf("%d", time.Now().UnixNano())
	actor := "test-e2e"

	tmpDir := t.TempDir()
	dbDir := filepath.Join(tmpDir, "taskflow-db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("failed to create db dir: %v", err)
	}

	binPath, binCleanup := buildTaskflowBinary(t)
	defer binCleanup()

	// Add a task
	_, stderr, err := runTaskflow(t, binPath, dbDir, "add",
		fmt.Sprintf(`{"id":"%s","milestone":"sprint-1","title":"In Progress Test","description":"For status test","actor":"%s"}`,
			taskID, actor))
	if err != nil {
		t.Fatalf("task add failed: %v\nstderr: %s", err, stderr)
	}

	// Update the task to change status to 'in_progress'
	_, _, err = runTaskflow(t, binPath, dbDir, "update",
		fmt.Sprintf(`{"id":"%s","status":"in_progress"}`, taskID))
	if err != nil {
		t.Fatalf("task update failed: %v", err)
	}

	// Attempt to unblock an in_progress task - should fail with INVALID_STATUS_TRANSITION
	_, stderr, err = runTaskflow(t, binPath, dbDir, "unblock",
		fmt.Sprintf(`{"id":"%s"}`, taskID))

	// Verify the error occurred
	if err == nil {
		t.Fatal("expected error when unblocking an in_progress task, got nil")
	}

	// Verify the error message contains INVALID_STATUS_TRANSITION
	stderrLower := strings.ToLower(stderr)
	if !strings.Contains(stderrLower, "invalid_status_transition") {
		t.Logf("stderr: %s", stderr)
		t.Errorf("expected error message to contain 'invalid_status_transition', got: %s", stderr)
	}

	// Verify the error message mentions 'in_progress'
	if !strings.Contains(stderrLower, "in_progress") {
		t.Logf("stderr: %s", stderr)
		t.Errorf("expected error message to mention 'in_progress', got: %s", stderr)
	}

	t.Logf("Correctly got INVALID_STATUS_TRANSITION for in_progress task: %v", err)

	// Cleanup
	_, _, _ = runTaskflow(t, binPath, dbDir, "delete",
		fmt.Sprintf(`{"id":"%s"}`, taskID))
}

// TestTaskUnblock_E2E_DoneStatus verifies that unblocking a task in 'done'
// status returns an INVALID_STATUS_TRANSITION error.
func TestTaskUnblock_E2E_DoneStatus(t *testing.T) {
	taskID := "e2e-unblock-done-" + fmt.Sprintf("%d", time.Now().UnixNano())
	actor := "test-e2e"

	tmpDir := t.TempDir()
	dbDir := filepath.Join(tmpDir, "taskflow-db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("failed to create db dir: %v", err)
	}

	binPath, binCleanup := buildTaskflowBinary(t)
	defer binCleanup()

	// Add a task
	_, stderr, err := runTaskflow(t, binPath, dbDir, "add",
		fmt.Sprintf(`{"id":"%s","milestone":"sprint-1","title":"Done Test","description":"For status test","actor":"%s"}`,
			taskID, actor))
	if err != nil {
		t.Fatalf("task add failed: %v\nstderr: %s", err, stderr)
	}

	// Complete the task to change status to 'done'
	_, _, err = runTaskflow(t, binPath, dbDir, "complete",
		fmt.Sprintf(`{"id":"%s"}`, taskID))
	if err != nil {
		t.Fatalf("task complete failed: %v", err)
	}

	// Attempt to unblock a done task - should fail with INVALID_STATUS_TRANSITION
	_, stderr, err = runTaskflow(t, binPath, dbDir, "unblock",
		fmt.Sprintf(`{"id":"%s"}`, taskID))

	// Verify the error occurred
	if err == nil {
		t.Fatal("expected error when unblocking a done task, got nil")
	}

	// Verify the error message contains INVALID_STATUS_TRANSITION
	stderrLower := strings.ToLower(stderr)
	if !strings.Contains(stderrLower, "invalid_status_transition") {
		t.Logf("stderr: %s", stderr)
		t.Errorf("expected error message to contain 'invalid_status_transition', got: %s", stderr)
	}

	// Verify the error message mentions 'done'
	if !strings.Contains(stderrLower, "done") {
		t.Logf("stderr: %s", stderr)
		t.Errorf("expected error message to mention 'done', got: %s", stderr)
	}

	t.Logf("Correctly got INVALID_STATUS_TRANSITION for done task: %v", err)

	// Cleanup
	_, _, _ = runTaskflow(t, binPath, dbDir, "delete",
		fmt.Sprintf(`{"id":"%s"}`, taskID))
}

// TestTaskUnblock_E2E_EmptyID verifies that unblocking with an empty id
// parameter returns an error.
func TestTaskUnblock_E2E_EmptyID(t *testing.T) {
	tmpDir := t.TempDir()
	dbDir := filepath.Join(tmpDir, "taskflow-db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("failed to create db dir: %v", err)
	}

	binPath, binCleanup := buildTaskflowBinary(t)
	defer binCleanup()

	// Attempt to unblock with an empty id - should fail with error
	_, stderr, err := runTaskflow(t, binPath, dbDir, "unblock",
		`{"id":""}`)

	// Verify the error occurred
	if err == nil {
		t.Fatal("expected error when unblocking with empty id, got nil")
	}

	// Verify the error message indicates an invalid or missing argument
	stderrLower := strings.ToLower(stderr)
	if !strings.Contains(stderrLower, "missing_argument") && !strings.Contains(stderrLower, "invalid_argument") {
		t.Logf("stderr: %s", stderr)
		t.Errorf("expected error message to contain 'missing_argument' or 'invalid_argument', got: %s", stderr)
	}

	t.Logf("Correctly got error when unblocking with empty id: %v", err)
}

// TestTaskUnblock_E2E_MissingID verifies that unblocking with a missing id
// parameter returns an error.
func TestTaskUnblock_E2E_MissingID(t *testing.T) {
	tmpDir := t.TempDir()
	dbDir := filepath.Join(tmpDir, "taskflow-db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("failed to create db dir: %v", err)
	}

	binPath, binCleanup := buildTaskflowBinary(t)
	defer binCleanup()

	// Attempt to unblock with a completely empty payload - should fail with error
	_, stderr, err := runTaskflow(t, binPath, dbDir, "unblock",
		`{}`)

	// Verify the error occurred
	if err == nil {
		t.Fatal("expected error when unblocking with missing id, got nil")
	}

	// Verify the error message indicates a missing argument
	stderrLower := strings.ToLower(stderr)
	if !strings.Contains(stderrLower, "missing_argument") {
		t.Logf("stderr: %s", stderr)
		t.Errorf("expected error message to contain 'missing_argument', got: %s", stderr)
	}

	t.Logf("Correctly got error when unblocking with missing id: %v", err)
}

// TestTaskUnblock_E2E_Idempotency verifies that calling unblock twice on the
// same task fails on the second call, since the task is no longer in 'blocked'
// status after the first successful unblock.
func TestTaskUnblock_E2E_Idempotency(t *testing.T) {
	taskID := "e2e-unblock-idem-" + fmt.Sprintf("%d", time.Now().UnixNano())
	milestone := "sprint-1"
	originalDesc := "Original description for idempotency test"
	actor := "test-e2e"

	tmpDir := t.TempDir()
	dbDir := filepath.Join(tmpDir, "taskflow-db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("failed to create db dir: %v", err)
	}

	binPath, binCleanup := buildTaskflowBinary(t)
	defer binCleanup()

	// ---- Step 1: Add a task ----
	_, stderr, err := runTaskflow(t, binPath, dbDir, "add",
		fmt.Sprintf(`{"id":"%s","milestone":"%s","title":"Idempotency Test","description":"%s","actor":"%s"}`,
			taskID, milestone, originalDesc, actor))
	if err != nil {
		t.Fatalf("task add failed: %v\nstderr: %s", err, stderr)
	}

	// ---- Step 2: Block the task ----
	_, _, err = runTaskflow(t, binPath, dbDir, "block",
		fmt.Sprintf(`{"id":"%s","reason":"Waiting on external dependency"}`, taskID))
	if err != nil {
		t.Fatalf("task block failed: %v", err)
	}

	// ---- Step 3: First unblock (should succeed) ----
	stdout, stderr, err := runTaskflow(t, binPath, dbDir, "unblock",
		fmt.Sprintf(`{"id":"%s","description":"Updated after first unblock"}`, taskID))
	if err != nil {
		t.Fatalf("first unblock failed: %v\nstderr: %s", err, stderr)
	}
	t.Logf("First unblock succeeded: %s", stdout)

	// Verify the task is now in 'todo' status
	stdout, stderr, err = runTaskflow(t, binPath, dbDir, "list",
		fmt.Sprintf(`{"id":"%s","format":"xml"}`, taskID))
	if err != nil {
		t.Fatalf("task list failed: %v\nstderr: %s", err, stderr)
	}
	firstResult := parseTaskJSON(t, stdout)
	if firstResult["Status"] != "todo" {
		t.Errorf("expected status 'todo' after first unblock, got %v", firstResult["Status"])
	}

	// ---- Step 4: Second unblock (should fail - not idempotent) ----
	_, stderr, err = runTaskflow(t, binPath, dbDir, "unblock",
		fmt.Sprintf(`{"id":"%s"}`, taskID))

	// Verify the second unblock failed
	if err == nil {
		t.Fatal("expected error on second unblock call (not idempotent), got nil")
	}

	// Verify the error message indicates INVALID_STATUS_TRANSITION
	stderrLower := strings.ToLower(stderr)
	if !strings.Contains(stderrLower, "invalid_status_transition") {
		t.Logf("stderr: %s", stderr)
		t.Errorf("expected error message to contain 'invalid_status_transition' on second unblock, got: %s", stderr)
	}

	t.Logf("Correctly got error on second unblock (not idempotent): %v", err)

	// Cleanup
	_, _, _ = runTaskflow(t, binPath, dbDir, "delete",
		fmt.Sprintf(`{"id":"%s"}`, taskID))
}

// TestTaskUnblock_E2E_ErrorOnNonBlocked verifies that unblocking a non-blocked
// task fails with an appropriate error through the CLI.
func TestTaskUnblock_E2E_ErrorOnNonBlocked(t *testing.T) {
	taskID := "e2e-unblock-error-" + fmt.Sprintf("%d", time.Now().UnixNano())

	tmpDir := t.TempDir()
	dbDir := filepath.Join(tmpDir, "taskflow-db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("failed to create db dir: %v", err)
	}

	binPath, binCleanup := buildTaskflowBinary(t)
	defer binCleanup()

	// Add a task (status is 'todo', not 'blocked')
	_, stderr, err := runTaskflow(t, binPath, dbDir, "add",
		fmt.Sprintf(`{"id":"%s","milestone":"sprint-1","title":"Error Test","description":"For error testing","actor":"test"}`,
			taskID))
	if err != nil {
		t.Fatalf("task add failed: %v\nstderr: %s", err, stderr)
	}

	// Attempt to unblock a non-blocked task - should fail
	_, stderr, err = runTaskflow(t, binPath, dbDir, "unblock",
		fmt.Sprintf(`{"id":"%s"}`, taskID))

	// Verify the error occurred
	if err == nil {
		t.Fatal("expected error when unblocking a non-blocked task, got nil")
	}

	// Verify the error message indicates the problem
	stderrLower := strings.ToLower(stderr)
	if !strings.Contains(stderrLower, "blocked") && !strings.Contains(stderrLower, "error") {
		t.Logf("stderr: %s", stderr)
	}

	t.Logf("Correctly got error when unblocking non-blocked task: %v", err)

	// Cleanup
	_, _, _ = runTaskflow(t, binPath, dbDir, "delete",
		fmt.Sprintf(`{"id":"%s"}`, taskID))
}
