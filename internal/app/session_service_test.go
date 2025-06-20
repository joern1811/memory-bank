package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/sirupsen/logrus"
)

func setupSessionServiceTest() (*SessionService, *MockSessionRepository, *MockProjectRepository) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	sessionRepo := NewMockSessionRepository()
	projectRepo := NewMockProjectRepository()
	service := NewSessionService(sessionRepo, projectRepo, logger)
	return service, sessionRepo, projectRepo
}

func TestNewSessionService(t *testing.T) {
	service, _, _ := setupSessionServiceTest()

	if service == nil {
		t.Fatal("Expected non-nil service")
	}
	if service.sessionRepo == nil {
		t.Error("Expected sessionRepo to be set")
	}
	if service.projectRepo == nil {
		t.Error("Expected projectRepo to be set")
	}
	if service.logger == nil {
		t.Error("Expected logger to be set")
	}
}

func TestSessionService_StartSession(t *testing.T) {
	service, sessionRepo, projectRepo := setupSessionServiceTest()
	ctx := context.Background()

	// Create a test project
	projectID := domain.ProjectID(generateUniqueTestID("proj"))
	testProject := &domain.Project{
		ID:          projectID,
		Name:        "Test Project",
		Path:        "/test/project",
		Description: "Test description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := projectRepo.Store(ctx, testProject)
	if err != nil {
		t.Fatalf("Failed to store test project: %v", err)
	}

	// Test starting a session
	req := ports.StartSessionRequest{
		ProjectID:       projectID,
		TaskDescription: "Implement user authentication",
	}

	session, err := service.StartSession(ctx, req)
	if err != nil {
		t.Fatalf("Failed to start session: %v", err)
	}

	// Verify session properties
	if session.ProjectID != req.ProjectID {
		t.Errorf("Expected ProjectID %s, got %s", req.ProjectID, session.ProjectID)
	}
	if session.TaskDescription != req.TaskDescription {
		t.Errorf("Expected TaskDescription %s, got %s", req.TaskDescription, session.TaskDescription)
	}
	if session.Status != domain.SessionStatusActive {
		t.Errorf("Expected Status %s, got %s", domain.SessionStatusActive, session.Status)
	}
	if session.EndTime != nil {
		t.Error("Expected EndTime to be nil for active session")
	}

	// Verify session was stored in repository
	stored, err := sessionRepo.GetByID(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve stored session: %v", err)
	}
	if stored.ID != session.ID {
		t.Error("Stored session ID mismatch")
	}
}

func TestSessionService_StartSession_ProjectNotFound(t *testing.T) {
	service, _, _ := setupSessionServiceTest()
	ctx := context.Background()

	// Test starting session with non-existent project
	nonExistentProjectID := domain.ProjectID(generateUniqueTestID("proj"))
	req := ports.StartSessionRequest{
		ProjectID:       nonExistentProjectID,
		TaskDescription: "Test task",
	}

	_, err := service.StartSession(ctx, req)
	if err == nil {
		t.Error("Expected error when starting session for non-existent project")
	}
	if !strings.Contains(err.Error(), "project not found") {
		t.Errorf("Expected 'project not found' error, got: %v", err)
	}
}

func TestSessionService_StartSession_AbortsExistingActiveSession(t *testing.T) {
	service, sessionRepo, projectRepo := setupSessionServiceTest()
	ctx := context.Background()

	// Create a test project
	projectID := domain.ProjectID(generateUniqueTestID("proj"))
	testProject := &domain.Project{
		ID:          projectID,
		Name:        "Test Project",
		Path:        "/test/project",
		Description: "Test description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := projectRepo.Store(ctx, testProject)
	if err != nil {
		t.Fatalf("Failed to store test project: %v", err)
	}

	// Create an active session directly in repository to avoid ID collision
	session1ID := domain.SessionID(generateUniqueTestID("sess"))
	session1 := &domain.Session{
		ID:              session1ID,
		ProjectID:       projectID,
		Name:            "First Session",
		TaskDescription: "First task " + generateUniqueTestID("task"),
		StartTime:       time.Now().Add(-time.Hour),
		Status:          domain.SessionStatusActive,
		Progress:        make([]domain.ProgressEntry, 0),
		Tags:            make(domain.Tags, 0),
	}

	err = sessionRepo.Store(ctx, session1)
	if err != nil {
		t.Fatalf("Failed to store first session directly: %v", err)
	}

	// Verify first session is active
	if !session1.IsActive() {
		t.Error("Expected first session to be active")
	}

	// Start second session via service (should abort first)
	req2 := ports.StartSessionRequest{
		ProjectID:       projectID,
		TaskDescription: "Second task " + generateUniqueTestID("task"),
	}

	session2, err := service.StartSession(ctx, req2)
	if err != nil {
		t.Fatalf("Failed to start second session: %v", err)
	}

	// Verify second session is active
	if !session2.IsActive() {
		t.Error("Expected second session to be active")
	}

	// Verify first session was aborted
	updated1, err := sessionRepo.GetByID(ctx, session1.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve first session: %v", err)
	}
	if updated1.Status != domain.SessionStatusAborted {
		t.Errorf("Expected first session to be aborted, got status: %s", updated1.Status)
	}
}

func TestSessionService_GetSession(t *testing.T) {
	service, sessionRepo, _ := setupSessionServiceTest()
	ctx := context.Background()

	// Create a test session directly in repository
	sessionID := domain.SessionID(generateUniqueTestID("sess"))
	projectID := domain.ProjectID(generateUniqueTestID("proj"))
	testSession := &domain.Session{
		ID:              sessionID,
		ProjectID:       projectID,
		Name:            "Test Session",
		TaskDescription: "Test task",
		StartTime:       time.Now(),
		Status:          domain.SessionStatusActive,
		Progress:        make([]domain.ProgressEntry, 0),
		Tags:            make(domain.Tags, 0),
	}

	err := sessionRepo.Store(ctx, testSession)
	if err != nil {
		t.Fatalf("Failed to store test session: %v", err)
	}

	// Test getting the session
	retrieved, err := service.GetSession(ctx, testSession.ID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if retrieved.ID != testSession.ID {
		t.Errorf("Expected ID %s, got %s", testSession.ID, retrieved.ID)
	}
	if retrieved.TaskDescription != testSession.TaskDescription {
		t.Errorf("Expected TaskDescription %s, got %s", testSession.TaskDescription, retrieved.TaskDescription)
	}
}

func TestSessionService_GetSession_NotFound(t *testing.T) {
	service, _, _ := setupSessionServiceTest()
	ctx := context.Background()

	nonExistentID := domain.SessionID(generateUniqueTestID("nonexistent"))
	_, err := service.GetSession(ctx, nonExistentID)
	if err == nil {
		t.Error("Expected error when getting non-existent session")
	}
}

func TestSessionService_GetActiveSession(t *testing.T) {
	service, sessionRepo, _ := setupSessionServiceTest()
	ctx := context.Background()

	projectID := domain.ProjectID(generateUniqueTestID("proj"))

	// Create an active session
	activeSession := &domain.Session{
		ID:              domain.SessionID(generateUniqueTestID("sess")),
		ProjectID:       projectID,
		Name:            "Active Session",
		TaskDescription: "Active task",
		StartTime:       time.Now(),
		Status:          domain.SessionStatusActive,
		Progress:        make([]domain.ProgressEntry, 0),
		Tags:            make(domain.Tags, 0),
	}

	err := sessionRepo.Store(ctx, activeSession)
	if err != nil {
		t.Fatalf("Failed to store active session: %v", err)
	}

	// Create a completed session for the same project
	completedSession := &domain.Session{
		ID:              domain.SessionID(generateUniqueTestID("sess")),
		ProjectID:       projectID,
		Name:            "Completed Session",
		TaskDescription: "Completed task",
		StartTime:       time.Now().Add(-time.Hour),
		Status:          domain.SessionStatusCompleted,
		Progress:        make([]domain.ProgressEntry, 0),
		Tags:            make(domain.Tags, 0),
	}
	endTime := time.Now().Add(-time.Minute)
	completedSession.EndTime = &endTime

	err = sessionRepo.Store(ctx, completedSession)
	if err != nil {
		t.Fatalf("Failed to store completed session: %v", err)
	}

	// Test getting active session
	retrieved, err := service.GetActiveSession(ctx, projectID)
	if err != nil {
		t.Fatalf("Failed to get active session: %v", err)
	}

	if retrieved.ID != activeSession.ID {
		t.Error("Expected to get the active session")
	}
	if !retrieved.IsActive() {
		t.Error("Retrieved session should be active")
	}
}

func TestSessionService_GetActiveSession_NotFound(t *testing.T) {
	service, _, _ := setupSessionServiceTest()
	ctx := context.Background()

	nonExistentProjectID := domain.ProjectID(generateUniqueTestID("proj"))
	_, err := service.GetActiveSession(ctx, nonExistentProjectID)
	if err == nil {
		t.Error("Expected error when getting active session for project with no sessions")
	}
}

func TestSessionService_LogProgress(t *testing.T) {
	service, sessionRepo, _ := setupSessionServiceTest()
	ctx := context.Background()

	// Create an active session
	projectID := domain.ProjectID(generateUniqueTestID("proj"))
	sessionID := domain.SessionID(generateUniqueTestID("sess"))
	testSession := &domain.Session{
		ID:              sessionID,
		ProjectID:       projectID,
		Name:            "Test Session",
		TaskDescription: "Test task",
		StartTime:       time.Now(),
		Status:          domain.SessionStatusActive,
		Progress:        make([]domain.ProgressEntry, 0),
		Tags:            make(domain.Tags, 0),
	}

	err := sessionRepo.Store(ctx, testSession)
	if err != nil {
		t.Fatalf("Failed to store test session: %v", err)
	}

	// Test logging progress
	progressMessage := "Implemented user authentication middleware"
	err = service.LogProgress(ctx, sessionID, progressMessage)
	if err != nil {
		t.Fatalf("Failed to log progress: %v", err)
	}

	// Verify progress was logged
	updated, err := sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated session: %v", err)
	}

	if len(updated.Progress) != 1 {
		t.Errorf("Expected 1 progress entry, got %d", len(updated.Progress))
	}

	if updated.Progress[0].Message != progressMessage {
		t.Errorf("Expected progress message %s, got %s", progressMessage, updated.Progress[0].Message)
	}

	if updated.Progress[0].Type != "info" {
		t.Errorf("Expected progress type 'info', got %s", updated.Progress[0].Type)
	}
}

func TestSessionService_LogProgress_SessionNotActive(t *testing.T) {
	service, sessionRepo, _ := setupSessionServiceTest()
	ctx := context.Background()

	// Create a completed session
	projectID := domain.ProjectID(generateUniqueTestID("proj"))
	sessionID := domain.SessionID(generateUniqueTestID("sess"))
	testSession := &domain.Session{
		ID:              sessionID,
		ProjectID:       projectID,
		Name:            "Completed Session",
		TaskDescription: "Completed task",
		StartTime:       time.Now().Add(-time.Hour),
		Status:          domain.SessionStatusCompleted,
		Progress:        make([]domain.ProgressEntry, 0),
		Tags:            make(domain.Tags, 0),
	}
	endTime := time.Now().Add(-time.Minute)
	testSession.EndTime = &endTime

	err := sessionRepo.Store(ctx, testSession)
	if err != nil {
		t.Fatalf("Failed to store completed session: %v", err)
	}

	// Test logging progress to inactive session
	err = service.LogProgress(ctx, sessionID, "Should fail")
	if err == nil {
		t.Error("Expected error when logging progress to inactive session")
	}
	if !strings.Contains(err.Error(), "session is not active") {
		t.Errorf("Expected 'session is not active' error, got: %v", err)
	}
}

func TestSessionService_CompleteSession(t *testing.T) {
	service, sessionRepo, _ := setupSessionServiceTest()
	ctx := context.Background()

	// Create an active session
	projectID := domain.ProjectID(generateUniqueTestID("proj"))
	sessionID := domain.SessionID(generateUniqueTestID("sess"))
	testSession := &domain.Session{
		ID:              sessionID,
		ProjectID:       projectID,
		Name:            "Test Session",
		TaskDescription: "Test task",
		StartTime:       time.Now().Add(-time.Hour), // Started an hour ago
		Status:          domain.SessionStatusActive,
		Progress:        make([]domain.ProgressEntry, 0),
		Tags:            make(domain.Tags, 0),
	}

	err := sessionRepo.Store(ctx, testSession)
	if err != nil {
		t.Fatalf("Failed to store test session: %v", err)
	}

	// Test completing the session
	outcome := "Successfully implemented user authentication"
	err = service.CompleteSession(ctx, sessionID, outcome)
	if err != nil {
		t.Fatalf("Failed to complete session: %v", err)
	}

	// Verify session was completed
	updated, err := sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated session: %v", err)
	}

	if updated.Status != domain.SessionStatusCompleted {
		t.Errorf("Expected status %s, got %s", domain.SessionStatusCompleted, updated.Status)
	}

	if updated.Outcome != outcome {
		t.Errorf("Expected outcome %s, got %s", outcome, updated.Outcome)
	}

	if updated.EndTime == nil {
		t.Error("Expected EndTime to be set for completed session")
	}

	if updated.SessionDuration == nil {
		t.Error("Expected SessionDuration to be set for completed session")
	}

	// Verify progress entry was added
	if len(updated.Progress) == 0 {
		t.Error("Expected progress entry to be added for completion")
	}

	lastProgress := updated.Progress[len(updated.Progress)-1]
	if lastProgress.Type != "milestone" {
		t.Errorf("Expected last progress type 'milestone', got %s", lastProgress.Type)
	}
}

func TestSessionService_CompleteSession_NotActive(t *testing.T) {
	service, sessionRepo, _ := setupSessionServiceTest()
	ctx := context.Background()

	// Create a completed session
	projectID := domain.ProjectID(generateUniqueTestID("proj"))
	sessionID := domain.SessionID(generateUniqueTestID("sess"))
	testSession := &domain.Session{
		ID:              sessionID,
		ProjectID:       projectID,
		Name:            "Already Completed Session",
		TaskDescription: "Already completed task",
		StartTime:       time.Now().Add(-time.Hour),
		Status:          domain.SessionStatusCompleted,
		Progress:        make([]domain.ProgressEntry, 0),
		Tags:            make(domain.Tags, 0),
	}
	endTime := time.Now().Add(-time.Minute)
	testSession.EndTime = &endTime

	err := sessionRepo.Store(ctx, testSession)
	if err != nil {
		t.Fatalf("Failed to store completed session: %v", err)
	}

	// Test completing already completed session
	err = service.CompleteSession(ctx, sessionID, "Should fail")
	if err == nil {
		t.Error("Expected error when completing already completed session")
	}
	if !strings.Contains(err.Error(), "session is not active") {
		t.Errorf("Expected 'session is not active' error, got: %v", err)
	}
}

func TestSessionService_AbortSession(t *testing.T) {
	service, sessionRepo, _ := setupSessionServiceTest()
	ctx := context.Background()

	// Create an active session
	projectID := domain.ProjectID(generateUniqueTestID("proj"))
	sessionID := domain.SessionID(generateUniqueTestID("sess"))
	testSession := &domain.Session{
		ID:              sessionID,
		ProjectID:       projectID,
		Name:            "Test Session",
		TaskDescription: "Test task",
		StartTime:       time.Now().Add(-time.Hour),
		Status:          domain.SessionStatusActive,
		Progress:        make([]domain.ProgressEntry, 0),
		Tags:            make(domain.Tags, 0),
	}

	err := sessionRepo.Store(ctx, testSession)
	if err != nil {
		t.Fatalf("Failed to store test session: %v", err)
	}

	// Test aborting the session
	err = service.AbortSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to abort session: %v", err)
	}

	// Verify session was aborted
	updated, err := sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated session: %v", err)
	}

	if updated.Status != domain.SessionStatusAborted {
		t.Errorf("Expected status %s, got %s", domain.SessionStatusAborted, updated.Status)
	}

	if updated.EndTime == nil {
		t.Error("Expected EndTime to be set for aborted session")
	}

	if updated.SessionDuration == nil {
		t.Error("Expected SessionDuration to be set for aborted session")
	}

	// Verify progress entry was added
	if len(updated.Progress) == 0 {
		t.Error("Expected progress entry to be added for abortion")
	}

	lastProgress := updated.Progress[len(updated.Progress)-1]
	if lastProgress.Type != "info" {
		t.Errorf("Expected last progress type 'info', got %s", lastProgress.Type)
	}
}

