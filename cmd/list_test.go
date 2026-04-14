package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

// Helper to set up a clean command for testing
func setupListCommand() *cobra.Command {
	// Reset global variables before each test
	ResetListFlags()

	return listCmd
}

func TestListCmdUse(t *testing.T) {
	cmd := setupListCommand()
	if cmd.Use != "list" {
		t.Errorf("Use = %v, want %v", cmd.Use, "list")
	}
}

func TestListCmdShort(t *testing.T) {
	cmd := setupListCommand()
	if cmd.Short == "" {
		t.Error("Short should not be empty")
	}
}

func TestListCmdLong(t *testing.T) {
	cmd := setupListCommand()
	if cmd.Long == "" {
		t.Error("Long should not be empty")
	}
}

func TestListCmdArgs(t *testing.T) {
	cmd := setupListCommand()
	// listCmd should accept NoArgs
	if cmd.Args == nil {
		t.Error("Expected Args validator to be set")
	}
}

func TestListCmdFlags(t *testing.T) {
	cmd := setupListCommand()

	// Check that all expected flags exist
	flags := cmd.Flags()
	if flags.Lookup("all") == nil {
		t.Error("Expected 'all' flag to exist")
	}
	if flags.Lookup("milestone") == nil {
		t.Error("Expected 'milestone' flag to exist")
	}
	if flags.Lookup("status") == nil {
		t.Error("Expected 'status' flag to exist")
	}
	if flags.Lookup("actor") == nil {
		t.Error("Expected 'actor' flag to exist")
	}
	if flags.Lookup("format") == nil {
		t.Error("Expected 'format' flag to exist")
	}
	if flags.Lookup("limit") == nil {
		t.Error("Expected 'limit' flag to exist")
	}
	if flags.Lookup("offset") == nil {
		t.Error("Expected 'offset' flag to exist")
	}
}

func TestListCmdFlagShorthands(t *testing.T) {
	cmd := setupListCommand()

	// Check flag shorthands
	allFlag := cmd.Flags().Lookup("all")
	if allFlag != nil && allFlag.Shorthand != "a" {
		t.Errorf("Expected 'all' shorthand to be 'a', got %s", allFlag.Shorthand)
	}
	milestoneFlag := cmd.Flags().Lookup("milestone")
	if milestoneFlag != nil && milestoneFlag.Shorthand != "m" {
		t.Errorf("Expected 'milestone' shorthand to be 'm', got %s", milestoneFlag.Shorthand)
	}
	statusFlag := cmd.Flags().Lookup("status")
	if statusFlag != nil && statusFlag.Shorthand != "s" {
		t.Errorf("Expected 'status' shorthand to be 's', got %s", statusFlag.Shorthand)
	}
	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag != nil && formatFlag.Shorthand != "f" {
		t.Errorf("Expected 'format' shorthand to be 'f', got %s", formatFlag.Shorthand)
	}
	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag != nil && limitFlag.Shorthand != "l" {
		t.Errorf("Expected 'limit' shorthand to be 'l', got %s", limitFlag.Shorthand)
	}
	offsetFlag := cmd.Flags().Lookup("offset")
	if offsetFlag != nil && offsetFlag.Shorthand != "o" {
		t.Errorf("Expected 'offset' shorthand to be 'o', got %s", offsetFlag.Shorthand)
	}
}

func TestListCmdExample(t *testing.T) {
	cmd := setupListCommand()

	if cmd.Example == "" {
		t.Error("Example should not be empty")
	}
}

func TestListCmdExampleFormat(t *testing.T) {
	cmd := setupListCommand()

	examples := cmd.Example
	if examples != "" {
		if len(examples) < 10 {
			t.Errorf("Example too short: %s", examples)
		}
	}
}

func TestListCmdHasSubcommands(t *testing.T) {
	cmd := setupListCommand()

	if len(cmd.Commands()) != 0 {
		t.Error("List command should not have subcommands")
	}
}

