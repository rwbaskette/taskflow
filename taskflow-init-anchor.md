# Design: `.taskflow` Anchor Resolution with Init Command

Author: praxis. Status: Implemented, revision 4. Date: 2026-09-15.
Supersedes the runtime part of `tmp/design/worktree-detection.md`.

Superseded note. This doc's revision 2 git-at-init rules, and the git-at-runtime detection and
boundary logic of `worktree-detection.md`, are superseded. That prior doc is fully historical: its
git facts, the busy_timeout DSN requirement, and its test conventions remain useful; no part of its
detection or boundary logic is used. Taskflow has zero git involvement: no binary, no `rev-parse`,
and no `.git` marker tests anywhere.

## 1. Problem and verdict

`DefaultDBPath()` in `internal/db/db.go` returns `.taskflow/tasks.db` relative to the cwd. Verified
fact: `taskflow add` in a deep subdirectory silently creates a new DB right there, so auto-create
scatters DBs; tasks in `<repo>/sub/a/.taskflow` are invisible from `<repo>/sub`. Maintainer
proposal: walk up looking for `.taskflow`; if not found, error with "run `taskflow init`".

Verdict: reasonable, simpler than the git-based detection in the prior doc, and fully git-free:
no git binary and no `.git` knowledge. Two mechanisms replace everything git did: the walk as a
guardrail (if an anchor already exists above the cwd, init refuses to create a second anchor;
`--force` overrides), and the pointer anchor for sharing (`--target <path>` writes a pointer file
explicitly, for example from a linked worktree). Taskflow supports worktrees only through the
pointer mechanism. Three weaknesses of the bare proposal need fixes:

1. Every checkout errors until init, and a naive init in a linked worktree makes an isolated DB;
   the pointer anchor (section 2) plus `--target` fixes this.
2. An unbounded walk can latch onto an unrelated ancestor `.taskflow`; section 5 documents this
   residual risk and its remedies honestly.
3. Breaking change: today every command auto-creates; the new plan errors until init. Existing
   users are unaffected (the walk finds their anchor at once); only fresh projects break (section 8).

## 2. Anchor model

**Directory anchor.** `.taskflow` is a directory. Its DB home is `<dir>/.taskflow/tasks.db`;
today's layout.

**Pointer-file anchor.** `.taskflow` is a regular file. It contains exactly one non-empty line:

```
database: <path>
```

`<path>` is absolute, or relative to the directory that holds the file (resolution:
`<dir-of-file>/<path>`). A valid target names an anchor home (a directory named `.taskflow`) or a
DB file (a file named `tasks.db`); taskflow appends `tasks.db` to a `.taskflow` directory target;
a pointer-file target is a hard error (no nested pointers). The pointer is the only cross-checkout
mechanism; taskflow has no worktree concept. Pointer rules: strip CR/LF and surrounding
whitespace, then require the `database:` prefix and a non-empty path; resolve with
`filepath.Abs`. Error tiers:

1. Malformed syntax (bad prefix, empty path, extra lines, nested pointer target): hard error; fix the file by hand.
2. Target missing on disk: "dangling", healable — expected in a fresh clone (pointer committed, DB
   gitignored). `taskflow init` repairs it (section 4); until then, commands fail with the
   pointer-error contract (section 6), never a silent fallback.
3. Target exists but with the wrong type (a plain directory or file that is neither a `.taskflow`
   directory nor a `tasks.db` file): hard error.

## 3. Resolution chain

Commands resolve the DB in this order:

1. `TASKFLOW_DIR` set and non-empty: `$TASKFLOW_DIR/tasks.db`. The walk is skipped entirely; env
   always wins; contract unchanged.
2. Walk up from the cwd (section 5) until an anchor is found or a stop rule fires. Dir anchor ->
   `<dir>/.taskflow/tasks.db`. Pointer anchor -> the resolved pointer target.
3. Not found, or a pointer error: exit code 2, error contract of section 6. No DB is created. Only
   `taskflow init` creates or repairs.

