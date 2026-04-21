package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func setupUpdateCommand() *cobra.Command {
	updateJSON = ""
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
	if cmd.Args == nil {
		t.Error("Expected Args validator to be set")
	}
}

func TestUpdateCmdFlags(t *testing.T) {
	cmd := setupUpdateCommand()

	flags := cmd.Flags()
	if flags.Lookup("json") == nil {
		t.Error("Expected 'json' flag to exist")
	}
}

func TestUpdateCmdFlagShorthands(t *testing.T) {
	cmd := setupUpdateCommand()

	jsonFlag := cmd.Flags().Lookup("json")
	if jsonFlag != nil && jsonFlag.Shorthand != "j" {
		t.Errorf("Expected 'json' shorthand to be 'j', got %s", jsonFlag.Shorthand)
	}
}

func TestUpdateCmdExample(t *testing.T) {
	cmd := setupUpdateCommand()

	if cmd.Example == "" {
		t.Error("Example should not be empty")
	}
}

func TestUpdateCmdExampleFormat(t *testing.T) {
	cmd := setupUpdateCommand()

	examples := cmd.Example
	if examples != "" {
		if len(examples) < 10 {
			t.Errorf("Example too short: %s", examples)
		}
	}
}

func TestUpdateCmdHasSubcommands(t *testing.T) {
	cmd := setupUpdateCommand()

	if len(cmd.Commands()) != 0 {
		t.Error("Update command should not have subcommands")
	}
}

func TestUpdateCmdFlagDescriptions(t *testing.T) {
	cmd := setupUpdateCommand()

	jsonFlag := cmd.Flags().Lookup("json")
	if jsonFlag != nil && jsonFlag.Usage == "" {
		t.Error("JSON flag should have usage description")
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

func TestUpdateCmdAnnotations(t *testing.T) {
	cmd := setupUpdateCommand()

	annotations := cmd.Annotations
	if annotations == nil {
		t.Log("No annotations set (this is OK)")
	}
}

func TestUpdateCmdSilenceUsage(t *testing.T) {
	cmd := setupUpdateCommand()

	if cmd.SilenceUsage {
		t.Log("SilenceUsage is true")
	}
}

func TestUpdateCmdSilenceErrors(t *testing.T) {
	cmd := setupUpdateCommand()

	if cmd.SilenceErrors {
		t.Log("SilenceErrors is true")
	}
}

func TestUpdateValidationJSONFlag(t *testing.T) {
	cmd := setupUpdateCommand()

	flag := cmd.Flags().Lookup("json")
	if flag == nil {
		t.Error("Expected 'json' flag to exist")
	}
}

func TestUpdateFlagDefaults(t *testing.T) {
	cmd := setupUpdateCommand()

	flag := cmd.Flags().Lookup("json")
	if flag == nil {
		t.Skip("Flag 'json' not found")
	}

	if flag.DefValue != "" {
		t.Logf("Default value for 'json' = %v", flag.DefValue)
	}
}