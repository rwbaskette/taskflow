package generator

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// ValidTypeScriptIdentifier checks if a string is a valid TypeScript identifier
// Must start with a letter, underscore, or dollar sign, followed by word characters
var ValidTypeScriptIdentifier = regexp.MustCompile(`^[a-zA-Z_$][\w$]*$`)

// ValidateIdentifier returns an error if the identifier is not valid
func ValidateIdentifier(identifier string) error {
	if identifier == "" {
		return fmt.Errorf("identifier cannot be empty")
	}
	if !ValidTypeScriptIdentifier.MatchString(identifier) {
		return fmt.Errorf("invalid TypeScript identifier: %s (must match ^[a-zA-Z_$][\\w$]*$)", identifier)
	}
	return nil
}

// CleanAndValidatePath sanitizes a file path and ensures it doesn't escape allowed directory
func CleanAndValidatePath(outputPath string, allowedDir string) (string, error) {
	// Convert allowed dir to absolute
	absAllowed, err := filepath.Abs(allowedDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve allowed directory: %w", err)
	}

	// If output path is absolute, clean and validate directly
	if filepath.IsAbs(outputPath) {
		cleaned := filepath.Clean(outputPath)
		absCleaned, err := filepath.Abs(cleaned)
		if err != nil {
			return "", fmt.Errorf("failed to resolve output path: %w", err)
		}
		// Check if path tries to escape the allowed directory
		if !strings.HasPrefix(absCleaned, absAllowed+string(filepath.Separator)) && absCleaned != absAllowed {
			return "", fmt.Errorf("path traversal detected: %s is not within %s", outputPath, allowedDir)
		}
		return cleaned, nil
	}

	// For relative paths, join with allowed dir, then clean and validate
	joined := filepath.Join(absAllowed, outputPath)
	cleaned := filepath.Clean(joined)

	// Verify the cleaned path is still within allowed dir
	if !strings.HasPrefix(cleaned, absAllowed+string(filepath.Separator)) && cleaned != absAllowed {
		return "", fmt.Errorf("path traversal detected: %s is not within %s", outputPath, allowedDir)
	}

	return cleaned, nil
}

// TypeScriptOptions contains options for TypeScript code generation
type TypeScriptOptions struct {
	PackageName string
	ClassName   string
}

// DefaultTypeScriptOptions returns default options for TypeScript generation
func DefaultTypeScriptOptions() *TypeScriptOptions {
	return &TypeScriptOptions{
		PackageName: "taskflow",
		ClassName:   "TaskManager",
	}
}

