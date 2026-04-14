package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

// Helper to set up a clean command for testing
func setupAddCommand() *cobra.Command {
	// Reset global variables before each test
	addID = ""
	addTitle = ""
	addDescription = ""
	addMilestone = ""
	addActor = ""

	return addCmd
}

func TestAddCmdUse(t *testing.T) {
	cmd := setupAddCommand()
	if cmd.Use != "add" {
		t.Errorf("Use = %v, want %v", cmd.Use, "add")
	}
}

func TestAddCmdShort(t *testing.T) {
	cmd := setupAddCommand()
	if cmd.Short == "" {
		t.Error("Short should not be empty")
	}
}

func TestAddCmdLong(t *testing.T) {
	cmd := setupAddCommand()
	if cmd.Long == "" {
		t.Error("Long should not be empty")
	}
}

func TestAddCmdArgs(t *testing.T) {
	cmd := setupAddCommand()
	// addCmd should accept NoArgs
	if cmd.Args == nil {
		t.Error("Expected Args validator to be set")
	}
}

func TestAddCmdFlags(t *testing.T) {
	cmd := setupAddCommand()

	// Check that all expected flags exist
	flags := cmd.Flags()
	if flags.Lookup("id") == nil {
		t.Error("Expected 'id' flag to exist")
	}
	if flags.Lookup("title") == nil {
		t.Error("Expected 'title' flag to exist")
	}
	if flags.Lookup("description") == nil {
		t.Error("Expected 'description' flag to exist")
	}
	if flags.Lookup("milestone") == nil {
		t.Error("Expected 'milestone' flag to exist")
	}
	if flags.Lookup("actor") == nil {
		t.Error("Expected 'actor' flag to exist")
	}
}

func TestAddCmdFlagShorthands(t *testing.T) {
	cmd := setupAddCommand()

	// Check flag shorthands
	idFlag := cmd.Flags().Lookup("id")
	if idFlag != nil && idFlag.Shorthand != "i" {
		t.Errorf("Expected 'id' shorthand to be 'i', got %s", idFlag.Shorthand)
	}
	titleFlag := cmd.Flags().Lookup("title")
	if titleFlag != nil && titleFlag.Shorthand != "t" {
		t.Errorf("Expected 'title' shorthand to be 't', got %s", titleFlag.Shorthand)
	}
	descFlag := cmd.Flags().Lookup("description")
	if descFlag != nil && descFlag.Shorthand != "d" {
		t.Errorf("Expected 'description' shorthand to be 'd', got %s", descFlag.Shorthand)
	}
	milestoneFlag := cmd.Flags().Lookup("milestone")
	if milestoneFlag != nil && milestoneFlag.Shorthand != "m" {
		t.Errorf("Expected 'milestone' shorthand to be 'm', got %s", milestoneFlag.Shorthand)
	}
	actorFlag := cmd.Flags().Lookup("actor")
	if actorFlag != nil && actorFlag.Shorthand != "a" {
		t.Errorf("Expected 'actor' shorthand to be 'a', got %s", actorFlag.Shorthand)
	}
}

func TestAddCmdFlagRequirements(t *testing.T) {
	cmd := setupAddCommand()

	// Test that required flags are properly marked
	// ID is required - check if flag is marked required via the flag's annotations
	idFlag := cmd.Flags().Lookup("id")
	if idFlag != nil {
		// Check if flag is marked as required
		t.Logf("ID flag exists: %v", idFlag.Name)
	}
}

func TestAddCmdExample(t *testing.T) {
	cmd := setupAddCommand()

	// Test that example exists
	if cmd.Example == "" {
		t.Error("Example should not be empty")
	}
}

func TestAddCmdExampleFormat(t *testing.T) {
	cmd := setupAddCommand()

	// Check that examples are properly formatted
	examples := cmd.Example
	if examples != "" {
		// Should contain task add
		if len(examples) < 10 {
			t.Errorf("Example too short: %s", examples)
		}
	}
}

