//go:build integration

package generator_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rwbaskette/taskflow/internal/generator"
)

// sharedDir is created once by TestMain. It contains node_modules after
// npm install runs, so any .mjs file written into a subdir can import
// @opencode/plugin via node's upward module resolution.
var sharedDir string

// testBinary is the taskflow binary built once by TestMain from the repo
// sources; execution tests point the generated plugin at it through plugin
// options.
var testBinary string

// fakeBinary is a tiny stand-in binary built once by TestMain; the
// version-gate tests drive it through the FAKE_* environment variables.
var fakeBinary string

// fakeBinarySource is the single-file, stdlib-only source of the fake
// taskflow binary. Behavior is driven entirely by the environment it
// inherits:
//
//   - FAKE_VERSION: output for the "version" subcommand (default
//     "taskflow version 1.0.0")
//   - FAKE_SLEEP_MS: sleep this many milliseconds before answering any
//     non-version subcommand (cancellation tests)
//   - FAKE_LOG: append one line per invocation (the subcommand) to this file
//
// Non-version subcommands print the same JSON shape the real CLI's list
// command prints (tasks/total/limit/offset/hasMore).
const fakeBinarySource = `package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	sub := ""
	if len(os.Args) > 1 {
		sub = os.Args[1]
	}
	if logPath := os.Getenv("FAKE_LOG"); logPath != "" {
		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err == nil {
			fmt.Fprintln(f, sub)
			f.Close()
		}
	}
	if sub == "version" {
		version := os.Getenv("FAKE_VERSION")
		if version == "" {
			version = "taskflow version 1.0.0"
		}
		fmt.Println(version)
		return
	}
	if ms := os.Getenv("FAKE_SLEEP_MS"); ms != "" {
		if n, err := strconv.Atoi(ms); err == nil && n > 0 {
			time.Sleep(time.Duration(n) * time.Millisecond)
		}
	}
	fmt.Println("{\"tasks\":[],\"total\":0,\"limit\":10,\"offset\":0,\"hasMore\":false}")
}
`