func TestSessionService_AbortSession_NotActive(t *testing.T) {
	service, sessionRepo, _ := setupSessionServiceTest()
	ctx := context.Background()

	// Create an aborted session
	projectID := domain.ProjectID(generateUniqueTestID("proj"))
	sessionID := domain.SessionID(generateUniqueTestID("sess"))
	testSession := &domain.Session{
		ID:              sessionID,
		ProjectID:       projectID,
		Name:            "Already Aborted Session",
		TaskDescription: "Already aborted task",
		StartTime:       time.Now().Add(-time.Hour),
		Status:          domain.SessionStatusAborted,
		Progress:        make([]domain.ProgressEntry, 0),
		Tags:            make(domain.Tags, 0),
	}
	endTime := time.Now().Add(-time.Minute)
	testSession.EndTime = &endTime

	err := sessionRepo.Store(ctx, testSession)
	if err != nil {
		t.Fatalf("Failed to store aborted session: %v", err)
	}

	// Test aborting already aborted session
	err = service.AbortSession(ctx, sessionID)
	if err == nil {
		t.Error("Expected error when aborting already aborted session")
	}
	if !strings.Contains(err.Error(), "session is not active") {
		t.Errorf("Expected 'session is not active' error, got: %v", err)
	}
}

func TestSessionService_ListSessions(t *testing.T) {
	service, sessionRepo, _ := setupSessionServiceTest()
	ctx := context.Background()

	projectID1 := domain.ProjectID(generateUniqueTestID("proj"))
	projectID2 := domain.ProjectID(generateUniqueTestID("proj"))

	// Create multiple sessions with collision handling
	sessions := []*domain.Session{
		{
			ID:              domain.SessionID(generateUniqueTestID("sess")),
			ProjectID:       projectID1,
			Name:            "Session 1",
			TaskDescription: "Task 1",
			StartTime:       time.Now().Add(-3 * time.Hour),
			Status:          domain.SessionStatusCompleted,
			Progress:        make([]domain.ProgressEntry, 0),
			Tags:            make(domain.Tags, 0),
		},
		{
			ID:              domain.SessionID(generateUniqueTestID("sess")),
			ProjectID:       projectID1,
			Name:            "Session 2",
			TaskDescription: "Task 2",
			StartTime:       time.Now().Add(-2 * time.Hour),
			Status:          domain.SessionStatusActive,
			Progress:        make([]domain.ProgressEntry, 0),
			Tags:            make(domain.Tags, 0),
		},
		{
			ID:              domain.SessionID(generateUniqueTestID("sess")),
			ProjectID:       projectID2,
			Name:            "Session 3",
			TaskDescription: "Task 3",
			StartTime:       time.Now().Add(-1 * time.Hour),
			Status:          domain.SessionStatusAborted,
			Progress:        make([]domain.ProgressEntry, 0),
			Tags:            make(domain.Tags, 0),
		},
	}

	// Set end times for completed/aborted sessions
	endTime1 := time.Now().Add(-2*time.Hour - 30*time.Minute)
	sessions[0].EndTime = &endTime1
	endTime3 := time.Now().Add(-30 * time.Minute)
	sessions[2].EndTime = &endTime3

	for i, session := range sessions {
		err := sessionRepo.Store(ctx, session)
		if err != nil {
			// If session already exists due to ID collision, create with unique ID
			if strings.Contains(err.Error(), "already exists") {
				session.ID = domain.SessionID(generateUniqueTestID("sess"))
				err = sessionRepo.Store(ctx, session)
				if err != nil {
					t.Fatalf("Failed to store session %d with unique ID: %v", i, err)
				}
				sessions[i] = session // Update the slice with new ID
			} else {
				t.Fatalf("Failed to store session %d: %v", i, err)
			}
		}
		time.Sleep(time.Microsecond) // Ensure different creation times
	}

	// Test listing all sessions
	allFilters := ports.SessionFilters{Limit: 10}
	allSessions, err := service.ListSessions(ctx, allFilters)
	if err != nil {
		t.Fatalf("Failed to list all sessions: %v", err)
	}

	if len(allSessions) != 3 {
		t.Errorf("Expected 3 sessions, got %d", len(allSessions))
	}

	// Verify sessions are sorted by start time (newest first)
	if len(allSessions) > 1 {
		if allSessions[0].StartTime.Before(allSessions[1].StartTime) {
			t.Error("Expected sessions to be sorted by start time (newest first)")
		}
	}

	// Test listing sessions for specific project
	projectFilters := ports.SessionFilters{
		ProjectID: &projectID1,
		Limit:     10,
	}
	projectSessions, err := service.ListSessions(ctx, projectFilters)
	if err != nil {
		t.Fatalf("Failed to list project sessions: %v", err)
	}

	// We should have at least 2 sessions for project 1 (may have more due to collision fallbacks)
	if len(projectSessions) < 2 {
		t.Errorf("Expected at least 2 sessions for project 1, got %d", len(projectSessions))
	}

	for _, session := range projectSessions {
		if session.ProjectID != projectID1 {
			t.Errorf("Expected all sessions to belong to project %s", projectID1)
		}
	}

	// Test listing sessions by status
	activeStatus := domain.SessionStatusActive
	statusFilters := ports.SessionFilters{
		Status: &activeStatus,
		Limit:  10,
	}
	activeSessions, err := service.ListSessions(ctx, statusFilters)
	if err != nil {
		t.Fatalf("Failed to list active sessions: %v", err)
	}

	if len(activeSessions) != 1 {
		t.Errorf("Expected 1 active session, got %d", len(activeSessions))
	}

	if activeSessions[0].Status != domain.SessionStatusActive {
		t.Errorf("Expected active status, got %s", activeSessions[0].Status)
	}
}

