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

func setupProjectServiceTest() (*ProjectService, *MockProjectRepository) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	projectRepo := NewMockProjectRepository()
	service := NewProjectService(projectRepo, logger)
	return service, projectRepo
}

func TestNewProjectService(t *testing.T) {
	service, _ := setupProjectServiceTest()

	if service == nil {
		t.Fatal("Expected non-nil service")
	}
	if service.projectRepo == nil {
		t.Error("Expected projectRepo to be set")
	}
	if service.logger == nil {
		t.Error("Expected logger to be set")
	}
}

func TestProjectService_CreateProject(t *testing.T) {
	service, projectRepo := setupProjectServiceTest()
	ctx := context.Background()

	// Test basic project creation
	req := ports.CreateProjectRequest{
		Name:        "Test Project",
		Path:        "/test/project/path",
		Description: "A test project for unit testing",
		Language:    "go",
		Framework:   "gin",
	}

	project, err := service.CreateProject(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Verify project properties
	if project.Name != req.Name {
		t.Errorf("Expected Name %s, got %s", req.Name, project.Name)
	}
	if project.Path != req.Path {
		t.Errorf("Expected Path %s, got %s", req.Path, project.Path)
	}
	if project.Description != req.Description {
		t.Errorf("Expected Description %s, got %s", req.Description, project.Description)
	}
	if project.Language != req.Language {
		t.Errorf("Expected Language %s, got %s", req.Language, project.Language)
	}
	if project.Framework != req.Framework {
		t.Errorf("Expected Framework %s, got %s", req.Framework, project.Framework)
	}

	// Verify project was stored in repository
	stored, err := projectRepo.GetByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve stored project: %v", err)
	}
	if stored.ID != project.ID {
		t.Error("Stored project ID mismatch")
	}

	// Verify project can be found by path
	byPath, err := projectRepo.GetByPath(ctx, req.Path)
	if err != nil {
		t.Fatalf("Failed to retrieve project by path: %v", err)
	}
	if byPath.ID != project.ID {
		t.Error("Project found by path ID mismatch")
	}
}

func TestProjectService_CreateProject_DuplicatePath(t *testing.T) {
	service, _ := setupProjectServiceTest()
	ctx := context.Background()

	path := "/test/duplicate/path"
	req := ports.CreateProjectRequest{
		Name:        "First Project",
		Path:        path,
		Description: "First project",
	}

	// Create first project
	_, err := service.CreateProject(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create first project: %v", err)
	}

	// Try to create second project with same path
	req2 := ports.CreateProjectRequest{
		Name:        "Second Project",
		Path:        path,
		Description: "Second project",
	}

	_, err = service.CreateProject(ctx, req2)
	if err == nil {
		t.Error("Expected error when creating project with duplicate path")
	}
	if !strings.Contains(err.Error(), "already exists at path") {
		t.Errorf("Expected duplicate path error, got: %v", err)
	}
}

