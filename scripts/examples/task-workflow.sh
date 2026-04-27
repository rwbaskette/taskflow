#!/bin/bash
#
# =============================================================================
# TASK-WORKFLOW.SH - Task Workflow Operations Script
# =============================================================================
#
# DESCRIPTION:
#   This script demonstrates comprehensive task workflow operations using
#   the opencode CLI tools. It provides commands for managing the complete
#   lifecycle of tasks including creation, status updates, blocking, 
#   completion, and various workflow patterns.
#
# USAGE:
#   ./task-workflow.sh [COMMAND] [OPTIONS]
#
# COMMANDS:
#   create TASKID TITLE DESC MILESTONE [ACTOR]  Create a new task
#   start TASKID                            Start working on a task
#   complete TASKID                          Mark a task as completed
#   block TASKID REASON                      Block a task with reason
#   unblock TASKID                           Unblock a task
#   update TASKID FIELD VALUE                Update task field
#   show TASKID                              Show task details
#   list-all                                List all tasks
#   list-todo                               List tasks with todo status
#   list-progress                           List tasks in progress
#   list-done                               List completed tasks
#   list-blocked                            List blocked tasks
#   list-milestone MILESTONE                 List tasks by milestone
#   list-actor ACTOR                         List tasks by actor
#   bulk-block TASKID... REASON              Block multiple tasks
#   bulk-complete TASKID...                  Complete multiple tasks
#   reset-timedout MINUTES                   Reset timed out tasks
#   help                                    Display this help message
#
# EXAMPLES:
#   # Create a new task
#   ./task-workflow.sh create "task-1" "Implement login" "Add login feature" "v1.0" "dev"
#
#   # Start working on a task
#   ./task-workflow.sh start "task-1"
#
#   # Complete a task
#   ./task-workflow.sh complete "task-1"
#
#   # Block a task
#   ./task-workflow.sh block "task-1" "Waiting for API documentation"
#
#   # Update task status
#   ./task-workflow.sh update "task-1" "status" "in_progress"
#
#   # List all tasks in progress
#   ./task-workflow.sh list-progress
#
#   # List tasks by milestone
#   ./task-workflow.sh list-milestone "sprint-1"
#
#   # Bulk complete multiple tasks
#   ./task-workflow.sh bulk-complete "task-1" "task-2" "task-3"
#
#   # Reset tasks that have been in progress for more than 30 minutes
#   ./task-workflow.sh reset-timedout 30
#
# VALID STATUS VALUES:
#   todo         - Task has not been started
#   in_progress  - Task is currently being worked on
#   done         - Task has been completed
#   blocked      - Task is blocked and cannot proceed
#
# VALID FIELDS FOR UPDATE:
#   title        - Task title
#   description  - Task description
#   milestone   - Task milestone/sprint
#   status       - Task status
#   actor        - Assigned team member
#
# =============================================================================

# -----------------------------------------------------------------------------
# SCRIPT CONFIGURATION
# -----------------------------------------------------------------------------

# ANSI color codes for terminal output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# Valid status values
VALID_STATUSES=("todo" "in_progress" "done" "blocked")

# Valid fields for update
VALID_FIELDS=("title" "description" "milestone" "status" "actor")

# -----------------------------------------------------------------------------
# FUNCTION: Display Help
# -----------------------------------------------------------------------------

