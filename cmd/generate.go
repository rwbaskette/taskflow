package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/user/project/pkg/generator"
)

var (
	agentType   string
	outputFile  string
	packageName string
	className   string
	binaryName  string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate wrapper code for task management",
	Long: `Generate wrapper code for various agent types (e.g., TypeScript, OpenCode).

This command creates client libraries that wrap the task management operations
for use in different programming environments.`,
	Example: `  # Generate TypeScript wrapper
  task generate --agent typescript

  # Generate TypeScript wrapper with custom package name
  task generate --agent typescript --package my-tasks

  # Generate TypeScript wrapper with custom class name
  task generate --agent typescript --class MyTaskManager

  # Generate OpenCode wrapper for integration
  task generate --agent opencode --binary-name task

  # Generate OpenCode wrapper and save to file
  task generate --agent opencode --output manage-tasks.sh`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate agent type
		validAgents := generator.GetSupportedAgents()
		valid := false
		for _, a := range validAgents {
			if a == agentType {
				valid = true
				break
			}
		}
		if !valid {
			fmt.Printf("Error: invalid agent type '%s'. Valid options: %v\n", agentType, validAgents)
			os.Exit(1)
		}

		// Build options
		opts := &generator.TypeScriptOptions{
			PackageName: packageName,
			ClassName:   className,
		}

		// Validate output file path if specified
		if outputFile != "" {
			// Get the current working directory as allowed base
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Printf("Error: could not determine working directory: %v\n", err)
				os.Exit(1)
			}
			// Clean and validate the path
			cleanPath, err := generator.CleanAndValidatePath(outputFile, cwd)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			// Ensure the directory exists
			dir := filepath.Dir(cleanPath)
			if dir != "." && dir != "" {
				if err := os.MkdirAll(dir, 0755); err != nil {
					fmt.Printf("Error creating directory: %v\n", err)
					os.Exit(1)
				}
			}
			outputFile = cleanPath
		}

		// Generate code
		code, err := generator.GenerateAgent(agentType, opts, binaryName)
		if err != nil {
			fmt.Printf("Error generating code: %v\n", err)
			os.Exit(1)
		}

		// Output to file or stdout
		if outputFile != "" {
			err := os.WriteFile(outputFile, []byte(code), 0644)
			if err != nil {
				fmt.Printf("Error writing to file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Generated %s wrapper to: %s\n", agentType, outputFile)
		} else {
			fmt.Print(code)
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&agentType, "agent", "a", "typescript", "Agent type to generate (typescript, opencode)")
	generateCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (prints to stdout if not specified)")
	generateCmd.Flags().StringVarP(&packageName, "package", "p", "taskflow", "Package name for TypeScript module")
	generateCmd.Flags().StringVarP(&className, "class", "c", "TaskManager", "Class name for TypeScript wrapper")
	generateCmd.Flags().StringVarP(&binaryName, "binary-name", "b", "task", "Binary name for OpenCode wrapper")
}
