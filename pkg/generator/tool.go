package generator

import (
	"fmt"
	"strings"
)

// ToolWrapperOptions contains options for tool wrapper generation
type ToolWrapperOptions struct {
	BinaryPath string
}

// DefaultToolWrapperOptions returns default options for tool wrapper generation
func DefaultToolWrapperOptions() *ToolWrapperOptions {
	return &ToolWrapperOptions{
		BinaryPath: "taskflow",
	}
}

// ToolArg represents a tool argument definition
type ToolArg struct {
	Name        string
	Type        string
	Description string
	Required    bool
}

// ToolCommand represents a taskflow command definition
type ToolCommand struct {
	Name        string
	Description string
	Args        []ToolArg
}

// getToolCommands returns the list of taskflow commands and their arguments
func getToolCommands() []ToolCommand {
	return []ToolCommand{
		{
			Name:        "add",
			Description: "Add a new task to the task list. Requires id, milestone, title, and description.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID (required)", Required: true},
				{Name: "title", Type: "string", Description: "Task title (required)", Required: true},
				{Name: "description", Type: "string", Description: "Task description (required)", Required: true},
				{Name: "milestone", Type: "string", Description: "Milestone for the task (required)", Required: true},
				{Name: "actor", Type: "string", Description: "Actor assigned to the task", Required: false},
			},
		},
		{
			Name:        "update",
			Description: "Update an existing task by its ID. At least one update field must be provided.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID (required)", Required: true},
				{Name: "title", Type: "string", Description: "New task title", Required: false},
				{Name: "description", Type: "string", Description: "New task description", Required: false},
				{Name: "status", Type: "string", Description: "New task status (todo, in_progress, done, blocked)", Required: false},
				{Name: "milestone", Type: "string", Description: "New milestone for the task", Required: false},
				{Name: "actor", Type: "string", Description: "New actor assigned to the task", Required: false},
			},
		},
		{
			Name:        "complete",
			Description: "Mark a task as completed by providing its ID.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID (required)", Required: true},
			},
		},
		{
			Name:        "block",
			Description: "Block a task by providing its ID and a reason.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID (required)", Required: true},
				{Name: "reason", Type: "string", Description: "Reason for blocking the task (required)", Required: true},
			},
		},
		{
			Name:        "list",
			Description: "List all tasks with optional filters.",
			Args: []ToolArg{
				{Name: "milestone", Type: "string", Description: "Filter by milestone", Required: false},
				{Name: "status", Type: "string", Description: "Filter by status (todo, in_progress, done, blocked)", Required: false},
				{Name: "actor", Type: "string", Description: "Filter by actor", Required: false},
				{Name: "format", Type: "string", Description: "Output format (table, markdown, xml)", Required: false},
				{Name: "limit", Type: "number", Description: "Maximum number of tasks to display", Required: false},
				{Name: "offset", Type: "number", Description: "Number of tasks to skip", Required: false},
				{Name: "all", Type: "boolean", Description: "Show all tasks including completed", Required: false},
			},
		},
		{
			Name:        "reset_timedout",
			Description: "Reset timed out tasks to todo status. Finds in-progress tasks that have exceeded the specified timeout duration.",
			Args: []ToolArg{
				{Name: "minutes", Type: "number", Description: "Timeout duration in minutes (default: 30)", Required: false},
			},
		},
	}
}

// generateZodSchema generates a Zod schema for a tool argument
func generateZodSchema(arg ToolArg) string {
	switch arg.Type {
	case "string":
		return fmt.Sprintf("tool.schema.string().describe(%q)", arg.Description)
	case "number":
		return fmt.Sprintf("tool.schema.number().describe(%q)", arg.Description)
	case "boolean":
		return fmt.Sprintf("tool.schema.boolean().describe(%q)", arg.Description)
	default:
		return fmt.Sprintf("tool.schema.string().describe(%q)", arg.Description)
	}
}