show_help() {
    cat << 'EOF'
Task Workflow Operations Script

DESCRIPTION:
    This script provides comprehensive task workflow management capabilities
    using the opencode task CLI. It supports the complete task lifecycle
    from creation to completion, with support for various workflow patterns
    including blocking, bulk operations, and filtering.

COMMANDS:
    create TASKID TITLE DESC MILESTONE [ACTOR]  Create a new task
    start TASKID                            Start working on a task
    complete TASKID                          Mark a task as completed
    block TASKID REASON                      Block a task with reason
    unblock TASKID                           Unblock a task
    update TASKID FIELD VALUE                Update task field
    show TASKID                              Show task details
    list-all                                List all tasks
    list-todo                               List tasks with todo status
    list-progress                           List tasks in progress
    list-done                               List completed tasks
    list-blocked                            List blocked tasks
    list-milestone MILESTONE                 List tasks by milestone
    list-actor ACTOR                         List tasks by actor
    bulk-block TASKID... REASON              Block multiple tasks
    bulk-complete TASKID...                  Complete multiple tasks
    reset-timedout MINUTES                   Reset timed out tasks
    help                                    Display this help message

EXAMPLES:
    # Create a new task
    ./task-workflow.sh create "task-1" "Implement login" "Add login feature" "v1.0" "dev"

    # Start working on a task
    ./task-workflow.sh start "task-1"

    # Complete a task
    ./task-workflow.sh complete "task-1"

    # Block a task
    ./task-workflow.sh block "task-1" "Waiting for API documentation"

    # Update task status
    ./task-workflow.sh update "task-1" "status" "in_progress"

    # List all tasks in progress
    ./task-workflow.sh list-progress

    # List tasks by milestone
    ./task-workflow.sh list-milestone "sprint-1"

    # Bulk complete multiple tasks
    ./task-workflow.sh bulk-complete "task-1" "task-2" "task-3"

    # Reset tasks that have been in progress for more than 30 minutes
    ./task-workflow.sh reset-timedout 30

VALID STATUS VALUES:
    todo         - Task has not been started
    in_progress  - Task is currently being worked on
    done         - Task has been completed
    blocked      - Task is blocked and cannot proceed

VALID FIELDS FOR UPDATE:
    title        - Task title
    description  - Task description
    milestone   - Task milestone/sprint
    status       - Task status
    actor        - Assigned team member

EOF
    exit 0
}

# -----------------------------------------------------------------------------
# FUNCTION: Log Messages
# -----------------------------------------------------------------------------

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# -----------------------------------------------------------------------------
# FUNCTION: Verify Task CLI Availability
# -----------------------------------------------------------------------------

verify_task_cli() {
    # Check if 'task' command is available in PATH
    if ! command -v task &> /dev/null; then
        log_error "Task CLI not found in PATH."
        log_info "Please build the project first: go build -o task-cli ./main.go"
        return 1
    fi
    return 0
}

# -----------------------------------------------------------------------------
# FUNCTION: Validate Status
# -----------------------------------------------------------------------------

validate_status() {
    local status="$1"
    
    for valid_status in "${VALID_STATUSES[@]}"; do
        if [ "$status" = "$valid_status" ]; then
            return 0
        fi
    done
    
    log_error "Invalid status: $status"
    log_info "Valid statuses: ${VALID_STATUSES[*]}"
    return 1
}

# -----------------------------------------------------------------------------
# FUNCTION: Validate Field
# -----------------------------------------------------------------------------

validate_field() {
    local field="$1"
    
    for valid_field in "${VALID_FIELDS[@]}"; do
        if [ "$field" = "$valid_field" ]; then
            return 0
        fi
    done
    
    log_error "Invalid field: $field"
    log_info "Valid fields: ${VALID_FIELDS[*]}"
    return 1
}

# -----------------------------------------------------------------------------
# COMMAND: Create Task
# -----------------------------------------------------------------------------

cmd_create() {
    local task_id="$1"
    local title="$2"
    local description="$3"
    local milestone="$4"
    local actor="${5:-}"
    
    # Validate required parameters
    if [ -z "$task_id" ]; then
        log_error "Task ID is required"
        echo "Usage: $0 create <task-id> <title> <description> <milestone> [actor]"
        exit 1
    fi
    
    if [ -z "$title" ]; then
        log_error "Title is required"
        echo "Usage: $0 create <task-id> <title> <description> <milestone> [actor]"
        exit 1
    fi
    
    if [ -z "$description" ]; then
        log_error "Description is required"
        echo "Usage: $0 create <task-id> <title> <description> <milestone> [actor]"
        exit 1
    fi
    
    if [ -z "$milestone" ]; then
        log_error "Milestone is required"
        echo "Usage: $0 create <task-id> <title> <description> <milestone> [actor]"
        exit 1
    fi
    
    log_info "Creating task: $task_id"
    
    # Build command arguments
    local cmd_args=(
        "task"
        "add"
        "--id" "$task_id"
        "--title" "$title"
        "--description" "$description"
        "--milestone" "$milestone"
    )
    
    # Add actor if provided
    if [ -n "$actor" ]; then
        cmd_args+=("--actor" "$actor")
    fi
    
    # Execute command
    "${cmd_args[@]}"
    
    if [ $? -eq 0 ]; then
        log_success "Task '$task_id' created successfully"
    else
        log_error "Failed to create task"
        exit 1
    fi
}

