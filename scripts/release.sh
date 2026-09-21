#!/bin/sh
# scripts/release.sh — bump the embedded VERSION, test, commit, and tag.
# Fully local: no network, no git host. Usage: scripts/release.sh X.Y.Z
set -eu

# Always operate from the repo root, so the script works from any cwd.
cd "$(dirname "$0")/.." || exit 1

if [ "$#" -ne 1 ]; then
	echo "usage: $0 X.Y.Z" >&2
	exit 2
fi
V="$1"
printf '%s\n' "$V" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$' || {
	echo "not semver: $V" >&2
	exit 2
}

if git rev-parse -q --verify "refs/tags/v$V" >/dev/null 2>&1; then
	echo "tag v$V already exists" >&2
	exit 2
fi

VERSION_FILE="internal/version/VERSION"
CUR="$(cat "$VERSION_FILE")"
if [ "$CUR" = "$V" ]; then
	echo "VERSION is already $V; nothing to do" >&2
	exit 2
fi

if [ "$(printf '%s\n%s\n' "$CUR" "$V" | sort -V | head -n 1)" = "$V" ]; then
	echo "new version $V is not greater than current $CUR" >&2
	exit 2
fi

if ! git diff --quiet || ! git diff --cached --quiet; then
	echo "uncommitted changes present; commit or stash first" >&2
	exit 2
fi

printf '%s\n' "$V" > "$VERSION_FILE"
if ! go test ./...; then
	printf '%s\n' "$CUR" > "$VERSION_FILE"
	echo "go test failed; VERSION restored to $CUR" >&2
	exit 1
fi

git add "$VERSION_FILE"
git commit -m "release: v$V"
git tag "v$V"
echo "tagged v$V (local only; publish later with: git push <remote> v$V)"
