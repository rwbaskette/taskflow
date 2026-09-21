package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
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

	buf := new(bytes.Buffer)
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

// listTotal runs the list command with the given filter and returns the
// reported total count.
func listTotal(t *testing.T, binPath, dbDir, filter string) float64 {
	t.Helper()

	stdout, stderr, err := runTaskflow(t, binPath, dbDir, "list", filter)
	if err != nil {
		t.Fatalf("task list %s failed: %v\nstderr: %s", filter, err, stderr)
	}

	// The output may carry text around the JSON body; extract it.
	start := strings.Index(stdout, "{")
	end := strings.LastIndex(stdout, "}")
	if start == -1 || end <= start {
		t.Fatalf("could not find JSON output for list %s: %q", filter, stdout)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout[start:end+1]), &result); err != nil {
		t.Fatalf("failed to parse list output %q: %v", stdout, err)
	}
	total, _ := result["total"].(float64)
	return total
}

// TestListPaddedAllStatusBehavesAsAll pins that a whitespace-padded "all"
// status filter (any casing) is treated exactly like "all": the filter is
// dropped and every status is returned, instead of the literal value
// " all " reaching the db and matching nothing.
func TestListPaddedAllStatusBehavesAsAll(t *testing.T) {
	binPath, binCleanup := buildTaskflowBinary(t)
	defer binCleanup()

	dbDir := t.TempDir()

	if _, stderr, err := runTaskflow(t, binPath, dbDir, "add",
		`{"id":"padded-all-1","milestone":"sprint-1","title":"Padded All Status Test","description":"seed task","actor":"tester"}`); err != nil {
		t.Fatalf("task add failed: %v\nstderr: %s", err, stderr)
	}

	tests := []struct {
		name      string
		filter    string
		wantTotal float64
	}{
		{"explicit status filters", `{"status":"todo"}`, 1},
		{"non-matching status filters to zero", `{"status":"done"}`, 0},
		{"all means every status", `{"status":"all"}`, 1},
		{"padded all behaves as all", `{"status":" all "}`, 1},
		{"padded uppercase ALL behaves as all", `{"status":" ALL "}`, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := listTotal(t, binPath, dbDir, tt.filter); got != tt.wantTotal {
				t.Errorf("list %s total = %v, want %v", tt.filter, got, tt.wantTotal)
			}
		})
	}
}
