package database

import (
	"context"
	"testing"
	"time"

	"github.com/joern1811/memory-bank/internal/domain"
)

func TestNewSQLiteMemoryRepository(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := setupTestLogger()
	repo := NewSQLiteMemoryRepository(db, logger)

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

func TestSQLiteMemoryRepository_Store(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteMemoryRepository(db, setupTestLogger())
	ctx := context.Background()

	// Test storing a memory
	memory := createTestMemory("proj_1", domain.MemoryTypeDecision)
	sessionID := domain.SessionID("sess_1")
	memory.SessionID = &sessionID
	memory.SetEmbedding()

	err := repo.Store(ctx, memory)
	if err != nil {
		t.Fatalf("Failed to store memory: %v", err)
	}

	// Verify the memory was stored by retrieving it
	retrieved, err := repo.GetByID(ctx, memory.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve stored memory: %v", err)
	}

	assertMemoryEqual(t, memory, retrieved)
}

func TestSQLiteMemoryRepository_Store_DuplicateID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteMemoryRepository(db, setupTestLogger())
	ctx := context.Background()

	memory := createTestMemory("proj_1", domain.MemoryTypePattern)

	// Store memory first time
	err := repo.Store(ctx, memory)
	if err != nil {
		t.Fatalf("Failed to store memory first time: %v", err)
	}

	// Try to store same memory again (should fail)
	err = repo.Store(ctx, memory)
	if err == nil {
		t.Error("Expected error when storing duplicate memory ID")
	}
}

func TestSQLiteMemoryRepository_GetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteMemoryRepository(db, setupTestLogger())
	ctx := context.Background()

	// Test getting non-existent memory
	nonExistent := domain.MemoryID("non_existent")
	_, err := repo.GetByID(ctx, nonExistent)
	if err == nil {
		t.Error("Expected error when getting non-existent memory")
	}

	// Store a memory and retrieve it
	memory := createTestMemory("proj_1", domain.MemoryTypeCode)
	err = repo.Store(ctx, memory)
	if err != nil {
		t.Fatalf("Failed to store memory: %v", err)
	}

	retrieved, err := repo.GetByID(ctx, memory.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve memory: %v", err)
	}

	assertMemoryEqual(t, memory, retrieved)
}

func TestSQLiteMemoryRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteMemoryRepository(db, setupTestLogger())
	ctx := context.Background()

	// Store original memory
	memory := createTestMemory("proj_1", domain.MemoryTypeDocumentation)
	err := repo.Store(ctx, memory)
	if err != nil {
		t.Fatalf("Failed to store memory: %v", err)
	}

	// Update memory
	memory.Title = "Updated Title"
	memory.Content = "Updated Content"
	memory.Context = "Updated Context"
	memory.AddTag("updated")
	memory.SetEmbedding()
	memory.UpdatedAt = time.Now()

	err = repo.Update(ctx, memory)
	if err != nil {
		t.Fatalf("Failed to update memory: %v", err)
	}

	// Retrieve and verify update
	retrieved, err := repo.GetByID(ctx, memory.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated memory: %v", err)
	}

	assertMemoryEqual(t, memory, retrieved)
}

func TestSQLiteMemoryRepository_Update_NonExistent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteMemoryRepository(db, setupTestLogger())
	ctx := context.Background()

	// Try to update non-existent memory
	memory := createTestMemory("proj_1", domain.MemoryTypePattern)
	err := repo.Update(ctx, memory)
	if err == nil {
		t.Error("Expected error when updating non-existent memory")
	}
}

func TestSQLiteMemoryRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteMemoryRepository(db, setupTestLogger())
	ctx := context.Background()

	// Store a memory
	memory := createTestMemory("proj_1", domain.MemoryTypeErrorSolution)
	err := repo.Store(ctx, memory)
	if err != nil {
		t.Fatalf("Failed to store memory: %v", err)
	}

	// Delete the memory
	err = repo.Delete(ctx, memory.ID)
	if err != nil {
		t.Fatalf("Failed to delete memory: %v", err)
	}

	// Verify memory is deleted
	_, err = repo.GetByID(ctx, memory.ID)
	if err == nil {
		t.Error("Expected error when getting deleted memory")
	}
}

func TestSQLiteMemoryRepository_Delete_NonExistent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteMemoryRepository(db, setupTestLogger())
	ctx := context.Background()

	// Try to delete non-existent memory
	nonExistent := domain.MemoryID("non_existent")
	err := repo.Delete(ctx, nonExistent)
	if err == nil {
		t.Error("Expected error when deleting non-existent memory")
	}
}

func TestSQLiteMemoryRepository_ListByProject(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteMemoryRepository(db, setupTestLogger())
	ctx := context.Background()

	projectID1 := domain.ProjectID("proj_1")
	projectID2 := domain.ProjectID("proj_2")

	// Store memories for different projects
	memory1 := createTestMemory(projectID1, domain.MemoryTypeDecision)
	memory2 := createTestMemory(projectID1, domain.MemoryTypePattern)
	memory3 := createTestMemory(projectID2, domain.MemoryTypeCode)

	memories := []*domain.Memory{memory1, memory2, memory3}
	for _, memory := range memories {
		err := repo.Store(ctx, memory)
		if err != nil {
			t.Fatalf("Failed to store memory: %v", err)
		}
	}

	// Test listing memories for project 1
	project1Memories, err := repo.ListByProject(ctx, projectID1)
	if err != nil {
		t.Fatalf("Failed to list memories for project 1: %v", err)
	}

	if len(project1Memories) != 2 {
		t.Errorf("Expected 2 memories for project 1, got %d", len(project1Memories))
	}

	// Test listing memories for project 2
	project2Memories, err := repo.ListByProject(ctx, projectID2)
	if err != nil {
		t.Fatalf("Failed to list memories for project 2: %v", err)
	}

	if len(project2Memories) != 1 {
		t.Errorf("Expected 1 memory for project 2, got %d", len(project2Memories))
	}
}

