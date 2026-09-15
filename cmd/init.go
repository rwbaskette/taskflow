package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rwbaskette/taskflow/internal/anchor"
	"github.com/rwbaskette/taskflow/internal/db"
	"github.com/spf13/cobra"
)

var (
	initTarget string
	initForce  bool
)

const initLongHelp = `Create or repair the taskflow anchor for the current project.

taskflow init is the only command that creates files; it never runs git.

What init creates:
  - a .taskflow directory at the current directory (the anchor), containing
    the task database tasks.db, created with the current schema.

Pointer files:
  - A .taskflow entry that is a regular file (not a directory) is a pointer
    anchor. It contains exactly one line:

        database: <path>

    where <path> is absolute, or relative to the directory that holds the
    pointer file. Runtime commands resolve the database through the pointer;
    the pointer is the only cross-checkout mechanism.

Cross-checkout sharing (--target):
  - Run ` + "`taskflow init --target <path>`" + ` from a second checkout (for example a
    linked worktree) to write a pointer file that shares one database with
    the first checkout. <path> must name an anchor home (a directory named
    .taskflow) or a database file (a file named tasks.db). A plain directory
    or plain file target is refused.

Recovery recipe for scattered databases:
  - Older taskflow versions auto-created .taskflow/tasks.db in whatever
    directory a command ran in, so databases can be scattered. To recover:
    run ` + "`taskflow init`" + ` in the project root, move each stray tasks.db into
    that root's .taskflow and delete the stray .taskflow directories, or
    leave a stray in place and point at it with a pointer file
    ` + "`database: <straydir>/.taskflow`" + `.

TASKFLOW_DIR still overrides everything: when set, init creates or opens
$TASKFLOW_DIR/tasks.db and no anchor is written.`

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create or repair the .taskflow anchor for the current project",
	Long:  initLongHelp,
	Example: `  taskflow init
  taskflow init --target /work/main/.taskflow/tasks.db
  taskflow init --force`,
	Args: cobra.NoArgs,
	Run:  runInit,
}

// runInit implements the section 4 algorithm: env override, anchor walk as
// guardrail, then repair-or-create. Exit codes follow the section 6 contract.
func runInit(cmd *cobra.Command, args []string) {
	// (a) TASKFLOW_DIR set and non-empty: create or open $TASKFLOW_DIR/tasks.db
	// and print the path; no anchor is written.
	if dir := os.Getenv("TASKFLOW_DIR"); dir != "" {
		path, err := db.DefaultDBPath()
		if err != nil {
			// Defensive: DefaultDBPath cannot fail while the env var is set.
			printAnchorError(err)
			return
		}
		database, err := db.NewDB(path)
		if err != nil {
			// Repair-failure rule: underlying error, exit 2, nothing touched.
			fmt.Fprintf(os.Stderr, "taskflow: %v\n", err)
			os.Exit(2)
		}
		dbPath := database.Path()
		_ = database.Close()
		fmt.Printf("db: %s\n", dbPath)
		os.Exit(0)
	}

	// (b) Anchor walk from the (symlink-resolved) cwd. This walk is the
	// guardrail against scattered anchors.
	a, aerr := anchor.Resolve("")
	start := startDir(a, aerr)

	if aerr != nil {
		switch {
		case anchor.IsNotFound(aerr):
			// (d) No anchor found.
			runInitFresh(start)
		case anchor.IsMalformedPointer(aerr) || anchor.IsWrongTargetType(aerr):
			// Malformed or wrong-type pointer. If it sits above the cwd and
			// --force is set, a shadow directory anchor at the cwd is the
			// remedy (design 4c); otherwise hard error, file untouched.
			pointerAtCwd := filepath.Dir(aerr.PointerPath) == start
			if !pointerAtCwd && initForce {
				createShadowAnchor(start, "shadow anchor created; the malformed pointer at "+
					aerr.PointerPath+" is untouched")
			}
			printAnchorError(aerr)
		default:
			printAnchorError(aerr)
		}
		return
	}

	// (c) Anchor found; step e edge: --target given but an anchor already
	// exists, so --target is ignored (step c behavior applies).
	if initTarget != "" {
		fmt.Printf("anchor already exists at %s; --target ignored\n", a.AnchorPath)
	}
	runInitExisting(a, start)
}

