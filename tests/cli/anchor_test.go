//go:build integration

package cli_test

// Integration tests for the .taskflow anchor feature (taskflow-init-anchor.md
// sections 4, 6, and 7): init, the error contract, pointer anchors, --force,
// --target, where, and the concurrency smoke test.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestAnchorErrorContract_NoAnchorAnywhere verifies the section 6 not-found
// contract: exit 2, plain text, the walked chain up to the filesystem root,
// the init remedy, the TASKFLOW_DIR note, and nothing created.
func TestAnchorErrorContract_NoAnchorAnywhere(t *testing.T) {
	dir := tempWorkdir(t)
	r := runCLI(t, dir, "add", addTaskJSON("anchor-1"))

	if r.exitCode != 2 {
		t.Errorf("expected exit 2, got %d; output:\n%s", r.exitCode, r.combined)
	}
	if !strings.HasPrefix(r.stderr, "taskflow: no .taskflow anchor found\n") {
		t.Errorf("stderr must start with the not-found message, got:\n%s", r.stderr)
	}
	assertContains(t, r.stderr, "searched:", "chain header")

	chain := chainFromDir(dir)
	for i, d := range chain {
		if i == len(chain)-1 {
			assertContains(t, r.stderr, "  "+d+" (filesystem root reached, no .taskflow)\n", "final chain line")
		} else {
			assertContains(t, r.stderr, "  "+d+"\n", "chain entry "+d)
		}
	}
	assertContains(t, r.stderr, "remedy: run `taskflow init` in the project root (--target <path> shares one DB across checkouts)", "init remedy with the --target sharing sentence")
	assertContains(t, r.stderr, "or set TASKFLOW_DIR to override", "TASKFLOW_DIR note")

	if anchorExistsAt(t, dir) {
		t.Error("nothing must be created: found a .taskflow entry at the workdir")
	}
}

// TestAnchorErrorContract_PointerErrors verifies the section 6 pointer-error
// contracts: malformed and dangling both exit 2, name the pointer file path,
// and carry their own remedies.
func TestAnchorErrorContract_PointerErrors(t *testing.T) {
	t.Run("malformed pointer", func(t *testing.T) {
		dir := tempWorkdir(t)
		pointerPath := filepath.Join(dir, ".taskflow")
		if err := os.WriteFile(pointerPath, []byte("garbage\n"), 0o644); err != nil {
			t.Fatalf("setup: %v", err)
		}

		r := runCLI(t, dir, "add", addTaskJSON("anchor-2"))
		if r.exitCode != 2 {
			t.Errorf("expected exit 2, got %d; output:\n%s", r.exitCode, r.combined)
		}
		assertContains(t, r.stderr, pointerPath, "stderr names the pointer file path")
		assertContains(t, r.stderr, "malformed", "reason malformed")
		assertContains(t, r.stderr, "fix by hand, or run `taskflow init --force` below it to shadow it", "malformed remedy")
	})

	t.Run("dangling pointer", func(t *testing.T) {
		dir := tempWorkdir(t)
		pointerPath := writePointerFile(t, dir, filepath.Join(dir, "missing", "tasks.db"))

		r := runCLI(t, dir, "add", addTaskJSON("anchor-3"))
		if r.exitCode != 2 {
			t.Errorf("expected exit 2, got %d; output:\n%s", r.exitCode, r.combined)
		}
		assertContains(t, r.stderr, pointerPath, "stderr names the pointer file path")
		assertContains(t, r.stderr, "dangling (target missing)", "reason dangling")
		assertContains(t, r.stderr, "remedy: run `taskflow init` to create or repair the database", "dangling remedy")
	})
}

// TestDanglingPointerLifecycle verifies the committed-pointer scenario: the
// pointer exists, the DB does not; commands fail with the dangling contract,
// init repairs, and the cycle repeats after the target DB is deleted.
func TestDanglingPointerLifecycle(t *testing.T) {
	dir := tempWorkdir(t)
	dbPath := filepath.Join(dir, "remote", "tasks.db")
	writePointerFile(t, dir, dbPath)

	r := runCLI(t, dir, "add", addTaskJSON("life-1"))
	if r.exitCode != 2 {
		t.Errorf("expected exit 2 with a dangling pointer, got %d; output:\n%s", r.exitCode, r.combined)
	}
	assertContains(t, r.stderr, "dangling (target missing)", "dangling contract")

	ri := runCLI(t, dir, "init")
	if ri.exitCode != 0 {
		t.Fatalf("init must repair a dangling pointer (exit %d): %s", ri.exitCode, ri.combined)
	}
	if _, err := os.Stat(dbPath); err != nil {
		t.Errorf("expected the DB at %s after repair: %v", dbPath, err)
	}

	r = runCLI(t, dir, "add", addTaskJSON("life-1"))
	assertExitZero(t, r, "add succeeds after repair")

	// Delete the target DB: dangling again; init repairs again.
	if err := os.Remove(dbPath); err != nil {
		t.Fatalf("setup: failed to remove target DB: %v", err)
	}

	r = runCLI(t, dir, "add", addTaskJSON("life-2"))
	if r.exitCode != 2 {
		t.Errorf("expected exit 2 after the target DB was deleted, got %d; output:\n%s", r.exitCode, r.combined)
	}
	assertContains(t, r.stderr, "dangling (target missing)", "dangling contract after deletion")

	ri = runCLI(t, dir, "init")
	if ri.exitCode != 0 {
		t.Fatalf("init must repair again (exit %d): %s", ri.exitCode, ri.combined)
	}

	r = runCLI(t, dir, "add", addTaskJSON("life-2"))
	assertExitZero(t, r, "add succeeds after the second repair")
}

