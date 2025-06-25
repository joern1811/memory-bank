package app

import (
	"context"
	"fmt"
	"sort"
	"strings"

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

	// Batch retrieve memories instead of individual lookups
	memoryIDs := make([]domain.MemoryID, len(searchResults))
	for i, result := range searchResults {
		memoryIDs[i] = domain.MemoryID(result.ID)
	}

	memories, err := s.memoryRepo.GetByIDs(ctx, memoryIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to batch retrieve memories: %w", err)
	}

	// Create lookup map for O(1) access
	memoryMap := make(map[domain.MemoryID]*domain.Memory)
	for _, memory := range memories {
		memoryMap[memory.ID] = memory
	}

	// Build results maintaining search order
	var results []ports.MemorySearchResult
	for _, searchResult := range searchResults {
		memoryID := domain.MemoryID(searchResult.ID)
		if memory, exists := memoryMap[memoryID]; exists && s.matchesFilters(memory, query) {
			results = append(results, ports.MemorySearchResult{
				Memory:     memory,
				Similarity: searchResult.Similarity,
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

// ListMemories lists memories based on filters
func (s *MemoryService) ListMemories(ctx context.Context, req ports.ListMemoriesRequest) ([]*domain.Memory, error) {
	s.logger.WithFields(logrus.Fields{
		"project_id": req.ProjectID,
		"type":       req.Type,
		"limit":      req.Limit,
	}).Info("Listing memories")

	// Use repository's ListByProject if project filter is specified
	if req.ProjectID != nil {
		memories, err := s.memoryRepo.ListByProject(ctx, *req.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to get memories by project: %w", err)
		}

		// Apply additional filters
		var filtered []*domain.Memory
		for _, memory := range memories {
			if s.matchesListFilters(memory, req) {
				filtered = append(filtered, memory)
			}
		}

		// Apply limit
		if req.Limit > 0 && len(filtered) > req.Limit {
			filtered = filtered[:req.Limit]
		}

		return filtered, nil
	}

	// For global listing, we would need to implement a GetAll method in repository
	// For now, return empty list with informative log
	s.logger.Info("Global memory listing requested but not yet implemented - requires GetAll repository method")
	return []*domain.Memory{}, nil
}

// matchesListFilters checks if a memory matches the list filters
func (s *MemoryService) matchesListFilters(memory *domain.Memory, req ports.ListMemoriesRequest) bool {
	// Type filter
	if req.Type != nil && memory.Type != *req.Type {
		return false
	}

	// Tags filter
	if len(req.Tags) > 0 {
		for _, requiredTag := range req.Tags {
			if !memory.Tags.Contains(requiredTag) {
				return false
			}
		}
	}

	return true
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

	// Time filter implementation would go here
	// if req.CreatedAfter != nil && memory.CreatedAt.Before(*req.CreatedAfter) {
	//     return false
	// }
	// if req.CreatedBefore != nil && memory.CreatedAt.After(*req.CreatedBefore) {
	//     return false
	// }

	return true
}

// FacetedSearch performs advanced search with faceting and filtering
func (s *MemoryService) FacetedSearch(ctx context.Context, req ports.FacetedSearchRequest) (*ports.FacetedSearchResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"query":          req.Query,
		"project_id":     req.ProjectID,
		"include_facets": req.IncludeFacets,
		"limit":          req.Limit,
	}).Info("Performing faceted search")

	// Convert to basic search request
	basicQuery := ports.SemanticSearchRequest{
		Query:     req.Query,
		ProjectID: req.ProjectID,
		Limit:     req.Limit * 2, // Get more results for filtering
		Threshold: req.Threshold,
	}

	// Perform basic search
	searchResults, err := s.SearchMemories(ctx, basicQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to perform basic search: %w", err)
	}

	// Apply additional filters
	filteredResults := s.applyAdvancedFilters(searchResults, req.Filters)

	// Sort results if requested
	if req.SortBy != nil {
		s.sortResults(filteredResults, *req.SortBy)
	}

	// Limit results
	if len(filteredResults) > req.Limit {
		filteredResults = filteredResults[:req.Limit]
	}

	response := &ports.FacetedSearchResponse{
		Results: filteredResults,
		Total:   len(filteredResults),
	}

	// Generate facets if requested
	if req.IncludeFacets {
		facets, err := s.generateFacets(ctx, searchResults, req.Filters)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to generate facets")
		} else {
			response.Facets = facets
		}
	}

	return response, nil
}

// SearchWithRelevanceScoring performs search with enhanced relevance scoring
func (s *MemoryService) SearchWithRelevanceScoring(ctx context.Context, query ports.SemanticSearchRequest) ([]ports.EnhancedMemorySearchResult, error) {
	s.logger.WithField("query", query.Query).Info("Performing enhanced relevance search")

	// Perform basic search
	basicResults, err := s.SearchMemories(ctx, query)
	if err != nil {
		return nil, err
	}

	// Enhance results with relevance scoring
	enhancedResults := make([]ports.EnhancedMemorySearchResult, len(basicResults))
	queryTerms := strings.Fields(strings.ToLower(query.Query))

	for i, result := range basicResults {
		enhancedResults[i] = ports.EnhancedMemorySearchResult{
			Memory:         result.Memory,
			Similarity:     result.Similarity,
			RelevanceScore: s.calculateRelevanceScore(result.Memory, queryTerms, float64(result.Similarity)),
			MatchReasons:   s.getMatchReasons(result.Memory, queryTerms),
			Highlights:     s.getHighlights(result.Memory, queryTerms),
		}
	}

	// Sort by relevance score
	sort.Slice(enhancedResults, func(i, j int) bool {
		return enhancedResults[i].RelevanceScore > enhancedResults[j].RelevanceScore
	})

	return enhancedResults, nil
}

// GetSearchSuggestions provides search suggestions based on existing content
func (s *MemoryService) GetSearchSuggestions(ctx context.Context, partialQuery string, projectID *domain.ProjectID) ([]string, error) {
	s.logger.WithField("partial_query", partialQuery).Info("Getting search suggestions")

	// Get recent memories for suggestions
	listReq := ports.ListMemoriesRequest{
		ProjectID: projectID,
		Limit:     100,
	}

	memories, err := s.ListMemories(ctx, listReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get memories for suggestions: %w", err)
	}

	suggestions := make(map[string]int)
	partialLower := strings.ToLower(partialQuery)

	// Extract suggestions from titles, tags, and content
	for _, memory := range memories {
		// From titles
		if strings.Contains(strings.ToLower(memory.Title), partialLower) {
			s.addWordSuggestions(memory.Title, partialLower, suggestions)
		}

		// From tags
		for _, tag := range memory.Tags {
			if strings.Contains(strings.ToLower(tag), partialLower) {
				suggestions[tag] = suggestions[tag] + 1
			}
		}

		// From content (extract key phrases)
		contentWords := strings.Fields(memory.Content)
		for _, word := range contentWords {
			if len(word) > 3 && strings.Contains(strings.ToLower(word), partialLower) {
				cleanWord := strings.ToLower(strings.Trim(word, ".,!?;:"))
				suggestions[cleanWord] = suggestions[cleanWord] + 1
			}
		}
	}

	// Convert to sorted slice
	type suggestion struct {
		text  string
		count int
	}

	var suggestionList []suggestion
	for text, count := range suggestions {
		if count > 0 {
			suggestionList = append(suggestionList, suggestion{text, count})
		}
	}

	// Sort by frequency
	sort.Slice(suggestionList, func(i, j int) bool {
		return suggestionList[i].count > suggestionList[j].count
	})

	// Return top 10 suggestions
	result := make([]string, 0, 10)
	for i, s := range suggestionList {
		if i >= 10 {
			break
		}
		result = append(result, s.text)
	}

	return result, nil
}

// Helper methods for enhanced search features

func (s *MemoryService) applyAdvancedFilters(results []ports.MemorySearchResult, filters *ports.SearchFilters) []ports.MemorySearchResult {
	if filters == nil {
		return results
	}

	var filtered []ports.MemorySearchResult

	for _, result := range results {
		memory := result.Memory

		// Type filter
		if len(filters.Types) > 0 {
			found := false
			for _, t := range filters.Types {
				if memory.Type == t {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Tags filter
		if len(filters.Tags) > 0 {
			hasAllTags := true
			for _, requiredTag := range filters.Tags {
				if !memory.Tags.Contains(requiredTag) {
					hasAllTags = false
					break
				}
			}
			if !hasAllTags {
				continue
			}
		}

		// Content length filters
		contentLen := len(memory.Content)
		if filters.MinLength != nil && contentLen < *filters.MinLength {
			continue
		}
		if filters.MaxLength != nil && contentLen > *filters.MaxLength {
			continue
		}

		// Has content filter
		if filters.HasContent && len(strings.TrimSpace(memory.Content)) == 0 {
			continue
		}

		// Session ID filter
		if len(filters.SessionIDs) > 0 && memory.SessionID != nil {
			found := false
			for _, sessionID := range filters.SessionIDs {
				if *memory.SessionID == sessionID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Time filter (basic implementation)
		// TODO: Implement time filtering when TimeFilter is properly defined
		// if filters.TimeFilter != nil {
		//     // Time filtering would be implemented here based on memory timestamps
		//     // For example: filter by CreatedAt, UpdatedAt ranges
		// }

		filtered = append(filtered, result)
	}

	return filtered
}

func (s *MemoryService) sortResults(results []ports.MemorySearchResult, sortBy ports.SortOption) {
	sort.Slice(results, func(i, j int) bool {
		switch sortBy.Field {
		case ports.SortByRelevance:
			if sortBy.Direction == ports.SortAsc {
				return results[i].Similarity < results[j].Similarity
			}
			return results[i].Similarity > results[j].Similarity

		case ports.SortByTitle:
			if sortBy.Direction == ports.SortAsc {
				return results[i].Memory.Title < results[j].Memory.Title
			}
			return results[i].Memory.Title > results[j].Memory.Title

		case ports.SortByType:
			if sortBy.Direction == ports.SortAsc {
				return results[i].Memory.Type < results[j].Memory.Type
			}
			return results[i].Memory.Type > results[j].Memory.Type

		case ports.SortByCreatedAt:
			// Sort by creation timestamp (newer first for descending)
			if sortBy.Direction == ports.SortDesc {
				return results[i].Memory.CreatedAt.After(results[j].Memory.CreatedAt)
			}
			return results[i].Memory.CreatedAt.Before(results[j].Memory.CreatedAt)

		case ports.SortByUpdatedAt:
			// Sort by update timestamp (newer first for descending)
			if sortBy.Direction == ports.SortDesc {
				return results[i].Memory.UpdatedAt.After(results[j].Memory.UpdatedAt)
			}
			return results[i].Memory.UpdatedAt.Before(results[j].Memory.UpdatedAt)

		default:
			return false
		}
	})
}

func (s *MemoryService) generateFacets(ctx context.Context, results []ports.MemorySearchResult, filters *ports.SearchFilters) (*ports.SearchFacets, error) {
	facets := &ports.SearchFacets{}

	// Type facets
	typeCounts := make(map[domain.MemoryType]int)
	tagCounts := make(map[string]int)

	for _, result := range results {
		memory := result.Memory

		// Count types
		typeCounts[memory.Type]++

		// Count tags
		for _, tag := range memory.Tags {
			tagCounts[tag]++
		}
	}

	// Convert type counts to facets
	for memoryType, count := range typeCounts {
		facets.Types = append(facets.Types, ports.TypeFacet{
			Type:  memoryType,
			Count: count,
		})
	}

	// Convert tag counts to facets (top 20)
	type tagCount struct {
		tag   string
		count int
	}
	var tagList []tagCount
	for tag, count := range tagCounts {
		tagList = append(tagList, tagCount{tag, count})
	}
	sort.Slice(tagList, func(i, j int) bool {
		return tagList[i].count > tagList[j].count
	})

	for i, tc := range tagList {
		if i >= 20 { // Limit to top 20 tags
			break
		}
		facets.Tags = append(facets.Tags, ports.TagFacet{
			Tag:   tc.tag,
			Count: tc.count,
		})
	}

	// Generate time period facets
	facets.TimePeriods = s.generateTimePeriodFacets(results)

	return facets, nil
}

func (s *MemoryService) generateTimePeriodFacets(results []ports.MemorySearchResult) []ports.TimePeriodFacet {
	// Simplified time period implementation
	// In a real implementation, this would use actual timestamps from memories

	periods := map[string]int{
		"Today":      0,
		"This week":  0,
		"This month": 0,
		"This year":  0,
		"Older":      0,
	}

	// For now, simulate distribution
	total := len(results)
	if total > 0 {
		periods["Today"] = total / 10
		periods["This week"] = total / 5
		periods["This month"] = total / 3
		periods["This year"] = total / 2
		periods["Older"] = total - periods["Today"] - periods["This week"] - periods["This month"] - periods["This year"]
	}

	var facets []ports.TimePeriodFacet
	for period, count := range periods {
		if count > 0 {
			facets = append(facets, ports.TimePeriodFacet{
				Period: period,
				Count:  count,
			})
		}
	}

	return facets
}

func (s *MemoryService) calculateRelevanceScore(memory *domain.Memory, queryTerms []string, similarity float64) float64 {
	score := similarity

	titleLower := strings.ToLower(memory.Title)
	contentLower := strings.ToLower(memory.Content)

	// Boost for title matches
	titleMatches := 0
	for _, term := range queryTerms {
		if strings.Contains(titleLower, term) {
			titleMatches++
		}
	}
	score += float64(titleMatches) * 0.3

	// Boost for exact phrase matches
	queryPhrase := strings.Join(queryTerms, " ")
	if strings.Contains(titleLower, queryPhrase) {
		score += 0.5
	}
	if strings.Contains(contentLower, queryPhrase) {
		score += 0.2
	}

	// Boost for tag matches
	tagMatches := 0
	for _, tag := range memory.Tags {
		tagLower := strings.ToLower(tag)
		for _, term := range queryTerms {
			if strings.Contains(tagLower, term) {
				tagMatches++
			}
		}
	}
	score += float64(tagMatches) * 0.1

	// Boost for memory type relevance
	if memory.Type == domain.MemoryTypeDecision {
		score += 0.1 // Decisions are often important
	}

	return score
}

func (s *MemoryService) getMatchReasons(memory *domain.Memory, queryTerms []string) []string {
	var reasons []string

	titleLower := strings.ToLower(memory.Title)
	contentLower := strings.ToLower(memory.Content)

	// Check title matches
	for _, term := range queryTerms {
		if strings.Contains(titleLower, term) {
			reasons = append(reasons, fmt.Sprintf("Title contains '%s'", term))
		}
	}

	// Check content matches
	contentMatches := 0
	for _, term := range queryTerms {
		if strings.Contains(contentLower, term) {
			contentMatches++
		}
	}
	if contentMatches > 0 {
		reasons = append(reasons, fmt.Sprintf("Content contains %d matching terms", contentMatches))
	}

	// Check tag matches
	tagMatches := []string{}
	for _, tag := range memory.Tags {
		tagLower := strings.ToLower(tag)
		for _, term := range queryTerms {
			if strings.Contains(tagLower, term) {
				tagMatches = append(tagMatches, tag)
				break
			}
		}
	}
	if len(tagMatches) > 0 {
		reasons = append(reasons, fmt.Sprintf("Tags match: %s", strings.Join(tagMatches, ", ")))
	}

	// Check memory type
	reasons = append(reasons, fmt.Sprintf("Memory type: %s", memory.Type))

	return reasons
}

func (s *MemoryService) getHighlights(memory *domain.Memory, queryTerms []string) []string {
	var highlights []string

	// Highlight title matches
	titleHighlight := s.highlightText(memory.Title, queryTerms)
	if titleHighlight != memory.Title {
		highlights = append(highlights, titleHighlight)
	}

	// Highlight content snippets (first 200 chars with matches)
	contentSnippets := s.extractContentSnippets(memory.Content, queryTerms, 200, 2)
	highlights = append(highlights, contentSnippets...)

	return highlights
}

func (s *MemoryService) highlightText(text string, queryTerms []string) string {
	result := text
	for _, term := range queryTerms {
		// Simple highlighting (in production, would use proper highlighting library)
		result = strings.ReplaceAll(result, term, fmt.Sprintf("**%s**", term))
		// Basic title case handling without deprecated strings.Title
		if len(term) > 0 {
			titleTerm := strings.ToUpper(term[:1]) + strings.ToLower(term[1:])
			result = strings.ReplaceAll(result, titleTerm, fmt.Sprintf("**%s**", titleTerm))
		}
	}
	return result
}

func (s *MemoryService) extractContentSnippets(content string, queryTerms []string, maxLength int, maxSnippets int) []string {
	var snippets []string
	contentLower := strings.ToLower(content)

	for _, term := range queryTerms {
		index := strings.Index(contentLower, term)
		if index != -1 {
			start := max(0, index-50)
			end := min(len(content), index+len(term)+50)

			snippet := content[start:end]
			if start > 0 {
				snippet = "..." + snippet
			}
			if end < len(content) {
				snippet = snippet + "..."
			}

			highlighted := s.highlightText(snippet, queryTerms)
			snippets = append(snippets, highlighted)

			if len(snippets) >= maxSnippets {
				break
			}
		}
	}

	return snippets
}

func (s *MemoryService) addWordSuggestions(text, partial string, suggestions map[string]int) {
	words := strings.Fields(text)
	for _, word := range words {
		cleanWord := strings.ToLower(strings.Trim(word, ".,!?;:"))
		if len(cleanWord) > 2 && strings.Contains(cleanWord, partial) {
			suggestions[cleanWord] = suggestions[cleanWord] + 1
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RegenerateEmbedding regenerates the embedding for a specific memory
func (s *MemoryService) RegenerateEmbedding(ctx context.Context, memoryID domain.MemoryID) error {
	// Get the memory
	memory, err := s.memoryRepo.GetByID(ctx, memoryID)
	if err != nil {
		return fmt.Errorf("failed to get memory: %w", err)
	}

	// Generate embedding
	embeddingText := fmt.Sprintf("%s %s", memory.Title, memory.Content)
	embedding, err := s.embeddingProvider.GenerateEmbedding(ctx, embeddingText)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to generate embedding, but memory exists")
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Store in vector store
	if err := s.vectorStore.Store(ctx, string(memory.ID), embedding, map[string]interface{}{
		"type":    string(memory.Type),
		"title":   memory.Title,
		"content": memory.Content,
	}); err != nil {
		s.logger.WithError(err).Warn("Failed to store embedding in vector store")
		return fmt.Errorf("failed to store embedding: %w", err)
	}

	// Update embedding flag
	memory.HasEmbedding = true
	if err := s.memoryRepo.Update(ctx, memory); err != nil {
		return fmt.Errorf("failed to update memory embedding flag: %w", err)
	}

	s.logger.WithField("memory_id", memory.ID).Info("Successfully regenerated embedding")
	return nil
}

// CleanupEmbeddings resets all embedding flags and regenerates embeddings for all memories in a project
func (s *MemoryService) CleanupEmbeddings(ctx context.Context, projectID domain.ProjectID) (*ports.CleanupResult, error) {
	s.logger.WithField("project_id", projectID).Info("Starting embedding cleanup")

	// Get all memories for the project
	memories, err := s.memoryRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list memories: %w", err)
	}

	result := &ports.CleanupResult{
		TotalMemories:       len(memories),
		MemoriesProcessed:   0,
		EmbeddingsGenerated: 0,
		Errors:              0,
		ErrorMessages:       []string{},
	}

	// Reset all embedding flags first
	if err := s.memoryRepo.ResetEmbeddingFlags(ctx, string(projectID)); err != nil {
		return result, fmt.Errorf("failed to reset embedding flags: %w", err)
	}

	s.logger.WithField("count", len(memories)).Info("Reset embedding flags for all memories")

	// Try to delete the ChromaDB collection to clean up old embeddings
	if err := s.vectorStore.DeleteCollection(ctx, "memory_bank"); err != nil {
		s.logger.WithError(err).Warn("Failed to delete ChromaDB collection, continuing anyway")
	}

	// Regenerate embeddings for all memories
	for _, memory := range memories {
		result.MemoriesProcessed++

		if err := s.RegenerateEmbedding(ctx, memory.ID); err != nil {
			result.Errors++
			errorMsg := fmt.Sprintf("Memory %s: %v", memory.ID, err)
			result.ErrorMessages = append(result.ErrorMessages, errorMsg)
			s.logger.WithError(err).WithField("memory_id", memory.ID).Error("Failed to regenerate embedding")
		} else {
			result.EmbeddingsGenerated++
		}
	}

	s.logger.WithFields(logrus.Fields{
		"total_memories":       result.TotalMemories,
		"embeddings_generated": result.EmbeddingsGenerated,
		"errors":               result.Errors,
	}).Info("Embedding cleanup completed")

	return result, nil
}
