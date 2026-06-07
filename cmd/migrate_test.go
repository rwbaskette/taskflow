package cmd

import (
	"path/filepath"
	"testing"

	"github.com/rwbaskette/taskflow/internal/db"
	"github.com/stretchr/testify/require"
)

// TestMigrateVersion tests the schema version functionality
func TestMigrateVersion(t *testing.T) {
	tests := []struct {
		name           string
		setupDB        func(t *testing.T, tmpDir string) (*db.DB, func())
		expectedErr    bool
		expectedErrMsg string
		verifyVersion  func(t *testing.T, version string)
	}{
		{
			name: "new database starts at latest version",
			setupDB: func(t *testing.T, tmpDir string) (*db.DB, func()) {
				dbPath := filepath.Join(tmpDir, "tasks.db")
				database, err := db.NewDB(dbPath)
				require.NoError(t, err)
				return database, func() { database.Close() }
			},
			expectedErr: false,
			verifyVersion: func(t *testing.T, version string) {
				require.NotEmpty(t, version, "schema version should not be empty for new database")
			},
		},
		{
			name: "schema version is non-empty",
			setupDB: func(t *testing.T, tmpDir string) (*db.DB, func()) {
				dbPath := filepath.Join(tmpDir, "tasks.db")
				database, err := db.NewDB(dbPath)
				require.NoError(t, err)
				return database, func() { database.Close() }
			},
			expectedErr: false,
			verifyVersion: func(t *testing.T, version string) {
				require.NotEmpty(t, version, "schema version should not be empty")
			},
		},
		{
			name: "schema version matches latest version",
			setupDB: func(t *testing.T, tmpDir string) (*db.DB, func()) {
				dbPath := filepath.Join(tmpDir, "tasks.db")
				database, err := db.NewDB(dbPath)
				require.NoError(t, err)
				return database, func() { database.Close() }
			},
			expectedErr: false,
			verifyVersion: func(t *testing.T, version string) {
				require.Equal(t, db.LatestVersion, version, "new database should be at latest version")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			database, cleanup := tt.setupDB(t, tmpDir)
			defer cleanup()

			version, err := database.GetSchemaVersion()
			if tt.expectedErr {
				if tt.expectedErrMsg != "" {
					require.Error(t, err)
					require.Contains(t, err.Error(), tt.expectedErrMsg)
				}
				return
			}

			require.NoError(t, err)
			if tt.verifyVersion != nil {
				tt.verifyVersion(t, version)
			}
		})
	}
}

// TestMigrateDryRun tests the dry-run functionality
func TestMigrateDryRun(t *testing.T) {
	tests := []struct {
		name          string
		setupDB       func(t *testing.T, tmpDir string) (*db.DB, func())
		expectPending bool
		pendingCount  int
	}{
		{
			name: "new database has no pending migrations",
			setupDB: func(t *testing.T, tmpDir string) (*db.DB, func()) {
				dbPath := filepath.Join(tmpDir, "tasks.db")
				database, err := db.NewDB(dbPath)
				require.NoError(t, err)
				return database, func() { database.Close() }
			},
			expectPending: false,
			pendingCount:  0,
		},
		{
			name: "get pending migrations returns empty slice when up to date",
			setupDB: func(t *testing.T, tmpDir string) (*db.DB, func()) {
				dbPath := filepath.Join(tmpDir, "tasks.db")
				database, err := db.NewDB(dbPath)
				require.NoError(t, err)
				return database, func() { database.Close() }
			},
			expectPending: false,
			pendingCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			database, cleanup := tt.setupDB(t, tmpDir)
			defer cleanup()

			pending, err := database.GetPendingMigrations()
			require.NoError(t, err)

			if tt.expectPending {
				require.NotEmpty(t, pending, "should have pending migrations")
			} else {
				require.Empty(t, pending, "should have no pending migrations")
			}
		})
	}
}

