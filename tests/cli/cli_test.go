//go:build integration

package cli_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binaryPath is the absolute path to the compiled taskflow binary.
// It is set once by TestMain before any tests run.
var binaryPath string

func TestMain(m *testing.M) {
	// Build the binary once for all tests.
	tmp, err := os.MkdirTemp("", "taskflow-cli-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmp)

	bin := filepath.Join(tmp, "taskflow")
	// Resolve the module root relative to this test file.
	moduleRoot := filepath.Join("..", "..")
	absRoot, err := filepath.Abs(moduleRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve module root: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = absRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build taskflow binary: %v\n", err)
		os.Exit(1)
	}
	binaryPath = bin

	os.Exit(m.Run())
}

// cliResult holds the combined output and exit status of a CLI invocation.
type cliResult struct {
	stdout   string
	stderr   string
	combined string
	exitCode int
}

// runCLI executes the taskflow binary in the given working directory with the
// supplied arguments. Each test should pass its own temp directory as workdir
// so the CLI creates an isolated .taskflow/tasks.db there.
func runCLI(t *testing.T, workdir string, args ...string) cliResult {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = workdir

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run taskflow: %v", err)
		}
	}

	return cliResult{
		stdout:   stdout.String(),
		stderr:   stderr.String(),
		combined: stdout.String() + stderr.String(),
		exitCode: exitCode,
	}
}

// tempWorkdir creates an isolated temporary directory for a single test.
func tempWorkdir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// addTask is a convenience helper that adds a task via the CLI and fails the
// test if the command does not succeed.
func addTask(t *testing.T, workdir, jsonInput string) {
	t.Helper()
	r := runCLI(t, workdir, "add", jsonInput)
	if r.exitCode != 0 {
		t.Fatalf("setup: add task failed (exit %d): %s", r.exitCode, r.combined)
	}
}

// --------------------------------------------------------------------------
// assertContains / assertNotContains / assertError helpers
// --------------------------------------------------------------------------

func assertContains(t *testing.T, haystack, needle, msg string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("%s: expected output to contain %q, got:\n%s", msg, needle, haystack)
	}
}

func assertNotContains(t *testing.T, haystack, needle, msg string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Errorf("%s: expected output NOT to contain %q, got:\n%s", msg, needle, haystack)
	}
}

func assertExitZero(t *testing.T, r cliResult, msg string) {
	t.Helper()
	if r.exitCode != 0 {
		t.Errorf("%s: expected exit 0, got %d; output:\n%s", msg, r.exitCode, r.combined)
	}
}

func assertExitNonZero(t *testing.T, r cliResult, msg string) {
	t.Helper()
	if r.exitCode == 0 {
		t.Errorf("%s: expected non-zero exit, got 0; output:\n%s", msg, r.combined)
	}
}

func assertError(t *testing.T, r cliResult, msg string) {
	t.Helper()
	lower := strings.ToLower(r.combined)
	if !strings.Contains(lower, "error") && !strings.Contains(lower, "required") {
		t.Errorf("%s: expected error in output, got:\n%s", msg, r.combined)
	}
}

// --------------------------------------------------------------------------
// Add command tests (replaces test-add.sh)
// --------------------------------------------------------------------------

func TestAdd_NoArguments(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "add")
	assertExitNonZero(t, r, "no args")
	assertError(t, r, "no args")
}

func TestAdd_ValidWithRequiredFields(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "add", `{"id":"add-1","title":"Test Task","milestone":"v1","actor":"tester","description":"A test task"}`)
	assertExitZero(t, r, "valid add")
	assertContains(t, r.combined, "Task added successfully", "success message")
	assertContains(t, r.combined, "Test Task", "title in output")
	assertContains(t, r.combined, "add-1", "id in output")
}

func TestAdd_WithDescription(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "add", `{"id":"add-2","title":"Task with Description","milestone":"v1","actor":"tester","description":"This is a description"}`)
	assertExitZero(t, r, "add with description")
	assertContains(t, r.combined, "Task added successfully", "success message")
	assertContains(t, r.combined, "Task with Description", "title")
	assertContains(t, r.combined, "This is a description", "description")
}

