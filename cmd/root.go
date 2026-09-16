package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rwbaskette/taskflow/internal/anchor"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	version = "0.1.0"
	commit  = ""
	date    = ""
)

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
	case anchor.IsNotFound(err):
		var b []byte
		b = append(b, "taskflow: no .taskflow anchor found\nsearched:\n"...)
		for i, dir := range aerr.Chain {
			if i == len(aerr.Chain)-1 {
				// The chain from AnchorError ends at the filesystem root;
				// the printer annotates the stop reason itself (AnchorError
				// has no StopReason field).
				b = append(b, fmt.Sprintf("  %s (filesystem root reached, no .taskflow)\n", dir)...)
			} else {
				b = append(b, fmt.Sprintf("  %s\n", dir)...)
			}
		}
		if len(aerr.Chain) == 1 && filepath.IsAbs(aerr.Chain[0]) {
			// A not-found chain of length one with an absolute start dir means
			// the walk started at the filesystem root: the launcher ran with
			// the root as its working directory and no anchor can exist above
			// it. (A length-one chain can also come from a non-absolute start
			// fallback such as "." when os.Getwd fails; the IsAbs check keeps
			// the note off that case.) Print the actual start dir, portable
			// across volume roots.
			b = append(b, fmt.Sprintf("note: the search started at the filesystem root; the process that started taskflow ran with cwd=%s and no anchor can exist above it; fix the launcher's working directory or set TASKFLOW_DIR\n", aerr.Chain[0])...)
		}
		b = append(b, "remedy: run `taskflow init` in the project root (--target <path> shares one DB across checkouts)\n"...)
		b = append(b, "or set TASKFLOW_DIR to override\n"...)
		return string(b)

	case aerr.Reason == anchor.ReasonMissingDB:
		// The anchor resolved but the DB file is gone (deleted by hand, or a
		// pointer to an anchor home whose tasks.db is absent). PointerPath
		// carries the missing DB path here, set by db.DefaultDBPath.
		return fmt.Sprintf("taskflow: %s is missing\nremedy: run `taskflow init` to create or repair the database\n", aerr.PointerPath)

	case anchor.IsDanglingPointer(err):
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
// cliErrors.HandleError (JSON, exit 1) or PrintError (colored text).
func printAnchorError(err error) {
	fmt.Fprint(os.Stderr, renderAnchorError(err))
	os.Exit(2)
}

func init() {
	cobra.OnInitialize()

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	// Set version - Cobra handles --version flag automatically
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("Task CLI version: {{.Version}}\n")
}
