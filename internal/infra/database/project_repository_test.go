package database

import (
	"context"
	"testing"

	"github.com/joern1811/memory-bank/internal/domain"
)

func TestNewSQLiteProjectRepository(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := setupTestLogger()
	repo := NewSQLiteProjectRepository(db, logger)

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

func TestSQLiteProjectRepository_Store(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteProjectRepository(db, setupTestLogger())
	ctx := context.Background()

	// Test storing a project
	project := createTestProject()

	err := repo.Store(ctx, project)
	if err != nil {
		t.Fatalf("Failed to store project: %v", err)
	}

	// Verify the project was stored by retrieving it
	retrieved, err := repo.GetByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve stored project: %v", err)
	}

	assertProjectEqual(t, project, retrieved)
}

func TestSQLiteProjectRepository_Store_DuplicateID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteProjectRepository(db, setupTestLogger())
	ctx := context.Background()

	project := createTestProject()

	// Store project first time
	err := repo.Store(ctx, project)
	if err != nil {
		t.Fatalf("Failed to store project first time: %v", err)
	}

	// Try to store same project again (should fail)
	err = repo.Store(ctx, project)
	if err == nil {
		t.Error("Expected error when storing duplicate project ID")
	}
}

func TestSQLiteProjectRepository_GetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteProjectRepository(db, setupTestLogger())
	ctx := context.Background()

	// Test getting non-existent project
	nonExistent := domain.ProjectID("non_existent")
	_, err := repo.GetByID(ctx, nonExistent)
	if err == nil {
		t.Error("Expected error when getting non-existent project")
	}

	// Store a project and retrieve it
	project := createTestProject()
	err = repo.Store(ctx, project)
	if err != nil {
		t.Fatalf("Failed to store project: %v", err)
	}

	retrieved, err := repo.GetByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve project: %v", err)
	}

	assertProjectEqual(t, project, retrieved)
}

func TestSQLiteProjectRepository_GetByPath(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteProjectRepository(db, setupTestLogger())
	ctx := context.Background()

	// Test getting project by non-existent path
	_, err := repo.GetByPath(ctx, "/non/existent/path")
	if err == nil {
		t.Error("Expected error when getting project by non-existent path")
	}

	// Store a project and retrieve it by path
	project := createTestProject()
	err = repo.Store(ctx, project)
	if err != nil {
		t.Fatalf("Failed to store project: %v", err)
	}

	retrieved, err := repo.GetByPath(ctx, project.Path)
	if err != nil {
		t.Fatalf("Failed to retrieve project by path: %v", err)
	}

	assertProjectEqual(t, project, retrieved)
}

func TestSQLiteProjectRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteProjectRepository(db, setupTestLogger())
	ctx := context.Background()

	// Store original project
	project := createTestProject()
	err := repo.Store(ctx, project)
	if err != nil {
		t.Fatalf("Failed to store project: %v", err)
	}

	// Update project (only persisted fields)
	project.Name = "Updated Project Name"
	project.Description = "Updated description"

	err = repo.Update(ctx, project)
	if err != nil {
		t.Fatalf("Failed to update project: %v", err)
	}

	// Retrieve and verify update
	retrieved, err := repo.GetByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated project: %v", err)
	}

	assertProjectEqual(t, project, retrieved)
}

func TestSQLiteProjectRepository_Update_NonExistent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteProjectRepository(db, setupTestLogger())
	ctx := context.Background()

	// Try to update non-existent project
	project := createTestProject()
	err := repo.Update(ctx, project)
	if err == nil {
		t.Error("Expected error when updating non-existent project")
	}
}

func TestSQLiteProjectRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteProjectRepository(db, setupTestLogger())
	ctx := context.Background()

	// Store a project
	project := createTestProject()
	err := repo.Store(ctx, project)
	if err != nil {
		t.Fatalf("Failed to store project: %v", err)
	}

	// Delete the project
	err = repo.Delete(ctx, project.ID)
	if err != nil {
		t.Fatalf("Failed to delete project: %v", err)
	}

	// Verify project is deleted
	_, err = repo.GetByID(ctx, project.ID)
	if err == nil {
		t.Error("Expected error when getting deleted project")
	}
}

func TestSQLiteProjectRepository_Delete_NonExistent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteProjectRepository(db, setupTestLogger())
	ctx := context.Background()

	// Try to delete non-existent project
	nonExistent := domain.ProjectID("non_existent")
	err := repo.Delete(ctx, nonExistent)
	if err == nil {
		t.Error("Expected error when deleting non-existent project")
	}
}

func TestSQLiteProjectRepository_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSQLiteProjectRepository(db, setupTestLogger())
	ctx := context.Background()

	// Store multiple projects
	project1 := createTestProject()
	project1.Name = "Project 1"
	project1.Path = "/path/to/project1"

	project2 := createTestProject()
	project2.Name = "Project 2"
	project2.Path = "/path/to/project2"

	project3 := createTestProject()
	project3.Name = "Project 3"
	project3.Path = "/path/to/project3"

	projects := []*domain.Project{project1, project2, project3}
	for _, project := range projects {
		err := repo.Store(ctx, project)
		if err != nil {
			t.Fatalf("Failed to store project: %v", err)
		}
	}

	// Test listing all projects
	allProjects, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list projects: %v", err)
	}

	if len(allProjects) != 3 {
		t.Errorf("Expected 3 projects, got %d", len(allProjects))
	}
}