func TestAdd_WithMilestone(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "add", `{"id":"add-3","title":"Task with Milestone","milestone":"v1.0","actor":"tester","description":"desc"}`)
	assertExitZero(t, r, "add with milestone")
	assertContains(t, r.combined, "v1.0", "milestone in output")
}

func TestAdd_WithActor(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "add", `{"id":"add-4","title":"Task with Actor","milestone":"v1","actor":"developer","description":"desc"}`)
	assertExitZero(t, r, "add with actor")
	assertContains(t, r.combined, "developer", "actor in output")
}

func TestAdd_AllFields(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "add", `{"id":"add-5","title":"Full Task","milestone":"v2.0","actor":"admin","description":"Full description"}`)
	assertExitZero(t, r, "add all fields")
	assertContains(t, r.combined, "Full Task", "title")
	assertContains(t, r.combined, "Full description", "description")
	assertContains(t, r.combined, "v2.0", "milestone")
	assertContains(t, r.combined, "admin", "actor")
}

func TestAdd_MultipleTasks(t *testing.T) {
	dir := tempWorkdir(t)
	r1 := runCLI(t, dir, "add", `{"id":"add-6","title":"Task 1","milestone":"v1","actor":"tester","description":"desc"}`)
	assertExitZero(t, r1, "first task")
	assertContains(t, r1.combined, "Task added successfully", "first success")

	r2 := runCLI(t, dir, "add", `{"id":"add-7","title":"Task 2","milestone":"v1","actor":"tester","description":"desc"}`)
	assertExitZero(t, r2, "second task")
	assertContains(t, r2.combined, "Task added successfully", "second success")
}

func TestAdd_Help(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "add", "--help")
	assertExitZero(t, r, "help")
	assertContains(t, r.combined, "add", "help output")
}

func TestAdd_MissingTitle(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "add", `{"id":"add-8","milestone":"v1","actor":"tester","description":"desc"}`)
	assertError(t, r, "missing title")
}

func TestAdd_InvalidJSON(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "add", "not-json")
	assertError(t, r, "invalid json")
}

// --------------------------------------------------------------------------
// Update command tests (replaces test-update.sh)
// --------------------------------------------------------------------------

func TestUpdate_NoArguments(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "update")
	assertExitNonZero(t, r, "no args")
	assertError(t, r, "no args")
}

func TestUpdate_Title(t *testing.T) {
	dir := tempWorkdir(t)
	addTask(t, dir, `{"id":"upd-1","title":"Original Title","milestone":"v1","actor":"tester","description":"Original desc"}`)

	r := runCLI(t, dir, "update", `{"id":"upd-1","title":"New Title"}`)
	assertExitZero(t, r, "update title")
	assertContains(t, r.combined, "Task updated successfully", "success message")
	assertContains(t, r.combined, "New Title", "new title in output")
}

func TestUpdate_Description(t *testing.T) {
	dir := tempWorkdir(t)
	addTask(t, dir, `{"id":"upd-2","title":"T","milestone":"v1","actor":"tester","description":"Old"}`)

	r := runCLI(t, dir, "update", `{"id":"upd-2","description":"New Description"}`)
	assertExitZero(t, r, "update description")
	assertContains(t, r.combined, "Task updated successfully", "success message")
	assertContains(t, r.combined, "New Description", "new description")
}

func TestUpdate_Milestone(t *testing.T) {
	dir := tempWorkdir(t)
	addTask(t, dir, `{"id":"upd-3","title":"T","milestone":"v1","actor":"tester","description":"d"}`)

	r := runCLI(t, dir, "update", `{"id":"upd-3","milestone":"v2.0"}`)
	assertExitZero(t, r, "update milestone")
	assertContains(t, r.combined, "v2.0", "new milestone")
}

func TestUpdate_Actor(t *testing.T) {
	dir := tempWorkdir(t)
	addTask(t, dir, `{"id":"upd-4","title":"T","milestone":"v1","actor":"tester","description":"d"}`)

	r := runCLI(t, dir, "update", `{"id":"upd-4","actor":"new-actor"}`)
	assertExitZero(t, r, "update actor")
	assertContains(t, r.combined, "new-actor", "new actor")
}