# -----------------------------------------------------------------------------
# COMMAND: Start Task
# -----------------------------------------------------------------------------

cmd_start() {
    local task_id="$1"
    
    if [ -z "$task_id" ]; then
        log_error "Task ID is required"
        echo "Usage: $0 start <task-id>"
        exit 1
    fi
    
    log_info "Starting task: $task_id"
    
    # Update task status to in_progress
    task update --id "$task_id" --status "in_progress"
    
    if [ $? -eq 0 ]; then
        log_success "Task '$task_id' started"
    else
        log_error "Failed to start task"
        exit 1
    fi
}

# -----------------------------------------------------------------------------
# COMMAND: Complete Task
# -----------------------------------------------------------------------------

cmd_complete() {
    local task_id="$1"
    
    if [ -z "$task_id" ]; then
        log_error "Task ID is required"
        echo "Usage: $0 complete <task-id>"
        exit 1
    fi
    
    log_info "Completing task: $task_id"
    
    # Use complete command to mark task as done
    task complete --id "$task_id"
    
    if [ $? -eq 0 ]; then
        log_success "Task '$task_id' completed"
    else
        log_error "Failed to complete task"
        exit 1
    fi
}

# -----------------------------------------------------------------------------
# COMMAND: Block Task
# -----------------------------------------------------------------------------

cmd_block() {
    local task_id="$1"
    local reason="$2"
    
    if [ -z "$task_id" ]; then
        log_error "Task ID is required"
        echo "Usage: $0 block <task-id> <reason>"
        exit 1
    fi
    
    if [ -z "$reason" ]; then
        log_error "Reason is required"
        echo "Usage: $0 block <task-id> <reason>"
        exit 1
    fi
    
    log_warning "Blocking task: $task_id"
    log_info "Reason: $reason"
    
    # Block the task with reason
    task block --id "$task_id" --reason "$reason"
    
    if [ $? -eq 0 ]; then
        log_success "Task '$task_id' blocked"
    else
        log_error "Failed to block task"
        exit 1
    fi
}

# -----------------------------------------------------------------------------
# COMMAND: Unblock Task
# -----------------------------------------------------------------------------

cmd_unblock() {
    local task_id="$1"
    
    if [ -z "$task_id" ]; then
        log_error "Task ID is required"
        echo "Usage: $0 unblock <task-id>"
        exit 1
    fi
    
    log_info "Unblocking task: $task_id"
    
    # Unblock by setting status back to todo
    task update --id "$task_id" --status "todo"
    
    if [ $? -eq 0 ]; then
        log_success "Task '$task_id' unblocked"
    else
        log_error "Failed to unblock task"
        exit 1
    fi
}

# -----------------------------------------------------------------------------
# COMMAND: Update Task Field
# -----------------------------------------------------------------------------

cmd_update() {
    local task_id="$1"
    local field="$2"
    local value="$3"
    
    if [ -z "$task_id" ]; then
        log_error "Task ID is required"
        echo "Usage: $0 update <task-id> <field> <value>"
        exit 1
    fi
    
    if [ -z "$field" ]; then
        log_error "Field is required"
        echo "Usage: $0 update <task-id> <field> <value>"
        exit 1
    fi
    
    if [ -z "$value" ]; then
        log_error "Value is required"
        echo "Usage: $0 update <task-id> <field> <value>"
        exit 1
    fi
    
    # Validate field
    validate_field "$field" || exit 1
    
    # Validate status if field is status
    if [ "$field" = "status" ]; then
        validate_status "$value" || exit 1
    fi
    
    log_info "Updating task: $task_id"
    log_info "Field: $field, Value: $value"
    
    # Build update command
    local cmd_args=(
        "task"
        "update"
        "--id" "$task_id"
    )
    
    # Add the field to update
    case "$field" in
        title)
            cmd_args+=("--title" "$value")
            ;;
        description)
            cmd_args+=("--description" "$value")
            ;;
        milestone)
            cmd_args+=("--milestone" "$value")
            ;;
        status)
            cmd_args+=("--status" "$value")
            ;;
        actor)
            cmd_args+=("--actor" "$value")
            ;;
    esac
    
    # Execute command
    "${cmd_args[@]}"
    
    if [ $? -eq 0 ]; then
        log_success "Task '$task_id' updated"
    else
        log_error "Failed to update task"
        exit 1
    fi
}

