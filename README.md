# TaskFlow - Task Management CLI Tool

A command-line interface tool for managing tasks with support for adding, updating, completing, blocking, deleting, listing, and resetting tasks. Built with Go using the Cobra framework and SQLite for persistent storage.

## Overview

TaskFlow is a CLI task management system designed for teams and individuals who prefer command-line workflows. It provides a robust set of commands for task lifecycle management including:

- **Add tasks** - Create new tasks with ID, title, description, milestone, and actor assignment
- **Update tasks** - Modify existing task properties
- **Complete tasks** - Mark tasks as done
- **Block tasks** - Block tasks with a reason for tracking dependencies
- **Delete tasks** - Soft delete tasks (moved to deleted_tasks table)
- **List tasks** - View all tasks with filtering, pagination, and sorting
- **Reset timed-out tasks** - Automatically reset tasks that have been in progress too long
- **Start tasks** - Move a task to in_progress status

## Installation

### Prerequisites

- Go 1.22.2 or later
- SQLite3 development libraries

### Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd taskflow

# Download dependencies
go mod download

# Build the binary
go build -o taskflow .
```

### Install

Installation is assumed to use standard go conventions:

```bash
go install .
```

### OpenCode 2.0 Plugin Setup

TaskFlow ships a generator for a complete [OpenCode](https://opencode.ai) 2.0 plugin. The plugin default-exports `Plugin.define` and registers the `taskflow_*` tools through `ctx.tool.transform`; each tool shells out to the taskflow CLI:

```bash
# Project-local plugin (auto-discovered by OpenCode 2.0)
taskflow tool-wrapper --output .opencode/plugins/taskflow.ts

# With a custom binary path
taskflow tool-wrapper --binary-path my-task --output .opencode/plugins/taskflow.ts

# Print to stdout instead
taskflow tool-wrapper
```

For a global install, save the generated file anywhere and add its path to the `"plugins"` array in `opencode.jsonc`. A per-project binary override is available through plugin options:

```json
"plugins": [{ "package": ".opencode/plugins/taskflow.ts", "options": { "binaryPath": "my-task" } }]
```

The `binaryPath` value must be an absolute path or a real executable on `PATH`; npm-style `.cmd`/`.bat` shims are not supported (the plugin spawns the binary directly).

Requirements and behavior:

- Requires OpenCode 2.0 and Node 18+.
- The plugin registers 12 tools namespaced `taskflow_*` (see the tool table under [Usage](#usage)).
- Each tool spawns taskflow in the session's working directory (worktree-aware), so taskflow finds the database by walking up from that directory for the `.taskflow` anchor.
- Tool calls are canceled when the session stops.
- After upgrading taskflow, rerun `taskflow tool-wrapper` to regenerate the plugin. The plugin refuses a taskflow with a different major version (or a different minor version while the major is 0), a missing binary, or unparseable version output, and the error says to regenerate the plugin.

### Upgrading from the 0.19-era tool wrapper

OpenCode 2.0 does not read the old tool wrapper file (for example `~/.config/opencode/tools/taskflow.ts`); the file can be deleted. The plugin registers the same operations under namespaced tool names (`taskflow_*`) and requires OpenCode 2.0.

### Database Setup

The application automatically creates the SQLite database at `.taskflow/tasks.db` on first run. The database schema includes:

- `tasks` table with columns: id, milestone, sprint, title, description, status, actor, blocked_by, created, last_updated
- `deleted_tasks` table with an additional `deleted_on` column for soft-deleted tasks
- Indexes on: milestone, status, sprint, deleted_on

## Usage

### Global Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--config` | - | Config file path | `./config.yaml` |
| `--verbose` | `-v` | Enable verbose output | `false` |
| `--version` | - | Display version info | - |

### Command: add

Add a new task to the task list.

```bash
taskflow add '{"id":"1","title":"Implement login feature","milestone":"v1.0","description":"Add login functionality"}'
taskflow add '{"id":"2","title":"Fix memory leak","milestone":"v2.0","description":"Memory leak in data processing"}'
taskflow add '{"id":"3","title":"Deploy to production","milestone":"v2.0 Release","actor":"devops"}'

# With JSON flag
taskflow add -j '{"id":"1","title":"Implement login","milestone":"v1","description":"Add login"}'

# From stdin
echo '{"id":"2","title":"Fix bug","milestone":"v1","description":"Fix memory leak"}' | taskflow add -
```

#### JSON Fields

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Unique task identifier |
| `title` | Yes | Task title |
| `description` | Yes | Task description |
| `milestone` | Yes | Milestone/sprint name |
| `actor` | No | Assigned actor/owner |

---

### Command: update

Update an existing task by its ID.