func TestSessionService_ListSessions_Empty(t *testing.T) {
	service, _, _ := setupSessionServiceTest()
	ctx := context.Background()

	// Test listing when no sessions exist
	filters := ports.SessionFilters{Limit: 10}
	sessions, err := service.ListSessions(ctx, filters)
	if err != nil {
		t.Fatalf("Failed to list empty sessions: %v", err)
	}

	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions for empty list, got %d", len(sessions))
	}
}

func TestSessionService_AbortActiveSessionsForProject(t *testing.T) {
	service, sessionRepo, _ := setupSessionServiceTest()
	ctx := context.Background()

	projectID := domain.ProjectID(generateUniqueTestID("proj"))

	// Create multiple sessions for the project with collision handling
	sessions := []*domain.Session{
		{
			ID:              domain.SessionID(generateUniqueTestID("sess")),
			ProjectID:       projectID,
			Name:            "Active Session 1",
			TaskDescription: "Active Task 1",
			StartTime:       time.Now().Add(-2 * time.Hour),
			Status:          domain.SessionStatusActive,
			Progress:        make([]domain.ProgressEntry, 0),
			Tags:            make(domain.Tags, 0),
		},
		{
			ID:              domain.SessionID(generateUniqueTestID("sess")),
			ProjectID:       projectID,
			Name:            "Active Session 2",
			TaskDescription: "Active Task 2",
			StartTime:       time.Now().Add(-1 * time.Hour),
			Status:          domain.SessionStatusActive,
			Progress:        make([]domain.ProgressEntry, 0),
			Tags:            make(domain.Tags, 0),
		},
		{
			ID:              domain.SessionID(generateUniqueTestID("sess")),
			ProjectID:       projectID,
			Name:            "Completed Session",
			TaskDescription: "Completed Task",
			StartTime:       time.Now().Add(-3 * time.Hour),
			Status:          domain.SessionStatusCompleted,
			Progress:        make([]domain.ProgressEntry, 0),
			Tags:            make(domain.Tags, 0),
		},
	}

	// Set end time for completed session
	endTime := time.Now().Add(-2*time.Hour - 30*time.Minute)
	sessions[2].EndTime = &endTime

	for i, session := range sessions {
		err := sessionRepo.Store(ctx, session)
		if err != nil {
			// If session already exists due to ID collision, create with unique ID
			if strings.Contains(err.Error(), "already exists") {
				session.ID = domain.SessionID(generateUniqueTestID("sess"))
				err = sessionRepo.Store(ctx, session)
				if err != nil {
					t.Fatalf("Failed to store session %d with unique ID: %v", i, err)
				}
				sessions[i] = session // Update the slice with new ID
			} else {
				t.Fatalf("Failed to store session %d: %v", i, err)
			}
		}
		time.Sleep(time.Microsecond) // Ensure different creation times
	}

	// Test aborting active sessions for the project
	abortedIDs, err := service.AbortActiveSessionsForProject(ctx, projectID)
	if err != nil {
		t.Fatalf("Failed to abort active sessions: %v", err)
	}

	// Should have aborted 2 active sessions
	if len(abortedIDs) != 2 {
		t.Errorf("Expected 2 aborted sessions, got %d", len(abortedIDs))
	}

	// Verify sessions were aborted
	for _, sessionID := range abortedIDs {
		session, err := sessionRepo.GetByID(ctx, sessionID)
		if err != nil {
			t.Fatalf("Failed to retrieve aborted session: %v", err)
		}

		if session.Status != domain.SessionStatusAborted {
			t.Errorf("Expected session %s to be aborted, got status: %s", sessionID, session.Status)
		}
	}

	// Verify completed session was not affected
	completedSession, err := sessionRepo.GetByID(ctx, sessions[2].ID)
	if err != nil {
		t.Fatalf("Failed to retrieve completed session: %v", err)
	}

	if completedSession.Status != domain.SessionStatusCompleted {
		t.Errorf("Expected completed session to remain completed, got status: %s", completedSession.Status)
	}
}