# -----------------------------------------------------------------------------
# COMMAND: Show Task Details
# -----------------------------------------------------------------------------

cmd_show() {
    local task_id="$1"
    
    if [ -z "$task_id" ]; then
        log_error "Task ID is required"
        echo "Usage: $0 show <task-id>"
        exit 1
    fi
    
    log_info "Showing task details: $task_id"
    echo ""
    
    # List all tasks and filter for the specific task using fixed-string matching
    local list_output
    list_output=$(task list --format table 2>&1)
    local list_status=$?
    
    if [ $list_status -ne 0 ]; then
        log_error "Failed to list tasks"
        exit 1
    fi
    
    echo "$list_output" | grep -F "$task_id"
    
    if [ $? -ne 0 ]; then
        log_warning "Task '$task_id' not found"
    fi
}

# -----------------------------------------------------------------------------
# COMMAND: List All Tasks
# -----------------------------------------------------------------------------

cmd_list_all() {
    log_info "Listing all tasks"
    echo ""
    task list --format table
}

# -----------------------------------------------------------------------------
# COMMAND: List Tasks by Status
# -----------------------------------------------------------------------------

cmd_list_status() {
    local status="$1"
    
    if [ -z "$status" ]; then
        log_error "Status is required"
        echo "Usage: $0 list-todo|list-progress|list-done|list-blocked"
        exit 1
    fi
    
    # Validate status
    validate_status "$status" || exit 1
    
    log_info "Listing tasks with status: $status"
    echo ""
    task list --status "$status" --format table
}

# -----------------------------------------------------------------------------
# COMMAND: List Tasks by Milestone
# -----------------------------------------------------------------------------

cmd_list_milestone() {
    local milestone="$1"
    
    if [ -z "$milestone" ]; then
        log_error "Milestone is required"
        echo "Usage: $0 list-milestone <milestone>"
        exit 1
    fi
    
    log_info "Listing tasks in milestone: $milestone"
    echo ""
    task list --milestone "$milestone" --format table
}

# -----------------------------------------------------------------------------
# COMMAND: List Tasks by Actor
# -----------------------------------------------------------------------------

cmd_list_actor() {
    local actor="$1"
    
    if [ -z "$actor" ]; then
        log_error "Actor is required"
        echo "Usage: $0 list-actor <actor>"
        exit 1
    fi
    
    log_info "Listing tasks assigned to: $actor"
    echo ""
    task list --actor "$actor" --format table
}

# -----------------------------------------------------------------------------
# COMMAND: Bulk Block Tasks
# -----------------------------------------------------------------------------

