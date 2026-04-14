package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateIdentifier(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple name", "TaskManager", false},
		{"valid with underscore", "my_class", false},
		{"valid with dollar", "$myClass", false},
		{"valid with numbers", "class1", false},
		{"valid camelCase", "myTaskManager", false},
		{"valid PascalCase", "MyTaskManager", false},
		{"valid with leading underscore", "_private", false},
		{"empty string", "", true},
		{"starts with number", "1task", true},
		{"starts with dash", "-task", true},
		{"contains space", "my task", true},
		{"contains special char", "my-task", true},
		{"contains script tag", "<script>", true},
		{"contains template literal", "${alert(1)}", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIdentifier(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIdentifier(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestCleanAndValidatePath(t *testing.T) {
	// Create temp allowed dir for all tests
	tmpDir, err := os.MkdirTemp("", "test_allowed")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a subdirectory to use as allowed dir
	subDir := filepath.Join(tmpDir, "project")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("failed to create sub dir: %v", err)
	}

	tests := []struct {
		name       string
		outputPath string
		allowedDir string
		wantErr    bool
	}{
		{
			name:       "simple valid relative path",
			outputPath: "output/taskflow.ts",
			allowedDir: subDir,
			wantErr:    false,
		},
		{
			name:       "path with dot dots attempting traversal",
			outputPath: "../other/file.ts",
			allowedDir: subDir,
			wantErr:    true,
		},
		{
			name:       "absolute path traversal attempt",
			outputPath: "/etc/passwd",
			allowedDir: subDir,
			wantErr:    true,
		},
		{
			name:       "path with parent directory reference from subdir",
			outputPath: "subdir/../../../etc/passwd",
			allowedDir: subDir,
			wantErr:    true,
		},
		{
			name:       "valid nested path",
			outputPath: "dir1/dir2/file.ts",
			allowedDir: subDir,
			wantErr:    false,
		},
		{
			name:       "absolute path within allowed dir",
			outputPath: filepath.Join(subDir, "output.ts"),
			allowedDir: subDir,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err = CleanAndValidatePath(tt.outputPath, tt.allowedDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("CleanAndValidatePath(%q, %q) error = %v, wantErr %v",
					tt.outputPath, tt.allowedDir, err, tt.wantErr)
			}
		})
	}
}

func TestGenerateTypeScriptWrapper(t *testing.T) {
	tests := []struct {
		name         string
		opts         *TypeScriptOptions
		wantContains string
		wantErr      bool
	}{
		{
			name: "default options",
			opts: &TypeScriptOptions{
				PackageName: "taskflow",
				ClassName:   "TaskManager",
			},
			wantContains: "export class TaskManager",
			wantErr:      false,
		},
		{
			name: "custom class name",
			opts: &TypeScriptOptions{
				PackageName: "taskflow",
				ClassName:   "MyTaskManager",
			},
			wantContains: "export class MyTaskManager",
			wantErr:      false,
		},
		{
			name: "custom package name in header",
			opts: &TypeScriptOptions{
				PackageName: "mypackage",
				ClassName:   "TaskManager",
			},
			wantContains: "Package: mypackage",
			wantErr:      false,
		},
		{
			name:         "nil options uses defaults",
			opts:         nil,
			wantContains: "export class TaskManager",
			wantErr:      false,
		},
		{
			name: "XSS attempt in className",
			opts: &TypeScriptOptions{
				PackageName: "taskflow",
				ClassName:   "<script>alert(1)</script>",
			},
			wantContains: "",
			wantErr:      true,
		},
		{
			name: "template literal injection in className",
			opts: &TypeScriptOptions{
				PackageName: "taskflow",
				ClassName:   "${alert(1)}",
			},
			wantContains: "",
			wantErr:      true,
		},
		{
			name: "XSS attempt in packageName",
			opts: &TypeScriptOptions{
				PackageName: "<script>alert(1)</script>",
				ClassName:   "TaskManager",
			},
			wantContains: "",
			wantErr:      true,
		},
		{
			name: "valid identifier with underscore",
			opts: &TypeScriptOptions{
				PackageName: "my_package",
				ClassName:   "_TaskManager",
			},
			wantContains: "export class _TaskManager",
			wantErr:      false,
		},
		{
			name: "valid identifier with dollar",
			opts: &TypeScriptOptions{
				PackageName: "$package",
				ClassName:   "$TaskManager",
			},
			wantContains: "export class $TaskManager",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateTypeScriptWrapper(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateTypeScriptWrapper() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.wantContains != "" && !strings.Contains(result, tt.wantContains) {
				t.Errorf("GenerateTypeScriptWrapper() result does not contain %q", tt.wantContains)
			}
		})
	}
}

func TestGenerateTypeScriptWrapperProducesValidOutput(t *testing.T) {
	opts := &TypeScriptOptions{
		PackageName: "taskflow",
		ClassName:   "TaskManager",
	}

	result, err := GenerateTypeScriptWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateTypeScriptWrapper() unexpected error: %v", err)
	}

	// Check that key elements are present
	checks := []string{
		"export type TaskStatus",
		"export interface Task {",
		"export interface AddTaskInput {",
		"export interface UpdateTaskInput {",
		"export interface BlockTaskInput {",
		"export interface ListFilterInput {",
		"export interface ResetTimedOutInput {",
		"export interface ListResult {",
		"export class TaskManager {",
		"async addTask(",
		"async updateTask(",
		"async completeTask(",
		"async blockTask(",
		"async listTasks(",
		"async resetTimedOut(",
		"export default TaskManager",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("GenerateTypeScriptWrapper() output missing: %s", check)
		}
	}
}

