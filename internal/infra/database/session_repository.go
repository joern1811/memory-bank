package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/sirupsen/logrus"
)

// SQLiteSessionRepository implements the SessionRepository interface using SQLite
// Maps to existing sessions table schema: (id, project_id, name, description, status, started_at, completed_at)
type SQLiteSessionRepository struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewSQLiteSessionRepository creates a new SQLite session repository
func NewSQLiteSessionRepository(db *sql.DB, logger *logrus.Logger) *SQLiteSessionRepository {
	return &SQLiteSessionRepository{
		db:     db,
		logger: logger,
	}
}

// Store stores a new session in the database
func (r *SQLiteSessionRepository) Store(ctx context.Context, session *domain.Session) error {
	query := `
		INSERT INTO sessions (id, project_id, name, description, status, started_at, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	var completedAt *time.Time
	if session.EndTime != nil {
		completedAt = session.EndTime
	}

	// Build description with outcome and progress information
	description := ""
	if session.Outcome != "" {
		description = "Outcome: " + session.Outcome
	}
	if len(session.ProgressLog) > 0 {
		if description != "" {
			description += " | "
		}
		description += fmt.Sprintf("Progress: %s", strings.Join(session.ProgressLog, "; "))
	}

	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.ProjectID,
		session.TaskDescription, // name = task description
		description,             // description = outcome + progress
		session.Status,
		session.StartTime, // started_at
		completedAt,       // completed_at
	)

	if err != nil {
		r.logger.WithError(err).WithField("session_id", session.ID).Error("Failed to store session")
		return fmt.Errorf("failed to store session: %w", err)
	}

	r.logger.WithField("session_id", session.ID).Debug("Session stored successfully")
	return nil
}

// GetByID retrieves a session by its ID
func (r *SQLiteSessionRepository) GetByID(ctx context.Context, id domain.SessionID) (*domain.Session, error) {
	query := `
		SELECT id, project_id, name, description, status, started_at, completed_at
		FROM sessions
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)

	session := &domain.Session{}
	var description sql.NullString
	var completedAt sql.NullTime

	err := row.Scan(
		&session.ID,
		&session.ProjectID,
		&session.TaskDescription, // name -> TaskDescription
		&description,
		&session.Status,
		&session.StartTime, // started_at -> StartTime
		&completedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found: %s", id)
		}
		r.logger.WithError(err).WithField("session_id", id).Error("Failed to get session")
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Handle completed_at -> EndTime
	if completedAt.Valid {
		session.EndTime = &completedAt.Time
	}

	// Parse description to extract outcome and progress
	r.parseDescription(description.String, session)

	r.logger.WithField("session_id", id).Debug("Session retrieved successfully")
	return session, nil
}

// parseDescription extracts outcome and progress from stored description
func (r *SQLiteSessionRepository) parseDescription(description string, session *domain.Session) {
	if description == "" {
		session.ProgressLog = []string{}
		return
	}

	parts := strings.Split(description, " | ")
	for _, part := range parts {
		if strings.HasPrefix(part, "Outcome: ") {
			session.Outcome = strings.TrimPrefix(part, "Outcome: ")
		} else if strings.HasPrefix(part, "Progress: ") {
			progressStr := strings.TrimPrefix(part, "Progress: ")
			if progressStr != "" {
				session.ProgressLog = strings.Split(progressStr, "; ")
			}
		}
	}
	
	// Ensure ProgressLog is never nil
	if session.ProgressLog == nil {
		session.ProgressLog = []string{}
	}
}

// buildDescription creates description string from outcome and progress
func (r *SQLiteSessionRepository) buildDescription(session *domain.Session) string {
	description := ""
	if session.Outcome != "" {
		description = "Outcome: " + session.Outcome
	}
	if len(session.ProgressLog) > 0 {
		if description != "" {
			description += " | "
		}
		description += fmt.Sprintf("Progress: %s", strings.Join(session.ProgressLog, "; "))
	}
	return description
}

// Update updates an existing session in the database
func (r *SQLiteSessionRepository) Update(ctx context.Context, session *domain.Session) error {
	query := `
		UPDATE sessions
		SET project_id = ?, name = ?, description = ?, status = ?, started_at = ?, completed_at = ?
		WHERE id = ?
	`

	var completedAt *time.Time
	if session.EndTime != nil {
		completedAt = session.EndTime
	}

	description := r.buildDescription(session)

	result, err := r.db.ExecContext(ctx, query,
		session.ProjectID,
		session.TaskDescription,
		description,
		session.Status,
		session.StartTime,
		completedAt,
		session.ID,
	)

	if err != nil {
		r.logger.WithError(err).WithField("session_id", session.ID).Error("Failed to update session")
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", session.ID)
	}

	r.logger.WithField("session_id", session.ID).Debug("Session updated successfully")
	return nil
}

// Delete deletes a session from the database
func (r *SQLiteSessionRepository) Delete(ctx context.Context, id domain.SessionID) error {
	query := `DELETE FROM sessions WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.WithError(err).WithField("session_id", id).Error("Failed to delete session")
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", id)
	}

	r.logger.WithField("session_id", id).Debug("Session deleted successfully")
	return nil
}

// ListByProject retrieves all sessions for a specific project
func (r *SQLiteSessionRepository) ListByProject(ctx context.Context, projectID domain.ProjectID) ([]*domain.Session, error) {
	query := `
		SELECT id, project_id, name, description, status, started_at, completed_at
		FROM sessions
		WHERE project_id = ?
		ORDER BY started_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		r.logger.WithError(err).WithField("project_id", projectID).Error("Failed to list sessions")
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*domain.Session

	for rows.Next() {
		session := &domain.Session{}
		var description sql.NullString
		var completedAt sql.NullTime

		err := rows.Scan(
			&session.ID,
			&session.ProjectID,
			&session.TaskDescription,
			&description,
			&session.Status,
			&session.StartTime,
			&completedAt,
		)

		if err != nil {
			r.logger.WithError(err).Error("Failed to scan session row")
			continue
		}

		// Handle completed_at
		if completedAt.Valid {
			session.EndTime = &completedAt.Time
		}

		// Parse description
		r.parseDescription(description.String, session)

		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating over session rows")
		return nil, fmt.Errorf("error iterating over session rows: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"project_id":     projectID,
		"sessions_count": len(sessions),
	}).Debug("Sessions listed successfully")

	return sessions, nil
}

// GetActiveSession retrieves the active session for a project
func (r *SQLiteSessionRepository) GetActiveSession(ctx context.Context, projectID domain.ProjectID) (*domain.Session, error) {
	query := `
		SELECT id, project_id, name, description, status, started_at, completed_at
		FROM sessions
		WHERE project_id = ? AND status = ?
		ORDER BY started_at DESC
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, projectID, domain.SessionStatusActive)

	session := &domain.Session{}
	var description sql.NullString
	var completedAt sql.NullTime

	err := row.Scan(
		&session.ID,
		&session.ProjectID,
		&session.TaskDescription,
		&description,
		&session.Status,
		&session.StartTime,
		&completedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active session found for project: %s", projectID)
		}
		r.logger.WithError(err).WithField("project_id", projectID).Error("Failed to get active session")
		return nil, fmt.Errorf("failed to get active session: %w", err)
	}

	// Handle completed_at
	if completedAt.Valid {
		session.EndTime = &completedAt.Time
	}

	// Parse description
	r.parseDescription(description.String, session)

	r.logger.WithFields(logrus.Fields{
		"project_id": projectID,
		"session_id": session.ID,
	}).Debug("Active session retrieved successfully")

	return session, nil
}