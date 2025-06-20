package database

import (
	"context"
	"testing"
	"time"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/ports"
)

func TestNewSQLiteSessionRepository(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := setupTestLogger()
	repo := NewSQLiteSessionRepository(db, logger)

	if repo == nil {
		t.Fatal("Expected non-nil repository")
	}
	if repo.db != db {
		t.Error("Expected repository to store database reference")
	}
	if repo.logger != logger {
		t.Error("Expected repository to store logger reference")
	}
}

func TestSQLiteSessionRepository_Store(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteSessionRepository(db, setupTestLogger())
	ctx := context.Background()

	// Test storing a session
	session := createTestSession("proj_1")
	session.AddTag("development")
	session.AddTag("feature")
	session.SetSummary("Test session summary")

	err := repo.Store(ctx, session)
	if err != nil {
		t.Fatalf("Failed to store session: %v", err)
	}

	// Verify the session was stored by retrieving it
	retrieved, err := repo.GetByID(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve stored session: %v", err)
	}

	assertSessionEqual(t, session, retrieved)
}

func TestSQLiteSessionRepository_Store_DuplicateID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteSessionRepository(db, setupTestLogger())
	ctx := context.Background()

	session := createTestSession("proj_1")

	// Store session first time
	err := repo.Store(ctx, session)
	if err != nil {
		t.Fatalf("Failed to store session first time: %v", err)
	}

	// Try to store same session again (should fail)
	err = repo.Store(ctx, session)
	if err == nil {
		t.Error("Expected error when storing duplicate session ID")
	}
}

func TestSQLiteSessionRepository_GetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteSessionRepository(db, setupTestLogger())
	ctx := context.Background()

	// Test getting non-existent session
	nonExistent := domain.SessionID("non_existent")
	_, err := repo.GetByID(ctx, nonExistent)
	if err == nil {
		t.Error("Expected error when getting non-existent session")
	}

	// Store a session and retrieve it
	session := createTestSession("proj_1")
	err = repo.Store(ctx, session)
	if err != nil {
		t.Fatalf("Failed to store session: %v", err)
	}

	retrieved, err := repo.GetByID(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve session: %v", err)
	}

	assertSessionEqual(t, session, retrieved)
}

func TestSQLiteSessionRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteSessionRepository(db, setupTestLogger())
	ctx := context.Background()

	// Store original session
	session := createTestSession("proj_1")
	err := repo.Store(ctx, session)
	if err != nil {
		t.Fatalf("Failed to store session: %v", err)
	}

	// Update session
	session.LogIssue("Encountered a bug in the authentication flow")
	session.LogSolution("Fixed by updating the JWT validation logic")
	session.AddTag("bugfix")
	session.SetSummary("Updated session summary")
	session.Complete("Successfully implemented authentication with bug fix")

	err = repo.Update(ctx, session)
	if err != nil {
		t.Fatalf("Failed to update session: %v", err)
	}

	// Retrieve and verify update
	retrieved, err := repo.GetByID(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated session: %v", err)
	}

	assertSessionEqual(t, session, retrieved)

	// Verify session is completed
	if retrieved.Status != domain.SessionStatusCompleted {
		t.Errorf("Expected session status to be completed, got %s", retrieved.Status)
	}
	if retrieved.Outcome != "Successfully implemented authentication with bug fix" {
		t.Errorf("Expected outcome to be set, got '%s'", retrieved.Outcome)
	}
	if retrieved.EndTime == nil {
		t.Error("Expected EndTime to be set for completed session")
	}
	if retrieved.SessionDuration == nil {
		t.Error("Expected SessionDuration to be set for completed session")
	}
}

func TestSQLiteSessionRepository_Update_NonExistent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteSessionRepository(db, setupTestLogger())
	ctx := context.Background()

	// Try to update non-existent session
	session := createTestSession("proj_1")
	err := repo.Update(ctx, session)
	if err == nil {
		t.Error("Expected error when updating non-existent session")
	}
}

func TestSQLiteSessionRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteSessionRepository(db, setupTestLogger())
	ctx := context.Background()

	// Store a session
	session := createTestSession("proj_1")
	err := repo.Store(ctx, session)
	if err != nil {
		t.Fatalf("Failed to store session: %v", err)
	}

	// Delete the session
	err = repo.Delete(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Verify session is deleted
	_, err = repo.GetByID(ctx, session.ID)
	if err == nil {
		t.Error("Expected error when getting deleted session")
	}
}

func TestSQLiteSessionRepository_Delete_NonExistent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteSessionRepository(db, setupTestLogger())
	ctx := context.Background()

	// Try to delete non-existent session
	nonExistent := domain.SessionID("non_existent")
	err := repo.Delete(ctx, nonExistent)
	if err == nil {
		t.Error("Expected error when deleting non-existent session")
	}
}

