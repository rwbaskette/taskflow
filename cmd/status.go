package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rwbaskette/taskflow/internal/db"
	cliErrors "github.com/rwbaskette/taskflow/pkg/errors"
)

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Show database schema status",
	Long:    "Show the current schema version, latest version, database path, and pending migrations.",
	Example: `  taskflow status
  taskflow status --verbose`,
	Run: func(cmd *cobra.Command, args []string) {
		database, err := db.NewDB(db.DefaultDBPath())
		if err != nil {
			cliErrors.HandleError(err)
			return
		}
		defer database.Close()

		// Get current version
		currentVersion, err := database.GetSchemaVersion()
		if err != nil {
			cliErrors.HandleError(err)
			return
		}

		// Get latest version
		latestVersion := db.LatestVersion

		// Get pending migrations
		pending, err := database.GetPendingMigrations()
		if err != nil {
			cliErrors.HandleError(err)
			return
		}

		// Get database path
		dbPath := database.Path()

		// Get backup path
		backupPath := database.GetBackupPath()
		backupExists := database.BackupExists()

		// Calculate pending count
		pendingCount := len(pending)

		// Print status
		fmt.Println()
		fmt.Println("  Schema Status")
		fmt.Println("  ─────────────")

		// Version comparison
		if currentVersion == latestVersion {
			fmt.Printf("  Schema Version:    %s ✓\n", currentVersion)
		} else if pendingCount > 0 {
			fmt.Printf("  Schema Version:    %s (behind)\n", currentVersion)
		} else if currentVersion == "pre-migration" {
			fmt.Printf("  Schema Version:    %s (legacy)\n", currentVersion)
		} else {
			fmt.Printf("  Schema Version:    %s\n", currentVersion)
		}
		fmt.Printf("  Latest Version:    %s\n", latestVersion)

		// Pending migrations
		if pendingCount > 0 {
			fmt.Printf("  Pending:           %d migration(s)\n", pendingCount)
			for _, p := range pending {
				fmt.Printf("    - %s\n", p)
			}
		} else if currentVersion == latestVersion {
			fmt.Printf("  Pending:           none (up to date)\n")
		} else {
			fmt.Printf("  Pending:           none\n")
		}

		fmt.Println()
		fmt.Println("  Database")
		fmt.Println("  ─────────────")
		fmt.Printf("  Database:          %s\n", dbPath)

		if backupExists {
			fmt.Printf("  Backup:            %s (exists)\n", backupPath)
		} else {
			fmt.Printf("  Backup:            %s\n", backupPath)
		}

		// Show database size
		if dbSize, err := database.GetDBSize(); err == nil {
			fmt.Printf("  Size:              %d bytes\n", dbSize)
		}

		// Show backup size if exists
		if backupExists {
			if backupSize, err := database.GetBackupSize(); err == nil && backupSize > 0 {
				fmt.Printf("  Backup Size:       %d bytes\n", backupSize)
			}
		}

		fmt.Println()

		// Suggestion if behind
		if pendingCount > 0 {
			fmt.Println("  To migrate:        taskflow migrate")
		}

		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}