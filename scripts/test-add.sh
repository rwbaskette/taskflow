#!/bin/bash
# Test script for the add command
# This CLI uses positional arguments for add

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
    
    if [[ "$output" == *"Adding task"* ]] || [[ "$output" == *"Added"* ]]; then
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
    
    if [[ "$output" == *"Error"* ]] || [[ "$output" == *"error"* ]] || [[ "$output" == *"required"* ]]; then
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
echo "Testing: add command"
echo "========================================="

# Test 1: Missing title (positional argument)
echo -e "\n${YELLOW}Test 1: Missing title (no arguments)${NC}"
output=$($CLI add 2>&1)
assert_error "$output" "Should error when title is missing"

# Test 2: Valid add with title only
echo -e "\n${YELLOW}Test 2: Valid add with title only${NC}"
output=$($CLI add "Test Task" 2>&1)
assert_success "$output" "Should add task successfully with title"
assert_contains "$output" "Test Task" "Should show correct title"

# Test 3: Add with description flag
echo -e "\n${YELLOW}Test 3: Add with description flag${NC}"
output=$($CLI add "Task with Description" -d "This is a description" 2>&1)
assert_success "$output" "Should add task with description flag"
assert_contains "$output" "Task with Description" "Should show title"
assert_contains "$output" "This is a description" "Should show description"

# Test 4: Add with milestone flag
echo -e "\n${YELLOW}Test 4: Add with milestone flag${NC}"
output=$($CLI add "Task with Milestone" -m "v1.0" 2>&1)
assert_success "$output" "Should add task with milestone flag"
assert_contains "$output" "v1.0" "Should show milestone"

# Test 5: Add with actor flag
echo -e "\n${YELLOW}Test 5: Add with actor flag${NC}"
output=$($CLI add "Task with Actor" -a "developer" 2>&1)
assert_success "$output" "Should add task with actor flag"
assert_contains "$output" "developer" "Should show actor"

# Test 6: Add with all fields
echo -e "\n${YELLOW}Test 6: Add with all fields${NC}"
output=$($CLI add "Full Task" -d "Full description" -m "v2.0" -a "admin" 2>&1)
assert_success "$output" "Should add task with all fields"
assert_contains "$output" "Full Task" "Should show title"
assert_contains "$output" "Full description" "Should show description"
assert_contains "$output" "v2.0" "Should show milestone"
assert_contains "$output" "admin" "Should show actor"

# Test 7: Add multiple tasks
echo -e "\n${YELLOW}Test 7: Add multiple tasks${NC}"
output=$($CLI add "Task 1" 2>&1)
assert_success "$output" "Should add first task"
output=$($CLI add "Task 2" 2>&1)
assert_success "$output" "Should add second task"

# Test 8: Verify add command help
echo -e "\n${YELLOW}Test 8: Verify add command help${NC}"
output=$($CLI add --help 2>&1)
assert_contains "$output" "add" "Help should contain add command"

# Test 9: Add with special characters in title
echo -e "\n${YELLOW}Test 9: Special characters in title${NC}"
output=$($CLI add "Special <chars>" 2>&1)
assert_success "$output" "Should handle special characters in title"

# Test 10: Add with numeric title
echo -e "\n${YELLOW}Test 10: Numeric title${NC}"
output=$($CLI add "12345" 2>&1)
assert_success "$output" "Should handle numeric title"

# Summary
echo ""
echo "========================================="
echo "SUMMARY: add command tests"
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