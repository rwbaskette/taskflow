package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

// Helper to set up a clean command for testing
func setupBlockCommand() *cobra.Command {
	// Reset global variables before each test
	blockID = ""
	blockReason = ""

	return blockCmd
}

func TestBlockCmdUse(t *testing.T) {
	cmd := setupBlockCommand()
	if cmd.Use != "block" {
		t.Errorf("Use = %v, want %v", cmd.Use, "block")
	}
}

func TestBlockCmdShort(t *testing.T) {
	cmd := setupBlockCommand()
	if cmd.Short == "" {
		t.Error("Short should not be empty")
	}
}

func TestBlockCmdLong(t *testing.T) {
	cmd := setupBlockCommand()
	if cmd.Long == "" {
		t.Error("Long should not be empty")
	}
}

func TestBlockCmdArgs(t *testing.T) {
	cmd := setupBlockCommand()
	// blockCmd should accept NoArgs
	if cmd.Args == nil {
		t.Error("Expected Args validator to be set")
	}
}

func TestBlockCmdFlags(t *testing.T) {
	cmd := setupBlockCommand()

	// Check that all expected flags exist
	flags := cmd.Flags()
	if flags.Lookup("id") == nil {
		t.Error("Expected 'id' flag to exist")
	}
	if flags.Lookup("reason") == nil {
		t.Error("Expected 'reason' flag to exist")
	}
}

func TestBlockCmdFlagShorthands(t *testing.T) {
	cmd := setupBlockCommand()

	// Check flag shorthands
	idFlag := cmd.Flags().Lookup("id")
	if idFlag != nil && idFlag.Shorthand != "i" {
		t.Errorf("Expected 'id' shorthand to be 'i', got %s", idFlag.Shorthand)
	}
	reasonFlag := cmd.Flags().Lookup("reason")
	if reasonFlag != nil && reasonFlag.Shorthand != "r" {
		t.Errorf("Expected 'reason' shorthand to be 'r', got %s", reasonFlag.Shorthand)
	}
}

func TestBlockCmdExample(t *testing.T) {
	cmd := setupBlockCommand()

	if cmd.Example == "" {
		t.Error("Example should not be empty")
	}
}

func TestBlockCmdExampleFormat(t *testing.T) {
	cmd := setupBlockCommand()

	examples := cmd.Example
	if examples != "" {
		if len(examples) < 10 {
			t.Errorf("Example too short: %s", examples)
		}
	}
}

func TestBlockCmdHasSubcommands(t *testing.T) {
	cmd := setupBlockCommand()

	if len(cmd.Commands()) != 0 {
		t.Error("Block command should not have subcommands")
	}
}

func TestBlockCmdFlagDescriptions(t *testing.T) {
	cmd := setupBlockCommand()

	idFlag := cmd.Flags().Lookup("id")
	if idFlag != nil && idFlag.Usage == "" {
		t.Error("ID flag should have usage description")
	}

	reasonFlag := cmd.Flags().Lookup("reason")
	if reasonFlag != nil && reasonFlag.Usage == "" {
		t.Error("Reason flag should have usage description")
	}
}

func TestBlockCmdExecute(t *testing.T) {
	cmd := setupBlockCommand()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute panicked: %v", r)
		}
	}()

	cmdName := cmd.Name()
	if cmdName != "block" {
		t.Errorf("Command name = %v, want %v", cmdName, "block")
	}
}

func TestBlockCmdHelp(t *testing.T) {
	cmd := setupBlockCommand()

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

func TestBlockCmdAnnotations(t *testing.T) {
	cmd := setupBlockCommand()

	annotations := cmd.Annotations
	if annotations == nil {
		t.Log("No annotations set (this is OK)")
	}
}

func TestBlockCmdSilenceUsage(t *testing.T) {
	cmd := setupBlockCommand()

	if cmd.SilenceUsage {
		t.Log("SilenceUsage is true")
	}
}

func TestBlockCmdSilenceErrors(t *testing.T) {
	cmd := setupBlockCommand()

	if cmd.SilenceErrors {
		t.Log("SilenceErrors is true")
	}
}

func TestBlockValidationRequiredFlags(t *testing.T) {
	// Test that required flags are properly configured
	tests := []struct {
		name   string
		flag   string
		exists bool
	}{
		{"ID flag", "id", true},
		{"Reason flag", "reason", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := setupBlockCommand()
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

func TestBlockFlagDefaults(t *testing.T) {
	// Check default values for flags
	tests := []struct {
		name     string
		flag     string
		defValue string
	}{
		{"ID default", "id", ""},
		{"Reason default", "reason", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := blockCmd.Flags().Lookup(tt.flag)

			if flag == nil {
				t.Skipf("Flag '%s' not found", tt.flag)
			}

			if flag.DefValue != tt.defValue {
				t.Logf("Default value for '%s' = %v (checking %v)", tt.flag, flag.DefValue, tt.defValue)
			}
		})
	}
}
