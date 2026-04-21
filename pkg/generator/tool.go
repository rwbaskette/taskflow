package generator

import (
	"fmt"
	"strings"
)

const (
	FormatTable    = "table"
	FormatMarkdown = "markdown"
	FormatXML      = "xml"
	ValidFormats   = "table, markdown, xml"
)

const (
	StatusTodo       = "todo"
	StatusInProgress = "in_progress"
	StatusDone       = "done"
	StatusBlocked    = "blocked"
	ValidStatuses    = "todo, in_progress, done, blocked"
)

const (
	SortByStatus      = "status"
	SortByPriority    = "priority"
	SortByMilestone   = "milestone"
	SortByCreated     = "created"
	SortByUpdated     = "updated"
	SortByID          = "id"
	SortBySprint      = "sprint"
	SortByTitle       = "title"
	SortByDescription = "description"
	SortByActor       = "actor"
	ValidSortBy       = "status, priority, milestone, created, updated, id, sprint, title, description, actor"
)

var validFormatValues = []string{FormatTable, FormatMarkdown, FormatXML}
var validStatusValues = []string{StatusTodo, StatusInProgress, StatusDone, StatusBlocked}
var validSortByValues = []string{SortByStatus, SortByPriority, SortByMilestone, SortByCreated, SortByUpdated, SortByID, SortBySprint, SortByTitle, SortByDescription, SortByActor}

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

// ToolEnumValue represents a single enum value for an argument
type ToolEnumValue struct {
	Value       string
	Description string
}

// ToolEnum represents an enumeration argument with predefined values
type ToolEnum struct {
	Name        string
	Values      []ToolEnumValue
	Description string
}

// ToolCommand represents a taskflow command definition
type ToolCommand struct {
	Name        string
	Description string
	Args        []ToolArg
	Enums       []ToolEnum
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
				{Name: "status", Type: "string", Description: "New task status (" + ValidStatuses + ")", Required: false},
				{Name: "milestone", Type: "string", Description: "New milestone for the task", Required: false},
				{Name: "actor", Type: "string", Description: "New actor assigned to the task", Required: false},
			},
		},
		{
			Name:        "complete",
			Description: "Mark a task as completed by providing its ID.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID (required)", Required: true},
				{Name: "title", Type: "string", Description: "New task title", Required: false},
				{Name: "description", Type: "string", Description: "New task description", Required: false},
				{Name: "status", Type: "string", Description: "New task status (" + ValidStatuses + ")", Required: false},
				{Name: "milestone", Type: "string", Description: "New milestone for the task", Required: false},
				{Name: "actor", Type: "string", Description: "New actor assigned to the task", Required: false},
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
				{Name: "actor", Type: "string", Description: "Filter by actor", Required: false},
				{Name: "limit", Type: "number", Description: "Maximum number of tasks to display", Required: false},
				{Name: "offset", Type: "number", Description: "Number of tasks to skip", Required: false},
				{Name: "all", Type: "boolean", Description: "Show all tasks including completed", Required: false},
				{Name: "sortBy", Type: "string", Description: "Sort by field (" + ValidSortBy + ")", Required: false},
			},
			Enums: []ToolEnum{
				{
					Name:        "status",
					Description: "Filter by status",
					Values: []ToolEnumValue{
						{Value: StatusTodo, Description: "Tasks in todo status"},
						{Value: StatusInProgress, Description: "Tasks in progress"},
						{Value: StatusDone, Description: "Completed tasks"},
						{Value: StatusBlocked, Description: "Blocked tasks"},
					},
				},
				{
					Name:        "format",
					Description: "Output format",
					Values: []ToolEnumValue{
						{Value: FormatTable, Description: "Table format output"},
						{Value: FormatMarkdown, Description: "Markdown format output"},
						{Value: FormatXML, Description: "XML format output"},
					},
				},
			},
		},
		{
			Name:        "delete",
			Description: "Soft delete a task by moving it to the deleted_tasks table.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID (required)", Required: true},
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

