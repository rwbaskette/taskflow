// Package version holds the taskflow version as a single source of truth.
//
// The value lives in the VERSION file next to this file and is embedded at
// compile time. Release flow: edit VERSION, run go test ./..., commit, and
// tag vX.Y.Z (scripts/release.sh does all three; tags are local-only).
package version

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var raw string

// Version is the taskflow version as a semantic version string (e.g. "0.1.0").
var Version = strings.TrimSpace(raw)
