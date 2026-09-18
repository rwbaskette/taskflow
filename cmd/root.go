package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rwbaskette/taskflow/internal/anchor"
	"github.com/rwbaskette/taskflow/internal/db"
	cliErrors "github.com/rwbaskette/taskflow/internal/errors"
	"github.com/rwbaskette/taskflow/internal/service"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "taskflow",
	Short: "A task management CLI tool",
	Long: `Task is a CLI tool for managing tasks with support for
adding, updating, completing, blocking, listing, and resetting tasks.

For more information, visit the project documentation.

Environment Variables:
  TASKFLOW_DIR  Custom directory for the database (default: .taskflow)
                  When set, the database will be created at $TASKFLOW_DIR/tasks.db
                  Example: TASKFLOW_DIR=/tmp/work/taskflow taskflow list`,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// fatal reports err via cliErrors.HandleError and exits. HandleError exits 1
// on every non-nil error (both the JSON and the marshal-fallback branches end
// in os.Exit(1); it returns only when err is nil). fatal is only ever called
// with a non-nil error, but the trailing os.Exit(1) keeps the helper
// provably noreturn regardless.
func fatal(err error) {
	cliErrors.HandleError(err)
	os.Exit(1) // unreachable for non-nil err; guarantees fatal never returns
}

// fatal2 prints err as a plain-text one-liner to stderr and exits 2. It is
// the init-style fatal path: human-readable text, exit code 2, no JSON.
func fatal2(err error) {
	fmt.Fprintf(os.Stderr, "taskflow: %v\n", err)
	os.Exit(2)
}

// openDB resolves the default database path and opens it, printing the
// section 6 error contract (exit 2) on anchor failures and the JSON error
// contract (exit 1) on open failures. It never returns on error: callers can
// rely on a non-nil *db.DB. openDB does not close the DB; callers defer
// database.Close().
func openDB() *db.DB {
	path, err := db.DefaultDBPath()
	if err != nil {
		printAnchorError(err) // always exits 2 (os.Exit(2) is its last statement)
		os.Exit(2)            // unreachable; keeps openDB provably noreturn here
	}
	database, err := db.NewDB(path)
	if err != nil {
		fatal(err)
	}
	return database
}

// statusAliases maps the accepted status aliases (lowercase) to their
// canonical storable values (db.validateTask only accepts todo, in_progress,
// done, blocked).
var statusAliases = map[string]string{
	"pending":     "todo",
	"in-progress": "in_progress",
	"inprogress":  "in_progress",
	"completed":   "done",
}

// canonicalStatus returns the canonical storable status for the given value:
// the alias's canonical value for an alias, the lowercased value unchanged
// for an already-canonical status, and "" for anything else ("all" included,
// which is a list-only concept and must never be stored).
func canonicalStatus(status string) string {
	s := strings.ToLower(strings.TrimSpace(status))
	switch s {
	case "todo", "in_progress", "done", "blocked":
		return s
	}
	return statusAliases[s]
}

// normalizeStatus validates the status via cliErrors.ValidateStatus (exiting
// via fatal on error), maps it to its canonical storable value, and returns
// it. An accepted value with no canonical mapping (only "all") is fatal.
func normalizeStatus(status string) string {
	if err := cliErrors.ValidateStatus(status); err != nil {
		fatal(err)
	}
	canonical := canonicalStatus(status)
	if canonical == "" {
		fatal(cliErrors.ValidationError("status",
			fmt.Sprintf("'%s' is not a valid task status", status),
			"Valid statuses: todo, in_progress, done, blocked"))
	}
	return canonical
}

// validateOptionalTaskFields validates the already-extracted optional task
// field values (title, milestone, actor), exiting via fatal on the first
// invalid value. Empty values are treated as absent and are valid; the
// callers extract the fields with service.GetStringFieldTrim, which trims
// each value and reports empty ones as absent. Status is not validated here:
// callers normalize it with normalizeStatus. Shared by update and complete.
func validateOptionalTaskFields(title, milestone, actor string) {
	if title != "" {
		if err := cliErrors.ValidateTitle(title); err != nil {
			fatal(err)
		}
	}
	if milestone != "" {
		if err := cliErrors.ValidateMilestone(milestone); err != nil {
			fatal(err)
		}
	}
	if actor != "" {
		if err := cliErrors.ValidateActor(actor); err != nil {
			fatal(err)
		}
	}
}

// jsonDoc parses the shared JSON document argument: the -j/--json flag wins,
// then args[0]; when both are empty, emptyDefault is used ("{}" for the
// tolerant commands list and reset) or MissingArgumentError is returned (the
// strict commands). ParseJSONFromArg handles the "-" stdin convention.
func jsonDoc(flagVal string, args []string, emptyDefault string) (map[string]interface{}, error) {
	arg := flagVal
	if arg == "" && len(args) > 0 {
		arg = args[0]
	}
	if arg == "" {
		if emptyDefault == "" {
			return nil, cliErrors.MissingArgumentError("json", "provide JSON document via argument or stdin")
		}
		arg = emptyDefault
	}
	return service.ParseJSONFromArg(arg)
}

