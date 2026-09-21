package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rwbaskette/taskflow/internal/version"
)

// TestToolWrapperEmbedsVersion pins the wiring from the cmd package version
// variable into the generated wrapper: tool-wrapper must emit the embedded
// version constant and the directory-based spawn cwd. The command writes its
// output through cmd.OutOrStdout, so the test captures it with a buffer.
func TestToolWrapperEmbedsVersion(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)

	rootCmd.SetArgs([]string{"tool-wrapper"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	code := buf.String()
	want := `const TASKFLOW_WRAPPER_VERSION = "` + version.Version + `";`
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