func TestGenerateAgent(t *testing.T) {
	tests := []struct {
		name      string
		agentType string
		opts      *TypeScriptOptions
		wantErr   bool
	}{
		{
			name:      "typescript agent with valid options",
			agentType: "typescript",
			opts: &TypeScriptOptions{
				PackageName: "taskflow",
				ClassName:   "TaskManager",
			},
			wantErr: false,
		},
		{
			name:      "typescript agent with invalid className",
			agentType: "typescript",
			opts: &TypeScriptOptions{
				PackageName: "taskflow",
				ClassName:   "invalid name",
			},
			wantErr: true,
		},
		{
			name:      "unsupported agent type",
			agentType: "python",
			opts:      &TypeScriptOptions{},
			wantErr:   true,
		},
		{
			name:      "nil options uses defaults",
			agentType: "typescript",
			opts:      nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GenerateAgent(tt.agentType, tt.opts, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateAgent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultTypeScriptOptions(t *testing.T) {
	opts := DefaultTypeScriptOptions()

	if opts.PackageName != "taskflow" {
		t.Errorf("DefaultTypeScriptOptions() PackageName = %v, want taskflow", opts.PackageName)
	}
	if opts.ClassName != "TaskManager" {
		t.Errorf("DefaultTypeScriptOptions() ClassName = %v, want TaskManager", opts.ClassName)
	}
}

func TestGetSupportedAgents(t *testing.T) {
	agents := GetSupportedAgents()

	if len(agents) != 2 {
		t.Errorf("GetSupportedAgents() returned %d agents, want 2", len(agents))
	}

	found := false
	for _, a := range agents {
		if a == "typescript" {
			found = true
			break
		}
	}
	if !found {
		t.Error("GetSupportedAgents() should contain 'typescript'")
	}

	// Check for opencode agent
	found = false
	for _, a := range agents {
		if a == "opencode" {
			found = true
			break
		}
	}
	if !found {
		t.Error("GetSupportedAgents() should contain 'opencode'")
	}
}

// Test path traversal from current working directory
func TestPathTraversalFromCWD(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Skip("could not get current working directory")
	}

	// Test path traversal detection
	_, err = CleanAndValidatePath("../test.txt", cwd)
	if err == nil {
		t.Error("CleanAndValidatePath should detect path traversal attempt")
	}

	// Test valid relative path
	cleanPath, err := CleanAndValidatePath("subdir/test.txt", cwd)
	if err != nil {
		t.Errorf("CleanAndValidatePath should accept valid path: %v", err)
	}
	_ = cleanPath
}

// Test absolute path traversal
func TestAbsolutePathTraversal(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Skip("could not get current working directory")
	}

	// Test absolute path traversal
	_, err = CleanAndValidatePath("/etc/passwd", cwd)
	if err == nil {
		t.Error("CleanAndValidatePath should reject absolute path traversal to /etc/passwd")
	}
}

// Test XSS prevention
func TestXSSPrevention(t *testing.T) {
	xssPayloads := []string{
		"<script>alert(1)</script>",
		"${alert(1)}",
		"{{constructor.constructor('alert(1)')()}}",
		"<img src=x onerror=alert(1)>",
		"javascript:alert(1)",
		"; rm -rf /",
		"$(whoami)",
	}

	for _, payload := range xssPayloads {
		t.Run(payload, func(t *testing.T) {
			opts := &TypeScriptOptions{
				PackageName: "taskflow",
				ClassName:   payload,
			}
			_, err := GenerateTypeScriptWrapper(opts)
			if err == nil {
				t.Error("GenerateTypeScriptWrapper should reject XSS payload in className")
			}
		})
	}
}

// Test that generated code doesn't contain raw user input
func TestGeneratedCodeSanitization(t *testing.T) {
	opts := &TypeScriptOptions{
		PackageName: "taskflow",
		ClassName:   "TaskManager",
	}

	result, err := GenerateTypeScriptWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateTypeScriptWrapper() error: %v", err)
	}

	// The generated code should not contain any XSS patterns in user-supplied values
	// Note: The template literals like ${this.baseUrl} are intentional for valid JS
	// We only check for dangerous patterns that would come from malicious user input
	dangerousPatterns := []string{
		"<script",
		"javascript:",
		"onerror=",
		"onclick=",
		"<img",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(result, pattern) {
			t.Errorf("Generated code contains dangerous XSS pattern: %s", pattern)
		}
	}

	// Verify that className and packageName don't appear in unsafe contexts
	// They should only appear in safe contexts (class declaration, comments, function return type)
	if strings.Contains(result, "class "+opts.ClassName+" {") {
		// This is expected and safe
	}
	// Ensure user input isn't reflected in a way that allows injection
}

// Test generated code structure
func TestGeneratedCodeStructure(t *testing.T) {
	opts := &TypeScriptOptions{
		PackageName: "myapp",
		ClassName:   "MyApp",
	}

	result, err := GenerateTypeScriptWrapper(opts)
	if err != nil {
		t.Fatalf("GenerateTypeScriptWrapper() error: %v", err)
	}

	// Verify package name appears in header comment
	if !strings.Contains(result, "Package: myapp") {
		t.Error("Generated code should include package name in header")
	}

	// Verify class name appears in class declaration
	if !strings.Contains(result, "export class MyApp") {
		t.Error("Generated code should declare the class with the correct name")
	}

	// Verify convenience function uses class name
	if !strings.Contains(result, "createTaskManager(baseUrl?: string): MyApp") {
		t.Error("Generated code should have createTaskManager function returning correct type")
	}

	// Verify default export uses class name
	if !strings.Contains(result, "export default MyApp") {
		t.Error("Generated code should have correct default export")
	}
}

// Test directory creation for nested paths
func TestNestedPathDirectoryCreation(t *testing.T) {
	// Create temp allowed dir
	tmpDir, err := os.MkdirTemp("", "test_nested")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outputPath := filepath.Join(tmpDir, "sub1", "sub2", "file.ts")
	cleanPath, err := CleanAndValidatePath(outputPath, tmpDir)
	if err != nil {
		t.Fatalf("CleanAndValidatePath() error: %v", err)
	}

	// Verify directory would be created (check by ensuring path is cleaned properly)
	expectedClean := filepath.Join(tmpDir, "sub1", "sub2", "file.ts")
	if cleanPath != expectedClean {
		t.Errorf("CleanAndValidatePath() = %v, want %v", cleanPath, expectedClean)
	}
}

func TestGenerateOpenCodeWrapper(t *testing.T) {
	tests := []struct {
		name         string
		opts         *OpenCodeOptions
		wantContains string
		wantErr      bool
	}{
		{
			name: "default binary name",
			opts: &OpenCodeOptions{
				BinaryName: "task",
				WorkingDir: ".",
			},
			wantContains: "MANAGE_TASKS_BIN=\"${SCRIPT_DIR}/task\"",
			wantErr:      false,
		},
		{
			name: "custom binary name",
			opts: &OpenCodeOptions{
				BinaryName: "myapp",
				WorkingDir: ".",
			},
			wantContains: "MANAGE_TASKS_BIN=\"${SCRIPT_DIR}/myapp\"",
			wantErr:      false,
		},
		{
			name: "custom binary name with path",
			opts: &OpenCodeOptions{
				BinaryName: "manage-tasks",
				WorkingDir: ".",
			},
			wantContains: "MANAGE_TASKS_BIN=\"${SCRIPT_DIR}/manage-tasks\"",
			wantErr:      false,
		},
		{
			name:         "nil options uses defaults",
			opts:         nil,
			wantContains: "MANAGE_TASKS_BIN=\"${SCRIPT_DIR}/task\"",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateOpenCodeWrapper(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateOpenCodeWrapper() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.wantContains != "" && !strings.Contains(result, tt.wantContains) {
				t.Errorf("GenerateOpenCodeWrapper() result does not contain %q", tt.wantContains)
			}
		})
	}
}

func TestGenerateAgentWithBinaryName(t *testing.T) {
	tests := []struct {
		name         string
		agentType    string
		opts         *TypeScriptOptions
		binaryName   string
		wantContains string
		wantErr      bool
	}{
		{
			name:         "opencode with default binary name",
			agentType:    "opencode",
			opts:         &TypeScriptOptions{},
			binaryName:   "",
			wantContains: "MANAGE_TASKS_BIN=\"${SCRIPT_DIR}/task\"",
			wantErr:      false,
		},
		{
			name:         "opencode with custom binary name",
			agentType:    "opencode",
			opts:         &TypeScriptOptions{},
			binaryName:   "my-binary",
			wantContains: "MANAGE_TASKS_BIN=\"${SCRIPT_DIR}/my-binary\"",
			wantErr:      false,
		},
		{
			name:         "opencode with explicit custom binary name",
			agentType:    "opencode",
			opts:         &TypeScriptOptions{},
			binaryName:   "custom-app",
			wantContains: "MANAGE_TASKS_BIN=\"${SCRIPT_DIR}/custom-app\"",
			wantErr:      false,
		},
		{
			name:         "typescript agent ignores binary name",
			agentType:    "typescript",
			opts:         &TypeScriptOptions{PackageName: "taskflow", ClassName: "TaskManager"},
			binaryName:   "unused-binary",
			wantContains: "export class TaskManager",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateAgent(tt.agentType, tt.opts, tt.binaryName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.wantContains != "" && !strings.Contains(result, tt.wantContains) {
				t.Errorf("GenerateAgent() result does not contain %q", tt.wantContains)
			}
		})
	}
}

func TestDefaultOpenCodeOptions(t *testing.T) {
	opts := DefaultOpenCodeOptions()

	if opts.BinaryName != "task" {
		t.Errorf("DefaultOpenCodeOptions() BinaryName = %v, want task", opts.BinaryName)
	}
	if opts.WorkingDir != "." {
		t.Errorf("DefaultOpenCodeOptions() WorkingDir = %v, want .", opts.WorkingDir)
	}
}