// TestMigrateStatus tests the migration status functionality
func TestMigrateStatus(t *testing.T) {
	tests := []struct {
		name             string
		setupDB          func(t *testing.T, tmpDir string) (*db.DB, func())
		expectedVersion string
		expectedLatest  string
		expectUpToDate  bool
		expectPending   int
	}{
		{
			name: "fresh database is at latest version",
			setupDB: func(t *testing.T, tmpDir string) (*db.DB, func()) {
				dbPath := filepath.Join(tmpDir, "tasks.db")
				database, err := db.NewDB(dbPath)
				require.NoError(t, err)
				return database, func() { database.Close() }
			},
			expectedVersion: db.LatestVersion,
			expectedLatest:  db.LatestVersion,
			expectUpToDate:  true,
			expectPending:   0,
		},
		{
			name: "version comparison works correctly",
			setupDB: func(t *testing.T, tmpDir string) (*db.DB, func()) {
				dbPath := filepath.Join(tmpDir, "tasks.db")
				database, err := db.NewDB(dbPath)
				require.NoError(t, err)
				return database, func() { database.Close() }
			},
			expectedVersion: db.LatestVersion,
			expectedLatest:  db.LatestVersion,
			expectUpToDate:  true,
			expectPending:   0,
		},
		{
			name: "applied migrations list is accessible",
			setupDB: func(t *testing.T, tmpDir string) (*db.DB, func()) {
				dbPath := filepath.Join(tmpDir, "tasks.db")
				database, err := db.NewDB(dbPath)
				require.NoError(t, err)
				return database, func() { database.Close() }
			},
			expectedVersion: db.LatestVersion,
			expectedLatest:  db.LatestVersion,
			expectUpToDate:  true,
			expectPending:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			database, cleanup := tt.setupDB(t, tmpDir)
			defer cleanup()

			currentVersion, err := database.GetSchemaVersion()
			require.NoError(t, err)
			require.Equal(t, tt.expectedVersion, currentVersion)

			latestVersion := db.LatestVersion
			require.Equal(t, tt.expectedLatest, latestVersion)

			require.Equal(t, tt.expectUpToDate, currentVersion == latestVersion)

			pending, err := database.GetPendingMigrations()
			require.NoError(t, err)
			require.Len(t, pending, tt.expectPending)

			applied, err := database.GetAppliedMigrations()
			require.NoError(t, err)
			require.NotEmpty(t, applied, "new database should have applied migrations")
		})
	}
}

