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

## Installation

### Prerequisites

- Go 1.22 or later
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

### OpenCode Tool Wrapper Setup

TaskFlow can be integrated with [OpenCode](https://opencode.ai) as a tool wrapper. Generate the TypeScript wrapper and install it:

```bash
# Generate and install the tool wrapper to OpenCode's tools directory
taskflow tool-wrapper > ~/.config/opencode/tools/taskflow.ts

# Or with a custom output path
taskflow tool-wrapper --output ~/.config/opencode/tools/taskflow.ts
```

For OpenCode to use the tool wrapper, ensure:
- The `~/.config/opencode/tools/` directory exists
- The `taskflow` binary is in your PATH

### Database Setup

The application automatically creates the SQLite database at `.taskflow/tasks.db` on first run. The database schema includes:

- `tasks` table with columns: id, title, description, milestone, status, actor, created_at, updated_at
- `deleted_tasks` table for soft-deleted tasks
- Proper indexes for efficient querying by status, milestone, and actor

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
taskflow add --id "1" --title "Implement login feature" --milestone "v1.0" --description "Add login functionality"
taskflow add --id "2" --title "Fix memory leak" --description "Memory leak in data processing" --milestone "v2.0"
taskflow add --id "3" --title "Deploy to production" --milestone "v2.0 Release" --actor "devops"
```

#### Flags

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--id` | `-i` | Yes | Unique task identifier |
| `--title` | `-t` | Yes | Task title |
| `--description` | `-d` | Yes | Task description |
| `--milestone` | `-m` | Yes | Milestone/sprint name |
| `--actor` | `-a` | No | Assigned actor/owner |

---

### Command: update

Update an existing task by its ID.

```bash
taskflow update --id "1" --title "New title"
taskflow update --id "abc" --description "Updated description" --milestone "v2.0"
taskflow update --id "1" --status "in_progress" --actor "new-owner"
taskflow update --id "1" --milestone "v2.0" --description "New description"
```

#### Flags

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--id` | `-i` | Yes | Task ID to update |
| `--title` | `-t` | No | New task title |
| `--description` | `-d` | No | New task description |
| `--status` | `-s` | No | New status (todo, in_progress, done, blocked) |
| `--milestone` | `-m` | No | New milestone |
| `--actor` | `-a` | No | New assigned actor |

**Note:** At least one of title, description, status, milestone, or actor must be provided.

---

### Command: complete

Mark a task as completed.

```bash
taskflow complete --id "1"
taskflow complete --id "abc123"
```

#### Flags

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--id` | `-i` | Yes | Task ID to complete |

---

### Command: block

Block a task with a reason. Blocked tasks cannot be worked on until unblocked.

```bash
taskflow block --id "1" --reason "Waiting for API documentation"
taskflow block --id "abc123" -r "Dependency not available"
```

#### Flags

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--id` | `-i` | Yes | Task ID to block |
| `--reason` | `-r` | Yes | Reason for blocking |

---

### Command: delete

Soft delete a task by moving it to the deleted_tasks table.

```bash
taskflow delete --id "1"
taskflow delete --id "abc123"
```

#### Flags

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--id` | `-i` | Yes | Task ID to delete |

---

### Command: list

List all tasks with optional filtering and formatting.

```bash
# List all tasks (default table format)
taskflow list

# Filter by milestone
taskflow list -m "sprint-1"

# Filter by status and actor
taskflow list -s pending -a john

# Output as markdown or xml
taskflow list --format markdown
taskflow list --format xml

# Sort by field
taskflow list --sort-by status
taskflow list --sort-by created

# Pagination
taskflow list --limit 10 --offset 0

# Include completed tasks
taskflow list --all

# Get specific task by ID
taskflow list --id "task-123"

# Combined example
taskflow list -m "v1.0" -s in_progress --format table --limit 50
```

#### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--all` | `-a` | `false` | Show all tasks including completed |
| `--milestone` | `-m` | - | Filter by milestone |
| `--sprint` | `-r` | - | Filter by sprint |
| `--status` | `-s` | - | Filter by status (todo, in_progress, done, blocked) |
| `--actor` | - | - | Filter by actor |
| `--id` | - | - | Get specific task by ID |
| `--sort-by` | - | - | Sort by field (status, priority, milestone, created, updated) |
| `--format` | `-f` | `table` | Output format (table\|markdown\|xml) |
| `--limit` | `-l` | `20` | Maximum tasks to display |
| `--offset` | `-o` | `0` | Number of tasks to skip |

---

### Command: reset-timedout

Reset in-progress tasks that have exceeded the specified timeout duration back to todo status.

```bash
# Reset tasks in progress longer than 30 minutes
taskflow reset-timedout --minutes 30

# Reset tasks in progress longer than 60 minutes
taskflow reset-timedout --minutes 60

# Custom timeout
taskflow reset-timedout -m 45
```

#### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--minutes` | `-m` | `30` | Timeout duration in minutes |

---

## Architecture

```
taskflow/
├── main.go                 # Application entry point
├── cmd/
│   ├── root.go            # Root command and global flags
│   ├── add.go             # Add task command
│   ├── update.go          # Update task command
│   ├── complete.go        # Complete task command
│   ├── block.go           # Block task command
│   ├── delete.go          # Delete task command
│   ├── list.go            # List tasks command
│   ├── reset.go           # Reset timed-out tasks command
│   ├── testutil.go        # Test utilities
│   └── *_test.go          # Unit tests for commands
├── internal/
│   ├── db/
│   │   ├── db.go          # Database connection and initialization
│   │   ├── operations.go  # CRUD operations
│   │   ├── query_builder.go  # Query building utilities
│   │   ├── filters.go     # Filtering logic
│   │   └── errors.go      # Database error types
│   ├── service/
│   │   ├── add.go         # Add task service logic
│   │   ├── update.go      # Update task service logic
│   │   ├── complete.go    # Complete task service logic
│   │   ├── block.go       # Block task service logic
│   │   ├── delete.go      # Delete task service logic
│   │   ├── list.go        # List tasks service logic
│   │   ├── reset.go       # Reset timed-out service logic
│   │   ├── timeout.go     # Timeout handling utilities
│   │   └── errors.go      # Service error types
│   └── validation/
│       ├── validator.go   # Input validation
│       └── status.go      # Status validation
├── pkg/
│   ├── output/
│   │   ├── formatter.go   # Output formatting (table, markdown, xml)
│   │   ├── table.go       # Table renderer
│   │   └── xml.go         # XML renderer
│   ├── errors/
│   │   └── errors.go      # Error handling utilities
│   └── generator/
│       ├── opencode.go    # OpenCode integration
│       └── tool.go        # Tool wrapper
├── scripts/
│   └── *.sh               # Integration test scripts
├── tests/
│   └── integration/       # Integration tests
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
- **pkg/generator/**: Code generation and tool integration

### Database Schema

```sql
CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    milestone TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'todo',
    actor TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_milestone ON tasks(milestone);
CREATE INDEX idx_tasks_actor ON tasks(actor);
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

These scripts validate the CLI functionality and can be used for continuous integration.

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

### Status Transitions

```
todo -> in_progress (update --status in_progress)
todo -> done (complete --id <id>)
in_progress -> done (complete --id <id>)
in_progress -> blocked (block --id <id> --reason "...")
blocked -> todo (reset -- not directly supported, use update)
done -> todo (update --status todo)
```

---

## Error Handling

The CLI provides descriptive error messages for common issues:

- **Missing required fields**: "at least one of --title, --description, --status, --milestone, or --actor is required"
- **Invalid ID format**: "ID must be a non-empty string"
- **Invalid status**: "status must be one of: todo, in_progress, done, blocked"
- **Database errors**: "failed to connect to database", "task not found"

---

## Configuration

### Verbose Mode

Enable verbose output for debugging:

```bash
taskflow --verbose add --id "1" --title "Test" --description "Test task" --milestone "v1.0"
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
taskflow add --id "TASK-1" --title "Design database schema" --description "Create ERD and schema definitions" --milestone "sprint-1" --actor "alice"
taskflow add --id "TASK-2" --title "Implement API endpoints" --description "REST API for task CRUD" --milestone "sprint-1" --actor "bob"
taskflow add --id "TASK-3" --title "Write unit tests" --description "Test coverage for API" --milestone "sprint-1" --actor "alice"

# 2. Start working on a task
taskflow update --id "TASK-1" --status in_progress

# 3. Block a task waiting on dependency
taskflow block --id "TASK-3" --reason "Waiting for API to be ready"

# 4. Check progress
taskflow list -m "sprint-1"
taskflow list -m "sprint-1" --format markdown

# 5. Complete tasks
taskflow complete --id "TASK-1"
taskflow complete --id "TASK-2"

# 6. View completed tasks
taskflow list --all -m "sprint-1"

# 7. Check for timed-out tasks
taskflow reset-timedout --minutes 30
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