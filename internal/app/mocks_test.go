package app

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/ports"
)

// MockMemoryRepository is a mock implementation of MemoryRepository
type MockMemoryRepository struct {
	mu       sync.RWMutex
	memories map[domain.MemoryID]*domain.Memory
	metadata map[domain.MemoryID]*ports.MemoryMetadata
}

func NewMockMemoryRepository() *MockMemoryRepository {
	return &MockMemoryRepository{
		memories: make(map[domain.MemoryID]*domain.Memory),
		metadata: make(map[domain.MemoryID]*ports.MemoryMetadata),
	}
}

func (m *MockMemoryRepository) Store(ctx context.Context, memory *domain.Memory) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.memories[memory.ID]; exists {
		return fmt.Errorf("memory with ID %s already exists", memory.ID)
	}

	m.memories[memory.ID] = memory
	m.metadata[memory.ID] = &ports.MemoryMetadata{
		ID:        memory.ID,
		ProjectID: memory.ProjectID,
		Type:      memory.Type,
		Title:     memory.Title,
		Tags:      memory.Tags,
		CreatedAt: memory.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	return nil
}

func (m *MockMemoryRepository) GetByID(ctx context.Context, id domain.MemoryID) (*domain.Memory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	memory, exists := m.memories[id]
	if !exists {
		return nil, fmt.Errorf("memory with ID %s not found", id)
	}
	return memory, nil
}

func (m *MockMemoryRepository) Update(ctx context.Context, memory *domain.Memory) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.memories[memory.ID]; !exists {
		return fmt.Errorf("memory with ID %s not found", memory.ID)
	}

	m.memories[memory.ID] = memory
	m.metadata[memory.ID] = &ports.MemoryMetadata{
		ID:        memory.ID,
		ProjectID: memory.ProjectID,
		Type:      memory.Type,
		Title:     memory.Title,
		Tags:      memory.Tags,
		CreatedAt: memory.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	return nil
}

func (m *MockMemoryRepository) Delete(ctx context.Context, id domain.MemoryID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.memories[id]; !exists {
		return fmt.Errorf("memory with ID %s not found", id)
	}

	delete(m.memories, id)
	delete(m.metadata, id)
	return nil
}

func (m *MockMemoryRepository) GetByIDs(ctx context.Context, ids []domain.MemoryID) ([]*domain.Memory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*domain.Memory
	for _, id := range ids {
		if memory, exists := m.memories[id]; exists {
			results = append(results, memory)
		}
	}
	return results, nil
}

func (m *MockMemoryRepository) GetMetadataByIDs(ctx context.Context, ids []domain.MemoryID) ([]*ports.MemoryMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*ports.MemoryMetadata
	for _, id := range ids {
		if metadata, exists := m.metadata[id]; exists {
			results = append(results, metadata)
		}
	}
	return results, nil
}

func (m *MockMemoryRepository) ListByProject(ctx context.Context, projectID domain.ProjectID) ([]*domain.Memory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*domain.Memory
	for _, memory := range m.memories {
		if memory.ProjectID == projectID {
			results = append(results, memory)
		}
	}

	// Sort by created time (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})

	return results, nil
}

func (m *MockMemoryRepository) ListByType(ctx context.Context, projectID domain.ProjectID, memoryType domain.MemoryType) ([]*domain.Memory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*domain.Memory
	for _, memory := range m.memories {
		if memory.ProjectID == projectID && memory.Type == memoryType {
			results = append(results, memory)
		}
	}

	// Sort by created time (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})

	return results, nil
}

func (m *MockMemoryRepository) ListByTags(ctx context.Context, projectID domain.ProjectID, tags domain.Tags) ([]*domain.Memory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*domain.Memory
	for _, memory := range m.memories {
		if memory.ProjectID == projectID {
			// Check if memory contains all required tags
			hasAllTags := true
			for _, requiredTag := range tags {
				if !memory.Tags.Contains(requiredTag) {
					hasAllTags = false
					break
				}
			}
			if hasAllTags {
				results = append(results, memory)
			}
		}
	}

	// Sort by created time (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})

	return results, nil
}

