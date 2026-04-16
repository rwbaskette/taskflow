package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/user/project/pkg/generator"
)

var (
	outputFile string
	binaryPath string
)

var toolWrapperCmd = &cobra.Command{
	Use:   "tool-wrapper",
	Short: "Generate OpenCode tool wrapper for taskflow CLI",
	Long: `Generate TypeScript tool wrapper code for OpenCode agent integration.

This command creates TypeScript tool wrappers that wrap the taskflow CLI operations
for use in the OpenCode environment using the tool() helper format.`,
	Example: `  # Generate TypeScript tool wrapper for OpenCode integration
  taskflow tool-wrapper

  # Generate TypeScript tool wrapper with custom binary name
  taskflow tool-wrapper --binary-path my-task

  # Generate TypeScript tool wrapper and save to OpenCode tools directory
  taskflow tool-wrapper --output .opencode/tools/taskflow.ts`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Build options
		toolOpts := &generator.ToolWrapperOptions{
			BinaryPath: binaryPath,
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
		code, err := generator.GenerateToolWrapper(toolOpts)
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
			fmt.Printf("Generated tool wrapper to: %s\n", outputFile)
		} else {
			fmt.Print(code)
		}
	},
}

func init() {
	rootCmd.AddCommand(toolWrapperCmd)

	toolWrapperCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (prints to stdout if not specified)")
	toolWrapperCmd.Flags().StringVarP(&binaryPath, "binary-path", "b", "taskflow", "Binary path for taskflow command")
}
