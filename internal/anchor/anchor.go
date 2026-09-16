// Package anchor resolves the .taskflow anchor for a working directory.
//
// Resolution walks from a start directory upward to the filesystem root,
// looking for an anchor: a directory named .taskflow (a directory anchor) or
// a regular file named .taskflow containing a single "database: <path>" line
// (a pointer anchor). See taskflow-init-anchor.md sections 2, 3, 5, and 9.
//
// The package has zero git involvement: no git binary, no exec.Command, and
// no .git marker tests anywhere.
package anchor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// anchorName is the anchor entry name inside every probed directory.
const anchorName = ".taskflow"

// Kind classifies the anchor that resolved (or the override a caller such as
// `taskflow where` applies on its own).
type Kind int

const (
	// KindDir means .taskflow is a directory; the DB lives inside it.
	KindDir Kind = iota
	// KindPointer means .taskflow is a regular pointer file.
	KindPointer
	// KindEnv means TASKFLOW_DIR override. Produced by callers like `where`,
	// never by Resolve.
	KindEnv
)

// PointerStatus values carried on Anchor.
const (
	StatusOK       = "ok"
	StatusDangling = "dangling"
)

// AnchorError reason values. Each maps to exactly one sentinel error.
const (
	ReasonNotFound  = "not-found"
	ReasonMalformed = "malformed"
	ReasonWrongType = "wrong-type"
	ReasonDangling  = "dangling"

	// ReasonMissingDB: the anchor resolved, but the DB file at DBPath is
	// gone (deleted by hand, or a pointer to an existing anchor home whose
	// tasks.db is absent). Produced by db.DefaultDBPath's walk branch so
	// runtime commands never silently auto-create a deleted DB (design
	// section 3: no DB is created; only taskflow init creates or repairs).
	ReasonMissingDB = "missing-db"
)

// Sentinel errors wrapped by AnchorError. Resolve itself never returns
// ErrDanglingPointer (a dangling pointer resolves to an Anchor with
// PointerStatus "dangling" and a nil error, so callers can diagnose it
// without a hard failure); the sentinel has two producers: init's
// repair-failure path, and db.DefaultDBPath for the missing-DB walk case
// (ReasonMissingDB), which is why ReasonMissingDB unwraps to it.
var (
	ErrNotFound         = errors.New("no .taskflow anchor found")
	ErrMalformedPointer = errors.New("malformed .taskflow pointer file")
	ErrWrongTargetType  = errors.New("pointer target has the wrong type")
	ErrDanglingPointer  = errors.New("pointer target is missing")
)

// Anchor is the result of a successful resolution. For a dangling pointer,
// Resolve returns an Anchor (not an error) with PointerStatus "dangling" and
// DBPath set to where the DB would live; the walk stops there.
type Anchor struct {
	Kind Kind

	// DBPath is the absolute path where tasks.db lives (or would live).
	DBPath string

	// AnchorPath is the absolute path of the directory that holds the anchor.
	AnchorPath string

	// PointerTarget is the resolved absolute pointer target (the .taskflow
	// directory or the tasks.db file as written/resolved); pointer anchors
	// only.
	PointerTarget string

	// PointerStatus is "ok", "dangling", "malformed", or "wrong-type".
	PointerStatus string

	// Chain lists the directories walked, starting dir first. On success it
	// ends at the anchor dir; on failure at the last probed dir. Required by
	// `taskflow where` (design section 7) for all kinds, including successful
	// resolution. This field is a deviation-in-addition to the section 9
	// struct; section 7 requires it and the chain also feeds AnchorError.
	Chain []string
}

// AnchorError is the failure result of Resolve. Reason is one of the Reason
// constants and selects the wrapped sentinel; Error consumers unwrap with
// errors.Is or errors.As.
type AnchorError struct {
	// Chain lists the directories walked, start dir first; the last entry is
	// the filesystem root when not-found.
	Chain []string

	// PointerPath is set for pointer errors: the path of the offending
	// pointer file (for missing-db it carries the DB file path instead).
	PointerPath string

	// Reason is "not-found", "malformed", "wrong-type", "dangling", or
	// "missing-db".
	Reason string
}

