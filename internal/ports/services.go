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
	SearchMemories(ctx context.Context, query SemanticSearchRequest) ([]MemorySearchResult, error)
	FacetedSearch(ctx context.Context, req FacetedSearchRequest) (*FacetedSearchResponse, error)
	FindSimilarMemories(ctx context.Context, memoryID domain.MemoryID, limit int) ([]MemorySearchResult, error)
	ListMemories(ctx context.Context, req ListMemoriesRequest) ([]*domain.Memory, error)

	// Advanced search operations
	SearchWithRelevanceScoring(ctx context.Context, query SemanticSearchRequest) ([]EnhancedMemorySearchResult, error)
	GetSearchSuggestions(ctx context.Context, partialQuery string, projectID *domain.ProjectID) ([]string, error)

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
	ListSessions(ctx context.Context, filters SessionFilters) ([]*domain.Session, error)
	AbortActiveSessionsForProject(ctx context.Context, projectID domain.ProjectID) ([]domain.SessionID, error)
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
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	Language          string            `json:"language,omitempty"`
	Framework         string            `json:"framework,omitempty"`
	EmbeddingProvider string            `json:"embedding_provider,omitempty"`
	VectorStore       string            `json:"vector_store,omitempty"`
	Config            map[string]string `json:"config,omitempty"`
}

// StartSessionRequest represents a request to start a session
type StartSessionRequest struct {
	ProjectID       domain.ProjectID `json:"project_id"`
	TaskDescription string           `json:"task_description"`
}

// SessionFilters represents filters for listing sessions
type SessionFilters struct {
	ProjectID *domain.ProjectID     `json:"project_id,omitempty"`
	Status    *domain.SessionStatus `json:"status,omitempty"`
	Limit     int                   `json:"limit"`
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

// ListMemoriesRequest represents a request to list memories
type ListMemoriesRequest struct {
	ProjectID *domain.ProjectID  `json:"project_id,omitempty"`
	Type      *domain.MemoryType `json:"type,omitempty"`
	Tags      domain.Tags        `json:"tags,omitempty"`
	Limit     int                `json:"limit"`
}

// FacetedSearchRequest represents an advanced search with faceting
type FacetedSearchRequest struct {
	Query         string            `json:"query"`
	ProjectID     *domain.ProjectID `json:"project_id,omitempty"`
	Filters       *SearchFilters    `json:"filters,omitempty"`
	Limit         int               `json:"limit"`
	Threshold     float32           `json:"threshold"`
	IncludeFacets bool              `json:"include_facets"`
	SortBy        *SortOption       `json:"sort_by,omitempty"`
}

// SearchFilters represents comprehensive search filters
type SearchFilters struct {
	Types      []domain.MemoryType `json:"types,omitempty"`
	Tags       domain.Tags         `json:"tags,omitempty"`
	SessionIDs []domain.SessionID  `json:"session_ids,omitempty"`
	TimeFilter *TimeFilter         `json:"time_filter,omitempty"`
	HasContent bool                `json:"has_content,omitempty"`
	MinLength  *int                `json:"min_length,omitempty"`
	MaxLength  *int                `json:"max_length,omitempty"`
}

// SortOption represents sorting options for search results
type SortOption struct {
	Field     SortField     `json:"field"`
	Direction SortDirection `json:"direction"`
}

type SortField string

const (
	SortByRelevance SortField = "relevance"
	SortByCreatedAt SortField = "created_at"
	SortByUpdatedAt SortField = "updated_at"
	SortByTitle     SortField = "title"
	SortByType      SortField = "type"
)

type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

// FacetedSearchResponse represents search results with facets
type FacetedSearchResponse struct {
	Results []MemorySearchResult `json:"results"`
	Facets  *SearchFacets        `json:"facets,omitempty"`
	Total   int                  `json:"total"`
}

// SearchFacets represents faceted search results
type SearchFacets struct {
	Types       []TypeFacet       `json:"types,omitempty"`
	Tags        []TagFacet        `json:"tags,omitempty"`
	Projects    []ProjectFacet    `json:"projects,omitempty"`
	Sessions    []SessionFacet    `json:"sessions,omitempty"`
	TimePeriods []TimePeriodFacet `json:"time_periods,omitempty"`
}

// TypeFacet represents a memory type facet
type TypeFacet struct {
	Type  domain.MemoryType `json:"type"`
	Count int               `json:"count"`
}

// TagFacet represents a tag facet
type TagFacet struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// ProjectFacet represents a project facet
type ProjectFacet struct {
	ProjectID   domain.ProjectID `json:"project_id"`
	ProjectName string           `json:"project_name"`
	Count       int              `json:"count"`
}

// SessionFacet represents a session facet
type SessionFacet struct {
	SessionID    domain.SessionID `json:"session_id"`
	SessionTitle string           `json:"session_title"`
	Count        int              `json:"count"`
}

// TimePeriodFacet represents a time period facet
type TimePeriodFacet struct {
	Period string `json:"period"`
	Count  int    `json:"count"`
}

// EnhancedMemorySearchResult represents a memory with enhanced relevance scoring
type EnhancedMemorySearchResult struct {
	Memory         *domain.Memory    `json:"memory"`
	Similarity     domain.Similarity `json:"similarity"`
	RelevanceScore float64           `json:"relevance_score"`
	MatchReasons   []string          `json:"match_reasons"`
	Highlights     []string          `json:"highlights"`
}

// SearchSuggestion represents a search suggestion
type SearchSuggestion struct {
	Query     string  `json:"query"`
	Frequency int     `json:"frequency"`
	Relevance float64 `json:"relevance"`
	Type      string  `json:"type"` // "tag", "title", "content", "type"
}