```bash
taskflow update '{"id":"1","title":"New title"}'
taskflow update '{"id":"abc","description":"Updated description","milestone":"v2.0"}'
taskflow update '{"id":"1","status":"in_progress","actor":"new-owner"}'
taskflow update '{"id":"1","milestone":"v2.0","description":"New description"}'

# With JSON flag
taskflow update -j '{"id":"1","status":"in_progress"}'
```

#### JSON Fields

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Task ID to update |
| `title` | No | New task title |
| `description` | No | New task description |
| `status` | No | New status (todo, in_progress, done, blocked) |
| `milestone` | No | New milestone |
| `actor` | No | New assigned actor |

**Note:** At least one of title, description, status, milestone, or actor must be provided.

---

### Command: complete

Mark a task as completed.

```bash
taskflow complete '{"id":"1"}'
taskflow complete '{"id":"abc123"}'

# With JSON flag
taskflow complete -j '{"id":"1"}'
```

#### JSON Fields

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Task ID to complete |

---

### Command: block

Block a task with a reason. Blocked tasks cannot be worked on until unblocked.

```bash
taskflow block '{"id":"1","reason":"Waiting for API documentation"}'
taskflow block '{"id":"abc123","reason":"Dependency not available"}'

# With JSON flag
taskflow block -j '{"id":"1","reason":"Waiting for API documentation"}'
```

#### JSON Fields

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Task ID to block |
| `reason` | Yes | Reason for blocking |

---

### Command: unblock

Unblocks a blocked task, transitioning it from 'blocked' to 'todo' status, with optional description overwrite.

```bash
taskflow unblock '{"id":"1"}'
taskflow unblock '{"id":"abc123"}'

# With description update
taskflow unblock '{"id":"1","description":"API docs now available, proceeding with implementation"}'

# With JSON flag
taskflow unblock -j '{"id":"1"}'
```

#### JSON Fields

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Task ID to unblock |
| `description` | No | New description to overwrite existing description |

---

### Command: delete

Soft delete a task by moving it to the deleted_tasks table.

```bash
taskflow delete '{"id":"1"}'
taskflow delete '{"id":"abc123"}'

# With JSON flag
taskflow delete -j '{"id":"1"}'
```

#### JSON Fields

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Task ID to delete |

---

### Command: list

List all tasks with optional filtering.

```bash
# List all tasks
taskflow list

# Filter by milestone
taskflow list '{"milestone":"sprint-1"}'

# Filter by status and actor
taskflow list '{"status":"todo","actor":"john"}'

# Sort by field
taskflow list '{"sort_by":"status"}'
taskflow list '{"sort_by":"milestone"}'
taskflow list '{"sort_by":"created"}'
taskflow list '{"sort_by":"updated"}'
taskflow list '{"sort_by":"id"}'
taskflow list '{"sort_by":"title"}'
taskflow list '{"sort_by":"description"}'
taskflow list '{"sort_by":"actor"}'

# Pagination
taskflow list '{"limit":10,"offset":0}'

# Get specific task by ID
taskflow list '{"id":"task-123"}'

# Combined example
taskflow list '{"milestone":"v1.0","status":"in_progress","limit":50}'
```

#### JSON Fields

| Field | Default | Description |
|-------|---------|-------------|
| `milestone` | - | Filter by milestone |
| `status` | - | Filter by status (todo, in_progress, done, blocked) |
| `actor` | - | Filter by actor |
| `id` | - | Get specific task by ID |
| `sort_by` | - | Sort by field (status, milestone, created, updated, id, title, description, actor) |
| `limit` | `20` | Maximum tasks to display |
| `offset` | `0` | Number of tasks to skip |

---

### Command: reset-timedout

Reset in-progress tasks that have exceeded the specified timeout duration back to todo status.

```bash
# Reset tasks in progress longer than 30 minutes
taskflow reset-timedout '{"minutes":30}'

# Reset tasks in progress longer than 60 minutes
taskflow reset-timedout '{"minutes":60}'

# With JSON flag
taskflow reset-timedout -j '{"minutes":30}'
```

#### JSON Fields

| Field | Default | Description |
|-------|---------|-------------|
| `minutes` | `30` | Timeout duration in minutes |

---

### OpenCode 2.0 Plugin Tools

The `tool-wrapper` command generates a complete OpenCode 2.0 plugin (a `Plugin.define` module that registers the tools below through `ctx.tool.transform`). The plugin spawns taskflow in the session's working directory; taskflow finds the database by walking up from that directory for the `.taskflow` anchor. The plugin embeds the taskflow version it was generated with. It refuses to run when the installed taskflow has a different major version, or a different minor version while the major is 0, when the binary is missing, or when the version output is not parseable. The error tells the agent to regenerate the plugin. The plugin exposes these tools that map to taskflow CLI calls:

