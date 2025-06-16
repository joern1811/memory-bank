package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/ports"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"strings"
)

// NewSQLiteDatabase creates a new SQLite database connection and runs migrations
func NewSQLiteDatabase(dbPath string, logger *logrus.Logger) (*sql.DB, error) {
	logger.WithField("db_path", dbPath).Info("Connecting to SQLite database")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run database migrations
	migrator := NewMigrator(db, logger)
	if err := migrator.Run(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("SQLite database connected and migrated")
	return db, nil
}

// initializeTables creates the necessary database tables
func initializeTables(db *sql.DB) error {
	createMemoriesTable := `
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
	);`

	createProjectsTable := `
	CREATE TABLE IF NOT EXISTS projects (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		path TEXT UNIQUE NOT NULL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	createSessionsTable := `
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
	);`

	tables := []string{createMemoriesTable, createProjectsTable, createSessionsTable}
	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

// SQLiteMemoryRepository implements MemoryRepository using SQLite
type SQLiteMemoryRepository struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewSQLiteMemoryRepository creates a new SQLite memory repository
func NewSQLiteMemoryRepository(db *sql.DB, logger *logrus.Logger) *SQLiteMemoryRepository {
	return &SQLiteMemoryRepository{
		db:     db,
		logger: logger,
	}
}

// Store saves a memory to the database
func (r *SQLiteMemoryRepository) Store(ctx context.Context, memory *domain.Memory) error {
	r.logger.WithField("memory_id", memory.ID).Debug("Storing memory")

	tagsJSON, err := json.Marshal(memory.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		INSERT INTO memories (
			id, project_id, session_id, type, title, content, context, 
			tags, created_at, updated_at, has_embedding
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var sessionID interface{}
	if memory.SessionID != nil {
		sessionID = string(*memory.SessionID)
	}

	_, err = r.db.ExecContext(ctx, query,
		string(memory.ID),
		string(memory.ProjectID),
		sessionID,
		string(memory.Type),
		memory.Title,
		memory.Content,
		memory.Context,
		string(tagsJSON),
		memory.CreatedAt,
		memory.UpdatedAt,
		memory.HasEmbedding,
	)

	if err != nil {
		return fmt.Errorf("failed to insert memory: %w", err)
	}

	return nil
}

// GetByID retrieves a memory by its ID
func (r *SQLiteMemoryRepository) GetByID(ctx context.Context, id domain.MemoryID) (*domain.Memory, error) {
	r.logger.WithField("memory_id", id).Debug("Getting memory by ID")

	query := `
		SELECT id, project_id, session_id, type, title, content, context, 
		       tags, created_at, updated_at, has_embedding
		FROM memories 
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, string(id))
	return r.scanMemory(row)
}

// Update updates an existing memory
func (r *SQLiteMemoryRepository) Update(ctx context.Context, memory *domain.Memory) error {
	r.logger.WithField("memory_id", memory.ID).Debug("Updating memory")

	tagsJSON, err := json.Marshal(memory.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		UPDATE memories 
		SET project_id = ?, session_id = ?, type = ?, title = ?, content = ?, 
		    context = ?, tags = ?, updated_at = ?, has_embedding = ?
		WHERE id = ?
	`

	var sessionID interface{}
	if memory.SessionID != nil {
		sessionID = string(*memory.SessionID)
	}

	result, err := r.db.ExecContext(ctx, query,
		string(memory.ProjectID),
		sessionID,
		string(memory.Type),
		memory.Title,
		memory.Content,
		memory.Context,
		string(tagsJSON),
		memory.UpdatedAt,
		memory.HasEmbedding,
		string(memory.ID),
	)

	if err != nil {
		return fmt.Errorf("failed to update memory: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("memory not found")
	}

	return nil
}

// Delete removes a memory from the database
func (r *SQLiteMemoryRepository) Delete(ctx context.Context, id domain.MemoryID) error {
	r.logger.WithField("memory_id", id).Debug("Deleting memory")

	query := `DELETE FROM memories WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, string(id))
	if err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("memory not found")
	}

	return nil
}

