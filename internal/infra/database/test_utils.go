package database

import (
	"database/sql"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/sirupsen/logrus"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// Use in-memory database for tests
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create test logger that outputs to nothing during tests
	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.SetLevel(logrus.ErrorLevel) // Only show errors in tests

	// Run migrations
	migrator := NewMigrator(db, logger)
	if err := migrator.Run(); err != nil {
		db.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

// setupTestLogger creates a test logger
func setupTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.SetLevel(logrus.ErrorLevel) // Only show errors in tests
	return logger
}

// createTestMemory creates a memory for testing
func createTestMemory(projectID domain.ProjectID, memoryType domain.MemoryType) *domain.Memory {
	memory := domain.NewMemory(
		projectID,
		memoryType,
		"Test Memory",
		"Test content for memory",
		"Test context",
	)
	// Override ID to ensure uniqueness
	memory.ID = domain.MemoryID(generateTestID("mem"))
	memory.AddTag("test")
	memory.AddTag("unit-test")
	return memory
}

// createTestProject creates a project for testing
func createTestProject() *domain.Project {
	project := domain.NewProject(
		"Test Project",
		"/path/to/test/project",
		"A project for unit testing",
	)
	// Override ID and path to ensure uniqueness
	project.ID = domain.ProjectID(generateTestID("proj"))
	project.Path = generateTestID("/path/to/test/project")
	return project
}

// createTestSession creates a session for testing
func createTestSession(projectID domain.ProjectID) *domain.Session {
	session := domain.NewSession(
		projectID,
		"Test Session",
		"Testing session functionality",
	)
	// Override ID to ensure uniqueness
	session.ID = domain.SessionID(generateTestID("sess"))
	session.LogInfo("Session started")
	session.LogMilestone("First milestone")
	session.AddTag("test")
	return session
}

// testIDCounter provides a counter for generating unique test IDs
var testIDCounter int64

// generateTestID generates a unique test ID
func generateTestID(prefix string) string {
	// Use atomic increment to ensure uniqueness even with rapid calls
	counter := atomic.AddInt64(&testIDCounter, 1)
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixNano(), counter)
}

// assertMemoryEqual checks if two memories are equal (ignoring timestamps)
func assertMemoryEqual(t *testing.T, expected, actual *domain.Memory) {
	if actual.ID != expected.ID {
		t.Errorf("Expected ID %s, got %s", expected.ID, actual.ID)
	}
	if actual.ProjectID != expected.ProjectID {
		t.Errorf("Expected ProjectID %s, got %s", expected.ProjectID, actual.ProjectID)
	}
	if actual.Type != expected.Type {
		t.Errorf("Expected Type %s, got %s", expected.Type, actual.Type)
	}
	if actual.Title != expected.Title {
		t.Errorf("Expected Title %s, got %s", expected.Title, actual.Title)
	}
	if actual.Content != expected.Content {
		t.Errorf("Expected Content %s, got %s", expected.Content, actual.Content)
	}
	if actual.Context != expected.Context {
		t.Errorf("Expected Context %s, got %s", expected.Context, actual.Context)
	}
	if actual.HasEmbedding != expected.HasEmbedding {
		t.Errorf("Expected HasEmbedding %t, got %t", expected.HasEmbedding, actual.HasEmbedding)
	}
	if len(actual.Tags) != len(expected.Tags) {
		t.Errorf("Expected %d tags, got %d", len(expected.Tags), len(actual.Tags))
	}
	for i, tag := range expected.Tags {
		if i >= len(actual.Tags) || actual.Tags[i] != tag {
			t.Errorf("Tag mismatch at index %d: expected %s, got %s", i, tag, actual.Tags[i])
		}
	}
}

// assertProjectEqual checks if two projects are equal (ignoring timestamps and non-persisted fields)
func assertProjectEqual(t *testing.T, expected, actual *domain.Project) {
	if actual.ID != expected.ID {
		t.Errorf("Expected ID %s, got %s", expected.ID, actual.ID)
	}
	if actual.Name != expected.Name {
		t.Errorf("Expected Name %s, got %s", expected.Name, actual.Name)
	}
	if actual.Path != expected.Path {
		t.Errorf("Expected Path %s, got %s", expected.Path, actual.Path)
	}
	if actual.Description != expected.Description {
		t.Errorf("Expected Description %s, got %s", expected.Description, actual.Description)
	}
	// Note: Language, Framework, EmbeddingProvider, VectorStore, and Config are not persisted
	// in the current schema, so we don't assert on them
}

// assertSessionEqual checks if two sessions are equal (ignoring timestamps and non-persisted fields)
func assertSessionEqual(t *testing.T, expected, actual *domain.Session) {
	if actual.ID != expected.ID {
		t.Errorf("Expected ID %s, got %s", expected.ID, actual.ID)
	}
	if actual.ProjectID != expected.ProjectID {
		t.Errorf("Expected ProjectID %s, got %s", expected.ProjectID, actual.ProjectID)
	}
	if actual.Name != expected.Name {
		t.Errorf("Expected Name %s, got %s", expected.Name, actual.Name)
	}
	// Note: TaskDescription is stored differently (combined with outcome and progress in description field)
	if actual.Status != expected.Status {
		t.Errorf("Expected Status %s, got %s", expected.Status, actual.Status)
	}
	// Note: Outcome, Summary, Tags, and Progress are stored in encoded format in description field
	// and reconstructed differently, so we don't assert exact equality here
}