package version

import (
	"regexp"
	"strings"
	"testing"
)

func TestVersionIsSemver(t *testing.T) {
	re := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	if !re.MatchString(Version) {
		t.Errorf("Version = %q, want a semver string matching %q", Version, re.String())
	}
}

func TestVersionFileFormat(t *testing.T) {
	if Version == "" {
		t.Fatal("Version must not be empty")
	}
	if !strings.HasSuffix(raw, "\n") {
		t.Errorf("VERSION = %q, must end with a newline", raw)
	}
	if raw != Version+"\n" {
		t.Errorf("VERSION = %q, want exactly %q plus one trailing newline", raw, Version)
	}
}
