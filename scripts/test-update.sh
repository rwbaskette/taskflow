#!/bin/bash
# Test script for the update command
# This CLI uses positional arguments for task ID

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
    
    if [[ "$output" == *"Updating task"* ]] || [[ "$output" == *"Updated"* ]]; then
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
    
    if [[ "$output" == *"Error"* ]] || [[ "$output" == *"error"* ]] || [[ "$output" == *"not found"* ]]; then
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
echo "Testing: update command"
echo "========================================="

# Test 1: Missing task ID
echo -e "\n${YELLOW}Test 1: Missing task ID${NC}"
output=$($CLI update 2>&1)
assert_error "$output" "Should error when task ID is missing"

# Test 2: Provide task ID (no flags)
echo -e "\n${YELLOW}Test 2: Provide task ID only${NC}"
output=$($CLI update "test-id-123" 2>&1)
assert_success "$output" "Should accept task ID"

# Test 3: Update with title flag
echo -e "\n${YELLOW}Test 3: Update with title flag${NC}"
output=$($CLI update "test-id-123" -t "New Title" 2>&1)
assert_success "$output" "Should update with title flag"
assert_contains "$output" "New Title" "Should show new title"

# Test 4: Update with description flag
echo -e "\n${YELLOW}Test 4: Update with description flag${NC}"
output=$($CLI update "test-id-123" -d "New Description" 2>&1)
assert_success "$output" "Should update with description flag"
assert_contains "$output" "New Description" "Should show new description"

# Test 5: Update with milestone flag
echo -e "\n${YELLOW}Test 5: Update with milestone flag${NC}"
output=$($CLI update "test-id-123" -m "v2.0" 2>&1)
assert_success "$output" "Should update with milestone flag"
assert_contains "$output" "v2.0" "Should show new milestone"

# Test 6: Update with actor flag
echo -e "\n${YELLOW}Test 6: Update with actor flag${NC}"
output=$($CLI update "test-id-123" -a "new-actor" 2>&1)
assert_success "$output" "Should update with actor flag"
assert_contains "$output" "new-actor" "Should show new actor"

# Test 7: Update multiple flags
echo -e "\n${YELLOW}Test 7: Update multiple flags${NC}"
output=$($CLI update "test-id-123" -t "Multi Update" -d "New Desc" -m "v3.0" -a "admin" 2>&1)
assert_success "$output" "Should update with multiple flags"
assert_contains "$output" "Multi Update" "Should show new title"
assert_contains "$output" "New Desc" "Should show new description"
assert_contains "$output" "v3.0" "Should show new milestone"
assert_contains "$output" "admin" "Should show new actor"

# Test 8: Verify update command help
echo -e "\n${YELLOW}Test 8: Verify update command help${NC}"
output=$($CLI update --help 2>&1)
assert_contains "$output" "update" "Help should contain update command"

# Test 9: Update with special characters
echo -e "\n${YELLOW}Test 9: Special characters in title${NC}"
output=$($CLI update "test-id-123" -t "Special <chars>" 2>&1)
assert_success "$output" "Should handle special characters"

# Summary
echo ""
echo "========================================="
echo "SUMMARY: update command tests"
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