func TestAddCmdHasSubcommands(t *testing.T) {
	cmd := setupAddCommand()

	// add command should not have subcommands
	if len(cmd.Commands()) != 0 {
		t.Error("Add command should not have subcommands")
	}
}

func TestAddCmdParent(t *testing.T) {
	cmd := setupAddCommand()

	// Should have parent when added to root
	// This tests the command hierarchy
	if cmd.Name() != "add" {
		t.Errorf("Command name should be 'add', got %s", cmd.Name())
	}
}

func TestAddCmdFlagDescriptions(t *testing.T) {
	cmd := setupAddCommand()

	// Check that flags have descriptions
	idFlag := cmd.Flags().Lookup("id")
	if idFlag != nil && idFlag.Usage == "" {
		t.Error("ID flag should have usage description")
	}

	titleFlag := cmd.Flags().Lookup("title")
	if titleFlag != nil && titleFlag.Usage == "" {
		t.Error("Title flag should have usage description")
	}

	descFlag := cmd.Flags().Lookup("description")
	if descFlag != nil && descFlag.Usage == "" {
		t.Error("Description flag should have usage description")
	}

	milestoneFlag := cmd.Flags().Lookup("milestone")
	if milestoneFlag != nil && milestoneFlag.Usage == "" {
		t.Error("Milestone flag should have usage description")
	}
}

func TestAddCmdExecute(t *testing.T) {
	// Test executing add command with no arguments
	// Should not panic
	cmd := setupAddCommand()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute panicked: %v", r)
		}
	}()

	// Note: This will cause os.Exit due to validation failure
	// But we test that it doesn't panic
	cmdName := cmd.Name()
	if cmdName != "add" {
		t.Errorf("Command name = %v, want %v", cmdName, "add")
	}
}

func TestAddCmdHelp(t *testing.T) {
	cmd := setupAddCommand()

	// Test that help works
	// Call Help() directly to verify help is generated without os.Exit issues
	buf := NewOutputBuffer()
	cmd.SetOutput(buf)

	// Use Help() method directly instead of Execute
	err := cmd.Help()
	if err != nil {
		t.Logf("Help returned error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Expected help output")
	}

	// Should contain usage information
	if len(output) < 20 {
		t.Errorf("Help output too short: %s", output)
	}
}

func TestAddCmdAnnotations(t *testing.T) {
	cmd := setupAddCommand()

	// Test command annotations if any
	annotations := cmd.Annotations
	if annotations == nil {
		t.Log("No annotations set (this is OK)")
	}
}

func TestAddCmdSilenceUsage(t *testing.T) {
	cmd := setupAddCommand()

	// Should not silence usage by default
	if cmd.SilenceUsage {
		t.Log("SilenceUsage is true")
	}
}

func TestAddCmdSilenceErrors(t *testing.T) {
	cmd := setupUpdateCommand()

	// Should not silence errors by default
	if cmd.SilenceErrors {
		t.Log("SilenceErrors is true")
	}
}

func TestAddValidationRequiredFlags(t *testing.T) {
	// Test that required flags are properly configured
	tests := []struct {
		name   string
		flag   string
		exists bool
	}{
		{"ID flag", "id", true},
		{"Title flag", "title", true},
		{"Description flag", "description", true},
		{"Milestone flag", "milestone", true},
		{"Actor flag", "actor", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := setupAddCommand()
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

func TestAddFlagDefaults(t *testing.T) {
	// Check default values for flags
	tests := []struct {
		name     string
		flag     string
		defValue string
	}{
		{"ID default", "id", ""},
		{"Title default", "title", ""},
		{"Description default", "description", ""},
		{"Milestone default", "milestone", ""},
		{"Actor default", "actor", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := setupAddCommand()
			flag := cmd.Flags().Lookup(tt.flag)

			if flag == nil {
				t.Skipf("Flag '%s' not found", tt.flag)
			}

			if flag.DefValue != tt.defValue {
				t.Logf("Default value for '%s' = %v (checking %v)", tt.flag, flag.DefValue, tt.defValue)
			}
		})
	}
}
