package generator

import (
	"bytes"
	"fmt"
	"text/template"
)

// ToolWrapperOptions contains options for tool wrapper generation
type ToolWrapperOptions struct {
	BinaryPath string
	// Version is the taskflow version the wrapper was generated with.
	// Empty means unknown ("0.0.0"), which disables the wrapper's
	// version-compatibility check.
	Version string
}

func DefaultToolWrapperOptions() *ToolWrapperOptions {
	return &ToolWrapperOptions{
		BinaryPath: "taskflow",
		Version:    "0.0.0",
	}
}

// ToolArg represents a tool argument definition
type ToolArg struct {
	Name        string
	Type        string
	Description string
}

// ToolCommand represents a taskflow command definition
type ToolCommand struct {
	Name        string
	Description string
	Args        []ToolArg
	// CLISubcommand is the subcommand passed to the binary (e.g. "list", "update").
	// Defaults to Name (with underscores replaced by hyphens) if empty.
	CLISubcommand string
	// FixedStatus, when non-empty, is serialised as a literal "status" field in
	// the JSON payload sent to the binary.
	FixedStatus string
}

// listFilterArgs holds the shared argument set for the five list commands
// (list_all, list_blocked, list_done, list_status_in_progress,
// list_status_todo). It is read-only: consumers (the wrapper template and
// tests) never append to or mutate it, so sharing one slice is safe.
var listFilterArgs = []ToolArg{
	{Name: "milestone", Type: "string", Description: "Filter by milestone"},
	{Name: "actor", Type: "string", Description: "Filter by actor"},
	{Name: "limit", Type: "number", Description: "Maximum number of tasks to display"},
	{Name: "offset", Type: "number", Description: "Number of tasks to skip"},
}

func getToolCommands() []ToolCommand {
	return []ToolCommand{
		{
			Name:        "add",
			Description: "Add a new task to the task list. Requires id, milestone, title, and description.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID (required)"},
				{Name: "title", Type: "string", Description: "Task title (required)"},
				{Name: "description", Type: "string", Description: "Task description (required)"},
				{Name: "milestone", Type: "string", Description: "Milestone for the task (required)"},
				{Name: "actor", Type: "string", Description: "Actor assigned to the task"},
			},
		},
		{
			Name:        "block",
			Description: "Block a task by providing its ID and a reason.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID (required)"},
				{Name: "reason", Type: "string", Description: "Reason for blocking the task (required)"},
			},
		},
		{
			Name:        "complete",
			Description: "Mark a task as completed by providing its ID.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID (required)"},
			},
		},
		{
			Name:        "delete",
			Description: "Soft delete a task by moving it to the deleted_tasks table.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID (required)"},
			},
		},
		{
			Name:          "list_all",
			CLISubcommand: "list",
			Description:   "List all tasks with optional milestone filter.",
			Args:          listFilterArgs,
		},
		{
			Name:          "list_blocked",
			CLISubcommand: "list",
			FixedStatus:   "blocked",
			Description:   "List blocked tasks with optional milestone filter.",
			Args:          listFilterArgs,
		},
		{
			Name:          "list_done",
			CLISubcommand: "list",
			FixedStatus:   "done",
			Description:   "List completed tasks with optional milestone filter.",
			Args:          listFilterArgs,
		},
		{
			Name:          "list_status_in_progress",
			CLISubcommand: "list",
			FixedStatus:   "in_progress",
			Description:   "List in-progress tasks with optional milestone filter.",
			Args:          listFilterArgs,
		},
		{
			Name:          "list_status_todo",
			CLISubcommand: "list",
			FixedStatus:   "todo",
			Description:   "List todo tasks with optional milestone filter.",
			Args:          listFilterArgs,
		},
		{
			Name:          "reset_timedout",
			CLISubcommand: "reset-timedout",
			Description:   "Reset timed out tasks to todo status. Finds in-progress tasks that have exceeded the specified timeout duration.",
			Args: []ToolArg{
				{Name: "minutes", Type: "number", Description: "Timeout duration in minutes (default: 30)"},
			},
		},
		{
			Name:          "start",
			CLISubcommand: "update",
			FixedStatus:   "in_progress",
			Description:   "Start working on a task by moving it to in-progress status.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID (required)"},
			},
		},
		{
			Name:          "unblock",
			CLISubcommand: "unblock",
			Description:   "Unblock a previously blocked task, transitioning it from blocked back to todo status. Optionally update the description.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "The unique identifier of the task to unblock"},
				{Name: "description", Type: "string", Description: "New description to overwrite the existing description"},
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
		default:
			return fmt.Sprintf("tool.schema.string().describe(%q)", arg.Description)
		}
	},
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