func TestMain(m *testing.M) {
	for _, bin := range []string{"node", "npm", "go"} {
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

	pkgJSON := `{"dependencies":{"@opencode/plugin":"2.0.26"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "write package.json: %v\n", err)
		os.Exit(1)
	}

	// --legacy-peer-deps: the package's solid-js/@opentui/* peers are not
	// needed for tool registration and some are heavy or native.
	install := exec.Command("npm", "install", "--silent", "--legacy-peer-deps")
	install.Dir = dir
	if out, err := install.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "npm install failed: %v\n%s\n", err, out)
		os.Exit(1)
	}

	// Build the real taskflow binary once from the repo root (the directory
	// of the go.mod file) so execution tests can drive the generated plugin
	// against genuine CLI runs.
	modOut, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "go env GOMOD: %v\n%s\n", err, modOut)
		os.Exit(1)
	}
	repoRoot := filepath.Dir(strings.TrimSpace(string(modOut)))
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	testBinary = filepath.Join(dir, "taskflow-test"+ext)
	build := exec.Command("go", "build", "-C", repoRoot, "-o", testBinary, ".")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "go build failed: %v\n%s\n", err, out)
		os.Exit(1)
	}

	// Build the fake taskflow binary for the version-gate tests. The
	// single-file, stdlib-only source builds outside the module.
	fakeSource := filepath.Join(dir, "fake-tf-main.go")
	if err := os.WriteFile(fakeSource, []byte(fakeBinarySource), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "write fake-tf-main.go: %v\n", err)
		os.Exit(1)
	}
	fakeBinary = filepath.Join(dir, "fake-tf"+ext)
	buildFake := exec.Command("go", "build", "-o", fakeBinary, fakeSource)
	buildFake.Dir = dir
	if out, err := buildFake.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "fake binary build failed: %v\n%s\n", err, out)
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

// writePlugin generates the plugin source into dir/plugin.mjs and returns its
// path. The plugin file itself carries no test driver; drivers are written as
// separate .mjs files next to it and import the default export.
func writePlugin(t *testing.T, dir string, opts *generator.ToolWrapperOptions) {
	t.Helper()
	src, err := generator.GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plugin.mjs"), []byte(src), 0600); err != nil {
		t.Fatalf("write plugin.mjs: %v", err)
	}
}

// runNode runs a .mjs file with node (cwd set to its directory so the ESM
// resolver finds sharedDir/node_modules) and returns stdout, failing the test
// on a non-zero exit with the captured stderr.
func runNode(t *testing.T, mjsPath string) string {
	t.Helper()
	cmd := exec.Command("node", mjsPath)
	cmd.Dir = filepath.Dir(mjsPath)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Fatalf("node exited non-zero:\n%s", exitErr.Stderr)
		}
		t.Fatalf("node: %v", err)
	}
	return string(out)
}

// runNodeExpectFailure runs a .mjs file that is expected to exit non-zero and
// returns the captured stderr.
func runNodeExpectFailure(t *testing.T, mjsPath string) string {
	t.Helper()
	cmd := exec.Command("node", mjsPath)
	cmd.Dir = filepath.Dir(mjsPath)
	out, err := cmd.Output()
	if err == nil {
		t.Fatalf("node exited zero, expected a failure; stdout:\n%s", out)
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("node: %v", err)
	}
	return string(exitErr.Stderr)
}

// ── introspection ────────────────────────────────────────────────────────────

const introspectDriver = `
import plugin from "./plugin.mjs";

const namespaces = [];
const added = [];
const editor = {
  namespace: (ns) => { namespaces.push(ns); },
  add: (tool) => { added.push(tool); },
  list: () => [...added],
  get: (id) => added.find((tool) => "taskflow_" + tool.name === id),
  update: () => {},
  remove: () => {},
};
const ctx = {
  options: {},
  location: { directory: process.cwd(), project: { id: "test", directory: process.cwd(), canonical: process.cwd() } },
  session: { get: async () => { throw new Error("unused in introspection"); } },
  tool: { transform: async (cb) => { cb(editor); } },
};
await plugin.setup(ctx);
console.log(JSON.stringify({
  namespaces,
  tools: added.map((tool) => {
    const props = tool.input && tool.input.properties ? tool.input.properties : {};
    const properties = {};
    for (const [argName, prop] of Object.entries(props)) {
      const shape = { type: prop.type };
      if (prop.enum) shape.enum = prop.enum;
      properties[argName] = shape;
    }
    return {
      name: "taskflow_" + tool.name,
      description: tool.description,
      namespace: tool.options && tool.options.namespace,
      properties,
      required: (tool.input && tool.input.required) || [],
    };
  }),
}));
`

// argShape mirrors one serialized property: its JSON Schema type and, for
// sort_by, the enum values.
type argShape struct {
	Type string   `json:"type"`
	Enum []string `json:"enum"`
}

// toolShape mirrors one serialized tool registration.
type toolShape struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Namespace   string              `json:"namespace"`
	Properties  map[string]argShape `json:"properties"`
	Required    []string            `json:"required"`
}

// namespaceShape mirrors one editor.namespace registration.
type namespaceShape struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// pluginIntrospection is the parsed introspection output.
type pluginIntrospection struct {
	Namespaces []namespaceShape `json:"namespaces"`
	Tools      []toolShape      `json:"tools"`
}

// toolMap indexes the introspection by effective tool name.
func toolMap(t *testing.T, intro pluginIntrospection) map[string]toolShape {
	t.Helper()
	m := make(map[string]toolShape, len(intro.Tools))
	for _, tool := range intro.Tools {
		if _, dup := m[tool.Name]; dup {
			t.Errorf("duplicate tool name %q", tool.Name)
		}
		m[tool.Name] = tool
	}
	return m
}

// runGeneratedPlugin generates the plugin into dir, runs the introspection
// driver against it in node, and returns the parsed registration shapes.
func runGeneratedPlugin(t *testing.T, dir string, opts *generator.ToolWrapperOptions) pluginIntrospection {
	t.Helper()
	writePlugin(t, dir, opts)
	driverPath := filepath.Join(dir, "driver.mjs")
	if err := os.WriteFile(driverPath, []byte(introspectDriver), 0600); err != nil {
		t.Fatalf("write driver.mjs: %v", err)
	}
	out := runNode(t, driverPath)
	var intro pluginIntrospection
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &intro); err != nil {
		t.Fatalf("parse node output: %v\nraw: %s", err, out)
	}
	return intro
}

// ── introspection tests ──────────────────────────────────────────────────────

func TestPluginLoadsInNode(t *testing.T) {
	intro := runGeneratedPlugin(t, testDir(t), generator.DefaultToolWrapperOptions())
	if len(intro.Tools) == 0 {
		t.Fatal("expected at least one registered tool, got none")
	}
}

func TestPluginRegistersAllTools(t *testing.T) {
	intro := runGeneratedPlugin(t, testDir(t), generator.DefaultToolWrapperOptions())

	want := []string{
		"taskflow_add", "taskflow_block", "taskflow_complete", "taskflow_delete",
		"taskflow_list_all", "taskflow_list_blocked", "taskflow_list_done",
		"taskflow_list_status_in_progress", "taskflow_list_status_todo",
		"taskflow_reset_timedout", "taskflow_start", "taskflow_unblock",
	}
	if len(intro.Tools) != len(want) {
		t.Errorf("registered %d tools, want %d", len(intro.Tools), len(want))
	}
	shapes := toolMap(t, intro)
	for _, name := range want {
		if _, ok := shapes[name]; !ok {
			t.Errorf("missing tool %q", name)
		}
	}
}

func TestPluginNamespaceRegistered(t *testing.T) {
	intro := runGeneratedPlugin(t, testDir(t), generator.DefaultToolWrapperOptions())

	if len(intro.Namespaces) != 1 {
		t.Fatalf("namespace registrations = %d, want 1", len(intro.Namespaces))
	}
	ns := intro.Namespaces[0]
	if ns.Name != "taskflow" {
		t.Errorf("namespace name = %q, want taskflow", ns.Name)
	}
	if !strings.Contains(ns.Description, "Taskflow task management backed by the taskflow CLI") {
		t.Errorf("namespace description %q missing intro text", ns.Description)
	}

	// Every tool must carry the namespace option so its effective name is
	// taskflow_<name>.
	shapes := toolMap(t, intro)
	for name, shape := range shapes {
		if shape.Namespace != "taskflow" {
			t.Errorf("tool %q namespace option = %q, want taskflow", name, shape.Namespace)
		}
	}
}

func TestPluginDescriptions(t *testing.T) {
	intro := runGeneratedPlugin(t, testDir(t), generator.DefaultToolWrapperOptions())
	shapes := toolMap(t, intro)

	cases := []struct {
		tool string
		want string
	}{
		{"taskflow_add", "Add a new task to the task list"},
		{"taskflow_block", "Block a task by providing its ID and a reason"},
		{"taskflow_complete", "Mark a task as completed"},
		{"taskflow_delete", "Soft delete a task"},
		{"taskflow_list_all", "List all tasks"},
		{"taskflow_list_blocked", "List blocked tasks"},
		{"taskflow_list_done", "List completed tasks"},
		{"taskflow_list_status_in_progress", "List in-progress tasks"},
		{"taskflow_list_status_todo", "List todo tasks"},
		{"taskflow_reset_timedout", "Reset timed out tasks"},
		{"taskflow_start", "Start working on a task"},
		{"taskflow_unblock", "Unblock a previously blocked task"},
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

func TestPluginArgTypes(t *testing.T) {
	intro := runGeneratedPlugin(t, testDir(t), generator.DefaultToolWrapperOptions())
	shapes := toolMap(t, intro)

	cases := []struct {
		tool    string
		arg     string
		wantTyp string
	}{
		{"taskflow_add", "id", "string"},
		{"taskflow_add", "title", "string"},
		{"taskflow_add", "description", "string"},
		{"taskflow_add", "milestone", "string"},
		{"taskflow_add", "actor", "string"},
		{"taskflow_block", "id", "string"},
		{"taskflow_block", "reason", "string"},
		{"taskflow_complete", "id", "string"},
		{"taskflow_delete", "id", "string"},
		{"taskflow_start", "id", "string"},
		{"taskflow_list_all", "milestone", "string"},
		{"taskflow_list_all", "actor", "string"},
		{"taskflow_list_all", "sort_by", "string"},
		{"taskflow_list_all", "limit", "number"},
		{"taskflow_list_all", "offset", "number"},
		{"taskflow_reset_timedout", "minutes", "number"},
		{"taskflow_unblock", "id", "string"},
		{"taskflow_unblock", "description", "string"},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s.%s", tc.tool, tc.arg), func(t *testing.T) {
			s, ok := shapes[tc.tool]
			if !ok {
				t.Fatalf("tool %q not found", tc.tool)
			}
			got, ok := s.Properties[tc.arg]
			if !ok {
				t.Fatalf("arg %q not found in tool %q (have: %v)", tc.arg, tc.tool, s.Properties)
			}
			if got.Type != tc.wantTyp {
				t.Errorf("arg %q type = %q, want %q", tc.arg, got.Type, tc.wantTyp)
			}
		})
	}
}

func TestPluginArgCounts(t *testing.T) {
	intro := runGeneratedPlugin(t, testDir(t), generator.DefaultToolWrapperOptions())
	shapes := toolMap(t, intro)

	cases := []struct {
		tool      string
		wantCount int
	}{
		{"taskflow_add", 5},          // id, title, description, milestone, actor
		{"taskflow_block", 2},        // id, reason
		{"taskflow_complete", 1},     // id
		{"taskflow_delete", 1},       // id
		{"taskflow_list_all", 5},     // milestone, actor, sort_by, limit, offset
		{"taskflow_list_blocked", 5}, // milestone, actor, sort_by, limit, offset
		{"taskflow_list_done", 5},
		{"taskflow_list_status_in_progress", 5},
		{"taskflow_list_status_todo", 5},
		{"taskflow_reset_timedout", 1}, // minutes
		{"taskflow_start", 1},          // id
		{"taskflow_unblock", 2},        // id, description
	}
	for _, tc := range cases {
		t.Run(tc.tool, func(t *testing.T) {
			s, ok := shapes[tc.tool]
			if !ok {
				t.Fatalf("tool %q not found", tc.tool)
			}
			if got := len(s.Properties); got != tc.wantCount {
				t.Errorf("arg count = %d, want %d (args: %v)", got, tc.wantCount, s.Properties)
			}
		})
	}
}

func TestPluginRequiredArgs(t *testing.T) {
	intro := runGeneratedPlugin(t, testDir(t), generator.DefaultToolWrapperOptions())
	shapes := toolMap(t, intro)

	cases := []struct {
		tool string
		want []string
	}{
		{"taskflow_add", []string{"id", "title", "description", "milestone"}},
		{"taskflow_block", []string{"id", "reason"}},
		{"taskflow_complete", []string{"id"}},
		{"taskflow_delete", []string{"id"}},
		{"taskflow_list_all", []string{}},
		{"taskflow_list_blocked", []string{}},
		{"taskflow_list_done", []string{}},
		{"taskflow_list_status_in_progress", []string{}},
		{"taskflow_list_status_todo", []string{}},
		{"taskflow_reset_timedout", []string{}},
		{"taskflow_start", []string{"id"}},
		{"taskflow_unblock", []string{"id"}},
	}
	for _, tc := range cases {
		t.Run(tc.tool, func(t *testing.T) {
			s, ok := shapes[tc.tool]
			if !ok {
				t.Fatalf("tool %q not found", tc.tool)
			}
			got := s.Required
			if got == nil {
				got = []string{}
			}
			if len(got) != len(tc.want) {
				t.Fatalf("required = %v, want %v", got, tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Fatalf("required = %v, want %v", got, tc.want)
				}
			}
		})
	}
}

func TestPluginSortByEnum(t *testing.T) {
	intro := runGeneratedPlugin(t, testDir(t), generator.DefaultToolWrapperOptions())
	shapes := toolMap(t, intro)

	// The sort_by enum must mirror db.ValidSortKeys exactly.
	wantEnum := []string{"status", "milestone", "created", "updated", "id", "title", "description", "actor"}

	listTools := []string{
		"taskflow_list_all",
		"taskflow_list_blocked",
		"taskflow_list_done",
		"taskflow_list_status_in_progress",
		"taskflow_list_status_todo",
	}
	for _, name := range listTools {
		t.Run(name, func(t *testing.T) {
			s, ok := shapes[name]
			if !ok {
				t.Fatalf("tool %q not found", name)
			}
			sortBy, ok := s.Properties["sort_by"]
			if !ok {
				t.Fatalf("tool %q has no sort_by arg", name)
			}
			if len(sortBy.Enum) != len(wantEnum) {
				t.Fatalf("sort_by enum = %v, want %v", sortBy.Enum, wantEnum)
			}
			for i := range wantEnum {
				if sortBy.Enum[i] != wantEnum[i] {
					t.Fatalf("sort_by enum = %v, want %v", sortBy.Enum, wantEnum)
				}
			}
		})
	}

	// Non-list tools must not carry enums.
	add, ok := shapes["taskflow_add"]
	if !ok {
		t.Fatal("tool taskflow_add not found")
	}
	for argName, prop := range add.Properties {
		if len(prop.Enum) != 0 {
			t.Errorf("taskflow_add arg %q carries an unexpected enum %v", argName, prop.Enum)
		}
	}
}

// ── execution tests ──────────────────────────────────────────────────────────

// TestPluginExecutesRealBinary drives the generated plugin's tools through
// full taskflow CLI runs against the binary built by TestMain. The DB lives
// under TASKFLOW_DIR (an absolute scratch path); the fake session resolver
// points the spawn at an empty scratch directory, and the test also proves
// the per-session directory really drives the spawn cwd by making one
// session's directory nonexistent (the spawn must fail) and another session's
// lookup throw (the plugin-directory fallback must be used).
func TestPluginExecutesRealBinary(t *testing.T) {
	dir := testDir(t)
	writePlugin(t, dir, generator.DefaultToolWrapperOptions())

	scratch := t.TempDir()
	sessionDir := filepath.Join(scratch, "sessiondir")
	missingDir := filepath.Join(scratch, "no-such-session-dir")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("create session dir: %v", err)
	}

	t.Setenv("TASKFLOW_DIR", filepath.Join(scratch, "data"))
	t.Setenv("TASKFLOW_TEST_BINARY", testBinary)
	t.Setenv("TASKFLOW_TEST_SESSION_DIR", sessionDir)
	t.Setenv("TASKFLOW_TEST_MISSING_DIR", missingDir)

	driver := `
import plugin from "./plugin.mjs";

const assert = (cond, msg) => { if (!cond) throw new Error("assert failed: " + msg); };

const added = [];
const editor = {
  namespace: () => {},
  add: (tool) => { added.push(tool); },
  list: () => [...added],
  get: (id) => added.find((tool) => "taskflow_" + tool.name === id),
  update: () => {},
  remove: () => {},
};
const byName = (name) => {
  const tool = added.find((x) => x.name === name);
  if (!tool) throw new Error("missing tool " + name);
  return tool;
};
const session = {
  get: async (input) => {
    if (input && input.sessionID === "ses_bogus_cwd") {
      return { location: { directory: process.env.TASKFLOW_TEST_MISSING_DIR } };
    }
    if (input && input.sessionID === "ses_throws") {
      throw new Error("session lookup unavailable");
    }
    return { location: { directory: process.env.TASKFLOW_TEST_SESSION_DIR } };
  },
};
const ctx = {
  options: { binaryPath: process.env.TASKFLOW_TEST_BINARY },
  location: { directory: process.cwd(), project: { id: "test", directory: process.cwd(), canonical: process.cwd() } },
  session,
  tool: { transform: async (cb) => { cb(editor); } },
};
await plugin.setup(ctx);

const baseCtx = { sessionID: "ses_test", agent: "build", messageID: "msg_1", id: "call_1", signal: new AbortController().signal, progress: async () => {} };
const parse = (text) => JSON.parse(text);

// 1. add: the CLI prints a human-readable report.
let res = await byName("add").execute({ id: "t1", title: "Sample", description: "Sample task", milestone: "demo" }, baseCtx);
assert(res.content.includes("Task added successfully"), "add output missing success line: " + res.content);
assert(res.content.includes("ID: t1"), "add output missing ID: t1");
assert(res.content.includes("Status: todo"), "add output missing Status: todo");

// 2. start moves the task to in_progress.
res = await byName("start").execute({ id: "t1" }, baseCtx);
assert(res.content.includes("Status: in_progress"), "start output missing in_progress: " + res.content);

// 3. list_status_in_progress shows the task (list output is JSON).
res = await byName("list_status_in_progress").execute({}, baseCtx);
const inProgress = parse(res.content);
assert(Array.isArray(inProgress.tasks) && inProgress.tasks.length === 1, "expected one in-progress task: " + res.content);
assert(inProgress.tasks[0].ID === "t1", "in-progress task ID mismatch: " + JSON.stringify(inProgress.tasks[0]));
assert(inProgress.tasks[0].Status === "in_progress", "in-progress task status mismatch");

// 4. complete marks the task done.
res = await byName("complete").execute({ id: "t1" }, baseCtx);
assert(res.content.includes("Status: done"), "complete output missing done: " + res.content);

// 5. list_blocked shows nothing.
res = await byName("list_blocked").execute({}, baseCtx);
const blocked = parse(res.content);
assert(Array.isArray(blocked.tasks) && blocked.tasks.length === 0, "expected zero blocked tasks: " + res.content);

// 6. reset_timedout succeeds with nothing timed out.
res = await byName("reset_timedout").execute({ minutes: 30 }, baseCtx);
assert(res.content.includes("No tasks were timed out.") || /Reset \d+ timed out/.test(res.content), "unexpected reset output: " + res.content);

// 7. A session directory that does not exist must drive the spawn cwd: the
// CLI spawn has to fail, which proves the per-session directory is used.
let bogusFailed = false;
try {
  await byName("complete").execute({ id: "t1" }, { ...baseCtx, sessionID: "ses_bogus_cwd" });
} catch (err) {
  bogusFailed = true;
  const msg = String(err && err.message ? err.message : err);
  assert(msg.startsWith("taskflow complete failed:"), "unexpected error for bogus session dir: " + msg);
}
assert(bogusFailed, "expected the bogus session directory to fail the spawn");

// 8. A throwing session lookup falls back to the plugin directory and the
// tool still runs.
res = await byName("list_done").execute({}, { ...baseCtx, sessionID: "ses_throws" });
const done = parse(res.content);
assert(Array.isArray(done.tasks) && done.tasks.length === 1 && done.tasks[0].ID === "t1", "expected t1 in done tasks: " + res.content);

console.log(JSON.stringify({ ok: true }));
`
	driverPath := filepath.Join(dir, "driver.mjs")
	if err := os.WriteFile(driverPath, []byte(driver), 0600); err != nil {
		t.Fatalf("write driver.mjs: %v", err)
	}
	out := runNode(t, driverPath)
	if !strings.Contains(out, `"ok":true`) {
		t.Fatalf("driver did not report ok: %s", out)
	}
}

// TestPluginSurfacesCliErrors checks that a failing CLI run surfaces as a
// rejected tool promise carrying the CLI's JSON error payload: executing
// complete with an unknown task ID must make node exit non-zero with the
// wrapped error on stderr.
func TestPluginSurfacesCliErrors(t *testing.T) {
	dir := testDir(t)
	writePlugin(t, dir, generator.DefaultToolWrapperOptions())

	scratch := t.TempDir()
	t.Setenv("TASKFLOW_DIR", filepath.Join(scratch, "data"))
	t.Setenv("TASKFLOW_TEST_BINARY", testBinary)

	driver := `
import plugin from "./plugin.mjs";

const added = [];
const editor = {
  namespace: () => {},
  add: (tool) => { added.push(tool); },
  list: () => [...added],
  get: (id) => added.find((tool) => "taskflow_" + tool.name === id),
  update: () => {},
  remove: () => {},
};
const ctx = {
  options: { binaryPath: process.env.TASKFLOW_TEST_BINARY },
  location: { directory: process.cwd(), project: { id: "test", directory: process.cwd(), canonical: process.cwd() } },
  session: { get: async () => ({ location: { directory: process.cwd() } }) },
  tool: { transform: async (cb) => { cb(editor); } },
};
await plugin.setup(ctx);

const baseCtx = { sessionID: "ses_test", agent: "build", messageID: "msg_1", id: "call_1", signal: new AbortController().signal, progress: async () => {} };

const tool = added.find((x) => x.name === "complete");
if (!tool) throw new Error("missing tool complete");
try {
  await tool.execute({ id: "nope-missing" }, baseCtx);
  console.error("EXPECTED_FAILURE: the tool call unexpectedly succeeded");
  process.exit(1);
} catch (err) {
  console.error("EXPECTED_FAILURE: " + (err && err.message ? err.message : String(err)));
  process.exit(1);
}
`
	driverPath := filepath.Join(dir, "driver.mjs")
	if err := os.WriteFile(driverPath, []byte(driver), 0600); err != nil {
		t.Fatalf("write driver.mjs: %v", err)
	}
	stderr := runNodeExpectFailure(t, driverPath)

	// The plugin wraps the CLI failure as "taskflow complete failed:" plus the
	// CLI's JSON error body ({"status":"error",...} with the task-not-found
	// message; verified against cmd/complete.go + internal/clierr).
	for _, want := range []string{
		"EXPECTED_FAILURE: taskflow complete failed:",
		`"status":"error"`,
		"task with ID",
	} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr missing %q; stderr:\n%s", want, stderr)
		}
	}
}

// TestPluginRejectsMissingBinary checks the version gate: with a binaryPath
// that cannot be spawned, executing any tool must reject with the plugin's
// missing-binary message before the CLI runs.
func TestPluginRejectsMissingBinary(t *testing.T) {
	dir := testDir(t)
	writePlugin(t, dir, generator.DefaultToolWrapperOptions())

	scratch := t.TempDir()
	t.Setenv("TASKFLOW_DIR", filepath.Join(scratch, "data"))

	driver := `
import plugin from "./plugin.mjs";

const added = [];
const editor = {
  namespace: () => {},
  add: (tool) => { added.push(tool); },
  list: () => [...added],
  get: (id) => added.find((tool) => "taskflow_" + tool.name === id),
  update: () => {},
  remove: () => {},
};
const ctx = {
  options: { binaryPath: "definitely-not-a-real-binary-tf" },
  location: { directory: process.cwd(), project: { id: "test", directory: process.cwd(), canonical: process.cwd() } },
  session: { get: async () => ({ location: { directory: process.cwd() } }) },
  tool: { transform: async (cb) => { cb(editor); } },
};
await plugin.setup(ctx);

const baseCtx = { sessionID: "ses_test", agent: "build", messageID: "msg_1", id: "call_1", signal: new AbortController().signal, progress: async () => {} };

const tool = added.find((x) => x.name === "complete");
if (!tool) throw new Error("missing tool complete");
try {
  await tool.execute({ id: "whatever" }, baseCtx);
  console.error("EXPECTED_FAILURE: the tool call unexpectedly succeeded");
  process.exit(1);
} catch (err) {
  console.error("EXPECTED_FAILURE: " + (err && err.message ? err.message : String(err)));
  process.exit(1);
}
`
	driverPath := filepath.Join(dir, "driver.mjs")
	if err := os.WriteFile(driverPath, []byte(driver), 0600); err != nil {
		t.Fatalf("write driver.mjs: %v", err)
	}
	stderr := runNodeExpectFailure(t, driverPath)

	if !strings.Contains(stderr, "is missing, not on PATH, or too old") {
		t.Errorf("stderr missing the missing-binary message; stderr:\n%s", stderr)
	}
}

// ── version-gate tests (fake binary) ─────────────────────────────────────────

// fakeEditorSetup is the driver preamble shared by the fake-binary tests: it
// imports the plugin, collects the registered tools through a fake editor,
// and exposes byName plus a base call context. Plugin options come from the
// TASKFLOW_TEST_PLUGIN_OPTIONS env var (an empty object when unset); the
// FAKE_* variables reach the fake binary through the inherited environment.
const fakeEditorSetup = `
import plugin from "./plugin.mjs";
import { readFileSync } from "node:fs";

const added = [];
const editor = {
  namespace: () => {},
  add: (tool) => { added.push(tool); },
  list: () => [...added],
  get: (id) => added.find((tool) => "taskflow_" + tool.name === id),
  update: () => {},
  remove: () => {},
};
const byName = (name) => {
  const tool = added.find((x) => x.name === name);
  if (!tool) throw new Error("missing tool " + name);
  return tool;
};
const pluginOptions = JSON.parse(process.env.TASKFLOW_TEST_PLUGIN_OPTIONS || "{}");
const ctx = {
  options: pluginOptions,
  location: { directory: process.cwd(), project: { id: "test", directory: process.cwd(), canonical: process.cwd() } },
  session: { get: async () => ({ location: { directory: process.cwd() } }) },
  tool: { transform: async (cb) => { cb(editor); } },
};
await plugin.setup(ctx);

const baseCtx = { sessionID: "ses_test", agent: "build", messageID: "msg_1", id: "call_1", signal: new AbortController().signal, progress: async () => {} };
`

// writeFakeBinaryDriver generates the plugin, writes driverSource next to it
// as driver.mjs, and returns the driver path. Call runNode or
// runNodeExpectFailure on the returned path; the FAKE_* and
// TASKFLOW_TEST_PLUGIN_OPTIONS env vars are set by the test with t.Setenv and
// reach node and the fake binary through process inheritance.
func writeFakeBinaryDriver(t *testing.T, dir string, opts *generator.ToolWrapperOptions, driverSource string) string {
	t.Helper()
	writePlugin(t, dir, opts)
	driverPath := filepath.Join(dir, "driver.mjs")
	if err := os.WriteFile(driverPath, []byte(driverSource), 0600); err != nil {
		t.Fatalf("write driver.mjs: %v", err)
	}
	return driverPath
}

// TestPluginVersionGateRejectsIncompatibleBinary generates the plugin for
// taskflow 1.0.0 and points it at the fake binary reporting 9.9.9: executing
// any tool must reject with the generated-for message and the regeneration
// instruction.
func TestPluginVersionGateRejectsIncompatibleBinary(t *testing.T) {
	dir := testDir(t)
	t.Setenv("TASKFLOW_TEST_PLUGIN_OPTIONS", fmt.Sprintf(`{"binaryPath":%q}`, fakeBinary))
	t.Setenv("FAKE_VERSION", "taskflow version 9.9.9")

	driver := fakeEditorSetup + `
try {
  await byName("complete").execute({ id: "whatever" }, baseCtx);
  console.error("EXPECTED_FAILURE: the tool call unexpectedly succeeded");
  process.exit(1);
} catch (err) {
  console.error("EXPECTED_FAILURE: " + (err && err.message ? err.message : String(err)));
  process.exit(1);
}
`
	stderr := runNodeExpectFailure(t, writeFakeBinaryDriver(t, dir, &generator.ToolWrapperOptions{BinaryPath: "taskflow", Version: "1.0.0"}, driver))

	for _, want := range []string{
		"generated for taskflow 1.0.0",
		"regenerate the plugin",
	} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr missing %q; stderr:\n%s", want, stderr)
		}
	}
}

// TestPluginVersionGateRejectsGarbageVersion points the plugin at the fake
// binary printing an unparseable version: executing a tool must reject with
// the not-a-semver message.
func TestPluginVersionGateRejectsGarbageVersion(t *testing.T) {
	dir := testDir(t)
	t.Setenv("TASKFLOW_TEST_PLUGIN_OPTIONS", fmt.Sprintf(`{"binaryPath":%q}`, fakeBinary))
	t.Setenv("FAKE_VERSION", "not a semver at all")

	driver := fakeEditorSetup + `
try {
  await byName("complete").execute({ id: "whatever" }, baseCtx);
  console.error("EXPECTED_FAILURE: the tool call unexpectedly succeeded");
  process.exit(1);
} catch (err) {
  console.error("EXPECTED_FAILURE: " + (err && err.message ? err.message : String(err)));
  process.exit(1);
}
`
	stderr := runNodeExpectFailure(t, writeFakeBinaryDriver(t, dir, &generator.ToolWrapperOptions{BinaryPath: "taskflow", Version: "1.0.0"}, driver))

	if !strings.Contains(stderr, "not a semver") {
		t.Errorf("stderr missing the unparseable-version message; stderr:\n%s", stderr)
	}
}

// TestPluginVersionGateProbesOnce runs one tool twice against the fake
// binary: the invocation log must show exactly one version probe (the
// per-process cached checkVersion) and two list runs.
func TestPluginVersionGateProbesOnce(t *testing.T) {
	dir := testDir(t)
	logPath := filepath.Join(dir, "fake-invocations.log")
	t.Setenv("TASKFLOW_TEST_PLUGIN_OPTIONS", fmt.Sprintf(`{"binaryPath":%q}`, fakeBinary))
	t.Setenv("FAKE_LOG", logPath)

	driver := fakeEditorSetup + `
const logPath = process.env.FAKE_LOG;
if (!logPath) throw new Error("FAKE_LOG not set");
const countLines = (name) => readFileSync(logPath, "utf-8").split("\n").filter((line) => line.trim() === name).length;
const first = await byName("list_all").execute({}, baseCtx);
if (!first.content.includes("\"tasks\"")) throw new Error("unexpected first list output: " + first.content);
const second = await byName("list_all").execute({}, baseCtx);
if (!second.content.includes("\"tasks\"")) throw new Error("unexpected second list output: " + second.content);
const versionLines = countLines("version");
const listLines = countLines("list");
if (versionLines !== 1) throw new Error("expected exactly one version probe, got " + versionLines);
if (listLines !== 2) throw new Error("expected two list invocations, got " + listLines);
console.log(JSON.stringify({ ok: true, versionProbes: versionLines, listCalls: listLines }));
`
	out := runNode(t, writeFakeBinaryDriver(t, dir, &generator.ToolWrapperOptions{BinaryPath: "taskflow", Version: "1.0.0"}, driver))
	if !strings.Contains(out, `"ok":true`) {
		t.Fatalf("driver did not report ok: %s", out)
	}

	// Double-check the log from the Go side: one version probe, two list runs.
	logBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read fake invocation log: %v", err)
	}
	logged := string(logBytes)
	if count := strings.Count(logged, "version\n"); count != 1 {
		t.Errorf("fake binary version probe count = %d, want 1 (log: %q)", count, logged)
	}
	if count := strings.Count(logged, "list\n"); count != 2 {
		t.Errorf("fake binary list invocation count = %d, want 2 (log: %q)", count, logged)
	}
}

// TestPluginUsesDefaultBinaryWhenOptionAbsent generates the plugin with the
// fake binary path as BinaryPath and sets up with empty plugin options: the
// DEFAULT_BINARY fallback must point execution at the fake binary.
func TestPluginUsesDefaultBinaryWhenOptionAbsent(t *testing.T) {
	dir := testDir(t)

	// TASKFLOW_TEST_PLUGIN_OPTIONS stays unset, so the driver's fake ctx
	// passes options: {}.
	driver := fakeEditorSetup + `
const res = await byName("list_all").execute({}, baseCtx);
if (!res.content.includes("\"tasks\"")) throw new Error("unexpected list output: " + res.content);
const listed = JSON.parse(res.content);
if (listed.total !== 0) throw new Error("expected total 0 from the fake binary");
console.log(JSON.stringify({ ok: true }));
`
	out := runNode(t, writeFakeBinaryDriver(t, dir, &generator.ToolWrapperOptions{BinaryPath: fakeBinary, Version: "0.0.0"}, driver))
	if !strings.Contains(out, `"ok":true`) {
		t.Fatalf("driver did not report ok: %s", out)
	}
}

// TestPluginCancelsOnAbortedSignal lets the fake binary stall for 2 seconds
// and aborts the call context's signal 300 ms in: the execute promise must
// reject with the cancellation message and the node process must finish
// normally (no wall-clock assertion).
func TestPluginCancelsOnAbortedSignal(t *testing.T) {
	dir := testDir(t)
	t.Setenv("TASKFLOW_TEST_PLUGIN_OPTIONS", fmt.Sprintf(`{"binaryPath":%q}`, fakeBinary))
	t.Setenv("FAKE_SLEEP_MS", "2000")

	driver := fakeEditorSetup + `
const controller = new AbortController();
setTimeout(() => controller.abort(), 300);
let message = "";
try {
  await byName("complete").execute({ id: "whatever" }, { ...baseCtx, signal: controller.signal });
  console.error("EXPECTED_FAILURE: the tool call unexpectedly succeeded");
  process.exit(1);
} catch (err) {
  message = String(err && err.message ? err.message : err);
}
if (!/was canceled/.test(message)) {
  console.error("EXPECTED_CANCELLATION but got: " + message);
  process.exit(1);
}
console.log(JSON.stringify({ ok: true, message: message }));
`
	out := runNode(t, writeFakeBinaryDriver(t, dir, generator.DefaultToolWrapperOptions(), driver))
	if !strings.Contains(out, `"ok":true`) {
		t.Fatalf("driver did not report ok: %s", out)
	}
	if !strings.Contains(out, "was canceled") {
		t.Errorf("driver output missing the cancellation message: %s", out)
	}
}