func TestSQLiteMemoryRepository_ListByType(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteMemoryRepository(db, setupTestLogger())
	ctx := context.Background()

	projectID := domain.ProjectID("proj_1")

	// Store memories of different types
	decision := createTestMemory(projectID, domain.MemoryTypeDecision)
	pattern := createTestMemory(projectID, domain.MemoryTypePattern)
	code := createTestMemory(projectID, domain.MemoryTypeCode)

	memories := []*domain.Memory{decision, pattern, code}
	for _, memory := range memories {
		err := repo.Store(ctx, memory)
		if err != nil {
			t.Fatalf("Failed to store memory: %v", err)
		}
	}

	// Test listing decisions
	decisions, err := repo.ListByType(ctx, projectID, domain.MemoryTypeDecision)
	if err != nil {
		t.Fatalf("Failed to list decisions: %v", err)
	}

	if len(decisions) != 1 {
		t.Errorf("Expected 1 decision, got %d", len(decisions))
	}
	if decisions[0].Type != domain.MemoryTypeDecision {
		t.Errorf("Expected decision type, got %s", decisions[0].Type)
	}

	// Test listing patterns
	patterns, err := repo.ListByType(ctx, projectID, domain.MemoryTypePattern)
	if err != nil {
		t.Fatalf("Failed to list patterns: %v", err)
	}

	if len(patterns) != 1 {
		t.Errorf("Expected 1 pattern, got %d", len(patterns))
	}

	// Test listing non-existent type
	errorSolutions, err := repo.ListByType(ctx, projectID, domain.MemoryTypeErrorSolution)
	if err != nil {
		t.Fatalf("Failed to list error solutions: %v", err)
	}

	if len(errorSolutions) != 0 {
		t.Errorf("Expected 0 error solutions, got %d", len(errorSolutions))
	}
}


func TestSQLiteMemoryRepository_GetByIDs(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteMemoryRepository(db, setupTestLogger())
	ctx := context.Background()

	// Store multiple memories
	memory1 := createTestMemory("proj_1", domain.MemoryTypeDecision)
	memory2 := createTestMemory("proj_1", domain.MemoryTypePattern)
	memory3 := createTestMemory("proj_1", domain.MemoryTypeCode)

	memories := []*domain.Memory{memory1, memory2, memory3}
	for _, memory := range memories {
		err := repo.Store(ctx, memory)
		if err != nil {
			t.Fatalf("Failed to store memory: %v", err)
		}
	}

	// Test getting multiple memories by IDs
	ids := []domain.MemoryID{memory1.ID, memory3.ID}
	results, err := repo.GetByIDs(ctx, ids)
	if err != nil {
		t.Fatalf("Failed to get memories by IDs: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 memories, got %d", len(results))
	}

	// Verify results contain the correct memories
	foundIDs := make(map[domain.MemoryID]bool)
	for _, memory := range results {
		foundIDs[memory.ID] = true
	}

	if !foundIDs[memory1.ID] {
		t.Error("Expected to find memory1 in results")
	}
	if !foundIDs[memory3.ID] {
		t.Error("Expected to find memory3 in results")
	}
	if foundIDs[memory2.ID] {
		t.Error("Did not expect to find memory2 in results")
	}

	// Test with empty IDs slice
	emptyResults, err := repo.GetByIDs(ctx, []domain.MemoryID{})
	if err != nil {
		t.Fatalf("Failed to get memories with empty IDs: %v", err)
	}

	if len(emptyResults) != 0 {
		t.Errorf("Expected 0 memories for empty IDs, got %d", len(emptyResults))
	}

	// Test with non-existent IDs
	nonExistentIDs := []domain.MemoryID{"non_existent_1", "non_existent_2"}
	noResults, err := repo.GetByIDs(ctx, nonExistentIDs)
	if err != nil {
		t.Fatalf("Failed to get memories with non-existent IDs: %v", err)
	}

	if len(noResults) != 0 {
		t.Errorf("Expected 0 memories for non-existent IDs, got %d", len(noResults))
	}
}

func TestSQLiteMemoryRepository_SessionHandling(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteMemoryRepository(db, setupTestLogger())
	ctx := context.Background()

	projectID := domain.ProjectID("proj_1")
	sessionID := domain.SessionID("sess_1")

	// Test memory without session
	memoryWithoutSession := createTestMemory(projectID, domain.MemoryTypeDecision)
	err := repo.Store(ctx, memoryWithoutSession)
	if err != nil {
		t.Fatalf("Failed to store memory without session: %v", err)
	}

	retrieved, err := repo.GetByID(ctx, memoryWithoutSession.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve memory without session: %v", err)
	}

	if retrieved.SessionID != nil {
		t.Error("Expected SessionID to be nil for memory without session")
	}

	// Test memory with session
	memoryWithSession := createTestMemory(projectID, domain.MemoryTypePattern)
	memoryWithSession.SessionID = &sessionID
	err = repo.Store(ctx, memoryWithSession)
	if err != nil {
		t.Fatalf("Failed to store memory with session: %v", err)
	}

	retrievedWithSession, err := repo.GetByID(ctx, memoryWithSession.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve memory with session: %v", err)
	}

	if retrievedWithSession.SessionID == nil {
		t.Error("Expected SessionID to be non-nil for memory with session")
	}
	if *retrievedWithSession.SessionID != sessionID {
		t.Errorf("Expected SessionID %s, got %s", sessionID, *retrievedWithSession.SessionID)
	}
}