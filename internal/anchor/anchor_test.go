package anchor

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writePointer writes a pointer file with the given content at dir/.taskflow.
func writePointer(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, anchorName), []byte(content), 0o644); err != nil {
		t.Fatalf("write pointer file: %v", err)
	}
}

// makeAnchorHome creates a directory named .taskflow (the anchor home).
func makeAnchorHome(t *testing.T, dir string) string {
	t.Helper()
	home := filepath.Join(dir, anchorName)
	if err := os.Mkdir(home, 0o755); err != nil {
		t.Fatalf("mkdir anchor home: %v", err)
	}
	return home
}

// chdirAt sets the cwd to dir for the duration of the test.
func chdirAt(t *testing.T, dir string) {
	t.Helper()
	t.Chdir(dir)
}

func TestResolve_FindsDirAnchorThreeLevelsUp(t *testing.T) {
	base := t.TempDir()
	makeAnchorHome(t, base)
	l1 := filepath.Join(base, "l1")
	l2 := filepath.Join(l1, "l2")
	l3 := filepath.Join(l2, "l3")
	for _, d := range []string{l1, l2, l3} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	a, aerr := Resolve(l3)
	if aerr != nil {
		t.Fatalf("Resolve: unexpected error: %v", aerr)
	}
	if a.Kind != KindDir {
		t.Errorf("Kind = %v, want KindDir", a.Kind)
	}
	if want := filepath.Join(base, anchorName, "tasks.db"); a.DBPath != want {
		t.Errorf("DBPath = %q, want %q", a.DBPath, want)
	}
	if a.AnchorPath != base {
		t.Errorf("AnchorPath = %q, want %q", a.AnchorPath, base)
	}
	if a.PointerStatus != StatusOK {
		t.Errorf("PointerStatus = %q, want %q", a.PointerStatus, StatusOK)
	}
	if a.PointerTarget != "" {
		t.Errorf("PointerTarget = %q, want empty", a.PointerTarget)
	}
	wantChain := []string{l3, l2, l1, base}
	if len(a.Chain) != len(wantChain) {
		t.Fatalf("chain = %v, want %v", a.Chain, wantChain)
	}
	for i, want := range wantChain {
		if a.Chain[i] != want {
			t.Errorf("chain[%d] = %q, want %q", i, a.Chain[i], want)
		}
	}
}

func TestResolve_PointerAnchorSeveralLevelsUp(t *testing.T) {
	// The shared anchor lives outside the walked tree; the pointer file sits
	// two levels up from the start dir.
	shared := t.TempDir()
	home := makeAnchorHome(t, shared)

	tree := t.TempDir()
	sub := filepath.Join(tree, "sub")
	deep := filepath.Join(sub, "deep")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writePointer(t, tree, "database: "+home+"\n")

	a, aerr := Resolve(deep)
	if aerr != nil {
		t.Fatalf("Resolve: unexpected error: %v", aerr)
	}
	if a.Kind != KindPointer {
		t.Errorf("Kind = %v, want KindPointer", a.Kind)
	}
	if a.PointerTarget != home {
		t.Errorf("PointerTarget = %q, want %q", a.PointerTarget, home)
	}
	if want := filepath.Join(home, "tasks.db"); a.DBPath != want {
		t.Errorf("DBPath = %q, want %q", a.DBPath, want)
	}
	if a.AnchorPath != tree {
		t.Errorf("AnchorPath = %q, want %q", a.AnchorPath, tree)
	}
	if a.PointerStatus != StatusOK {
		t.Errorf("PointerStatus = %q, want %q", a.PointerStatus, StatusOK)
	}
	wantChain := []string{deep, sub, tree}
	if len(a.Chain) != len(wantChain) {
		t.Fatalf("chain = %v, want %v", a.Chain, wantChain)
	}
	for i, want := range wantChain {
		if a.Chain[i] != want {
			t.Errorf("chain[%d] = %q, want %q", i, a.Chain[i], want)
		}
	}
}

