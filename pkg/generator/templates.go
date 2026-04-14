package generator

import (
	"fmt"
	"strings"
	"text/template"
)

// TypeScript templates for code generation

// headerTemplate generates the file header comment
const headerTemplate = "/**\n" +
	" * TaskFlow TypeScript Client\n" +
	" * Auto-generated wrapper for task management operations\n" +
	" * Package: {{.PackageName}}\n" +
	" */\n"

// typesTemplate generates TypeScript type definitions
const typesTemplate = `// Task status types
export type TaskStatus = 'todo' | 'in_progress' | 'done' | 'blocked';

// Task interface
export interface Task {
  id: string;
  title: string;
  description: string;
  milestone: string;
  status: TaskStatus;
  actor?: string;
  lastUpdated: string;
}

// Add task input
export interface AddTaskInput {
  id: string;
  title: string;
  description: string;
  milestone: string;
  actor?: string;
}

// Update task input
export interface UpdateTaskInput {
  id: string;
  title?: string;
  description?: string;
  milestone?: string;
  status?: TaskStatus;
  actor?: string;
}

// Block task input
export interface BlockTaskInput {
  id: string;
  reason: string;
}

// List filter input
export interface ListFilterInput {
  milestone?: string;
  status?: TaskStatus;
  actor?: string;
  limit?: number;
  offset?: number;
  showAll?: boolean;
}

// Reset timed out input
export interface ResetTimedOutInput {
  timeoutMinutes: number;
}

// List result interface
export interface ListResult {
  tasks: Task[];
  total: number;
  limit: number;
  offset: number;
  hasMore: boolean;
}

// Add task result
export interface AddTaskResult extends Task {}

// Update task result
export interface UpdateTaskResult extends Task {}

// Complete task result
export interface CompleteTaskResult extends Task {}

// Block task result
export interface BlockTaskResult extends Task {}

// Reset task result
export interface ResetTaskResult extends Task {}

// Reset timed out result
export interface ResetTimedOutResult {
  resetTasks: ResetTaskResult[];
}
`

// classTemplate generates the TaskManager class
const classTemplate = "/**\n" +
	" * TaskManager - Main class for task management operations\n" +
	" */\n" +
	"export class {{.ClassName}} {\n" +
	"  private baseUrl: string;\n\n" +
	"  /**\n" +
	"   * Create a new TaskManager instance\n" +
	"   * @param baseUrl - Base URL for the TaskFlow API (default: 'http://localhost:8080')\n" +
	"   */\n" +
	"  constructor(baseUrl: string = 'http://localhost:8080') {\n" +
	"    this.baseUrl = baseUrl;\n" +
	"  }\n\n" +
	"  /**\n" +
	"   * Add a new task\n" +
	"   * @param input - Task input data\n" +
	"   * @returns Promise resolving to the created task\n" +
	"   */\n" +
	"  async addTask(input: AddTaskInput): Promise<AddTaskResult> {\n" +
	"    const response = await fetch(`${this.baseUrl}/tasks/add`, {\n" +
	"      method: 'POST',\n" +
	"      headers: { 'Content-Type': 'application/json' },\n" +
	"      body: JSON.stringify(input),\n" +
	"    });\n" +
	"    if (!response.ok) {\n" +
	"      throw new Error(`Failed to add task: ${response.statusText}`);\n" +
	"    }\n" +
	"    return response.json();\n" +
	"  }\n\n" +
	"  /**\n" +
	"   * Update an existing task\n" +
	"   * @param input - Task input data with id\n" +
	"   * @returns Promise resolving to the updated task\n" +
	"   */\n" +
	"  async updateTask(input: UpdateTaskInput): Promise<UpdateTaskResult> {\n" +
	"    const response = await fetch(`${this.baseUrl}/tasks/update`, {\n" +
	"      method: 'POST',\n" +
	"      headers: { 'Content-Type': 'application/json' },\n" +
	"      body: JSON.stringify(input),\n" +
	"    });\n" +
	"    if (!response.ok) {\n" +
	"      throw new Error(`Failed to update task: ${response.statusText}`);\n" +
	"    }\n" +
	"    return response.json();\n" +
	"  }\n\n" +
	"  /**\n" +
	"   * Mark a task as completed\n" +
	"   * @param taskId - ID of the task to complete\n" +
	"   * @returns Promise resolving to the completed task\n" +
	"   */\n" +
	"  async completeTask(taskId: string): Promise<CompleteTaskResult> {\n" +
	"    const response = await fetch(`${this.baseUrl}/tasks/complete`, {\n" +
	"      method: 'POST',\n" +
	"      headers: { 'Content-Type': 'application/json' },\n" +
	"      body: JSON.stringify({ id: taskId }),\n" +
	"    });\n" +
	"    if (!response.ok) {\n" +
	"      throw new Error(`Failed to complete task: ${response.statusText}`);\n" +
	"    }\n" +
	"    return response.json();\n" +
	"  }\n\n" +
	"  /**\n" +
	"   * Block a task with a reason\n" +
	"   * @param input - Block task input with id and reason\n" +
	"   * @returns Promise resolving to the blocked task\n" +
	"   */\n" +
	"  async blockTask(input: BlockTaskInput): Promise<BlockTaskResult> {\n" +
	"    const response = await fetch(`${this.baseUrl}/tasks/block`, {\n" +
	"      method: 'POST',\n" +
	"      headers: { 'Content-Type': 'application/json' },\n" +
	"      body: JSON.stringify(input),\n" +
	"    });\n" +
	"    if (!response.ok) {\n" +
	"      throw new Error(`Failed to block task: ${response.statusText}`);\n" +
	"    }\n" +
	"    return response.json();\n" +
	"  }\n\n" +
	"  /**\n" +
	"   * List tasks with optional filters\n" +
	"   * @param filter - Optional filter parameters\n" +
	"   * @returns Promise resolving to list result\n" +
	"   */\n" +
	"  async listTasks(filter?: ListFilterInput): Promise<ListResult> {\n" +
	"    const params = new URLSearchParams();\n" +
	"    if (filter) {\n" +
	"      if (filter.milestone) params.append('milestone', filter.milestone);\n" +
	"      if (filter.status) params.append('status', filter.status);\n" +
	"      if (filter.actor) params.append('actor', filter.actor);\n" +
	"      if (filter.limit) params.append('limit', filter.limit.toString());\n" +
	"      if (filter.offset) params.append('offset', filter.offset.toString());\n" +
	"      if (filter.showAll) params.append('showAll', 'true');\n" +
	"    }\n" +
	"    const response = await fetch(`${this.baseUrl}/tasks/list?${params}`);\n" +
	"    if (!response.ok) {\n" +
	"      throw new Error(`Failed to list tasks: ${response.statusText}`);\n" +
	"    }\n" +
	"    return response.json();\n" +
	"  }\n\n" +
	"  /**\n" +
	"   * Reset tasks that have exceeded the timeout\n" +
	"   * @param input - Reset timed out input with timeout in minutes\n" +
	"   * @returns Promise resolving to reset result\n" +
	"   */\n" +
	"  async resetTimedOut(input: ResetTimedOutInput): Promise<ResetTimedOutResult> {\n" +
	"    const response = await fetch(`${this.baseUrl}/tasks/reset`, {\n" +
	"      method: 'POST',\n" +
	"      headers: { 'Content-Type': 'application/json' },\n" +
	"      body: JSON.stringify(input),\n" +
	"    });\n" +
	"    if (!response.ok) {\n" +
	"      throw new Error(`Failed to reset timed out tasks: ${response.statusText}`);\n" +
	"    }\n" +
	"    return response.json();\n" +
	"  }\n" +
	"}\n"

