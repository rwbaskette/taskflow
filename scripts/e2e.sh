#!/usr/bin/env bash
# scripts/e2e.sh — standalone, dependency-free end-to-end test for taskflow.
#
# Complements the Go e2e tests under tests/cli (tag-gated). Requires only
# bash, go, and coreutils. Runs the full task lifecycle plus error paths
# against a throwaway database.
#
# Verified against the code before writing:
#   - Isolation env var is TASKFLOW_DIR (cmd/init.go, internal/db/db.go):
#     when set and non-empty, every command resolves the DB at
#     abs($TASKFLOW_DIR/tasks.db) and the .taskflow anchor walk is skipped
#     entirely, so the repo's own .taskflow anchor can never be touched.
#   - Pinned error codes (internal/clierr/errors.go, cmd/unblock_e2e_test.go):
#     INVALID_STATUS_TRANSITION, RESOURCE_NOT_FOUND, INVALID_ARGUMENT,
#     MISSING_ARGUMENT. Errors print JSON {"status":"error",...} to stderr,
#     exit 1. Anchor errors are plain text, exit 2.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# ---------------------------------------------------------------------------
# Scratch environment: one temp dir for the binary and one for the DB, both
# cleaned up on exit. Kept under $ROOT/tmp so nothing is written outside the
# project tree.
# ---------------------------------------------------------------------------
SCRATCH="$(mktemp -d "${ROOT}/tmp/e2e-scratch.XXXXXX")"
export TASKFLOW_DIR="$SCRATCH/taskflow-data"
trap 'rm -rf "$SCRATCH"' EXIT

BIN="$SCRATCH/taskflow"

# CLI runner: always executes with cwd = scratch dir (belt and braces: even
# without TASKFLOW_DIR the anchor walk would start inside the scratch tree,
# not the repo). Captures merged stdout+stderr into TF_OUT and the exit code
# into TF_RC.
TF_RC=0
TF_OUT=""
tf() {
  TF_RC=0
  TF_OUT="$(cd "$SCRATCH" && "$BIN" "$@" 2>&1)" || TF_RC=$?
}

# ---------------------------------------------------------------------------
# Assertion helpers (fail fast: echo the failing command and actual output).
# ---------------------------------------------------------------------------
COUNT=0
fail() {
  echo "FAIL: $1" >&2
  echo "  command: $2" >&2
  echo "  exit: $3 (expected $4)" >&2
  echo "  output:" >&2
  printf '%s\n' "$TF_OUT" | sed 's/^/    /' >&2
  exit 1
}

# assert_ok <desc> [grep-pattern] <cmd...> : exit 0, and output matches
# the ERE pattern when one is given.
assert_ok() {
  local desc="$1" pattern="$2"
  shift 2
  tf "$@"
  if [ "$TF_RC" -ne 0 ]; then
    fail "$desc" "$*" "$TF_RC" 0
  fi
  if [ -n "$pattern" ] && ! printf '%s' "$TF_OUT" | grep -qE -- "$pattern"; then
    TF_RC="pattern-miss"
    fail "$desc (pattern: $pattern)" "$*" "no match" "match"
  fi
  COUNT=$((COUNT + 1))
}

# assert_fail <desc> <grep-pattern> <cmd...> : exit non-zero, and output
# matches the ERE pattern (error code, message fragment).
assert_fail() {
  local desc="$1" pattern="$2"
  shift 2
  tf "$@"
  if [ "$TF_RC" -eq 0 ]; then
    TF_RC="zero"
    fail "$desc (expected failure)" "$*" 0 "non-zero"
  fi
  if ! printf '%s' "$TF_OUT" | grep -qE -- "$pattern"; then
    TF_RC="pattern-miss"
    fail "$desc (pattern: $pattern)" "$*" "no match" "match"
  fi
  COUNT=$((COUNT + 1))
}

# ---------------------------------------------------------------------------
# 1. Build the binary once.
# ---------------------------------------------------------------------------
echo "== building taskflow binary =="
(cd "$ROOT" && go build -o "$BIN" .)

# ---------------------------------------------------------------------------
# 2. CLI basics: help and version.
# ---------------------------------------------------------------------------
echo "== help / version =="
assert_ok "--help exits 0" '.' --help
assert_ok "--version exits 0" 'taskflow version' --version
assert_ok "version exits 0" 'taskflow version' version

# ---------------------------------------------------------------------------
# 3. Init: TASKFLOW_DIR set -> creates $TASKFLOW_DIR/tasks.db, no anchor.
# ---------------------------------------------------------------------------
echo "== init =="
assert_ok "init creates the db and prints its path" "db: .*/tasks\.db" init