func TestSessionService_AbortActiveSessionsForProject_NoActiveSessions(t *testing.T) {
	service, _, _ := setupSessionServiceTest()
	ctx := context.Background()

	// Test aborting active sessions for project with no active sessions
	projectID := domain.ProjectID(generateUniqueTestID("proj"))
	abortedIDs, err := service.AbortActiveSessionsForProject(ctx, projectID)
	if err != nil {
		t.Fatalf("Failed to abort active sessions: %v", err)
	}

	if len(abortedIDs) != 0 {
		t.Errorf("Expected 0 aborted sessions, got %d", len(abortedIDs))
	}
}

func TestSessionService_Update(t *testing.T) {
	service, sessionRepo, _ := setupSessionServiceTest()
	ctx := context.Background()

	// Create a test session
	projectID := domain.ProjectID(generateUniqueTestID("proj"))
	sessionID := domain.SessionID(generateUniqueTestID("sess"))
	testSession := &domain.Session{
		ID:              sessionID,
		ProjectID:       projectID,
		Name:            "Original Session",
		TaskDescription: "Original task",
		StartTime:       time.Now(),
		Status:          domain.SessionStatusActive,
		Progress:        make([]domain.ProgressEntry, 0),
		Tags:            make(domain.Tags, 0),
	}

	err := sessionRepo.Store(ctx, testSession)
	if err != nil {
		t.Fatalf("Failed to store test session: %v", err)
	}

	// Update session
	testSession.Name = "Updated Session"
	testSession.TaskDescription = "Updated task"
	testSession.AddTag("updated")

	err = service.Update(ctx, testSession)
	if err != nil {
		t.Fatalf("Failed to update session: %v", err)
	}

	// Verify updates
	updated, err := sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated session: %v", err)
	}

	if updated.Name != "Updated Session" {
		t.Errorf("Expected Name 'Updated Session', got %s", updated.Name)
	}

	if updated.TaskDescription != "Updated task" {
		t.Errorf("Expected TaskDescription 'Updated task', got %s", updated.TaskDescription)
	}

	if len(updated.Tags) != 1 || updated.Tags[0] != "updated" {
		t.Errorf("Expected tag 'updated', got %v", updated.Tags)
	}
}
