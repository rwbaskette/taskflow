package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rwbaskette/taskflow/internal/anchor"
	"github.com/spf13/cobra"
)

var whereCmd = &cobra.Command{
	Use:   "where",
	Short: "Show the resolved database, anchor, and search chain",
	Long:  "Show which database taskflow would use, the anchor that resolved it, and the directories searched.\n\nOne item per line: db, anchor, anchor path, pointer target, pointer status, search chain. Exit 0 when an anchor resolves to an openable database; exit 2 with the section 6 error contract otherwise. No git context is printed.",
	Args:  cobra.NoArgs,
	Run:   runWhere,
}

// anchorKindLabel maps an anchor Kind to its section 7 report label.
func anchorKindLabel(k anchor.Kind) string {
	switch k {
	case anchor.KindDir:
		return "dir"
	case anchor.KindPointer:
		return "pointer"
	case anchor.KindEnv:
		return "env"
	default:
		return "unknown"
	}
}

// runWhere implements the section 7 debug command. It prints the resolution
// report to stdout first, then (on failure) the section 6 contract to stderr
// via the same rendering printAnchorError uses, exiting 2.
func runWhere(cmd *cobra.Command, args []string) {
	// TASKFLOW_DIR override: the walk is skipped entirely; env always wins
	// (design section 3 rule 1). where is diagnostic, so this branch keeps
	// exit 0 with the report even when the DB file does not exist yet: the
	// env contract is "TASKFLOW_DIR names where the DB lives", and the
	// auto-create contract there is unchanged.
	if dir := os.Getenv("TASKFLOW_DIR"); dir != "" {
		abs, err := filepath.Abs(dir)
		if err != nil {
			abs = dir
		}
		fmt.Printf("db: %s\n", filepath.Join(abs, "tasks.db"))
		fmt.Printf("anchor: %s\n", anchorKindLabel(anchor.KindEnv))
		fmt.Printf("anchor path: %s\n", abs)
		fmt.Println("search chain:")
		fmt.Println("  (TASKFLOW_DIR set; walk skipped)")
		os.Exit(0)
	}

	a, aerr := anchor.Resolve("")

	if aerr == nil {
		// Success: directory anchor, or a valid/dangling pointer.
		fmt.Printf("db: %s\n", a.DBPath)
		fmt.Printf("anchor: %s\n", anchorKindLabel(a.Kind))
		fmt.Printf("anchor path: %s\n", a.AnchorPath)
		if a.Kind == anchor.KindPointer {
			fmt.Printf("pointer target: %s\n", a.PointerTarget)
			fmt.Printf("pointer status: %s\n", a.PointerStatus)
		}
		fmt.Println("search chain:")
		printChain(a.Chain, "(anchor found)")

		// A dangling pointer resolves but no DB exists: exit 2 with the
		// pointer-error contract so where diagnoses stray latches.
		if a.PointerStatus == anchor.StatusDangling {
			aerr := &anchor.AnchorError{
				Chain:       a.Chain,
				PointerPath: filepath.Join(a.AnchorPath, ".taskflow"),
				Reason:      anchor.ReasonDangling,
			}
			fmt.Fprint(os.Stderr, renderAnchorError(aerr))
			os.Exit(2)
		}

		// A resolved anchor whose DB file is gone (deleted by hand) is not an
		// openable DB: exit 2 with the missing-DB contract (design section 7).
		// No stat-auto-create happens here: where never opens the DB.
		if _, serr := os.Stat(a.DBPath); serr != nil && os.IsNotExist(serr) {
			fmt.Println("db status: missing")
			aerr := &anchor.AnchorError{
				Chain:       a.Chain,
				PointerPath: a.DBPath,
				Reason:      anchor.ReasonMissingDB,
			}
			fmt.Fprint(os.Stderr, renderAnchorError(aerr))
			os.Exit(2)
		}
		os.Exit(0)
	}

	// Failure: print the chain and status info to stdout, then the section 6
	// contract to stderr, exit 2.
	switch {
	case anchor.IsNotFound(aerr):
		fmt.Println("anchor: none")
		fmt.Println("search chain:")
		printChain(aerr.Chain, "(filesystem root reached, no .taskflow)")
	case anchor.IsMalformedPointer(aerr) || anchor.IsWrongTargetType(aerr):
		fmt.Println("anchor: pointer")
		fmt.Printf("anchor path: %s\n", filepath.Dir(aerr.PointerPath))
		fmt.Printf("pointer status: %s\n", aerr.Reason)
		fmt.Println("search chain:")
		printChain(aerr.Chain, "(pointer error)")
	default:
		fmt.Println("anchor: none")
		fmt.Println("search chain:")
		printChain(aerr.Chain, "(filesystem root reached, no .taskflow)")
	}
	fmt.Fprint(os.Stderr, renderAnchorError(aerr))
	os.Exit(2)
}

// printChain renders the walked directories, one per line, annotating the
// final entry with the stop reason.
func printChain(chain []string, lastNote string) {
	for i, dir := range chain {
		if i == len(chain)-1 {
			fmt.Printf("  %s %s\n", dir, lastNote)
		} else {
			fmt.Printf("  %s\n", dir)
		}
	}
}

func init() {
	rootCmd.AddCommand(whereCmd)
}
