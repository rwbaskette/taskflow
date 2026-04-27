#!/bin/bash
#
# =============================================================================
# SPRINT-MANAGEMENT.SH - Sprint and Task Workflow Management Script
# =============================================================================
#
# DESCRIPTION:
#   This script demonstrates how to manage sprints and task workflows using
#   the opencode CLI tools. It provides commands for creating sprints,
#   managing sprint tasks, tracking sprint progress, and performing
#   sprint-related operations.
#
# USAGE:
#   ./sprint-management.sh [COMMAND] [OPTIONS]
#
# COMMANDS:
#   start-sprint NAME       Start a new sprint with the given name
#   list-sprints            List all sprints and their tasks
#   sprint-progress SPRINT  Show progress for a specific sprint
#   end-sprint NAME         End/close a sprint
#   assign-task TASKID SPRINT Assign a task to a sprint
#   unassign-task TASKID   Remove task from its sprint
#   block-sprint-tasks REASON   Block all in-progress tasks in a sprint
#   help                   Display this help message
#
# EXAMPLES:
#   # Start a new sprint
#   ./sprint-management.sh start-sprint "Sprint 1"
#
#   # List all sprints
#   ./sprint-management.sh list-sprints
#
#   # Show sprint progress
#   ./sprint-management.sh sprint-progress "Sprint 1"
#
#   # Assign task to sprint
#   ./sprint-management.sh assign-task "task-1" "Sprint 1"
#
#   # Block all tasks in a sprint with a reason
#   ./sprint-management.sh block-sprint-tasks "Waiting for API"
#
# NOTES:
#   - Tasks are associated with sprints via the milestone field
#   - Sprint names are stored as milestones in the task database
#   - Use the milestone filter when listing tasks to see sprint tasks
#
# =============================================================================

# -----------------------------------------------------------------------------
# SCRIPT CONFIGURATION
# -----------------------------------------------------------------------------

# Configuration constants
SPRINT_MILESTONE_PREFIX="sprint-"

# ANSI color codes for terminal output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# Sprint statuses
STATUS_NOT_STARTED="not_started"
STATUS_IN_PROGRESS="in_progress"
STATUS_COMPLETED="completed"

# -----------------------------------------------------------------------------
# FUNCTION: Display Help
# -----------------------------------------------------------------------------