func TestResolve_RelativePointerTarget(t *testing.T) {
	base := t.TempDir()
	sharedHome := filepath.Join(base, "shared", anchorName)
	if err := os.MkdirAll(sharedHome, 0o755); err != nil {
		t.Fatalf("mkdir shared home: %v", err)
	}

	t.Run("relative to the pointer file's directory", func(t *testing.T) {
		writePointer(t, base, "database: shared/.taskflow\n")
		a, aerr := Resolve(base)
		if aerr != nil {
			t.Fatalf("Resolve: unexpected error: %v", aerr)
		}
		if a.PointerTarget != sharedHome {
			t.Errorf("PointerTarget = %q, want %q", a.PointerTarget, sharedHome)
		}
		if want := filepath.Join(sharedHome, "tasks.db"); a.DBPath != want {
			t.Errorf("DBPath = %q, want %q", a.DBPath, want)
		}
		if a.PointerStatus != StatusOK {
			t.Errorf("PointerStatus = %q, want %q", a.PointerStatus, StatusOK)
		}
	})

	t.Run("dot-dot relative from a subdirectory", func(t *testing.T) {
		sub := filepath.Join(base, "sub")
		if err := os.Mkdir(sub, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		// ../shared/.taskflow from the pointer file's directory (sub)
		// resolves to base/shared/.taskflow.
		writePointer(t, sub, "database: ../shared/.taskflow\n")
		a, aerr := Resolve(sub)
		if aerr != nil {
			t.Fatalf("Resolve: unexpected error: %v", aerr)
		}
		if a.PointerTarget != sharedHome {
			t.Errorf("PointerTarget = %q, want %q", a.PointerTarget, sharedHome)
		}
		if a.AnchorPath != sub {
			t.Errorf("AnchorPath = %q, want %q", a.AnchorPath, sub)
		}
	})
}

func TestParsePointer_SyntaxAndWhitespace(t *testing.T) {
	tmp := t.TempDir()
	validHome := makeAnchorHome(t, tmp)

	tests := []struct {
		name    string
		content string
		want    string // resolved target; empty means an error is expected
	}{
		{
			name:    "plain absolute",
			content: "database: " + validHome + "\n",
			want:    validHome,
		},
		{
			name:    "crlf and padded whitespace",
			content: "database:  " + validHome + "  \r\n",
			want:    validHome,
		},
		{
			name:    "no space after prefix",
			content: "database:" + validHome + "\n",
			want:    validHome,
		},
		{
			name:    "one blank trailing line ok",
			content: "database: " + validHome + "\n\n",
			want:    validHome,
		},
		{
			name:    "no trailing newline ok",
			content: "database: " + validHome,
			want:    validHome,
		},
		{
			name:    "bad prefix",
			content: "path: " + validHome + "\n",
			want:    "",
		},
		{
			name:    "empty path",
			content: "database:   \n",
			want:    "",
		},
		{
			name:    "empty file",
			content: "",
			want:    "",
		},
		{
			name:    "only blank lines",
			content: "\n\n\r\n",
			want:    "",
		},
		{
			name:    "two non-empty lines",
			content: "database: " + validHome + "\ndatabase: " + validHome + "\n",
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			file := filepath.Join(dir, anchorName)
			writePointer(t, dir, tt.content)

			a, aerr := parsePointer(file)
			if tt.want == "" {
				if aerr == nil {
					t.Fatalf("parsePointer(%q) = success, want malformed error", tt.content)
				}
				if aerr.Reason != ReasonMalformed {
					t.Errorf("Reason = %q, want %q", aerr.Reason, ReasonMalformed)
				}
				if aerr.PointerPath != file {
					t.Errorf("PointerPath = %q, want %q", aerr.PointerPath, file)
				}
				return
			}
			if aerr != nil {
				t.Fatalf("parsePointer(%q): unexpected error: %v", tt.content, aerr)
			}
			if a.PointerTarget != tt.want {
				t.Errorf("PointerTarget = %q, want %q", a.PointerTarget, tt.want)
			}
			if want := filepath.Join(tt.want, "tasks.db"); a.DBPath != want {
				t.Errorf("DBPath = %q, want %q", a.DBPath, want)
			}
			if a.PointerStatus != StatusOK {
				t.Errorf("PointerStatus = %q, want %q", a.PointerStatus, StatusOK)
			}
			if a.AnchorPath != dir {
				t.Errorf("AnchorPath = %q, want %q", a.AnchorPath, dir)
			}
		})
	}
}

