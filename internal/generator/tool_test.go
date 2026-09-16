package generator

import (
	"strings"
	"testing"
)

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
			wantContains: "export const task_add = tool({",
			wantErr:      false,
		},
		{
			name: "custom binary path",
			opts: &ToolWrapperOptions{
				BinaryPath: "myapp",
			},
			wantContains: "execFileSync(`myapp`",
			wantErr:      false,
		},
		{
			name:         "nil options uses defaults",
			opts:         nil,
			wantContains: "execFileSync(`taskflow`",
			wantErr:      false,
		},
		{
			name: "empty binary path uses default",
			opts: &ToolWrapperOptions{
				BinaryPath: "",
			},
			wantContains: "execFileSync(`taskflow`",
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

func TestGenerateToolWrapperSpawnCwdDirectory(t *testing.T) {
	opts := DefaultToolWrapperOptions()
	result, err := GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
	}

	wantTools := len(getToolCommands())

	// The wrapper is worktree-unaware: no path import, no worktree reads,
	// no dirname guard. taskflow resolves the anchor by walking up from its
	// process cwd.
	if strings.Contains(result, "import path") {
		t.Error(`GenerateToolWrapper() output must not contain: import path`)
	}
	if strings.Contains(result, "context.worktree") {
		t.Error("GenerateToolWrapper() output must not contain: context.worktree")
	}
	if strings.Contains(result, "path.dirname") {
		t.Error("GenerateToolWrapper() output must not contain: path.dirname")
	}

	// Each tool must spawn with the opencode session directory as cwd.
	if count := strings.Count(result, "cwd: context.directory"); count != wantTools {
		t.Errorf("GenerateToolWrapper() cwd: context.directory count = %d, want %d", count, wantTools)
	}

	// The options object must contain exactly one cwd option per tool.
	if count := strings.Count(result, "cwd:"); count != wantTools {
		t.Errorf("GenerateToolWrapper() cwd option count = %d, want %d", count, wantTools)
	}

	// Every tool's execute must call checkVersion before building cmdArgs.
	if count := strings.Count(result, `checkVersion("taskflow");`); count != wantTools {
		t.Errorf("GenerateToolWrapper() checkVersion call count = %d, want %d", count, wantTools)
	}
}

func TestGenerateToolWrapperVersionEmbed(t *testing.T) {
	t.Run("embedded version", func(t *testing.T) {
		result, err := GenerateToolWrapper(&ToolWrapperOptions{BinaryPath: "taskflow", Version: "0.1.0"})
		if err != nil {
			t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
		}
		if !strings.Contains(result, `const TASKFLOW_WRAPPER_VERSION = "0.1.0";`) {
			t.Error(`GenerateToolWrapper() output missing: const TASKFLOW_WRAPPER_VERSION = "0.1.0";`)
		}
	})

	t.Run("nil options default to 0.0.0", func(t *testing.T) {
		result, err := GenerateToolWrapper(nil)
		if err != nil {
			t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
		}
		if !strings.Contains(result, `const TASKFLOW_WRAPPER_VERSION = "0.0.0";`) {
			t.Error(`GenerateToolWrapper() output missing: const TASKFLOW_WRAPPER_VERSION = "0.0.0";`)
		}
	})

	t.Run("empty version defaults to 0.0.0", func(t *testing.T) {
		result, err := GenerateToolWrapper(&ToolWrapperOptions{BinaryPath: "taskflow"})
		if err != nil {
			t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
		}
		if !strings.Contains(result, `const TASKFLOW_WRAPPER_VERSION = "0.0.0";`) {
			t.Error(`GenerateToolWrapper() output missing: const TASKFLOW_WRAPPER_VERSION = "0.0.0";`)
		}
	})

	t.Run("helper block", func(t *testing.T) {
		result, err := GenerateToolWrapper(DefaultToolWrapperOptions())
		if err != nil {
			t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
		}
		for _, want := range []string{
			"function parseSemver(",
			"function compatible(",
			"function checkVersion(",
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
}

func TestGenerateToolWrapperCustomBinaryVersionCheck(t *testing.T) {
	opts := &ToolWrapperOptions{BinaryPath: "mybin", Version: "0.1.0"}
	result, err := GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
	}

	if !strings.Contains(result, `checkVersion("mybin")`) {
		t.Error(`GenerateToolWrapper() should call checkVersion("mybin")`)
	}
	if !strings.Contains(result, "execFileSync(`mybin`") {
		t.Error("GenerateToolWrapper() should use custom binary path in execFileSync")
	}
}

func TestGenerateToolWrapperContainsToolHelper(t *testing.T) {
	opts := DefaultToolWrapperOptions()
	result, err := GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
	}

	// Check for tool import
	if !strings.Contains(result, "import { tool } from \"@opencode-ai/plugin\"") {
		t.Error("GenerateToolWrapper() output missing: import { tool } from \"@opencode-ai/plugin\"")
	}

	// Check for tool({ pattern
	if !strings.Contains(result, "tool({") {
		t.Error("GenerateToolWrapper() output missing: tool({")
	}
}

func TestGenerateToolWrapperContainsDescription(t *testing.T) {
	opts := DefaultToolWrapperOptions()
	result, err := GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
	}

	// Check for description field
	if !strings.Contains(result, "description:") {
		t.Error("GenerateToolWrapper() output missing: description:")
	}

	// Check for specific command descriptions
	descriptions := []string{
		"Add a new task to the task list",
		"Mark a task as completed",
		"Block a task by providing its ID and a reason",
		"Soft delete a task",
		"Reset timed out tasks to todo status",
		"Start working on a task",
	}

	for _, desc := range descriptions {
		if !strings.Contains(result, desc) {
			t.Errorf("GenerateToolWrapper() output missing description: %s", desc)
		}
	}
}

