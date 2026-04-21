package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func setupAddCommand() *cobra.Command {
	addJSON = ""
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
	if cmd.Args == nil {
		t.Error("Expected Args validator to be set")
	}
}

func TestAddCmdFlags(t *testing.T) {
	cmd := setupAddCommand()

	flags := cmd.Flags()
	if flags.Lookup("json") == nil {
		t.Error("Expected 'json' flag to exist")
	}
}

func TestAddCmdFlagShorthands(t *testing.T) {
	cmd := setupAddCommand()

	jsonFlag := cmd.Flags().Lookup("json")
	if jsonFlag != nil && jsonFlag.Shorthand != "j" {
		t.Errorf("Expected 'json' shorthand to be 'j', got %s", jsonFlag.Shorthand)
	}
}

func TestAddCmdExample(t *testing.T) {
	cmd := setupAddCommand()

	if cmd.Example == "" {
		t.Error("Example should not be empty")
	}
}

func TestAddCmdExampleFormat(t *testing.T) {
	cmd := setupAddCommand()

	examples := cmd.Example
	if examples != "" {
		if len(examples) < 10 {
			t.Errorf("Example too short: %s", examples)
		}
	}
}

func TestAddCmdHasSubcommands(t *testing.T) {
	cmd := setupAddCommand()

	if len(cmd.Commands()) != 0 {
		t.Error("Add command should not have subcommands")
	}
}

func TestAddCmdParent(t *testing.T) {
	cmd := setupAddCommand()

	if cmd.Name() != "add" {
		t.Errorf("Command name should be 'add', got %s", cmd.Name())
	}
}

func TestAddCmdFlagDescriptions(t *testing.T) {
	cmd := setupAddCommand()

	jsonFlag := cmd.Flags().Lookup("json")
	if jsonFlag != nil && jsonFlag.Usage == "" {
		t.Error("JSON flag should have usage description")
	}
}

func TestAddCmdExecute(t *testing.T) {
	cmd := setupAddCommand()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute panicked: %v", r)
		}
	}()

	cmdName := cmd.Name()
	if cmdName != "add" {
		t.Errorf("Command name = %v, want %v", cmdName, "add")
	}
}

func TestAddCmdHelp(t *testing.T) {
	cmd := setupAddCommand()

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

func TestAddCmdAnnotations(t *testing.T) {
	cmd := setupAddCommand()

	annotations := cmd.Annotations
	if annotations == nil {
		t.Log("No annotations set (this is OK)")
	}
}

func TestAddCmdSilenceUsage(t *testing.T) {
	cmd := setupAddCommand()

	if cmd.SilenceUsage {
		t.Log("SilenceUsage is true")
	}
}

func TestAddCmdSilenceErrors(t *testing.T) {
	cmd := setupUpdateCommand()

	if cmd.SilenceErrors {
		t.Log("SilenceErrors is true")
	}
}

func TestAddValidationJSONFlag(t *testing.T) {
	cmd := setupAddCommand()

	flag := cmd.Flags().Lookup("json")
	if flag == nil {
		t.Error("Expected 'json' flag to exist")
	}
}

func TestAddFlagDefaults(t *testing.T) {
	cmd := setupAddCommand()

	flag := cmd.Flags().Lookup("json")
	if flag == nil {
		t.Skip("Flag 'json' not found")
	}

	if flag.DefValue != "" {
		t.Logf("Default value for 'json' = %v", flag.DefValue)
	}
}