show_help() {
    cat << EOF
Sprint Management Script

DESCRIPTION:
    This script provides comprehensive sprint management capabilities
    using the opencode task CLI. It allows you to create, manage,
    and track sprints and their associated tasks.

COMMANDS:
    start-sprint NAME       Start a new sprint with the given name
    list-sprints            List all sprints and their tasks
    sprint-progress SPRINT  Show progress for a specific sprint
    end-sprint NAME         End/close a sprint
    assign-task TASKID SPRINT Assign a task to a sprint
    unassign-task TASKID    Remove task from its sprint
    block-sprint-tasks REASON   Block all in-progress tasks in a sprint
    help                   Display this help message

EXAMPLES:
    # Start a new sprint
    ./sprint-management.sh start-sprint "Sprint 1"

    # List all sprints
    ./sprint-management.sh list-sprints

    # Show sprint progress
    ./sprint-management.sh sprint-progress "Sprint 1"

    # Assign task to sprint
    ./sprint-management.sh assign-task "task-1" "Sprint 1"

    # Block all tasks in a sprint
    ./sprint-management.sh block-sprint-tasks "Waiting for API"

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

log_sprint() {
    echo -e "${MAGENTA}[SPRINT]${NC} $1"
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
# FUNCTION: Validate Sprint Name
# -----------------------------------------------------------------------------

validate_sprint_name() {
    local sprint_name="$1"
    
    if [ -z "$sprint_name" ]; then
        log_error "Sprint name cannot be empty"
        return 1
    fi
    
    # Sanitize sprint name for use as milestone
    # Convert to lowercase, replace spaces with hyphens
    echo "$sprint_name" | tr '[:upper:]' '[:lower:]' | tr ' ' '-'
    return 0
}

# -----------------------------------------------------------------------------
# COMMAND: Start Sprint
# -----------------------------------------------------------------------------

cmd_start_sprint() {
    local sprint_name="$1"
    
    if [ -z "$sprint_name" ]; then
        log_error "Sprint name is required"
        echo "Usage: $0 start-sprint <sprint-name>"
        exit 1
    fi
    
    log_sprint "Starting sprint: $sprint_name"
    
    # Validate and sanitize sprint name for milestone
    local milestone
    milestone=$(validate_sprint_name "$sprint_name")
    
    # Add sprint planning tasks
    # These are typical tasks for starting a new sprint
    
    log_info "Creating sprint planning tasks..."
    
    local task_failures=0
    
    # Task 1: Sprint planning meeting
    if ! task add --id "${milestone}-planning" \
             --title "Sprint planning meeting" \
             --description "Conduct sprint planning meeting to define sprint goals and backlog" \
             --milestone "$milestone" \
             --actor "scrum-master" 2>/dev/null; then
        log_warning "Failed to create sprint planning task"
        ((task_failures++))
    fi
    
    # Task 2: Backlog refinement
    if ! task add --id "${milestone}-backlog" \
             --title "Backlog refinement" \
             --description "Refine and prioritize product backlog for this sprint" \
             --milestone "$milestone" \
             --actor "product-owner" 2>/dev/null; then
        log_warning "Failed to create backlog refinement task"
        ((task_failures++))
    fi
    
    # Task 3: Task estimation
    if ! task add --id "${milestone}-estimation" \
             --title "Task estimation" \
             --description "Estimate effort for all sprint backlog items" \
             --milestone "$milestone" \
             --actor "team" 2>/dev/null; then
        log_warning "Failed to create task estimation task"
        ((task_failures++))
    fi
    
    # Task 4: Sprint goal definition
    if ! task add --id "${milestone}-goal" \
             --title "Define sprint goal" \
             --description "Define clear sprint goal and success criteria" \
             --milestone "$milestone" \
             --actor "product-owner" 2>/dev/null; then
        log_warning "Failed to create sprint goal task"
        ((task_failures++))
    fi
    
    if [ $task_failures -gt 0 ]; then
        log_warning "Sprint created with $task_failures task creation failures"
    fi
    
    log_success "Sprint '$sprint_name' started successfully!"
    log_info "Sprint milestone: $milestone"
    echo ""
    log_info "Sprint tasks:"
    task list --milestone "$milestone" --format table
}

# -----------------------------------------------------------------------------
# COMMAND: List Sprints
# -----------------------------------------------------------------------------

cmd_list_sprints() {
    log_sprint "Listing all sprints..."
    echo ""
    
    # Get all unique milestones (which represent sprints)
    local sprints
    local list_output
    list_output=$(task list --format markdown 2>&1)
    local list_status=$?
    
    if [ $list_status -ne 0 ]; then
        log_error "Failed to list tasks: $list_output"
        return 1
    fi
    
    sprints=$(echo "$list_output" | grep -E "^\|" | awk -F'|' '{gsub(/^[ \t]+|[ \t]+$/, "", $3); if(NR>2 && $3 != "") print $3}' | sort -u)
    
    if [ -z "$sprints" ]; then
        log_warning "No sprints found"
        return 0
    fi
    
    echo "=============================================="
    echo "  AVAILABLE SPRINTS"
    echo "=============================================="
    echo ""
    
    # Display each sprint with its task count
    while IFS= read -r sprint; do
        if [ -n "$sprint" ]; then
            # Skip header and empty lines
            if [[ "$sprint" != "ID" && "$sprint" != "---" && ${#sprint} -gt 2 ]]; then
                echo -e "${CYAN}Sprint: $sprint${NC}"
                
                # Count tasks in this sprint - subtract 2 for header rows (separator and header)
                local list_output
                list_output=$(task list --milestone "$sprint" 2>&1)
                local list_status=$?
                
                local count=0
                if [ $list_status -eq 0 ]; then
                    count=$(echo "$list_output" | grep -c "|" || echo "0")
                    # Subtract 2 for the separator line and header line in markdown table
                    count=$((count - 2))
                    [ $count -lt 0 ] && count=0
                fi
                
                local todo_count=0
                local in_progress_count=0
                local done_count=0
                local blocked_count=0
                
                # Get status-specific counts
                todo_count=$(task list --milestone "$sprint" --status "todo" 2>/dev/null | grep -c "|" || echo "0")
                in_progress_count=$(task list --milestone "$sprint" --status "in_progress" 2>/dev/null | grep -c "|" || echo "0")
                done_count=$(task list --milestone "$sprint" --status "done" 2>/dev/null | grep -c "|" || echo "0")
                blocked_count=$(task list --milestone "$sprint" --status "blocked" 2>/dev/null | grep -c "|" || echo "0")
                
                # Each status query returns its own table - subtract 2 for header rows
                todo_count=$((todo_count - 2))
                in_progress_count=$((in_progress_count - 2))
                done_count=$((done_count - 2))
                blocked_count=$((blocked_count - 2))
                
                # Ensure no negative counts
                [ $todo_count -lt 0 ] && todo_count=0
                [ $in_progress_count -lt 0 ] && in_progress_count=0
                [ $done_count -lt 0 ] && done_count=0
                [ $blocked_count -lt 0 ] && blocked_count=0
                
                echo "  Total tasks: $count"
                echo "  - Todo: $todo_count"
                echo "  - In Progress: $in_progress_count"
                echo "  - Done: $done_count"
                echo "  - Blocked: $blocked_count"
                echo ""
            fi
        fi
    done <<< "$sprints"
}

# -----------------------------------------------------------------------------
# COMMAND: Sprint Progress
# -----------------------------------------------------------------------------

cmd_sprint_progress() {
    local sprint_name="$1"
    
    if [ -z "$sprint_name" ]; then
        log_error "Sprint name is required"
        echo "Usage: $0 sprint-progress <sprint-name>"
        exit 1
    fi
    
    local milestone
    milestone=$(validate_sprint_name "$sprint_name")
    
    log_sprint "Showing progress for sprint: $sprint_name"
    echo ""
    
    # Get task counts for each status - each query is already filtered, so subtract 2 for headers
    local todo_count
    todo_count=$(task list --milestone "$milestone" --status "todo" 2>/dev/null | grep -c "|" || echo "0")
    todo_count=$((todo_count - 2))
    [ $todo_count -lt 0 ] && todo_count=0
    
    local in_progress_count
    in_progress_count=$(task list --milestone "$milestone" --status "in_progress" 2>/dev/null | grep -c "|" || echo "0")
    in_progress_count=$((in_progress_count - 2))
    [ $in_progress_count -lt 0 ] && in_progress_count=0
    
    local done_count
    done_count=$(task list --milestone "$milestone" --status "done" 2>/dev/null | grep -c "|" || echo "0")
    done_count=$((done_count - 2))
    [ $done_count -lt 0 ] && done_count=0
    
    local blocked_count
    blocked_count=$(task list --milestone "$milestone" --status "blocked" 2>/dev/null | grep -c "|" || echo "0")
    blocked_count=$((blocked_count - 2))
    [ $blocked_count -lt 0 ] && blocked_count=0
    
    # Sum of already-filtered counts (no subtraction needed - each query is separate)
    local total_tasks=$((todo_count + in_progress_count + done_count + blocked_count))
    
    # Calculate percentage with bounds checking
    local progress=0
    if [ $total_tasks -gt 0 ]; then
        progress=$((done_count * 100 / total_tasks))
    fi
    
    echo "=============================================="
    echo "  SPRINT PROGRESS: $sprint_name"
    echo "=============================================="
    echo ""
    echo "Sprint Milestone: $milestone"
    echo ""
    echo "Task Status Breakdown:"
    echo "----------------------"
    echo -e "  ${GREEN}Todo:${NC}          $todo_count"
    echo -e "  ${YELLOW}In Progress:${NC}  $in_progress_count"
    echo -e "  ${GREEN}Done:${NC}          $done_count"
    echo -e "  ${RED}Blocked:${NC}        $blocked_count"
    echo ""
    echo "Total Tasks: $total_tasks"
    echo "Progress: $progress%"
    echo ""
    
    # Show task list
    echo "Sprint Tasks:"
    echo "-------------"
    task list --milestone "$milestone" --format table
    
    # Check for blocked tasks (threshold adjusted since we're not subtracting headers)
    if [ $blocked_count -gt 0 ]; then
        echo ""
        log_warning "This sprint has blocked tasks that may need attention!"
    fi
    
    # Check for tasks in progress for too long
    if [ $in_progress_count -gt 0 ]; then
        echo ""
        log_info "Consider using: $0 block-sprint-tasks 'reason' to block in-progress tasks"
    fi
}

# -----------------------------------------------------------------------------
# COMMAND: End Sprint
# -----------------------------------------------------------------------------

cmd_end_sprint() {
    local sprint_name="$1"
    
    if [ -z "$sprint_name" ]; then
        log_error "Sprint name is required"
        echo "Usage: $0 end-sprint <sprint-name>"
        exit 1
    fi
    
    local milestone
    milestone=$(validate_sprint_name "$sprint_name")
    
    log_sprint "Ending sprint: $sprint_name"
    
    # Get all remaining tasks
    local remaining_tasks
    remaining_tasks=$(task list --milestone "$milestone" --status "todo" 2>/dev/null)
    
    if [ -n "$remaining_tasks" ]; then
        log_warning "There are uncompleted tasks in this sprint:"
        task list --milestone "$milestone" --status "todo" --format table
        echo ""
        read -p "Do you want to move remaining tasks to the next sprint? (y/n): " confirm
        
        if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
            log_info "Please specify the next sprint name:"
            read -p "Next sprint name: " next_sprint
            
            if [ -n "$next_sprint" ]; then
                local next_milestone
                next_milestone=$(validate_sprint_name "$next_sprint")
                
                log_info "Moving tasks to sprint: $next_sprint"
                # Update each task's milestone
                task list --milestone "$milestone" --status "todo" --format markdown 2>/dev/null | \
                    grep -E "^\|" | awk -F'|' '{gsub(/^[ \t]+|[ \t]+$/, "", $2); print $2}' | \
                    while read -r task_id; do
                        if [ -n "$task_id" ] && [ "$task_id" != "ID" ] && [ "$task_id" != "---" ]; then
                            task update --id "$task_id" --milestone "$next_milestone" 2>/dev/null
                        fi
                    done
                
                log_success "Tasks moved to sprint: $next_sprint"
            fi
        fi
    fi
    
    log_success "Sprint '$sprint_name' has been ended"
}

# -----------------------------------------------------------------------------
# COMMAND: Assign Task to Sprint
# -----------------------------------------------------------------------------

cmd_assign_task() {
    local task_id="$1"
    local sprint_name="$2"
    
    if [ -z "$task_id" ] || [ -z "$sprint_name" ]; then
        log_error "Task ID and Sprint name are required"
        echo "Usage: $0 assign-task <task-id> <sprint-name>"
        exit 1
    fi
    
    local milestone
    milestone=$(validate_sprint_name "$sprint_name")
    
    log_info "Assigning task '$task_id' to sprint '$sprint_name'"
    
    # Update task's milestone to assign to sprint
    task update --id "$task_id" --milestone "$milestone"
    
    if [ $? -eq 0 ]; then
        log_success "Task '$task_id' assigned to sprint '$sprint_name'"
    else
        log_error "Failed to assign task to sprint"
        exit 1
    fi
}

# -----------------------------------------------------------------------------
# COMMAND: Unassign Task
# -----------------------------------------------------------------------------

cmd_unassign_task() {
    local task_id="$1"
    
    if [ -z "$task_id" ]; then
        log_error "Task ID is required"
        echo "Usage: $0 unassign-task <task-id>"
        exit 1
    fi
    
    log_info "Unassigning task '$task_id' from its sprint"
    
    # Update task's milestone to empty (or use a default like 'backlog')
    task update --id "$task_id" --milestone "backlog"
    
    if [ $? -eq 0 ]; then
        log_success "Task '$task_id' has been unassigned from sprint"
    else
        log_error "Failed to unassign task"
        exit 1
    fi
}

# -----------------------------------------------------------------------------
# COMMAND: Block Sprint Tasks
# -----------------------------------------------------------------------------

cmd_block_sprint_tasks() {
    local reason="$1"
    
    if [ -z "$reason" ]; then
        log_error "Reason is required"
        echo "Usage: $0 block-sprint-tasks <reason>"
        exit 1
    fi
    
    log_warning "Blocking all in-progress tasks in all sprints..."
    log_info "Reason: $reason"
    
    # Get all in-progress tasks
    local in_progress_tasks
    in_progress_tasks=$(task list --status "in_progress" --format markdown 2>/dev/null)
    
    if [ -z "$in_progress_tasks" ]; then
        log_info "No in-progress tasks found"
        return 0
    fi
    
    # Block each in-progress task
    echo "$in_progress_tasks" | grep -E "^\|" | awk -F'|' '{gsub(/^[ \t]+|[ \t]+$/, "", $2); print $2}' | \
        while read -r task_id; do
            if [ -n "$task_id" ] && [ "$task_id" != "ID" ] && [ "$task_id" != "---" ]; then
                echo "Blocking task: $task_id"
                task block --id "$task_id" --reason "$reason" 2>/dev/null
            fi
        done
    
    log_success "All in-progress tasks have been blocked"
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
        start-sprint)
            cmd_start_sprint "$@"
            ;;
        list-sprints)
            cmd_list_sprints
            ;;
        sprint-progress)
            cmd_sprint_progress "$@"
            ;;
        end-sprint)
            cmd_end_sprint "$@"
            ;;
        assign-task)
            cmd_assign_task "$@"
            ;;
        unassign-task)
            cmd_unassign_task "$@"
            ;;
        block-sprint-tasks)
            cmd_block_sprint_tasks "$@"
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