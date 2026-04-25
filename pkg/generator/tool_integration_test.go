//go:build integration

package generator_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rwbaskette/taskflow/pkg/generator"
)

// sharedDir is created once by TestMain. It contains node_modules after
// npm install runs, so any .mjs file written into a subdir can import
// @opencode-ai/plugin via node's upward module resolution.
var sharedDir string

func TestMain(m *testing.M) {
	for _, bin := range []string{"node", "npm"} {
		if _, err := exec.LookPath(bin); err != nil {
			fmt.Fprintf(os.Stderr, "integration tests require %s in PATH\n", bin)
			os.Exit(1)
		}
	}

	dir, err := os.MkdirTemp("", "taskflow-integration-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	pkgJSON := `{"dependencies":{"@opencode-ai/plugin":"1.4.10"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "write package.json: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("npm", "install", "--silent")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "npm install failed: %v\n%s\n", err, out)
		os.Exit(1)
	}

	sharedDir = dir
	os.Exit(m.Run())
}

// testDir creates a subdirectory inside sharedDir for a single test. Because
// it lives under sharedDir, node's ESM resolver walks up and finds
// sharedDir/node_modules automatically. Cleanup is registered via t.Cleanup.
func testDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp(sharedDir, "test-*")
	if err != nil {
		t.Fatalf("create test dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

// runGeneratedToolWrapper writes the generated TypeScript to dir/tools.mjs,
// appends an introspection snippet, runs it with node, and returns the parsed
// JSON result keyed by tool name.
func runGeneratedToolWrapper(t *testing.T, dir string, opts *generator.ToolWrapperOptions) map[string]toolShape {
	t.Helper()

	ts, err := generator.GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper: %v", err)
	}

	// Append an introspection block that serialises every exported tool into:
	// { toolName: { description, args: { argName: zodType } } }
	// Zod v4 exposes the type as schema.def.type (a plain string).
	introspect := `
// -- introspection --
const tools = {
  task_add, task_block, task_complete, task_delete,
  task_list_all, task_list_blocked, task_list_done,
  task_list_status_in_progress, task_list_status_todo,
  task_reset_timedout, task_start,
};
function zodType(schema) { return schema?.def?.type ?? ''; }
const result = {};
for (const [name, t] of Object.entries(tools)) {
  result[name] = {
    description: t.description,
    args: Object.fromEntries(Object.entries(t.args).map(([k, v]) => [k, zodType(v)])),
  };
}
console.log(JSON.stringify(result));
`

	mjsPath := filepath.Join(dir, "tools.mjs")
	if err := os.WriteFile(mjsPath, []byte(ts+introspect), 0600); err != nil {
		t.Fatalf("write tools.mjs: %v", err)
	}

	out, err := exec.Command("node", mjsPath).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Fatalf("node exited non-zero:\n%s", exitErr.Stderr)
		}
		t.Fatalf("node: %v", err)
	}

	var shapes map[string]toolShape
	if err := json.Unmarshal(out, &shapes); err != nil {
		t.Fatalf("parse node output: %v\nraw: %s", err, out)
	}
	return shapes
}

type toolShape struct {
	Description string            `json:"description"`
	Args        map[string]string `json:"args"`
}

// ── tests ────────────────────────────────────────────────────────────────────

func TestToolWrapperLoadsInNode(t *testing.T) {
	shapes := runGeneratedToolWrapper(t, testDir(t), generator.DefaultToolWrapperOptions())
	if len(shapes) == 0 {
		t.Fatal("expected at least one tool shape, got none")
	}
}

func TestToolWrapperAllToolsPresent(t *testing.T) {
	shapes := runGeneratedToolWrapper(t, testDir(t), generator.DefaultToolWrapperOptions())

	want := []string{
		"task_add", "task_block", "task_complete", "task_delete",
		"task_list_all", "task_list_blocked", "task_list_done",
		"task_list_status_in_progress", "task_list_status_todo",
		"task_reset_timedout", "task_start",
	}
	for _, name := range want {
		if _, ok := shapes[name]; !ok {
			t.Errorf("missing tool %q", name)
		}
	}
}

