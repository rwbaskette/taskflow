package generator

import (
	"strings"
	"testing"
)

// generateToolWrapper is a helper that fails the test on a generation error
// and returns the generated plugin source.
func generateToolWrapper(t *testing.T, opts *ToolWrapperOptions) string {
	t.Helper()
	result, err := GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
	}
	return result
}

func TestGenerateToolWrapper(t *testing.T) {
	tests := []struct {
		name         string
		opts         *ToolWrapperOptions
		wantContains string
		wantErr      bool
	}{
		{
			name: "default options",
			opts: &ToolWrapperOptions{
				BinaryPath: "taskflow",
			},
			wantContains: "export default Plugin.define({",
			wantErr:      false,
		},
		{
			name: "custom binary path",
			opts: &ToolWrapperOptions{
				BinaryPath: "myapp",
			},
			wantContains: `const DEFAULT_BINARY = "myapp";`,
			wantErr:      false,
		},
		{
			name:         "nil options uses defaults",
			opts:         nil,
			wantContains: `const DEFAULT_BINARY = "taskflow";`,
			wantErr:      false,
		},
		{
			name: "empty binary path uses default",
			opts: &ToolWrapperOptions{
				BinaryPath: "",
			},
			wantContains: `const DEFAULT_BINARY = "taskflow";`,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateToolWrapper(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToolWrapper() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.wantContains != "" && !strings.Contains(result, tt.wantContains) {
				t.Errorf("GenerateToolWrapper() result does not contain %q", tt.wantContains)
			}
		})
	}
}

func TestGenerateToolWrapperPluginEntrypoints(t *testing.T) {
	result := generateToolWrapper(t, DefaultToolWrapperOptions())

	// The output must be an OpenCode 2.0 plugin: a default-exported
	// Plugin.define whose setup registers tools through ctx.tool.transform.
	wantContains := []string{
		`import { Plugin } from "@opencode/plugin";`,
		`import { execFile } from "node:child_process";`,
		`export default Plugin.define({`,
		`id: "taskflow"`,
		`await ctx.tool.transform((editor) =>`,
		`editor.namespace({ name: NAMESPACE`,
		`const NAMESPACE = "taskflow";`,
	}
	for _, want := range wantContains {
		if !strings.Contains(result, want) {
			t.Errorf("GenerateToolWrapper() output missing: %s", want)
		}
	}

	// The V1 wrapper format must not come back in any form.
	mustNotContain := []string{
		"@opencode-ai/plugin",
		"tool({",
		"tool.schema.",
		"execFileSync",
	}
	for _, bad := range mustNotContain {
		if strings.Contains(result, bad) {
			t.Errorf("GenerateToolWrapper() output must not contain: %s", bad)
		}
	}
}

func TestGenerateToolWrapperToolBlockCounts(t *testing.T) {
	result := generateToolWrapper(t, DefaultToolWrapperOptions())
	wantTools := len(getToolCommands())

	// Every registered tool emits exactly one of each structural element.
	patterns := []string{
		"editor.add({",
		`options: { namespace: NAMESPACE }`,
		"async execute(args, context)",
		"await checkVersion(binary);",
		"signal: context.signal",
		"cwd: await directoryFor(context.sessionID)",
		"input: {",
		"additionalProperties: false",
		`"type": "object"`,
		"JSON.stringify(payload)",
		"{ content: result }",
		"return { content: result };",
	}
	for _, pattern := range patterns {
		if count := strings.Count(result, pattern); count != wantTools {
			t.Errorf("GenerateToolWrapper() %q count = %d, want %d", pattern, count, wantTools)
		}
	}
}

func TestGenerateToolWrapperRequiredArrays(t *testing.T) {
	result := generateToolWrapper(t, DefaultToolWrapperOptions())

	// Exactly six tools carry a required array: add, block, complete,
	// delete, start, and unblock. The five list tools and reset_timedout
	// have none.
	if count := strings.Count(result, "required: ["); count != 6 {
		t.Errorf("GenerateToolWrapper() required array count = %d, want 6", count)
	}
	wantArrays := []string{
		`required: ["id", "title", "description", "milestone"],`, // add
		`required: ["id", "reason"],`,                            // block
	}
	for _, want := range wantArrays {
		if !strings.Contains(result, want) {
			t.Errorf("GenerateToolWrapper() output missing: %s", want)
		}
	}
	// complete, delete, start, and unblock each require only id.
	if count := strings.Count(result, `required: ["id"],`); count != 4 {
		t.Errorf(`GenerateToolWrapper() required: ["id"] count = %d, want 4`, count)
	}
}