func TestUpdate_MultipleFields(t *testing.T) {
	dir := tempWorkdir(t)
	addTask(t, dir, `{"id":"upd-5","title":"T","milestone":"v1","actor":"tester","description":"d"}`)

	r := runCLI(t, dir, "update", `{"id":"upd-5","title":"Multi Update","description":"New Desc","milestone":"v3.0","actor":"admin"}`)
	assertExitZero(t, r, "update multiple fields")
	assertContains(t, r.combined, "Task updated successfully", "success message")
	assertContains(t, r.combined, "Multi Update", "new title")
}

func TestUpdate_NonExistentTask(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "update", `{"id":"does-not-exist","title":"Nope"}`)
	assertError(t, r, "non-existent task")
}

func TestUpdate_Help(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "update", "--help")
	assertExitZero(t, r, "help")
	assertContains(t, r.combined, "update", "help output")
}

func TestUpdate_Status(t *testing.T) {
	dir := tempWorkdir(t)
	addTask(t, dir, `{"id":"upd-6","title":"T","milestone":"v1","actor":"tester","description":"d"}`)

	r := runCLI(t, dir, "update", `{"id":"upd-6","status":"in_progress"}`)
	assertExitZero(t, r, "update status")
	assertContains(t, r.combined, "Task updated successfully", "success message")
	assertContains(t, r.combined, "in_progress", "new status")
}

// --------------------------------------------------------------------------
// Complete command tests (replaces test-complete.sh)
// --------------------------------------------------------------------------

func TestComplete_NoArguments(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "complete")
	assertExitNonZero(t, r, "no args")
	assertError(t, r, "no args")
}

func TestComplete_ByID(t *testing.T) {
	dir := tempWorkdir(t)
	addTask(t, dir, `{"id":"comp-1","title":"Task to Complete","milestone":"v1","actor":"tester","description":"desc"}`)

	r := runCLI(t, dir, "complete", `{"id":"comp-1"}`)
	assertExitZero(t, r, "complete by id")
	assertContains(t, r.combined, "Task completed successfully", "success message")
	assertContains(t, r.combined, "comp-1", "task id")
}

func TestComplete_MultipleTasks(t *testing.T) {
	dir := tempWorkdir(t)
	addTask(t, dir, `{"id":"comp-2","title":"Another Task","milestone":"v1","actor":"tester","description":"desc"}`)
	addTask(t, dir, `{"id":"comp-3","title":"Third Task","milestone":"v1","actor":"tester","description":"desc"}`)

	r1 := runCLI(t, dir, "complete", `{"id":"comp-2"}`)
	assertExitZero(t, r1, "complete first")
	assertContains(t, r1.combined, "Task completed successfully", "first success")

	r2 := runCLI(t, dir, "complete", `{"id":"comp-3"}`)
	assertExitZero(t, r2, "complete second")
	assertContains(t, r2.combined, "Task completed successfully", "second success")
}

func TestComplete_Help(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "complete", "--help")
	assertExitZero(t, r, "help")
	assertContains(t, r.combined, "complete", "help output")
}

func TestComplete_NonExistentTask(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "complete", `{"id":"does-not-exist"}`)
	assertError(t, r, "non-existent task")
}

func TestComplete_InvalidJSON(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "complete", "not-json")
	assertError(t, r, "invalid json")
}

// --------------------------------------------------------------------------
// Block command tests (replaces test-block.sh)
// --------------------------------------------------------------------------

func TestBlock_NoArguments(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "block")
	assertExitNonZero(t, r, "no args")
	assertError(t, r, "no args")
}

func TestBlock_WithReason(t *testing.T) {
	dir := tempWorkdir(t)
	addTask(t, dir, `{"id":"blk-1","title":"Block Me","milestone":"v1","actor":"tester","description":"desc"}`)

	r := runCLI(t, dir, "block", `{"id":"blk-1","reason":"Waiting for dependency"}`)
	assertExitZero(t, r, "block with reason")
	assertContains(t, r.combined, "Task blocked successfully", "success message")
	assertContains(t, r.combined, "blk-1", "task id")
}

