package generator

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"
)

// ToolWrapperOptions contains options for tool wrapper generation
type ToolWrapperOptions struct {
	BinaryPath string
	// Version is the taskflow version the plugin was generated with.
	// Empty means unknown ("0.0.0"), which disables the plugin's
	// version-compatibility check (the binary is still probed).
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
	// Required marks the argument as mandatory. Required arguments are
	// collected into the generated JSON Schema "required" array; the
	// description text no longer carries a "(required)" marker.
	Required bool
	// Enum, when non-empty, restricts the argument to exactly these string
	// values and is rendered as a JSON Schema "enum".
	Enum []string
	// Minimum holds the literal JSON Schema minimum number text (for example
	// "0" or "1") for number arguments; empty means no minimum is emitted.
	Minimum string
}

// ToolCommand represents a taskflow command definition
type ToolCommand struct {
	Name        string
	Description string
	Args        []ToolArg
	// CLISubcommand is the subcommand passed to the binary (e.g. "list", "update").
	// Defaults to Name if empty.
	CLISubcommand string
	// FixedStatus, when non-empty, is serialised as a literal "status" field in
	// the JSON payload sent to the binary.
	FixedStatus string
}

// validSortKeys are the sort keys advertised by the list tools' sort_by
// argument. It must match db.ValidSortKeys (internal/db/operations.go); the
// list is duplicated here so the generator stays dependency-free.
var validSortKeys = []string{
	"status",
	"milestone",
	"created",
	"updated",
	"id",
	"title",
	"description",
	"actor",
}

// listFilterArgs holds the shared argument set for the five list commands
// (list_all, list_blocked, list_done, list_status_in_progress,
// list_status_todo). It is read-only: consumers (the wrapper template and
// tests) never append to or mutate it, so sharing one slice is safe.
var listFilterArgs = []ToolArg{
	{Name: "milestone", Type: "string", Description: "Filter by milestone"},
	{Name: "actor", Type: "string", Description: "Filter by actor"},
	{Name: "sort_by", Type: "string", Description: "Sort key: one of status, milestone, created, updated, id, title, description, actor", Enum: validSortKeys},
	{Name: "limit", Type: "number", Description: "Maximum number of tasks to display", Minimum: "0"},
	{Name: "offset", Type: "number", Description: "Number of tasks to skip", Minimum: "0"},
}

func getToolCommands() []ToolCommand {
	return []ToolCommand{
		{
			// milestone is Required here even though the CLI itself accepts an
			// empty milestone: the tool asks for an explicit milestone (V1
			// wrapper parity).
			Name:        "add",
			Description: "Add a new task to the task list. Requires id, milestone, title, and description.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID", Required: true},
				{Name: "title", Type: "string", Description: "Task title", Required: true},
				{Name: "description", Type: "string", Description: "Task description", Required: true},
				{Name: "milestone", Type: "string", Description: "Milestone for the task", Required: true},
				{Name: "actor", Type: "string", Description: "Actor assigned to the task"},
			},
		},
		{
			Name:        "block",
			Description: "Block a task by providing its ID and a reason.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID", Required: true},
				{Name: "reason", Type: "string", Description: "Reason for blocking the task", Required: true},
			},
		},
		{
			Name:        "complete",
			Description: "Mark a task as completed by providing its ID.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID", Required: true},
			},
		},
		{
			Name:        "delete",
			Description: "Soft delete a task by moving it to the deleted_tasks table.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID", Required: true},
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
				{Name: "minutes", Type: "number", Description: "Timeout duration in minutes (default: 30)", Minimum: "1"},
			},
		},
		{
			Name:          "start",
			CLISubcommand: "update",
			FixedStatus:   "in_progress",
			Description:   "Start working on a task by moving it to in-progress status.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "Task ID", Required: true},
			},
		},
		{
			Name:          "unblock",
			CLISubcommand: "unblock",
			Description:   "Unblock a previously blocked task, transitioning it from blocked back to todo status. Optionally update the description.",
			Args: []ToolArg{
				{Name: "id", Type: "string", Description: "The unique identifier of the task to unblock", Required: true},
				{Name: "description", Type: "string", Description: "New description to overwrite the existing description"},
			},
		},
	}
}

