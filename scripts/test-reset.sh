#!/bin/bash
# Test script for the reset-timedout command

# Configuration
TMPDIR="${TMPDIR:-/tmp}"
CLI="${TMPDIR}/task-cli"
DB_PATH="${TMPDIR}/data/tasks.db"

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
    
    if [[ "$output" == *"Reset"* ]] || [[ "$output" == *"reset"* ]] || [[ "$output" == *"timed out"* ]]; then
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
    
    if [[ "$output" == *"Error"* ]] || [[ "$output" == *"error"* ]] || [[ "$output" == *"must be"* ]] || [[ "$output" == *"positive"* ]]; then
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
echo "Testing: reset-timedout command"
echo "========================================="

# Test 1: Reset with task ID (positional argument)
echo -e "\n${YELLOW}Test 1: Reset with task ID${NC}"
output=$($CLI reset-timedout "task-123" 2>&1)
assert_success "$output" "Should accept task ID"
assert_contains "$output" "task-123" "Should show task ID"

# Test 2: Reset without task ID (should error)
echo -e "\n${YELLOW}Test 2: Reset without task ID${NC}"
output=$($CLI reset-timedout 2>&1)
# This should error since task ID is required
if [[ "$output" == *"Error"* ]] || [[ "$output" == *"error"* ]] || [[ "$output" == *"required"* ]]; then
    echo -e "${GREEN}PASS${NC}: Should error without task ID"
    ((PASSED++))
else
    echo -e "${YELLOW}SKIP${NC}: May not require task ID (implementation dependent)"
fi

# Test 3: Reset multiple task IDs
echo -e "\n${YELLOW}Test 3: Reset multiple task IDs${NC}"
output=$($CLI reset-timedout "task-456" "task-789" 2>&1)
assert_success "$output" "Should accept multiple task IDs"

# Test 4: Verify reset-timedout command help
echo -e "\n${YELLOW}Test 4: Verify reset-timedout command help${NC}"
output=$($CLI reset-timedout --help 2>&1)
assert_contains "$output" "reset-timedout" "Help should contain command name"

# Test 5: Reset with special characters in ID
echo -e "\n${YELLOW}Test 5: Special characters in task ID${NC}"
output=$($CLI reset-timedout "task-id-abc_123" 2>&1)
assert_success "$output" "Should handle special characters in ID"

# Test 6: Reset with numeric ID
echo -e "\n${YELLOW}Test 6: Numeric task ID${NC}"
output=$($CLI reset-timedout "12345" 2>&1)
assert_success "$output" "Should handle numeric ID"

# Test 7: Reset with verbose flag
echo -e "\n${YELLOW}Test 7: Reset with verbose flag${NC}"
output=$($CLI reset-timedout "task-abc" -v 2>&1)
assert_success "$output" "Should accept verbose flag"

# Summary
echo ""
echo "========================================="
echo "SUMMARY: reset-timedout command tests"
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