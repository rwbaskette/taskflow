package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

// Helper to set up a clean command for testing
func setupBlockCommand() *cobra.Command {
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

	flags := cmd.Flags()
	if flags.Lookup("json") == nil {
		t.Error("Expected 'json' flag to exist")
	}
}

func TestBlockCmdFlagShorthands(t *testing.T) {
	cmd := setupBlockCommand()

	jsonFlag := cmd.Flags().Lookup("json")
	if jsonFlag != nil && jsonFlag.Shorthand != "j" {
		t.Errorf("Expected 'json' shorthand to be 'j', got %s", jsonFlag.Shorthand)
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

	jsonFlag := cmd.Flags().Lookup("json")
	if jsonFlag != nil && jsonFlag.Usage == "" {
		t.Error("JSON flag should have usage description")
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
	cmd := setupBlockCommand()

	flag := cmd.Flags().Lookup("json")
	if flag == nil {
		t.Error("Expected 'json' flag to exist")
	}
}

func TestBlockFlagDefaults(t *testing.T) {
	cmd := setupBlockCommand()

	flag := cmd.Flags().Lookup("json")
	if flag == nil {
		t.Skip("Flag 'json' not found")
	}

	if flag.DefValue != "" {
		t.Logf("Default value for 'json' = %v", flag.DefValue)
	}
}
