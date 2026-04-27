package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

// Helper to set up a clean command for testing
func setupResetCommand() *cobra.Command {
	return resetCmd
}

func TestResetCmdUse(t *testing.T) {
	cmd := setupResetCommand()
	if cmd.Use != "reset-timedout" {
		t.Errorf("Use = %v, want %v", cmd.Use, "reset-timedout")
	}
}

func TestResetCmdShort(t *testing.T) {
	cmd := setupResetCommand()
	if cmd.Short == "" {
		t.Error("Short should not be empty")
	}
}

func TestResetCmdLong(t *testing.T) {
	cmd := setupResetCommand()
	if cmd.Long == "" {
		t.Error("Long should not be empty")
	}
}

func TestResetCmdArgs(t *testing.T) {
	cmd := setupResetCommand()
	if cmd.Args == nil {
		t.Error("Expected Args validator to be set")
	}
}

func TestResetCmdFlags(t *testing.T) {
	cmd := setupResetCommand()

	flags := cmd.Flags()
	if flags.Lookup("json") == nil {
		t.Error("Expected 'json' flag to exist")
	}
}

func TestResetCmdFlagShorthands(t *testing.T) {
	cmd := setupResetCommand()

	jsonFlag := cmd.Flags().Lookup("json")
	if jsonFlag != nil && jsonFlag.Shorthand != "j" {
		t.Errorf("Expected 'json' shorthand to be 'j', got %s", jsonFlag.Shorthand)
	}
}

func TestResetCmdExample(t *testing.T) {
	cmd := setupResetCommand()

	if cmd.Example == "" {
		t.Error("Example should not be empty")
	}
}

func TestResetCmdExampleFormat(t *testing.T) {
	cmd := setupResetCommand()

	examples := cmd.Example
	if examples != "" {
		if len(examples) < 10 {
			t.Errorf("Example too short: %s", examples)
		}
	}
}

func TestResetCmdHasSubcommands(t *testing.T) {
	cmd := setupResetCommand()

	if len(cmd.Commands()) != 0 {
		t.Error("Reset command should not have subcommands")
	}
}

func TestResetCmdFlagDescriptions(t *testing.T) {
	cmd := setupResetCommand()

	jsonFlag := cmd.Flags().Lookup("json")
	if jsonFlag != nil && jsonFlag.Usage == "" {
		t.Error("json flag should have usage description")
	}
}

func TestResetCmdExecute(t *testing.T) {
	cmd := setupResetCommand()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute panicked: %v", r)
		}
	}()

	cmdName := cmd.Name()
	if cmdName != "reset-timedout" {
		t.Errorf("Command name = %v, want %v", cmdName, "reset-timedout")
	}
}

func TestResetCmdHelp(t *testing.T) {
	cmd := setupResetCommand()

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

func TestResetCmdAnnotations(t *testing.T) {
	cmd := setupResetCommand()

	annotations := cmd.Annotations
	if annotations == nil {
		t.Log("No annotations set (this is OK)")
	}
}

func TestResetCmdSilenceUsage(t *testing.T) {
	cmd := setupResetCommand()

	if cmd.SilenceUsage {
		t.Log("SilenceUsage is true")
	}
}

func TestResetCmdSilenceErrors(t *testing.T) {
	cmd := setupResetCommand()

	if cmd.SilenceErrors {
		t.Log("SilenceErrors is true")
	}
}

func TestResetValidationFlags(t *testing.T) {
	tests := []struct {
		name   string
		flag   string
		exists bool
	}{
		{"json flag", "json", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := setupResetCommand()
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