// startDir extracts the symlink-resolved start directory (the first entry of
// the walk chain) from whichever result Resolve produced.
func startDir(a *anchor.Anchor, aerr *anchor.AnchorError) string {
	if a != nil && len(a.Chain) > 0 {
		return a.Chain[0]
	}
	if aerr != nil && len(aerr.Chain) > 0 {
		return aerr.Chain[0]
	}
	cwd, _ := os.Getwd()
	return cwd
}

// runInitExisting handles step (c): an anchor was found. Init never creates a
// second anchor without --force, and --force never deletes a database.
func runInitExisting(a *anchor.Anchor, start string) {
	switch a.Kind {
	case anchor.KindDir:
		atCwd := a.AnchorPath == start

		// Repair is additive: recreate a missing tasks.db (also under
		// --force; --force never deletes a database).
		if err := ensureDB(a.DBPath); err != nil {
			fmt.Fprintf(os.Stderr, "taskflow: %v\n", err)
			os.Exit(2)
		}

		fmt.Printf("anchor: %s\n", filepath.Join(a.AnchorPath, ".taskflow"))
		fmt.Printf("db: %s\n", a.DBPath)

		if atCwd {
			// Report and exit (with or without --force: report-and-no-op on
			// a directory anchor at the cwd).
			os.Exit(0)
		}
		if !initForce {
			fmt.Printf("anchor already exists at %s; no new anchor created\n", a.AnchorPath)
			fmt.Println("pass --force to create a shadow anchor at the current directory")
			os.Exit(0)
		}
		// Above the cwd with --force: shadow directory anchor at the cwd.
		createShadowAnchor(start, "shadow anchor created; the outer anchor at "+
			a.AnchorPath+" remains and is shadowed for all subdirectories below "+
			filepath.Join(start, ".taskflow"))

	case anchor.KindPointer:
		pointerAtCwd := a.AnchorPath == start
		switch a.PointerStatus {
		case anchor.StatusOK:
			// Repair first (design section 4c: "recreate a missing
			// tasks.db (repair)"; section 3: only init creates or
			// repairs): the pointer resolves, but its DB file may be gone
			// (deleted by hand), which runtime commands now refuse to
			// auto-create. ensureDB creates the file only when missing and
			// is a no-op open otherwise (the schema is idempotent).
			if err := ensureDB(a.DBPath); err != nil {
				fmt.Fprintf(os.Stderr, "taskflow: %v\n", err)
				os.Exit(2)
			}
			if pointerAtCwd {
				if !initForce {
					fmt.Printf("pointer anchor at %s\n", filepath.Join(a.AnchorPath, ".taskflow"))
					fmt.Printf("db: %s\n", a.DBPath)
					os.Exit(0)
				}
				// --force at the cwd: replace the pointer with a directory
				// anchor; the pointer's target DB stays untouched. The
				// replacement restores the pointer bytes on failure
				// (repair-failure rule, design section 6).
				if err := replacePointerWithDirAnchor(a.AnchorPath, nil); err != nil {
					fmt.Fprintf(os.Stderr, "taskflow: %v\n", err)
					os.Exit(2)
				}
				fmt.Printf("anchor: %s\n", filepath.Join(a.AnchorPath, ".taskflow"))
				fmt.Printf("db: %s\n", filepath.Join(a.AnchorPath, ".taskflow", "tasks.db"))
				fmt.Println("pointer file replaced with a directory anchor; " +
					"the previous target database at " + a.DBPath + " is untouched")
				return
			}
			// Pointer above the cwd.
			if initForce {
				createShadowAnchor(start, "shadow anchor created; the pointer at "+
					filepath.Join(a.AnchorPath, ".taskflow")+" is untouched")
			}
			fmt.Printf("pointer anchor at %s\n", filepath.Join(a.AnchorPath, ".taskflow"))
			fmt.Printf("db: %s\n", a.DBPath)
			fmt.Printf("anchor already exists at %s; no new anchor created\n", a.AnchorPath)
			fmt.Println("pass --force to create a shadow anchor at the current directory")
			os.Exit(0)

		case anchor.StatusDangling:
			// --force with a dangling pointer ABOVE the cwd: create a shadow
			// directory anchor at the cwd instead of healing. At the cwd,
			// healing is local, so heal even under --force.
			if !pointerAtCwd && initForce {
				createShadowAnchor(start, "shadow anchor created; the dangling pointer at "+
					filepath.Join(a.AnchorPath, ".taskflow")+" is untouched")
			}
			// Repair: create the DB at the anchor's DBPath. The pointer and
			// the anchor stay untouched on failure.
			if err := ensureDB(a.DBPath); err != nil {
				fmt.Fprintf(os.Stderr, "taskflow: %v\n", err)
				os.Exit(2)
			}
			fmt.Printf("pointer anchor at %s\n", filepath.Join(a.AnchorPath, ".taskflow"))
			fmt.Printf("dangling pointer repaired; db: %s\n", a.DBPath)
			os.Exit(0)

		default:
			// Resolve never returns a malformed or wrong-type pointer as an
			// Anchor; defensive.
			printAnchorError(&anchor.AnchorError{
				Chain:       a.Chain,
				PointerPath: filepath.Join(a.AnchorPath, ".taskflow"),
				Reason:      a.PointerStatus,
			})
		}
	}
}

