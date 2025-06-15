package app

import (
	"context"
	"fmt"
	"path/filepath"
	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/sirupsen/logrus"
)

// ProjectService implements the project service use cases
type ProjectService struct {
	projectRepo ports.ProjectRepository
	logger      *logrus.Logger
}

// NewProjectService creates a new project service
func NewProjectService(
	projectRepo ports.ProjectRepository,
	logger *logrus.Logger,
) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		logger:      logger,
	}
}

// CreateProject creates a new project
func (s *ProjectService) CreateProject(ctx context.Context, req ports.CreateProjectRequest) (*domain.Project, error) {
	s.logger.WithFields(logrus.Fields{
		"name": req.Name,
		"path": req.Path,
	}).Info("Creating project")

	// Check if project already exists by path
	existing, err := s.projectRepo.GetByPath(ctx, req.Path)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("project already exists at path: %s", req.Path)
	}

	// Create project
	project := domain.NewProject(req.Name, req.Path, req.Description)
	project.Language = req.Language
	project.Framework = req.Framework

	// Store project
	if err := s.projectRepo.Store(ctx, project); err != nil {
		s.logger.WithError(err).Error("Failed to store project")
		return nil, fmt.Errorf("failed to store project: %w", err)
	}

	s.logger.WithField("project_id", project.ID).Info("Project created successfully")
	return project, nil
}

// GetProject retrieves a project by ID
func (s *ProjectService) GetProject(ctx context.Context, id domain.ProjectID) (*domain.Project, error) {
	return s.projectRepo.GetByID(ctx, id)
}

// GetProjectByPath retrieves a project by path
func (s *ProjectService) GetProjectByPath(ctx context.Context, path string) (*domain.Project, error) {
	// Normalize path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	return s.projectRepo.GetByPath(ctx, absPath)
}

// UpdateProject updates an existing project
func (s *ProjectService) UpdateProject(ctx context.Context, project *domain.Project) error {
	s.logger.WithField("project_id", project.ID).Info("Updating project")

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	return nil
}

// DeleteProject deletes a project
func (s *ProjectService) DeleteProject(ctx context.Context, id domain.ProjectID) error {
	s.logger.WithField("project_id", id).Info("Deleting project")

	if err := s.projectRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// ListProjects lists all projects
func (s *ProjectService) ListProjects(ctx context.Context) ([]*domain.Project, error) {
	return s.projectRepo.List(ctx)
}

// InitializeProject initializes a new project from a path
func (s *ProjectService) InitializeProject(ctx context.Context, path string, req ports.InitializeProjectRequest) (*domain.Project, error) {
	s.logger.WithFields(logrus.Fields{
		"path": path,
		"name": req.Name,
	}).Info("Initializing project")

	// Normalize path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if project already exists
	existing, err := s.projectRepo.GetByPath(ctx, absPath)
	if err == nil && existing != nil {
		s.logger.WithField("project_id", existing.ID).Info("Project already exists, returning existing")
		return existing, nil
	}

	// Auto-detect project details if not provided
	if req.Name == "" {
		req.Name = filepath.Base(absPath)
	}

	if req.Language == "" {
		req.Language = s.detectLanguage(absPath)
	}

	if req.Framework == "" {
		req.Framework = s.detectFramework(absPath, req.Language)
	}

	// Create project
	project := domain.NewProject(req.Name, absPath, req.Description)
	project.Language = req.Language
	project.Framework = req.Framework

	// Set configuration
	if req.EmbeddingProvider != "" {
		project.EmbeddingProvider = req.EmbeddingProvider
	}
	if req.VectorStore != "" {
		project.VectorStore = req.VectorStore
	}

	// Set custom config
	for key, value := range req.Config {
		project.SetConfig(key, value)
	}

	// Store project
	if err := s.projectRepo.Store(ctx, project); err != nil {
		s.logger.WithError(err).Error("Failed to store initialized project")
		return nil, fmt.Errorf("failed to store project: %w", err)
	}

	s.logger.WithField("project_id", project.ID).Info("Project initialized successfully")
	return project, nil
}

// detectLanguage attempts to detect the primary language of a project
func (s *ProjectService) detectLanguage(path string) string {
	// Simple heuristics based on common files
	commonFiles := map[string]string{
		"go.mod":         "go",
		"package.json":   "javascript",
		"pom.xml":        "java",
		"build.gradle":   "java",
		"Cargo.toml":     "rust",
		"requirements.txt": "python",
		"setup.py":       "python",
		"Gemfile":        "ruby",
		"composer.json":  "php",
	}

	for file, language := range commonFiles {
		if s.fileExists(filepath.Join(path, file)) {
			return language
		}
	}

	return "unknown"
}

// detectFramework attempts to detect the framework used in a project
func (s *ProjectService) detectFramework(path, language string) string {
	switch language {
	case "go":
		if s.fileExists(filepath.Join(path, "go.mod")) {
			// Could check go.mod content for specific frameworks
			return "gin" // Default assumption
		}
	case "javascript":
		if s.fileExists(filepath.Join(path, "package.json")) {
			// Could parse package.json for framework detection
			return "react" // Default assumption
		}
	case "java":
		if s.fileExists(filepath.Join(path, "pom.xml")) {
			return "spring-boot" // Default assumption for Maven projects
		}
		if s.fileExists(filepath.Join(path, "build.gradle")) {
			return "spring-boot" // Default assumption for Gradle projects
		}
	case "python":
		if s.fileExists(filepath.Join(path, "manage.py")) {
			return "django"
		}
		if s.fileExists(filepath.Join(path, "app.py")) {
			return "flask"
		}
	}

	return ""
}

// fileExists checks if a file exists (simplified for now)
func (s *ProjectService) fileExists(path string) bool {
	// TODO: Implement proper file existence check
	// For now, just return false to avoid external dependencies
	return false
}