func (m *MockMemoryRepository) ListBySession(ctx context.Context, sessionID domain.SessionID) ([]*domain.Memory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*domain.Memory
	for _, memory := range m.memories {
		if memory.SessionID != nil && *memory.SessionID == sessionID {
			results = append(results, memory)
		}
	}

	// Sort by created time (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})

	return results, nil
}

// MockEmbeddingProvider is a mock implementation of EmbeddingProvider
type MockEmbeddingProvider struct {
	mu               sync.RWMutex
	embeddings       map[string]domain.EmbeddingVector
	failOnText       map[string]error
	batchFailOnTexts map[int]error // fail on specific batch index
}

func NewMockEmbeddingProvider() *MockEmbeddingProvider {
	return &MockEmbeddingProvider{
		embeddings:       make(map[string]domain.EmbeddingVector),
		failOnText:       make(map[string]error),
		batchFailOnTexts: make(map[int]error),
	}
}

func (m *MockEmbeddingProvider) SetEmbedding(text string, vector domain.EmbeddingVector) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.embeddings[text] = vector
}

func (m *MockEmbeddingProvider) SetFailure(text string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failOnText[text] = err
}

func (m *MockEmbeddingProvider) SetBatchFailure(index int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.batchFailOnTexts[index] = err
}

func (m *MockEmbeddingProvider) GenerateEmbedding(ctx context.Context, text string) (domain.EmbeddingVector, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err, exists := m.failOnText[text]; exists {
		return nil, err
	}

	if vector, exists := m.embeddings[text]; exists {
		return vector, nil
	}

	// Generate deterministic embedding based on text length and content
	dimensions := 384 // Standard embedding dimension
	vector := make(domain.EmbeddingVector, dimensions)

	// Simple hash-based embedding generation
	hash := 0
	for _, char := range text {
		hash = hash*31 + int(char)
	}

	for i := 0; i < dimensions; i++ {
		vector[i] = float32((hash+i)%1000) / 1000.0
	}

	return vector, nil
}

func (m *MockEmbeddingProvider) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([]domain.EmbeddingVector, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []domain.EmbeddingVector
	for i, text := range texts {
		if err, exists := m.batchFailOnTexts[i]; exists {
			return nil, err
		}

		vector, err := m.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, err
		}
		results = append(results, vector)
	}

	return results, nil
}

func (m *MockEmbeddingProvider) GetDimensions() int {
	return 384
}

func (m *MockEmbeddingProvider) GetModelName() string {
	return "mock-embedding-model"
}

// MockVectorStore is a mock implementation of VectorStore
type MockVectorStore struct {
	mu      sync.RWMutex
	vectors map[string]vectorEntry
	failOn  map[string]error
}

type vectorEntry struct {
	Vector   domain.EmbeddingVector
	Metadata map[string]interface{}
}

func NewMockVectorStore() *MockVectorStore {
	return &MockVectorStore{
		vectors: make(map[string]vectorEntry),
		failOn:  make(map[string]error),
	}
}

func (m *MockVectorStore) SetFailure(operation string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failOn[operation] = err
}

func (m *MockVectorStore) Store(ctx context.Context, id string, vector domain.EmbeddingVector, metadata map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err, exists := m.failOn["store"]; exists {
		return err
	}

	m.vectors[id] = vectorEntry{
		Vector:   vector,
		Metadata: metadata,
	}
	return nil
}

func (m *MockVectorStore) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err, exists := m.failOn["delete"]; exists {
		return err
	}

	delete(m.vectors, id)
	return nil
}

func (m *MockVectorStore) Update(ctx context.Context, id string, vector domain.EmbeddingVector, metadata map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err, exists := m.failOn["update"]; exists {
		return err
	}

	m.vectors[id] = vectorEntry{
		Vector:   vector,
		Metadata: metadata,
	}
	return nil
}

func (m *MockVectorStore) BatchStore(ctx context.Context, items []ports.BatchStoreItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err, exists := m.failOn["batch_store"]; exists {
		return err
	}

	for _, item := range items {
		m.vectors[item.ID] = vectorEntry{
			Vector:   item.Vector,
			Metadata: item.Metadata,
		}
	}
	return nil
}