// TestStatusCommand tests the status command functionality
func TestStatusCommand(t *testing.T) {
	tests := []struct {
		name            string
		setupDB         func(t *testing.T, tmpDir string) (*db.DB, func())
		expectNoError   bool
		checkDB         func(t *testing.T, database *db.DB)
		expectDBPath    bool
		expectBackupPath bool
	}{
		{
			name: "status command returns database info",
			setupDB: func(t *testing.T, tmpDir string) (*db.DB, func()) {
				dbPath := filepath.Join(tmpDir, "tasks.db")
				database, err := db.NewDB(dbPath)
				require.NoError(t, err)
				return database, func() { database.Close() }
			},
			expectNoError:   true,
			expectDBPath:    true,
			expectBackupPath: true,
			checkDB: func(t *testing.T, database *db.DB) {
				require.NotEmpty(t, database.Path())
			},
		},
		{
			name: "status command shows backup info",
			setupDB: func(t *testing.T, tmpDir string) (*db.DB, func()) {
				dbPath := filepath.Join(tmpDir, "tasks.db")
				database, err := db.NewDB(dbPath)
				require.NoError(t, err)
				return database, func() { database.Close() }
			},
			expectNoError:   true,
			expectDBPath:    true,
			expectBackupPath: true,
			checkDB: func(t *testing.T, database *db.DB) {
				backupPath := database.GetBackupPath()
				require.NotEmpty(t, backupPath, "backup path should be available")
				require.True(t, len(backupPath) > 0, "backup path should be set")
			},
		},
		{
			name: "status command shows backup exists status",
			setupDB: func(t *testing.T, tmpDir string) (*db.DB, func()) {
				dbPath := filepath.Join(tmpDir, "tasks.db")
				database, err := db.NewDB(dbPath)
				require.NoError(t, err)
				return database, func() { database.Close() }
			},
			expectNoError:    true,
			expectDBPath:     true,
			expectBackupPath: true,
			checkDB: func(t *testing.T, database *db.DB) {
				backupExists := database.BackupExists()
				require.False(t, backupExists, "backup should not exist initially")
			},
		},
		{
			name: "status command shows database size",
			setupDB: func(t *testing.T, tmpDir string) (*db.DB, func()) {
				dbPath := filepath.Join(tmpDir, "tasks.db")
				database, err := db.NewDB(dbPath)
				require.NoError(t, err)
				return database, func() { database.Close() }
			},
			expectNoError:   true,
			expectDBPath:    true,
			expectBackupPath: true,
			checkDB: func(t *testing.T, database *db.DB) {
				dbSize, err := database.GetDBSize()
				require.NoError(t, err)
				require.Greater(t, dbSize, int64(0), "database size should be greater than 0")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			database, cleanup := tt.setupDB(t, tmpDir)
			defer cleanup()

			if tt.checkDB != nil {
				tt.checkDB(t, database)
			}

			if tt.expectDBPath {
				require.NotEmpty(t, database.Path())
			}

			if tt.expectBackupPath {
				backupPath := database.GetBackupPath()
				require.NotEmpty(t, backupPath)
			}

			if tt.checkDB != nil {
				tt.checkDB(t, database)
			}
		})
	}
}

// TestBackupAndRestore tests backup and restore functionality
func TestBackupAndRestore(t *testing.T) {
	tests := []struct {
		name        string
		setupDB     func(t *testing.T, tmpDir string) (*db.DB, func())
		verifyBackup func(t *testing.T, database *db.DB)
	}{
		{
			name: "backup can be created",
			setupDB: func(t *testing.T, tmpDir string) (*db.DB, func()) {
				dbPath := filepath.Join(tmpDir, "tasks.db")
				database, err := db.NewDB(dbPath)
				require.NoError(t, err)
				return database, func() { database.Close() }
			},
			verifyBackup: func(t *testing.T, database *db.DB) {
				err := database.BackupDB()
				require.NoError(t, err)

				exists := database.BackupExists()
				require.True(t, exists, "backup should exist after creation")
			},
		},
		{
			name: "backup prevents duplicate backup",
			setupDB: func(t *testing.T, tmpDir string) (*db.DB, func()) {
				dbPath := filepath.Join(tmpDir, "tasks.db")
				database, err := db.NewDB(dbPath)
				require.NoError(t, err)
				return database, func() { database.Close() }
			},
			verifyBackup: func(t *testing.T, database *db.DB) {
				err := database.BackupDB()
				require.NoError(t, err)

				err = database.BackupDB()
				require.Error(t, err, "creating duplicate backup should fail")
				require.Contains(t, err.Error(), "already exists")
			},
		},
		{
			name: "backup path is correctly formed",
			setupDB: func(t *testing.T, tmpDir string) (*db.DB, func()) {
				dbPath := filepath.Join(tmpDir, "tasks.db")
				database, err := db.NewDB(dbPath)
				require.NoError(t, err)
				return database, func() { database.Close() }
			},
			verifyBackup: func(t *testing.T, database *db.DB) {
				backupPath := database.GetBackupPath()
				require.NotEmpty(t, backupPath)
				require.Contains(t, backupPath, ".backup", "backup path should have .backup extension")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			database, cleanup := tt.setupDB(t, tmpDir)
			defer cleanup()

			tt.verifyBackup(t, database)
		})
	}
}