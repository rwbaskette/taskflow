# TaskFlow - Task Management CLI Tool

A command-line interface tool for managing tasks with support for adding, updating, completing, blocking, deleting, listing, and resetting tasks. Built with Go using the Cobra framework and SQLite for persistent storage.

## Overview

TaskFlow is a CLI task management system designed for teams and individuals who prefer command-line workflows. It provides a robust set of commands for task lifecycle management including:

- **Add tasks** - Create new tasks with ID, title, description, milestone, and actor assignment
- **Update tasks** - Modify existing task properties
- **Complete tasks** - Mark tasks as done
- **Block tasks** - Block tasks with a reason for tracking dependencies
- **Delete tasks** - Soft delete tasks (moved to deleted_tasks table)
- **List tasks** - View all tasks with filtering, pagination, sorting, and multiple output formats
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

### Install to PATH

To make `taskflow` available system-wide:

```bash
# Install the binary to a directory in your PATH
# Common locations:
#   ~/.local/bin/       (user-local binaries)
#   /usr/local/bin/     (system-wide binaries, requires sudo)

# Example: Install to user-local bin
install -Dm755 taskflow ~/.local/bin/taskflow

# Or for system-wide installation:
sudo install -Dm755 taskflow /usr/local/bin/taskflow
```

### One-Line Install

```bash
./install.sh
```

This builds the binary to `~/.local/bin/taskflow` and generates the OpenCode tool wrapper.

### OpenCode Tool Wrapper Setup

