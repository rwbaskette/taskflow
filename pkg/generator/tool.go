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
	SortByTitle       = "title"
	SortByDescription = "description"
	SortByActor       = "actor"
	ValidSortBy       = "status, priority, milestone, created, updated, id, title, description, actor"
)

var validFormatValues = []string{FormatTable, FormatMarkdown, FormatXML}
var validStatusValues = []string{StatusTodo, StatusInProgress, StatusDone, StatusBlocked}
var validSortByValues = []string{SortByStatus, SortByPriority, SortByMilestone, SortByCreated, SortByUpdated, SortByID, SortByTitle, SortByDescription, SortByActor}

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
			Name:        "block",
			Description: "Block a task by providing its ID and a reason.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID (required)", Required: true},
				{Name: "reason", Type: "string", Description: "Reason for blocking the task (required)", Required: true},
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
			Name:        "delete",
			Description: "Soft delete a task by moving it to the deleted_tasks table.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID (required)", Required: true},
			},
		},
		{
			Name:        "list_all",
			Description: "List all tasks with optional milestone filter.",
			Args: []ToolArg{
				{Name: "milestone", Type: "string", Description: "Filter by milestone", Required: false},
				{Name: "actor", Type: "string", Description: "Filter by actor", Required: false},
				{Name: "limit", Type: "number", Description: "Maximum number of tasks to display", Required: false},
				{Name: "offset", Type: "number", Description: "Number of tasks to skip", Required: false},
				{Name: "all", Type: "boolean", Description: "Show all tasks including completed", Required: false},
			},
		},
		{
			Name:        "list_blocked",
			Description: "List blocked tasks with optional milestone filter.",
			Args: []ToolArg{
				{Name: "milestone", Type: "string", Description: "Filter by milestone", Required: false},
				{Name: "actor", Type: "string", Description: "Filter by actor", Required: false},
				{Name: "limit", Type: "number", Description: "Maximum number of tasks to display", Required: false},
				{Name: "offset", Type: "number", Description: "Number of tasks to skip", Required: false},
			},
		},
		{
			Name:        "list_done",
			Description: "List completed tasks with optional milestone filter.",
			Args: []ToolArg{
				{Name: "milestone", Type: "string", Description: "Filter by milestone", Required: false},
				{Name: "actor", Type: "string", Description: "Filter by actor", Required: false},
				{Name: "limit", Type: "number", Description: "Maximum number of tasks to display", Required: false},
				{Name: "offset", Type: "number", Description: "Number of tasks to skip", Required: false},
			},
		},
		{
			Name:        "list_status_in_progress",
			Description: "List in-progress tasks with optional milestone filter.",
			Args: []ToolArg{
				{Name: "milestone", Type: "string", Description: "Filter by milestone", Required: false},
				{Name: "actor", Type: "string", Description: "Filter by actor", Required: false},
				{Name: "limit", Type: "number", Description: "Maximum number of tasks to display", Required: false},
				{Name: "offset", Type: "number", Description: "Number of tasks to skip", Required: false},
			},
		},
		{
			Name:        "list_status_todo",
			Description: "List todo tasks with optional milestone filter.",
			Args: []ToolArg{
				{Name: "milestone", Type: "string", Description: "Filter by milestone", Required: false},
				{Name: "actor", Type: "string", Description: "Filter by actor", Required: false},
				{Name: "limit", Type: "number", Description: "Maximum number of tasks to display", Required: false},
				{Name: "offset", Type: "number", Description: "Number of tasks to skip", Required: false},
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
			Name:        "block",
			Description: "Block a task by providing its ID and a reason.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID (required)", Required: true},
				{Name: "reason", Type: "string", Description: "Reason for blocking the task (required)", Required: true},
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
			Name:        "delete",
			Description: "Soft delete a task by moving it to the deleted_tasks table.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID (required)", Required: true},
			},
		},
		{
			Name:        "list_all",
			Description: "List all tasks with optional milestone filter.",
			Args: []ToolArg{
				{Name: "milestone", Type: "string", Description: "Filter by milestone", Required: false},
				{Name: "actor", Type: "string", Description: "Filter by actor", Required: false},
				{Name: "limit", Type: "number", Description: "Maximum number of tasks to display", Required: false},
				{Name: "offset", Type: "number", Description: "Number of tasks to skip", Required: false},
				{Name: "all", Type: "boolean", Description: "Show all tasks including completed", Required: false},
			},
		},
		{
			Name:        "list_blocked",
			Description: "List blocked tasks with optional milestone filter.",
			Args: []ToolArg{
				{Name: "milestone", Type: "string", Description: "Filter by milestone", Required: false},
				{Name: "actor", Type: "string", Description: "Filter by actor", Required: false},
				{Name: "limit", Type: "number", Description: "Maximum number of tasks to display", Required: false},
				{Name: "offset", Type: "number", Description: "Number of tasks to skip", Required: false},
			},
		},
		{
			Name:        "list_done",
			Description: "List completed tasks with optional milestone filter.",
			Args: []ToolArg{
				{Name: "milestone", Type: "string", Description: "Filter by milestone", Required: false},
				{Name: "actor", Type: "string", Description: "Filter by actor", Required: false},
				{Name: "limit", Type: "number", Description: "Maximum number of tasks to display", Required: false},
				{Name: "offset", Type: "number", Description: "Number of tasks to skip", Required: false},
			},
		},
		{
			Name:        "list_status_in_progress",
			Description: "List in-progress tasks with optional milestone filter.",
			Args: []ToolArg{
				{Name: "milestone", Type: "string", Description: "Filter by milestone", Required: false},
				{Name: "actor", Type: "string", Description: "Filter by actor", Required: false},
				{Name: "limit", Type: "number", Description: "Maximum number of tasks to display", Required: false},
				{Name: "offset", Type: "number", Description: "Number of tasks to skip", Required: false},
			},
		},
		{
			Name:        "list_status_todo",
			Description: "List todo tasks with optional milestone filter.",
			Args: []ToolArg{
				{Name: "milestone", Type: "string", Description: "Filter by milestone", Required: false},
				{Name: "actor", Type: "string", Description: "Filter by actor", Required: false},
				{Name: "limit", Type: "number", Description: "Maximum number of tasks to display", Required: false},
				{Name: "offset", Type: "number", Description: "Number of tasks to skip", Required: false},
			},
		},
		{
			Name:        "reset_timedout",
			Description: "Reset timed out tasks to todo status. Finds in-progress tasks that have exceeded the specified timeout duration.",
			Args: []ToolArg{
				{Name: "minutes", Type: "number", Description: "Timeout duration in minutes (default: 30)", Required: false},
			},
		},
		{
			Name:        "start",
			Description: "Start working on a task by moving it to in-progress status.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID (required)", Required: true},
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
		toolName := fmt.Sprintf("task_%s", cmd.Name)
		b.WriteString(fmt.Sprintf("export const %s = tool({\n", toolName))
		b.WriteString(fmt.Sprintf("  description: %q,\n", cmd.Description))
		b.WriteString(generateArgsSchema(cmd.Args, cmd.Enums))

		b.WriteString("  async execute(args, context) {\n")
		b.WriteString("    const cmdArgs = [];\n")

		switch cmd.Name {
		case "add":
			b.WriteString("    cmdArgs.push(\"add\");\n")
			b.WriteString("    const addJSON = JSON.stringify({\n")
			b.WriteString("      id: args.id,\n")
			b.WriteString("      title: args.title,\n")
			b.WriteString("      description: args.description,\n")
			b.WriteString("      milestone: args.milestone,\n")
			b.WriteString("      actor: args.actor\n")
			b.WriteString("    });\n")
			b.WriteString("    cmdArgs.push(addJSON);\n")

		case "block":
			b.WriteString("    cmdArgs.push(\"block\");\n")
			b.WriteString("    const blockJSON = JSON.stringify({\n")
			b.WriteString("      id: args.id,\n")
			b.WriteString("      reason: args.reason\n")
			b.WriteString("    });\n")
			b.WriteString("    cmdArgs.push(blockJSON);\n")

		case "complete":
			b.WriteString("    cmdArgs.push(\"complete\");\n")
			b.WriteString("    const completeJSON = JSON.stringify({ id: args.id });\n")
			b.WriteString("    cmdArgs.push(completeJSON);\n")

		case "delete":
			b.WriteString("    cmdArgs.push(\"delete\");\n")
			b.WriteString("    const deleteJSON = JSON.stringify({ id: args.id });\n")
			b.WriteString("    cmdArgs.push(deleteJSON);\n")

		case "list_all":
			b.WriteString("    cmdArgs.push(\"list\");\n")
			b.WriteString("    const listAllJSON = JSON.stringify({\n")
			b.WriteString("      milestone: args.milestone,\n")
			b.WriteString("      actor: args.actor,\n")
			b.WriteString("      limit: args.limit,\n")
			b.WriteString("      offset: args.offset,\n")
			b.WriteString("      all: args.all\n")
			b.WriteString("    });\n")
			b.WriteString("    cmdArgs.push(listAllJSON);\n")

		case "list_blocked":
			b.WriteString("    cmdArgs.push(\"list\");\n")
			b.WriteString("    const listBlockedJSON = JSON.stringify({\n")
			b.WriteString("      milestone: args.milestone,\n")
			b.WriteString("      actor: args.actor,\n")
			b.WriteString("      limit: args.limit,\n")
			b.WriteString("      offset: args.offset,\n")
			b.WriteString("      status: \"blocked\"\n")
			b.WriteString("    });\n")
			b.WriteString("    cmdArgs.push(listBlockedJSON);\n")

		case "list_done":
			b.WriteString("    cmdArgs.push(\"list\");\n")
			b.WriteString("    const listDoneJSON = JSON.stringify({\n")
			b.WriteString("      milestone: args.milestone,\n")
			b.WriteString("      actor: args.actor,\n")
			b.WriteString("      limit: args.limit,\n")
			b.WriteString("      offset: args.offset,\n")
			b.WriteString("      status: \"done\"\n")
			b.WriteString("    });\n")
			b.WriteString("    cmdArgs.push(listDoneJSON);\n")

		case "list_status_in_progress":
			b.WriteString("    cmdArgs.push(\"list\");\n")
			b.WriteString("    const listInProgressJSON = JSON.stringify({\n")
			b.WriteString("      milestone: args.milestone,\n")
			b.WriteString("      actor: args.actor,\n")
			b.WriteString("      limit: args.limit,\n")
			b.WriteString("      offset: args.offset,\n")
			b.WriteString("      status: \"in_progress\"\n")
			b.WriteString("    });\n")
			b.WriteString("    cmdArgs.push(listInProgressJSON);\n")

		case "list_status_todo":
			b.WriteString("    cmdArgs.push(\"list\");\n")
			b.WriteString("    const listTodoJSON = JSON.stringify({\n")
			b.WriteString("      milestone: args.milestone,\n")
			b.WriteString("      actor: args.actor,\n")
			b.WriteString("      limit: args.limit,\n")
			b.WriteString("      offset: args.offset,\n")
			b.WriteString("      status: \"todo\"\n")
			b.WriteString("    });\n")
			b.WriteString("    cmdArgs.push(listTodoJSON);\n")

		case "reset_timedout":
			b.WriteString("    cmdArgs.push(\"reset-timedout\");\n")
			b.WriteString("    const resetJSON = JSON.stringify({ minutes: args.minutes });\n")
			b.WriteString("    cmdArgs.push(resetJSON);\n")

		case "start":
			b.WriteString("    cmdArgs.push(\"update\");\n")
			b.WriteString("    const updateJSON = JSON.stringify({ id: args.id, status: \"in_progress\" });\n")
			b.WriteString("    cmdArgs.push(updateJSON);\n")
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

	return b.String(), nil
}