// GenerateTypeScriptWrapper generates a TypeScript wrapper for task management operations
// Returns an error if className or packageName contain invalid characters
func GenerateTypeScriptWrapper(opts *TypeScriptOptions) (string, error) {
	if opts == nil {
		opts = DefaultTypeScriptOptions()
	}

	// Validate identifiers to prevent XSS
	if err := ValidateIdentifier(opts.ClassName); err != nil {
		return "", fmt.Errorf("invalid className: %w", err)
	}
	if err := ValidateIdentifier(opts.PackageName); err != nil {
		return "", fmt.Errorf("invalid packageName: %w", err)
	}

	var b strings.Builder

	// File header
	b.WriteString("/**\n")
	b.WriteString(" * TaskFlow TypeScript Client\n")
	b.WriteString(" * Auto-generated wrapper for task management operations\n")
	b.WriteString(" * Package: " + opts.PackageName + "\n")
	b.WriteString(" */\n\n")

	// Interface definitions
	b.WriteString("// Task status types\nexport type TaskStatus = 'todo' | 'in_progress' | 'done' | 'blocked';\n\n")
	b.WriteString("// Task interface\nexport interface Task {\n")
	b.WriteString("  id: string;\n")
	b.WriteString("  title: string;\n")
	b.WriteString("  description: string;\n")
	b.WriteString("  milestone: string;\n")
	b.WriteString("  status: TaskStatus;\n")
	b.WriteString("  actor?: string;\n")
	b.WriteString("  lastUpdated: string;\n")
	b.WriteString("}\n\n")

	// Input interfaces
	b.WriteString("// Add task input\nexport interface AddTaskInput {\n")
	b.WriteString("  id: string;\n")
	b.WriteString("  title: string;\n")
	b.WriteString("  description: string;\n")
	b.WriteString("  milestone: string;\n")
	b.WriteString("  actor?: string;\n")
	b.WriteString("}\n\n")

	b.WriteString("// Update task input\nexport interface UpdateTaskInput {\n")
	b.WriteString("  id: string;\n")
	b.WriteString("  title?: string;\n")
	b.WriteString("  description?: string;\n")
	b.WriteString("  milestone?: string;\n")
	b.WriteString("  status?: TaskStatus;\n")
	b.WriteString("  actor?: string;\n")
	b.WriteString("}\n\n")

	b.WriteString("// Block task input\nexport interface BlockTaskInput {\n")
	b.WriteString("  id: string;\n")
	b.WriteString("  reason: string;\n")
	b.WriteString("}\n\n")

	b.WriteString("// List filter input\nexport interface ListFilterInput {\n")
	b.WriteString("  milestone?: string;\n")
	b.WriteString("  status?: TaskStatus;\n")
	b.WriteString("  actor?: string;\n")
	b.WriteString("  limit?: number;\n")
	b.WriteString("  offset?: number;\n")
	b.WriteString("  showAll?: boolean;\n")
	b.WriteString("}\n\n")

	b.WriteString("// Reset timed out input\nexport interface ResetTimedOutInput {\n")
	b.WriteString("  timeoutMinutes: number;\n")
	b.WriteString("}\n\n")

	// List result interface
	b.WriteString("// List result interface\nexport interface ListResult {\n")
	b.WriteString("  tasks: Task[];\n")
	b.WriteString("  total: number;\n")
	b.WriteString("  limit: number;\n")
	b.WriteString("  offset: number;\n")
	b.WriteString("  hasMore: boolean;\n")
	b.WriteString("}\n\n")

	// Result interfaces
	b.WriteString("// Add task result\nexport interface AddTaskResult extends Task {}\n\n")
	b.WriteString("// Update task result\nexport interface UpdateTaskResult extends Task {}\n\n")
	b.WriteString("// Complete task result\nexport interface CompleteTaskResult extends Task {}\n\n")
	b.WriteString("// Block task result\nexport interface BlockTaskResult extends Task {}\n\n")
	b.WriteString("// Reset task result\nexport interface ResetTaskResult extends Task {}\n\n")
	b.WriteString("// Reset timed out result\nexport interface ResetTimedOutResult {\n")
	b.WriteString("  resetTasks: ResetTaskResult[];\n")
	b.WriteString("}\n\n")

	// TaskManager class
	b.WriteString("/**\n")
	b.WriteString(" * TaskManager - Main class for task management operations\n")
	b.WriteString(" */\nexport class " + opts.ClassName + " {\n")
	b.WriteString("  private baseUrl: string;\n\n")
	b.WriteString("  /**\n")
	b.WriteString("   * Create a new TaskManager instance\n")
	b.WriteString("   * @param baseUrl - Base URL for the TaskFlow API (default: 'http://localhost:8080')\n")
	b.WriteString("   */\n")
	b.WriteString("  constructor(baseUrl: string = 'http://localhost:8080') {\n")
	b.WriteString("    this.baseUrl = baseUrl;\n")
	b.WriteString("  }\n\n")

	// Add task method
	b.WriteString("  /**\n")
	b.WriteString("   * Add a new task\n")
	b.WriteString("   * @param input - Task input data\n")
	b.WriteString("   * @returns Promise resolving to the created task\n")
	b.WriteString("   */\n")
	b.WriteString("  async addTask(input: AddTaskInput): Promise<AddTaskResult> {\n")
	b.WriteString("    const response = await fetch(`${this.baseUrl}/tasks/add`, {\n")
	b.WriteString("      method: 'POST',\n")
	b.WriteString("      headers: { 'Content-Type': 'application/json' },\n")
	b.WriteString("      body: JSON.stringify(input),\n")
	b.WriteString("    });\n")
	b.WriteString("    if (!response.ok) {\n")
	b.WriteString("      throw new Error(`Failed to add task: ${response.statusText}`);\n")
	b.WriteString("    }\n")
	b.WriteString("    return response.json();\n")
	b.WriteString("  }\n\n")

	// Update task method
	b.WriteString("  /**\n")
	b.WriteString("   * Update an existing task\n")
	b.WriteString("   * @param input - Task input data with id\n")
	b.WriteString("   * @returns Promise resolving to the updated task\n")
	b.WriteString("   */\n")
	b.WriteString("  async updateTask(input: UpdateTaskInput): Promise<UpdateTaskResult> {\n")
	b.WriteString("    const response = await fetch(`${this.baseUrl}/tasks/update`, {\n")
	b.WriteString("      method: 'POST',\n")
	b.WriteString("      headers: { 'Content-Type': 'application/json' },\n")
	b.WriteString("      body: JSON.stringify(input),\n")
	b.WriteString("    });\n")
	b.WriteString("    if (!response.ok) {\n")
	b.WriteString("      throw new Error(`Failed to update task: ${response.statusText}`);\n")
	b.WriteString("    }\n")
	b.WriteString("    return response.json();\n")
	b.WriteString("  }\n\n")

	// Complete task method
	b.WriteString("  /**\n")
	b.WriteString("   * Mark a task as completed\n")
	b.WriteString("   * @param taskId - ID of the task to complete\n")
	b.WriteString("   * @returns Promise resolving to the completed task\n")
	b.WriteString("   */\n")
	b.WriteString("  async completeTask(taskId: string): Promise<CompleteTaskResult> {\n")
	b.WriteString("    const response = await fetch(`${this.baseUrl}/tasks/complete`, {\n")
	b.WriteString("      method: 'POST',\n")
	b.WriteString("      headers: { 'Content-Type': 'application/json' },\n")
	b.WriteString("      body: JSON.stringify({ id: taskId }),\n")
	b.WriteString("    });\n")
	b.WriteString("    if (!response.ok) {\n")
	b.WriteString("      throw new Error(`Failed to complete task: ${response.statusText}`);\n")
	b.WriteString("    }\n")
	b.WriteString("    return response.json();\n")
	b.WriteString("  }\n\n")

	// Block task method
	b.WriteString("  /**\n")
	b.WriteString("   * Block a task with a reason\n")
	b.WriteString("   * @param input - Block task input with id and reason\n")
	b.WriteString("   * @returns Promise resolving to the blocked task\n")
	b.WriteString("   */\n")
	b.WriteString("  async blockTask(input: BlockTaskInput): Promise<BlockTaskResult> {\n")
	b.WriteString("    const response = await fetch(`${this.baseUrl}/tasks/block`, {\n")
	b.WriteString("      method: 'POST',\n")
	b.WriteString("      headers: { 'Content-Type': 'application/json' },\n")
	b.WriteString("      body: JSON.stringify(input),\n")
	b.WriteString("    });\n")
	b.WriteString("    if (!response.ok) {\n")
	b.WriteString("      throw new Error(`Failed to block task: ${response.statusText}`);\n")
	b.WriteString("    }\n")
	b.WriteString("    return response.json();\n")
	b.WriteString("  }\n\n")

	// List tasks method
	b.WriteString("  /**\n")
	b.WriteString("   * List tasks with optional filters\n")
	b.WriteString("   * @param filter - Optional filter parameters\n")
	b.WriteString("   * @returns Promise resolving to list result\n")
	b.WriteString("   */\n")
	b.WriteString("  async listTasks(filter?: ListFilterInput): Promise<ListResult> {\n")
	b.WriteString("    const params = new URLSearchParams();\n")
	b.WriteString("    if (filter) {\n")
	b.WriteString("      if (filter.milestone) params.append('milestone', filter.milestone);\n")
	b.WriteString("      if (filter.status) params.append('status', filter.status);\n")
	b.WriteString("      if (filter.actor) params.append('actor', filter.actor);\n")
	b.WriteString("      if (filter.limit) params.append('limit', filter.limit.toString());\n")
	b.WriteString("      if (filter.offset) params.append('offset', filter.offset.toString());\n")
	b.WriteString("      if (filter.showAll) params.append('showAll', 'true');\n")
	b.WriteString("    }\n")
	b.WriteString("    const response = await fetch(`${this.baseUrl}/tasks/list?${params}`);\n")
	b.WriteString("    if (!response.ok) {\n")
	b.WriteString("      throw new Error(`Failed to list tasks: ${response.statusText}`);\n")
	b.WriteString("    }\n")
	b.WriteString("    return response.json();\n")
	b.WriteString("  }\n\n")

	// Reset timed out method
	b.WriteString("  /**\n")
	b.WriteString("   * Reset tasks that have exceeded the timeout\n")
	b.WriteString("   * @param input - Reset timed out input with timeout in minutes\n")
	b.WriteString("   * @returns Promise resolving to reset result\n")
	b.WriteString("   */\n")
	b.WriteString("  async resetTimedOut(input: ResetTimedOutInput): Promise<ResetTimedOutResult> {\n")
	b.WriteString("    const response = await fetch(`${this.baseUrl}/tasks/reset`, {\n")
	b.WriteString("      method: 'POST',\n")
	b.WriteString("      headers: { 'Content-Type': 'application/json' },\n")
	b.WriteString("      body: JSON.stringify(input),\n")
	b.WriteString("    });\n")
	b.WriteString("    if (!response.ok) {\n")
	b.WriteString("      throw new Error(`Failed to reset timed out tasks: ${response.statusText}`);\n")
	b.WriteString("    }\n")
	b.WriteString("    return response.json();\n")
	b.WriteString("  }\n")
	b.WriteString("}\n\n")

	// Export default instance helper
	b.WriteString("// Convenience function to create a TaskManager instance\nexport function createTaskManager(baseUrl?: string): " + opts.ClassName + " {\n")
	b.WriteString("  return new " + opts.ClassName + "(baseUrl);\n")
	b.WriteString("}\n\n")

	// Export module
	b.WriteString("export default " + opts.ClassName + ";\n")

	return b.String(), nil
}

