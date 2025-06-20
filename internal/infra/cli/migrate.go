package cli

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joern1811/memory-bank/internal/infra/database"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration management",
	Long:  `Manage database schema migrations including running, rolling back, and checking status.`,
}

// migrateUpCmd runs pending migrations
var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Run pending database migrations",
	Long:  `Execute all pending database migrations to bring the schema up to date.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := logrus.New()
		logger.SetLevel(logrus.InfoLevel)

		// Get database path from environment or use default
		dbPath := os.Getenv("MEMORY_BANK_DB_PATH")
		if dbPath == "" {
			dbPath = "./memory_bank.db"
		}

		// Connect to database
		db, err := database.NewSQLiteDatabase(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer func() {
			if err := db.Close(); err != nil {
				logger.WithError(err).Error("Failed to close database")
			}
		}()

		// Create migrator and run migrations
		migrator := database.NewMigrator(db, logger)
		if err := migrator.Run(); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}

		fmt.Println("All migrations completed successfully")
		return nil
	},
}

// migrateDownCmd rolls back the last migration
var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback the last database migration",
	Long:  `Roll back the most recently applied database migration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := logrus.New()
		logger.SetLevel(logrus.InfoLevel)

		// Get database path from environment or use default
		dbPath := os.Getenv("MEMORY_BANK_DB_PATH")
		if dbPath == "" {
			dbPath = "./memory_bank.db"
		}

		// Connect to database without running migrations
		db, err := connectToDatabase(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer func() {
			if err := db.Close(); err != nil {
				logger.WithError(err).Error("Failed to close database")
			}
		}()

		// Create migrator and rollback
		migrator := database.NewMigrator(db, logger)
		if err := migrator.Rollback(); err != nil {
			return fmt.Errorf("failed to rollback migration: %w", err)
		}

		fmt.Println("Migration rolled back successfully")
		return nil
	},
}

// migrateStatusCmd shows current migration status
var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current migration status",
	Long:  `Display the current database schema version and migration status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := logrus.New()
		logger.SetLevel(logrus.WarnLevel) // Reduce logging for status check

		// Get database path from environment or use default
		dbPath := os.Getenv("MEMORY_BANK_DB_PATH")
		if dbPath == "" {
			dbPath = "./memory_bank.db"
		}

		// Connect to database without running migrations
		db, err := connectToDatabase(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer func() {
			if err := db.Close(); err != nil {
				logger.WithError(err).Error("Failed to close database")
			}
		}()

		// Create migrator and get current version
		migrator := database.NewMigrator(db, logger)
		version, err := migrator.GetCurrentVersion()
		if err != nil {
			return fmt.Errorf("failed to get current version: %w", err)
		}

		fmt.Printf("Current schema version: %d\n", version)
		fmt.Printf("Database path: %s\n", dbPath)
		return nil
	},
}

// connectToDatabase creates a database connection without running migrations
func connectToDatabase(dbPath string, logger *logrus.Logger) (*sql.DB, error) {
	logger.WithField("db_path", dbPath).Info("Connecting to SQLite database")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			logger.WithError(closeErr).Error("Failed to close database after ping failure")
		}
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func init() {
	// Add migrate subcommands
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateStatusCmd)

	// Add migrate command to root
	rootCmd.AddCommand(migrateCmd)
}