# ---------------------------------------------------------------------------
# 4. Add two tasks; JSON-argument style.
# ---------------------------------------------------------------------------
echo "== add =="
# Note: output is multiline and grep is line-oriented, so each marker
# ("Task added successfully:", "Status: todo") is asserted separately.
assert_ok "add task 1" 'Task added successfully:' \
  add '{"id":"e2e-1","milestone":"e2e-m1","title":"First task","actor":"e2e","description":"Do the first thing"}'
assert_ok "add task 1 status is todo" '"Status": "todo"' list '{"id":"e2e-1"}'
assert_ok "add task 2" 'Task added successfully:' \
  add '{"id":"e2e-2","milestone":"e2e-m1","title":"Second task","actor":"e2e","description":"Do the second thing"}'
assert_ok "add task 2 status is todo" '"Status": "todo"' list '{"id":"e2e-2"}'

# ---------------------------------------------------------------------------
# 5. List: both tasks present, total 2.
# ---------------------------------------------------------------------------
echo "== list =="
assert_ok "list shows total 2 and both ids" '"total": 2' list '{}'
assert_ok "list contains task 1" '"ID": "e2e-1"' list '{}'
assert_ok "list contains task 2" '"ID": "e2e-2"' list '{}'

# ---------------------------------------------------------------------------
# 6. Update and complete task 1.
# ---------------------------------------------------------------------------
echo "== update / complete =="
assert_ok "update task 1 title" 'Task updated successfully:' \
  update '{"id":"e2e-1","title":"First task (renamed)"}'
assert_ok "update took effect" 'First task \(renamed\)' list '{"id":"e2e-1"}'
assert_ok "complete task 1" 'Status: done' complete '{"id":"e2e-1"}'
assert_ok "list status=done shows only task 1" '"Status": "done"' list '{"status":"done"}'

# ---------------------------------------------------------------------------
# 7. Block / unblock lifecycle on task 2.
# ---------------------------------------------------------------------------
echo "== block / unblock =="
assert_ok "block task 2 with reason" 'Status: blocked' \
  block '{"id":"e2e-2","reason":"waiting on review-777"}'
assert_ok "list status=blocked shows task 2" '"ID": "e2e-2"' list '{"status":"blocked"}'
assert_ok "blocked list carries the block reason" 'review-777' list '{"status":"blocked"}'
assert_ok "unblock task 2" 'Task unblocked successfully:' unblock '{"id":"e2e-2"}'
assert_ok "unblock task 2 status is todo" '"Status": "todo"' list '{"id":"e2e-2"}'

# Error paths (codes pinned from cmd/unblock_e2e_test.go):
assert_fail "unblock a non-blocked task fails" 'INVALID_STATUS_TRANSITION' unblock '{"id":"e2e-2"}'
assert_fail "unblock a nonexistent task fails" 'RESOURCE_NOT_FOUND' unblock '{"id":"no-such-task"}'

# ---------------------------------------------------------------------------
# 8. Delete task 2 (soft delete), then re-list.
# ---------------------------------------------------------------------------
echo "== delete =="
# One call only: the delete is a soft delete, so a second delete of the
# same id would fail with RESOURCE_NOT_FOUND. Assert both markers here.
assert_ok "delete task 2 (soft delete, shows Deleted On)" 'Deleted On:' delete '{"id":"e2e-2"}'
assert_ok "list shows total 1" '"total": 1' list '{}'
if printf '%s' "$TF_OUT" | grep -q '"ID": "e2e-2"'; then
  TF_RC="still-listed"
  fail "deleted task 2 must not appear in list" "list '{}'" 0 0
fi
assert_ok "task 1 survives the delete" '"ID": "e2e-1"' list '{}'

# ---------------------------------------------------------------------------
# 9. Error paths: bad JSON, missing required fields.
# ---------------------------------------------------------------------------
echo "== error paths =="
# Invalid JSON: ParseJSON fails; HandleError wraps it as a JSON error body
# ("status":"error") on stderr, exit 1.
assert_fail "list with invalid JSON fails with JSON error" '"status": ?"error"' list '{not json'
# Missing required fields on add: id/title/milestone empty -> INVALID_ARGUMENT;
# missing description -> MISSING_ARGUMENT. Either code is a contract failure.
assert_fail "add with missing title fails" '"status": ?"error"' \
  add '{"id":"e2e-bad","milestone":"m","actor":"a","description":"d"}'
assert_fail "add with missing description fails" 'MISSING_ARGUMENT' \
  add '{"id":"e2e-bad2","milestone":"m","title":"t","actor":"a"}'

# ---------------------------------------------------------------------------
# 10. where: with TASKFLOW_DIR set it reports the env override, walk skipped.
# ---------------------------------------------------------------------------
echo "== where =="
assert_ok "where reports the env-resolved db" 'db: .*/tasks\.db' where
assert_ok "where reports the env anchor kind" 'anchor: env' where

# ---------------------------------------------------------------------------
# Summary.
# ---------------------------------------------------------------------------
echo "e2e: ALL PASSED ($COUNT assertions)"
exit 0