func TestGenerateToolWrapperContainsArgsWithZodSchema(t *testing.T) {
	opts := DefaultToolWrapperOptions()
	result, err := GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
	}

	// Check for args field
	if !strings.Contains(result, "args:") {
		t.Error("GenerateToolWrapper() output missing: args:")
	}

	// Check for tool.schema patterns
	schemaPatterns := []string{
		"tool.schema.string()",
		"tool.schema.number()",
	}

	for _, pattern := range schemaPatterns {
		if !strings.Contains(result, pattern) {
			t.Errorf("GenerateToolWrapper() output missing: %s", pattern)
		}
	}

	// Check for .describe() calls (part of Zod schema)
	if !strings.Contains(result, ".describe(") {
		t.Error("GenerateToolWrapper() output missing: .describe(")
	}
}

func TestGenerateToolWrapperContainsExecuteFunction(t *testing.T) {
	opts := DefaultToolWrapperOptions()
	result, err := GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
	}

	// Check for execute function
	if !strings.Contains(result, "async execute(args, context)") {
		t.Error("GenerateToolWrapper() output missing: async execute(args, context)")
	}

	// Check that multiple tools have execute functions
	count := strings.Count(result, "async execute(args, context)")
	if count != 12 {
		t.Errorf("GenerateToolWrapper() should have exactly 12 execute functions, found %d", count)
	}
}

func TestGenerateToolWrapperProducesValidOutput(t *testing.T) {
	opts := &ToolWrapperOptions{
		BinaryPath: "taskflow",
	}

	result, err := GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
	}

	// Check that key elements are present
	checks := []struct {
		pattern string
		desc    string
	}{
		{"import { tool } from \"@opencode-ai/plugin\"", "tool import"},
		{"export const task_add = tool({", "add tool definition"},
		{"export const task_complete = tool({", "complete tool definition"},
		{"export const task_block = tool({", "block tool definition"},
		{"export const task_delete = tool({", "delete tool definition"},
		{"export const task_list_all = tool({", "list_all tool definition"},
		{"export const task_list_blocked = tool({", "list_blocked tool definition"},
		{"export const task_list_done = tool({", "list_done tool definition"},
		{"export const task_list_status_in_progress = tool({", "list_status_in_progress tool definition"},
		{"export const task_list_status_todo = tool({", "list_status_todo tool definition"},
		{"export const task_reset_timedout = tool({", "reset_timedout tool definition"},
		{"export const task_start = tool({", "start tool definition"},
		{"description:", "description field"},
		{"args: {", "args object start"},
		{"async execute(args, context)", "execute function"},
		{"const cmdArgs = [", "command args array"},
		{"execFileSync", "execFileSync call"},
	}

	for _, check := range checks {
		if !strings.Contains(result, check.pattern) {
			t.Errorf("GenerateToolWrapper() output missing: %s", check.desc)
		}
	}
}

func TestGenerateToolWrapperWithCustomBinaryPath(t *testing.T) {
	opts := &ToolWrapperOptions{
		BinaryPath: "custom-task-cli",
	}

	result, err := GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
	}

	// Verify custom binary path is used in execFileSync calls
	if !strings.Contains(result, "execFileSync(`custom-task-cli`") {
		t.Error("GenerateToolWrapper() should use custom binary path")
	}
}

