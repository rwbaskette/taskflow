package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rwbaskette/taskflow/internal/anchor"
	"github.com/spf13/cobra"
)

func TestRootCmdVersion(t *testing.T) {
	// Test that version is set correctly
	if rootCmd.Version != "0.1.0" {
		t.Errorf("Version = %v, want %v", rootCmd.Version, "0.1.0")
	}
}

func TestRootCmdUse(t *testing.T) {
	if rootCmd.Use != "taskflow" {
		t.Errorf("Use = %v, want %v", rootCmd.Use, "taskflow")
	}
}

func TestRootCmdShort(t *testing.T) {
	expectedShort := "A task management CLI tool"
	if rootCmd.Short != expectedShort {
		t.Errorf("Short = %v, want %v", rootCmd.Short, expectedShort)
	}
}

func TestRootCmdLong(t *testing.T) {
	if rootCmd.Long == "" {
		t.Error("Long should not be empty")
	}
}

func TestRootCmdArgs(t *testing.T) {
	// Test that running with no args triggers help
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{})
	rootCmd.Execute()

	// Should have called help
	output := buf.String()
	if output == "" {
		t.Log("Root command with no args - checking help output")
	}
}

func TestRootCmdVersionTemplate(t *testing.T) {
	// Test version template is set
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--version"})
	rootCmd.Execute()

	output := buf.String()
	if output == "" {
		t.Error("Expected version output")
	}
}

func TestPersistentFlags(t *testing.T) {
	// The dead --config and --verbose flags were removed: neither was ever
	// read (cfgFile referenced a config.yaml that does not exist).
	if flag := rootCmd.PersistentFlags().Lookup("config"); flag != nil {
		t.Error("Expected 'config' flag to be removed")
	}
	if flag := rootCmd.PersistentFlags().Lookup("verbose"); flag != nil {
		t.Error("Expected 'verbose' flag to be removed")
	}
}

func TestRootCmdHasSubcommands(t *testing.T) {
	subcommands := []string{"add", "update", "complete", "block", "list", "reset-timedout"}

	for _, sub := range subcommands {
		found := false
		for _, c := range rootCmd.Commands() {
			if c.Name() == sub {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand '%s' to exist", sub)
		}
	}
}

func TestExecute(t *testing.T) {
	// Test that Execute doesn't panic with valid args
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute panicked: %v", r)
		}
	}()

	// Note: This will call os.Exit in normal execution
	// We'll just test that it doesn't panic with --help
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--help"})
	Execute()
}

func TestAddCommand(t *testing.T) {
	// Test that add command exists and has correct configuration
	addCmd := rootCmd.Commands()[0] // Based on insertion order
	if addCmd.Name() != "add" {
		t.Errorf("Expected first command to be 'add', got '%s'", addCmd.Name())
	}

	// Test Args validator
	if addCmd.Args == nil {
		t.Error("Expected add command to have Args validator")
	}
}

func TestListCommand(t *testing.T) {
	var listCmd *cobra.Command
	for _, c := range rootCmd.Commands() {
		if c.Name() == "list" {
			listCmd = c
			break
		}
	}

	if listCmd == nil {
		t.Fatal("Expected 'list' command to exist")
	}

	// List should accept no args
	if listCmd.Args == nil {
		t.Error("Expected list command to have Args validator")
	}
}

func TestCompleteCommand(t *testing.T) {
	var completeCmd *cobra.Command
	for _, c := range rootCmd.Commands() {
		if c.Name() == "complete" {
			completeCmd = c
			break
		}
	}

	if completeCmd == nil {
		t.Fatal("Expected 'complete' command to exist")
	}
}

func TestBlockCommand(t *testing.T) {
	var blockCmd *cobra.Command
	for _, c := range rootCmd.Commands() {
		if c.Name() == "block" {
			blockCmd = c
			break
		}
	}

	if blockCmd == nil {
		t.Fatal("Expected 'block' command to exist")
	}
}

func TestResetCommand(t *testing.T) {
	var resetCmd *cobra.Command
	for _, c := range rootCmd.Commands() {
		if c.Name() == "reset-timedout" {
			resetCmd = c
			break
		}
	}

	if resetCmd == nil {
		t.Fatal("Expected 'reset-timedout' command to exist")
	}
}

func TestUpdateCommand(t *testing.T) {
	var updateCmd *cobra.Command
	for _, c := range rootCmd.Commands() {
		if c.Name() == "update" {
			updateCmd = c
			break
		}
	}

	if updateCmd == nil {
		t.Fatal("Expected 'update' command to exist")
	}
}

func TestRenderAnchorErrorNotFoundRootChain(t *testing.T) {
	// Chain length 1 with an absolute start dir: the walk started at the
	// filesystem root, so the launcher-cwd note must appear, before the
	// remedy lines.
	err := &anchor.AnchorError{Chain: []string{"/"}, Reason: anchor.ReasonNotFound}
	out := renderAnchorError(err)

	note := "note: the search started at the filesystem root; the process that started taskflow ran with cwd=/ and no anchor can exist above it; fix the launcher's working directory or set TASKFLOW_DIR\n"
	if !strings.Contains(out, note) {
		t.Errorf("renderAnchorError() output missing root-start note:\n%s", out)
	}
	if !strings.Contains(out, "filesystem root reached") {
		t.Errorf("renderAnchorError() output missing 'filesystem root reached':\n%s", out)
	}
	if !strings.Contains(out, "remedy: run `taskflow init` in the project root") {
		t.Errorf("renderAnchorError() output missing init remedy line:\n%s", out)
	}
	if !strings.Contains(out, "or set TASKFLOW_DIR to override") {
		t.Errorf("renderAnchorError() output missing TASKFLOW_DIR remedy line:\n%s", out)
	}
	if noteIdx := strings.Index(out, "note:"); noteIdx == -1 || noteIdx > strings.Index(out, "remedy:") {
		t.Errorf("renderAnchorError() note must come before the remedy lines:\n%s", out)
	}
}

func TestRenderAnchorErrorNotFoundNonAbsSingleEntry(t *testing.T) {
	// Chain length 1 with a non-absolute start dir ("." fallback when
	// os.Getwd fails): the note must NOT appear; the start dir is not a
	// filesystem root.
	err := &anchor.AnchorError{Chain: []string{"."}, Reason: anchor.ReasonNotFound}
	out := renderAnchorError(err)

	if strings.Contains(out, "note:") {
		t.Errorf("renderAnchorError() output must not contain 'note:' for a non-absolute single-entry chain:\n%s", out)
	}
}

func TestRenderAnchorErrorNotFoundMultiEntryChain(t *testing.T) {
	// Multi-entry chain: the walk started below the root, so no note must
	// appear.
	err := &anchor.AnchorError{
		Chain:  []string{"/home/u/proj", "/home/u", "/"},
		Reason: anchor.ReasonNotFound,
	}
	out := renderAnchorError(err)

	if strings.Contains(out, "note:") {
		t.Errorf("renderAnchorError() output must not contain 'note:' for a multi-entry chain:\n%s", out)
	}
	if !strings.Contains(out, "filesystem root reached") {
		t.Errorf("renderAnchorError() output missing 'filesystem root reached':\n%s", out)
	}
	if !strings.Contains(out, "remedy: run `taskflow init` in the project root") {
		t.Errorf("renderAnchorError() output missing init remedy line:\n%s", out)
	}
	if !strings.Contains(out, "or set TASKFLOW_DIR to override") {
		t.Errorf("renderAnchorError() output missing TASKFLOW_DIR remedy line:\n%s", out)
	}
}