func TestGenerateToolWrapperJSONSchemaTyping(t *testing.T) {
	result := generateToolWrapper(t, DefaultToolWrapperOptions())

	// Property objects are plain JSON Schema, not zod.
	if !strings.Contains(result, `{ type: "string"`) {
		t.Error(`GenerateToolWrapper() output missing: { type: "string"`)
	}
	if !strings.Contains(result, `{ type: "number"`) {
		t.Error(`GenerateToolWrapper() output missing: { type: "number"`)
	}

	// sort_by carries the db.ValidSortKeys enum on each of the five list
	// tools.
	wantEnum := `enum: ["status", "milestone", "created", "updated", "id", "title", "description", "actor"]`
	if count := strings.Count(result, wantEnum); count != 5 {
		t.Errorf("GenerateToolWrapper() sort_by enum count = %d, want 5", count)
	}

	// limit and offset use minimum 0 (2 args x 5 list tools); reset_timedout's
	// minutes uses minimum 1.
	if count := strings.Count(result, "minimum: 0"); count != 10 {
		t.Errorf("GenerateToolWrapper() minimum: 0 count = %d, want 10", count)
	}
	if count := strings.Count(result, "minimum: 1"); count != 1 {
		t.Errorf("GenerateToolWrapper() minimum: 1 count = %d, want 1", count)
	}
}

func TestGenerateToolWrapperVersionEmbed(t *testing.T) {
	t.Run("embedded version", func(t *testing.T) {
		result := generateToolWrapper(t, &ToolWrapperOptions{BinaryPath: "taskflow", Version: "0.1.0"})
		if !strings.Contains(result, `const TASKFLOW_PLUGIN_VERSION = "0.1.0";`) {
			t.Error(`GenerateToolWrapper() output missing: const TASKFLOW_PLUGIN_VERSION = "0.1.0";`)
		}
	})

	t.Run("nil options default to 0.0.0", func(t *testing.T) {
		result := generateToolWrapper(t, nil)
		if !strings.Contains(result, `const TASKFLOW_PLUGIN_VERSION = "0.0.0";`) {
			t.Error(`GenerateToolWrapper() output missing: const TASKFLOW_PLUGIN_VERSION = "0.0.0";`)
		}
	})

	t.Run("empty version defaults to 0.0.0", func(t *testing.T) {
		result := generateToolWrapper(t, &ToolWrapperOptions{BinaryPath: "taskflow"})
		if !strings.Contains(result, `const TASKFLOW_PLUGIN_VERSION = "0.0.0";`) {
			t.Error(`GenerateToolWrapper() output missing: const TASKFLOW_PLUGIN_VERSION = "0.0.0";`)
		}
	})

	t.Run("helper block", func(t *testing.T) {
		result := generateToolWrapper(t, DefaultToolWrapperOptions())
		for _, want := range []string{
			"function parseSemver(",
			"function compatible(",
			"async function checkVersion(",
			// Semver rule lines, pinned verbatim: 0.0.0 disables the rule;
			// majors must match; a 0-major also pins the minor.
			"if (runtimeVersion !== null) return;",
			"if (want.major === 0 && want.minor === 0 && want.patch === 0) return true;",
			"if (want.major !== got.major) return false;",
			"if (want.major === 0 && want.minor !== got.minor) return false;",
			// Cache only on success.
			"runtimeVersion = got;",
		} {
			if !strings.Contains(result, want) {
				t.Errorf("GenerateToolWrapper() output missing: %s", want)
			}
		}
	})

	t.Run("version errors name the regeneration command", func(t *testing.T) {
		result := generateToolWrapper(t, DefaultToolWrapperOptions())
		// All three version-gate errors (missing binary, unparseable output,
		// incompatible version) end with the regeneration instruction.
		want := "regenerate the plugin: taskflow tool-wrapper --output .opencode/plugins/taskflow.ts"
		if count := strings.Count(result, want); count != 3 {
			t.Errorf("GenerateToolWrapper() regeneration instruction count = %d, want 3", count)
		}
	})
}

func TestGenerateToolWrapperPluginOptions(t *testing.T) {
	result := generateToolWrapper(t, DefaultToolWrapperOptions())

	// The plugin reads the taskflow binary from plugin options, falling back
	// to the embedded default.
	wantContains := []string{
		"ctx.options ? ctx.options.binaryPath : undefined",
		"optionBinary.trim() : DEFAULT_BINARY",
	}
	for _, want := range wantContains {
		if !strings.Contains(result, want) {
			t.Errorf("GenerateToolWrapper() output missing: %s", want)
		}
	}
}