func TestGenerateToolWrapperContainsAllCommands(t *testing.T) {
	opts := DefaultToolWrapperOptions()
	result, err := GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
	}

	// Individual command tools
	commandNames := []string{
		"task_add",
		"task_block",
		"task_complete",
		"task_delete",
		"task_list_all",
		"task_list_blocked",
		"task_list_done",
		"task_list_status_in_progress",
		"task_list_status_todo",
		"task_reset_timedout",
		"task_start",
		"task_unblock",
	}

	for _, cmd := range commandNames {
		pattern := "export const " + cmd + " = tool({"
		if !strings.Contains(result, pattern) {
			t.Errorf("GenerateToolWrapper() output missing: %s", cmd)
		}
	}
}

func TestGenerateToolWrapperCommandArgumentHandling(t *testing.T) {
	opts := DefaultToolWrapperOptions()
	result, err := GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
	}

	// Verify JSON format for arguments
	if !strings.Contains(result, "args.") {
		t.Error("GenerateToolWrapper() should reference args in JSON serialization")
	}

	// Verify JSON format for arguments
	if !strings.Contains(result, "JSON.stringify") {
		t.Error("GenerateToolWrapper() should use JSON format for arguments")
	}
}

func TestGenerateToolWrapperSchemaTypes(t *testing.T) {
	opts := DefaultToolWrapperOptions()
	result, err := GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
	}

	// Verify different Zod schema types are used
	schemaTypes := []struct {
		pattern string
		desc    string
	}{
		{"tool.schema.string()", "string schema"},
		{"tool.schema.number()", "number schema"},
	}

	for _, st := range schemaTypes {
		if !strings.Contains(result, st.pattern) {
			t.Errorf("GenerateToolWrapper() missing schema type: %s", st.desc)
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
	opts := DefaultToolWrapperOptions()
	result, err := GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
	}

	if len(result) == 0 {
		t.Error("GenerateToolWrapper() should return non-empty string")
	}
}

func TestGenerateToolWrapperReturnsResult(t *testing.T) {
	opts := DefaultToolWrapperOptions()
	result, err := GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
	}

	// Verify the result includes return statement
	if !strings.Contains(result, "return result;") {
		t.Error("GenerateToolWrapper() should return result from execSync")
	}
}

func TestGenerateToolWrapperUnblockCommand(t *testing.T) {
	opts := DefaultToolWrapperOptions()
	result, err := GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
	}

	// Verify task_unblock tool definition exists
	pattern := "export const task_unblock = tool({"
	if !strings.Contains(result, pattern) {
		t.Errorf("GenerateToolWrapper() output missing: %s", pattern)
	}

	// Verify description includes unblock-specific text
	if !strings.Contains(result, "Unblock a previously blocked task") {
		t.Error("GenerateToolWrapper() output missing unblock command description")
	}

	// Verify the CLI subcommand "unblock" is used
	if !strings.Contains(result, `cmdArgs.push("unblock")`) {
		t.Error("GenerateToolWrapper() should use 'unblock' as CLI subcommand")
	}

	// Verify id argument is present
	if !strings.Contains(result, "id: args.id") {
		t.Error("GenerateToolWrapper() should reference args.id in payload")
	}

	// Verify description argument is present
	if !strings.Contains(result, "description: args.description") {
		t.Error("GenerateToolWrapper() should reference args.description in payload")
	}
}

func TestGenerateToolWrapperHandlesDashesInCommandNames(t *testing.T) {
	opts := DefaultToolWrapperOptions()
	result, err := GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
	}

	// Verify that underscores are preserved (reset_timedout -> task_reset_timedout)
	if strings.Contains(result, "task_reset_timedout") {
		// This is the expected behavior - underscores preserved
	} else if strings.Contains(result, "task_resetTimedout") {
		t.Error("Command name with underscores should be preserved")
	}
}

func TestGenerateToolWrapperDoesNotAdvertiseAllKnob(t *testing.T) {
	// cmd/list.go no longer reads an "all" JSON field, so the generated
	// wrapper must not advertise one.
	opts := DefaultToolWrapperOptions()
	result, err := GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
	}

	if strings.Contains(result, "Show all tasks including completed") {
		t.Error("GenerateToolWrapper() output must not advertise the removed \"all\" list parameter")
	}
	if strings.Contains(result, "args.all") {
		t.Error("GenerateToolWrapper() output must not reference args.all in a payload")
	}
	if strings.Contains(result, "all: tool.schema.boolean()") {
		t.Error("GenerateToolWrapper() output must not define an \"all\" argument")
	}
}