// TestRuntimeCommandsDoNotRecreateDeletedDB verifies the missing-DB walk
// guard (review FIX 1): when an anchor resolves but the DB FILE is gone,
// runtime commands exit 2 with the missing-DB contract and create nothing;
// only `taskflow init` recreates the DB.
func TestRuntimeCommandsDoNotRecreateDeletedDB(t *testing.T) {
	t.Run("dir anchor with deleted tasks.db", func(t *testing.T) {
		dir := tempWorkdir(t)
		initProject(t, dir)
		dbPath := filepath.Join(dir, ".taskflow", "tasks.db")
		if err := os.Remove(dbPath); err != nil {
			t.Fatalf("setup: failed to remove tasks.db: %v", err)
		}

		r := runCLI(t, dir, "add", addTaskJSON("missing-db-1"))
		if r.exitCode != 2 {
			t.Errorf("expected exit 2 with the DB file deleted, got %d; output:\n%s", r.exitCode, r.combined)
		}
		assertContains(t, r.stderr, "taskflow: "+dbPath+" is missing", "missing-DB message names the DB path")
		assertContains(t, r.stderr, "remedy: run `taskflow init` to create or repair the database", "init remedy")
		if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
			t.Errorf("add must not recreate a deleted tasks.db (stat err=%v)", err)
		}

		ri := runCLI(t, dir, "init")
		if ri.exitCode != 0 {
			t.Fatalf("init must recreate a missing tasks.db (exit %d): %s", ri.exitCode, ri.combined)
		}
		if _, err := os.Stat(dbPath); err != nil {
			t.Errorf("expected the DB at %s after repair: %v", dbPath, err)
		}

		r = runCLI(t, dir, "add", addTaskJSON("missing-db-1"))
		assertExitZero(t, r, "add works again after the repair")
	})

	t.Run("pointer to an anchor home with deleted tasks.db", func(t *testing.T) {
		dir := tempWorkdir(t)
		main := filepath.Join(dir, "main")
		if err := os.MkdirAll(main, 0o755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		initProject(t, main)
		// The pointer names the anchor home (the canonical form); taskflow
		// appends tasks.db. (A pointer naming the tasks.db file itself makes
		// the pointer target missing when the file is deleted, which is the
		// dangling contract, not the missing-DB case.)
		home := filepath.Join(main, ".taskflow")
		target := filepath.Join(home, "tasks.db")
		pointerPath := writePointerFile(t, dir, home)

		// Sanity: the pointer resolves and the DB is openable.
		ra := runCLI(t, dir, "add", addTaskJSON("missing-db-2"))
		assertExitZero(t, ra, "add through the pointer before deletion")

		if err := os.Remove(target); err != nil {
			t.Fatalf("setup: failed to remove the target DB: %v", err)
		}

		r := runCLI(t, dir, "add", addTaskJSON("missing-db-3"))
		if r.exitCode != 2 {
			t.Errorf("expected exit 2 with the pointer target DB deleted, got %d; output:\n%s", r.exitCode, r.combined)
		}
		assertContains(t, r.stderr, "taskflow: "+target+" is missing", "missing-DB message names the DB path")
		assertContains(t, r.stderr, "remedy: run `taskflow init` to create or repair the database", "init remedy")
		if _, err := os.Stat(target); !os.IsNotExist(err) {
			t.Errorf("add must not recreate a deleted target DB (stat err=%v)", err)
		}

		ri := runCLI(t, dir, "init")
		if ri.exitCode != 0 {
			t.Fatalf("init must recreate the pointer target DB (exit %d): %s", ri.exitCode, ri.combined)
		}
		if _, err := os.Stat(target); err != nil {
			t.Errorf("expected the DB at %s after repair: %v", target, err)
		}
		pointerBytes := readFileBytes(t, pointerPath)
		assertContains(t, string(pointerBytes), "database: "+home, "the pointer file is untouched")

		r = runCLI(t, dir, "add", addTaskJSON("missing-db-3"))
		assertExitZero(t, r, "add works again after the repair")
	})
}

// TestInitFromChildHealsDanglingParentPointer verifies the healing flow: a
// dangling pointer at a parent directory is repaired by plain `init` from a
// child directory (no --force): the DB is created at the target, the pointer
// file stays untouched, and subsequent adds from the child work.
func TestInitFromChildHealsDanglingParentPointer(t *testing.T) {
	dir := tempWorkdir(t)
	dbPath := filepath.Join(dir, "remote", "tasks.db")
	pointerPath := writePointerFile(t, dir, dbPath)
	before := readFileBytes(t, pointerPath)

	child := filepath.Join(dir, "child")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	ri := runCLI(t, child, "init")
	if ri.exitCode != 0 {
		t.Fatalf("init from the child must heal the parent's dangling pointer (exit %d): %s", ri.exitCode, ri.combined)
	}
	if _, err := os.Stat(dbPath); err != nil {
		t.Errorf("expected the DB created at the target %s: %v", dbPath, err)
	}
	assertBytesEqual(t, "the pointer file after healing", before, readFileBytes(t, pointerPath))

	// Sanity: the pointer still resolves, and no anchor was created in the child.
	if anchorExistsAt(t, child) {
		t.Error("healing must not create a second anchor in the child directory")
	}

	r := runCLI(t, child, "add", addTaskJSON("heal-child"))
	assertExitZero(t, r, "add from the child after healing")

	lr := listAll(t, child)
	assertContains(t, lr.stdout, "heal-child", "the task added from the child is listed")
}

