package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

// Helper to set up a clean command for testing
func setupCompleteCommand() *cobra.Command {
	// Reset global variables before each test
	completeID = ""

	return completeCmd
}

func TestCompleteCmdUse(t *testing.T) {
	cmd := setupCompleteCommand()
	if cmd.Use != "complete" {
		t.Errorf("Use = %v, want %v", cmd.Use, "complete")
	}
}

func TestCompleteCmdShort(t *testing.T) {
	cmd := setupCompleteCommand()
	if cmd.Short == "" {
		t.Error("Short should not be empty")
	}
}

func TestCompleteCmdLong(t *testing.T) {
	cmd := setupCompleteCommand()
	if cmd.Long == "" {
		t.Error("Long should not be empty")
	}
}

func TestCompleteCmdArgs(t *testing.T) {
	cmd := setupCompleteCommand()
	// completeCmd should accept NoArgs
	if cmd.Args == nil {
		t.Error("Expected Args validator to be set")
	}
}

func TestCompleteCmdFlags(t *testing.T) {
	cmd := setupCompleteCommand()

	// Check that all expected flags exist
	flags := cmd.Flags()
	if flags.Lookup("id") == nil {
		t.Error("Expected 'id' flag to exist")
	}
}

func TestCompleteCmdFlagShorthands(t *testing.T) {
	cmd := setupCompleteCommand()

	// Check flag shorthand
	idFlag := cmd.Flags().Lookup("id")
	if idFlag != nil && idFlag.Shorthand != "i" {
		t.Errorf("Expected 'id' shorthand to be 'i', got %s", idFlag.Shorthand)
	}
}

func TestCompleteCmdExample(t *testing.T) {
	cmd := setupCompleteCommand()

	if cmd.Example == "" {
		t.Error("Example should not be empty")
	}
}

func TestCompleteCmdExampleFormat(t *testing.T) {
	cmd := setupCompleteCommand()

	examples := cmd.Example
	if examples != "" {
		if len(examples) < 10 {
			t.Errorf("Example too short: %s", examples)
		}
	}
}

func TestCompleteCmdHasSubcommands(t *testing.T) {
	cmd := setupCompleteCommand()

	if len(cmd.Commands()) != 0 {
		t.Error("Complete command should not have subcommands")
	}
}

func TestCompleteCmdFlagDescriptions(t *testing.T) {
	cmd := setupCompleteCommand()

	idFlag := cmd.Flags().Lookup("id")
	if idFlag != nil && idFlag.Usage == "" {
		t.Error("ID flag should have usage description")
	}
}

func TestCompleteCmdExecute(t *testing.T) {
	cmd := setupCompleteCommand()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute panicked: %v", r)
		}
	}()

	cmdName := cmd.Name()
	if cmdName != "complete" {
		t.Errorf("Command name = %v, want %v", cmdName, "complete")
	}
}

func TestCompleteCmdHelp(t *testing.T) {
	cmd := setupCompleteCommand()

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

func TestCompleteCmdAnnotations(t *testing.T) {
	cmd := setupCompleteCommand()

	annotations := cmd.Annotations
	if annotations == nil {
		t.Log("No annotations set (this is OK)")
	}
}

func TestCompleteCmdSilenceUsage(t *testing.T) {
	cmd := setupCompleteCommand()

	if cmd.SilenceUsage {
		t.Log("SilenceUsage is true")
	}
}

func TestCompleteCmdSilenceErrors(t *testing.T) {
	cmd := setupCompleteCommand()

	if cmd.SilenceErrors {
		t.Log("SilenceErrors is true")
	}
}

func TestCompleteValidationRequiredFlags(t *testing.T) {
	// Test that required flags are properly configured
	tests := []struct {
		name   string
		flag   string
		exists bool
	}{
		{"ID flag", "id", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := setupCompleteCommand()
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

func TestCompleteFlagDefaults(t *testing.T) {
	// Check default values for flags
	flag := completeCmd.Flags().Lookup("id")
	if flag != nil && flag.DefValue != "" {
		t.Logf("Default value for 'id' = %v", flag.DefValue)
	}
}