// ListByProject retrieves all memories for a project
func (r *SQLiteMemoryRepository) ListByProject(ctx context.Context, projectID domain.ProjectID) ([]*domain.Memory, error) {
	r.logger.WithField("project_id", projectID).Debug("Listing memories by project")

	query := `
		SELECT id, project_id, session_id, type, title, content, context, 
		       tags, created_at, updated_at, has_embedding
		FROM memories 
		WHERE project_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, string(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to query memories: %w", err)
	}
	defer rows.Close()

	return r.scanMemories(rows)
}

// ListByType retrieves memories by type for a project
func (r *SQLiteMemoryRepository) ListByType(ctx context.Context, projectID domain.ProjectID, memoryType domain.MemoryType) ([]*domain.Memory, error) {
	r.logger.WithFields(logrus.Fields{
		"project_id": projectID,
		"type":       memoryType,
	}).Debug("Listing memories by type")

	query := `
		SELECT id, project_id, session_id, type, title, content, context, 
		       tags, created_at, updated_at, has_embedding
		FROM memories 
		WHERE project_id = ? AND type = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, string(projectID), string(memoryType))
	if err != nil {
		return nil, fmt.Errorf("failed to query memories: %w", err)
	}
	defer rows.Close()

	return r.scanMemories(rows)
}

// ListByTags retrieves memories that have all specified tags
func (r *SQLiteMemoryRepository) ListByTags(ctx context.Context, projectID domain.ProjectID, tags domain.Tags) ([]*domain.Memory, error) {
	r.logger.WithFields(logrus.Fields{
		"project_id": projectID,
		"tags":       tags,
	}).Debug("Listing memories by tags")

	// For SQLite, we'll use JSON functions to query tags
	// This is a simplified implementation - in production you might want a tags table
	query := `
		SELECT id, project_id, session_id, type, title, content, context, 
		       tags, created_at, updated_at, has_embedding
		FROM memories 
		WHERE project_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, string(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to query memories: %w", err)
	}
	defer rows.Close()

	memories, err := r.scanMemories(rows)
	if err != nil {
		return nil, err
	}

	// Filter memories that contain all required tags
	var filtered []*domain.Memory
	for _, memory := range memories {
		hasAllTags := true
		for _, requiredTag := range tags {
			if !memory.Tags.Contains(requiredTag) {
				hasAllTags = false
				break
			}
		}
		if hasAllTags {
			filtered = append(filtered, memory)
		}
	}

	return filtered, nil
}

// ListBySession retrieves all memories for a session
func (r *SQLiteMemoryRepository) ListBySession(ctx context.Context, sessionID domain.SessionID) ([]*domain.Memory, error) {
	r.logger.WithField("session_id", sessionID).Debug("Listing memories by session")

	query := `
		SELECT id, project_id, session_id, type, title, content, context, 
		       tags, created_at, updated_at, has_embedding
		FROM memories 
		WHERE session_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, string(sessionID))
	if err != nil {
		return nil, fmt.Errorf("failed to query memories: %w", err)
	}
	defer rows.Close()

	return r.scanMemories(rows)
}

// scanMemory scans a single memory from a row
func (r *SQLiteMemoryRepository) scanMemory(row *sql.Row) (*domain.Memory, error) {
	var memory domain.Memory
	var sessionID sql.NullString
	var tagsJSON string

	err := row.Scan(
		&memory.ID,
		&memory.ProjectID,
		&sessionID,
		&memory.Type,
		&memory.Title,
		&memory.Content,
		&memory.Context,
		&tagsJSON,
		&memory.CreatedAt,
		&memory.UpdatedAt,
		&memory.HasEmbedding,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("memory not found")
		}
		return nil, fmt.Errorf("failed to scan memory: %w", err)
	}

	// Handle nullable session ID
	if sessionID.Valid {
		sid := domain.SessionID(sessionID.String)
		memory.SessionID = &sid
	}

	// Unmarshal tags
	if err := json.Unmarshal([]byte(tagsJSON), &memory.Tags); err != nil {
		r.logger.WithError(err).Warn("Failed to unmarshal tags, using empty tags")
		memory.Tags = make(domain.Tags, 0)
	}

	return &memory, nil
}