func TestSQLiteSessionRepository_ListByProject(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteSessionRepository(db, setupTestLogger())
	ctx := context.Background()

	projectID1 := domain.ProjectID("proj_1")
	projectID2 := domain.ProjectID("proj_2")

	// Store sessions for different projects
	session1 := createTestSession(projectID1)
	session1.Name = "Session 1"

	session2 := createTestSession(projectID1)
	session2.Name = "Session 2"

	session3 := createTestSession(projectID2)
	session3.Name = "Session 3"

	sessions := []*domain.Session{session1, session2, session3}
	for _, session := range sessions {
		err := repo.Store(ctx, session)
		if err != nil {
			t.Fatalf("Failed to store session: %v", err)
		}
	}

	// Test listing sessions for project 1
	project1Sessions, err := repo.ListByProject(ctx, projectID1)
	if err != nil {
		t.Fatalf("Failed to list sessions for project 1: %v", err)
	}

	if len(project1Sessions) != 2 {
		t.Errorf("Expected 2 sessions for project 1, got %d", len(project1Sessions))
	}

	// Test listing sessions for project 2
	project2Sessions, err := repo.ListByProject(ctx, projectID2)
	if err != nil {
		t.Fatalf("Failed to list sessions for project 2: %v", err)
	}

	if len(project2Sessions) != 1 {
		t.Errorf("Expected 1 session for project 2, got %d", len(project2Sessions))
	}
}

func TestSQLiteSessionRepository_ListWithFilters(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteSessionRepository(db, setupTestLogger())
	ctx := context.Background()

	projectID := domain.ProjectID("proj_1")

	// Create sessions with different statuses
	activeSession := createTestSession(projectID)
	activeSession.Name = "Active Session"

	pausedSession := createTestSession(projectID)
	pausedSession.Name = "Paused Session"
	pausedSession.Pause()

	completedSession := createTestSession(projectID)
	completedSession.Name = "Completed Session"
	completedSession.Complete("Task completed successfully")

	abortedSession := createTestSession(projectID)
	abortedSession.Name = "Aborted Session"
	abortedSession.Abort("User cancelled")

	sessions := []*domain.Session{activeSession, pausedSession, completedSession, abortedSession}
	for _, session := range sessions {
		err := repo.Store(ctx, session)
		if err != nil {
			t.Fatalf("Failed to store session: %v", err)
		}
	}

	// Import SessionFilters from the correct package
	activeStatus := domain.SessionStatusActive
	completedStatus := domain.SessionStatusCompleted

	// Test listing active sessions
	activeFilters := ports.SessionFilters{
		ProjectID: &projectID,
		Status:    &activeStatus,
		Limit:     10,
	}
	activeSessions, err := repo.ListWithFilters(ctx, activeFilters)
	if err != nil {
		t.Fatalf("Failed to list active sessions: %v", err)
	}

	if len(activeSessions) != 1 {
		t.Errorf("Expected 1 active session, got %d", len(activeSessions))
	}
	if activeSessions[0].Status != domain.SessionStatusActive {
		t.Errorf("Expected active status, got %s", activeSessions[0].Status)
	}

	// Test listing completed sessions
	completedFilters := ports.SessionFilters{
		ProjectID: &projectID,
		Status:    &completedStatus,
		Limit:     10,
	}
	completedSessions, err := repo.ListWithFilters(ctx, completedFilters)
	if err != nil {
		t.Fatalf("Failed to list completed sessions: %v", err)
	}

	if len(completedSessions) != 1 {
		t.Errorf("Expected 1 completed session, got %d", len(completedSessions))
	}
	if completedSessions[0].Status != domain.SessionStatusCompleted {
		t.Errorf("Expected completed status, got %s", completedSessions[0].Status)
	}

	// Test listing all sessions for project (no status filter)
	allFilters := ports.SessionFilters{
		ProjectID: &projectID,
		Limit:     10,
	}
	allSessions, err := repo.ListWithFilters(ctx, allFilters)
	if err != nil {
		t.Fatalf("Failed to list all sessions: %v", err)
	}

	if len(allSessions) != 4 {
		t.Errorf("Expected 4 sessions total, got %d", len(allSessions))
	}
}