func TestProjectService_GetProject(t *testing.T) {
	service, projectRepo := setupProjectServiceTest()
	ctx := context.Background()

	// Create a test project directly in repository
	projectID := domain.ProjectID(generateUniqueTestID("proj"))
	testProject := &domain.Project{
		ID:          projectID,
		Name:        "Test Project",
		Path:        "/test/project",
		Description: "Test description",
		Language:    "go",
		Framework:   "gin",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := projectRepo.Store(ctx, testProject)
	if err != nil {
		t.Fatalf("Failed to store test project: %v", err)
	}

	// Test getting the project
	retrieved, err := service.GetProject(ctx, testProject.ID)
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	if retrieved.ID != testProject.ID {
		t.Errorf("Expected ID %s, got %s", testProject.ID, retrieved.ID)
	}
	if retrieved.Name != testProject.Name {
		t.Errorf("Expected Name %s, got %s", testProject.Name, retrieved.Name)
	}
}

func TestProjectService_GetProject_NotFound(t *testing.T) {
	service, _ := setupProjectServiceTest()
	ctx := context.Background()

	nonExistentID := domain.ProjectID(generateUniqueTestID("nonexistent"))
	_, err := service.GetProject(ctx, nonExistentID)
	if err == nil {
		t.Error("Expected error when getting non-existent project")
	}
}

func TestProjectService_GetByPath(t *testing.T) {
	service, _ := setupProjectServiceTest()
	ctx := context.Background()

	// Create a project
	req := ports.CreateProjectRequest{
		Name:        "Path Test Project",
		Path:        "/path/test/project",
		Description: "Testing path lookup",
	}

	created, err := service.CreateProject(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Test getting by path
	retrieved, err := service.GetByPath(ctx, req.Path)
	if err != nil {
		t.Fatalf("Failed to get project by path: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Error("Retrieved project ID mismatch")
	}
}

func TestProjectService_GetProjectByPath_Deprecated(t *testing.T) {
	service, _ := setupProjectServiceTest()
	ctx := context.Background()

	// Create a project
	req := ports.CreateProjectRequest{
		Name:        "Deprecated Test Project",
		Path:        "/deprecated/test/project",
		Description: "Testing deprecated alias",
	}

	created, err := service.CreateProject(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Test deprecated alias
	retrieved, err := service.GetProjectByPath(ctx, req.Path)
	if err != nil {
		t.Fatalf("Failed to get project by deprecated method: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Error("Retrieved project ID mismatch using deprecated method")
	}
}

func TestProjectService_GetByPath_NotFound(t *testing.T) {
	service, _ := setupProjectServiceTest()
	ctx := context.Background()

	_, err := service.GetByPath(ctx, "/nonexistent/path")
	if err == nil {
		t.Error("Expected error when getting project by non-existent path")
	}
}

func TestProjectService_UpdateProject(t *testing.T) {
	service, projectRepo := setupProjectServiceTest()
	ctx := context.Background()

	// Create initial project
	req := ports.CreateProjectRequest{
		Name:        "Original Project",
		Path:        "/original/path",
		Description: "Original description",
		Language:    "go",
	}

	project, err := service.CreateProject(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create initial project: %v", err)
	}

	originalUpdatedAt := project.UpdatedAt

	// Wait a bit to ensure updated time changes
	time.Sleep(time.Millisecond)

	// Update project
	project.Name = "Updated Project"
	project.Description = "Updated description"
	project.Language = "javascript"
	project.Framework = "react"

	err = service.UpdateProject(ctx, project)
	if err != nil {
		t.Fatalf("Failed to update project: %v", err)
	}

	// Retrieve updated project to verify changes
	updatedProject, err := projectRepo.GetByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated project: %v", err)
	}

	// Verify updates
	if updatedProject.Name != "Updated Project" {
		t.Errorf("Expected Name 'Updated Project', got %s", updatedProject.Name)
	}
	if updatedProject.Description != "Updated description" {
		t.Errorf("Expected Description 'Updated description', got %s", updatedProject.Description)
	}
	if updatedProject.Language != "javascript" {
		t.Errorf("Expected Language 'javascript', got %s", updatedProject.Language)
	}
	if updatedProject.Framework != "react" {
		t.Errorf("Expected Framework 'react', got %s", updatedProject.Framework)
	}
	if !updatedProject.UpdatedAt.After(originalUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestProjectService_UpdateProject_NotFound(t *testing.T) {
	service, _ := setupProjectServiceTest()
	ctx := context.Background()

	// Create a non-existent project object
	nonExistentID := domain.ProjectID(generateUniqueTestID("nonexistent"))
	nonExistentProject := &domain.Project{
		ID:          nonExistentID,
		Name:        "Non-existent Project",
		Path:        "/nonexistent/path",
		Description: "This project doesn't exist",
	}

	err := service.UpdateProject(ctx, nonExistentProject)
	if err == nil {
		t.Error("Expected error when updating non-existent project")
	}
}

func TestProjectService_DeleteProject(t *testing.T) {
	service, projectRepo := setupProjectServiceTest()
	ctx := context.Background()

	// Create a project
	req := ports.CreateProjectRequest{
		Name:        "Delete Test Project",
		Path:        "/delete/test/project",
		Description: "Testing project deletion",
	}

	project, err := service.CreateProject(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Delete the project
	err = service.DeleteProject(ctx, project.ID)
	if err != nil {
		t.Fatalf("Failed to delete project: %v", err)
	}

	// Verify project was deleted from repository
	_, err = projectRepo.GetByID(ctx, project.ID)
	if err == nil {
		t.Error("Expected error when getting deleted project from repository")
	}

	// Verify project cannot be found by path
	_, err = projectRepo.GetByPath(ctx, req.Path)
	if err == nil {
		t.Error("Expected error when getting deleted project by path")
	}
}

func TestProjectService_DeleteProject_NotFound(t *testing.T) {
	service, _ := setupProjectServiceTest()
	ctx := context.Background()

	nonExistentID := domain.ProjectID(generateUniqueTestID("nonexistent"))
	err := service.DeleteProject(ctx, nonExistentID)
	if err == nil {
		t.Error("Expected error when deleting non-existent project")
	}
}

func TestProjectService_ListProjects(t *testing.T) {
	service, projectRepo := setupProjectServiceTest()
	ctx := context.Background()

	// Create multiple projects
	projects := []ports.CreateProjectRequest{
		{
			Name:        "Project 1 " + generateUniqueTestID("name"),
			Path:        "/project/1/" + generateUniqueTestID("path"),
			Description: "First project",
			Language:    "go",
		},
		{
			Name:        "Project 2 " + generateUniqueTestID("name"),
			Path:        "/project/2/" + generateUniqueTestID("path"),
			Description: "Second project",
			Language:    "javascript",
		},
		{
			Name:        "Project 3 " + generateUniqueTestID("name"),
			Path:        "/project/3/" + generateUniqueTestID("path"),
			Description: "Third project",
			Language:    "python",
		},
	}

	var createdProjects []*domain.Project
	for i, req := range projects {
		project, err := service.CreateProject(ctx, req)
		if err != nil {
			// If project already exists due to ID collision, create manually with unique ID
			if strings.Contains(err.Error(), "already exists") {
				uniqueID := generateUniqueTestID("proj")
				project := domain.NewProject(req.Name, req.Path, req.Description)
				project.ID = domain.ProjectID(uniqueID)
				project.Language = req.Language

				err = projectRepo.Store(ctx, project)
				if err != nil {
					t.Fatalf("Failed to store project %d manually: %v", i, err)
				}
				createdProjects = append(createdProjects, project)
			} else {
				t.Fatalf("Failed to create project %d: %v", i, err)
			}
		} else {
			createdProjects = append(createdProjects, project)
		}
		time.Sleep(time.Microsecond) // Ensure different creation times
	}

	// Test listing projects
	listedProjects, err := service.ListProjects(ctx)
	if err != nil {
		t.Fatalf("Failed to list projects: %v", err)
	}

	if len(listedProjects) != 3 {
		t.Errorf("Expected 3 projects, got %d", len(listedProjects))
	}

	// Verify projects are sorted by creation time (newest first)
	if len(listedProjects) > 1 {
		if listedProjects[0].CreatedAt.Before(listedProjects[1].CreatedAt) {
			t.Error("Expected projects to be sorted by creation time (newest first)")
		}
	}

	// Verify all created projects are in the list
	foundIDs := make(map[domain.ProjectID]bool)
	for _, project := range listedProjects {
		foundIDs[project.ID] = true
	}

	for _, created := range createdProjects {
		if !foundIDs[created.ID] {
			t.Errorf("Created project %s not found in list", created.ID)
		}
	}
}

func TestProjectService_ListProjects_Empty(t *testing.T) {
	service, _ := setupProjectServiceTest()
	ctx := context.Background()

	// Test listing when no projects exist
	projects, err := service.ListProjects(ctx)
	if err != nil {
		t.Fatalf("Failed to list empty projects: %v", err)
	}

	if len(projects) != 0 {
		t.Errorf("Expected 0 projects for empty list, got %d", len(projects))
	}
}

func TestProjectService_InitializeProject(t *testing.T) {
	service, _ := setupProjectServiceTest()
	ctx := context.Background()

	// Test project initialization
	path := "/test/init/project"
	req := ports.InitializeProjectRequest{
		Name:              "Initialized Project",
		Description:       "A project created via initialization",
		Language:          "go",
		Framework:         "gin",
		EmbeddingProvider: "ollama",
		VectorStore:       "chromadb",
		Config: map[string]string{
			"theme":       "dark",
			"debug_mode":  "true",
			"max_results": "50",
		},
	}

	project, err := service.InitializeProject(ctx, path, req)
	if err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Verify project properties
	if project.Name != req.Name {
		t.Errorf("Expected Name %s, got %s", req.Name, project.Name)
	}
	if project.Description != req.Description {
		t.Errorf("Expected Description %s, got %s", req.Description, project.Description)
	}
	if project.Language != req.Language {
		t.Errorf("Expected Language %s, got %s", req.Language, project.Language)
	}
	if project.Framework != req.Framework {
		t.Errorf("Expected Framework %s, got %s", req.Framework, project.Framework)
	}
	if project.EmbeddingProvider != req.EmbeddingProvider {
		t.Errorf("Expected EmbeddingProvider %s, got %s", req.EmbeddingProvider, project.EmbeddingProvider)
	}
	if project.VectorStore != req.VectorStore {
		t.Errorf("Expected VectorStore %s, got %s", req.VectorStore, project.VectorStore)
	}

	// Verify config was set
	for key, expectedValue := range req.Config {
		actualValue, exists := project.GetConfig(key)
		if !exists {
			t.Errorf("Expected config key %s to exist", key)
		}
		if actualValue != expectedValue {
			t.Errorf("Expected config[%s] = %s, got %s", key, expectedValue, actualValue)
		}
	}
}

func TestProjectService_InitializeProject_ExistingProject(t *testing.T) {
	service, _ := setupProjectServiceTest()
	ctx := context.Background()

	path := "/test/existing/project"

	// Create initial project
	req1 := ports.CreateProjectRequest{
		Name:        "Existing Project",
		Path:        path,
		Description: "Already exists",
	}

	existing, err := service.CreateProject(ctx, req1)
	if err != nil {
		t.Fatalf("Failed to create existing project: %v", err)
	}

	// Try to initialize project at same path
	req2 := ports.InitializeProjectRequest{
		Name:        "New Project",
		Description: "Should not overwrite existing",
	}

	initialized, err := service.InitializeProject(ctx, path, req2)
	if err != nil {
		t.Fatalf("Failed to initialize existing project: %v", err)
	}

	// Should return the existing project, not create a new one
	if initialized.ID != existing.ID {
		t.Error("Expected to return existing project, not create new one")
	}
	if initialized.Name != existing.Name {
		t.Error("Expected existing project name to be preserved")
	}
}

func TestProjectService_InitializeProject_AutoDetection(t *testing.T) {
	service, _ := setupProjectServiceTest()
	ctx := context.Background()

	// Test with minimal request (should trigger auto-detection)
	path := "/test/auto/detection"
	req := ports.InitializeProjectRequest{
		// Name and Language will be auto-detected
		Description: "Testing auto-detection",
	}

	project, err := service.InitializeProject(ctx, path, req)
	if err != nil {
		t.Fatalf("Failed to initialize project with auto-detection: %v", err)
	}

	// Verify auto-detected name (should be basename of path)
	expectedName := "detection"
	if project.Name != expectedName {
		t.Errorf("Expected auto-detected name %s, got %s", expectedName, project.Name)
	}

	// Verify auto-detected language (since fileExists returns false, should be "unknown")
	expectedLanguage := "unknown"
	if project.Language != expectedLanguage {
		t.Errorf("Expected auto-detected language %s, got %s", expectedLanguage, project.Language)
	}

	// Framework should be empty since language is unknown
	if project.Framework != "" {
		t.Errorf("Expected empty framework for unknown language, got %s", project.Framework)
	}
}

func TestProjectService_DetectLanguage(t *testing.T) {
	service, _ := setupProjectServiceTest()

	// Since fileExists always returns false in the current implementation,
	// detectLanguage should always return "unknown"
	language := service.detectLanguage("/any/path")
	if language != "unknown" {
		t.Errorf("Expected language 'unknown', got %s", language)
	}
}

func TestProjectService_DetectFramework(t *testing.T) {
	service, _ := setupProjectServiceTest()

	// Since fileExists always returns false in the current implementation,
	// detectFramework should always return empty string
	framework := service.detectFramework("/any/path", "go")
	if framework != "" {
		t.Errorf("Expected empty framework, got %s", framework)
	}

	framework = service.detectFramework("/any/path", "javascript")
	if framework != "" {
		t.Errorf("Expected empty framework, got %s", framework)
	}

	framework = service.detectFramework("/any/path", "unknown")
	if framework != "" {
		t.Errorf("Expected empty framework, got %s", framework)
	}
}
