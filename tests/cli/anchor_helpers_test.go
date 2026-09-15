//go:build integration

package cli_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// writePointerFile writes a .taskflow pointer file at dir containing exactly
// one line "database: <target>" and returns the pointer file path.
func writePointerFile(t *testing.T, dir, target string) string {
	t.Helper()
	pointerPath := filepath.Join(dir, ".taskflow")
	line := "database: " + target + "\n"
	if err := os.WriteFile(pointerPath, []byte(line), 0o644); err != nil {
		t.Fatalf("setup: failed to write pointer file %s: %v", pointerPath, err)
	}
	return pointerPath
}

// readFileBytes reads path, failing the test on error.
func readFileBytes(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("setup: failed to read %s: %v", path, err)
	}
	return data
}

// assertBytesEqual fails the test when the two byte slices differ.
func assertBytesEqual(t *testing.T, label string, want, got []byte) {
	t.Helper()
	if !bytes.Equal(want, got) {
		t.Errorf("%s: file content changed (%d bytes before, %d bytes after)", label, len(want), len(got))
	}
}

// listAll runs `taskflow list` with all:true in dir so completed, blocked, and
// in-progress tasks are visible too.
func listAll(t *testing.T, dir string) cliResult {
	t.Helper()
	return runCLI(t, dir, "list", `{"all":true}`)
}

// runCLIWithMinimalEnv executes the taskflow binary with a minimal
// environment: only HOME, TMPDIR, and an empty PATH; no variable is inherited.
// binaryPath is absolute, so the binary itself needs no PATH. This exercises
// the git-free property: taskflow must never shell out.
func runCLIWithMinimalEnv(t *testing.T, workdir string, args ...string) cliResult {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = workdir
	cmd.Env = []string{"HOME=" + workdir, "TMPDIR=" + os.TempDir(), "PATH="}

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

// runCLIConcurrent executes the taskflow binary without touching the testing
// state, so it is safe to call from goroutines. A start failure is reported as
// exit code -1; the caller fails the test from its own goroutine.
func runCLIConcurrent(workdir string, args ...string) cliResult {
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
			exitCode = -1
		}
	}
	return cliResult{
		stdout:   stdout.String(),
		stderr:   stderr.String(),
		combined: stdout.String() + stderr.String(),
		exitCode: exitCode,
	}
}

// anchorExistsAt reports whether dir contains a .taskflow entry (file or
// directory).
func anchorExistsAt(t *testing.T, dir string) bool {
	t.Helper()
	_, err := os.Stat(filepath.Join(dir, ".taskflow"))
	return err == nil
}

// chainFromDir returns the directories the anchor walk visits from dir up to
// the filesystem root, start dir first, mirroring internal/anchor.Resolve
// (EvalSymlinks on the start dir, then repeated filepath.Dir until the root).
func chainFromDir(dir string) []string {
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		if abs, err := filepath.Abs(resolved); err == nil {
			dir = abs
		}
	}
	chain := []string{}
	d := dir
	for {
		chain = append(chain, d)
		next := filepath.Dir(d)
		if next == d {
			break
		}
		d = next
	}
	return chain
}

// addTaskJSON builds a taskflow add JSON document with the given id.
func addTaskJSON(id string) string {
	return `{"id":"` + id + `","title":"Task ` + id + `","milestone":"v1","actor":"tester","description":"desc"}`
}
