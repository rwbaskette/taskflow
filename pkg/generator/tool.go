package generator

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
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

// ToolWrapperOptions contains options for tool wrapper generation
type ToolWrapperOptions struct {
	BinaryPath string
}

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
	// CLISubcommand is the subcommand passed to the binary (e.g. "list", "update").
	// Defaults to Name (with underscores replaced by hyphens) if empty.
	CLISubcommand string
	// FixedStatus, when non-empty, is serialised as a literal "status" field in
	// the JSON payload sent to the binary.
	FixedStatus string
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
			Name:          "list_all",
			CLISubcommand: "list",
			Description:   "List all tasks with optional milestone filter.",
			Args: []ToolArg{
				{Name: "milestone", Type: "string", Description: "Filter by milestone", Required: false},
				{Name: "actor", Type: "string", Description: "Filter by actor", Required: false},
				{Name: "limit", Type: "number", Description: "Maximum number of tasks to display", Required: false},
				{Name: "offset", Type: "number", Description: "Number of tasks to skip", Required: false},
				{Name: "all", Type: "boolean", Description: "Show all tasks including completed", Required: false},
			},
		},
		{
			Name:          "list_blocked",
			CLISubcommand: "list",
			FixedStatus:   "blocked",
			Description:   "List blocked tasks with optional milestone filter.",
			Args: []ToolArg{
				{Name: "milestone", Type: "string", Description: "Filter by milestone", Required: false},
				{Name: "actor", Type: "string", Description: "Filter by actor", Required: false},
				{Name: "limit", Type: "number", Description: "Maximum number of tasks to display", Required: false},
				{Name: "offset", Type: "number", Description: "Number of tasks to skip", Required: false},
			},
		},
		{
			Name:          "list_done",
			CLISubcommand: "list",
			FixedStatus:   "done",
			Description:   "List completed tasks with optional milestone filter.",
			Args: []ToolArg{
				{Name: "milestone", Type: "string", Description: "Filter by milestone", Required: false},
				{Name: "actor", Type: "string", Description: "Filter by actor", Required: false},
				{Name: "limit", Type: "number", Description: "Maximum number of tasks to display", Required: false},
				{Name: "offset", Type: "number", Description: "Number of tasks to skip", Required: false},
			},
		},
		{
			Name:          "list_status_in_progress",
			CLISubcommand: "list",
			FixedStatus:   "in_progress",
			Description:   "List in-progress tasks with optional milestone filter.",
			Args: []ToolArg{
				{Name: "milestone", Type: "string", Description: "Filter by milestone", Required: false},
				{Name: "actor", Type: "string", Description: "Filter by actor", Required: false},
				{Name: "limit", Type: "number", Description: "Maximum number of tasks to display", Required: false},
				{Name: "offset", Type: "number", Description: "Number of tasks to skip", Required: false},
			},
		},
		{
			Name:          "list_status_todo",
			CLISubcommand: "list",
			FixedStatus:   "todo",
			Description:   "List todo tasks with optional milestone filter.",
			Args: []ToolArg{
				{Name: "milestone", Type: "string", Description: "Filter by milestone", Required: false},
				{Name: "actor", Type: "string", Description: "Filter by actor", Required: false},
				{Name: "limit", Type: "number", Description: "Maximum number of tasks to display", Required: false},
				{Name: "offset", Type: "number", Description: "Number of tasks to skip", Required: false},
			},
		},
		{
			Name:          "reset_timedout",
			CLISubcommand: "reset-timedout",
			Description:   "Reset timed out tasks to todo status. Finds in-progress tasks that have exceeded the specified timeout duration.",
			Args: []ToolArg{
				{Name: "minutes", Type: "number", Description: "Timeout duration in minutes (default: 30)", Required: false},
			},
		},
		{
			Name:          "start",
			CLISubcommand: "update",
			FixedStatus:   "in_progress",
			Description:   "Start working on a task by moving it to in-progress status.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID (required)", Required: true},
			},
		},
		{
			Name:          "unblock",
			CLISubcommand: "unblock",
			Description:   "Unblock a previously blocked task, transitioning it from blocked back to todo status. Optionally update the description.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "The unique identifier of the task to unblock", Required: true},
				{Name: "description", Type: "string", Description: "New description to overwrite the existing description", Required: false},
			},
		},
		{
			Name:        "wait",
			Description: "Wait for one or more tasks to complete. Blocks until all specified tasks reach 'done' status or the timeout is reached.",
			Args: []ToolArg{
				{Name: "task_ids", Type: "string_array", Description: "List of task IDs to wait on (required)", Required: true},
				{Name: "timeout", Type: "number", Description: "Timeout in milliseconds (0 = wait forever)", Required: false},
			},
		},
	}
}

