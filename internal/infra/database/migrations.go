package database

import (
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
)

// Migration represents a database schema migration
type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}

// Migrator handles database migrations
type Migrator struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewMigrator creates a new database migrator
func NewMigrator(db *sql.DB, logger *logrus.Logger) *Migrator {
	return &Migrator{
		db:     db,
		logger: logger,
	}
}

// Run executes all pending migrations
func (m *Migrator) Run() error {
	m.logger.Info("Starting database migrations")

	// Create migrations table if it doesn't exist
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current schema version
	currentVersion, err := m.getCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Get all migrations
	migrations := m.getAllMigrations()
	
	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	// Run pending migrations
	executed := 0
	for _, migration := range migrations {
		if migration.Version > currentVersion {
			if err := m.runMigration(migration); err != nil {
				return fmt.Errorf("failed to run migration %d: %w", migration.Version, err)
			}
			executed++
		}
	}

	m.logger.WithField("executed", executed).Info("Database migrations completed")
	return nil
}

// Rollback rolls back the last migration
func (m *Migrator) Rollback() error {
	m.logger.Info("Rolling back last migration")

	currentVersion, err := m.getCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if currentVersion == 0 {
		m.logger.Info("No migrations to rollback")
		return nil
	}

	// Find the migration to rollback
	migrations := m.getAllMigrations()
	var migrationToRollback *Migration
	for _, migration := range migrations {
		if migration.Version == currentVersion {
			migrationToRollback = &migration
			break
		}
	}

	if migrationToRollback == nil {
		return fmt.Errorf("migration with version %d not found", currentVersion)
	}

	// Execute rollback
	if err := m.rollbackMigration(*migrationToRollback); err != nil {
		return fmt.Errorf("failed to rollback migration %d: %w", currentVersion, err)
	}

	m.logger.WithField("version", currentVersion).Info("Migration rolled back successfully")
	return nil
}

// GetCurrentVersion returns the current schema version
func (m *Migrator) GetCurrentVersion() (int, error) {
	if err := m.createMigrationsTable(); err != nil {
		return 0, fmt.Errorf("failed to create migrations table: %w", err)
	}
	return m.getCurrentVersion()
}

// createMigrationsTable creates the migrations tracking table
func (m *Migrator) createMigrationsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := m.db.Exec(query)
	return err
}

// getCurrentVersion gets the current schema version from the database
func (m *Migrator) getCurrentVersion() (int, error) {
	query := "SELECT COALESCE(MAX(version), 0) FROM schema_migrations"
	
	var version int
	err := m.db.QueryRow(query).Scan(&version)
	if err != nil {
		return 0, err
	}
	
	return version, nil
}

// runMigration executes a single migration
func (m *Migrator) runMigration(migration Migration) error {
	m.logger.WithFields(logrus.Fields{
		"version": migration.Version,
		"name":    migration.Name,
	}).Info("Running migration")

	// Start transaction
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			m.logger.WithError(err).Warn("Failed to rollback transaction")
		}
	}()

	// Execute migration SQL
	if _, err := tx.Exec(migration.Up); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration in schema_migrations table
	insertQuery := "INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)"
	if _, err := tx.Exec(insertQuery, migration.Version, migration.Name, time.Now()); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	m.logger.WithFields(logrus.Fields{
		"version": migration.Version,
		"name":    migration.Name,
	}).Info("Migration completed successfully")

	return nil
}

// rollbackMigration rolls back a single migration
func (m *Migrator) rollbackMigration(migration Migration) error {
	m.logger.WithFields(logrus.Fields{
		"version": migration.Version,
		"name":    migration.Name,
	}).Info("Rolling back migration")

	// Start transaction
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			m.logger.WithError(err).Warn("Failed to rollback transaction")
		}
	}()

	// Execute rollback SQL
	if _, err := tx.Exec(migration.Down); err != nil {
		return fmt.Errorf("failed to execute rollback SQL: %w", err)
	}

	// Remove migration record from schema_migrations table
	deleteQuery := "DELETE FROM schema_migrations WHERE version = ?"
	if _, err := tx.Exec(deleteQuery, migration.Version); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	m.logger.WithFields(logrus.Fields{
		"version": migration.Version,
		"name":    migration.Name,
	}).Info("Migration rollback completed successfully")

	return nil
}

// getAllMigrations returns all available migrations
func (m *Migrator) getAllMigrations() []Migration {
	return []Migration{
		{
			Version: 1,
			Name:    "initial_schema",
			Up: `
			CREATE TABLE IF NOT EXISTS memories (
				id TEXT PRIMARY KEY,
				project_id TEXT NOT NULL,
				session_id TEXT,
				type TEXT NOT NULL,
				title TEXT NOT NULL,
				content TEXT NOT NULL,
				context TEXT,
				tags TEXT, -- JSON array
				metadata TEXT, -- JSON object
				embedding BLOB, -- Binary embedding vector
				has_embedding BOOLEAN DEFAULT FALSE,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);

			CREATE TABLE IF NOT EXISTS projects (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				path TEXT UNIQUE NOT NULL,
				description TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);

			CREATE TABLE IF NOT EXISTS sessions (
				id TEXT PRIMARY KEY,
				project_id TEXT NOT NULL,
				name TEXT NOT NULL,
				description TEXT,
				status TEXT DEFAULT 'active',
				started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				completed_at DATETIME,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (project_id) REFERENCES projects (id)
			);
			`,
			Down: `
			DROP TABLE IF EXISTS sessions;
			DROP TABLE IF EXISTS memories;
			DROP TABLE IF EXISTS projects;
			`,
		},
		{
			Version: 2,
			Name:    "add_memory_indexes",
			Up: `
			CREATE INDEX IF NOT EXISTS idx_memories_project_id ON memories(project_id);
			CREATE INDEX IF NOT EXISTS idx_memories_session_id ON memories(session_id);
			CREATE INDEX IF NOT EXISTS idx_memories_type ON memories(type);
			CREATE INDEX IF NOT EXISTS idx_memories_created_at ON memories(created_at);
			CREATE INDEX IF NOT EXISTS idx_memories_has_embedding ON memories(has_embedding);
			
			CREATE INDEX IF NOT EXISTS idx_projects_path ON projects(path);
			CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects(created_at);
			
			CREATE INDEX IF NOT EXISTS idx_sessions_project_id ON sessions(project_id);
			CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
			CREATE INDEX IF NOT EXISTS idx_sessions_started_at ON sessions(started_at);
			`,
			Down: `
			DROP INDEX IF EXISTS idx_memories_project_id;
			DROP INDEX IF EXISTS idx_memories_session_id;
			DROP INDEX IF EXISTS idx_memories_type;
			DROP INDEX IF EXISTS idx_memories_created_at;
			DROP INDEX IF EXISTS idx_memories_has_embedding;
			
			DROP INDEX IF EXISTS idx_projects_path;
			DROP INDEX IF EXISTS idx_projects_created_at;
			
			DROP INDEX IF EXISTS idx_sessions_project_id;
			DROP INDEX IF EXISTS idx_sessions_status;
			DROP INDEX IF EXISTS idx_sessions_started_at;
			`,
		},
	}
}