func TestSQLiteSessionRepository_GetActiveSession(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteSessionRepository(db, setupTestLogger())
	ctx := context.Background()

	projectID := domain.ProjectID("proj_1")

	// Test when no active session exists
	activeSession, err := repo.GetActiveSession(ctx, projectID)
	if err == nil {
		t.Error("Expected error when no active session exists")
	}
	if activeSession != nil {
		t.Error("Expected nil session when no active session exists")
	}

	// Create and store multiple sessions with different statuses
	completedSession := createTestSession(projectID)
	completedSession.Complete("Completed task")

	pausedSession := createTestSession(projectID)
	pausedSession.Pause()

	// Store completed and paused sessions
	err = repo.Store(ctx, completedSession)
	if err != nil {
		t.Fatalf("Failed to store completed session: %v", err)
	}

	err = repo.Store(ctx, pausedSession)
	if err != nil {
		t.Fatalf("Failed to store paused session: %v", err)
	}

	// Still no active session
	activeSession, err = repo.GetActiveSession(ctx, projectID)
	if err == nil {
		t.Error("Expected error when no active session exists (only completed and paused)")
	}

	// Now create an active session
	newActiveSession := createTestSession(projectID)
	err = repo.Store(ctx, newActiveSession)
	if err != nil {
		t.Fatalf("Failed to store active session: %v", err)
	}

	// Should now find the active session
	activeSession, err = repo.GetActiveSession(ctx, projectID)
	if err != nil {
		t.Fatalf("Failed to get active session: %v", err)
	}

	if activeSession == nil {
		t.Fatal("Expected to find active session")
	}

	if activeSession.ID != newActiveSession.ID {
		t.Errorf("Expected active session ID %s, got %s", newActiveSession.ID, activeSession.ID)
	}
	if activeSession.Status != domain.SessionStatusActive {
		t.Errorf("Expected active status, got %s", activeSession.Status)
	}
}

func TestSQLiteSessionRepository_ProgressHandling(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteSessionRepository(db, setupTestLogger())
	ctx := context.Background()

	// Create session with detailed progress
	session := createTestSession("proj_1")
	session.LogInfo("Starting authentication implementation")
	session.LogMilestone("Created user model")
	session.LogMilestone("Implemented JWT middleware")
	session.LogIssue("Token validation failing")
	session.LogSolution("Fixed token validation by updating secret key")
	session.LogMilestone("Authentication flow completed")
	session.LogInfo("Running final tests")

	err := repo.Store(ctx, session)
	if err != nil {
		t.Fatalf("Failed to store session with progress: %v", err)
	}

	// Retrieve and verify progress is preserved
	retrieved, err := repo.GetByID(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve session with progress: %v", err)
	}

	if len(retrieved.Progress) != len(session.Progress) {
		t.Errorf("Expected %d progress entries, got %d", len(session.Progress), len(retrieved.Progress))
	}

	// Verify specific progress entries
	milestones := retrieved.GetMilestones()
	if len(milestones) != 4 {
		t.Errorf("Expected 4 milestones, got %d", len(milestones))
		for i, milestone := range milestones {
			t.Logf("Milestone %d: %s", i, milestone.Message)
		}
	}

	issues := retrieved.GetIssues()
	if len(issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(issues))
	}

	solutions := retrieved.GetSolutions()
	if len(solutions) != 1 {
		t.Errorf("Expected 1 solution, got %d", len(solutions))
	}

	// Verify progress entry content
	for i, entry := range retrieved.Progress {
		originalEntry := session.Progress[i]
		if entry.Message != originalEntry.Message {
			t.Errorf("Progress entry %d message mismatch: expected '%s', got '%s'",
				i, originalEntry.Message, entry.Message)
		}
		if entry.Type != originalEntry.Type {
			t.Errorf("Progress entry %d type mismatch: expected '%s', got '%s'",
				i, originalEntry.Type, entry.Type)
		}
		if entry.Timestamp != originalEntry.Timestamp {
			t.Errorf("Progress entry %d timestamp mismatch: expected '%s', got '%s'",
				i, originalEntry.Timestamp, entry.Timestamp)
		}
	}
}

func TestSQLiteSessionRepository_DurationHandling(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteSessionRepository(db, setupTestLogger())
	ctx := context.Background()

	// Test active session (no end time or duration)
	activeSession := createTestSession("proj_1")
	err := repo.Store(ctx, activeSession)
	if err != nil {
		t.Fatalf("Failed to store active session: %v", err)
	}

	retrievedActive, err := repo.GetByID(ctx, activeSession.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve active session: %v", err)
	}

	if retrievedActive.EndTime != nil {
		t.Error("Expected EndTime to be nil for active session")
	}
	if retrievedActive.SessionDuration != nil {
		t.Error("Expected SessionDuration to be nil for active session")
	}

	// Test completed session (with end time and duration)
	completedSession := createTestSession("proj_1")
	time.Sleep(10 * time.Millisecond) // Ensure some duration
	completedSession.Complete("Task finished")

	err = repo.Store(ctx, completedSession)
	if err != nil {
		t.Fatalf("Failed to store completed session: %v", err)
	}

	retrievedCompleted, err := repo.GetByID(ctx, completedSession.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve completed session: %v", err)
	}

	if retrievedCompleted.EndTime == nil {
		t.Error("Expected EndTime to be set for completed session")
	}
	if retrievedCompleted.SessionDuration == nil {
		t.Error("Expected SessionDuration to be set for completed session")
	}

	// Verify duration is positive
	if *retrievedCompleted.SessionDuration <= 0 {
		t.Error("Expected positive session duration")
	}

	// Verify duration matches calculation
	expectedDuration := retrievedCompleted.EndTime.Sub(retrievedCompleted.StartTime)
	if *retrievedCompleted.SessionDuration != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, *retrievedCompleted.SessionDuration)
	}
}
