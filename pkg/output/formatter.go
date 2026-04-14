// Package output provides consistent output formatting for the CLI.
package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	cliErrors "github.com/user/project/pkg/errors"
)

// Re-export colors and error types from errors package
var (
	ColorRed    = cliErrors.ColorRed
	ColorGreen  = cliErrors.ColorGreen
	ColorYellow = cliErrors.ColorYellow
	ColorBlue   = cliErrors.ColorBlue
	ColorReset  = cliErrors.ColorReset
	ColorBold   = cliErrors.ColorBold
)

// PrintError prints a formatted CLI error to stderr.
// This function delegates to the errors package PrintError.
func PrintError(err error) {
	cliErrors.PrintError(err)
}

// PrintSuccess prints a success message to stdout.
func PrintSuccess(message string) {
	fmt.Printf("%s✓ %s%s\n", ColorGreen, message, ColorReset)
}

// PrintInfo prints an info message to stdout.
func PrintInfo(message string) {
	fmt.Printf("%sℹ %s%s\n", ColorBlue, message, ColorReset)
}

// PrintWarning prints a warning message to stdout.
func PrintWarning(message string) {
	fmt.Printf("%s⚠ %s%s\n", ColorYellow, message, ColorReset)
}

// PrintHelp prints help text with consistent formatting.
func PrintHelp(cmd *cobra.Command) {
	if cmd == nil {
		return
	}

	fmt.Println()
	fmt.Printf("%sUsage:%s\n", ColorBold, ColorReset)
	fmt.Printf("  %s\n\n", cmd.UseLine())

	if cmd.Short != "" {
		fmt.Printf("%sDescription:%s\n", ColorBold, ColorReset)
		fmt.Printf("  %s\n\n", cmd.Short)
	}

	if cmd.Long != "" {
		fmt.Printf("%sDetails:%s\n", ColorBold, ColorReset)
		fmt.Printf("  %s\n\n", cmd.Long)
	}

	if len(cmd.Flags().FlagUsages()) > 0 {
		fmt.Printf("%sOptions:%s\n", ColorBold, ColorReset)
		fmt.Printf("  %s\n\n", cmd.Flags().FlagUsages())
	}

	if cmd.Example != "" {
		fmt.Printf("%sExamples:%s\n", ColorBold, ColorReset)
		fmt.Printf("  %s\n\n", cmd.Example)
	}
}

// PrintVersion prints version information.
func PrintVersion(name, version, commit, date string) {
	fmt.Printf("%s %s\n", name, version)
	if commit != "" {
		fmt.Printf("  commit: %s\n", commit)
	}
	if date != "" {
		fmt.Printf("  date: %s\n", date)
	}
}

// FormatTaskList formats a list of tasks for display.
func FormatTaskList(tasks []map[string]interface{}) {
	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return
	}

	fmt.Println()
	fmt.Printf("%s%-5s %-20s %-15s %-15s %s%s\n",
		ColorBold, "ID", "Title", "Status", "Milestone", "Actor", ColorReset)
	fmt.Println(strings.Repeat("-", 80))

	for _, task := range tasks {
		id := fmt.Sprintf("%v", task["id"])
		title := fmt.Sprintf("%v", task["title"])
		status := fmt.Sprintf("%v", task["status"])
		milestone := fmt.Sprintf("%v", task["milestone"])
		actor := fmt.Sprintf("%v", task["actor"])

		// Truncate long titles
		if len(title) > 18 {
			title = title[:15] + "..."
		}

		fmt.Printf("%-5s %-20s %-15s %-15s %s\n",
			id, title, status, milestone, actor)
	}
	fmt.Println()
}

// FormatTaskDetails formats a single task for detailed display.
func FormatTaskDetails(task map[string]interface{}) {
	fmt.Println()
	fmt.Printf("%sTask Details:%s\n", ColorBold, ColorReset)
	fmt.Println(strings.Repeat("-", 40))

	fields := []string{"id", "title", "description", "status", "milestone", "actor", "created_at", "updated_at"}
	for _, field := range fields {
		if val, ok := task[field]; ok {
			fmt.Printf("  %-15s: %v\n", field, val)
		}
	}
	fmt.Println()
}

// PrintUsageError prints a usage error with help.
func PrintUsageError(command string, message string) {
	fmt.Fprintf(os.Stderr, "\n%sError: %s%s\n\n", ColorRed, message, ColorReset)
	fmt.Fprintf(os.Stderr, "Run 'task %s --help' for usage information.\n\n", command)
}

// Confirm prompts the user for confirmation.
func Confirm(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