cmd_bulk_block() {
    local reason="$1"
    shift
    local task_ids=("$@")
    
    if [ -z "$reason" ]; then
        log_error "Reason is required"
        echo "Usage: $0 bulk-block <task-id-1> <task-id-2> ... <reason>"
        exit 1
    fi
    
    if [ ${#task_ids[@]} -eq 0 ]; then
        log_error "At least one task ID is required"
        echo "Usage: $0 bulk-block <task-id-1> <task-id-2> ... <reason>"
        exit 1
    fi
    
    log_warning "Blocking ${#task_ids[@]} tasks..."
    log_info "Reason: $reason"
    echo ""
    
    # Block each task
    local success_count=0
    local fail_count=0
    
    for task_id in "${task_ids[@]}"; do
        if [ -n "$task_id" ]; then
            echo "Blocking: $task_id"
            task block --id "$task_id" --reason "$reason" 2>/dev/null
            
            if [ $? -eq 0 ]; then
                ((success_count++))
            else
                ((fail_count++))
            fi
        fi
    done
    
    echo ""
    log_success "Blocked $success_count tasks successfully"
    
    if [ $fail_count -gt 0 ]; then
        log_error "Failed to block $fail_count tasks"
    fi
}

# -----------------------------------------------------------------------------
# COMMAND: Bulk Complete Tasks
# -----------------------------------------------------------------------------

cmd_bulk_complete() {
    local task_ids=("$@")
    
    if [ ${#task_ids[@]} -eq 0 ]; then
        log_error "At least one task ID is required"
        echo "Usage: $0 bulk-complete <task-id-1> <task-id-2> ..."
        exit 1
    fi
    
    log_info "Completing ${#task_ids[@]} tasks..."
    echo ""
    
    # Complete each task
    local success_count=0
    local fail_count=0
    
    for task_id in "${task_ids[@]}"; do
        if [ -n "$task_id" ]; then
            echo "Completing: $task_id"
            task complete --id "$task_id" 2>/dev/null
            
            if [ $? -eq 0 ]; then
                ((success_count++))
            else
                ((fail_count++))
            fi
        fi
    done
    
    echo ""
    log_success "Completed $success_count tasks successfully"
    
    if [ $fail_count -gt 0 ]; then
        log_error "Failed to complete $fail_count tasks"
    fi
}

# -----------------------------------------------------------------------------
# COMMAND: Reset Timed Out Tasks
# -----------------------------------------------------------------------------

cmd_reset_timedout() {
    local minutes="$1"
    
    if [ -z "$minutes" ]; then
        minutes=30  # Default timeout
        log_info "Using default timeout: $minutes minutes"
    fi
    
    if ! [[ "$minutes" =~ ^[0-9]+$ ]]; then
        log_error "Invalid timeout value: $minutes"
        echo "Usage: $0 reset-timedout <minutes>"
        exit 1
    fi
    
    log_warning "Resetting tasks in progress for more than $minutes minutes..."
    echo ""
    
    # Execute reset command
    task reset-timedout --minutes "$minutes"
    
    if [ $? -eq 0 ]; then
        log_success "Timed out tasks reset complete"
    else
        log_error "Failed to reset timed out tasks"
        exit 1
    fi
}

# -----------------------------------------------------------------------------
# MAIN SCRIPT LOGIC
# -----------------------------------------------------------------------------

main() {
    # Verify task CLI is available
    verify_task_cli || exit 1
    
    # Check for minimum arguments
    if [ $# -eq 0 ]; then
        show_help
    fi
    
    # Parse command
    local command="$1"
    shift
    
    case "$command" in
        create)
            cmd_create "$@"
            ;;
        start)
            cmd_start "$@"
            ;;
        complete)
            cmd_complete "$@"
            ;;
        block)
            cmd_block "$@"
            ;;
        unblock)
            cmd_unblock "$@"
            ;;
        update)
            cmd_update "$@"
            ;;
        show)
            cmd_show "$@"
            ;;
        list-all)
            cmd_list_all
            ;;
        list-todo)
            cmd_list_status "todo"
            ;;
        list-progress)
            cmd_list_status "in_progress"
            ;;
        list-done)
            cmd_list_status "done"
            ;;
        list-blocked)
            cmd_list_status "blocked"
            ;;
        list-milestone)
            cmd_list_milestone "$@"
            ;;
        list-actor)
            cmd_list_actor "$@"
            ;;
        bulk-block)
            # Get reason (last argument) and all task IDs before it
            local reason="${@: -1}"
            local task_ids=("${@:1:$#-1}")
            cmd_bulk_block "$reason" "${task_ids[@]}"
            ;;
        bulk-complete)
            cmd_bulk_complete "$@"
            ;;
        reset-timedout)
            cmd_reset_timedout "$@"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "Unknown command: $command"
            echo ""
            show_help
            ;;
    esac
}

# -----------------------------------------------------------------------------
# SCRIPT ENTRY POINT
# -----------------------------------------------------------------------------

main "$@"