// getToolCommandsWithEnums returns commands with enum values expanded into specialized tools
func getToolCommandsWithEnums() []ToolCommand {
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
				{Name: "status", Type: "string", Description: "New task status (" + ValidStatuses + ")", Required: false},
				{Name: "milestone", Type: "string", Description: "New milestone for the task", Required: false},
				{Name: "actor", Type: "string", Description: "New actor assigned to the task", Required: false},
			},
		},
		{
			Name:        "complete",
			Description: "Mark a task as completed by providing its ID.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID (required)", Required: true},
				{Name: "title", Type: "string", Description: "New task title", Required: false},
				{Name: "description", Type: "string", Description: "New task description", Required: false},
				{Name: "status", Type: "string", Description: "New task status (" + ValidStatuses + ")", Required: false},
				{Name: "milestone", Type: "string", Description: "New milestone for the task", Required: false},
				{Name: "actor", Type: "string", Description: "New actor assigned to the task", Required: false},
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
				{Name: "actor", Type: "string", Description: "Filter by actor", Required: false},
				{Name: "limit", Type: "number", Description: "Maximum number of tasks to display", Required: false},
				{Name: "offset", Type: "number", Description: "Number of tasks to skip", Required: false},
				{Name: "all", Type: "boolean", Description: "Show all tasks including completed", Required: false},
				{Name: "sortBy", Type: "string", Description: "Sort by field (" + ValidSortBy + ")", Required: false},
			},
			Enums: []ToolEnum{
				{
					Name:        "status",
					Description: "Filter by status",
					Values: []ToolEnumValue{
						{Value: StatusTodo, Description: "Tasks in todo status"},
						{Value: StatusInProgress, Description: "Tasks in progress"},
						{Value: StatusDone, Description: "Completed tasks"},
						{Value: StatusBlocked, Description: "Blocked tasks"},
					},
				},
				{
					Name:        "format",
					Description: "Output format",
					Values: []ToolEnumValue{
						{Value: FormatTable, Description: "Table format output"},
						{Value: FormatMarkdown, Description: "Markdown format output"},
						{Value: FormatXML, Description: "XML format output"},
					},
				},
			},
		},
		{
			Name:        "list_status",
			Description: "List tasks filtered by status.",
			Args: []ToolArg{
				{Name: "milestone", Type: "string", Description: "Filter by milestone", Required: false},
				{Name: "actor", Type: "string", Description: "Filter by actor", Required: false},
				{Name: "limit", Type: "number", Description: "Maximum number of tasks to display", Required: false},
				{Name: "offset", Type: "number", Description: "Number of tasks to skip", Required: false},
				{Name: "all", Type: "boolean", Description: "Show all tasks including completed", Required: false},
				{Name: "sortBy", Type: "string", Description: "Sort by field (" + ValidSortBy + ")", Required: false},
			},
		},
		{
			Name:        "delete",
			Description: "Soft delete a task by moving it to the deleted_tasks table.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID (required)", Required: true},
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

// generateEnumSchema generates a Zod enum schema
func generateEnumSchema(enum ToolEnum) string {
	values := make([]string, len(enum.Values))
	descs := make([]string, len(enum.Values))
	for i, v := range enum.Values {
		values[i] = fmt.Sprintf("%q", v.Value)
		descs[i] = fmt.Sprintf("%s: %s", v.Value, v.Description)
	}
	return fmt.Sprintf("tool.schema.enum([%s]).describe(%q)", strings.Join(values, ", "), enum.Description)
}