| Plugin Tool | Maps To | Description |
|-------------|---------|-------------|
| `taskflow_add` | `taskflow add '{"id":"...","title":"...","description":"...","milestone":"..."}'` | Add a new task |
| `taskflow_block` | `taskflow block '{"id":"...","reason":"..."}'` | Block a task with a reason |
| `taskflow_complete` | `taskflow complete '{"id":"..."}'` | Mark a task as done |
| `taskflow_delete` | `taskflow delete '{"id":"..."}'` | Soft delete a task |
| `taskflow_list_all` | `taskflow list` with filters | List all tasks |
| `taskflow_list_blocked` | `taskflow list '{"status":"blocked"}'` | List blocked tasks |
| `taskflow_list_done` | `taskflow list '{"status":"done"}'` | List completed tasks |
| `taskflow_list_status_in_progress` | `taskflow list '{"status":"in_progress"}'` | List in-progress tasks |
| `taskflow_list_status_todo` | `taskflow list '{"status":"todo"}'` | List todo tasks |
| `taskflow_reset_timedout` | `taskflow reset-timedout '{"minutes":30}'` | Reset timed-out tasks to todo |
| `taskflow_start` | `taskflow update '{"id":"...","status":"in_progress"}'` | Start working on a task |
| `taskflow_unblock` | `taskflow unblock '{"id":"..."}'` | Unblock a blocked task |

### Starting a Task

To start working on a task via the CLI, use the update command:

```bash
taskflow update '{"id":"1","status":"in_progress"}'
```

---

## Architecture

```
taskflow/
├── main.go                 # Application entry point
├── install.sh              # One-line install script
├── project.md              # Project brief
├── LICENSE.md              # License file
├── cmd/
│   ├── root.go            # Root command and global flags
│   ├── add.go             # Add task command
│   ├── update.go          # Update task command
│   ├── complete.go        # Complete task command
│   ├── block.go           # Block task command
│   ├── delete.go          # Delete task command
│   ├── list.go            # List tasks command
│   ├── reset.go           # Reset timed-out tasks command
│   ├── tool-wrapper.go    # Generate OpenCode 2.0 plugin
│   └── *_test.go          # Unit tests for commands
├── internal/
│   ├── db/
│   │   ├── db.go          # Database connection and initialization
│   │   ├── operations.go  # CRUD operations
│   │   ├── errors.go      # Database error types
│   │   └── schema.sql     # Database schema definition
│   ├── service/
│   │   ├── add.go         # Add task service logic
│   │   ├── update.go      # Update task service logic
│   │   ├── complete.go    # Complete task service logic
│   │   ├── block.go       # Block task service logic
│   │   ├── delete.go      # Delete task service logic
│   │   ├── list.go        # List tasks service logic
│   │   ├── reset.go       # Reset timed-out service logic
│   │   ├── timeout.go     # Timeout handling utilities
│   │   ├── json.go        # JSON parsing utilities
│   │   ├── errors.go      # Service error types
│   │   └── *_test.go      # Service unit tests
│   └── README.md          # Internal package documentation
├── scripts/
│   ├── test-add.sh        # Test adding tasks
│   ├── test-update.sh     # Test updating tasks
│   ├── test-complete.sh   # Test completing tasks
│   ├── test-block.sh      # Test blocking tasks
│   ├── test-list.sh       # Test listing tasks
│   ├── test-reset.sh      # Test resetting timed-out tasks
│   ├── run-all-tests.sh   # Run all integration tests
│   └── examples/
│       ├── sprint-management.sh
│       ├── task-workflow.sh
│       └── project-setup.sh
├── tests/
│   └── integration/
│       ├── integration_test.go  # Integration tests
│       └── setup_test.go        # Test database setup
└── .taskflow/
    └── tasks.db           # SQLite database
```

### Component Responsibilities