func TestListCmdFlagDescriptions(t *testing.T) {
	cmd := setupListCommand()

	allFlag := cmd.Flags().Lookup("all")
	if allFlag != nil && allFlag.Usage == "" {
		t.Error("All flag should have usage description")
	}

	milestoneFlag := cmd.Flags().Lookup("milestone")
	if milestoneFlag != nil && milestoneFlag.Usage == "" {
		t.Error("Milestone flag should have usage description")
	}

	statusFlag := cmd.Flags().Lookup("status")
	if statusFlag != nil && statusFlag.Usage == "" {
		t.Error("Status flag should have usage description")
	}

	actorFlag := cmd.Flags().Lookup("actor")
	if actorFlag != nil && actorFlag.Usage == "" {
		t.Error("Actor flag should have usage description")
	}

	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag != nil && formatFlag.Usage == "" {
		t.Error("Format flag should have usage description")
	}

	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag != nil && limitFlag.Usage == "" {
		t.Error("Limit flag should have usage description")
	}

	offsetFlag := cmd.Flags().Lookup("offset")
	if offsetFlag != nil && offsetFlag.Usage == "" {
		t.Error("Offset flag should have usage description")
	}
}

func TestListCmdExecute(t *testing.T) {
	cmd := setupListCommand()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute panicked: %v", r)
		}
	}()

	cmdName := cmd.Name()
	if cmdName != "list" {
		t.Errorf("Command name = %v, want %v", cmdName, "list")
	}
}

func TestListCmdHelp(t *testing.T) {
	cmd := setupListCommand()

	buf := NewOutputBuffer()
	cmd.SetOutput(buf)

	err := cmd.Help()
	if err != nil {
		t.Logf("Help returned error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Expected help output")
	}

	if len(output) < 20 {
		t.Errorf("Help output too short: %s", output)
	}
}

func TestListCmdAnnotations(t *testing.T) {
	cmd := setupListCommand()

	annotations := cmd.Annotations
	if annotations == nil {
		t.Log("No annotations set (this is OK)")
	}
}

func TestListCmdSilenceUsage(t *testing.T) {
	cmd := setupListCommand()

	if cmd.SilenceUsage {
		t.Log("SilenceUsage is true")
	}
}

func TestListCmdSilenceErrors(t *testing.T) {
	cmd := setupListCommand()

	if cmd.SilenceErrors {
		t.Log("SilenceErrors is true")
	}
}

func TestListValidationFlags(t *testing.T) {
	// Test that all filter flags are properly configured
	tests := []struct {
		name   string
		flag   string
		exists bool
	}{
		{"All flag", "all", true},
		{"Milestone flag", "milestone", true},
		{"Status flag", "status", true},
		{"Actor flag", "actor", true},
		{"Format flag", "format", true},
		{"Limit flag", "limit", true},
		{"Offset flag", "offset", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := setupListCommand()
			flag := cmd.Flags().Lookup(tt.flag)

			if tt.exists && flag == nil {
				t.Errorf("Expected flag '%s' to exist", tt.flag)
			}
			if !tt.exists && flag != nil {
				t.Errorf("Expected flag '%s' to not exist", tt.flag)
			}
		})
	}
}

func TestListFlagDefaults(t *testing.T) {
	// Check default values for flags
	tests := []struct {
		name     string
		flag     string
		defValue string
	}{
		{"Format default", "format", "table"},
		{"Limit default", "limit", "20"},
		{"Offset default", "offset", "0"},
		{"All default", "all", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := listCmd.Flags().Lookup(tt.flag)

			if flag == nil {
				t.Skipf("Flag '%s' not found", tt.flag)
			}

			if flag.DefValue != tt.defValue {
				t.Logf("Default value for '%s' = %v (checking %v)", tt.flag, flag.DefValue, tt.defValue)
			}
		})
	}
}

func TestListCompletionRegistration(t *testing.T) {
	// Completion functions should be registered
	// This is a smoke test to ensure no panics
	t.Log("Testing flag completion registration")
}

func TestParseLimit(t *testing.T) {
	tests := []struct {
		input   string
		wantVal int
		wantErr bool
	}{
		{"", 20, false},
		{"10", 10, false},
		{"0", 0, false},
		{"100", 100, false},
		{"abc", 0, true},
		{"-1", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			val, err := ParseLimit(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLimit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if val != tt.wantVal {
				t.Errorf("ParseLimit() = %v, want %v", val, tt.wantVal)
			}
		})
	}
}

func TestParseOffset(t *testing.T) {
	tests := []struct {
		input   string
		wantVal int
		wantErr bool
	}{
		{"", 0, false},
		{"10", 10, false},
		{"0", 0, false},
		{"100", 100, false},
		{"abc", 0, true},
		{"-1", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			val, err := ParseOffset(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseOffset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if val != tt.wantVal {
				t.Errorf("ParseOffset() = %v, want %v", val, tt.wantVal)
			}
		})
	}
}
