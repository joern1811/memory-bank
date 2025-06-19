package app

import (
	"context"
	"fmt"
	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/sirupsen/logrus"
)

// SessionService implements the session service use cases
type SessionService struct {
	sessionRepo ports.SessionRepository
	projectRepo ports.ProjectRepository
	logger      *logrus.Logger
}

// NewSessionService creates a new session service
func NewSessionService(
	sessionRepo ports.SessionRepository,
	projectRepo ports.ProjectRepository,
	logger *logrus.Logger,
) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		projectRepo: projectRepo,
		logger:      logger,
	}
}

// StartSession starts a new development session
func (s *SessionService) StartSession(ctx context.Context, req ports.StartSessionRequest) (*domain.Session, error) {
	s.logger.WithFields(logrus.Fields{
		"project_id": req.ProjectID,
		"task":       req.TaskDescription,
	}).Info("Starting session")

	// Verify project exists
	project, err := s.projectRepo.GetByID(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Check if there's already an active session for this project
	activeSession, err := s.sessionRepo.GetActiveSession(ctx, req.ProjectID)
	if err == nil && activeSession != nil {
		s.logger.WithField("existing_session_id", activeSession.ID).Warn("Active session already exists, aborting it")
		activeSession.Abort("Starting new session")
		if err := s.sessionRepo.Update(ctx, activeSession); err != nil {
			s.logger.WithError(err).Warn("Failed to abort existing session")
		}
	}

	// Create new session
	session := domain.NewSession(req.ProjectID, req.TaskDescription, req.TaskDescription)

	// Store session
	if err := s.sessionRepo.Store(ctx, session); err != nil {
		s.logger.WithError(err).Error("Failed to store session")
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"session_id": session.ID,
		"project":    project.Name,
	}).Info("Session started successfully")

	return session, nil
}

// GetSession retrieves a session by ID
func (s *SessionService) GetSession(ctx context.Context, id domain.SessionID) (*domain.Session, error) {
	return s.sessionRepo.GetByID(ctx, id)
}

// GetActiveSession retrieves the active session for a project
func (s *SessionService) GetActiveSession(ctx context.Context, projectID domain.ProjectID) (*domain.Session, error) {
	return s.sessionRepo.GetActiveSession(ctx, projectID)
}

// LogProgress adds a progress entry to a session
func (s *SessionService) LogProgress(ctx context.Context, sessionID domain.SessionID, entry string) error {
	s.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"entry":      entry,
	}).Debug("Logging session progress")

	// Get session
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Check if session is active
	if !session.IsActive() {
		return fmt.Errorf("session is not active")
	}

	// Add progress entry
	session.LogInfo(entry)

	// Update session
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// CompleteSession marks a session as completed
func (s *SessionService) CompleteSession(ctx context.Context, sessionID domain.SessionID, outcome string) error {
	s.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"outcome":    outcome,
	}).Info("Completing session")

	// Get session
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Check if session is active
	if !session.IsActive() {
		return fmt.Errorf("session is not active")
	}

	// Complete session
	session.Complete(outcome)

	// Update session
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"session_id": session.ID,
		"duration":   session.Duration(),
	}).Info("Session completed successfully")

	return nil
}

// AbortSession marks a session as aborted
func (s *SessionService) AbortSession(ctx context.Context, sessionID domain.SessionID) error {
	s.logger.WithField("session_id", sessionID).Info("Aborting session")

	// Get session
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Check if session is active
	if !session.IsActive() {
		return fmt.Errorf("session is not active")
	}

	// Abort session
	session.Abort("Manually aborted")

	// Update session
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"session_id": session.ID,
		"duration":   session.Duration(),
	}).Info("Session aborted")

	return nil
}

// Update updates an existing session
func (s *SessionService) Update(ctx context.Context, session *domain.Session) error {
	return s.sessionRepo.Update(ctx, session)
}

// ListSessions lists sessions with filters
func (s *SessionService) ListSessions(ctx context.Context, filters ports.SessionFilters) ([]*domain.Session, error) {
	return s.sessionRepo.ListWithFilters(ctx, filters)
}

// AbortActiveSessionsForProject aborts all active sessions for a project
func (s *SessionService) AbortActiveSessionsForProject(ctx context.Context, projectID domain.ProjectID) ([]domain.SessionID, error) {
	s.logger.WithField("project_id", projectID).Info("Aborting all active sessions for project")

	// Get all active sessions for the project
	filters := ports.SessionFilters{
		ProjectID: &projectID,
		Status:    &[]domain.SessionStatus{domain.SessionStatusActive}[0],
		Limit:     100, // Reasonable limit for batch abort
	}

	sessions, err := s.sessionRepo.ListWithFilters(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list active sessions: %w", err)
	}

	var abortedIDs []domain.SessionID
	for _, session := range sessions {
		if session.IsActive() {
			session.Abort("Project-wide session abort")
			if err := s.sessionRepo.Update(ctx, session); err != nil {
				s.logger.WithError(err).WithField("session_id", session.ID).Warn("Failed to abort session")
				continue
			}
			abortedIDs = append(abortedIDs, session.ID)
		}
	}

	s.logger.WithFields(logrus.Fields{
		"project_id":     projectID,
		"aborted_count":  len(abortedIDs),
	}).Info("Completed aborting active sessions")

	return abortedIDs, nil
}