const TASKFLOW_WRAPPER_VERSION = "{{.Version}}";

let runtimeVersion = null;

function parseSemver(text) {
  const m = String(text).trim().match(/(\d+)\.(\d+)\.(\d+)/);
  return m ? { major: Number(m[1]), minor: Number(m[2]), patch: Number(m[3]) } : null;
}

function compatible(want, got) {
  if (!want || !got) return false;
  if (want.major === 0 && want.minor === 0 && want.patch === 0) return true;
  if (want.major !== got.major) return false;
  if (want.major === 0 && want.minor !== got.minor) return false;
  return true;
}

function checkVersion(bin) {
  if (runtimeVersion !== null) return;
  let out;
  try {
    out = execFileSync(bin, ["version"], { encoding: "utf-8", stdio: ["pipe", "pipe", "pipe"] });
  } catch (err) {
    throw new Error(` + "`" + `taskflow binary "${bin}" is missing, not on PATH, or too old to report a version (${String(err.message).split("\n")[0]}). Upgrade taskflow and run: taskflow tool-wrapper --output ~/.config/opencode/tools/taskflow.ts` + "`" + `);
  }
  const got = parseSemver(out);
  if (!got) {
    throw new Error(` + "`" + `taskflow version output is not a semver: ${String(out).trim().split("\n")[0]}. Upgrade taskflow and run: taskflow tool-wrapper --output ~/.config/opencode/tools/taskflow.ts` + "`" + `);
  }
  if (!compatible(parseSemver(TASKFLOW_WRAPPER_VERSION), got)) {
    throw new Error(` + "`" + `taskflow wrapper was generated for taskflow ${TASKFLOW_WRAPPER_VERSION} but the installed taskflow reports ${got.major}.${got.minor}.${got.patch}; rebuild taskflow and regenerate the wrapper: taskflow tool-wrapper --output ~/.config/opencode/tools/taskflow.ts` + "`" + `);
  }
  runtimeVersion = got;
}
{{range $cmd := .Commands}}
export const task_{{$cmd.Name}} = tool({
  description: {{printf "%q" $cmd.Description}},
  args: {
{{- range $arg := $cmd.Args}}
    {{$arg.Name}}: {{zodSchema $arg}},
{{- end}}
  },
  async execute(args, context) {
    checkVersion({{printf "%q" $bin}});
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

    const result = execFileSync(` + "`" + `{{$bin}}` + "`" + `, cmdArgs, { encoding: "utf-8", stdio: ["pipe", "pipe", "pipe"], cwd: context.directory });

    return result;
  },
});
{{end}}`))

// toolWrapperTmplData is the data passed to toolWrapperTmpl.
type toolWrapperTmplData struct {
	BinaryPath string
	Version    string
	Commands   []ToolCommand
}

// GenerateToolWrapper generates a TypeScript tool wrapper using the tool() helper format.
// An empty Version in opts is normalized to "0.0.0". The generated wrapper treats an
// embedded 0.0.0 as unknown: the semver compatibility rule is disabled, but the wrapper
// still probes the binary with `version` and refuses a missing binary.
func GenerateToolWrapper(opts *ToolWrapperOptions) (string, error) {
	if opts == nil {
		opts = DefaultToolWrapperOptions()
	}
	if opts.BinaryPath == "" {
		opts.BinaryPath = "taskflow"
	}
	if opts.Version == "" {
		opts.Version = "0.0.0"
	}

	var buf bytes.Buffer
	if err := toolWrapperTmpl.Execute(&buf, toolWrapperTmplData{
		BinaryPath: opts.BinaryPath,
		Version:    opts.Version,
		Commands:   getToolCommands(),
	}); err != nil {
		return "", fmt.Errorf("tool wrapper template: %w", err)
	}
	return buf.String(), nil
}