func TestParsePointer_NestedPointerTarget(t *testing.T) {
	tmp := t.TempDir()
	// The target is itself a .taskflow regular file: nested pointer, malformed.
	nested := filepath.Join(tmp, "other")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nested, anchorName), []byte("database: /x\n"), 0o644); err != nil {
		t.Fatalf("write nested pointer: %v", err)
	}

	dir := t.TempDir()
	writePointer(t, dir, "database: "+filepath.Join(nested, anchorName)+"\n")
	_, aerr := parsePointer(filepath.Join(dir, anchorName))
	if aerr == nil {
		t.Fatal("parsePointer = success, want malformed error for nested pointer target")
	}
	if aerr.Reason != ReasonMalformed {
		t.Errorf("Reason = %q, want %q", aerr.Reason, ReasonMalformed)
	}
}

func TestParsePointer_WrongTypeTargets(t *testing.T) {
	tmp := t.TempDir()
	plainDir := filepath.Join(tmp, "plaindir")
	if err := os.Mkdir(plainDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	plainFile := filepath.Join(tmp, "plainfile")
	if err := os.WriteFile(plainFile, []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	dbDir := filepath.Join(tmp, "tasks.db")
	if err := os.Mkdir(dbDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	tests := []struct {
		name   string
		target string
	}{
		{name: "plain directory with wrong name", target: plainDir},
		{name: "plain file with wrong name", target: plainFile},
		{name: "directory named tasks.db", target: dbDir},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writePointer(t, dir, "database: "+tt.target+"\n")
			_, aerr := parsePointer(filepath.Join(dir, anchorName))
			if aerr == nil {
				t.Fatal("parsePointer = success, want wrong-type error")
			}
			if aerr.Reason != ReasonWrongType {
				t.Errorf("Reason = %q, want %q", aerr.Reason, ReasonWrongType)
			}
		})
	}
}

func TestParsePointer_ExistingDBFileTargetIsValid(t *testing.T) {
	tmp := t.TempDir()
	dbFile := filepath.Join(tmp, "shared", "tasks.db")
	if err := os.MkdirAll(filepath.Dir(dbFile), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(dbFile, []byte("placeholder"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	dir := t.TempDir()
	writePointer(t, dir, "database: "+dbFile+"\n")
	a, aerr := parsePointer(filepath.Join(dir, anchorName))
	if aerr != nil {
		t.Fatalf("parsePointer: unexpected error: %v", aerr)
	}
	if a.PointerStatus != StatusOK {
		t.Errorf("PointerStatus = %q, want %q", a.PointerStatus, StatusOK)
	}
	if a.DBPath != dbFile {
		t.Errorf("DBPath = %q, want %q", a.DBPath, dbFile)
	}
	if a.PointerTarget != dbFile {
		t.Errorf("PointerTarget = %q, want %q", a.PointerTarget, dbFile)
	}
	if a.Kind != KindPointer {
		t.Errorf("Kind = %v, want %v", a.Kind, KindPointer)
	}
}

func TestParsePointer_MissingValidNamesDangling(t *testing.T) {
	tmp := t.TempDir()
	missingHome := filepath.Join(tmp, "gone", anchorName)
	missingDB := filepath.Join(tmp, "elsewhere", "tasks.db")

	tests := []struct {
		name       string
		target     string
		wantDBPath string
	}{
		{
			name:       "missing .taskflow home",
			target:     missingHome,
			wantDBPath: filepath.Join(missingHome, "tasks.db"),
		},
		{
			name:       "missing tasks.db file",
			target:     missingDB,
			wantDBPath: missingDB,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writePointer(t, dir, "database: "+tt.target+"\n")
			a, aerr := parsePointer(filepath.Join(dir, anchorName))
			if aerr != nil {
				t.Fatalf("parsePointer: unexpected error: %v", aerr)
			}
			if a.PointerStatus != StatusDangling {
				t.Errorf("PointerStatus = %q, want %q", a.PointerStatus, StatusDangling)
			}
			if a.DBPath != tt.wantDBPath {
				t.Errorf("DBPath = %q, want %q", a.DBPath, tt.wantDBPath)
			}
			if a.PointerTarget != tt.target {
				t.Errorf("PointerTarget = %q, want %q", a.PointerTarget, tt.target)
			}
		})
	}
}

func TestParsePointer_MissingWrongNameMalformed(t *testing.T) {
	dir := t.TempDir()
	writePointer(t, dir, "database: "+filepath.Join(t.TempDir(), "straydb.sqlite")+"\n")
	_, aerr := parsePointer(filepath.Join(dir, anchorName))
	if aerr == nil {
		t.Fatal("parsePointer = success, want malformed error for wrong-named missing target")
	}
	if aerr.Reason != ReasonMalformed {
		t.Errorf("Reason = %q, want %q", aerr.Reason, ReasonMalformed)
	}
}

func TestParsePointer_ReadErrorAfterStatIsMalformed(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root: permission checks do not apply")
	}
	dir := t.TempDir()
	file := filepath.Join(dir, anchorName)
	writePointer(t, dir, "database: /x\n")
	if err := os.Chmod(file, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(file, 0o644)
	})
	_, aerr := parsePointer(file)
	if aerr == nil {
		t.Fatal("parsePointer = success, want malformed error on unreadable file")
	}
	if aerr.Reason != ReasonMalformed {
		t.Errorf("Reason = %q, want %q", aerr.Reason, ReasonMalformed)
	}
}

func TestResolve_DanglingPointerStopsWalk(t *testing.T) {
	base := t.TempDir()
	// Target missing on disk; the pointer sits at base.
	writePointer(t, base, "database: "+filepath.Join(base, "gone", anchorName)+"\n")
	start := filepath.Join(base, "sub")
	if err := os.Mkdir(start, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	a, aerr := Resolve(start)
	if aerr != nil {
		t.Fatalf("Resolve: unexpected error for dangling pointer: %v", aerr)
	}
	if a.Kind != KindPointer {
		t.Errorf("Kind = %v, want KindPointer", a.Kind)
	}
	if a.PointerStatus != StatusDangling {
		t.Errorf("PointerStatus = %q, want %q", a.PointerStatus, StatusDangling)
	}
	if want := filepath.Join(base, "gone", anchorName, "tasks.db"); a.DBPath != want {
		t.Errorf("DBPath = %q, want where the DB would live: %q", a.DBPath, want)
	}
	// The walk stops at the pointer: base's parent must not be in the chain.
	wantChain := []string{start, base}
	if len(a.Chain) != len(wantChain) {
		t.Fatalf("chain = %v, want %v (walk must stop at the pointer)", a.Chain, wantChain)
	}
}

func TestResolve_SymlinkedStartDir(t *testing.T) {
	base := t.TempDir()
	real := filepath.Join(base, "real")
	if err := os.Mkdir(real, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	makeAnchorHome(t, real)
	link := filepath.Join(base, "link")
	if err := os.Symlink(real, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	chdirAt(t, link)
	a, aerr := Resolve("")
	if aerr != nil {
		t.Fatalf("Resolve(\"\"): unexpected error: %v", aerr)
	}
	// EvalSymlinks on the start dir must resolve link to real.
	if a.AnchorPath != real {
		t.Errorf("AnchorPath = %q, want resolved real path %q", a.AnchorPath, real)
	}
	if a.Chain[0] != real {
		t.Errorf("chain[0] = %q, want %q (symlink resolved)", a.Chain[0], real)
	}
}

func TestEvalStart_FallsBackToRawString(t *testing.T) {
	base := t.TempDir()
	real := filepath.Join(base, "real")
	if err := os.Mkdir(real, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	dangling := filepath.Join(real, "dangling")
	if err := os.Symlink(filepath.Join(base, "no-such-target"), dangling); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	// EvalSymlinks fails on a path with a dangling component; the raw string
	// is used as fallback.
	if got := evalStart(dangling); got != dangling {
		t.Errorf("evalStart(%q) = %q, want the raw path (fallback)", dangling, got)
	}
}

func TestResolve_DanglingSymlinkInStartPathDoesNotAbort(t *testing.T) {
	base := t.TempDir()
	makeAnchorHome(t, base)
	real := filepath.Join(base, "real")
	if err := os.Mkdir(real, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	dangling := filepath.Join(real, "dangling")
	if err := os.Symlink(filepath.Join(base, "no-such-target"), dangling); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	a, aerr := Resolve(dangling)
	if aerr != nil {
		t.Fatalf("Resolve: unexpected error: %v", aerr)
	}
	if a.Kind != KindDir {
		t.Errorf("Kind = %v, want KindDir", a.Kind)
	}
	if a.AnchorPath != base {
		t.Errorf("AnchorPath = %q, want %q", a.AnchorPath, base)
	}
	// The raw (unresolved) start path leads the chain; the walk continued
	// past the dangling component.
	if a.Chain[0] != dangling {
		t.Errorf("chain[0] = %q, want raw start path %q", a.Chain[0], dangling)
	}
	if a.Chain[len(a.Chain)-1] != base {
		t.Errorf("last chain entry = %q, want %q", a.Chain[len(a.Chain)-1], base)
	}
}

func TestResolve_PermissionDeniedAncestorDoesNotAbort(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root: permission-denied ancestor cannot be simulated")
	}
	base := t.TempDir()
	makeAnchorHome(t, base)
	secret := filepath.Join(base, "secret")
	deep := filepath.Join(secret, "deep")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Chmod(secret, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(secret, 0o755)
	})

	a, aerr := Resolve(deep)
	if aerr != nil {
		t.Fatalf("Resolve: unexpected error (walk must not abort): %v", aerr)
	}
	if a.AnchorPath != base {
		t.Errorf("AnchorPath = %q, want %q", a.AnchorPath, base)
	}
	// The denied ancestor is probed (and recorded) but does not stop the walk.
	found := false
	for _, d := range a.Chain {
		if d == secret {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("chain %v does not include the denied ancestor %q", a.Chain, secret)
	}
}

func TestResolve_PermissionDeniedAncestorNotFoundAtRoot(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root: permission-denied ancestor cannot be simulated")
	}
	// No anchor anywhere on the way up (the temp tree is bare); the denied
	// ancestor is skipped and the walk ends with not-found at the root.
	base := t.TempDir()
	secret := filepath.Join(base, "secret")
	deep := filepath.Join(secret, "deep")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Chmod(secret, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(secret, 0o755)
	})

	_, aerr := Resolve(deep)
	if aerr == nil {
		t.Fatal("Resolve = success, want not-found")
	}
	if aerr.Reason != ReasonNotFound {
		t.Errorf("Reason = %q, want %q", aerr.Reason, ReasonNotFound)
	}
	if last := aerr.Chain[len(aerr.Chain)-1]; filepath.Dir(last) != last {
		t.Errorf("last chain entry = %q, want the filesystem root", last)
	}
}

func TestResolve_NotFoundChainEndsAtRoot(t *testing.T) {
	// A bare temp dir: no anchor anywhere on the way up (the repo's own
	// .taskflow is not an ancestor of t.TempDir()).
	tmp := t.TempDir()
	_, aerr := Resolve(tmp)
	if aerr == nil {
		t.Fatal("Resolve = success, want not-found")
	}
	if aerr.PointerPath != "" {
		t.Errorf("PointerPath = %q, want empty", aerr.PointerPath)
	}
	if len(aerr.Chain) == 0 || aerr.Chain[0] != tmp {
		t.Fatalf("chain = %v, want to start at %q", aerr.Chain, tmp)
	}
	last := aerr.Chain[len(aerr.Chain)-1]
	if filepath.Dir(last) != last {
		t.Errorf("last chain entry = %q, want the filesystem root", last)
	}
}

func TestIsProtectedDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	tests := []struct {
		name string
		dir  string
		want bool
	}{
		{name: "home directory", dir: home, want: true},
		{name: "home directory with trailing separator", dir: home + string(os.PathSeparator), want: true},
		{name: "filesystem root", dir: "/", want: true},
		{name: "temp dir", dir: t.TempDir(), want: false},
		{name: "subdirectory of home", dir: filepath.Join(home, "sub"), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsProtectedDir(tt.dir); got != tt.want {
				t.Errorf("IsProtectedDir(%q) = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}

func TestAnchorError_ErrorAndUnwrapping(t *testing.T) {
	// Not-found from a bare tree.
	tmp := t.TempDir()
	_, err := Resolve(tmp)
	if err == nil {
		t.Fatal("Resolve = success, want not-found")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("errors.Is(err, ErrNotFound) = false, want true (err = %v)", err)
	}
	var ae *AnchorError
	if !errors.As(err, &ae) {
		t.Fatalf("errors.As(err, *AnchorError) = false, want true")
	}
	if ae.Reason != ReasonNotFound {
		t.Errorf("Reason = %q, want %q", ae.Reason, ReasonNotFound)
	}
	if ae.Error() == "" {
		t.Error("Error() renders an empty message")
	}

	// Malformed pointer.
	malDir := t.TempDir()
	writePointer(t, malDir, "path: /x\n")
	_, err = Resolve(malDir)
	if err == nil {
		t.Fatal("Resolve = success, want malformed")
	}
	if !errors.Is(err, ErrMalformedPointer) {
		t.Errorf("errors.Is(err, ErrMalformedPointer) = false (err = %v)", err)
	}
	ae = nil
	if !errors.As(err, &ae) {
		t.Fatal("errors.As(err, *AnchorError) = false")
	}
	if ae.PointerPath == "" {
		t.Error("PointerPath empty for a pointer error")
	}

	// Wrong-type pointer target.
	wtDir := t.TempDir()
	plain := filepath.Join(wtDir, "plaindir")
	if mkErr := os.Mkdir(plain, 0o755); mkErr != nil {
		t.Fatalf("mkdir: %v", mkErr)
	}
	writePointer(t, wtDir, "database: "+plain+"\n")
	_, err = Resolve(wtDir)
	if err == nil {
		t.Fatal("Resolve = success, want wrong-type")
	}
	if !errors.Is(err, ErrWrongTargetType) {
		t.Errorf("errors.Is(err, ErrWrongTargetType) = false (err = %v)", err)
	}

	// Dangling resolves to an anchor with a nil error, never ErrDanglingPointer.
	dgDir := t.TempDir()
	writePointer(t, dgDir, "database: "+filepath.Join(dgDir, "gone", anchorName)+"\n")
	a, err := Resolve(dgDir)
	if err != nil {
		t.Fatalf("Resolve on dangling pointer returned error %v, want nil", err)
	}
	if !IsDanglingPointer(ErrDanglingPointer) {
		t.Error("IsDanglingPointer does not detect ErrDanglingPointer")
	}
	if a.PointerStatus != StatusDangling {
		t.Errorf("PointerStatus = %q, want %q", a.PointerStatus, StatusDangling)
	}
}

func TestAnchorError_ErrorMessages(t *testing.T) {
	tests := []struct {
		name string
		e    *AnchorError
	}{
		{
			name: "not-found",
			e:    &AnchorError{Chain: []string{"/a/b", "/a", "/"}, Reason: ReasonNotFound},
		},
		{
			name: "malformed",
			e:    &AnchorError{Chain: []string{"/a"}, PointerPath: "/a/.taskflow", Reason: ReasonMalformed},
		},
		{
			name: "wrong-type",
			e:    &AnchorError{Chain: []string{"/a"}, PointerPath: "/a/.taskflow", Reason: ReasonWrongType},
		},
		{
			name: "dangling",
			e:    &AnchorError{Chain: []string{"/a"}, PointerPath: "/a/.taskflow", Reason: ReasonDangling},
		},
		{
			name: "missing-db",
			e:    &AnchorError{Chain: []string{"/a"}, PointerPath: "/a/.taskflow/tasks.db", Reason: ReasonMissingDB},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.e.Error()
			if msg == "" {
				t.Fatal("Error() renders an empty message")
			}
			if tt.e.PointerPath != "" && !strings.Contains(msg, tt.e.PointerPath) {
				t.Errorf("Error() = %q, want it to name the pointer path", msg)
			}
			if tt.e.Reason == ReasonNotFound && !strings.Contains(msg, "/a/b") {
				t.Errorf("Error() = %q, want it to name the start dir", msg)
			}
		})
	}
}

func TestKindConstants(t *testing.T) {
	if KindDir != 0 || KindPointer != 1 || KindEnv != 2 {
		t.Errorf("Kind constants drifted: Dir=%d Pointer=%d Env=%d", KindDir, KindPointer, KindEnv)
	}
}
