package app

import (
	"context"
	"fmt"
	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/sirupsen/logrus"
)

// MemoryService implements the memory service use cases
type MemoryService struct {
	memoryRepo        ports.MemoryRepository
	embeddingProvider ports.EmbeddingProvider
	vectorStore       ports.VectorStore
	logger            *logrus.Logger
}

// NewMemoryService creates a new memory service
func NewMemoryService(
	memoryRepo ports.MemoryRepository,
	embeddingProvider ports.EmbeddingProvider,
	vectorStore ports.VectorStore,
	logger *logrus.Logger,
) *MemoryService {
	return &MemoryService{
		memoryRepo:        memoryRepo,
		embeddingProvider: embeddingProvider,
		vectorStore:       vectorStore,
		logger:            logger,
	}
}

// CreateMemory creates a new memory entry with embedding
func (s *MemoryService) CreateMemory(ctx context.Context, req ports.CreateMemoryRequest) (*domain.Memory, error) {
	s.logger.WithFields(logrus.Fields{
		"project_id": req.ProjectID,
		"type":       req.Type,
		"title":      req.Title,
	}).Info("Creating memory")

	// Create memory entity
	memory := domain.NewMemory(req.ProjectID, req.Type, req.Title, req.Content, req.Context)
	if req.SessionID != nil {
		memory.SessionID = req.SessionID
	}
	
	// Add tags
	for _, tag := range req.Tags {
		memory.AddTag(tag)
	}

	// Store in database first
	if err := s.memoryRepo.Store(ctx, memory); err != nil {
		s.logger.WithError(err).Error("Failed to store memory")
		return nil, fmt.Errorf("failed to store memory: %w", err)
	}

	// Generate and store embedding
	if err := s.generateAndStoreEmbedding(ctx, memory); err != nil {
		s.logger.WithError(err).Warn("Failed to generate embedding, but memory was stored")
		// Don't fail the entire operation if embedding fails
	}

	s.logger.WithField("memory_id", memory.ID).Info("Memory created successfully")
	return memory, nil
}

// GetMemory retrieves a memory by ID
func (s *MemoryService) GetMemory(ctx context.Context, id domain.MemoryID) (*domain.Memory, error) {
	return s.memoryRepo.GetByID(ctx, id)
}

// UpdateMemory updates an existing memory
func (s *MemoryService) UpdateMemory(ctx context.Context, memory *domain.Memory) error {
	s.logger.WithField("memory_id", memory.ID).Info("Updating memory")

	// Update in database
	if err := s.memoryRepo.Update(ctx, memory); err != nil {
		return fmt.Errorf("failed to update memory: %w", err)
	}

	// Regenerate embedding if content changed
	if err := s.generateAndStoreEmbedding(ctx, memory); err != nil {
		s.logger.WithError(err).Warn("Failed to update embedding")
	}

	return nil
}

// DeleteMemory deletes a memory entry
func (s *MemoryService) DeleteMemory(ctx context.Context, id domain.MemoryID) error {
	s.logger.WithField("memory_id", id).Info("Deleting memory")

	// Delete from vector store first
	if err := s.vectorStore.Delete(ctx, string(id)); err != nil {
		s.logger.WithError(err).Warn("Failed to delete from vector store")
	}

	// Delete from database
	if err := s.memoryRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}

	return nil
}

