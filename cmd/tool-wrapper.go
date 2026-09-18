package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rwbaskette/taskflow/internal/generator"
)

// cleanAndValidatePath sanitizes a file path and ensures it doesn't escape
// the allowed directory. (Moved here from pkg/generator: the tool-wrapper
// command is its only caller.)
func cleanAndValidatePath(outputPath string, allowedDir string) (string, error) {
	// Convert allowed dir to absolute
	absAllowed, err := filepath.Abs(allowedDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve allowed directory: %w", err)
	}

	// For relative output paths, join with the allowed dir first; absolute
	// paths are validated as-is. (filepath.Abs never fails on an absolute
	// path, so the old separate abs-branch add-on call was a no-op.)
	joined := outputPath
	if !filepath.IsAbs(outputPath) {
		joined = filepath.Join(absAllowed, outputPath)
	}
	cleaned := filepath.Clean(joined)

	// Verify the cleaned path is still within allowed dir
	if !strings.HasPrefix(cleaned, absAllowed+string(filepath.Separator)) && cleaned != absAllowed {
		return "", fmt.Errorf("path traversal detected: %s is not within %s", outputPath, allowedDir)
	}

	return cleaned, nil
}

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
	RunE: func(cmd *cobra.Command, args []string) error {
		// Build options
		toolOpts := &generator.ToolWrapperOptions{
			BinaryPath: binaryPath,
			Version:    version,
		}

		// Validate output file path if specified
		if outputFile != "" {
			// Get the current working directory as allowed base
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("could not determine working directory: %w", err)
			}
			// Clean and validate the path
			cleanPath, err := cleanAndValidatePath(outputFile, cwd)
			if err != nil {
				return err
			}
			// Ensure the directory exists
			dir := filepath.Dir(cleanPath)
			if dir != "." && dir != "" {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return fmt.Errorf("creating directory: %w", err)
				}
			}
			outputFile = cleanPath
		}

		// Generate code
		code, err := generator.GenerateToolWrapper(toolOpts)
		if err != nil {
			return fmt.Errorf("generating code: %w", err)
		}

		// Output to file or stdout
		if outputFile != "" {
			if err := os.WriteFile(outputFile, []byte(code), 0644); err != nil {
				return fmt.Errorf("writing to file: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Generated tool wrapper to: %s\n", outputFile)
			return nil
		}
		_, err = cmd.OutOrStdout().Write([]byte(code))
		return err
	},
}

func init() {
	rootCmd.AddCommand(toolWrapperCmd)

	toolWrapperCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (prints to stdout if not specified)")
	toolWrapperCmd.Flags().StringVarP(&binaryPath, "binary-path", "b", "taskflow", "Binary path for taskflow command")
}
