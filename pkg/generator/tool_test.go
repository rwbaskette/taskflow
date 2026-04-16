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
			wantContains: "export const addTool = tool({",
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
		"Update an existing task by its ID",
		"Mark a task as completed",
		"Block a task by providing its ID and a reason",
		"List all tasks with optional filters",
		"Reset timed out tasks to todo status",
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
		"tool.schema.boolean()",
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
	if count < 7 {
		t.Errorf("GenerateToolWrapper() should have at least 7 execute functions, found %d", count)
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
		{"export const addTool = tool({", "add tool definition"},
		{"export const updateTool = tool({", "update tool definition"},
		{"export const completeTool = tool({", "complete tool definition"},
		{"export const blockTool = tool({", "block tool definition"},
		{"export const listTool = tool({", "list tool definition"},
		{"export const reset_timedoutTool = tool({", "reset_timedout tool definition"},
		{"export const taskflowTool = tool({", "combined taskflow tool definition"},
		{"description:", "description field"},
		{"args: {", "args object start"},
		{"async execute(args, context)", "execute function"},
		{"const cmdArgs = [", "command args array"},
		{"execFileSync", "execFileSync call"},
		{"export default taskflowTool", "default export"},
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

	// All 6 individual tools plus the combined tool
	commandNames := []string{
		"addTool",
		"updateTool",
		"completeTool",
		"blockTool",
		"listTool",
		"reset_timedoutTool",
		"taskflowTool",
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

	// Verify required argument handling
	if !strings.Contains(result, "if (args.") {
		t.Error("GenerateToolWrapper() should handle arguments")
	}

	// Verify --flag format for arguments
	if !strings.Contains(result, "--") {
		t.Error("GenerateToolWrapper() should use --flag format for arguments")
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
		{"tool.schema.boolean()", "boolean schema"},
		{"tool.schema.enum(", "enum schema"},
		{".optional()", "optional modifier"},
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

func TestGenerateToolWrapperContainsChildProcessImport(t *testing.T) {
	opts := DefaultToolWrapperOptions()
	result, err := GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
	}

	// Verify child_process is imported dynamically
	if !strings.Contains(result, "await import(\"child_process\")") {
		t.Error("GenerateToolWrapper() should dynamically import child_process")
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

func TestGenerateToolWrapperHandlesDashesInCommandNames(t *testing.T) {
	opts := DefaultToolWrapperOptions()
	result, err := GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
	}

	// Verify that dashes are handled (reset-timedout -> resetTimedout)
	if strings.Contains(result, "resetTimedoutTool") {
		// This is the expected behavior - dashes removed for valid JS identifier
	} else if strings.Contains(result, "reset-timedoutTool") {
		t.Error("Command name with dashes should be converted to valid JS identifier")
	}
}

func TestGenerateToolWrapperCombinedTool(t *testing.T) {
	opts := DefaultToolWrapperOptions()
	result, err := GenerateToolWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateToolWrapper() unexpected error: %v", err)
	}

	// Verify combined tool has all command options
	combinedChecks := []string{
		"taskflowTool = tool({",
		"action: tool.schema.enum(",
		"args.id",
		"args.title",
		"args.description",
		"args.milestone",
		"args.status",
		"args.actor",
		"args.reason",
		"args.minutes",
		"args.limit",
		"args.offset",
		"args.all",
	}

	for _, check := range combinedChecks {
		if !strings.Contains(result, check) {
			t.Errorf("GenerateToolWrapper() combined tool missing: %s", check)
		}
	}
}
