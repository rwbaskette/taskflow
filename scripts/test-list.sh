#!/bin/bash
# Test script for the list command

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

assert_not_contains() {
    local haystack="$1"
    local needle="$2"
    local message="$3"
    
    if [[ "$haystack" != *"$needle"* ]]; then
        echo -e "${GREEN}PASS${NC}: $message"
        ((PASSED++))
    else
        echo -e "${RED}FAIL${NC}: $message"
        echo "  Expected NOT to contain: $needle"
        echo "  Actual:                  $haystack"
        ((FAILED++))
    fi
}

# Test function that returns true/false
check_contains() {
    local haystack="$1"
    local needle="$2"
    [[ "$haystack" == *"$needle"* ]]
}

# Test cases
echo "========================================="
echo "Testing: list command"
echo "========================================="

# Test 1: Basic list command
echo -e "\n${YELLOW}Test 1: Basic list command${NC}"
output=$($CLI list 2>&1)
assert_contains "$output" "Listing tasks" "Should show listing header"

# Test 2: List with --all flag
echo -e "\n${YELLOW}Test 2: List with --all flag${NC}"
output=$($CLI list --all 2>&1)
assert_contains "$output" "Listing tasks" "Should show listing header"
assert_contains "$output" "all tasks" "Should indicate all tasks"

# Test 3: Filter by milestone
echo -e "\n${YELLOW}Test 3: Filter by milestone${NC}"
output=$($CLI list -m "v1.0" 2>&1)
if check_contains "$output" "Listing tasks"; then
    echo -e "${GREEN}PASS${NC}: Should accept milestone filter"
    ((PASSED++))
else
    echo -e "${RED}FAIL${NC}: Should accept milestone filter"
    ((FAILED++))
fi

# Test 4: Filter by actor
echo -e "\n${YELLOW}Test 4: Filter by actor${NC}"
output=$($CLI list --actor "john" 2>&1)
if check_contains "$output" "Listing tasks"; then
    echo -e "${GREEN}PASS${NC}: Should accept actor filter"
    ((PASSED++))
else
    echo -e "${RED}FAIL${NC}: Should accept actor filter"
    ((FAILED++))
fi

# Test 5: Filter by status
echo -e "\n${YELLOW}Test 5: Filter by status${NC}"
output=$($CLI list --status "todo" 2>&1)
if check_contains "$output" "Listing tasks"; then
    echo -e "${GREEN}PASS${NC}: Should accept status filter"
    ((PASSED++))
else
    echo -e "${RED}FAIL${NC}: Should accept status filter"
    ((FAILED++))
fi

# Test 6: Combined filters (milestone + actor)
echo -e "\n${YELLOW}Test 6: Combined filters${NC}"
output=$($CLI list -m "v1.0" --actor "john" 2>&1)
if check_contains "$output" "Listing tasks"; then
    echo -e "${GREEN}PASS${NC}: Should accept combined filters"
    ((PASSED++))
else
    echo -e "${RED}FAIL${NC}: Should accept combined filters"
    ((FAILED++))
fi

# Test 7: Combined filters (milestone + status)
echo -e "\n${YELLOW}Test 7: Combined filters (milestone + status)${NC}"
output=$($CLI list -m "v1.0" --status "todo" 2>&1)
if check_contains "$output" "Listing tasks"; then
    echo -e "${GREEN}PASS${NC}: Should accept milestone + status filters"
    ((PASSED++))
else
    echo -e "${RED}FAIL${NC}: Should accept milestone + status filters"
    ((FAILED++))
fi

# Test 8: Combined filters (actor + status)
echo -e "\n${YELLOW}Test 8: Combined filters (actor + status)${NC}"
output=$($CLI list --actor "john" --status "in_progress" 2>&1)
if check_contains "$output" "Listing tasks"; then
    echo -e "${GREEN}PASS${NC}: Should accept actor + status filters"
    ((PASSED++))
else
    echo -e "${RED}FAIL${NC}: Should accept actor + status filters"
    ((FAILED++))
fi

# Test 9: Verify list command help
echo -e "\n${YELLOW}Test 9: Verify list command help${NC}"
output=$($CLI list --help 2>&1)
assert_contains "$output" "list" "Help should contain list command"

# Test 10: Test all three filters together
echo -e "\n${YELLOW}Test 10: All three filters together${NC}"
output=$($CLI list -m "v1.0" --actor "john" --status "todo" 2>&1)
if check_contains "$output" "Listing tasks"; then
    echo -e "${GREEN}PASS${NC}: Should accept all three filters"
    ((PASSED++))
else
    echo -e "${RED}FAIL${NC}: Should accept all three filters"
    ((FAILED++))
fi

# Test 11: Empty list (no tasks)
echo -e "\n${YELLOW}Test 11: List with no matching tasks${NC}"
output=$($CLI list -m "nonexistent-milestone" 2>&1)
if check_contains "$output" "Listing tasks"; then
    echo -e "${GREEN}PASS${NC}: Should handle non-existent filter"
    ((PASSED++))
else
    echo -e "${RED}FAIL${NC}: Should handle non-existent filter"
    ((FAILED++))
fi

# Test 12: List command exists and runs
echo -e "\n${YELLOW}Test 12: List command is executable${NC}"
output=$($CLI list 2>&1)
if [[ -n "$output" ]]; then
    echo -e "${GREEN}PASS${NC}: List command produces output"
    ((PASSED++))
else
    echo -e "${RED}FAIL${NC}: List command should produce output"
    ((FAILED++))
fi

# Summary
echo ""
echo "========================================="
echo "SUMMARY: list command tests"
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