// toolWrapperTmpl is the text/template used to render all tool wrappers.
// Template data is toolWrapperTmplData.
var toolWrapperTmpl = template.Must(template.New("tool-wrapper").Funcs(template.FuncMap{
	// zodSchema returns the tool.schema.* call for a plain ToolArg.
	"zodSchema": func(arg ToolArg) string {
		switch arg.Type {
		case "number":
			return fmt.Sprintf("tool.schema.number().describe(%q)", arg.Description)
		case "boolean":
			return fmt.Sprintf("tool.schema.boolean().describe(%q)", arg.Description)
		case "string_array":
			return fmt.Sprintf("tool.schema.array(tool.schema.string()).describe(%q)", arg.Description)
		default:
			return fmt.Sprintf("tool.schema.string().describe(%q)", arg.Description)
		}
	},
	// enumSchema returns the tool.schema.enum([...]).describe(...) call.
	"enumSchema": func(enum ToolEnum) string {
		vals := make([]string, len(enum.Values))
		for i, v := range enum.Values {
			vals[i] = fmt.Sprintf("%q", v.Value)
		}
		return fmt.Sprintf("tool.schema.enum([%s]).describe(%q)", strings.Join(vals, ", "), enum.Description)
	},
	// last returns true when i is the last index of a slice of length n.
	"last": func(i, n int) bool { return i == n-1 },
	// cliSub returns the CLI subcommand for a command, falling back to Name.
	"cliSub": func(cmd ToolCommand) string {
		if cmd.CLISubcommand != "" {
			return cmd.CLISubcommand
		}
		return cmd.Name
	},
}).Parse(`import { tool } from "@opencode-ai/plugin";
import { execFileSync } from "child_process";
{{- $bin := .BinaryPath}}
{{range $i, $cmd := .Commands}}
export const task_{{$cmd.Name}} = tool({
  description: {{printf "%q" $cmd.Description}},
  args: {
{{- range $cmd.Enums}}
    {{.Name}}: {{enumSchema .}},
{{- end}}
{{- range $i, $arg := $cmd.Args}}
    {{$arg.Name}}: {{zodSchema $arg}}{{if not (last $i (len $cmd.Args))}},{{end}}
{{- end}}
  },
  async execute(args, context) {
    const cmdArgs = [];
    cmdArgs.push("{{cliSub $cmd}}");
    const payload = {
{{- range $cmd.Args}}
      {{.Name}}: args.{{.Name}},
{{- end}}
{{- if $cmd.FixedStatus}}
      status: "{{$cmd.FixedStatus}}",
{{- end}}
    };
    cmdArgs.push(JSON.stringify(payload));

    const result = execFileSync(` + "`" + `{{$bin}}` + "`" + `, cmdArgs, {
      encoding: "utf-8", stdio: ["pipe", "pipe", "pipe"], cwd: context.worktree
    });

    return result;
  },
});
{{end}}`))

// toolWrapperTmplData is the data passed to toolWrapperTmpl.
type toolWrapperTmplData struct {
	BinaryPath string
	Commands   []ToolCommand
}

// GenerateToolWrapper generates a TypeScript tool wrapper using the tool() helper format.
func GenerateToolWrapper(opts *ToolWrapperOptions) (string, error) {
	if opts == nil {
		opts = DefaultToolWrapperOptions()
	}
	if opts.BinaryPath == "" {
		opts.BinaryPath = "taskflow"
	}

	var buf bytes.Buffer
	if err := toolWrapperTmpl.Execute(&buf, toolWrapperTmplData{
		BinaryPath: opts.BinaryPath,
		Commands:   getToolCommandsWithEnums(),
	}); err != nil {
		return "", fmt.Errorf("tool wrapper template: %w", err)
	}
	return buf.String(), nil
}
