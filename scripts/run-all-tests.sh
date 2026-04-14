#!/bin/bash
# Master test runner script

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLI="/home/rwbaskette/tmp/task-cli"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
TOTAL_PASSED=0
TOTAL_FAILED=0
TOTAL_SCRIPTS=0

# Test results storage
declare -A TEST_RESULTS

# Print header
print_header() {
    echo ""
    echo "======================================================"
    echo -e "${BLUE}$1${NC}"
    echo "======================================================"
}

# Check if CLI exists
check_cli() {
    if [[ ! -f "$CLI" ]]; then
        echo -e "${RED}ERROR: CLI binary not found at $CLI${NC}"
        echo "Please build the project first."
        exit 1
    fi
    
    # Check if CLI is executable
    if [[ ! -x "$CLI" ]]; then
        echo -e "${YELLOW}WARNING: CLI is not executable, attempting to fix...${NC}"
        chmod +x "$CLI"
    fi
    
    echo -e "${GREEN}CLI binary found: $CLI${NC}"
}

# Run a single test script
run_test() {
    local test_script="$1"
    local test_name=$(basename "$test_script" .sh)
    
    ((TOTAL_SCRIPTS++))
    
    echo ""
    print_header "Running: $test_name"
    echo "Script: $test_script"
    echo "------------------------------------------------------"
    
    # Run the test script and capture exit code
    local start_time=$(date +%s)
    "$test_script"
    local exit_code=$?
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    # Store result
    TEST_RESULTS["$test_name"]=$exit_code
    
    # Update counters
    if [[ $exit_code -eq 0 ]]; then
        ((TOTAL_PASSED++))
        echo ""
        echo -e "${GREEN}✓ $test_name PASSED${NC} (${duration}s)"
    else
        ((TOTAL_FAILED++))
        echo ""
        echo -e "${RED}✗ $test_name FAILED${NC} (${duration}s)"
    fi
}

# Print final summary
print_summary() {
    print_header "TEST EXECUTION SUMMARY"
    
    echo "Test Scripts Executed: $TOTAL_SCRIPTS"
    echo -e "Passed: ${GREEN}$TOTAL_PASSED${NC}"
    echo -e "Failed: ${RED}$TOTAL_FAILED${NC}"
    echo ""
    echo "======================================================"
    echo "DETAILED RESULTS"
    echo "======================================================"
    
    for test_name in "${!TEST_RESULTS[@]}"; do
        local result=${TEST_RESULTS[$test_name]}
        if [[ $result -eq 0 ]]; then
            echo -e "${GREEN}[PASS]${NC} $test_name"
        else
            echo -e "${RED}[FAIL]${NC} $test_name"
        fi
    done | sort
    
    echo ""
    echo "======================================================"
    
    if [[ $TOTAL_FAILED -gt 0 ]]; then
        echo -e "${RED}OVERALL STATUS: FAILED${NC}"
        echo ""
        echo "$TOTAL_FAILED test(s) failed. Please review the output above."
        return 1
    else
        echo -e "${GREEN}OVERALL STATUS: ALL TESTS PASSED${NC}"
        echo ""
        echo "All $TOTAL_SCRIPTS test scripts passed successfully!"
        return 0
    fi
}

# Main execution
main() {
    echo ""
    echo "╔══════════════════════════════════════════════════════╗"
    echo "║         MANAGE TASKS CLI - TEST SUITE              ║"
    echo "╚══════════════════════════════════════════════════════╝"
    echo ""
    echo "Starting test execution at $(date)"
    echo "Working directory: $(pwd)"
    
    # Check CLI exists
    check_cli
    
    # Find all test scripts
    TEST_SCRIPTS=(
        "$SCRIPT_DIR/test-add.sh"
        "$SCRIPT_DIR/test-update.sh"
        "$SCRIPT_DIR/test-complete.sh"
        "$SCRIPT_DIR/test-block.sh"
        "$SCRIPT_DIR/test-list.sh"
        "$SCRIPT_DIR/test-reset.sh"
    )
    
    # Check if all test scripts exist
    MISSING_SCRIPTS=()
    for script in "${TEST_SCRIPTS[@]}"; do
        if [[ ! -f "$script" ]]; then
            MISSING_SCRIPTS+=("$script")
        fi
    done
    
    if [[ ${#MISSING_SCRIPTS[@]} -gt 0 ]]; then
        echo -e "${RED}ERROR: Missing test scripts:${NC}"
        for script in "${MISSING_SCRIPTS[@]}"; do
            echo "  - $script"
        done
        exit 1
    fi
    
    # Make scripts executable
    for script in "${TEST_SCRIPTS[@]}"; do
        chmod +x "$script"
    done
    
    # Run all tests
    for script in "${TEST_SCRIPTS[@]}"; do
        run_test "$script"
    done
    
    # Print summary and exit with appropriate code
    print_summary
    exit $?
}

# Run main function
main "$@"