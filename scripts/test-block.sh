#!/bin/bash
# Test script for the block command
# This CLI uses positional argument for task ID

# Configuration
CLI="/home/rwbaskette/tmp/task-cli"
DB_PATH="/home/rwbaskette/tmp/data/tasks.db"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counter
PASSED=0
FAILED=0

# Assertion function
assert_success() {
    local output="$1"
    local message="$2"
    
    if [[ "$output" == *"Blocking task"* ]] || [[ "$output" == *"blocked"* ]]; then
        echo -e "${GREEN}PASS${NC}: $message"
        ((PASSED++))
    else
        echo -e "${RED}FAIL${NC}: $message"
        echo "  Output: $output"
        ((FAILED++))
    fi
}

assert_error() {
    local output="$1"
    local message="$2"
    
    if [[ "$output" == *"Error"* ]] || [[ "$output" == *"error"* ]]; then
        echo -e "${GREEN}PASS${NC}: $message"
        ((PASSED++))
    else
        echo -e "${RED}FAIL${NC}: $message"
        echo "  Expected error but got: $output"
        ((FAILED++))
    fi
}

assert_contains() {
    local haystack="$1"
    local needle="$2"
    local message="$3"
    
    if [[ "$haystack" == *"$needle"* ]]; then
        echo -e "${GREEN}PASS${NC}: $message"
        ((PASSED++))
    else
        echo -e "${RED}FAIL${NC}: $message"
        echo "  Expected to contain: $needle"
        echo "  Actual:              $haystack"
        ((FAILED++))
    fi
}

# Test cases
echo "========================================="
echo "Testing: block command"
echo "========================================="

# Test 1: Missing task ID
echo -e "\n${YELLOW}Test 1: Missing task ID${NC}"
output=$($CLI block 2>&1)
assert_error "$output" "Should error when task ID is missing"

# Test 2: Provide task ID with reason
echo -e "\n${YELLOW}Test 2: Block with task ID and reason${NC}"
output=$($CLI block "task-123" -r "Waiting for dependency" 2>&1)
assert_success "$output" "Should block task with reason"
assert_contains "$output" "task-123" "Should show task ID"
assert_contains "$output" "Waiting for dependency" "Should show reason"

# Test 3: Block with task ID without reason
echo -e "\n${YELLOW}Test 3: Block with task ID only${NC}"
output=$($CLI block "task-456" 2>&1)
assert_success "$output" "Should block task without reason"

# Test 4: Block multiple tasks
echo -e "\n${YELLOW}Test 4: Block multiple tasks${NC}"
output=$($CLI block "task-789" -r "Reason 1" 2>&1)
assert_success "$output" "Should block first task"
output=$($CLI block "task-abc" -r "Reason 2" 2>&1)
assert_success "$output" "Should block second task"

# Test 5: Verify block command help
echo -e "\n${YELLOW}Test 5: Verify block command help${NC}"
output=$($CLI block --help 2>&1)
assert_contains "$output" "block" "Help should contain block command"

# Test 6: Block with special characters in ID
echo -e "\n${YELLOW}Test 6: Special characters in task ID${NC}"
output=$($CLI block "task-id-abc_123" -r "Test reason" 2>&1)
assert_success "$output" "Should handle special characters in ID"

# Test 7: Block with special characters in reason
echo -e "\n${YELLOW}Test 7: Special characters in reason${NC}"
output=$($CLI block "task-def" -r "Reason with <special> chars" 2>&1)
assert_success "$output" "Should handle special characters in reason"

# Test 8: Block with long reason
echo -e "\n${YELLOW}Test 8: Long reason text${NC}"
long_reason="This is a very long reason for blocking a task that explains in detail why the task cannot be worked on at this time."
output=$($CLI block "task-ghi" -r "$long_reason" 2>&1)
assert_success "$output" "Should handle long reason text"

# Summary
echo ""
echo "========================================="
echo "SUMMARY: block command tests"
echo "========================================="
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo "========================================="

# Exit with appropriate code
if [[ $FAILED -gt 0 ]]; then
    exit 1
else
    exit 0
fi