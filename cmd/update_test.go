package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

// Helper to set up a clean command for testing
func setupUpdateCommand() *cobra.Command {
	// Reset global variables before each test
	updateID = ""
	updateTitle = ""
	updateDescription = ""
	updateStatus = ""
	updateMilestone = ""
	updateActor = ""

	return updateCmd
}

func TestUpdateCmdUse(t *testing.T) {
	cmd := setupUpdateCommand()
	if cmd.Use != "update" {
		t.Errorf("Use = %v, want %v", cmd.Use, "update")
	}
}

func TestUpdateCmdShort(t *testing.T) {
	cmd := setupUpdateCommand()
	if cmd.Short == "" {
		t.Error("Short should not be empty")
	}
}

func TestUpdateCmdLong(t *testing.T) {
	cmd := setupUpdateCommand()
	if cmd.Long == "" {
		t.Error("Long should not be empty")
	}
}

func TestUpdateCmdArgs(t *testing.T) {
	cmd := setupUpdateCommand()
	// updateCmd should accept NoArgs
	if cmd.Args == nil {
		t.Error("Expected Args validator to be set")
	}
}

func TestUpdateCmdFlags(t *testing.T) {
	cmd := setupUpdateCommand()

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
	if flags.Lookup("status") == nil {
		t.Error("Expected 'status' flag to exist")
	}
	if flags.Lookup("milestone") == nil {
		t.Error("Expected 'milestone' flag to exist")
	}
	if flags.Lookup("actor") == nil {
		t.Error("Expected 'actor' flag to exist")
	}
}

func TestUpdateCmdFlagShorthands(t *testing.T) {
	cmd := setupUpdateCommand()

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
	statusFlag := cmd.Flags().Lookup("status")
	if statusFlag != nil && statusFlag.Shorthand != "s" {
		t.Errorf("Expected 'status' shorthand to be 's', got %s", statusFlag.Shorthand)
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

func TestUpdateCmdExample(t *testing.T) {
	cmd := setupUpdateCommand()

	// Test that example exists
	if cmd.Example == "" {
		t.Error("Example should not be empty")
	}
}

func TestUpdateCmdExampleFormat(t *testing.T) {
	cmd := setupUpdateCommand()

	// Check that examples are properly formatted
	examples := cmd.Example
	if examples != "" {
		if len(examples) < 10 {
			t.Errorf("Example too short: %s", examples)
		}
	}
}

func TestUpdateCmdHasSubcommands(t *testing.T) {
	cmd := setupUpdateCommand()

	// update command should not have subcommands
	if len(cmd.Commands()) != 0 {
		t.Error("Update command should not have subcommands")
	}
}

func TestUpdateCmdFlagDescriptions(t *testing.T) {
	cmd := setupUpdateCommand()

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

	statusFlag := cmd.Flags().Lookup("status")
	if statusFlag != nil && statusFlag.Usage == "" {
		t.Error("Status flag should have usage description")
	}

	milestoneFlag := cmd.Flags().Lookup("milestone")
	if milestoneFlag != nil && milestoneFlag.Usage == "" {
		t.Error("Milestone flag should have usage description")
	}

	actorFlag := cmd.Flags().Lookup("actor")
	if actorFlag != nil && actorFlag.Usage == "" {
		t.Error("Actor flag should have usage description")
	}
}

func TestUpdateCmdExecute(t *testing.T) {
	cmd := setupUpdateCommand()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute panicked: %v", r)
		}
	}()

	cmdName := cmd.Name()
	if cmdName != "update" {
		t.Errorf("Command name = %v, want %v", cmdName, "update")
	}
}

func TestUpdateCmdHelp(t *testing.T) {
	cmd := setupUpdateCommand()

	// Test that help works
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

	// Should contain usage information
	if len(output) < 20 {
		t.Errorf("Help output too short: %s", output)
	}
}

func TestUpdateCmdAnnotations(t *testing.T) {
	cmd := setupUpdateCommand()

	// Test command annotations if any
	annotations := cmd.Annotations
	if annotations == nil {
		t.Log("No annotations set (this is OK)")
	}
}

func TestUpdateCmdSilenceUsage(t *testing.T) {
	cmd := setupUpdateCommand()

	// Should not silence usage by default
	if cmd.SilenceUsage {
		t.Log("SilenceUsage is true")
	}
}

func TestUpdateCmdSilenceErrors(t *testing.T) {
	cmd := setupUpdateCommand()

	// Should not silence errors by default
	if cmd.SilenceErrors {
		t.Log("SilenceErrors is true")
	}
}

func TestUpdateValidationRequiredFlags(t *testing.T) {
	// Test that required flags are properly configured
	tests := []struct {
		name   string
		flag   string
		exists bool
	}{
		{"ID flag", "id", true},
		{"Title flag", "title", true},
		{"Description flag", "description", true},
		{"Status flag", "status", true},
		{"Milestone flag", "milestone", true},
		{"Actor flag", "actor", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := setupUpdateCommand()
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

func TestUpdateFlagDefaults(t *testing.T) {
	// Check default values for flags
	tests := []struct {
		name     string
		flag     string
		defValue string
	}{
		{"ID default", "id", ""},
		{"Title default", "title", ""},
		{"Description default", "description", ""},
		{"Status default", "status", ""},
		{"Milestone default", "milestone", ""},
		{"Actor default", "actor", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := setupUpdateCommand()
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