// Error renders a sane single-line message. The section 6 printer in cmd/
// renders the full multi-line contract itself; this method is for generic
// error paths.
func (e *AnchorError) Error() string {
	switch e.Reason {
	case ReasonNotFound:
		start := ""
		if len(e.Chain) > 0 {
			start = e.Chain[0]
		}
		return fmt.Sprintf("%s (searched from %s up to the filesystem root)", ErrNotFound, start)
	case ReasonMalformed:
		return fmt.Sprintf("%s: %s", ErrMalformedPointer, e.PointerPath)
	case ReasonWrongType:
		return fmt.Sprintf("%s: %s", ErrWrongTargetType, e.PointerPath)
	case ReasonDangling:
		return fmt.Sprintf("%s: %s", ErrDanglingPointer, e.PointerPath)
	case ReasonMissingDB:
		// PointerPath carries the missing DB path in this case (set by
		// db.DefaultDBPath's walk branch), not a pointer file path.
		return fmt.Sprintf("database file is missing: %s", e.PointerPath)
	default:
		return fmt.Sprintf("anchor error (%s)", e.Reason)
	}
}

// Unwrap maps the reason to its sentinel.
func (e *AnchorError) Unwrap() error {
	switch e.Reason {
	case ReasonNotFound:
		return ErrNotFound
	case ReasonMalformed:
		return ErrMalformedPointer
	case ReasonWrongType:
		return ErrWrongTargetType
	case ReasonDangling:
		return ErrDanglingPointer
	case ReasonMissingDB:
		return ErrDanglingPointer
	default:
		return nil
	}
}

// IsNotFound reports whether err is (or wraps) an anchor not-found error.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsMalformedPointer reports whether err is (or wraps) a malformed-pointer
// error.
func IsMalformedPointer(err error) bool {
	return errors.Is(err, ErrMalformedPointer)
}

// IsWrongTargetType reports whether err is (or wraps) a wrong-target-type
// error.
func IsWrongTargetType(err error) bool {
	return errors.Is(err, ErrWrongTargetType)
}

// IsDanglingPointer reports whether err is (or wraps) a dangling-pointer
// error.
func IsDanglingPointer(err error) bool {
	return errors.Is(err, ErrDanglingPointer)
}

// Resolve walks upward from dir looking for a .taskflow anchor. An empty dir
// means the cwd, with filepath.EvalSymlinks applied to the start dir first
// (macOS maps /tmp to /private/tmp); if EvalSymlinks errors, the raw start
// string is used. The walk never aborts: stat and EvalSymlinks errors mean
// "no anchor here, continue upward".
//
// Returns (anchor, nil) for a directory anchor, a valid pointer anchor, and
// a dangling pointer (Anchor.PointerStatus "dangling", walk stopped).
// Returns (nil, *AnchorError) for not-found, malformed, and wrong-type; the
// error carries the walked chain and, for pointer errors, the pointer path.
func Resolve(dir string) (*Anchor, *AnchorError) {
	return walkUp(evalStart(dir))
}

// evalStart determines the start directory: cwd when dir is empty, then
// EvalSymlinks with a fallback to the raw string on error. The result is
// always absolute so the root-termination rule works.
func evalStart(dir string) string {
	start := dir
	if start == "" {
		if cwd, err := os.Getwd(); err == nil {
			start = cwd
		} else {
			start = "."
		}
	}
	if resolved, err := filepath.EvalSymlinks(start); err == nil {
		start = resolved
	}
	if abs, err := filepath.Abs(start); err == nil {
		start = abs
	}
	return start
}

// walkUp probes each directory from start upward. Two stop conditions only:
// an anchor (section 5 rule 1) and the filesystem root (section 5 rule 2).
// See Resolve for the return contract.
func walkUp(start string) (*Anchor, *AnchorError) {
	chain := make([]string, 0, 8)
	d := start
	for {
		chain = append(chain, d)
		anchorPath := filepath.Join(d, anchorName)
		fi, err := os.Stat(anchorPath)
		switch {
		case err == nil && fi.IsDir():
			// Directory anchor.
			return &Anchor{
				Kind:          KindDir,
				DBPath:        filepath.Join(anchorPath, "tasks.db"),
				AnchorPath:    d,
				PointerStatus: StatusOK,
				Chain:         chain,
			}, nil
		case err == nil && fi.Mode().IsRegular():
			// Pointer anchor. Malformed or wrong-type: hard error, stop.
			// Dangling: anchor with status dangling, stop.
			a, aerr := parsePointer(anchorPath)
			if aerr != nil {
				return nil, &AnchorError{
					Chain:       chain,
					PointerPath: anchorPath,
					Reason:      aerr.Reason,
				}
			}
			a.Chain = chain
			return a, nil
		}
		// Not found, or a stat error, or an entry that is neither a directory
		// nor a regular file (for example a device): no anchor here, continue
		// upward. The walk never aborts (section 5).
		next := filepath.Dir(d)
		if next == d {
			return nil, &AnchorError{Chain: chain, Reason: ReasonNotFound}
		}
		d = next
	}
}