// toolWrapperTmpl is the text/template used to render the OpenCode 2.0
// plugin. Template data is toolWrapperTmplData.
var toolWrapperTmpl = template.Must(template.New("tool-wrapper").Funcs(template.FuncMap{
	// jsonProp returns the JSON Schema property object text for a ToolArg:
	// a string or number schema, or a string schema with an enum when the
	// argument carries one.
	"jsonProp": func(arg ToolArg) string {
		desc := fmt.Sprintf("description: %q", arg.Description)
		switch {
		case len(arg.Enum) > 0:
			quoted := make([]string, len(arg.Enum))
			for i, value := range arg.Enum {
				quoted[i] = strconv.Quote(value)
			}
			return fmt.Sprintf(`{ type: "string", enum: [%s], %s }`, strings.Join(quoted, ", "), desc)
		case arg.Type == "number":
			if arg.Minimum != "" {
				return fmt.Sprintf(`{ type: "number", minimum: %s, %s }`, arg.Minimum, desc)
			}
			return fmt.Sprintf(`{ type: "number", %s }`, desc)
		default:
			return fmt.Sprintf(`{ type: "string", %s }`, desc)
		}
	},
	// requiredLine returns the JSON Schema "required" array line for a
	// command's arguments, or the empty string when no argument is required.
	"requiredLine": func(cmd ToolCommand) string {
		var names []string
		for _, arg := range cmd.Args {
			if arg.Required {
				names = append(names, strconv.Quote(arg.Name))
			}
		}
		if len(names) == 0 {
			return ""
		}
		return "required: [" + strings.Join(names, ", ") + "],"
	},
	// cliSub returns the CLI subcommand for a command, falling back to Name.
	"cliSub": func(cmd ToolCommand) string {
		if cmd.CLISubcommand != "" {
			return cmd.CLISubcommand
		}
		return cmd.Name
	},
}).Parse(`// Code generated by taskflow tool-wrapper for OpenCode 2.0. DO NOT EDIT.
//
// This file is a complete OpenCode 2.0 plugin: it default-exports
// Plugin.define and registers the taskflow_* tools through ctx.tool.transform.
// Every tool shells out to the taskflow CLI in the session's working
// directory, so taskflow finds the project's task database by walking up from
// there for the .taskflow anchor.
//
// Project-local install (auto-discovered by OpenCode 2.0):
//   taskflow tool-wrapper --output .opencode/plugins/taskflow.ts
//
// For a plugin stored outside the project, save the file anywhere and add its
// path to the "plugins" array in opencode.jsonc (a "binaryPath" plugin option
// overrides the taskflow binary per project).

import { Plugin } from "@opencode/plugin";
import { execFile } from "node:child_process";

const TASKFLOW_PLUGIN_VERSION = {{printf "%q" .Version}};
const DEFAULT_BINARY = {{printf "%q" .BinaryPath}};
const NAMESPACE = "taskflow";

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

function runBinary(bin, args, options) {
  return new Promise((resolve, reject) => {
    execFile(bin, args, {
      encoding: "utf-8",
      windowsHide: true,
      maxBuffer: 16 * 1024 * 1024,
      cwd: options.cwd,
      signal: options.signal,
    }, (err, stdout, stderr) => {
      if (err) {
        if (options.signal?.aborted || err.signal || err.name === "AbortError") {
          reject(new Error(` + "`" + `taskflow ${args[0]} was canceled` + "`" + `));
          return;
        }
        const detail = String(stderr || stdout || err.message).trim();
        reject(new Error(` + "`" + `taskflow ${args[0]} failed: ${detail}` + "`" + `));
        return;
      }
      resolve(stdout);
    });
  });
}

async function checkVersion(bin) {
  if (runtimeVersion !== null) return;
  let out;
  try {
    out = await runBinary(bin, ["version"], {});
  } catch (err) {
    throw new Error(` + "`" + `taskflow binary "${bin}" is missing, not on PATH, or too old to report a version (${String(err.message).split("\n")[0]}). Upgrade taskflow and regenerate the plugin: taskflow tool-wrapper --output .opencode/plugins/taskflow.ts` + "`" + `);
  }
  const got = parseSemver(out);
  if (!got) {
    throw new Error(` + "`" + `taskflow version output is not a semver: ${String(out).trim().split("\n")[0]}. Upgrade taskflow and regenerate the plugin: taskflow tool-wrapper --output .opencode/plugins/taskflow.ts` + "`" + `);
  }
  if (!compatible(parseSemver(TASKFLOW_PLUGIN_VERSION), got)) {
    throw new Error(` + "`" + `taskflow plugin was generated for taskflow ${TASKFLOW_PLUGIN_VERSION} but the installed taskflow reports ${got.major}.${got.minor}.${got.patch}; rebuild taskflow and regenerate the plugin: taskflow tool-wrapper --output .opencode/plugins/taskflow.ts` + "`" + `);
  }
  runtimeVersion = got;
}

function makeDirectoryResolver(ctx, fallback) {
  // Session directories are cached for the process lifetime; a directory
  // move mid-session is not observed.
  const cache = new Map();
  return async (sessionID) => {
    if (!sessionID) return fallback;
    if (cache.has(sessionID)) return cache.get(sessionID);
    let dir = fallback;
    try {
      const session = await ctx.session.get({ sessionID });
      const sessionDir = session && session.location ? session.location.directory : undefined;
      if (typeof sessionDir === "string" && sessionDir !== "") {
        dir = sessionDir;
      }
    } catch {
      // Keep the plugin directory when the session lookup fails.
    }
    cache.set(sessionID, dir);
    return dir;
  };
}

export default Plugin.define({
  id: "taskflow",
  async setup(ctx) {
    const optionBinary = ctx.options ? ctx.options.binaryPath : undefined;
    const binary = typeof optionBinary === "string" && optionBinary.trim() !== "" ? optionBinary.trim() : DEFAULT_BINARY;
    const pluginDirectory = ctx.location && typeof ctx.location.directory === "string" ? ctx.location.directory : process.cwd();
    const directoryFor = makeDirectoryResolver(ctx, pluginDirectory);
    await ctx.tool.transform((editor) => {
      editor.namespace({ name: NAMESPACE, description: "Taskflow task management backed by the taskflow CLI: add, start, complete, block, unblock, delete, and list tasks stored in the project's task database." });
{{- range $cmd := .Commands}}
      editor.add({
        name: {{printf "%q" $cmd.Name}},
        description: {{printf "%q" $cmd.Description}},
        input: {
          "type": "object",
          properties: {
{{- range $arg := $cmd.Args}}
            {{$arg.Name}}: {{jsonProp $arg}},
{{- end}}
          },
{{- with requiredLine $cmd}}
          {{.}}
{{- end}}
          additionalProperties: false,
        },
        options: { namespace: NAMESPACE },
        async execute(args, context) {
          await checkVersion(binary);
          const payload = {
{{- range $arg := $cmd.Args}}
            {{$arg.Name}}: args.{{$arg.Name}},
{{- end}}
{{- if $cmd.FixedStatus}}
            status: "{{$cmd.FixedStatus}}",
{{- end}}
          };
          const result = await runBinary(binary, ["{{cliSub $cmd}}", JSON.stringify(payload)], { cwd: await directoryFor(context.sessionID), signal: context.signal });
          return { content: result };
        },
      });
{{- end}}
    });
  },
});`))

// toolWrapperTmplData is the data passed to toolWrapperTmpl.
type toolWrapperTmplData struct {
	BinaryPath string
	Version    string
	Commands   []ToolCommand
}

// GenerateToolWrapper generates a complete OpenCode 2.0 plugin: a file that
// default-exports Plugin.define and registers the taskflow_* tools through
// ctx.tool.transform. An empty Version in opts is normalized to "0.0.0". The
// generated plugin treats an embedded 0.0.0 as unknown: the semver
// compatibility rule is disabled, but the plugin still probes the binary with
// `version` and refuses a missing binary.
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