## 4. `taskflow init` spec

`taskflow init` creates or repairs the anchor for the current project. It is the only command that
creates files; it never runs git. Algorithm:

a. `TASKFLOW_DIR` set: create or open `$TASKFLOW_DIR/tasks.db`, print the path, exit 0; no anchor
   is written.
b. Run the anchor walk from the cwd (section 5), `filepath.EvalSymlinks` on the start dir first.
   This walk is the guardrail against scattered anchors.
c. Anchor found. Init never creates a second (shadow) anchor without `--force`:
   - Directory anchor: report its path; recreate a missing `tasks.db` (repair), exit 0. Pointer
     anchor: report the target; dangling -> create the DB at the target (repair), exit 0;
     malformed or wrong type -> pointer error, exit 2, file untouched.
   - Refusal message: when the found anchor sits above the cwd, init states "anchor already exists
     at <path>; no new anchor created" and names `--force`; exit 0.
   - `--force` never deletes a database. On a valid directory anchor AT the cwd, `--force` is a
     report-and-no-op: print the current anchor, exit 0. On a pointer file AT the cwd, `--force`
     replaces the pointer with a directory anchor (the pointer's target DB stays untouched). For
     an anchor above the cwd, `--force` creates a shadow directory anchor at the cwd and its DB;
     for a malformed pointer above the cwd this shadow is the remedy (the walk finds the shadow
     first). A nested anchor shadows the outer one for all subdirectories below it.
d. No anchor found:
   - `--target <path>`: validate the target kind BEFORE writing anything. The target must name an
     anchor home (a directory named `.taskflow`) or a DB file (a file named `tasks.db`). Plain
     directory, plain file, or pointer-file targets are refused (clear error, exit 2, nothing
     written). An existing `.taskflow` directory or `tasks.db` file is a valid anchor:
     create-if-missing, exit 0; a missing target dangles until created. After validation, write
     the pointer file FIRST: `.taskflow` at the cwd, one line `database: <path>` (absolute or
     cwd-relative). Then attempt DB creation at the target: writable parent -> create it with the
     schema, exit 0; unwritable parent -> keep the pointer and warn that it dangles.
   - Default: first the cheap guard. Init refuses to create an anchor at the user's home directory
     (`os.UserHomeDir`; plain Go, not git knowledge) or at the filesystem root, unless `--force`
     (this blocks agent mis-init and CI-cache strays, section 5): print "refusing to create an
     anchor here; pass --force to override", exit 2. Otherwise create a `.taskflow/` directory
     plus `tasks.db` at the cwd (reuse `NewDB`, which runs `migrate()`), print the resolved DB
     path, exit 0. Help text explains pointer files and cross-checkout sharing, so users discover
     `--target`.
   - Residual scatter, disclosed: the guardrail only refuses when an anchor exists above the cwd;
     `init` from a subdirectory with no anchor anywhere still creates the anchor at that cwd; the
     error text guides users to the project root (section 8).
e. Flag surface: only `--target <path>` and `--force`. There is no `--local`: everything that is
   not a pointer is local by default. There is no worktree detection in init.

Idempotency. Re-run must not wipe data; `--force` never deletes a database. Runtime commands stay
git-free. Repair failure is defined in section 6.

## 5. Walk rules

From the cwd upward, with `filepath.EvalSymlinks` on the start dir first (macOS maps `/tmp` to
`/private/tmp`; without this, path comparisons diverge). The walk has exactly two stop conditions:

1. At each directory D, test `<D>/.taskflow`. Directory anchor -> use it, stop. Pointer anchor ->
   resolve it: valid -> use it; dangling -> use it, and the command fails at open time with the
   pointer-error contract (section 6); malformed or wrong type -> hard error. Stop.
2. Filesystem root reached without an anchor: not found. Root termination is
   `filepath.Dir(D) == D`; repeated `Dir` calls return `C:\` on Windows, so the rule is portable.

Unreadable ancestors: a stat error on `<D>/.taskflow`, or an `EvalSymlinks` error (which falls
back to the raw cwd string), means "no anchor here, continue upward"; the walk never aborts.

Rationale for the unbounded walk. Anchors are explicit: only `init` creates them. An ancestor
`.taskflow` is almost always intentional; climbing to the project root is the desired behavior.

Residual latch risks, stated honestly. A stray ancestor anchor can still capture a nested project:
old auto-created anchors from before init existed (`$HOME/.taskflow`, or a subdirectory anchor
from the old cwd-relative default).
- Nested independent project capture: a separate checkout or repo nested under another project's
  tree, with no anchor of its own, latches the OUTER anchor and writes tasks into the wrong task
  list. Remedy: init every nested project at its root, so the walk finds its anchor first; mixed
  trees must anchor each project root.
- Agent mis-init and CI-cache strays: an agent may run init in `$HOME`, `/tmp`, or a workspace
  parent, and a CI cache can persist the stray anchor. The init guard against `$HOME` and the
  filesystem root (section 4) blocks the two worst spots.

Remedies for all cases: the recovery recipe (section 8), `taskflow where` for diagnosis, and
`TASKFLOW_DIR` as the override.

Edge cases: cwd equals a directory with no `.taskflow` above it -> not found. A pointer may escape
the project (`database: ../../elsewhere/db`); taskflow only resolves what the user wrote.

## 6. Error contract

When resolution fails, every command behaves the same. Two resolution variants and one init-only
variant; all exit code 2, all create no files (no silent auto-create anywhere):

- Not found: stderr says "no .taskflow anchor found"; then the searched directory chain, one per
  line, ending with "filesystem root reached"; then the remedy "run `taskflow init` in the project
  root; use `taskflow init --target <path>` to share one database across multiple checkouts"; then
  "set TASKFLOW_DIR to override".
- Pointer error: stderr names the pointer file path and the reason ("malformed", "dangling (target
  missing)", or "wrong target type"); remedy: dangling -> "run `taskflow init` to create or repair
  the database"; malformed or wrong type -> "fix by hand, or run `taskflow init --force` below it
  to shadow it".
- Repair failure (init only): a failed repair (a dangling heal or a missing `tasks.db` recreation
  that errors, for example on an unwritable parent) prints the underlying error, exits 2; the
  pointer and the anchor stay untouched.

Agent guidance: plugin users' cwd is the checkout root, so the remedy targets the cwd; manual
agents should init at the topmost directory of the project they edit.

Example, not-found (dangling pointers print the same shape):

```
taskflow: no .taskflow anchor found
searched:
  /proj/myrepo/sub
  /proj/myrepo
  /proj
  / (filesystem root reached, no .taskflow)
remedy: run `taskflow init` in the project root (--target <path> shares one DB across checkouts)
or set TASKFLOW_DIR to override
```

## 7. Debug command

Add `taskflow where`. One item per line: `db:` the absolute DB path; `anchor:` dir, pointer, or
env; `anchor path:` the directory that holds the anchor; `pointer target:` the resolved target;
`pointer status:` ok, dangling, malformed, or wrong-type; `search chain:` the directories walked,
with the stop reason. No git context is printed.

Exit 0 when an anchor resolves to an openable DB; exit 2 with the section 6 message otherwise. A
dangling pointer reports exit 2 with chain and status, so `where` diagnoses stray latches.

## 8. Backward compatibility

Resolution chain, in order: `TASKFLOW_DIR` -> anchor walk -> error. Today the chain is
`TASKFLOW_DIR` -> auto-create in cwd.

- Git-free property: taskflow has zero git involvement — no binary, no `.git` marker tests.
- Existing users keep working unchanged: their `.taskflow` sits in the project root, the walk finds
  it at once, dir anchor, same DB path as today.
- Scattered-DB users now get exit 2 with a clear remedy; the stray DBs still exist. Recovery recipe
  (also printed by `taskflow init --help`): run `taskflow init` in the project root; move each
  stray `tasks.db` into it and delete the stray dirs, or point at it with a pointer file
  `database: <straydir>/.taskflow`.
- Fresh projects: `taskflow add` no longer creates a DB; the error message carries the fix.
- Residual scatter and sharing: `init` from a subdirectory of a project with no anchor still makes
  a local anchor there (the error text guides users to the project root); a second checkout (a
  linked worktree) gets an isolated DB from plain init, and sharing requires
  `taskflow init --target <path>`, named in the help and the error remedy; taskflow never detects
  checkout relations.
- Stray anchors and agents: plugin users' cwd is the checkout root, so the remedy targets the cwd;
  manual agents should init at the topmost directory of the project they edit. Init refuses `$HOME`
  and the filesystem root as anchor sites (without `--force`); nested projects must anchor each
  root (section 5).
- `worktree-detection.md` status: fully historical; still relevant: the verified git facts,
  busy_timeout, and the test approach.

## 9. Implementation notes

- New package `internal/anchor`:
  - `type Kind int`: `KindDir`, `KindPointer`, `KindEnv`.
  - `type Anchor struct { Kind; DBPath string; AnchorPath string; PointerTarget string;
    PointerStatus string }` — `PointerStatus` is ok, dangling, malformed, or wrong-type; consumers
    (error printer, init repair, `where`) act on the status `Resolve` sets.
  - `Resolve(dir string) (*Anchor, *AnchorError)` — walks upward from `dir` (empty means the
    symlink-resolved cwd); not-found, malformed, and wrong-type return `(nil, *AnchorError)`;
    dangling returns the anchor with status dangling. Root termination is `filepath.Dir(D) == D`;
    stat/`EvalSymlinks` errors mean continue (section 5).
  - `type AnchorError struct { Chain []string; PointerPath string; Reason string }` — carries the
    chain for the section 6 printer; the printer renders the "filesystem root reached" line itself
    (no StopReason field). Sentinels `ErrNotFound`, `ErrMalformedPointer`, `ErrWrongTargetType`;
    `ErrDanglingPointer` only for init's repair-failure path.
  - `parsePointer(file string) (*Anchor, *AnchorError)` — the pointer-anchor rules of section 2,
    including CR/LF and whitespace trimming and the three error tiers.
  - `IsProtectedDir(dir string) bool` — true for the user's home (`os.UserHomeDir`) and the
    filesystem root; init's guard uses it. No git helper, no boundary test, no depth cap; no
    `exec.Command` in the package or `cmd/`.
- `internal/db/db.go`: `DefaultDBPath()` becomes `DefaultDBPath() (string, error)`, with the
  `TASKFLOW_DIR` branch inside it. Add `_pragma=busy_timeout(5000)` to the DSN in `NewDB`
  (mandatory for the shared-DB pointer case); other pragmas stay.
- Error plumbing. The existing sites call `cliErrors.HandleError` (`pkg/errors/errors.go`), which
  hardwires exit 1; `PrintError` adds colored text; anchor errors must not go through either. Each
  of the eight `cmd/*.go` call sites does `path, err := db.DefaultDBPath()`; on error it calls
  `cmd/root.go: printAnchorError(err)` unwraps the error, prints the section 6 variant, and exits 2
  directly; all other errors keep `HandleError` and exit 1. The plain
  section-6 format is deliberate; do not unify it with `HandleError`'s JSON or `PrintError`'s
  colored text (agents parse plain text). New `cmd/init.go` (section 4) and `cmd/where.go`
  (section 7).

## 10. Test plan

Unit tests, `internal/anchor/*_test.go`, temp dirs, no network:

- Unbounded walk: an anchor three levels up found, chain recorded; pointer form several levels up
  resolves. Env precedence and symlinked start dir covered. Unreadable ancestors: a
  permission-denied ancestor is skipped and the walk continues; an `EvalSymlinks` failure falls
  back to the raw cwd string; the walk never aborts.
- Pointer file, absolute target: resolves; relative target: resolves against the file's directory.
  Pointer parsing: CRLF and padded whitespace trimmed (one blank trailing line OK); malformed
  pointers (bad prefix, empty path, two lines, nested target) and wrong-type targets: hard error.
  Dangling: `Resolve` returns status dangling, and the command error carries the repair remedy.

Integration tests, `tests/cli/cli_test.go` style (built binary, `runCLI`, temp dirs):

- Error contract: `add` with no anchor anywhere: exit 2; stderr lists the chain up to the
  filesystem root and asserts the final "filesystem root reached" line, the init remedy with the
  `--target` sentence, and the `TASKFLOW_DIR` note; no files created. Pointer error contracts:
  malformed and dangling exit 2 with pointer path and reason.
- Dangling-pointer lifecycle: committed pointer, gitignored DB, fresh clone: commands exit 2;
  `init` repairs (exit 0); deleted target -> dangling contract, `init` repairs.
- Init walk-before-create refusal: anchor above the cwd; `init` exits 0, reports the existing
  anchor, creates no new anchor, and the message names `--force`. Default init in a fresh-repo
  subdirectory (residual scatter): `init` in `<repo>/sub` (no anchor anywhere) creates the anchor
  there; the DB lands at `<repo>/sub/.taskflow`.
- Protected-directory guard and stray latch: `init` in `$HOME` (temp home via env) and at a fake
  filesystem root refuses without `--force` (exit 2, clear message); with `--force` it creates the
  anchor. A nested unanchored project under a tree with a stray ancestor anchor latches the outer
  anchor; the recovery recipe restores correct resolution. A nested project init'd at its own root
  is found first, as designed.
- `--force`: shadow above the cwd (commands below use the nested DB, above the outer DB);
  replacement of a pointer file at the cwd (target DB untouched); report-and-no-op on a
  data-bearing directory anchor at the cwd (exit 0, `tasks.db` byte-identical); no touch of a
  malformed pointer above the cwd.
- `--target` validation and write, relative and absolute: plain directory refused (exit 2, nothing
  written); plain file refused; existing `.taskflow` directory or `tasks.db` -> create-if-missing,
  exit 0. Pointer written FIRST (exact `database:` line); DB created at the target when the parent
  is writable; unwritable -> pointer kept, warning printed, dangling contract; repair via init
  from another directory. Failed repair: `init` prints the underlying error, exits 2, and the
  pointer file is byte-identical afterwards.
- Init with `TASKFLOW_DIR` set: DB there, no anchor in the tree. Idempotency: add a task, run
  `init` again, task still listed.
- Sibling-worktree sharing, git-free: main repo anchor; fresh linked worktree (git for setup
  only); `init --target <main>/.taskflow/tasks.db`; tasks visible in both; taskflow never calls
  git (empty `PATH` works).
- `where`: anchor type, pointer target, pointer status, chain for all kinds; exit 0 on resolution,
  exit 2 on errors. Concurrency smoke: N parallel `add` calls on one shared pointer DB, no DB file
  at start; all N tasks exist; no `SQLITE_BUSY` (busy_timeout pragma and WAL).

Existing tests that need updates (name them):

- `internal/db/default_path_test.go`: all six tests break when `DefaultDBPath()` returns two
  values — `TestDefaultDBPath_WithEnvVar`, `TestDefaultDBPath_WithoutEnvVar`,
  `TestDefaultDBPath_WithEmptyEnvVar`, `TestDefaultDBPath_AbsolutePath`,
  `TestDefaultDBPath_RelativePath`, `TestDefaultDBPath_TrailingSlash`. Rework the whole file
  around the two-valued return; env cases add an `err == nil` check; no-env cases move to
  `internal/anchor` tests with temp-dir anchors.
- `tests/cli/cli_test.go`: the `runCLI` docs say the CLI creates an isolated `.taskflow/tasks.db`;
  tests relying on auto-create must call `init` first or set `TASKFLOW_DIR` (same for
  `cmd/*_test.go`).

## 11. Open questions

1. Should `--target` create the target DB when missing? Current design: yes, when the parent is
   writable; else a dangling pointer with a warning. Alternative: never create.
2. Pointer file committed or gitignored? Recommendation: commit it, with a relative `database:`
   path (relative paths travel with clones). This repo's `.gitignore` entry is `.taskflow/`
   (trailing slash), which matches directories only: a committed pointer file named `.taskflow` is
   not ignored, and `tasks.db` stays ignored (the dangling-repair flow in section 4 covers the
   fresh-clone gap).
3. Plugin wrapper auto-init? Recommendation: no auto-init; surface the exit 2 text (the remedy
   tells the agent to run init). If auto behavior is wanted later, restrict it to
   opencode-worktree-root == anchor-directory; comparable without git.
4. Read-only commands: softer error? Recommendation: no; one contract is simpler to teach.
   Revisit only if users complain.
5. Pointer target forms: the name rule (`.taskflow` home, or `tasks.db` file) settles much of
   this. Remaining: drop the DB-file form? Simpler, but it breaks pointers aimed at a DB outside
   any `.taskflow` home. Decide before implementation.
6. Walk depth cap: none. The only stop is the filesystem root (one cheap `stat` per directory); a
   cap is optional future hardening.

## 12. Implementation record

Implemented 2026-09-15 as specified, with these recorded resolutions (review-confirmed):

- `Anchor` also carries `Chain []string` (section 9's struct did not list it): section 7 requires
  `where` to print the search chain for all kinds, including successful resolution, and
  `AnchorError` alone cannot carry it on success. The "printer renders the root line itself, no
  StopReason field" rule is honored.
- New error reason `missing-db`: resolution succeeds but the DB file under the anchor is absent
  (a deleted `tasks.db`, or a pointer to an existing `.taskflow` home whose DB was removed).
  Runtime commands exit 2 with "taskflow: <db> is missing" plus the init remedy; only `init`
  repairs (it also repairs this case for status-ok pointers, mirroring the section 4c dir-anchor
  repair). The `TASKFLOW_DIR` branch is untouched: env keeps today's auto-create contract. This
  enforces section 3's "no DB is created" rule against `NewDB`'s open-create behavior; the
  section 10 concurrency test therefore pre-creates the shared DB instead of starting without one.
- Relative pointer targets resolve with `filepath.Join(holder, path)` + `filepath.Clean` —
  equivalent to, and more correct than, the literal `filepath.Abs` wording of section 2 (Abs
  would resolve against the process cwd, not the pointer's directory).
- A missing pointer target whose base name is neither `.taskflow` nor `tasks.db` is malformed
  (tier 1), not dangling: it can never become valid, so dangling would be a dead end.
- Init ambiguity resolutions: unwritable `--target` parent keeps the pointer, warns, exit 0
  (open question 1's "create when the parent is writable; else a dangling pointer with a
  warning"); the protected-dir guard applies to the `--target` branch too (a pointer at `$HOME`
  is the same stray-anchor hazard); `--target` is ignored, with a note, when an anchor already
  exists; `--force` on a dangling pointer at the cwd heals it, above the cwd it shadows;
  `--force` replaces only a well-formed pointer at the cwd and restores the pointer bytes if the
  DB creation fails — the restore is self-healing (a run-created anchor directory is removed
  first, pre-existing directories are never removed, and restore failures surface in the error
  message); malformed pointers are hard errors everywhere — shadowing is the only
  `--force` remedy above the cwd.
- `where` with `TASKFLOW_DIR` set exits 0 without an existence check (env always wins; `where`
  is diagnostic, never a creator). For walk-derived anchors, `where` stats the DB and reports
  `db status: missing` with exit 2 when the file is absent.
- `ErrDanglingPointer` now has two producers via `Unwrap`: the init repair-failure path and the
  `missing-db` walk case (the section 9 "only init's repair-failure path" note is superseded).