// parsePointer parses a pointer file (a regular file named .taskflow) and
// returns the pointer anchor. The target must exist with the right type: a
// directory named .taskflow (anchor home) or a file named tasks.db (DB
// file). Missing targets yield an anchor with PointerStatus "dangling";
// syntax violations and unusable target types yield an AnchorError. A read
// error after a successful stat is treated as malformed (hard error).
func parsePointer(file string) (*Anchor, *AnchorError) {
	holder := filepath.Dir(file)
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, &AnchorError{PointerPath: file, Reason: ReasonMalformed}
	}
	target, aerr := pointerTarget(file, string(data), holder)
	if aerr != nil {
		return nil, aerr
	}
	a, aerr := classifyTarget(file, target)
	if aerr != nil {
		return nil, aerr
	}
	a.AnchorPath = holder
	return a, nil
}

// pointerTarget extracts the "database: <path>" line and resolves the path.
// Relative paths resolve against the directory that holds the pointer file.
func pointerTarget(file, content, holder string) (string, *AnchorError) {
	malformed := func() (string, *AnchorError) {
		return "", &AnchorError{PointerPath: file, Reason: ReasonMalformed}
	}
	var line string
	nonEmpty := 0
	for _, raw := range strings.Split(content, "\n") {
		// Strip CR/LF and surrounding whitespace per line; ignore lines that
		// become empty (one blank trailing line is OK).
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		nonEmpty++
		if nonEmpty > 1 {
			return malformed()
		}
		line = raw
	}
	if nonEmpty == 0 {
		return malformed()
	}
	if !strings.HasPrefix(line, "database:") {
		return malformed()
	}
	path := strings.TrimSpace(strings.TrimPrefix(line, "database:"))
	if path == "" {
		return malformed()
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(holder, path)
	}
	return filepath.Clean(path), nil
}

// classifyTarget applies the target name rule (design section 2) and the
// on-disk type checks, mapping each combination to ok, dangling, malformed,
// or wrong-type. A stat error other than not-exist is treated as "missing":
// the walk never aborts, and the dangling status is healable.
func classifyTarget(file, target string) (*Anchor, *AnchorError) {
	base := filepath.Base(target)
	fi, err := os.Stat(target)

	switch {
	case err == nil && fi.IsDir():
		if base == anchorName {
			return &Anchor{
				Kind:          KindPointer,
				DBPath:        filepath.Join(target, "tasks.db"),
				PointerTarget: target,
				PointerStatus: StatusOK,
			}, nil
		}
		// A directory that is not an anchor home, including a directory
		// named tasks.db.
		return nil, &AnchorError{PointerPath: file, Reason: ReasonWrongType}

	case err == nil:
		// Regular file, or another non-directory entry (device, socket, ...).
		switch base {
		case anchorName:
			// A pointer-file target is a hard error: no nested pointers.
			return nil, &AnchorError{PointerPath: file, Reason: ReasonMalformed}
		case "tasks.db":
			// An existing DB file is a valid target (design section 2: a
			// valid target names an anchor home or a DB file named tasks.db).
			return &Anchor{
				Kind:          KindPointer,
				DBPath:        target,
				PointerTarget: target,
				PointerStatus: StatusOK,
			}, nil
		default:
			return nil, &AnchorError{PointerPath: file, Reason: ReasonWrongType}
		}

	default:
		// Missing (or an unreadable target treated as missing).
		switch base {
		case anchorName:
			return &Anchor{
				Kind:          KindPointer,
				DBPath:        filepath.Join(target, "tasks.db"),
				PointerTarget: target,
				PointerStatus: StatusDangling,
			}, nil
		case "tasks.db":
			return &Anchor{
				Kind:          KindPointer,
				DBPath:        target,
				PointerTarget: target,
				PointerStatus: StatusDangling,
			}, nil
		default:
			// A wrong-named missing target can never become valid.
			return nil, &AnchorError{PointerPath: file, Reason: ReasonMalformed}
		}
	}
}

// IsProtectedDir reports whether dir is a place init must not anchor without
// --force: the user's home directory or the filesystem root. The home check
// reads os.UserHomeDir per call (not cached) so tests can set HOME; an error
// there skips the home check. The root test is portable: it holds for "/" on
// POSIX and for volume roots like "C:\" on Windows.
func IsProtectedDir(dir string) bool {
	cleaned := filepath.Clean(dir)
	if home, err := os.UserHomeDir(); err == nil && home != "" && cleaned == filepath.Clean(home) {
		return true
	}
	return filepath.Dir(cleaned) == cleaned
}