// runInitFresh handles step (d): no anchor found anywhere above the cwd.
func runInitFresh(start string) {
	if initTarget != "" {
		runInitTarget(start, initTarget)
		return
	}

	// Default: refuse protected directories without --force.
	if anchor.IsProtectedDir(start) && !initForce {
		fmt.Fprintln(os.Stderr, "refusing to create an anchor here; pass --force to override")
		os.Exit(2)
	}

	createShadowAnchor(start, "")
}

// runInitTarget implements the --target branch of step (d): validate the
// target kind, write the pointer file FIRST, then attempt DB creation at the
// resolved target (create-if-missing).
func runInitTarget(start, target string) {
	// The protected-dir guard also applies here: writing a pointer at $HOME
	// or the filesystem root is the same stray-anchor hazard.
	if anchor.IsProtectedDir(start) && !initForce {
		fmt.Fprintln(os.Stderr, "refusing to create an anchor here; pass --force to override")
		os.Exit(2)
	}

	// Validate the target kind BEFORE writing anything.
	base := filepath.Base(target)
	if base != ".taskflow" && base != "tasks.db" {
		fmt.Fprintf(os.Stderr, "taskflow: --target %q is invalid: the target must name an anchor home (a directory named .taskflow) or a database file (a file named tasks.db)\n", target)
		os.Exit(2)
	}

	resolved := target
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(start, resolved)
	}

	// An existing target with the wrong on-disk type is refused before
	// anything is written (design section 4d): a pointer file named .taskflow
	// cannot be a target (no nested pointers), and a directory named tasks.db
	// is not a database. The name check above cannot catch these because the
	// name is one of the two valid ones.
	if fi, err := os.Stat(resolved); err == nil {
		if (base == ".taskflow" && !fi.IsDir()) || (base == "tasks.db" && fi.IsDir()) {
			fmt.Fprintf(os.Stderr, "taskflow: --target %q is invalid: the target exists with the wrong type; a pointer file named .taskflow or a directory named tasks.db cannot be an anchor target\n", target)
			os.Exit(2)
		}
	}

	// Write the pointer file FIRST, with the user-given path as-is (absolute
	// stays absolute; relative stays relative and resolves against the
	// pointer's directory, which is the cwd).
	pointerPath := filepath.Join(start, ".taskflow")
	line := "database: " + target + "\n"
	if err := os.WriteFile(pointerPath, []byte(line), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "taskflow: %v\n", err)
		os.Exit(2)
	}

	// Attempt DB creation at the resolved target.
	dbPath := resolved
	if base == ".taskflow" {
		dbPath = filepath.Join(resolved, "tasks.db")
	}

	database, err := db.NewDB(dbPath)
	if err != nil {
		// Unwritable parent (or otherwise uncreatable target): KEEP the
		// pointer and warn that it dangles until the target is creatable.
		fmt.Fprintf(os.Stderr, "warning: pointer file %s written; it dangles until the target is creatable: %v\n", pointerPath, err)
		fmt.Fprintln(os.Stderr, "remedy: run `taskflow init` again once the target is writable")
		os.Exit(0)
	}
	p := database.Path()
	_ = database.Close()
	fmt.Printf("pointer: %s\n", pointerPath)
	fmt.Printf("db: %s\n", p)
	os.Exit(0)
}