// OpenCodeOptions contains options for OpenCode wrapper generation
type OpenCodeOptions struct {
	BinaryName string
	WorkingDir string
}

// DefaultOpenCodeOptions returns default options for OpenCode generation
func DefaultOpenCodeOptions() *OpenCodeOptions {
	return &OpenCodeOptions{
		BinaryName: "task",
		WorkingDir: ".",
	}
}

// GenerateOpenCodeWrapper generates a shell script wrapper for OpenCode integration
// This allows OpenCode to call the manage-tasks binary properly
func GenerateOpenCodeWrapper(opts *OpenCodeOptions) (string, error) {
	if opts == nil {
		opts = DefaultOpenCodeOptions()
	}

	var b strings.Builder

	// Shebang
	b.WriteString("#!/bin/bash\n\n")
	b.WriteString("# OpenCode wrapper for TaskFlow CLI\n")
	b.WriteString("# Generated by manage-tasks --agent opencode\n\n")

	// Get the directory where this script is located
	b.WriteString("# Get the directory where this script is located\n")
	b.WriteString("SCRIPT_DIR=\"$(cd \"$(dirname \"${BASH_SOURCE[0]}\")\"\" && pwd)\"\n\n")

	// Determine the binary path
	b.WriteString("# Determine the binary path\n")
	b.WriteString("if [ -z \"$MANAGE_TASKS_BIN\" ]; then\n")
	b.WriteString("  MANAGE_TASKS_BIN=\"${SCRIPT_DIR}/" + opts.BinaryName + "\"\n")
	b.WriteString("fi\n\n")

	// Helper function to call the binary
	b.WriteString("# Helper function to execute manage-tasks commands\n")
	b.WriteString("manage_tasks() {\n")
	b.WriteString("  \"$MANAGE_TASKS_BIN\" \"$@\"\n")
	b.WriteString("}\n\n")

	// Parse command-line arguments and route to appropriate manage-tasks commands
	b.WriteString("# Parse command and forward to manage-tasks\n")
	b.WriteString("COMMAND=\"$1\"\n")
	b.WriteString("shift || true\n\n")

	b.WriteString("case \"$COMMAND\" in\n")
	b.WriteString("  add)\n")
	b.WriteString("    manage_tasks add \"$@\"\n")
	b.WriteString("    ;;\n")
	b.WriteString("  update)\n")
	b.WriteString("    manage_tasks update \"$@\"\n")
	b.WriteString("    ;;\n")
	b.WriteString("  complete)\n")
	b.WriteString("    manage_tasks complete \"$@\"\n")
	b.WriteString("    ;;\n")
	b.WriteString("  block)\n")
	b.WriteString("    manage_tasks block \"$@\"\n")
	b.WriteString("    ;;\n")
	b.WriteString("  list)\n")
	b.WriteString("    manage_tasks list \"$@\"\n")
	b.WriteString("    ;;\n")
	b.WriteString("  reset-timedout|reset)\n")
	b.WriteString("    manage_tasks reset-timedout \"$@\"\n")
	b.WriteString("    ;;\n")
	b.WriteString("  help|--help|-h)\n")
	b.WriteString("    manage_tasks --help\n")
	b.WriteString("    ;;\n")
	b.WriteString("  *)\n")
	b.WriteString("    manage_tasks \"$@\"\n")
	b.WriteString("    ;;\n")
	b.WriteString("esac\n")

	return b.String(), nil
}

// GetSupportedAgents returns the list of supported agent types
func GetSupportedAgents() []string {
	return []string{
		"typescript",
		"opencode",
	}
}

// OpenCodeAgent generates code for the opencode agent type
type OpenCodeAgent struct {
	BinaryName string
	WorkingDir string
}

// GenerateAgent generates code for the specified agent type
// binaryName is only used for opencode agent type, defaults to "task" if empty
func GenerateAgent(agentType string, opts *TypeScriptOptions, binaryName string) (string, error) {
	// Set default binary name for opencode agent
	if binaryName == "" {
		binaryName = "task"
	}

	switch agentType {
	case "typescript":
		return GenerateTypeScriptWrapper(opts)
	case "opencode":
		opencodeOpts := &OpenCodeOptions{
			BinaryName: binaryName,
			WorkingDir: ".",
		}
		return GenerateOpenCodeWrapper(opencodeOpts)
	default:
		return "", fmt.Errorf("unsupported agent type: %s", agentType)
	}
}