// TestInitWalkBeforeCreateRefusal verifies the guardrail: with an anchor above
// the cwd, plain init reports the existing anchor, creates nothing, and names
// --force; commands from the child still resolve to the parent DB.
func TestInitWalkBeforeCreateRefusal(t *testing.T) {
	dir := tempWorkdir(t)
	initProject(t, dir)
	addTask(t, dir, addTaskJSON("walk-1"))

	child := filepath.Join(dir, "sub")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	r := runCLI(t, child, "init")
	if r.exitCode != 0 {
		t.Fatalf("expected exit 0, got %d; output:\n%s", r.exitCode, r.combined)
	}
	assertContains(t, r.stdout, "anchor already exists at "+dir, "refusal names the existing anchor")
	assertContains(t, r.stdout, "no new anchor created", "refusal message")
	assertContains(t, r.stdout, "--force", "refusal names --force")
	if anchorExistsAt(t, child) {
		t.Error("no new .taskflow must be created in the child directory")
	}

	lr := runCLI(t, child, "list", `{"all":true}`)
	assertExitZero(t, lr, "list from the child")
	assertContains(t, lr.stdout, "walk-1", "the parent's task is visible from the child")
}

// TestInitResidualScatterInSubdir verifies the disclosed residual scatter:
// init from a subdirectory with no anchor anywhere creates the anchor there.
func TestInitResidualScatterInSubdir(t *testing.T) {
	dir := tempWorkdir(t)
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	r := runCLI(t, sub, "init")
	assertExitZero(t, r, "init in a fresh subdirectory")
	assertContains(t, r.stdout, filepath.Join(sub, ".taskflow"), "stdout reports the anchor path")

	dbPath := filepath.Join(sub, ".taskflow", "tasks.db")
	if _, err := os.Stat(dbPath); err != nil {
		t.Errorf("expected the DB at %s: %v", dbPath, err)
	}
}

// TestInitProtectedDirGuard verifies the protected-directory guard against a
// temp HOME: init refuses without --force and creates the anchor with --force.
// The plain refusal at the real filesystem root is covered by
// TestInitRefusesFilesystemRoot (the guard fires before any write, so no
// privileges are needed); only the --force-at-real-root case is infeasible
// for tests, because it would write a real anchor at /.
func TestInitProtectedDirGuard(t *testing.T) {
	tmpHome := tempWorkdir(t)
	t.Setenv("HOME", tmpHome)

	r := runCLI(t, tmpHome, "init")
	if r.exitCode != 2 {
		t.Errorf("expected exit 2 at the home directory, got %d; output:\n%s", r.exitCode, r.combined)
	}
	assertContains(t, r.stderr, "refusing to create an anchor here; pass --force to override", "refusal message")
	if anchorExistsAt(t, tmpHome) {
		t.Error("refused init must not create a .taskflow entry")
	}

	rf := runCLI(t, tmpHome, "init", "--force")
	if rf.exitCode != 0 {
		t.Fatalf("expected exit 0 with --force, got %d; output:\n%s", rf.exitCode, rf.combined)
	}
	dbPath := filepath.Join(tmpHome, ".taskflow", "tasks.db")
	if _, err := os.Stat(dbPath); err != nil {
		t.Errorf("expected the anchor created with --force at %s: %v", dbPath, err)
	}
}

// TestInitRefusesFilesystemRoot verifies the protected-dir guard at the real
// filesystem root: the guard fires before any write, so no privileges are
// needed. Only the --force-at-real-root case is infeasible for tests (it
// would write a real anchor at /).
func TestInitRefusesFilesystemRoot(t *testing.T) {
	_, statErr := os.Stat("/.taskflow")
	rootAnchorExisted := statErr == nil

	r := runCLI(t, "/", "init")
	if r.exitCode != 2 {
		t.Errorf("expected exit 2 at the filesystem root, got %d; output:\n%s", r.exitCode, r.combined)
	}
	assertContains(t, r.stderr, "refusing to create an anchor here; pass --force to override", "refusal message")

	// Guard the assertion: if /.taskflow already exists in this environment,
	// the stat check is skipped (the refusal itself is still asserted above).
	if !rootAnchorExisted {
		if _, err := os.Stat("/.taskflow"); !os.IsNotExist(err) {
			t.Errorf("refused init must not create /.taskflow (stat err=%v)", err)
		}
	}
}

// TestStrayLatchAndRecovery verifies the residual latch risk of section 5: a
// nested unanchored project latches the outer anchor; init --force at the
// nested project root restores correct resolution.
func TestStrayLatchAndRecovery(t *testing.T) {
	dir := tempWorkdir(t)
	initProject(t, dir)
	addTask(t, dir, addTaskJSON("latch-parent"))

	nested := filepath.Join(dir, "project")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Latch: the nested dir with no own anchor writes into the parent DB.
	r := runCLI(t, nested, "add", addTaskJSON("latch-nested"))
	assertExitZero(t, r, "add from the nested dir")
	plr := listAll(t, dir)
	assertContains(t, plr.stdout, "latch-nested", "the task latched into the parent DB")

	// Plain init refuses to create a second anchor (guardrail).
	ri := runCLI(t, nested, "init")
	if ri.exitCode != 0 {
		t.Errorf("expected exit 0, got %d; output:\n%s", ri.exitCode, ri.combined)
	}
	assertContains(t, ri.stdout, "no new anchor created", "plain init refuses")
	if anchorExistsAt(t, nested) {
		t.Error("plain init must not create a nested anchor while the outer one is found")
	}

	// Recovery: init --force at the nested project root creates its own anchor.
	rf := runCLI(t, nested, "init", "--force")
	assertExitZero(t, rf, "init --force at the nested project root")
	if !anchorExistsAt(t, nested) {
		t.Error("init --force must create the nested anchor")
	}

	r = runCLI(t, nested, "add", addTaskJSON("latch-recovered"))
	assertExitZero(t, r, "add from the nested dir after recovery")

	nlr := listAll(t, nested)
	assertContains(t, nlr.stdout, "latch-recovered", "the task lands in the nested DB")

	plr = listAll(t, dir)
	assertNotContains(t, plr.stdout, "latch-recovered", "the parent no longer sees the nested task")
}