func TestGenerateToolWrapperDirectoryResolution(t *testing.T) {
	opts := DefaultToolWrapperOptions()
	result := generateToolWrapper(t, opts)

	// The spawn cwd is resolved per session from the session's worktree-aware
	// working directory, falling back to the plugin load location.
	wantContains := []string{
		"function makeDirectoryResolver(ctx, fallback)",
		"ctx.session.get({ sessionID })",
		"session.location.directory",
		"const pluginDirectory",
	}
	for _, want := range wantContains {
		if !strings.Contains(result, want) {
			t.Errorf("GenerateToolWrapper() output missing: %s", want)
		}
	}

	// The V1 directory plumbing must be gone.
	mustNotContain := []string{
		"cwd: context.directory",
		"context.worktree",
		"import path",
		"path.dirname",
	}
	for _, bad := range mustNotContain {
		if strings.Contains(result, bad) {
			t.Errorf("GenerateToolWrapper() output must not contain: %s", bad)
		}
	}
}

func TestGenerateToolWrapperCustomBinaryVersionCheck(t *testing.T) {
	result := generateToolWrapper(t, &ToolWrapperOptions{BinaryPath: "mybin", Version: "0.1.0"})

	if !strings.Contains(result, `const DEFAULT_BINARY = "mybin";`) {
		t.Error(`GenerateToolWrapper() should embed the custom binary: const DEFAULT_BINARY = "mybin";`)
	}

	// The default stays "taskflow".
	defaultResult := generateToolWrapper(t, DefaultToolWrapperOptions())
	if !strings.Contains(defaultResult, `const DEFAULT_BINARY = "taskflow";`) {
		t.Error(`GenerateToolWrapper() should default to: const DEFAULT_BINARY = "taskflow";`)
	}
}

func TestGenerateToolWrapperSubcommands(t *testing.T) {
	result := generateToolWrapper(t, DefaultToolWrapperOptions())

	// The five list tools share the "list" subcommand; every other tool maps
	// to its own CLI subcommand.
	if count := strings.Count(result, `["list", JSON.stringify(payload)]`); count != 5 {
		t.Errorf(`GenerateToolWrapper() ["list", ...] count = %d, want 5`, count)
	}
	singles := []string{
		`["reset-timedout", JSON.stringify(payload)]`,
		`["update", JSON.stringify(payload)]`,
		`["unblock", JSON.stringify(payload)]`,
		`["add", JSON.stringify(payload)]`,
		`["block", JSON.stringify(payload)]`,
		`["complete", JSON.stringify(payload)]`,
		`["delete", JSON.stringify(payload)]`,
	}
	for _, want := range singles {
		if count := strings.Count(result, want); count != 1 {
			t.Errorf("GenerateToolWrapper() %q count = %d, want 1", want, count)
		}
	}
}

func TestGenerateToolWrapperFixedStatuses(t *testing.T) {
	result := generateToolWrapper(t, DefaultToolWrapperOptions())

	// Fixed payload statuses: one per filtered list tool, plus start's
	// in_progress (so in_progress counts twice).
	statusCounts := []struct {
		pattern string
		want    int
	}{
		{`status: "blocked"`, 1},
		{`status: "done"`, 1},
		{`status: "in_progress"`, 2},
		{`status: "todo"`, 1},
	}
	for _, sc := range statusCounts {
		if count := strings.Count(result, sc.pattern); count != sc.want {
			t.Errorf("GenerateToolWrapper() %q count = %d, want %d", sc.pattern, count, sc.want)
		}
	}
}

func TestGenerateToolWrapperToolNames(t *testing.T) {
	result := generateToolWrapper(t, DefaultToolWrapperOptions())

	// Every tool name, with underscores preserved (reset_timedout).
	commandNames := []string{
		"add",
		"block",
		"complete",
		"delete",
		"list_all",
		"list_blocked",
		"list_done",
		"list_status_in_progress",
		"list_status_todo",
		"reset_timedout",
		"start",
		"unblock",
	}
	for _, name := range commandNames {
		pattern := `name: "` + name + `"`
		if !strings.Contains(result, pattern) {
			t.Errorf("GenerateToolWrapper() output missing: %s", pattern)
		}
	}

	// Twelve tool name fields; the namespace registration uses the
	// NAMESPACE constant, so it carries no quoted name field.
	if count := strings.Count(result, `name: "`); count != len(commandNames) {
		t.Errorf("GenerateToolWrapper() name field count = %d, want %d", count, len(commandNames))
	}
}

