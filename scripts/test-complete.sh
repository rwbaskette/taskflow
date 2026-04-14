#!/bin/bash
# Test script for the complete command
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
    
    if [[ "$output" == *"Completing task"* ]] || [[ "$output" == *"completed"* ]]; then
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
echo "Testing: complete command"
echo "========================================="

# Test 1: Missing task ID
echo -e "\n${YELLOW}Test 1: Missing task ID${NC}"
output=$($CLI complete 2>&1)
assert_error "$output" "Should error when task ID is missing"

# Test 2: Provide task ID
echo -e "\n${YELLOW}Test 2: Provide task ID${NC}"
output=$($CLI complete "task-123" 2>&1)
assert_success "$output" "Should accept task ID"
assert_contains "$output" "task-123" "Should show task ID"

# Test 3: Complete multiple tasks
echo -e "\n${YELLOW}Test 3: Complete multiple tasks${NC}"
output=$($CLI complete "task-456" 2>&1)
assert_success "$output" "Should complete first task"
output=$($CLI complete "task-789" 2>&1)
assert_success "$output" "Should complete second task"

# Test 4: Verify complete command help
echo -e "\n${YELLOW}Test 4: Verify complete command help${NC}"
output=$($CLI complete --help 2>&1)
assert_contains "$output" "complete" "Help should contain complete command"

# Test 5: Complete with special characters in ID
echo -e "\n${YELLOW}Test 5: Special characters in task ID${NC}"
output=$($CLI complete "task-id-abc_123" 2>&1)
assert_success "$output" "Should handle special characters in ID"

# Test 6: Complete with numeric ID
echo -e "\n${YELLOW}Test 6: Numeric task ID${NC}"
output=$($CLI complete "12345" 2>&1)
assert_success "$output" "Should handle numeric ID"

# Summary
echo ""
echo "========================================="
echo "SUMMARY: complete command tests"
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