// TestWhereNestedAnchorFoundFirst verifies that a nested project anchored at
// its own root resolves to its own anchor, while an unanchored sibling still
// resolves to the parent anchor.
func TestWhereNestedAnchorFoundFirst(t *testing.T) {
	dir := tempWorkdir(t)
	initProject(t, dir)

	nested := filepath.Join(dir, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	rf := runCLI(t, nested, "init", "--force")
	assertExitZero(t, rf, "init --force at the nested root")

	wr := runCLI(t, nested, "where")
	assertExitZero(t, wr, "where from the nested dir")
	assertContains(t, wr.stdout, "db: "+filepath.Join(nested, ".taskflow", "tasks.db"), "the nested DBPath is reported")
	assertContains(t, wr.stdout, "anchor: dir", "anchor kind")

	sib := filepath.Join(dir, "sibling")
	if err := os.MkdirAll(sib, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	ws := runCLI(t, sib, "where")
	assertExitZero(t, ws, "where from the unanchored sibling")
	assertContains(t, ws.stdout, "db: "+filepath.Join(dir, ".taskflow", "tasks.db"), "the parent DBPath is reported from the sibling")
}

// TestForceShadowAboveCwd verifies --force on an anchor above the cwd: the
// child gets a shadow anchor; the two DBs are independent.
func TestForceShadowAboveCwd(t *testing.T) {
	dir := tempWorkdir(t)
	initProject(t, dir)
	addTask(t, dir, addTaskJSON("force-outer"))

	child := filepath.Join(dir, "child")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	rf := runCLI(t, child, "init", "--force")
	assertExitZero(t, rf, "init --force in the child")
	if !anchorExistsAt(t, child) {
		t.Error("expected a .taskflow anchor in the child")
	}

	r := runCLI(t, child, "add", addTaskJSON("force-inner"))
	assertExitZero(t, r, "add in the child")

	clr := listAll(t, child)
	assertContains(t, clr.stdout, "force-inner", "the child DB has its task")
	assertNotContains(t, clr.stdout, "force-outer", "the child DB does not see the parent task")

	plr := listAll(t, dir)
	assertContains(t, plr.stdout, "force-outer", "the parent DB keeps its task")
	assertNotContains(t, plr.stdout, "force-inner", "the parent DB does not see the child task")
}

// TestForceReplacesPointerAtCwd verifies that --force on a pointer anchor at
// the cwd replaces the pointer with a directory anchor and leaves the
// pointer's target DB untouched.
func TestForceReplacesPointerAtCwd(t *testing.T) {
	dir := tempWorkdir(t)
	main := filepath.Join(dir, "main")
	if err := os.MkdirAll(main, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	initProject(t, main)
	addTask(t, main, addTaskJSON("ptr-target"))

	wt := filepath.Join(dir, "wt")
	if err := os.MkdirAll(wt, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	writePointerFile(t, wt, filepath.Join(main, ".taskflow", "tasks.db"))

	// Sanity: an add through the pointer lands in the main DB.
	r := runCLI(t, wt, "add", addTaskJSON("ptr-shared"))
	assertExitZero(t, r, "add through the pointer")

	targetDB := filepath.Join(main, ".taskflow", "tasks.db")
	before := readFileBytes(t, targetDB)

	rf := runCLI(t, wt, "init", "--force")
	assertExitZero(t, rf, "init --force at the pointer anchor")

	fi, err := os.Stat(filepath.Join(wt, ".taskflow"))
	if err != nil || !fi.IsDir() {
		t.Errorf("expected .taskflow to be a directory after --force, err=%v", err)
	}
	assertBytesEqual(t, "the pointer's target DB", before, readFileBytes(t, targetDB))

	mlr := listAll(t, main)
	assertContains(t, mlr.stdout, "ptr-target", "the target DB still contains its task")

	// The new local DB is separate from the pointer target.
	r = runCLI(t, wt, "add", addTaskJSON("ptr-local"))
	assertExitZero(t, r, "add into the new local DB")

	wlr := listAll(t, wt)
	assertContains(t, wlr.stdout, "ptr-local", "the local DB has its task")
	assertNotContains(t, wlr.stdout, "ptr-target", "the local DB is separate from the former target")
}

// TestForceReportAndNoOpOnDataBearingAnchor verifies that --force on a valid
// directory anchor at the cwd is a report-and-no-op: exit 0, tasks.db
// byte-identical, data intact.
func TestForceReportAndNoOpOnDataBearingAnchor(t *testing.T) {
	dir := tempWorkdir(t)
	initProject(t, dir)
	addTask(t, dir, addTaskJSON("noop-1"))

	dbPath := filepath.Join(dir, ".taskflow", "tasks.db")
	before := readFileBytes(t, dbPath)

	rf := runCLI(t, dir, "init", "--force")
	assertExitZero(t, rf, "init --force on a directory anchor at the cwd")

	assertBytesEqual(t, "tasks.db after --force", before, readFileBytes(t, dbPath))

	lr := listAll(t, dir)
	assertContains(t, lr.stdout, "noop-1", "the task is still listed")
}

// TestForceShadowOverMalformedPointerAboveCwd verifies the malformed-pointer
// remedy: init --force below a malformed pointer creates a shadow anchor and
// leaves the pointer file untouched.
func TestForceShadowOverMalformedPointerAboveCwd(t *testing.T) {
	dir := tempWorkdir(t)
	pointerPath := filepath.Join(dir, ".taskflow")
	if err := os.WriteFile(pointerPath, []byte("garbage\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	before := readFileBytes(t, pointerPath)

	child := filepath.Join(dir, "sub")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	rf := runCLI(t, child, "init", "--force")
	assertExitZero(t, rf, "init --force below the malformed pointer")
	if !anchorExistsAt(t, child) {
		t.Error("expected a shadow directory anchor in the child")
	}
	assertBytesEqual(t, "the malformed pointer file", before, readFileBytes(t, pointerPath))
}

// TestInitTargetValidationAndWrite covers the --target branch of init: target
// kind validation before anything is written, create-if-missing, the exact
// pointer line, relative targets, unwritable parents, and failed repair.
func TestInitTargetValidationAndWrite(t *testing.T) {
	t.Run("plain directory target refused", func(t *testing.T) {
		dir := tempWorkdir(t)
		plain := filepath.Join(dir, "plaindir")
		if err := os.MkdirAll(plain, 0o755); err != nil {
			t.Fatalf("setup: %v", err)
		}

		r := runCLI(t, dir, "init", "--target", plain)
		if r.exitCode != 2 {
			t.Errorf("expected exit 2, got %d; output:\n%s", r.exitCode, r.combined)
		}
		assertContains(t, r.combined, "invalid", "clear error")
		if anchorExistsAt(t, dir) {
			t.Error("nothing must be written: no pointer at the cwd")
		}
		if entries, err := os.ReadDir(plain); err != nil || len(entries) != 0 {
			t.Errorf("nothing must be written at the target (entries=%d, err=%v)", len(entries), err)
		}
	})

	t.Run("plain file target refused", func(t *testing.T) {
		dir := tempWorkdir(t)
		plain := filepath.Join(dir, "plain.txt")
		if err := os.WriteFile(plain, []byte("hello"), 0o644); err != nil {
			t.Fatalf("setup: %v", err)
		}

		r := runCLI(t, dir, "init", "--target", plain)
		if r.exitCode != 2 {
			t.Errorf("expected exit 2, got %d; output:\n%s", r.exitCode, r.combined)
		}
		assertContains(t, r.combined, "invalid", "clear error")
		if anchorExistsAt(t, dir) {
			t.Error("nothing must be written: no pointer at the cwd")
		}
	})

	t.Run("pointer-file target refused", func(t *testing.T) {
		dir := tempWorkdir(t)
		other := filepath.Join(dir, "other")
		if err := os.MkdirAll(other, 0o755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		pointerTarget := filepath.Join(other, ".taskflow")
		if err := os.WriteFile(pointerTarget, []byte("database: somewhere/tasks.db\n"), 0o644); err != nil {
			t.Fatalf("setup: %v", err)
		}
		before := readFileBytes(t, pointerTarget)

		r := runCLI(t, dir, "init", "--target", pointerTarget)
		if r.exitCode != 2 {
			t.Errorf("expected exit 2 for a pointer-file target, got %d; output:\n%s", r.exitCode, r.combined)
		}
		assertContains(t, r.combined, "invalid", "clear error")
		if anchorExistsAt(t, dir) {
			t.Error("nothing must be written: no pointer at the cwd")
		}
		assertBytesEqual(t, "the pointer-file target", before, readFileBytes(t, pointerTarget))
	})

	t.Run("existing .taskflow directory target", func(t *testing.T) {
		dir := tempWorkdir(t)
		target := filepath.Join(dir, "shared", ".taskflow")
		if err := os.MkdirAll(target, 0o755); err != nil {
			t.Fatalf("setup: %v", err)
		}

		r := runCLI(t, dir, "init", "--target", target)
		assertExitZero(t, r, "init --target with an existing .taskflow directory")

		if fi, err := os.Stat(filepath.Join(dir, ".taskflow")); err != nil || fi.IsDir() {
			t.Errorf("expected a pointer file at the cwd, err=%v", err)
		}
		dbPath := filepath.Join(target, "tasks.db")
		if _, err := os.Stat(dbPath); err != nil {
			t.Errorf("expected the DB created inside the target: %v", err)
		}
	})

	t.Run("existing tasks.db file target", func(t *testing.T) {
		dir := tempWorkdir(t)
		home := filepath.Join(dir, "home")
		if err := os.MkdirAll(home, 0o755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		initProject(t, home)
		addTask(t, home, addTaskJSON("target-data"))

		target := filepath.Join(home, ".taskflow", "tasks.db")
		before := readFileBytes(t, target)

		r := runCLI(t, dir, "init", "--target", target)
		assertExitZero(t, r, "init --target with an existing tasks.db file")

		assertBytesEqual(t, "the existing tasks.db target", before, readFileBytes(t, target))
		assertContains(t, r.stdout, "pointer: "+filepath.Join(dir, ".taskflow"), "pointer reported")

		lr := listAll(t, dir)
		assertContains(t, lr.stdout, "target-data", "the existing data is reachable through the pointer")
	})

	t.Run("missing .taskflow-named target", func(t *testing.T) {
		dir := tempWorkdir(t)
		target := filepath.Join(dir, "shared", ".taskflow")

		r := runCLI(t, dir, "init", "--target", target)
		assertExitZero(t, r, "init --target with a missing .taskflow target")

		pointerPath := filepath.Join(dir, ".taskflow")
		want := "database: " + target + "\n"
		if got := string(readFileBytes(t, pointerPath)); got != want {
			t.Errorf("pointer file must contain exactly %q, got %q", want, got)
		}
		dbPath := filepath.Join(target, "tasks.db")
		if _, err := os.Stat(dbPath); err != nil {
			t.Errorf("expected the DB created at the resolved target %s: %v", dbPath, err)
		}
	})

	t.Run("missing tasks.db-named target", func(t *testing.T) {
		dir := tempWorkdir(t)
		target := filepath.Join(dir, "shared2", "tasks.db")

		r := runCLI(t, dir, "init", "--target", target)
		assertExitZero(t, r, "init --target with a missing tasks.db target")

		pointerPath := filepath.Join(dir, ".taskflow")
		want := "database: " + target + "\n"
		if got := string(readFileBytes(t, pointerPath)); got != want {
			t.Errorf("pointer file must contain exactly %q, got %q", want, got)
		}
		if _, err := os.Stat(target); err != nil {
			t.Errorf("expected the DB created at the target itself %s: %v", target, err)
		}
	})

	t.Run("relative target form", func(t *testing.T) {
		root := tempWorkdir(t)
		shared := filepath.Join(root, "shared")
		wt1 := filepath.Join(root, "wt1")
		wt2 := filepath.Join(root, "wt2")
		for _, d := range []string{wt1, wt2} {
			if err := os.MkdirAll(d, 0o755); err != nil {
				t.Fatalf("setup: %v", err)
			}
		}

		r := runCLI(t, wt1, "init", "--target", "../shared/.taskflow")
		assertExitZero(t, r, "init --target with a relative path from wt1")

		pointerPath := filepath.Join(wt1, ".taskflow")
		want := "database: ../shared/.taskflow\n"
		if got := string(readFileBytes(t, pointerPath)); got != want {
			t.Errorf("pointer file must keep the relative path as given (%q), got %q", want, got)
		}
		dbPath := filepath.Join(shared, ".taskflow", "tasks.db")
		if _, err := os.Stat(dbPath); err != nil {
			t.Errorf("expected the DB at the resolved target %s: %v", dbPath, err)
		}

		r = runCLI(t, wt2, "init", "--target", "../shared/.taskflow")
		assertExitZero(t, r, "init --target with a relative path from wt2")

		// Sibling sharing, git-free: both dirs read and write the shared DB.
		r = runCLI(t, wt1, "add", addTaskJSON("rel-wt1"))
		assertExitZero(t, r, "add from wt1")
		r = runCLI(t, wt2, "add", addTaskJSON("rel-wt2"))
		assertExitZero(t, r, "add from wt2")

		lr1 := listAll(t, wt1)
		assertContains(t, lr1.stdout, "rel-wt1", "wt1 sees its task")
		assertContains(t, lr1.stdout, "rel-wt2", "wt1 sees the shared task")
		lr2 := listAll(t, wt2)
		assertContains(t, lr2.stdout, "rel-wt1", "wt2 sees the shared task")
		assertContains(t, lr2.stdout, "rel-wt2", "wt2 sees its task")
	})

	t.Run("unwritable parent keeps pointer with warning", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("permission test requires a non-root user")
		}
		dir := tempWorkdir(t)
		locked := filepath.Join(dir, "locked")
		if err := os.MkdirAll(locked, 0o755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		if err := os.Chmod(locked, 0o555); err != nil {
			t.Fatalf("setup: %v", err)
		}
		t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })

		target := filepath.Join(locked, "tasks.db")
		r := runCLI(t, dir, "init", "--target", target)
		if r.exitCode != 0 {
			t.Fatalf("expected exit 0 with the dangling warning, got %d; output:\n%s", r.exitCode, r.combined)
		}
		assertContains(t, r.stderr, "dangle", "the warning mentions dangling")
		if !anchorExistsAt(t, dir) {
			t.Error("the pointer must be kept at the cwd")
		}

		// Commands fail with the dangling contract while the target dangles.
		ra := runCLI(t, dir, "add", addTaskJSON("unwr-1"))
		if ra.exitCode != 2 {
			t.Errorf("expected exit 2, got %d; output:\n%s", ra.exitCode, ra.combined)
		}
		assertContains(t, ra.stderr, "dangling (target missing)", "dangling contract")

		// Make the parent writable and repair from the workdir.
		if err := os.Chmod(locked, 0o755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		ri := runCLI(t, dir, "init")
		if ri.exitCode != 0 {
			t.Fatalf("init must repair from the workdir (exit %d): %s", ri.exitCode, ri.combined)
		}
		if _, err := os.Stat(target); err != nil {
			t.Errorf("expected the DB at %s after repair: %v", target, err)
		}
		ra = runCLI(t, dir, "add", addTaskJSON("unwr-1"))
		assertExitZero(t, ra, "add works after the repair")
	})

	t.Run("failed repair prints underlying error and keeps pointer", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("permission test requires a non-root user")
		}
		dir := tempWorkdir(t)
		locked := filepath.Join(dir, "locked")
		if err := os.MkdirAll(locked, 0o755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		if err := os.Chmod(locked, 0o555); err != nil {
			t.Fatalf("setup: %v", err)
		}
		t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })

		target := filepath.Join(locked, "tasks.db")
		pointerPath := writePointerFile(t, dir, target)
		before := readFileBytes(t, pointerPath)

		r := runCLI(t, dir, "init")
		if r.exitCode != 2 {
			t.Fatalf("expected exit 2, got %d; output:\n%s", r.exitCode, r.combined)
		}
		assertContains(t, r.stderr, "taskflow: ", "the underlying error is printed")
		assertBytesEqual(t, "the pointer file after a failed repair", before, readFileBytes(t, pointerPath))
		if _, err := os.Stat(target); !os.IsNotExist(err) {
			t.Errorf("no DB must be created at the target (stat err=%v)", err)
		}
	})
}

// TestInitWithTaskflowDirEnv verifies the env branch of init: the DB is
// created at $TASKFLOW_DIR/tasks.db, no anchor is written, and where reports
// the env override.
func TestInitWithTaskflowDirEnv(t *testing.T) {
	dir := tempWorkdir(t)
	envDir := filepath.Join(dir, "envdb")
	env := []string{"TASKFLOW_DIR=" + envDir}

	ri := runCLIWithEnv(t, dir, env, "init")
	assertExitZero(t, ri, "init with TASKFLOW_DIR")
	assertContains(t, ri.stdout, filepath.Join(envDir, "tasks.db"), "the DB path is printed")

	var anchors []string
	if err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err == nil && d.Name() == ".taskflow" {
			anchors = append(anchors, path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk failed: %v", err)
	}
	if len(anchors) > 0 {
		t.Errorf("no .taskflow anchor must be created with TASKFLOW_DIR set; found %v", anchors)
	}

	wr := runCLIWithEnv(t, dir, env, "where")
	assertExitZero(t, wr, "where with TASKFLOW_DIR")
	assertContains(t, wr.stdout, "anchor: env", "the env anchor is reported")

	ra := runCLIWithEnv(t, dir, env, "add", addTaskJSON("env-1"))
	assertExitZero(t, ra, "add with TASKFLOW_DIR")

	lr := runCLIWithEnv(t, dir, env, "list", `{"all":true}`)
	assertContains(t, lr.stdout, "env-1", "the task is listed")
}

// TestInitIdempotency verifies that re-running init does not wipe data.
func TestInitIdempotency(t *testing.T) {
	dir := tempWorkdir(t)
	initProject(t, dir)
	addTask(t, dir, addTaskJSON("idem-1"))

	ri := runCLI(t, dir, "init")
	assertExitZero(t, ri, "second init")

	lr := listAll(t, dir)
	assertContains(t, lr.stdout, "idem-1", "the task is still listed after re-init")
}

// TestSiblingWorktreeSharing verifies cross-checkout sharing through a pointer
// anchor, using git only to set up a linked worktree. The git-free property is
// asserted by running an add with a minimal environment (no PATH).
func TestSiblingWorktreeSharing(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available; worktree setup requires it")
	}
	dir := tempWorkdir(t)
	main := filepath.Join(dir, "main")
	if err := os.MkdirAll(main, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	initProject(t, main)
	addTask(t, main, addTaskJSON("wt-main"))

	// Set up a throwaway git repo and a linked worktree (setup only).
	repo := filepath.Join(dir, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	gitHome := filepath.Join(dir, "githome")
	if err := os.MkdirAll(gitHome, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		cmd.Env = append([]string{
			"HOME=" + gitHome,
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
		}, "PATH="+os.Getenv("PATH"))
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	git("init", ".")
	git("-c", "commit.gpgsign=false", "commit", "--allow-empty", "-m", "init")
	git("worktree", "add", filepath.Join(dir, "wt"))

	wt := filepath.Join(dir, "wt")
	ri := runCLI(t, wt, "init", "--target", filepath.Join(main, ".taskflow", "tasks.db"))
	assertExitZero(t, ri, "init --target from the worktree")

	lr := listAll(t, wt)
	assertContains(t, lr.stdout, "wt-main", "the main task is visible in the worktree")

	ra := runCLI(t, wt, "add", addTaskJSON("wt-shared"))
	assertExitZero(t, ra, "add from the worktree")

	mlr := listAll(t, main)
	assertContains(t, mlr.stdout, "wt-main", "the main task is still visible from main")
	assertContains(t, mlr.stdout, "wt-shared", "the worktree task is visible from main")

	// Git-free property: taskflow works with a minimal environment.
	mr := runCLIWithMinimalEnv(t, wt, "add", addTaskJSON("wt-minimal"))
	assertExitZero(t, mr, "add with a minimal environment (no PATH)")
	mlr = listAll(t, main)
	assertContains(t, mlr.stdout, "wt-minimal", "the minimal-env task reached the shared DB")
}

// TestWhereOutput covers the section 7 report for all anchor kinds and both
// exit codes.
func TestWhereOutput(t *testing.T) {
	t.Run("dir anchor", func(t *testing.T) {
		dir := tempWorkdir(t)
		initProject(t, dir)

		r := runCLI(t, dir, "where")
		assertExitZero(t, r, "where with a dir anchor")
		assertContains(t, r.stdout, "db: "+filepath.Join(dir, ".taskflow", "tasks.db"), "absolute DB path ending in .taskflow/tasks.db")
		assertContains(t, r.stdout, "anchor: dir", "anchor kind dir")
		assertContains(t, r.stdout, "anchor path: "+dir, "anchor path")
		assertContains(t, r.stdout, "  "+dir+" (anchor found)", "chain ends at the anchor dir")
	})

	t.Run("pointer anchor", func(t *testing.T) {
		dir := tempWorkdir(t)
		main := filepath.Join(dir, "main")
		if err := os.MkdirAll(main, 0o755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		initProject(t, main)
		target := filepath.Join(main, ".taskflow", "tasks.db")
		writePointerFile(t, dir, target)

		r := runCLI(t, dir, "where")
		assertExitZero(t, r, "where with a pointer anchor")
		assertContains(t, r.stdout, "anchor: pointer", "anchor kind pointer")
		assertContains(t, r.stdout, "pointer target: "+target, "resolved pointer target")
		assertContains(t, r.stdout, "pointer status: ok", "pointer status ok")
	})

	t.Run("env override", func(t *testing.T) {
		dir := tempWorkdir(t)
		env := []string{"TASKFLOW_DIR=" + filepath.Join(dir, "envdb")}

		r := runCLIWithEnv(t, dir, env, "where")
		assertExitZero(t, r, "where with TASKFLOW_DIR")
		assertContains(t, r.stdout, "anchor: env", "anchor kind env")
		assertContains(t, r.stdout, "(TASKFLOW_DIR set; walk skipped)", "chain note")
	})

	t.Run("no anchor", func(t *testing.T) {
		dir := tempWorkdir(t)

		r := runCLI(t, dir, "where")
		if r.exitCode != 2 {
			t.Errorf("expected exit 2, got %d; output:\n%s", r.exitCode, r.combined)
		}
		assertContains(t, r.stdout, "anchor: none", "stdout reports no anchor")
		assertContains(t, r.stdout, "search chain:", "stdout has the chain")
		assertContains(t, r.stderr, "taskflow: no .taskflow anchor found", "stderr has the section 6 message")
	})

	t.Run("dangling pointer", func(t *testing.T) {
		dir := tempWorkdir(t)
		writePointerFile(t, dir, filepath.Join(dir, "missing", "tasks.db"))

		r := runCLI(t, dir, "where")
		if r.exitCode != 2 {
			t.Errorf("expected exit 2, got %d; output:\n%s", r.exitCode, r.combined)
		}
		assertContains(t, r.stdout, "pointer status: dangling", "stdout reports the dangling status")
		assertContains(t, r.stderr, "dangling (target missing)", "stderr has the dangling contract")
	})

	t.Run("dir anchor with missing tasks.db", func(t *testing.T) {
		dir := tempWorkdir(t)
		initProject(t, dir)
		dbPath := filepath.Join(dir, ".taskflow", "tasks.db")
		if err := os.Remove(dbPath); err != nil {
			t.Fatalf("setup: failed to remove tasks.db: %v", err)
		}

		r := runCLI(t, dir, "where")
		if r.exitCode != 2 {
			t.Errorf("expected exit 2 with the DB file missing, got %d; output:\n%s", r.exitCode, r.combined)
		}
		assertContains(t, r.stdout, "anchor: dir", "the anchor still resolves and is reported")
		assertContains(t, r.stdout, "db status: missing", "stdout reports the missing DB status")
		assertContains(t, r.stderr, "taskflow: "+dbPath+" is missing", "stderr has the missing-DB contract")
		assertContains(t, r.stderr, "remedy: run `taskflow init` to create or repair the database", "stderr has the init remedy")
	})
}

// TestWherePointerErrorStatuses verifies the stdout branches of where for
// hard pointer errors (review FIX 5b): the report names the pointer anchor,
// the malformed/wrong-type status, the holder dir, and the chain, while the
// stderr carries the section 6 pointer-error message; both exit 2.
func TestWherePointerErrorStatuses(t *testing.T) {
	t.Run("malformed pointer", func(t *testing.T) {
		dir := tempWorkdir(t)
		pointerPath := filepath.Join(dir, ".taskflow")
		if err := os.WriteFile(pointerPath, []byte("garbage\n"), 0o644); err != nil {
			t.Fatalf("setup: %v", err)
		}

		r := runCLI(t, dir, "where")
		if r.exitCode != 2 {
			t.Errorf("expected exit 2, got %d; output:\n%s", r.exitCode, r.combined)
		}
		assertContains(t, r.stdout, "anchor: pointer", "stdout reports the pointer anchor")
		assertContains(t, r.stdout, "anchor path: "+dir, "stdout reports the holder dir")
		assertContains(t, r.stdout, "pointer status: malformed", "stdout reports the malformed status")
		assertContains(t, r.stdout, "search chain:", "stdout has the chain")
		assertContains(t, r.stderr, "taskflow: "+pointerPath+" is malformed", "stderr names the pointer file and the reason")
		assertContains(t, r.stderr, "remedy: fix by hand, or run `taskflow init --force` below it to shadow it", "stderr has the pointer-error remedy")
	})

	t.Run("wrong-type pointer target", func(t *testing.T) {
		dir := tempWorkdir(t)
		plain := filepath.Join(dir, "plaindir")
		if err := os.MkdirAll(plain, 0o755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		pointerPath := writePointerFile(t, dir, plain)

		r := runCLI(t, dir, "where")
		if r.exitCode != 2 {
			t.Errorf("expected exit 2, got %d; output:\n%s", r.exitCode, r.combined)
		}
		assertContains(t, r.stdout, "anchor: pointer", "stdout reports the pointer anchor")
		assertContains(t, r.stdout, "anchor path: "+dir, "stdout reports the holder dir")
		assertContains(t, r.stdout, "pointer status: wrong-type", "stdout reports the wrong-type status")
		assertContains(t, r.stdout, "search chain:", "stdout has the chain")
		assertContains(t, r.stderr, "taskflow: "+pointerPath+" is wrong target type", "stderr names the pointer file and the reason")
		assertContains(t, r.stderr, "remedy: fix by hand, or run `taskflow init --force` below it to shadow it", "stderr has the pointer-error remedy")
	})
}

// TestConcurrencySmokeSharedPointerDB verifies that N parallel add invocations
// on one shared pointer DB all succeed and no lock errors surface
// (busy_timeout pragma and WAL). The DB is created by `init` before the
// parallel adds: runtime commands no longer auto-create a missing DB (review
// FIX 1), so the earlier lazy-creation-by-first-writer setup would now fail
// every add with the missing-DB contract.
func TestConcurrencySmokeSharedPointerDB(t *testing.T) {
	dir := tempWorkdir(t)
	anchorHome := filepath.Join(dir, "shared", ".taskflow")
	if err := os.MkdirAll(anchorHome, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	wt := filepath.Join(dir, "wt")
	if err := os.MkdirAll(wt, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	writePointerFile(t, wt, anchorHome)

	// Create the shared DB before the parallel adds (init is the only
	// command allowed to create; review FIX 1).
	ri := runCLI(t, wt, "init")
	if ri.exitCode != 0 {
		t.Fatalf("setup: init failed to create the shared DB (exit %d): %s", ri.exitCode, ri.combined)
	}

	const n = 6
	results := make([]cliResult, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i] = runCLIConcurrent(wt, "add", addTaskJSON(fmt.Sprintf("conc-%d", i)))
		}(i)
	}
	wg.Wait()

	for i, r := range results {
		if r.exitCode != 0 {
			t.Errorf("parallel add %d failed (exit %d):\n%s", i, r.exitCode, r.combined)
		}
		assertNotContains(t, strings.ToUpper(r.combined), "SQLITE_BUSY", "no SQLITE_BUSY in parallel add output")
		assertNotContains(t, strings.ToUpper(r.combined), "DATABASE IS LOCKED", "no lock error in parallel add output")
	}

	lr := listAll(t, wt)
	assertExitZero(t, lr, "list after the concurrent adds")
	for i := 0; i < n; i++ {
		assertContains(t, lr.stdout, fmt.Sprintf("conc-%d", i), "task from the parallel add is present")
	}
	assertNotContains(t, strings.ToUpper(lr.combined), "SQLITE_BUSY", "no SQLITE_BUSY in list output")
}
