package cmd

import (
	"io"
	"os"
	"strings"
	"testing"
)

// TestToolWrapperEmbedsVersion pins the wiring from the cmd package version
// variable into the generated wrapper: tool-wrapper must emit the embedded
// version constant and the directory-based spawn cwd. tool-wrapper prints the
// generated code to stdout with fmt.Print (not cmd.OutOrStdout), so the test
// captures the process stdout with a pipe.
func TestToolWrapperEmbedsVersion(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	old := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = old }()

	rootCmd.SetArgs([]string{"tool-wrapper"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	os.Stdout = old
	if err := w.Close(); err != nil {
		t.Fatalf("close pipe writer: %v", err)
	}

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}

	code := string(out)
	want := `const TASKFLOW_WRAPPER_VERSION = "` + version + `";`
	if !strings.Contains(code, want) {
		t.Errorf("tool-wrapper output missing %q; first 200 chars:\n%s", want, code[:min(len(code), 200)])
	}
	if !strings.Contains(code, "cwd: context.directory") {
		t.Error("tool-wrapper output missing: cwd: context.directory")
	}
	if strings.Contains(code, "context.worktree") {
		t.Error("tool-wrapper output must not reference context.worktree")
	}
}
