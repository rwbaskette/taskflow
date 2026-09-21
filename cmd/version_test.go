package cmd

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

func TestVersionCmdExists(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "version" {
			return
		}
	}
	t.Error("Expected 'version' subcommand to exist on rootCmd")
}

func TestVersionCmdConfig(t *testing.T) {
	if versionCmd.Use != "version" {
		t.Errorf("Use = %v, want %v", versionCmd.Use, "version")
	}
	if versionCmd.Short != "Print the taskflow version" {
		t.Errorf("Short = %v, want %v", versionCmd.Short, "Print the taskflow version")
	}
	if versionCmd.Args == nil {
		t.Error("Expected Args validator to be set")
	}
}

func TestVersionCmdOutput(t *testing.T) {
	// Both `version` and `--version` must print the same unified output.
	run := func(args []string) string {
		// Other tests (e.g. TestExecute) run the shared rootCmd with
		// --help; pflag keeps flag values across parses on the same flag
		// set, so clear the leftover help flag to keep --version from
		// being treated as help.
		if h := rootCmd.Flags().Lookup("help"); h != nil {
			_ = h.Value.Set("false")
			h.Changed = false
		}

		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs(args)
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("Execute(%v) error: %v", args, err)
		}
		return strings.TrimSpace(buf.String())
	}

	versionOut := run([]string{"version"})
	flagOut := run([]string{"--version"})

	for _, output := range []string{versionOut, flagOut} {
		if strings.Contains(output, "\n") {
			t.Errorf("version output must be a single line, got:\n%s", output)
		}
	}

	if versionOut != flagOut {
		t.Errorf("version output = %q, --version output = %q; both must be equal", versionOut, flagOut)
	}

	re := regexp.MustCompile(`^taskflow version \d+\.\d+\.\d+$`)
	if !re.MatchString(versionOut) {
		t.Errorf("version output = %q, want match for %v", versionOut, re)
	}
	if !re.MatchString(flagOut) {
		t.Errorf("--version output = %q, want match for %v", flagOut, re)
	}
}