- **cmd/**: Cobra command implementations - parse flags, validate inputs, call services
- **internal/db/**: Database layer - connection management, queries, schema
- **internal/service/**: Business logic - task operations, validation, state transitions
- **internal/clierr/**: Error handling and input validation - CLI error types, ID/status/field validation
- **internal/generator/**: Code generation - OpenCode 2.0 plugin for taskflow tools

### Database Schema

```sql
CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    milestone TEXT,
    sprint TEXT,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'todo',
    actor TEXT,
    blocked_by TEXT,
    created TEXT NOT NULL,
    last_updated TEXT NOT NULL
);

CREATE INDEX idx_tasks_milestone ON tasks(milestone);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_sprint ON tasks(sprint);

CREATE TABLE deleted_tasks (
    id TEXT PRIMARY KEY,
    milestone TEXT,
    sprint TEXT,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL,
    actor TEXT,
    blocked_by TEXT,
    created TEXT NOT NULL,
    last_updated TEXT NOT NULL,
    deleted_on TEXT NOT NULL
);

CREATE INDEX idx_deleted_tasks_deleted_on ON deleted_tasks(deleted_on);
```

---

## Sprint Management Integration

TaskFlow integrates with the project's sprint management system through shell scripts in the `scripts/` directory:

| Script | Purpose |
|--------|---------|
| `test-add.sh` | Test adding tasks |
| `test-update.sh` | Test updating tasks |
| `test-complete.sh` | Test completing tasks |
| `test-block.sh` | Test blocking tasks |
| `test-list.sh` | Test listing tasks |
| `test-reset.sh` | Test resetting timed-out tasks |
| `run-all-tests.sh` | Run all integration tests |
| `examples/sprint-management.sh` | Sprint management examples |
| `examples/task-workflow.sh` | Task workflow examples |
| `examples/project-setup.sh` | Project setup examples |

### Integration Usage

```bash
# Run all tests
./scripts/run-all-tests.sh

# Test individual commands
./scripts/test-add.sh
./scripts/test-list.sh --milestone "sprint-1"
```

---

## Status Values

Tasks can have one of four statuses:

| Status | Description |
|--------|-------------|
| `todo` | Task created, not yet started |
| `in_progress` | Task is actively being worked on |
| `done` | Task completed |
| `blocked` | Task blocked by external dependency |

### Status Aliases

The CLI accepts aliases that are normalized to canonical values:

| Alias | Canonical |
|-------|-----------|
| `pending` | `todo` |
| `in-progress` | `in_progress` |
| `inprogress` | `in_progress` |
| `completed` | `done` |
| `timed-out` | `blocked` |
| `timedout` | `blocked` |

### Status Transitions

```
todo -> in_progress (update --status in_progress)
todo -> done (complete --id <id>)
in_progress -> done (complete --id <id>)
in_progress -> blocked (block --id <id> --reason "...")
blocked -> todo (update --status todo)
done -> todo (update --status todo)
```

---

## Error Handling

The CLI provides descriptive error messages for common issues:

- **Missing required fields**: "Missing required argument: <field>\nUsage: <usage>"
- **Invalid ID format**: "Invalid value for task-id: cannot be empty"
- **Invalid status**: "Invalid value for status: '<value>' is not valid"
- **Database errors**: "failed to connect to database", "task not found"
- **Validation errors**: Include error codes (INVALID_ARGUMENT, MISSING_ARGUMENT, INVALID_FORMAT, RESOURCE_NOT_FOUND, RESOURCE_EXISTS, PERMISSION_DENIED, DATABASE_ERROR, FILE_NOT_FOUND, CONFIGURATION_ERROR, UNEXPECTED_ERROR)

---

## Configuration

### Verbose Mode

Enable verbose output for debugging:

```bash
taskflow --verbose add '{"id":"1","title":"Test","description":"Test task","milestone":"v1.0"}'
```

### Config File

Specify a custom configuration file:

```bash
taskflow --config /path/to/config.yaml <command>
```

---

## Version Information

Run `taskflow version` for the current version.

---

## Examples

### Complete Workflow

```bash
# 1. Add tasks to a sprint
taskflow add '{"id":"TASK-1","title":"Design database schema","description":"Create ERD and schema definitions","milestone":"sprint-1","actor":"alice"}'
taskflow add '{"id":"TASK-2","title":"Implement API endpoints","description":"REST API for task CRUD","milestone":"sprint-1","actor":"bob"}'
taskflow add '{"id":"TASK-3","title":"Write unit tests","description":"Test coverage for API","milestone":"sprint-1","actor":"alice"}'

# 2. Start working on a task
taskflow update '{"id":"TASK-1","status":"in_progress"}'

# 3. Block a task waiting on dependency
taskflow block '{"id":"TASK-3","reason":"Waiting for API to be ready"}'

# 4. Check progress
taskflow list '{"milestone":"sprint-1"}'

# 5. Complete tasks
taskflow complete '{"id":"TASK-1"}'
taskflow complete '{"id":"TASK-2"}'

# 6. View completed tasks
taskflow list '{"status":"done","milestone":"sprint-1"}'

# 7. Check for timed-out tasks
taskflow reset-timedout '{"minutes":30}'
```

---

## Development

### Running Tests

```bash
# All tests
go test ./...

# Unit tests for commands
go test ./cmd/...

# Unit tests for internal packages
go test ./internal/...

# Integration tests
go test ./tests/integration/...

# Standalone CLI e2e (no Go test deps; builds the binary and runs a full lifecycle)
bash scripts/e2e.sh
```

### Adding New Commands

1. Create `cmd/newcommand.go` with Cobra command definition
2. Create `internal/service/newcommand.go` with business logic (if needed)
3. Register the command in `cmd/root.go` init()
4. Add corresponding test files

### Code Linting

```bash
# Run go vet
go vet ./...

# Run go fmt
gofmt -l .
```

---

## License

This project is part of the task management system. All rights reserved.