// createShadowAnchor creates a .taskflow directory anchor plus its database
// at dir via db.NewDB (which runs migrate), prints the new anchor and DB
// paths, and exits 0. On failure it exits 2 and nothing but the failed DB
// creation is left behind, keeping the anchor site untouched (repair-failure
// rule, design section 6).
func createShadowAnchor(dir, note string) {
	database, err := db.NewDB(filepath.Join(dir, ".taskflow", "tasks.db"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "taskflow: %v\n", err)
		os.Exit(2)
	}
	dbPath := database.Path()
	_ = database.Close()
	fmt.Printf("anchor: %s\n", filepath.Join(dir, ".taskflow"))
	fmt.Printf("db: %s\n", dbPath)
	if note != "" {
		fmt.Println(note)
	}
	os.Exit(0)
}

// replacePointerWithDirAnchor implements the `init --force` pointer
// replacement at the cwd: the pointer file at dir/.taskflow is replaced with
// a directory anchor plus its database. create is the DB-creation seam; nil
// means ensureDB (db.NewDB running migrate, connection closed). It returns an
// error instead of exiting so the failure paths stay unit-testable.
//
// Failure handling follows the repair-failure rule (design section 6: "the
// pointer and the anchor stay untouched"): if the DB creation fails after
// db.NewDB already created dir/.taskflow as a directory (its MkdirAll runs
// before the DB is opened), that directory is removed before the pointer
// bytes are rewritten, so the restore cannot hit EISDIR and a fresh empty DB
// cannot silently shadow the lost pointer. A .taskflow directory that
// pre-existed is never removed. If the restore itself fails, the error says
// so and instructs a manual restore.
func replacePointerWithDirAnchor(dir string, create func(path string) error) error {
	if create == nil {
		create = ensureDB
	}
	pointerPath := filepath.Join(dir, ".taskflow")

	// Capture up front whether .taskflow pre-existed as a directory. At
	// this call site .taskflow is the pointer file (a regular file), so a
	// directory found at this path AFTER the replacement started was
	// created by this run (db.NewDB's MkdirAll) and is safe to remove on
	// failure; a directory that pre-existed is never removed.
	preExistedAsDir := false
	if fi, err := os.Stat(pointerPath); err == nil && fi.IsDir() {
		preExistedAsDir = true
	}

	// Keep the pointer bytes: if the DB creation fails after the removal,
	// the pointer file is restored byte-for-byte before exit 2.
	pointerBytes, err := os.ReadFile(pointerPath)
	if err != nil {
		return err
	}

	if err := os.Remove(pointerPath); err != nil {
		return err
	}

	if dbErr := create(filepath.Join(dir, ".taskflow", "tasks.db")); dbErr != nil {
		// Undo the removal. If .taskflow is now a directory that this run
		// created, remove it first so the pointer bytes can be rewritten;
		// a pre-existing .taskflow directory is never removed. Every
		// restore failure is surfaced in the returned error.
		if !preExistedAsDir {
			if fi, serr := os.Stat(pointerPath); serr == nil && fi.IsDir() {
				if rmErr := os.RemoveAll(pointerPath); rmErr != nil {
					return fmt.Errorf("%w; pointer restore failed: %v; restore the pointer file %s by hand", dbErr, rmErr, pointerPath)
				}
			}
		}
		if werr := os.WriteFile(pointerPath, pointerBytes, 0o644); werr != nil {
			return fmt.Errorf("%w; pointer restore failed: %v; restore the pointer file %s by hand", dbErr, werr, pointerPath)
		}
		return dbErr
	}
	return nil
}

// ensureDB creates or opens the database at path (running migrate) and
// closes the connection; init only needs the file to exist with the schema.
func ensureDB(path string) error {
	database, err := db.NewDB(path)
	if err != nil {
		return err
	}
	return database.Close()
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVar(&initTarget, "target", "", "create a pointer anchor to the database at <path> (anchor home .taskflow or database file tasks.db)")
	initCmd.Flags().BoolVar(&initForce, "force", false, "override the anchor guardrail: shadow an anchor above the cwd, or replace a pointer at the cwd")
}
