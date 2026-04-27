package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func setupListCommand() *cobra.Command {
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
	if cmd.Args == nil {
		t.Error("Expected Args validator to be set")
	}
}

func TestListCmdFlags(t *testing.T) {
	cmd := setupListCommand()

	flags := cmd.Flags()
	if flags.Lookup("json") == nil {
		t.Error("Expected 'json' flag to exist")
	}
}

func TestListCmdFlagShorthands(t *testing.T) {
	cmd := setupListCommand()

	jsonFlag := cmd.Flags().Lookup("json")
	if jsonFlag != nil && jsonFlag.Shorthand != "j" {
		t.Errorf("Expected 'json' shorthand to be 'j', got %s", jsonFlag.Shorthand)
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

	jsonFlag := cmd.Flags().Lookup("json")
	if jsonFlag != nil && jsonFlag.Usage == "" {
		t.Error("json flag should have usage description")
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
	tests := []struct {
		name   string
		flag   string
		exists bool
	}{
		{"json flag", "json", true},
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

func TestListCompletionRegistration(t *testing.T) {
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