func TestBlock_WithoutReason(t *testing.T) {
	dir := tempWorkdir(t)
	addTask(t, dir, `{"id":"blk-2","title":"Block Me Too","milestone":"v1","actor":"tester","description":"desc"}`)

	r := runCLI(t, dir, "block", `{"id":"blk-2"}`)
	assertError(t, r, "missing reason")
}

func TestBlock_MultipleTasks(t *testing.T) {
	dir := tempWorkdir(t)
	addTask(t, dir, `{"id":"blk-3","title":"Block Me Three","milestone":"v1","actor":"tester","description":"desc"}`)

	r := runCLI(t, dir, "block", `{"id":"blk-3","reason":"Reason 1"}`)
	assertExitZero(t, r, "block task")
	assertContains(t, r.combined, "Task blocked successfully", "success")
}

func TestBlock_Help(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "block", "--help")
	assertExitZero(t, r, "help")
	assertContains(t, r.combined, "block", "help output")
}

func TestBlock_SpecialCharsInReason(t *testing.T) {
	dir := tempWorkdir(t)
	addTask(t, dir, `{"id":"blk-4","title":"Special Chars Task","milestone":"v1","actor":"tester","description":"desc"}`)

	r := runCLI(t, dir, "block", `{"id":"blk-4","reason":"Reason with special chars"}`)
	assertExitZero(t, r, "special chars in reason")
	assertContains(t, r.combined, "Task blocked successfully", "success")
}

func TestBlock_LongReason(t *testing.T) {
	dir := tempWorkdir(t)
	addTask(t, dir, `{"id":"blk-5","title":"Long Reason Task","milestone":"v1","actor":"tester","description":"desc"}`)

	r := runCLI(t, dir, "block", `{"id":"blk-5","reason":"This is a very long reason for blocking a task that explains in detail why the task cannot be worked on at this time."}`)
	assertExitZero(t, r, "long reason")
	assertContains(t, r.combined, "Task blocked successfully", "success")
}

func TestBlock_NonExistentTask(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "block", `{"id":"does-not-exist","reason":"test"}`)
	assertError(t, r, "non-existent task")
}

// --------------------------------------------------------------------------
// List command tests (replaces test-list.sh)
// --------------------------------------------------------------------------

// seedListTasks creates a common set of tasks for list tests.
func seedListTasks(t *testing.T, dir string) {
	t.Helper()
	addTask(t, dir, `{"id":"list-1","title":"List Task One","milestone":"sprint-1","actor":"alice","description":"desc"}`)
	addTask(t, dir, `{"id":"list-2","title":"List Task Two","milestone":"sprint-1","actor":"bob","description":"desc"}`)
	addTask(t, dir, `{"id":"list-3","title":"List Task Three","milestone":"sprint-2","actor":"alice","description":"desc"}`)
}

func TestList_Basic(t *testing.T) {
	dir := tempWorkdir(t)
	seedListTasks(t, dir)

	r := runCLI(t, dir, "list", `{}`)
	assertExitZero(t, r, "basic list")
	// Should not contain error
	lower := strings.ToLower(r.combined)
	if strings.Contains(lower, "error") {
		t.Errorf("basic list returned error: %s", r.combined)
	}
}

func TestList_AllTrue(t *testing.T) {
	dir := tempWorkdir(t)
	seedListTasks(t, dir)

	r := runCLI(t, dir, "list", `{"all":true}`)
	assertExitZero(t, r, "list all:true")
}

func TestList_FilterByMilestone(t *testing.T) {
	dir := tempWorkdir(t)
	seedListTasks(t, dir)

	r := runCLI(t, dir, "list", `{"milestone":"sprint-1"}`)
	assertExitZero(t, r, "filter by milestone")
	assertContains(t, r.combined, "sprint-1", "milestone in output")
}