// scanMemories scans multiple memories from rows
func (r *SQLiteMemoryRepository) scanMemories(rows *sql.Rows) ([]*domain.Memory, error) {
	var memories []*domain.Memory

	for rows.Next() {
		var memory domain.Memory
		var sessionID sql.NullString
		var tagsJSON string

		err := rows.Scan(
			&memory.ID,
			&memory.ProjectID,
			&sessionID,
			&memory.Type,
			&memory.Title,
			&memory.Content,
			&memory.Context,
			&tagsJSON,
			&memory.CreatedAt,
			&memory.UpdatedAt,
			&memory.HasEmbedding,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan memory: %w", err)
		}

		// Handle nullable session ID
		if sessionID.Valid {
			sid := domain.SessionID(sessionID.String)
			memory.SessionID = &sid
		}

		// Unmarshal tags
		if err := json.Unmarshal([]byte(tagsJSON), &memory.Tags); err != nil {
			r.logger.WithError(err).Warn("Failed to unmarshal tags, using empty tags")
			memory.Tags = make(domain.Tags, 0)
		}

		memories = append(memories, &memory)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return memories, nil
}

// GetByIDs retrieves multiple memories by their IDs in a single batch query
func (r *SQLiteMemoryRepository) GetByIDs(ctx context.Context, ids []domain.MemoryID) ([]*domain.Memory, error) {
	if len(ids) == 0 {
		return []*domain.Memory{}, nil
	}

	r.logger.WithField("batch_size", len(ids)).Debug("Batch retrieving memories by IDs")

	// Build IN clause with placeholders
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = string(id)
	}

	query := fmt.Sprintf(`
		SELECT id, project_id, session_id, type, title, content, context, 
		       tags, created_at, updated_at, has_embedding
		FROM memories 
		WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to batch query memories: %w", err)
	}
	defer rows.Close()

	return r.scanMemories(rows)
}

// GetMetadataByIDs retrieves lightweight metadata for multiple memories by their IDs
func (r *SQLiteMemoryRepository) GetMetadataByIDs(ctx context.Context, ids []domain.MemoryID) ([]*ports.MemoryMetadata, error) {
	if len(ids) == 0 {
		return []*ports.MemoryMetadata{}, nil
	}

	r.logger.WithField("batch_size", len(ids)).Debug("Batch retrieving memory metadata by IDs")

	// Build IN clause with placeholders
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = string(id)
	}

	query := fmt.Sprintf(`
		SELECT id, project_id, type, title, tags, created_at
		FROM memories 
		WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to batch query memory metadata: %w", err)
	}
	defer rows.Close()

	var metadata []*ports.MemoryMetadata
	for rows.Next() {
		var meta ports.MemoryMetadata
		var tagsJSON string

		err := rows.Scan(
			&meta.ID,
			&meta.ProjectID,
			&meta.Type,
			&meta.Title,
			&tagsJSON,
			&meta.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metadata: %w", err)
		}

		// Parse tags JSON
		if err := json.Unmarshal([]byte(tagsJSON), &meta.Tags); err != nil {
			r.logger.WithField("memory_id", meta.ID).Warn("Failed to parse tags JSON, using empty tags")
			meta.Tags = domain.Tags{}
		}

		metadata = append(metadata, &meta)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating metadata rows: %w", err)
	}

	r.logger.WithField("metadata_count", len(metadata)).Debug("Batch metadata retrieval completed")
	return metadata, nil
}

// InitializeSchema creates the necessary tables
func (r *SQLiteMemoryRepository) InitializeSchema(ctx context.Context) error {
	r.logger.Info("Initializing memory database schema")

	schema := `
	CREATE TABLE IF NOT EXISTS memories (
		id TEXT PRIMARY KEY,
		project_id TEXT NOT NULL,
		session_id TEXT,
		type TEXT NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		context TEXT NOT NULL,
		tags TEXT NOT NULL DEFAULT '[]',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		has_embedding BOOLEAN NOT NULL DEFAULT FALSE,
		
		INDEX idx_memories_project_id (project_id),
		INDEX idx_memories_type (type),
		INDEX idx_memories_session_id (session_id),
		INDEX idx_memories_created_at (created_at)
	);
	`

	_, err := r.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to create memories table: %w", err)
	}

	r.logger.Info("Memory database schema initialized successfully")
	return nil
}