// generateArgsSchema generates the args schema for a tool
func generateArgsSchema(args []ToolArg) string {
	if len(args) == 0 {
		return "  args: {},\n"
	}

	var b strings.Builder
	b.WriteString("  args: {\n")
	for i, arg := range args {
		b.WriteString(fmt.Sprintf("    %s: %s", arg.Name, generateZodSchema(arg)))
		if i < len(args)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString("  },\n")
	return b.String()
}

// generateCommandArgs generates the command arguments for exec
func generateCommandArgs(cmd ToolCommand, argsVar string) string {
	var b strings.Builder
	b.WriteString("const cmdArgs = [\"")
	b.WriteString(cmd.Name)
	b.WriteString("\"];\n")
	b.WriteString("\n")

	for _, arg := range cmd.Args {
		if arg.Required {
			b.WriteString(fmt.Sprintf("if (args.%s) cmdArgs.push(\"--%s\", args.%s);\n",
				arg.Name, arg.Name, arg.Name))
		} else {
			b.WriteString(fmt.Sprintf("if (args.%s !== undefined) cmdArgs.push(\"--%s\", String(args.%s));\n",
				arg.Name, arg.Name, arg.Name))
		}
	}

	// Handle boolean flags specially (--flag without value)
	for _, arg := range cmd.Args {
		if arg.Type == "boolean" {
			b.WriteString(fmt.Sprintf("if (args.%s) cmdArgs.push(\"--%s\");\n",
				arg.Name, arg.Name))
		}
	}

	return b.String()
}

// GenerateToolWrapper generates a TypeScript tool wrapper using the tool() helper format
func GenerateToolWrapper(opts *ToolWrapperOptions) (string, error) {
	if opts == nil {
		opts = DefaultToolWrapperOptions()
	}

	// Validate binary path
	if opts.BinaryPath == "" {
		opts.BinaryPath = "taskflow"
	}

	commands := getToolCommands()

	var b strings.Builder

	// File header with import
	b.WriteString("import { tool } from \"@opencode-ai/plugin\";\n\n")

	// Generate tool for each command
	for i, cmd := range commands {
		// Tool definition start - convert command name to camelCase for valid JS identifier
		toolName := strings.ReplaceAll(cmd.Name, "-", "")
		b.WriteString(fmt.Sprintf("export const %sTool = tool({\n", toolName))
		b.WriteString(fmt.Sprintf("  description: %q,\n", cmd.Description))

		// Generate args schema
		b.WriteString(generateArgsSchema(cmd.Args))

		// Generate execute function
		b.WriteString("  async execute(args, context) {\n")
		b.WriteString("    // Build command arguments\n")
		b.WriteString("    const actionMap: Record<string, string> = {\n")
		b.WriteString("      \"reset_timedout\": \"reset-timedout\",\n")
		b.WriteString("    };\n")
		b.WriteString(fmt.Sprintf("    const cmdAction = actionMap[\"%s\"] || \"%s\";\n", cmd.Name, cmd.Name))
		b.WriteString("    const cmdArgs = [cmdAction];\n")

		// Add arguments
		for _, arg := range cmd.Args {
			if arg.Type == "boolean" {
				b.WriteString(fmt.Sprintf("    if (args.%s) cmdArgs.push(\"--%s\");\n",
					arg.Name, arg.Name))
			} else if arg.Required {
				b.WriteString(fmt.Sprintf("    if (args.%s) cmdArgs.push(\"--%s\", args.%s);\n",
					arg.Name, arg.Name, arg.Name))
			} else {
				b.WriteString(fmt.Sprintf("    if (args.%s !== undefined) cmdArgs.push(\"--%s\", String(args.%s));\n",
					arg.Name, arg.Name, arg.Name))
			}
		}

		// Default format to xml for list command
		if cmd.Name == "list" {
			b.WriteString("    if (args.format === undefined) cmdArgs.push(\"--format\", \"xml\");\n")
		}

		// Add exec import and call
		b.WriteString("\n")
		b.WriteString("    // Execute taskflow command\n")
		b.WriteString("    const { execFileSync } = await import(\"child_process\");\n")
		b.WriteString(fmt.Sprintf("    const result = execFileSync(`%s`, cmdArgs, {\n", opts.BinaryPath))
		b.WriteString("      encoding: \"utf-8\", stdio: [\"pipe\", \"pipe\", \"pipe\"]\n")
		b.WriteString("    });\n\n")
		b.WriteString("    return result;\n")
		b.WriteString("  },\n")
		b.WriteString("});\n")

		// Add blank line between tools (except for the last one)
		if i < len(commands)-1 {
			b.WriteString("\n")
		}
	}

	// Generate combined tool that accepts an action parameter
	b.WriteString("// Combined tool that routes to specific commands based on 'action' parameter\nexport const taskflowTool = tool({\n")
	b.WriteString("  description: \"TaskFlow CLI - Manage tasks with add, update, complete, block, list, and reset_timedout commands\",\n")
	b.WriteString("  args: {\n")
	b.WriteString("    action: tool.schema.enum([\"add\", \"update\", \"complete\", \"block\", \"list\", \"reset_timedout\"]).describe(\"The taskflow command to execute\"),\n")
	b.WriteString("    id: tool.schema.string().optional().describe(\"Task ID\"),\n")
	b.WriteString("    title: tool.schema.string().optional().describe(\"Task title\"),\n")
	b.WriteString("    description: tool.schema.string().optional().describe(\"Task description\"),\n")
	b.WriteString("    milestone: tool.schema.string().optional().describe(\"Milestone for the task\"),\n")
	b.WriteString("    status: tool.schema.string().optional().describe(\"Task status (todo, in_progress, done, blocked)\"),\n")
	b.WriteString("    actor: tool.schema.string().optional().describe(\"Actor assigned to the task\"),\n")
	b.WriteString("    reason: tool.schema.string().optional().describe(\"Reason for blocking the task\"),\n")
	b.WriteString("    minutes: tool.schema.number().optional().describe(\"Timeout duration in minutes\"),\n")
	b.WriteString("    format: tool.schema.string().optional().describe(\"Output format (table, markdown, xml)\"),\n")
	b.WriteString("    limit: tool.schema.number().optional().describe(\"Maximum number of tasks to display\"),\n")
	b.WriteString("    offset: tool.schema.number().optional().describe(\"Number of tasks to skip\"),\n")
	b.WriteString("    all: tool.schema.boolean().optional().describe(\"Show all tasks including completed\"),\n")
	b.WriteString("    filter_milestone: tool.schema.string().optional().describe(\"Filter list by milestone\"),\n")
	b.WriteString("    filter_status: tool.schema.string().optional().describe(\"Filter list by status\"),\n")
	b.WriteString("    filter_actor: tool.schema.string().optional().describe(\"Filter list by actor\"),\n")
	b.WriteString("    formatted: tool.schema.boolean().optional().describe(\"Output as markdown table for list action\"),\n")
	b.WriteString("  },\n")
	b.WriteString("  async execute(args, context) {\n")
	b.WriteString("    const actionMap: Record<string, string> = {\n")
	b.WriteString("      \"reset_timedout\": \"reset-timedout\",\n")
	b.WriteString("    };\n")
	b.WriteString("    const cmdAction = actionMap[args.action] || args.action;\n")
	b.WriteString("    const cmdArgs = [cmdAction];\n")
	b.WriteString("\n")
	b.WriteString("    // Add optional arguments\n")
	b.WriteString("    if (args.id) cmdArgs.push(\"--id\", args.id);\n")
	b.WriteString("    if (args.title) cmdArgs.push(\"--title\", args.title);\n")
	b.WriteString("    if (args.description) cmdArgs.push(\"--description\", args.description);\n")
	b.WriteString("    if (args.reason) cmdArgs.push(\"--reason\", args.reason);\n")
	b.WriteString("    if (args.minutes !== undefined) cmdArgs.push(\"--minutes\", String(args.minutes));\n")
	b.WriteString("    if (args.format) cmdArgs.push(\"--format\", args.format);\n")
	b.WriteString("    if (args.action === \"list\" && !args.format) cmdArgs.push(\"--format\", \"xml\");\n")
	b.WriteString("    if (args.limit !== undefined) cmdArgs.push(\"--limit\", String(args.limit));\n")
	b.WriteString("    if (args.offset !== undefined) cmdArgs.push(\"--offset\", String(args.offset));\n")
	b.WriteString("    if (args.all) cmdArgs.push(\"--all\");\n")
	b.WriteString("    if (args.formatted) cmdArgs.push(\"--formatted\");\n")
	b.WriteString("\n")
	b.WriteString("    // Action-specific arguments\n")
	b.WriteString("    if (args.action === \"list\") {\n")
	b.WriteString("      if (args.filter_milestone) cmdArgs.push(\"--milestone\", args.filter_milestone);\n")
	b.WriteString("      if (args.filter_status) cmdArgs.push(\"--status\", args.filter_status);\n")
	b.WriteString("      if (args.filter_actor) cmdArgs.push(\"--actor\", args.filter_actor);\n")
	b.WriteString("    } else {\n")
	b.WriteString("      if (args.milestone) cmdArgs.push(\"--milestone\", args.milestone);\n")
	b.WriteString("      if (args.status) cmdArgs.push(\"--status\", args.status);\n")
	b.WriteString("      if (args.actor) cmdArgs.push(\"--actor\", args.actor);\n")
	b.WriteString("    }\n")
	b.WriteString("    // Execute taskflow command\n")
	b.WriteString("    const { execFileSync } = await import(\"child_process\");\n")
	b.WriteString(fmt.Sprintf("    const result = execFileSync(`%s`, cmdArgs,\n", opts.BinaryPath))
	b.WriteString("      { encoding: \"utf-8\", stdio: [\"pipe\", \"pipe\", \"pipe\"] }\n")
	b.WriteString("    );\n\n")
	b.WriteString("    return result;\n")
	b.WriteString("  },\n")
	b.WriteString("});\n\n")
	b.WriteString("// Export default\n")
	b.WriteString("export default taskflowTool;\n")

	return b.String(), nil
}