func (m *MockVectorStore) BatchDelete(ctx context.Context, ids []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err, exists := m.failOn["batch_delete"]; exists {
		return err
	}

	for _, id := range ids {
		delete(m.vectors, id)
	}
	return nil
}

func (m *MockVectorStore) Search(ctx context.Context, vector domain.EmbeddingVector, limit int, threshold float32) ([]ports.SearchResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err, exists := m.failOn["search"]; exists {
		return nil, err
	}

	type result struct {
		ID         string
		Similarity domain.Similarity
		Metadata   map[string]interface{}
	}

	var results []result
	for id, entry := range m.vectors {
		// Calculate dot product similarity
		similarity := calculateDotProduct(vector, entry.Vector)
		if float32(similarity) >= threshold {
			results = append(results, result{
				ID:         id,
				Similarity: similarity,
				Metadata:   entry.Metadata,
			})
		}
	}

	// Sort by similarity (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// Limit results
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	// Convert to SearchResult
	var searchResults []ports.SearchResult
	for _, r := range results {
		searchResults = append(searchResults, ports.SearchResult{
			ID:         r.ID,
			Similarity: r.Similarity,
			Metadata:   r.Metadata,
		})
	}

	return searchResults, nil
}

func (m *MockVectorStore) SearchByText(ctx context.Context, text string, limit int, threshold float32) ([]ports.SearchResult, error) {
	return nil, fmt.Errorf("SearchByText not implemented in mock")
}

func (m *MockVectorStore) CreateCollection(ctx context.Context, name string) error {
	if err, exists := m.failOn["create_collection"]; exists {
		return err
	}
	return nil
}

func (m *MockVectorStore) DeleteCollection(ctx context.Context, name string) error {
	if err, exists := m.failOn["delete_collection"]; exists {
		return err
	}
	return nil
}

func (m *MockVectorStore) ListCollections(ctx context.Context) ([]string, error) {
	if err, exists := m.failOn["list_collections"]; exists {
		return nil, err
	}
	return []string{"mock_collection"}, nil
}

// calculateDotProduct calculates dot product similarity
func calculateDotProduct(a, b domain.EmbeddingVector) domain.Similarity {
	if len(a) != len(b) {
		return 0
	}

	var sum float32
	for i := range a {
		sum += a[i] * b[i]
	}
	return domain.Similarity(sum)
}

// MockProjectRepository is a mock implementation of ProjectRepository
type MockProjectRepository struct {
	mu       sync.RWMutex
	projects map[domain.ProjectID]*domain.Project
	pathMap  map[string]domain.ProjectID // path -> project ID mapping
}

func NewMockProjectRepository() *MockProjectRepository {
	return &MockProjectRepository{
		projects: make(map[domain.ProjectID]*domain.Project),
		pathMap:  make(map[string]domain.ProjectID),
	}
}

func (m *MockProjectRepository) Store(ctx context.Context, project *domain.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.projects[project.ID]; exists {
		return fmt.Errorf("project with ID %s already exists", project.ID)
	}

	// Check for path conflicts
	if existingID, exists := m.pathMap[project.Path]; exists {
		return fmt.Errorf("project already exists at path %s with ID %s", project.Path, existingID)
	}

	m.projects[project.ID] = project
	m.pathMap[project.Path] = project.ID
	return nil
}

func (m *MockProjectRepository) GetByID(ctx context.Context, id domain.ProjectID) (*domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	project, exists := m.projects[id]
	if !exists {
		return nil, fmt.Errorf("project with ID %s not found", id)
	}
	return project, nil
}

func (m *MockProjectRepository) GetByPath(ctx context.Context, path string) (*domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	projectID, exists := m.pathMap[path]
	if !exists {
		return nil, fmt.Errorf("project not found at path: %s", path)
	}

	project, exists := m.projects[projectID]
	if !exists {
		return nil, fmt.Errorf("project with ID %s not found", projectID)
	}

	return project, nil
}

