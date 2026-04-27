#!/bin/bash
#
# =============================================================================
# PROJECT-SETUP.SH - Opencode Project Initialization Script
# =============================================================================
#
# DESCRIPTION:
#   This script sets up a new opencode project with a predefined structure
#   of tasks, milestones, and team assignments. It demonstrates how to use
#   the opencode CLI tools to bootstrap a new project from scratch.
#
# USAGE:
#   ./project-setup.sh [OPTIONS]
#
# OPTIONS:
#   -h, --help              Display this help message and exit
#   -p, --project NAME      Set project name (default: "my-project")
#   -d, --data-dir PATH     Set data directory for tasks.db (default: "./data")
#   -v, --verbose           Enable verbose output
#
# EXAMPLES:
#   # Basic project setup with defaults
#   ./project-setup.sh
#
#   # Custom project with specific name and data directory
#   ./project-setup.sh --project "webapp" --data-dir "./my-data"
#
#   # Verbose mode for debugging
#   ./project-setup.sh --verbose
#
# NOTES:
#   - Requires the 'task' CLI tool to be built and available in PATH
#   - Creates a tasks.db database if it doesn't exist
#   - All tasks are added with "todo" status by default
#
# =============================================================================

# -----------------------------------------------------------------------------
# SCRIPT CONFIGURATION
# -----------------------------------------------------------------------------

# Default configuration values
PROJECT_NAME="my-project"
DATA_DIR="./data"
VERBOSE=false

# ANSI color codes for terminal output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# -----------------------------------------------------------------------------
# FUNCTION: Display Help
# -----------------------------------------------------------------------------

show_help() {
    # Extract and display the documentation from the script header
    head -60 "$0" | tail -50
    exit 0
}

# -----------------------------------------------------------------------------
# FUNCTION: Log Messages
# -----------------------------------------------------------------------------

log_info() {
    if [ "$VERBOSE" = true ]; then
        echo -e "${BLUE}[INFO]${NC} $1"
    fi
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
    log_info "Verifying task CLI availability..."
    
    # Check if 'task' command is available in PATH
    if ! command -v task &> /dev/null; then
        log_error "Task CLI not found in PATH. Please build the project first."
        log_info "Run: go build -o task-cli ./main.go"
        return 1
    fi
    
    log_success "Task CLI found: $(which task)"
    return 0
}

# -----------------------------------------------------------------------------
# FUNCTION: Initialize Project Structure
# -----------------------------------------------------------------------------

initialize_project_structure() {
    log_info "Initializing project structure..."
    
    # Create data directory if it doesn't exist
    if [ ! -d "$DATA_DIR" ]; then
        mkdir -p "$DATA_DIR"
        log_info "Created data directory: $DATA_DIR"
    fi
    
    log_success "Project structure initialized"
}

# -----------------------------------------------------------------------------
# FUNCTION: Add Milestone Tasks
# -----------------------------------------------------------------------------

add_milestone_tasks() {
    local milestone="$1"
    local description="$2"
    
    log_info "Adding milestone: $milestone"
    
    # Track task creation results
    local task_failures=0
    local task_success=0
    
    # Add multiple tasks for each milestone
    # Using task add command with required parameters: id, title, description, milestone
    
    # Task 1: Initial setup task
    if task add --id "${milestone}-1" \
             --title "Project setup and configuration" \
             --description "$description: Initialize project configuration and dependencies" \
             --milestone "$milestone" \
             --actor "lead-dev" 2>/dev/null; then
        ((task_success++))
    else
        log_warning "Failed to create task ${milestone}-1"
        ((task_failures++))
    fi
    
    # Task 2: Core feature development
    if task add --id "${milestone}-2" \
             --title "Implement core feature" \
             --description "$description: Develop the main feature set for this milestone" \
             --milestone "$milestone" \
             --actor "senior-dev" 2>/dev/null; then
        ((task_success++))
    else
        log_warning "Failed to create task ${milestone}-2"
        ((task_failures++))
    fi
    
    # Task 3: Testing task
    if task add --id "${milestone}-3" \
             --title "Write unit and integration tests" \
             --description "$description: Ensure code quality with comprehensive tests" \
             --milestone "$milestone" \
             --actor "qa-engineer" 2>/dev/null; then
        ((task_success++))
    else
        log_warning "Failed to create task ${milestone}-3"
        ((task_failures++))
    fi
    
    # Task 4: Documentation task
    if task add --id "${milestone}-4" \
             --title "Update documentation" \
             --description "$description: Document all new features and APIs" \
             --milestone "$milestone" \
             --actor "tech-writer" 2>/dev/null; then
        ((task_success++))
    else
        log_warning "Failed to create task ${milestone}-4"
        ((task_failures++))
    fi
    
    if [ $task_failures -gt 0 ]; then
        log_warning "Added $task_success tasks, $task_failures failed for milestone: $milestone"
    else
        log_success "Added $task_success tasks for milestone: $milestone"
    fi
}

# -----------------------------------------------------------------------------
# FUNCTION: Display Project Summary
# -----------------------------------------------------------------------------

display_project_summary() {
    echo ""
    echo "=============================================="
    echo "  PROJECT SETUP COMPLETE: $PROJECT_NAME"
    echo "=============================================="
    echo ""
    
    # List all tasks to show the project structure
    echo "Project Tasks:"
    echo "-------------"
    task list --format table
    
    echo ""
    log_success "Project '$PROJECT_NAME' has been successfully set up!"
    echo ""
    echo "Next steps:"
    echo "  1. Review tasks with: task list"
    echo "  2. Start working on tasks with: task update --id <id> --status in_progress"
    echo "  3. Complete tasks with: task complete --id <id>"
    echo "  4. Block tasks when needed: task block --id <id> --reason <reason>"
    echo ""
}

# -----------------------------------------------------------------------------
# MAIN SCRIPT LOGIC
# -----------------------------------------------------------------------------

main() {
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                shift
                ;;
            -p|--project)
                PROJECT_NAME="$2"
                shift 2
                ;;
            -d|--data-dir)
                DATA_DIR="$2"
                shift 2
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                ;;
        esac
    done
    
    echo "=============================================="
    echo "  OPENCODE PROJECT SETUP SCRIPT"
    echo "=============================================="
    echo ""
    echo "Project Name: $PROJECT_NAME"
    echo "Data Directory: $DATA_DIR"
    echo ""
    
    # Step 1: Verify task CLI availability
    if ! verify_task_cli; then
        exit 1
    fi
    
    # Step 2: Initialize project structure
    initialize_project_structure
    
    # Step 3: Add tasks for different milestones
    # Adding tasks for Foundation milestone
    add_milestone_tasks "foundation" "Foundation phase - setting up the basic infrastructure"
    
    # Adding tasks for Development milestone
    add_milestone_tasks "development" "Development phase - building the main functionality"
    
    # Adding tasks for Testing milestone
    add_milestone_tasks "testing" "Testing phase - ensuring quality and reliability"
    
    # Adding tasks for Deployment milestone
    add_milestone_tasks "deployment" "Deployment phase - releasing to production"
    
    # Step 4: Display project summary
    display_project_summary
}

# -----------------------------------------------------------------------------
# SCRIPT ENTRY POINT
# -----------------------------------------------------------------------------

# Run main function with all script arguments
main "$@"