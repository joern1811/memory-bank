package ports

import (
	"context"
	"github.com/joern1811/memory-bank/internal/domain"
)

// MemoryRepository defines the interface for memory storage
type MemoryRepository interface {
	// Basic CRUD operations
	Store(ctx context.Context, memory *domain.Memory) error
	GetByID(ctx context.Context, id domain.MemoryID) (*domain.Memory, error)
	Update(ctx context.Context, memory *domain.Memory) error
	Delete(ctx context.Context, id domain.MemoryID) error
	
	// Batch operations for performance
	GetByIDs(ctx context.Context, ids []domain.MemoryID) ([]*domain.Memory, error)
	GetMetadataByIDs(ctx context.Context, ids []domain.MemoryID) ([]*MemoryMetadata, error)
	
	// Query operations
	ListByProject(ctx context.Context, projectID domain.ProjectID) ([]*domain.Memory, error)
	ListByType(ctx context.Context, projectID domain.ProjectID, memoryType domain.MemoryType) ([]*domain.Memory, error)
	ListByTags(ctx context.Context, projectID domain.ProjectID, tags domain.Tags) ([]*domain.Memory, error)
	
	// Session-related operations
	ListBySession(ctx context.Context, sessionID domain.SessionID) ([]*domain.Memory, error)
}

// MemoryMetadata represents lightweight memory metadata for efficient queries
type MemoryMetadata struct {
	ID        domain.MemoryID    `json:"id"`
	ProjectID domain.ProjectID   `json:"project_id"`
	Type      domain.MemoryType  `json:"type"`
	Title     string             `json:"title"`
	Tags      domain.Tags        `json:"tags"`
	CreatedAt string             `json:"created_at"`
}

// ProjectRepository defines the interface for project storage
type ProjectRepository interface {
	Store(ctx context.Context, project *domain.Project) error
	GetByID(ctx context.Context, id domain.ProjectID) (*domain.Project, error)
	GetByPath(ctx context.Context, path string) (*domain.Project, error)
	Update(ctx context.Context, project *domain.Project) error
	Delete(ctx context.Context, id domain.ProjectID) error
	List(ctx context.Context) ([]*domain.Project, error)
}

// SessionRepository defines the interface for session storage
type SessionRepository interface {
	Store(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, id domain.SessionID) (*domain.Session, error)
	Update(ctx context.Context, session *domain.Session) error
	Delete(ctx context.Context, id domain.SessionID) error
	
	// Query operations
	ListByProject(ctx context.Context, projectID domain.ProjectID) ([]*domain.Session, error)
	ListWithFilters(ctx context.Context, filters SessionFilters) ([]*domain.Session, error)
	GetActiveSession(ctx context.Context, projectID domain.ProjectID) (*domain.Session, error)
}

// EmbeddingProvider defines the interface for generating embeddings
type EmbeddingProvider interface {
	GenerateEmbedding(ctx context.Context, text string) (domain.EmbeddingVector, error)
	GenerateBatchEmbeddings(ctx context.Context, texts []string) ([]domain.EmbeddingVector, error)
	GetDimensions() int
	GetModelName() string
}

// VectorStore defines the interface for vector storage and similarity search
type VectorStore interface {
	// Store operations
	Store(ctx context.Context, id string, vector domain.EmbeddingVector, metadata map[string]interface{}) error
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, id string, vector domain.EmbeddingVector, metadata map[string]interface{}) error
	
	// Batch operations for performance
	BatchStore(ctx context.Context, items []BatchStoreItem) error
	BatchDelete(ctx context.Context, ids []string) error
	
	// Search operations
	Search(ctx context.Context, vector domain.EmbeddingVector, limit int, threshold float32) ([]SearchResult, error)
	SearchByText(ctx context.Context, text string, limit int, threshold float32) ([]SearchResult, error)
	
	// Management operations
	CreateCollection(ctx context.Context, name string) error
	DeleteCollection(ctx context.Context, name string) error
	ListCollections(ctx context.Context) ([]string, error)
}

// BatchStoreItem represents an item for batch storage operations
type BatchStoreItem struct {
	ID       string                 `json:"id"`
	Vector   domain.EmbeddingVector `json:"vector"`
	Metadata map[string]interface{} `json:"metadata"`
}

// SearchResult represents a result from vector similarity search
type SearchResult struct {
	ID         string                 `json:"id"`
	Similarity domain.Similarity     `json:"similarity"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// SearchQuery represents a search query with filters
type SearchQuery struct {
	Text         string                `json:"text"`
	ProjectID    *domain.ProjectID     `json:"project_id,omitempty"`
	MemoryType   *domain.MemoryType    `json:"memory_type,omitempty"`
	Tags         domain.Tags           `json:"tags,omitempty"`
	Limit        int                   `json:"limit"`
	Threshold    float32               `json:"threshold"`
	TimeFilter   *TimeFilter           `json:"time_filter,omitempty"`
}

// TimeFilter represents time-based filtering options
type TimeFilter struct {
	After  *string `json:"after,omitempty"`  // ISO 8601 format
	Before *string `json:"before,omitempty"` // ISO 8601 format
}

// MemoryFilters represents filters for memory queries
type MemoryFilters struct {
	ProjectID  *domain.ProjectID  `json:"project_id,omitempty"`
	Type       *domain.MemoryType `json:"type,omitempty"`
	Tags       domain.Tags        `json:"tags,omitempty"`
	SessionID  *domain.SessionID  `json:"session_id,omitempty"`
	TimeFilter *TimeFilter        `json:"time_filter,omitempty"`
}
