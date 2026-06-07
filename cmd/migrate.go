package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rwbaskette/taskflow/internal/db"
	cliErrors "github.com/rwbaskette/taskflow/pkg/errors"
)

var (
	migrateDryRun bool
	migrateForce  bool
)

var migrateCmd = &cobra.Command{
	Use:     "migrate",
	Short:   "Migrate the database schema",
	Long:    "Migrate the database schema to the latest version. Shows status if database is behind.",
	Example: `  taskflow migrate
  taskflow migrate --dry-run
  taskflow migrate --version 1.0.0`,
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

		// Check for --version flag (show version and exit)
		if showVersion {
			fmt.Printf("Schema version: %s\n", currentVersion)
			return
		}

		// Get latest version
		latestVersion := db.LatestVersion

		// Compare versions
		if currentVersion == latestVersion {
			fmt.Printf("Database is already at the latest version: %s\n", latestVersion)
			return
		}

		// Check if current is "pre-migration" (old database before migrations existed)
		if currentVersion == "pre-migration" {
			fmt.Println("Legacy database detected (pre-migration format).")
			fmt.Printf("Latest schema version: %s\n", latestVersion)
			fmt.Println()
			fmt.Println("To migrate this database:")
			fmt.Println("  1. A backup will be created at: " + database.GetBackupPath())
			fmt.Println("  2. A new database will be created with the updated schema")
			fmt.Println("  3. Your data will be preserved if possible")
			fmt.Println()
			if !migrateDryRun && !migrateForce {
				fmt.Print("Proceed with migration? (y/N): ")
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" {
					fmt.Println("Migration cancelled.")
					return
				}
			}
		}

		// Check major version mismatch
		currentMajor := db.MajorVersion(currentVersion)
		latestMajor := db.MajorVersion(latestVersion)
		if currentMajor != latestMajor && currentVersion != "pre-migration" && !migrateForce {
			fmt.Printf("Cannot migrate across major versions (%s -> %s). Use --force to override.\n", currentVersion, latestVersion)
			os.Exit(1)
		}

		// Get pending migrations
		pending, err := database.GetPendingMigrations()
		if err != nil {
			cliErrors.HandleError(err)
			return
		}

		if len(pending) == 0 {
			fmt.Printf("Database is already at version %s\n", latestVersion)
			return
		}

		// Show what would happen
		fmt.Printf("Database version: %s\n", currentVersion)
		fmt.Printf("Latest version: %s\n", latestVersion)
		fmt.Printf("Pending migrations: %d\n", len(pending))
		for _, v := range pending {
			fmt.Printf("  - %s\n", v)
		}
		fmt.Println()

		if migrateDryRun {
			fmt.Println("Dry run - no changes were made.")
			fmt.Println("To apply migrations, run: taskflow migrate")
			return
		}

		// Check for existing backup
		if database.BackupExists() {
			fmt.Printf("ERROR: Backup already exists at %s\n", database.GetBackupPath())
			fmt.Println("Remove the backup to allow re-migration:")
			fmt.Printf("  rm %s\n", database.GetBackupPath())
			os.Exit(1)
		}

		// Confirm migration
		if !migrateForce {
			fmt.Print("Proceed with migration? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" {
				fmt.Println("Migration cancelled.")
				return
			}
		}

		// Perform migration
		fmt.Println()
		fmt.Println("Starting migration...")

		for _, version := range pending {
			fmt.Printf("Applying migration %s...\n", version)

			// Create backup before migration
			if err := database.BackupDB(); err != nil {
				// If backup already exists, that's ok
				if !strings.Contains(err.Error(), "backup already exists") {
					fmt.Printf("ERROR: Failed to create backup: %v\n", err)
					fmt.Printf("Backup location: %s\n", database.GetBackupPath())
					os.Exit(1)
				}
			} else {
				fmt.Printf("  Backup created: %s\n", database.GetBackupPath())
			}

			// Apply migration
			if err := database.RunMigration(version); err != nil {
				fmt.Printf("ERROR: Migration failed: %v\n", err)
				fmt.Println("Your database is unchanged, but backup was created.")
				fmt.Printf("Backup location: %s\n", database.GetBackupPath())
				os.Exit(1)
			}
			fmt.Printf("  Applied: %s\n", version)
		}

		// Get new version
		newVersion, _ := database.GetSchemaVersion()
		fmt.Println()
		fmt.Printf("Migration complete. Schema is now at version %s\n", newVersion)
	},
}

var showVersion bool

func init() {
	rootCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().BoolVar(&showVersion, "version", false, "Show current schema version and exit")
	migrateCmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "Show what migrations would be applied without applying them")
	migrateCmd.Flags().BoolVar(&migrateForce, "force", false, "Force migration even across major versions")
}