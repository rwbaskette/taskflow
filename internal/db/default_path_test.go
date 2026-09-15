package db

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/rwbaskette/taskflow/internal/anchor"
)

// Env tests keep the t.Setenv restore discipline; assertions add err == nil
// since DefaultDBPath now returns two values. The no-env walk behavior is
// covered by internal/anchor tests (temp-dir anchors); the only no-env unit
// test here is TestDefaultDBPath_NoEnvFindsDirAnchor (design section 10).

func TestDefaultDBPath_WithEnvVar(t *testing.T) {
	base := t.TempDir()
	t.Setenv("TASKFLOW_DIR", filepath.Join(base, "custom", "taskflow"))

	got, err := DefaultDBPath()
	if err != nil {
		t.Fatalf("DefaultDBPath() error = %v, want nil", err)
	}
	want := filepath.Join(base, "custom", "taskflow", "tasks.db")
	if got != want {
		t.Errorf("DefaultDBPath() = %v, want %v", got, want)
	}
}

func TestDefaultDBPath_WithEmptyEnvVar(t *testing.T) {
	// Empty env must behave exactly like unset: the TASKFLOW_DIR branch is
	// skipped and the anchor walk runs (design section 3).
	t.Run("bare temp dir, no anchor: not-found", func(t *testing.T) {
		t.Setenv("TASKFLOW_DIR", "")
		dir := t.TempDir()
		t.Chdir(dir)

		got, err := DefaultDBPath()
		if got != "" {
			t.Errorf("DefaultDBPath() = %q, want an empty path on failure", got)
		}
		var aerr *anchor.AnchorError
		if !errors.As(err, &aerr) {
			t.Fatalf("DefaultDBPath() error = %v, want *AnchorError", err)
		}
		if aerr.Reason != anchor.ReasonNotFound {
			t.Errorf("Reason = %q, want %q (empty env must not short-circuit to an env path)", aerr.Reason, anchor.ReasonNotFound)
		}
	})

	t.Run("temp dir with a .taskflow dir anchor: resolved", func(t *testing.T) {
		t.Setenv("TASKFLOW_DIR", "")
		dir := t.TempDir()
		t.Chdir(dir)
		if err := os.MkdirAll(filepath.Join(dir, ".taskflow"), 0o755); err != nil {
			t.Fatalf("create anchor dir: %v", err)
		}
		// The walk branch stats the resolved DB (missing-db guard), so the
		// DB file must exist for resolution to succeed.
		if err := os.WriteFile(filepath.Join(dir, ".taskflow", "tasks.db"), []byte("placeholder"), 0o644); err != nil {
			t.Fatalf("create tasks.db: %v", err)
		}

		got, err := DefaultDBPath()
		if err != nil {
			t.Fatalf("DefaultDBPath() error = %v, want nil", err)
		}
		want := filepath.Join(dir, ".taskflow", "tasks.db")
		if got != want {
			t.Errorf("DefaultDBPath() = %v, want %v", got, want)
		}
	})
}

func TestDefaultDBPath_AbsolutePath(t *testing.T) {
	base := t.TempDir()
	t.Setenv("TASKFLOW_DIR", filepath.Join(base, "opt", "data"))

	got, err := DefaultDBPath()
	if err != nil {
		t.Fatalf("DefaultDBPath() error = %v, want nil", err)
	}
	want := filepath.Join(base, "opt", "data", "tasks.db")
	if got != want {
		t.Errorf("DefaultDBPath() = %v, want %v", got, want)
	}
}

func TestDefaultDBPath_RelativePath(t *testing.T) {
	t.Setenv("TASKFLOW_DIR", "mydata")

	got, err := DefaultDBPath()
	if err != nil {
		t.Fatalf("DefaultDBPath() error = %v, want nil", err)
	}
	// Should resolve to absolute path
	if got == filepath.Join("mydata", "tasks.db") {
		t.Errorf("DefaultDBPath() = %v, expected absolute path but got relative", got)
	}
	// Should end with tasks.db
	if got[len(got)-len("tasks.db"):] != "tasks.db" {
		t.Errorf("DefaultDBPath() = %v, expected path ending with tasks.db", got)
	}
}

func TestDefaultDBPath_TrailingSlash(t *testing.T) {
	base := t.TempDir()
	t.Setenv("TASKFLOW_DIR", filepath.Join(base, "work")+string(filepath.Separator))

	got, err := DefaultDBPath()
	if err != nil {
		t.Fatalf("DefaultDBPath() error = %v, want nil", err)
	}
	want := filepath.Join(base, "work", "tasks.db")
	if got != want {
		t.Errorf("DefaultDBPath() = %v, want %v", got, want)
	}
}

// TestDefaultDBPath_NoEnvFindsDirAnchor covers the walk branch with env
// unset: a .taskflow directory anchor in the (chdir'ed) temp cwd resolves to
// <dir>/.taskflow/tasks.db with a nil error. The broader no-env walk behavior
// lives in internal/anchor tests; this only pins the wiring.
func TestDefaultDBPath_NoEnvFindsDirAnchor(t *testing.T) {
	// t.Setenv cannot unset; unset manually with restore discipline.
	orig, had := os.LookupEnv("TASKFLOW_DIR")
	if had {
		t.Setenv("TASKFLOW_DIR", "")
		os.Unsetenv("TASKFLOW_DIR")
	}
	t.Cleanup(func() {
		if had {
			os.Setenv("TASKFLOW_DIR", orig)
		}
	})

	dir := t.TempDir()
	t.Chdir(dir)

	if err := os.MkdirAll(filepath.Join(dir, ".taskflow"), 0o755); err != nil {
		t.Fatalf("create anchor dir: %v", err)
	}
	// The walk branch stats the resolved DB path (missing-db guard), so the
	// DB file must exist for DefaultDBPath to return the path.
	if err := os.WriteFile(filepath.Join(dir, ".taskflow", "tasks.db"), []byte("placeholder"), 0o644); err != nil {
		t.Fatalf("create tasks.db: %v", err)
	}

	got, err := DefaultDBPath()
	if err != nil {
		t.Fatalf("DefaultDBPath() error = %v, want nil", err)
	}
	want := filepath.Join(dir, ".taskflow", "tasks.db")
	if got != want {
		t.Errorf("DefaultDBPath() = %v, want %v", got, want)
	}
}