// exportTemplate generates the export statements
const exportTemplate = `// Convenience function to create a TaskManager instance
export function createTaskManager(baseUrl?: string): {{.ClassName}} {
  return new {{.ClassName}}(baseUrl);
}

// Export module
export default {{.ClassName}};
`

// templateData represents the data structure for templates
type templateData struct {
	PackageName string
	ClassName   string
}

// headerFunc returns the header template function
func headerFunc() *template.Template {
	return template.Must(template.New("header").Parse(headerTemplate))
}

// typesFunc returns the types template function
func typesFunc() *template.Template {
	return template.Must(template.New("types").Parse(typesTemplate))
}

// classFunc returns the class template function
func classFunc() *template.Template {
	return template.Must(template.New("class").Parse(classTemplate))
}

// exportFunc returns the export template function
func exportFunc() *template.Template {
	return template.Must(template.New("export").Parse(exportTemplate))
}

// generateFromTemplate generates code from a template with given data
func generateFromTemplate(tmpl *template.Template, data *templateData) (string, error) {
	var sb strings.Builder
	err := tmpl.Execute(&sb, data)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

// GenerateTypeScriptFromTemplates generates TypeScript code using templates
// This is an alternative to the direct string building approach in typescript.go
func GenerateTypeScriptFromTemplates(opts *TypeScriptOptions) (string, error) {
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

	data := &templateData{
		PackageName: opts.PackageName,
		ClassName:   opts.ClassName,
	}

	var result strings.Builder

	// Generate header
	header, err := generateFromTemplate(headerFunc(), data)
	if err != nil {
		return "", fmt.Errorf("failed to generate header: %w", err)
	}
	result.WriteString(header)
	result.WriteString("\n")

	// Generate types (static - no template needed, but include for consistency)
	result.WriteString(typesTemplate)
	result.WriteString("\n")

	// Generate class
	class, err := generateFromTemplate(classFunc(), data)
	if err != nil {
		return "", fmt.Errorf("failed to generate class: %w", err)
	}
	result.WriteString(class)
	result.WriteString("\n")

	// Generate exports
	export, err := generateFromTemplate(exportFunc(), data)
	if err != nil {
		return "", fmt.Errorf("failed to generate export: %w", err)
	}
	result.WriteString(export)

	return result.String(), nil
}