func TestToolWrapperDescriptions(t *testing.T) {
	shapes := runGeneratedToolWrapper(t, testDir(t), generator.DefaultToolWrapperOptions())

	cases := []struct {
		tool string
		want string
	}{
		{"task_add", "Add a new task to the task list"},
		{"task_block", "Block a task by providing its ID and a reason"},
		{"task_complete", "Mark a task as completed"},
		{"task_delete", "Soft delete a task"},
		{"task_list_all", "List all tasks"},
		{"task_list_blocked", "List blocked tasks"},
		{"task_list_done", "List completed tasks"},
		{"task_list_status_in_progress", "List in-progress tasks"},
		{"task_list_status_todo", "List todo tasks"},
		{"task_reset_timedout", "Reset timed out tasks"},
		{"task_start", "Start working on a task"},
	}
	for _, tc := range cases {
		t.Run(tc.tool, func(t *testing.T) {
			s, ok := shapes[tc.tool]
			if !ok {
				t.Fatalf("tool %q not found", tc.tool)
			}
			if !strings.Contains(s.Description, tc.want) {
				t.Errorf("description %q does not contain %q", s.Description, tc.want)
			}
		})
	}
}

func TestToolWrapperArgTypes(t *testing.T) {
	shapes := runGeneratedToolWrapper(t, testDir(t), generator.DefaultToolWrapperOptions())

	cases := []struct {
		tool    string
		arg     string
		wantTyp string
	}{
		{"task_add", "id", "string"},
		{"task_add", "title", "string"},
		{"task_add", "description", "string"},
		{"task_add", "milestone", "string"},
		{"task_add", "actor", "string"},
		{"task_block", "id", "string"},
		{"task_block", "reason", "string"},
		{"task_complete", "id", "string"},
		{"task_delete", "id", "string"},
		{"task_start", "id", "string"},
		{"task_list_all", "limit", "number"},
		{"task_list_all", "offset", "number"},
		{"task_reset_timedout", "minutes", "number"},
		{"task_list_all", "all", "boolean"},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s.%s", tc.tool, tc.arg), func(t *testing.T) {
			s, ok := shapes[tc.tool]
			if !ok {
				t.Fatalf("tool %q not found", tc.tool)
			}
			got, ok := s.Args[tc.arg]
			if !ok {
				t.Fatalf("arg %q not found in tool %q (have: %v)", tc.arg, tc.tool, s.Args)
			}
			if got != tc.wantTyp {
				t.Errorf("arg %q type = %q, want %q", tc.arg, got, tc.wantTyp)
			}
		})
	}
}

func TestToolWrapperArgCounts(t *testing.T) {
	shapes := runGeneratedToolWrapper(t, testDir(t), generator.DefaultToolWrapperOptions())

	cases := []struct {
		tool      string
		wantCount int
	}{
		{"task_add", 5},                  // id, title, description, milestone, actor
		{"task_block", 2},                // id, reason
		{"task_complete", 1},             // id
		{"task_delete", 1},               // id
		{"task_list_all", 5},             // milestone, actor, limit, offset, all
		{"task_list_blocked", 4},         // milestone, actor, limit, offset
		{"task_list_done", 4},            // milestone, actor, limit, offset
		{"task_list_status_in_progress", 4},
		{"task_list_status_todo", 4},
		{"task_reset_timedout", 1},       // minutes
		{"task_start", 1},                // id
	}
	for _, tc := range cases {
		t.Run(tc.tool, func(t *testing.T) {
			s, ok := shapes[tc.tool]
			if !ok {
				t.Fatalf("tool %q not found", tc.tool)
			}
			if got := len(s.Args); got != tc.wantCount {
				t.Errorf("arg count = %d, want %d (args: %v)", got, tc.wantCount, s.Args)
			}
		})
	}
}

func TestToolWrapperCustomBinaryPath(t *testing.T) {
	dir := testDir(t)
	shapes := runGeneratedToolWrapper(t, dir, &generator.ToolWrapperOptions{BinaryPath: "my-custom-bin"})
	if _, ok := shapes["task_add"]; !ok {
		t.Error("task_add missing from output with custom binary path")
	}

	// Also verify the binary path appears literally in the generated source.
	ts, err := generator.GenerateToolWrapper(&generator.ToolWrapperOptions{BinaryPath: "my-custom-bin"})
	if err != nil {
		t.Fatalf("GenerateToolWrapper: %v", err)
	}
	if !strings.Contains(ts, "my-custom-bin") {
		t.Error("generated output does not contain custom binary path")
	}
}