TaskFlow can be integrated with [OpenCode](https://opencode.ai) as a tool wrapper. Generate the TypeScript wrapper and install it:

```bash
# Generate and install the tool wrapper to OpenCode's tools directory
taskflow tool-wrapper > ~/.config/opencode/tools/taskflow.ts

# Or with a custom output path
taskflow tool-wrapper --output ~/.config/opencode/tools/taskflow.ts

# With a custom binary path
taskflow tool-wrapper --binary-path my-task --output ~/.config/opencode/tools/taskflow.ts
```

For OpenCode to use the tool wrapper, ensure:
- The `~/.config/opencode/tools/` directory exists
- The `taskflow` binary is in your PATH

### Database Setup

The application automatically creates the SQLite database at `.taskflow/tasks.db` on first run. The database schema includes:

- `tasks` table with columns: id, milestone, sprint, title, description, status, priority, actor, blocked_by, created, last_updated
- `deleted_tasks` table with an additional `deleted_on` column for soft-deleted tasks
- Indexes on: milestone, status, sprint, priority, deleted_on

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

List all tasks with optional filtering and formatting.

```bash
# List all tasks (default table format)
taskflow list

# Filter by milestone
taskflow list '{"milestone":"sprint-1"}'

# Filter by status and actor
taskflow list '{"status":"todo","actor":"john"}'

# Output as markdown or xml
taskflow list '{"format":"markdown"}'
taskflow list '{"format":"xml"}'

# Sort by field
taskflow list '{"sort_by":"status"}'
taskflow list '{"sort_by":"priority"}'
taskflow list '{"sort_by":"milestone"}'
taskflow list '{"sort_by":"created"}'
taskflow list '{"sort_by":"updated"}'
taskflow list '{"sort_by":"id"}'
taskflow list '{"sort_by":"sprint"}'
taskflow list '{"sort_by":"title"}'
taskflow list '{"sort_by":"description"}'
taskflow list '{"sort_by":"actor"}'

# Pagination
taskflow list '{"limit":10,"offset":0}'

# Include completed tasks
taskflow list '{"all":true}'

# Get specific task by ID
taskflow list '{"id":"task-123"}'

# Combined example
taskflow list '{"milestone":"v1.0","status":"in_progress","format":"table","limit":50}'
```

#### JSON Fields

| Field | Default | Description |
|-------|---------|-------------|
| `milestone` | - | Filter by milestone |
| `sprint` | - | Filter by sprint |
| `status` | - | Filter by status (todo, in_progress, done, blocked) |
| `actor` | - | Filter by actor |
| `id` | - | Get specific task by ID |
| `sort_by` | - | Sort by field (status, priority, milestone, created, updated, id, sprint, title, description, actor) |
| `format` | `table` | Output format (table, markdown, xml) |
| `limit` | `20` | Maximum tasks to display |
| `offset` | `0` | Number of tasks to skip |
| `all` | `false` | Show all tasks including completed |

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

### OpenCode Tool Wrapper Commands

The `tool-wrapper` command generates a TypeScript wrapper for OpenCode integration. The wrapper exposes these additional commands that map to taskflow CLI calls:

| Wrapper Command | Maps To | Description |
|-----------------|---------|-------------|
| `start` | `taskflow update '{"id":"...","status":"in_progress"}'` | Start working on a task |
| `list_all` | `taskflow list` with filters | List all tasks |
| `list_blocked` | `taskflow list '{"status":"blocked"}'` | List blocked tasks |
| `list_done` | `taskflow list '{"status":"done"}'` | List completed tasks |
| `list_status_in_progress` | `taskflow list '{"status":"in_progress"}'` | List in-progress tasks |
| `list_status_todo` | `taskflow list '{"status":"todo"}'` | List todo tasks |

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
│   ├── tool-wrapper.go    # Generate OpenCode tool wrapper
│   ├── testutil.go        # Test utilities
│   └── *_test.go          # Unit tests for commands
├── internal/
│   ├── db/
│   │   ├── db.go          # Database connection and initialization
│   │   ├── operations.go  # CRUD operations
│   │   ├── query_builder.go  # Query building utilities
│   │   ├── filters.go     # Filtering logic
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
│   ├── validation/
│   │   ├── validator.go   # Input validation
│   │   ├── status.go      # Status validation constants and helpers
│   │   └── *_test.go      # Validation tests
│   └── README.md          # Internal package documentation
├── pkg/
│   ├── output/
│   │   ├── formatter.go   # Output formatting (table, markdown, xml)
│   │   ├── table.go       # Table renderer
│   │   └── *_test.go      # Formatter tests
│   ├── errors/
│   │   ├── errors.go      # Error handling utilities and CLI error types
│   │   └── *_test.go      # Error tests
│   ├── generator/
│   │   ├── opencode.go    # OpenCode shell wrapper generator
│   │   ├── tool.go        # TypeScript tool wrapper generator
│   │   └── tool_test.go   # Generator tests
│   └── README.md          # Package documentation
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
- **internal/validation/**: Input validation - ID format, status values, field constraints
- **pkg/output/**: Output formatting - table, markdown, and xml renderers
- **pkg/errors/**: Error handling - custom error types and formatting
- **pkg/generator/**: Code generation - TypeScript tool wrappers and OpenCode shell wrappers

### Database Schema

```sql
CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    milestone TEXT,
    sprint TEXT,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'todo',
    priority INTEGER DEFAULT 0,
    actor TEXT,
    blocked_by TEXT,
    created TEXT NOT NULL,
    last_updated TEXT NOT NULL
);

CREATE INDEX idx_tasks_milestone ON tasks(milestone);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_sprint ON tasks(sprint);
CREATE INDEX idx_tasks_priority ON tasks(priority);

CREATE TABLE deleted_tasks (
    id TEXT PRIMARY KEY,
    milestone TEXT,
    sprint TEXT,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL,
    priority INTEGER DEFAULT 0,
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

```
Task CLI version: 0.1.0
```

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
taskflow list '{"milestone":"sprint-1","format":"markdown"}'

# 5. Complete tasks
taskflow complete '{"id":"TASK-1"}'
taskflow complete '{"id":"TASK-2"}'

# 6. View completed tasks
taskflow list '{"all":true,"milestone":"sprint-1"}'

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

# Unit tests for pkg packages
go test ./pkg/...

# Integration tests
go test ./tests/integration/...
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
