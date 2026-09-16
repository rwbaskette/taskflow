package cmd

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

func TestVersionCmdExists(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "version" {
			return
		}
	}
	t.Error("Expected 'version' subcommand to exist on rootCmd")
}

func TestVersionCmdConfig(t *testing.T) {
	if versionCmd.Use != "version" {
		t.Errorf("Use = %v, want %v", versionCmd.Use, "version")
	}
	if versionCmd.Short != "Print the taskflow version" {
		t.Errorf("Short = %v, want %v", versionCmd.Short, "Print the taskflow version")
	}
	if versionCmd.Args == nil {
		t.Error("Expected Args validator to be set")
	}
}

func TestVersionCmdOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"version"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if strings.Contains(output, "\n") {
		t.Errorf("version output must be a single line, got:\n%s", output)
	}

	re := regexp.MustCompile(`^taskflow version \d+\.\d+\.\d+$`)
	if !re.MatchString(output) {
		t.Errorf("version output = %q, want match for %v", output, re)
	}
}