func TestGenerateToolWrapperContainsDescription(t *testing.T) {
	result := generateToolWrapper(t, DefaultToolWrapperOptions())

	descriptions := []string{
		"Add a new task to the task list",
		"Block a task by providing its ID and a reason",
		"Mark a task as completed",
		"Soft delete a task",
		"List all tasks",
		"List blocked tasks",
		"List completed tasks",
		"List in-progress tasks",
		"List todo tasks",
		"Reset timed out tasks to todo status",
		"Start working on a task",
		"Unblock a previously blocked task",
	}

	for _, desc := range descriptions {
		if !strings.Contains(result, desc) {
			t.Errorf("GenerateToolWrapper() output missing description: %s", desc)
		}
	}
}

func TestGenerateToolWrapperDoesNotAdvertiseAllOrListIDKnobs(t *testing.T) {
	// cmd/list.go accepts an "all" status and an "id" filter, but the
	// generated plugin must not advertise either: the filtered list_* tools
	// and the individual task tools cover those cases.
	result := generateToolWrapper(t, DefaultToolWrapperOptions())

	if strings.Contains(result, "Show all tasks including completed") {
		t.Error(`GenerateToolWrapper() output must not advertise the removed "all" list parameter`)
	}
	if strings.Contains(result, "args.all") {
		t.Error("GenerateToolWrapper() output must not reference args.all in a payload")
	}
	if strings.Contains(result, `all: { type: "boolean"`) {
		t.Error(`GenerateToolWrapper() output must not define an "all" argument`)
	}
	if strings.Contains(result, "specific task") {
		t.Error("GenerateToolWrapper() output must not advertise a get-task-by-id list knob")
	}

	// id: args.id belongs only to add, block, complete, delete, start, and
	// unblock - never to a list payload.
	if count := strings.Count(result, "id: args.id"); count != 6 {
		t.Errorf("GenerateToolWrapper() id: args.id count = %d, want 6", count)
	}
	// The sort knob is advertised on the list tools instead.
	if count := strings.Count(result, "sort_by: args.sort_by"); count != 5 {
		t.Errorf("GenerateToolWrapper() sort_by: args.sort_by count = %d, want 5", count)
	}
}

func TestGenerateToolWrapperCancellationAndErrorHelpers(t *testing.T) {
	result := generateToolWrapper(t, DefaultToolWrapperOptions())

	wantContains := []string{
		"was canceled",
		`taskflow ${args[0]} failed:`,
		"maxBuffer: 16 * 1024 * 1024",
		"windowsHide: true",
		// The abort/kill detection in runBinary.
		"options.signal?.aborted",
		`err.name === "AbortError"`,
	}
	for _, want := range wantContains {
		if !strings.Contains(result, want) {
			t.Errorf("GenerateToolWrapper() output missing: %s", want)
		}
	}
}

func TestDefaultToolWrapperOptions(t *testing.T) {
	opts := DefaultToolWrapperOptions()

	if opts.BinaryPath != "taskflow" {
		t.Errorf("DefaultToolWrapperOptions() BinaryPath = %v, want taskflow", opts.BinaryPath)
	}
	if opts.Version != "0.0.0" {
		t.Errorf("DefaultToolWrapperOptions() Version = %v, want 0.0.0", opts.Version)
	}
}

func TestGenerateToolWrapperNoErrorsOnValidInput(t *testing.T) {
	testCases := []struct {
		name string
		opts *ToolWrapperOptions
	}{
		{"default options", DefaultToolWrapperOptions()},
		{"custom binary", &ToolWrapperOptions{BinaryPath: "mybin"}},
		{"nil options", nil},
		{"empty binary", &ToolWrapperOptions{BinaryPath: ""}},
		{"empty version", &ToolWrapperOptions{BinaryPath: "taskflow", Version: ""}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := GenerateToolWrapper(tc.opts)
			if err != nil {
				t.Errorf("GenerateToolWrapper() unexpected error on valid input: %v", err)
			}
		})
	}
}

func TestGenerateToolWrapperReturnsNonEmpty(t *testing.T) {
	result := generateToolWrapper(t, DefaultToolWrapperOptions())

	if len(result) == 0 {
		t.Error("GenerateToolWrapper() should return non-empty string")
	}
}
