package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rwbaskette/taskflow/internal/version"
)

// TestToolWrapperEmbedsVersion pins the wiring from the cmd package version
// variable into the generated plugin: tool-wrapper must emit the OpenCode 2.0
// plugin (a default-exported Plugin.define registering the taskflow_* tools)
// with the embedded version constant, and it must resolve the spawn cwd from
// the session's working directory. The command writes its output through
// cmd.OutOrStdout, so the test captures it with a buffer.
func TestToolWrapperEmbedsVersion(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)

	rootCmd.SetArgs([]string{"tool-wrapper"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	code := buf.String()
	want := `const TASKFLOW_PLUGIN_VERSION = "` + version.Version + `";`
	if !strings.Contains(code, want) {
		t.Errorf("tool-wrapper output missing %q; first 200 chars:\n%s", want, code[:min(len(code), 200)])
	}
	if !strings.Contains(code, "import { Plugin } from \"@opencode/plugin\";") {
		t.Error("tool-wrapper output missing: import { Plugin } from \"@opencode/plugin\";")
	}
	if !strings.Contains(code, `editor.namespace({ name: NAMESPACE`) {
		t.Error(`tool-wrapper output missing: editor.namespace({ name: NAMESPACE`)
	}
	if strings.Contains(code, "@opencode-ai/plugin") {
		t.Error("tool-wrapper output must not reference the V1 @opencode-ai/plugin import")
	}
	if strings.Contains(code, "context.directory") {
		t.Error("tool-wrapper output must not reference context.directory")
	}
	if !strings.Contains(code, "session.location.directory") {
		t.Error("tool-wrapper output missing: session.location.directory")
	}
}