// printTaskResult prints the shared success report for add and update.
func printTaskResult(verb string, task *db.Task) {
	fmt.Println("Task " + verb + " successfully:")
	fmt.Printf("  ID: %s\n", task.ID)
	fmt.Printf("  Title: %s\n", task.Title)
	fmt.Printf("  Description: %s\n", task.Description)
	fmt.Printf("  Milestone: %s\n", task.Milestone)
	if task.Actor != "" {
		fmt.Printf("  Actor: %s\n", task.Actor)
	}
	fmt.Printf("  Status: %s\n", task.Status)
}

// printStatusResult prints the short success report (ID, Title, Status only)
// shared by block, complete, and unblock.
func printStatusResult(verb string, task *db.Task) {
	fmt.Println("Task " + verb + " successfully:")
	fmt.Printf("  ID: %s\n", task.ID)
	fmt.Printf("  Title: %s\n", task.Title)
	fmt.Printf("  Status: %s\n", task.Status)
}

// pointerReasonText maps an anchor reason to the section 6 pointer-error
// reason text.
func pointerReasonText(reason string) string {
	switch reason {
	case anchor.ReasonMalformed:
		return "malformed"
	case anchor.ReasonDangling:
		return "dangling (target missing)"
	case anchor.ReasonWrongType:
		return "wrong target type"
	default:
		return reason
	}
}

// renderAnchorError renders the section 6 error contract in plain text (no
// color, no JSON; agents parse plain text). It returns the full multi-line
// message including a trailing newline, so `where` can print stdout before
// the same rendering goes to stderr with exit 2.
func renderAnchorError(err error) string {
	var aerr *anchor.AnchorError
	if !errors.As(err, &aerr) {
		// Defensive: not an anchor error. Plain text, exit 2 path.
		return fmt.Sprintf("taskflow: %s\n", err.Error())
	}

	switch {
	case errors.Is(err, anchor.ErrNotFound):
		// The chain from AnchorError ends at the filesystem root; the
		// printer annotates the stop reason itself (AnchorError has no
		// StopReason field).
		var b strings.Builder
		fmt.Fprint(&b, "taskflow: no .taskflow anchor found\nsearched:\n")
		printChain(&b, aerr.Chain, "(filesystem root reached, no .taskflow)")
		if len(aerr.Chain) == 1 && filepath.IsAbs(aerr.Chain[0]) {
			// A not-found chain of length one with an absolute start dir means
			// the walk started at the filesystem root: the launcher ran with
			// the root as its working directory and no anchor can exist above
			// it. (A length-one chain can also come from a non-absolute start
			// fallback such as "." when os.Getwd fails; the IsAbs check keeps
			// the note off that case.) Print the actual start dir, portable
			// across volume roots.
			fmt.Fprintf(&b, "note: the search started at the filesystem root; the process that started taskflow ran with cwd=%s and no anchor can exist above it; fix the launcher's working directory or set TASKFLOW_DIR\n", aerr.Chain[0])
		}
		fmt.Fprint(&b, "remedy: run `taskflow init` in the project root (--target <path> shares one DB across checkouts)\n")
		fmt.Fprint(&b, "or set TASKFLOW_DIR to override\n")
		return b.String()

	case aerr.Reason == anchor.ReasonMissingDB:
		// The anchor resolved but the DB file is gone (deleted by hand, or a
		// pointer to an anchor home whose tasks.db is absent). PointerPath
		// carries the missing DB path here, set by db.DefaultDBPath.
		return fmt.Sprintf("taskflow: %s is missing\nremedy: run `taskflow init` to create or repair the database\n", aerr.PointerPath)

	case errors.Is(err, anchor.ErrDanglingPointer):
		b := fmt.Sprintf("taskflow: %s is %s\n", aerr.PointerPath, pointerReasonText(aerr.Reason))
		b += "remedy: run `taskflow init` to create or repair the database\n"
		return b

	default:
		// malformed or wrong-type
		b := fmt.Sprintf("taskflow: %s is %s\n", aerr.PointerPath, pointerReasonText(aerr.Reason))
		b += "remedy: fix by hand, or run `taskflow init --force` below it to shadow it\n"
		return b
	}
}

// printAnchorError prints the section 6 error contract to stderr in plain
// text and exits 2 directly. Anchor errors must never go through
// cliErrors.HandleError (JSON, exit 1); they render as plain text here.
func printAnchorError(err error) {
	fmt.Fprint(os.Stderr, renderAnchorError(err))
	os.Exit(2)
}

func init() {
	rootCmd.SetVersionTemplate("Task CLI version: {{.Version}}\n")
}