// generateArgsSchema generates the args schema for a tool
func generateArgsSchema(args []ToolArg, enums []ToolEnum) string {
	if len(args) == 0 && len(enums) == 0 {
		return "  args: {},\n"
	}

	var b strings.Builder
	b.WriteString("  args: {\n")

	// Generate enum args first (status, format)
	for _, e := range enums {
		b.WriteString(fmt.Sprintf("    %s: %s", e.Name, generateEnumSchema(e)))
		b.WriteString(",\n")
	}

	// Generate regular args
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

// sanitizeName converts a command name to a valid TypeScript identifier
func sanitizeName(name string) string {
	name = strings.ReplaceAll(name, "-", "_")
	return name
}

// toCamelCase converts a name to camelCase for tool names
func toCamelCase(name string) string {
	parts := strings.Split(name, "_")
	if len(parts) == 1 {
		return parts[0]
	}
	var b strings.Builder
	b.WriteString(strings.ToLower(parts[0]))
	for _, p := range parts[1:] {
		b.WriteString(strings.ToUpper(string(p[0])))
		b.WriteString(p[1:])
	}
	return b.String()
}

// GenerateToolWrapper generates a TypeScript tool wrapper using the tool() helper format
func GenerateToolWrapper(opts *ToolWrapperOptions) (string, error) {
	if opts == nil {
		opts = DefaultToolWrapperOptions()
	}

	if opts.BinaryPath == "" {
		opts.BinaryPath = "taskflow"
	}

	commands := getToolCommandsWithEnums()

	var b strings.Builder

	b.WriteString("import { tool } from \"@opencode-ai/plugin\";\n\n")

	for i, cmd := range commands {
		toolName := fmt.Sprintf("taskflow_%s", strings.ReplaceAll(cmd.Name, "_", "_"))
		b.WriteString(fmt.Sprintf("export const %sTool = tool({\n", toolName))
		b.WriteString(fmt.Sprintf("  description: %q,\n", cmd.Description))
		b.WriteString(generateArgsSchema(cmd.Args, cmd.Enums))

		b.WriteString("  async execute(args, context) {\n")
		b.WriteString("    const actionMap: Record<string, string> = {\n")
		b.WriteString("      \"reset_timedout\": \"reset-timedout\",\n")
		b.WriteString("    };\n")
		b.WriteString(fmt.Sprintf("    const cmdAction = actionMap[\"%s\"] || \"%s\";\n", cmd.Name, cmd.Name))
		b.WriteString("    const cmdArgs = [cmdAction];\n")

		for _, arg := range cmd.Args {
			if arg.Name == "sortBy" && (cmd.Name == "list" || cmd.Name == "list_status") {
				continue
			}
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

		if cmd.Name == "list" {
			b.WriteString("    if (args.status) cmdArgs.push(\"--status\", args.status);\n")
			b.WriteString("    if (args.format) cmdArgs.push(\"--format\", args.format);\n")
			b.WriteString("    if (!args.format) cmdArgs.push(\"--format\", \"xml\");\n")
			b.WriteString("    if (args.sortBy) cmdArgs.push(\"--sort-by\", args.sortBy);\n")
		} else if cmd.Name == "list_status" {
			b.WriteString("    if (args.status) cmdArgs.push(\"--status\", args.status);\n")
			b.WriteString("    if (args.format) cmdArgs.push(\"--format\", args.format);\n")
			b.WriteString("    if (args.sortBy) cmdArgs.push(\"--sort-by\", args.sortBy);\n")
		}

		b.WriteString("\n")
		b.WriteString("    const { execFileSync } = await import(\"child_process\");\n")
		b.WriteString(fmt.Sprintf("    const result = execFileSync(`%s`, cmdArgs, {\n", opts.BinaryPath))
		b.WriteString("      encoding: \"utf-8\", stdio: [\"pipe\", \"pipe\", \"pipe\"]\n")
		b.WriteString("    });\n\n")
		b.WriteString("    return result;\n")
		b.WriteString("  },\n")
		b.WriteString("});\n")

		if i < len(commands)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n// Enum-based tools for list with format\n")
	for _, format := range validFormatValues {
		toolName := fmt.Sprintf("taskflow_list_format_%s", format)
		b.WriteString(fmt.Sprintf("export const %sTool = tool({\n", toolName))
		b.WriteString(fmt.Sprintf("  description: %q,\n", fmt.Sprintf("List tasks in %s format", format)))
		b.WriteString("  args: {\n")
		b.WriteString("    milestone: tool.schema.string().optional().describe(\"Filter by milestone\"),\n")
		b.WriteString("    status: tool.schema.string().optional().describe(\"Filter by status\"),\n")
		b.WriteString("    actor: tool.schema.string().optional().describe(\"Filter by actor\"),\n")
		b.WriteString("    limit: tool.schema.number().optional().describe(\"Maximum number of tasks to display\"),\n")
		b.WriteString("    offset: tool.schema.number().optional().describe(\"Number of tasks to skip\"),\n")
		b.WriteString("    all: tool.schema.boolean().optional().describe(\"Show all tasks including completed\"),\n")
		b.WriteString("    sortBy: tool.schema.string().optional().describe(\"Sort by field\"),\n")
		b.WriteString("  },\n")
		b.WriteString("  async execute(args, context) {\n")
		b.WriteString("    const cmdArgs = [\"list\"];\n")
		b.WriteString("    if (args.milestone) cmdArgs.push(\"--milestone\", args.milestone);\n")
		b.WriteString("    if (args.status) cmdArgs.push(\"--status\", args.status);\n")
		b.WriteString("    if (args.actor) cmdArgs.push(\"--actor\", args.actor);\n")
		b.WriteString("    if (args.limit !== undefined) cmdArgs.push(\"--limit\", String(args.limit));\n")
		b.WriteString("    if (args.offset !== undefined) cmdArgs.push(\"--offset\", String(args.offset));\n")
		b.WriteString("    if (args.all) cmdArgs.push(\"--all\");\n")
		b.WriteString("    if (args.sortBy) cmdArgs.push(\"--sort-by\", args.sortBy);\n")
		b.WriteString(fmt.Sprintf("    cmdArgs.push(\"--format\", %q);\n", format))
		b.WriteString("    const { execFileSync } = await import(\"child_process\");\n")
		b.WriteString(fmt.Sprintf("    const result = execFileSync(`%s`, cmdArgs, {\n", opts.BinaryPath))
		b.WriteString("      encoding: \"utf-8\", stdio: [\"pipe\", \"pipe\", \"pipe\"]\n")
		b.WriteString("    });\n\n")
		b.WriteString("    return result;\n")
		b.WriteString("  },\n")
		b.WriteString("});\n\n")
	}

	b.WriteString("// Status-based tools for list\n")
	for _, status := range validStatusValues {
		toolName := fmt.Sprintf("taskflow_list_status_%s", status)
		b.WriteString(fmt.Sprintf("export const %sTool = tool({\n", toolName))
		b.WriteString(fmt.Sprintf("  description: %q,\n", fmt.Sprintf("List tasks with status %s", status)))
		b.WriteString("  args: {\n")
		b.WriteString("    milestone: tool.schema.string().optional().describe(\"Filter by milestone\"),\n")
		b.WriteString("    actor: tool.schema.string().optional().describe(\"Filter by actor\"),\n")
		b.WriteString("    limit: tool.schema.number().optional().describe(\"Maximum number of tasks to display\"),\n")
		b.WriteString("    offset: tool.schema.number().optional().describe(\"Number of tasks to skip\"),\n")
		b.WriteString("    all: tool.schema.boolean().optional().describe(\"Show all tasks including completed\"),\n")
		b.WriteString("    sortBy: tool.schema.string().optional().describe(\"Sort by field\"),\n")
		b.WriteString("    format: tool.schema.string().optional().describe(\"Output format\"),\n")
		b.WriteString("  },\n")
		b.WriteString("  async execute(args, context) {\n")
		b.WriteString("    const cmdArgs = [\"list\"];\n")
		b.WriteString("    if (args.milestone) cmdArgs.push(\"--milestone\", args.milestone);\n")
		b.WriteString(fmt.Sprintf("    cmdArgs.push(\"--status\", %q);\n", status))
		b.WriteString("    if (args.actor) cmdArgs.push(\"--actor\", args.actor);\n")
		b.WriteString("    if (args.limit !== undefined) cmdArgs.push(\"--limit\", String(args.limit));\n")
		b.WriteString("    if (args.offset !== undefined) cmdArgs.push(\"--offset\", String(args.offset));\n")
		b.WriteString("    if (args.all) cmdArgs.push(\"--all\");\n")
		b.WriteString("    if (args.sortBy) cmdArgs.push(\"--sort-by\", args.sortBy);\n")
		b.WriteString("    if (args.format) cmdArgs.push(\"--format\", args.format);\n")
		b.WriteString("    if (!args.format) cmdArgs.push(\"--format\", \"xml\");\n")
		b.WriteString("    const { execFileSync } = await import(\"child_process\");\n")
		b.WriteString(fmt.Sprintf("    const result = execFileSync(`%s`, cmdArgs, {\n", opts.BinaryPath))
		b.WriteString("      encoding: \"utf-8\", stdio: [\"pipe\", \"pipe\", \"pipe\"]\n")
		b.WriteString("    });\n\n")
		b.WriteString("    return result;\n")
		b.WriteString("  },\n")
		b.WriteString("});\n\n")
	}

	b.WriteString("// Sprint tools\n")
	b.WriteString("export const taskflow_sprint_completeTool = tool({\n")
	b.WriteString("  description: \"Mark a task as completed by providing its ID.\",\n")
	b.WriteString("  args: {\n")
	b.WriteString("    id: tool.schema.string().describe(\"Task ID (required)\"),\n")
	b.WriteString("    title: tool.schema.string().optional().describe(\"New task title\"),\n")
	b.WriteString("    description: tool.schema.string().optional().describe(\"New task description\"),\n")
	b.WriteString("    status: tool.schema.string().optional().describe(\"New task status\"),\n")
	b.WriteString("    milestone: tool.schema.string().optional().describe(\"New milestone for the task\"),\n")
	b.WriteString("    actor: tool.schema.string().optional().describe(\"New actor assigned to the task\"),\n")
	b.WriteString("  },\n")
	b.WriteString("  async execute(args, context) {\n")
	b.WriteString("    const cmdArgs = [\"complete\"];\n")
	b.WriteString("    if (args.id) cmdArgs.push(\"--id\", args.id);\n")
	b.WriteString("    if (args.title) cmdArgs.push(\"--title\", args.title);\n")
	b.WriteString("    if (args.description) cmdArgs.push(\"--description\", args.description);\n")
	b.WriteString("    if (args.status) cmdArgs.push(\"--status\", args.status);\n")
	b.WriteString("    if (args.milestone) cmdArgs.push(\"--milestone\", args.milestone);\n")
	b.WriteString("    if (args.actor) cmdArgs.push(\"--actor\", args.actor);\n")
	b.WriteString("    const { execFileSync } = await import(\"child_process\");\n")
	b.WriteString(fmt.Sprintf("    const result = execFileSync(`%s`, cmdArgs, {\n", opts.BinaryPath))
	b.WriteString("      encoding: \"utf-8\", stdio: [\"pipe\", \"pipe\", \"pipe\"]\n")
	b.WriteString("    });\n\n")
	b.WriteString("    return result;\n")
	b.WriteString("  },\n")
	b.WriteString("});\n")

	return b.String(), nil
}