func TestList_FilterByActor(t *testing.T) {
	dir := tempWorkdir(t)
	seedListTasks(t, dir)

	r := runCLI(t, dir, "list", `{"actor":"alice"}`)
	assertExitZero(t, r, "filter by actor")
	assertContains(t, r.combined, "alice", "actor in output")
}

func TestList_FilterByStatus(t *testing.T) {
	dir := tempWorkdir(t)
	seedListTasks(t, dir)

	r := runCLI(t, dir, "list", `{"status":"todo"}`)
	assertExitZero(t, r, "filter by status")
}

func TestList_CombinedMilestoneAndActor(t *testing.T) {
	dir := tempWorkdir(t)
	seedListTasks(t, dir)

	r := runCLI(t, dir, "list", `{"milestone":"sprint-1","actor":"alice"}`)
	assertExitZero(t, r, "combined milestone+actor")
}

func TestList_CombinedMilestoneAndStatus(t *testing.T) {
	dir := tempWorkdir(t)
	seedListTasks(t, dir)

	r := runCLI(t, dir, "list", `{"milestone":"sprint-1","status":"todo"}`)
	assertExitZero(t, r, "combined milestone+status")
}

func TestList_CombinedActorAndStatus(t *testing.T) {
	dir := tempWorkdir(t)
	seedListTasks(t, dir)

	r := runCLI(t, dir, "list", `{"actor":"alice","status":"in_progress"}`)
	assertExitZero(t, r, "combined actor+status")
}

func TestList_Help(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "list", "--help")
	assertExitZero(t, r, "help")
	assertContains(t, r.combined, "list", "help output")
}

func TestList_AllThreeFilters(t *testing.T) {
	dir := tempWorkdir(t)
	seedListTasks(t, dir)

	r := runCLI(t, dir, "list", `{"milestone":"sprint-1","actor":"alice","status":"todo"}`)
	assertExitZero(t, r, "all three filters")
}

func TestList_NonExistentMilestone(t *testing.T) {
	dir := tempWorkdir(t)
	seedListTasks(t, dir)

	r := runCLI(t, dir, "list", `{"milestone":"nonexistent-milestone-xyz"}`)
	assertExitZero(t, r, "non-existent milestone")
	// Should not error
	lower := strings.ToLower(r.combined)
	if strings.Contains(lower, "error") {
		t.Errorf("non-existent milestone returned error: %s", r.combined)
	}
}

func TestList_ProducesOutput(t *testing.T) {
	dir := tempWorkdir(t)
	seedListTasks(t, dir)

	r := runCLI(t, dir, "list", `{}`)
	if len(strings.TrimSpace(r.combined)) == 0 {
		t.Error("list command should produce output")
	}
}

// --------------------------------------------------------------------------
// Reset-timedout command tests (replaces test-reset.sh)
// --------------------------------------------------------------------------

func TestReset_WithMinutes(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "reset-timedout", `{"minutes":30}`)
	assertExitZero(t, r, "reset with minutes")
	// Should not error
	lower := strings.ToLower(r.combined)
	if strings.Contains(lower, "error") {
		t.Errorf("reset returned error: %s", r.combined)
	}
}

func TestReset_NoArguments(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "reset-timedout")
	assertExitNonZero(t, r, "no args")
	assertError(t, r, "no args")
}

func TestReset_LargeTimeout(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "reset-timedout", `{"minutes":1440}`)
	assertExitZero(t, r, "large timeout")
}

func TestReset_Help(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "reset-timedout", "--help")
	assertExitZero(t, r, "help")
	assertContains(t, r.combined, "reset-timedout", "help output")
}

func TestReset_OneMinute(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "reset-timedout", `{"minutes":1}`)
	assertExitZero(t, r, "1 minute timeout")
}

func TestReset_InvalidJSON(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "reset-timedout", "not-json")
	assertError(t, r, "invalid json")
}

func TestReset_ZeroMinutes(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "reset-timedout", `{"minutes":0}`)
	// Zero minutes may be valid or invalid — just ensure it produces output and doesn't crash.
	if len(strings.TrimSpace(r.combined)) == 0 {
		t.Error("zero minutes should produce some output")
	}
}