// SearchMemories performs semantic search on memories
func (s *MemoryService) SearchMemories(ctx context.Context, query ports.SemanticSearchRequest) ([]ports.MemorySearchResult, error) {
	s.logger.WithFields(logrus.Fields{
		"query":      query.Query,
		"project_id": query.ProjectID,
		"limit":      query.Limit,
	}).Info("Searching memories")

	// Generate embedding for query
	queryVector, err := s.embeddingProvider.GenerateEmbedding(ctx, query.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search in vector store
	searchResults, err := s.vectorStore.Search(ctx, queryVector, query.Limit, query.Threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to search vector store: %w", err)
	}

	// Convert to memory search results
	var results []ports.MemorySearchResult
	for _, result := range searchResults {
		memoryID := domain.MemoryID(result.ID)
		memory, err := s.memoryRepo.GetByID(ctx, memoryID)
		if err != nil {
			s.logger.WithError(err).WithField("memory_id", memoryID).Warn("Failed to fetch memory for search result")
			continue
		}

		// Apply filters
		if s.matchesFilters(memory, query) {
			results = append(results, ports.MemorySearchResult{
				Memory:     memory,
				Similarity: result.Similarity,
			})
		}
	}

	s.logger.WithField("result_count", len(results)).Info("Search completed")
	return results, nil
}

// FindSimilarMemories finds memories similar to a given memory
func (s *MemoryService) FindSimilarMemories(ctx context.Context, memoryID domain.MemoryID, limit int) ([]ports.MemorySearchResult, error) {
	// Get the original memory
	memory, err := s.memoryRepo.GetByID(ctx, memoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}

	// Search for similar memories
	query := ports.SemanticSearchRequest{
		Query:     memory.GetEmbeddingText(),
		ProjectID: &memory.ProjectID,
		Limit:     limit + 1, // +1 because we'll exclude the original
		Threshold: 0.1,       // Low threshold for similarity
	}

	results, err := s.SearchMemories(ctx, query)
	if err != nil {
		return nil, err
	}

	// Filter out the original memory
	var filtered []ports.MemorySearchResult
	for _, result := range results {
		if result.Memory.ID != memoryID {
			filtered = append(filtered, result)
		}
	}

	return filtered, nil
}

// CreateDecision creates a new decision memory
func (s *MemoryService) CreateDecision(ctx context.Context, req ports.CreateDecisionRequest) (*domain.Decision, error) {
	decision := domain.NewDecision(
		req.ProjectID,
		req.Title,
		req.Content,
		req.Context,
		req.Rationale,
		req.Options,
	)
	decision.Outcome = req.Outcome

	if req.SessionID != nil {
		decision.Memory.SessionID = req.SessionID
	}

	// Add tags
	for _, tag := range req.Tags {
		decision.Memory.AddTag(tag)
	}

	// Store the underlying memory
	if err := s.memoryRepo.Store(ctx, decision.Memory); err != nil {
		return nil, fmt.Errorf("failed to store decision: %w", err)
	}

	// Generate embedding
	if err := s.generateAndStoreEmbedding(ctx, decision.Memory); err != nil {
		s.logger.WithError(err).Warn("Failed to generate embedding for decision")
	}

	return decision, nil
}

// CreatePattern creates a new pattern memory
func (s *MemoryService) CreatePattern(ctx context.Context, req ports.CreatePatternRequest) (*domain.Pattern, error) {
	pattern := domain.NewPattern(
		req.ProjectID,
		req.Title,
		req.PatternType,
		req.Implementation,
		req.UseCase,
	)
	pattern.Language = req.Language

	if req.SessionID != nil {
		pattern.Memory.SessionID = req.SessionID
	}

	// Add tags
	for _, tag := range req.Tags {
		pattern.Memory.AddTag(tag)
	}

	// Store the underlying memory
	if err := s.memoryRepo.Store(ctx, pattern.Memory); err != nil {
		return nil, fmt.Errorf("failed to store pattern: %w", err)
	}

	// Generate embedding
	if err := s.generateAndStoreEmbedding(ctx, pattern.Memory); err != nil {
		s.logger.WithError(err).Warn("Failed to generate embedding for pattern")
	}

	return pattern, nil
}

// CreateErrorSolution creates a new error solution memory
func (s *MemoryService) CreateErrorSolution(ctx context.Context, req ports.CreateErrorSolutionRequest) (*domain.ErrorSolution, error) {
	errorSolution := domain.NewErrorSolution(
		req.ProjectID,
		req.Title,
		req.ErrorSignature,
		req.Solution,
		req.Context,
	)
	errorSolution.StackTrace = req.StackTrace
	errorSolution.Language = req.Language

	if req.SessionID != nil {
		errorSolution.Memory.SessionID = req.SessionID
	}

	// Add tags
	for _, tag := range req.Tags {
		errorSolution.Memory.AddTag(tag)
	}

	// Store the underlying memory
	if err := s.memoryRepo.Store(ctx, errorSolution.Memory); err != nil {
		return nil, fmt.Errorf("failed to store error solution: %w", err)
	}

	// Generate embedding
	if err := s.generateAndStoreEmbedding(ctx, errorSolution.Memory); err != nil {
		s.logger.WithError(err).Warn("Failed to generate embedding for error solution")
	}

	return errorSolution, nil
}

// generateAndStoreEmbedding generates and stores embedding for a memory
func (s *MemoryService) generateAndStoreEmbedding(ctx context.Context, memory *domain.Memory) error {
	// Generate embedding
	text := memory.GetEmbeddingText()
	vector, err := s.embeddingProvider.GenerateEmbedding(ctx, text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Prepare metadata
	metadata := map[string]interface{}{
		"memory_id":  memory.ID,
		"project_id": memory.ProjectID,
		"type":       memory.Type,
		"title":      memory.Title,
		"tags":       memory.Tags,
		"created_at": memory.CreatedAt,
	}

	if memory.SessionID != nil {
		metadata["session_id"] = *memory.SessionID
	}

	// Store in vector store
	if err := s.vectorStore.Store(ctx, string(memory.ID), vector, metadata); err != nil {
		return fmt.Errorf("failed to store embedding: %w", err)
	}

	// Mark memory as having embedding
	memory.SetEmbedding()
	if err := s.memoryRepo.Update(ctx, memory); err != nil {
		s.logger.WithError(err).Warn("Failed to update memory embedding flag")
	}

	return nil
}

// matchesFilters checks if a memory matches the search filters
func (s *MemoryService) matchesFilters(memory *domain.Memory, query ports.SemanticSearchRequest) bool {
	// Project filter
	if query.ProjectID != nil && memory.ProjectID != *query.ProjectID {
		return false
	}

	// Type filter
	if query.Type != nil && memory.Type != *query.Type {
		return false
	}

	// Tags filter
	if len(query.Tags) > 0 {
		for _, requiredTag := range query.Tags {
			if !memory.Tags.Contains(requiredTag) {
				return false
			}
		}
	}

	// TODO: Implement time filter

	return true
}