func (m *MockProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.projects[project.ID]; !exists {
		return fmt.Errorf("project with ID %s not found", project.ID)
	}

	// Update path mapping if path changed
	oldProject := m.projects[project.ID]
	if oldProject.Path != project.Path {
		delete(m.pathMap, oldProject.Path)

		// Check for path conflicts
		if existingID, exists := m.pathMap[project.Path]; exists && existingID != project.ID {
			return fmt.Errorf("project already exists at path %s with ID %s", project.Path, existingID)
		}

		m.pathMap[project.Path] = project.ID
	}

	m.projects[project.ID] = project
	return nil
}

func (m *MockProjectRepository) Delete(ctx context.Context, id domain.ProjectID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	project, exists := m.projects[id]
	if !exists {
		return fmt.Errorf("project with ID %s not found", id)
	}

	delete(m.projects, id)
	delete(m.pathMap, project.Path)
	return nil
}

func (m *MockProjectRepository) List(ctx context.Context) ([]*domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var projects []*domain.Project
	for _, project := range m.projects {
		projects = append(projects, project)
	}

	// Sort by created time (newest first)
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].CreatedAt.After(projects[j].CreatedAt)
	})

	return projects, nil
}

// MockSessionRepository is a mock implementation of SessionRepository
type MockSessionRepository struct {
	mu           sync.RWMutex
	sessions     map[domain.SessionID]*domain.Session
	idsByProject map[domain.ProjectID][]domain.SessionID
}

func NewMockSessionRepository() *MockSessionRepository {
	return &MockSessionRepository{
		sessions:     make(map[domain.SessionID]*domain.Session),
		idsByProject: make(map[domain.ProjectID][]domain.SessionID),
	}
}

func (m *MockSessionRepository) Store(ctx context.Context, session *domain.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sessions[session.ID]; exists {
		return fmt.Errorf("session with ID %s already exists", session.ID)
	}

	m.sessions[session.ID] = session
	m.idsByProject[session.ProjectID] = append(m.idsByProject[session.ProjectID], session.ID)
	return nil
}

func (m *MockSessionRepository) GetByID(ctx context.Context, id domain.SessionID) (*domain.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session with ID %s not found", id)
	}
	return session, nil
}

func (m *MockSessionRepository) GetActiveSession(ctx context.Context, projectID domain.ProjectID) (*domain.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessionIDs := m.idsByProject[projectID]
	for _, sessionID := range sessionIDs {
		session := m.sessions[sessionID]
		if session.IsActive() {
			return session, nil
		}
	}

	return nil, fmt.Errorf("no active session found for project %s", projectID)
}

func (m *MockSessionRepository) Update(ctx context.Context, session *domain.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sessions[session.ID]; !exists {
		return fmt.Errorf("session with ID %s not found", session.ID)
	}

	m.sessions[session.ID] = session
	return nil
}

func (m *MockSessionRepository) Delete(ctx context.Context, id domain.SessionID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[id]
	if !exists {
		return fmt.Errorf("session with ID %s not found", id)
	}

	delete(m.sessions, id)

	// Remove from project index
	projectIDs := m.idsByProject[session.ProjectID]
	for i, sessionID := range projectIDs {
		if sessionID == id {
			m.idsByProject[session.ProjectID] = append(projectIDs[:i], projectIDs[i+1:]...)
			break
		}
	}

	return nil
}

func (m *MockSessionRepository) ListByProject(ctx context.Context, projectID domain.ProjectID) ([]*domain.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*domain.Session
	sessionIDs := m.idsByProject[projectID]

	for _, sessionID := range sessionIDs {
		if session, exists := m.sessions[sessionID]; exists {
			results = append(results, session)
		}
	}

	// Sort by start time (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].StartTime.After(results[j].StartTime)
	})

	return results, nil
}

func (m *MockSessionRepository) ListWithFilters(ctx context.Context, filters ports.SessionFilters) ([]*domain.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*domain.Session

	for _, session := range m.sessions {
		// Apply filters
		if filters.ProjectID != nil && session.ProjectID != *filters.ProjectID {
			continue
		}
		if filters.Status != nil && session.Status != *filters.Status {
			continue
		}

		results = append(results, session)
	}

	// Sort by start time (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].StartTime.After(results[j].StartTime)
	})

	// Apply limit
	if filters.Limit > 0 && len(results) > filters.Limit {
		results = results[:filters.Limit]
	}

	return results, nil
}
