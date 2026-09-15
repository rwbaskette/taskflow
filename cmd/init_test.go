package cmd

// Unit tests for replacePointerWithDirAnchor (the `init --force` pointer
// replacement at the cwd, taskflow-init-anchor.md section 4c). The
// DB-creation seam lets the tests inject a failure that occurs AFTER the
// anchor directory is created, which is the mid-creation window where the
// pointer must still be restored byte-for-byte (repair-failure rule, design
// section 6).

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTestPointer writes a pointer file at dir/.taskflow and returns its
// exact bytes.
func writeTestPointer(t *testing.T, dir string) []byte {
	t.Helper()
	want := []byte("database: /somewhere/else/tasks.db\n")
	if err := os.WriteFile(filepath.Join(dir, ".taskflow"), want, 0o644); err != nil {
		t.Fatalf("setup: write pointer: %v", err)
	}
	return want
}

// TestReplacePointerWithDirAnchorSuccess verifies the success path: the
// pointer file is replaced by a directory anchor whose tasks.db exists.
func TestReplacePointerWithDirAnchorSuccess(t *testing.T) {
	dir := t.TempDir()
	writeTestPointer(t, dir)

	if err := replacePointerWithDirAnchor(dir, nil); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	fi, err := os.Stat(filepath.Join(dir, ".taskflow"))
	if err != nil || !fi.IsDir() {
		t.Fatalf("expected .taskflow to be a directory after the replacement (err=%v)", err)
	}
	dbPath := filepath.Join(dir, ".taskflow", "tasks.db")
	if _, err := os.Stat(dbPath); err != nil {
		t.Errorf("expected the DB created at %s: %v", dbPath, err)
	}
	// The created DB must open with the schema (ensureDB is a no-op open).
	if err := ensureDB(dbPath); err != nil {
		t.Errorf("the created DB must open cleanly: %v", err)
	}
}

// TestReplacePointerWithDirAnchorRestoreOnCreateFailure injects a DB failure
// in both windows and verifies the pointer is restored byte-for-byte:
//   - before the anchor directory is created (the already-tested restore
//     case), and
//   - after the anchor directory is created (the mid-creation window:
//     db.NewDB MkdirAll's dir/.taskflow before opening the DB, so a later
//     failure would previously leave the pointer lost under an empty
//     directory DB).
func TestReplacePointerWithDirAnchorRestoreOnCreateFailure(t *testing.T) {
	cases := map[string]func(path string) error{
		"failure before directory creation": func(string) error {
			return errors.New("injected db failure")
		},
		"failure after directory creation": func(path string) error {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			return errors.New("injected db failure")
		},
	}
	for name, create := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			want := writeTestPointer(t, dir)

			err := replacePointerWithDirAnchor(dir, create)
			if err == nil {
				t.Fatal("expected the injected DB failure to surface")
			}
			// The bare DB error must surface (no restore note on success).
			if err.Error() != "injected db failure" {
				t.Errorf("expected the bare DB error, got %q", err)
			}

			// Pointer bytes restored byte-for-byte, no leftover directory.
			got, rerr := os.ReadFile(filepath.Join(dir, ".taskflow"))
			if rerr != nil {
				t.Fatalf("the pointer file must be restored: %v", rerr)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("pointer bytes must be restored byte-for-byte, got %q want %q", got, want)
			}
			if fi, serr := os.Stat(filepath.Join(dir, ".taskflow")); serr == nil && fi.IsDir() {
				t.Error("no anchor directory may be left behind after the restore")
			}
		})
	}
}

// TestReplacePointerWithDirAnchorRestoreFailureSurfaced verifies that a
// failing restore is NOT discarded: the injected create func makes the
// holder dir unwritable after creating the anchor directory, so neither the
// directory removal nor the pointer rewrite can succeed, and the returned
// error must say the pointer needs a manual restore.
func TestReplacePointerWithDirAnchorRestoreFailureSurfaced(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("permission test requires a non-root user")
	}
	dir := t.TempDir()
	writeTestPointer(t, dir)

	create := func(path string) error {
		// db.NewDB's MkdirAll first (dir still writable), then lock the
		// holder dir so the restore cannot remove the directory nor
		// rewrite the pointer.
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.Chmod(dir, 0o555); err != nil {
			return err
		}
		return errors.New("injected db failure after the directory was created")
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	err := replacePointerWithDirAnchor(dir, create)
	if err == nil {
		t.Fatal("expected an error from the failing create and restore")
	}
	msg := err.Error()
	if !strings.Contains(msg, "pointer restore failed") {
		t.Errorf("the error must surface the restore failure, got %q", msg)
	}
	if !strings.Contains(msg, "restore the pointer file") || !strings.Contains(msg, "by hand") {
		t.Errorf("the error must instruct a manual restore, got %q", msg)
	}
	if !strings.Contains(msg, "injected db failure") {
		t.Errorf("the underlying DB error must stay part of the message, got %q", msg)
	}
}

// TestReplacePointerWithDirAnchorNeverRemovesPreexistingDir verifies the
// removal guard: a .taskflow directory that pre-existed is never removed.
// With a pre-existing directory there is no pointer file to read, so the
// replacement must fail before touching anything and the directory and its
// contents must survive intact.
func TestReplacePointerWithDirAnchorNeverRemovesPreexistingDir(t *testing.T) {
	dir := t.TempDir()
	anchorDir := filepath.Join(dir, ".taskflow")
	if err := os.MkdirAll(anchorDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	keep := filepath.Join(anchorDir, "keep.txt")
	if err := os.WriteFile(keep, []byte("keep"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	err := replacePointerWithDirAnchor(dir, func(string) error {
		t.Error("create must not run: reading the pointer fails before anything is removed")
		return errors.New("create must not be called")
	})
	if err == nil {
		t.Fatal("expected an error: there is no pointer file to replace")
	}

	if _, serr := os.Stat(keep); serr != nil {
		t.Errorf("the pre-existing .taskflow directory and its contents must survive: %v", serr)
	}
}
