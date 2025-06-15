package ports

import (
	"context"
	"github.com/joern1811/memory-bank/internal/domain"
)

// MemoryService defines the primary port for memory operations
type MemoryService interface {
	// Memory CRUD operations
	CreateMemory(ctx context.Context, req CreateMemoryRequest) (*domain.Memory, error)
	GetMemory(ctx context.Context, id domain.MemoryID) (*domain.Memory, error)
	UpdateMemory(ctx context.Context, memory *domain.Memory) error
	DeleteMemory(ctx context.Context, id domain.MemoryID) error
	
	// Search operations
	SearchMemories(ctx context.Context, query SearchQuery) ([]MemorySearchResult, error)
	FindSimilarMemories(ctx context.Context, memoryID domain.MemoryID, limit int) ([]MemorySearchResult, error)
	
	// Specialized operations
	CreateDecision(ctx context.Context, req CreateDecisionRequest) (*domain.Decision, error)
	CreatePattern(ctx context.Context, req CreatePatternRequest) (*domain.Pattern, error)
	CreateErrorSolution(ctx context.Context, req CreateErrorSolutionRequest) (*domain.ErrorSolution, error)
}

// ProjectService defines the primary port for project operations
type ProjectService interface {
	CreateProject(ctx context.Context, req CreateProjectRequest) (*domain.Project, error)
	GetProject(ctx context.Context, id domain.ProjectID) (*domain.Project, error)
	GetProjectByPath(ctx context.Context, path string) (*domain.Project, error)
	UpdateProject(ctx context.Context, project *domain.Project) error
	DeleteProject(ctx context.Context, id domain.ProjectID) error
	ListProjects(ctx context.Context) ([]*domain.Project, error)
	
	// Project initialization
	InitializeProject(ctx context.Context, path string, req InitializeProjectRequest) (*domain.Project, error)
}

// SessionService defines the primary port for session operations
type SessionService interface {
	StartSession(ctx context.Context, req StartSessionRequest) (*domain.Session, error)
	GetSession(ctx context.Context, id domain.SessionID) (*domain.Session, error)
	GetActiveSession(ctx context.Context, projectID domain.ProjectID) (*domain.Session, error)
	LogProgress(ctx context.Context, sessionID domain.SessionID, entry string) error
	CompleteSession(ctx context.Context, sessionID domain.SessionID, outcome string) error
	AbortSession(ctx context.Context, sessionID domain.SessionID) error
	ListSessions(ctx context.Context, projectID domain.ProjectID) ([]*domain.Session, error)
}

// Requests and Responses

// CreateMemoryRequest represents a request to create a memory
type CreateMemoryRequest struct {
	ProjectID domain.ProjectID  `json:"project_id"`
	SessionID *domain.SessionID `json:"session_id,omitempty"`
	Type      domain.MemoryType `json:"type"`
	Title     string            `json:"title"`
	Content   string            `json:"content"`
	Context   string            `json:"context"`
	Tags      domain.Tags       `json:"tags,omitempty"`
}

// CreateDecisionRequest represents a request to create a decision memory
type CreateDecisionRequest struct {
	CreateMemoryRequest
	Rationale string   `json:"rationale"`
	Options   []string `json:"options"`
	Outcome   string   `json:"outcome,omitempty"`
}

// CreatePatternRequest represents a request to create a pattern memory
type CreatePatternRequest struct {
	CreateMemoryRequest
	PatternType    string `json:"pattern_type"`
	Implementation string `json:"implementation"`
	UseCase        string `json:"use_case"`
	Language       string `json:"language,omitempty"`
}

// CreateErrorSolutionRequest represents a request to create an error solution memory
type CreateErrorSolutionRequest struct {
	CreateMemoryRequest
	ErrorSignature string `json:"error_signature"`
	Solution       string `json:"solution"`
	StackTrace     string `json:"stack_trace,omitempty"`
	Language       string `json:"language,omitempty"`
}

// CreateProjectRequest represents a request to create a project
type CreateProjectRequest struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description"`
	Language    string `json:"language,omitempty"`
	Framework   string `json:"framework,omitempty"`
}

// InitializeProjectRequest represents a request to initialize a project
type InitializeProjectRequest struct {
	Name                string            `json:"name"`
	Description         string            `json:"description"`
	Language            string            `json:"language,omitempty"`
	Framework           string            `json:"framework,omitempty"`
	EmbeddingProvider   string            `json:"embedding_provider,omitempty"`
	VectorStore         string            `json:"vector_store,omitempty"`
	Config              map[string]string `json:"config,omitempty"`
}

// StartSessionRequest represents a request to start a session
type StartSessionRequest struct {
	ProjectID       domain.ProjectID `json:"project_id"`
	TaskDescription string           `json:"task_description"`
}

// MemorySearchResult represents a memory with similarity score
type MemorySearchResult struct {
	Memory     *domain.Memory    `json:"memory"`
	Similarity domain.Similarity `json:"similarity"`
}

// SemanticSearchRequest represents a semantic search request
type SemanticSearchRequest struct {
	Query      string             `json:"query"`
	ProjectID  *domain.ProjectID  `json:"project_id,omitempty"`
	Type       *domain.MemoryType `json:"type,omitempty"`
	Tags       domain.Tags        `json:"tags,omitempty"`
	Limit      int                `json:"limit"`
	Threshold  float32            `json:"threshold"`
	TimeFilter *TimeFilter        `json:"time_filter,omitempty"`
}
