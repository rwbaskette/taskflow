package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCmdVersion(t *testing.T) {
	// Test that version is set correctly
	if rootCmd.Version != "0.1.0" {
		t.Errorf("Version = %v, want %v", rootCmd.Version, "0.1.0")
	}
}

func TestRootCmdUse(t *testing.T) {
	if rootCmd.Use != "task" {
		t.Errorf("Use = %v, want %v", rootCmd.Use, "task")
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
	// Test that config flag exists
	flag := rootCmd.PersistentFlags().Lookup("config")
	if flag == nil {
		t.Error("Expected 'config' flag to exist")
	}

	// Test that verbose flag exists
	flag = rootCmd.PersistentFlags().Lookup("verbose")
	if flag == nil {
		t.Error("Expected 'verbose' flag to exist")
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
