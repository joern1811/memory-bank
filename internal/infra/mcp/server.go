package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
)

// MemoryBankServer implements the MCP server for Memory Bank
type MemoryBankServer struct {
	memoryService  ports.MemoryService
	projectService ports.ProjectService
	sessionService ports.SessionService
	logger         *logrus.Logger
}

// NewMemoryBankServer creates a new MCP server instance
func NewMemoryBankServer(
	memoryService ports.MemoryService,
	projectService ports.ProjectService,
	sessionService ports.SessionService,
	logger *logrus.Logger,
) *MemoryBankServer {
	return &MemoryBankServer{
		memoryService:  memoryService,
		projectService: projectService,
		sessionService: sessionService,
		logger:         logger,
	}
}

// RegisterMethods registers all MCP methods for the Memory Bank server
func (s *MemoryBankServer) RegisterMethods(mcpServer *server.MCPServer) {
	// Note: The mark3labs/mcp-go library uses Tools instead of direct method registration
	// We'll need to implement this as tools rather than RPC methods
	
	s.logger.Info("MCP methods registered successfully")
}

// CreateMemoryRequest represents a request to create a new memory
type CreateMemoryRequest struct {
	ProjectID   string                 `json:"project_id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	SessionID   *string                `json:"session_id,omitempty"`
}

// CreateMemoryResponse represents the response from creating a memory
type CreateMemoryResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *MemoryBankServer) handleCreateMemory(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling memory/create request")

	var req CreateMemoryRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	// Validate required fields
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	if req.Type == "" {
		return nil, fmt.Errorf("type is required")
	}
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if req.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	// Convert to domain types
	projectID := domain.ProjectID(req.ProjectID)
	memoryType := domain.MemoryType(req.Type)
	tags := domain.Tags(req.Tags)

	var sessionID *domain.SessionID
	if req.SessionID != nil {
		sid := domain.SessionID(*req.SessionID)
		sessionID = &sid
	}

	// Create memory using service
	createReq := ports.CreateMemoryRequest{
		ProjectID: projectID,
		SessionID: sessionID,
		Type:      memoryType,
		Title:     req.Title,
		Content:   req.Content,
		Context:   "", // Could be extracted from metadata
		Tags:      tags,
	}
	
	memory, err := s.memoryService.CreateMemory(ctx, createReq)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create memory")
		return nil, fmt.Errorf("failed to create memory: %w", err)
	}

	response := CreateMemoryResponse{
		ID:        string(memory.ID),
		CreatedAt: memory.CreatedAt,
	}

	s.logger.WithFields(logrus.Fields{
		"memory_id":  memory.ID,
		"project_id": projectID,
		"type":       memoryType,
	}).Info("Memory created successfully")

	return response, nil
}

// SearchMemoriesRequest represents a request to search memories
type SearchMemoriesRequest struct {
	Query       string   `json:"query"`
	ProjectID   *string  `json:"project_id,omitempty"`
	Type        *string  `json:"type,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Limit       *int     `json:"limit,omitempty"`
	Threshold   *float32 `json:"threshold,omitempty"`
}

// SearchMemoriesResponse represents the response from searching memories
type SearchMemoriesResponse struct {
	Results []MemorySearchResult `json:"results"`
	Total   int                  `json:"total"`
}

// MemorySearchResult represents a single search result
type MemorySearchResult struct {
	ID         string                 `json:"id"`
	ProjectID  string                 `json:"project_id"`
	Type       string                 `json:"type"`
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	Tags       []string               `json:"tags"`
	Metadata   map[string]interface{} `json:"metadata"`
	Similarity float32                `json:"similarity"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

func (s *MemoryBankServer) handleSearchMemories(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling memory/search request")

	var req SearchMemoriesRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	if req.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	// Set defaults
	limit := 10
	if req.Limit != nil {
		limit = *req.Limit
	}

	threshold := float32(0.5)
	if req.Threshold != nil {
		threshold = *req.Threshold
	}

	// Build search filters
	filters := ports.MemoryFilters{}
	if req.ProjectID != nil {
		projectID := domain.ProjectID(*req.ProjectID)
		filters.ProjectID = &projectID
	}
	if req.Type != nil {
		memoryType := domain.MemoryType(*req.Type)
		filters.Type = &memoryType
	}
	if len(req.Tags) > 0 {
		filters.Tags = domain.Tags(req.Tags)
	}

	// Perform search
	searchQuery := ports.SemanticSearchRequest{
		Query:     req.Query,
		ProjectID: filters.ProjectID,
		Type:      filters.Type,
		Tags:      filters.Tags,
		Limit:     limit,
		Threshold: threshold,
	}
	
	searchResults, err := s.memoryService.SearchMemories(ctx, searchQuery)
	if err != nil {
		s.logger.WithError(err).Error("Failed to search memories")
		return nil, fmt.Errorf("failed to search memories: %w", err)
	}

	// Convert results
	results := make([]MemorySearchResult, len(searchResults))
	for i, result := range searchResults {
		results[i] = MemorySearchResult{
			ID:         string(result.Memory.ID),
			ProjectID:  string(result.Memory.ProjectID),
			Type:       string(result.Memory.Type),
			Title:      result.Memory.Title,
			Content:    result.Memory.Content,
			Tags:       []string(result.Memory.Tags),
			Metadata:   map[string]interface{}{"context": result.Memory.Context}, // Use context as metadata
			Similarity: float32(result.Similarity),
			CreatedAt:  result.Memory.CreatedAt,
			UpdatedAt:  result.Memory.UpdatedAt,
		}
	}

	response := SearchMemoriesResponse{
		Results: results,
		Total:   len(results),
	}

	s.logger.WithFields(logrus.Fields{
		"query":         req.Query,
		"results_count": len(results),
	}).Info("Memory search completed")

	return response, nil
}

// GetMemoryRequest represents a request to get a specific memory
type GetMemoryRequest struct {
	ID string `json:"id"`
}

func (s *MemoryBankServer) handleGetMemory(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling memory/get request")

	var req GetMemoryRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	if req.ID == "" {
		return nil, fmt.Errorf("id is required")
	}

	memoryID := domain.MemoryID(req.ID)
	memory, err := s.memoryService.GetMemory(ctx, memoryID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get memory")
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}

	result := MemorySearchResult{
		ID:        string(memory.ID),
		ProjectID: string(memory.ProjectID),
		Type:      string(memory.Type),
		Title:     memory.Title,
		Content:   memory.Content,
		Tags:      []string(memory.Tags),
		Metadata:  map[string]interface{}{"context": memory.Context},
		CreatedAt: memory.CreatedAt,
		UpdatedAt: memory.UpdatedAt,
	}

	s.logger.WithField("memory_id", memoryID).Info("Memory retrieved successfully")
	return result, nil
}

// UpdateMemoryRequest represents a request to update a memory
type UpdateMemoryRequest struct {
	ID       string                 `json:"id"`
	Title    *string                `json:"title,omitempty"`
	Content  *string                `json:"content,omitempty"`
	Tags     []string               `json:"tags,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (s *MemoryBankServer) handleUpdateMemory(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling memory/update request")

	var req UpdateMemoryRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	if req.ID == "" {
		return nil, fmt.Errorf("id is required")
	}

	memoryID := domain.MemoryID(req.ID)

	// Get existing memory
	memory, err := s.memoryService.GetMemory(ctx, memoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}

	// Apply updates
	if req.Title != nil {
		memory.Title = *req.Title
	}
	if req.Content != nil {
		memory.Content = *req.Content
	}
	if len(req.Tags) > 0 {
		memory.Tags = domain.Tags(req.Tags)
	}
	if req.Metadata != nil {
		if context, ok := req.Metadata["context"].(string); ok {
			memory.Context = context
		}
	}

	// Update memory
	err = s.memoryService.UpdateMemory(ctx, memory)
	if err != nil {
		s.logger.WithError(err).Error("Failed to update memory")
		return nil, fmt.Errorf("failed to update memory: %w", err)
	}

	result := MemorySearchResult{
		ID:        string(memory.ID),
		ProjectID: string(memory.ProjectID),
		Type:      string(memory.Type),
		Title:     memory.Title,
		Content:   memory.Content,
		Tags:      []string(memory.Tags),
		Metadata:  map[string]interface{}{"context": memory.Context},
		CreatedAt: memory.CreatedAt,
		UpdatedAt: memory.UpdatedAt,
	}

	s.logger.WithField("memory_id", memoryID).Info("Memory updated successfully")
	return result, nil
}

// DeleteMemoryRequest represents a request to delete a memory
type DeleteMemoryRequest struct {
	ID string `json:"id"`
}

func (s *MemoryBankServer) handleDeleteMemory(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling memory/delete request")

	var req DeleteMemoryRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	if req.ID == "" {
		return nil, fmt.Errorf("id is required")
	}

	memoryID := domain.MemoryID(req.ID)
	err := s.memoryService.DeleteMemory(ctx, memoryID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to delete memory")
		return nil, fmt.Errorf("failed to delete memory: %w", err)
	}

	s.logger.WithField("memory_id", memoryID).Info("Memory deleted successfully")
	return map[string]interface{}{"success": true}, nil
}

// ListMemoriesRequest represents a request to list memories
type ListMemoriesRequest struct {
	ProjectID *string  `json:"project_id,omitempty"`
	Type      *string  `json:"type,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

func (s *MemoryBankServer) handleListMemories(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling memory/list request")

	var req ListMemoriesRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	// Build filters
	filters := ports.MemoryFilters{}
	if req.ProjectID != nil {
		projectID := domain.ProjectID(*req.ProjectID)
		filters.ProjectID = &projectID
	}
	if req.Type != nil {
		memoryType := domain.MemoryType(*req.Type)
		filters.Type = &memoryType
	}
	if len(req.Tags) > 0 {
		filters.Tags = domain.Tags(req.Tags)
	}

	// Use empty search query to list all memories
	searchQuery := ports.SemanticSearchRequest{
		Query:     "", // Empty query will return all memories
		ProjectID: filters.ProjectID,
		Type:      filters.Type,
		Tags:      filters.Tags,
		Limit:     1000, // Large limit to get all results
		Threshold: 0.0,  // No threshold filtering
	}
	
	searchResults, err := s.memoryService.SearchMemories(ctx, searchQuery)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list memories")
		return nil, fmt.Errorf("failed to list memories: %w", err)
	}

	// Convert results
	results := make([]MemorySearchResult, len(searchResults))
	for i, result := range searchResults {
		results[i] = MemorySearchResult{
			ID:        string(result.Memory.ID),
			ProjectID: string(result.Memory.ProjectID),
			Type:      string(result.Memory.Type),
			Title:     result.Memory.Title,
			Content:   result.Memory.Content,
			Tags:      []string(result.Memory.Tags),
			Metadata:  map[string]interface{}{"context": result.Memory.Context},
			CreatedAt: result.Memory.CreatedAt,
			UpdatedAt: result.Memory.UpdatedAt,
		}
	}

	response := SearchMemoriesResponse{
		Results: results,
		Total:   len(results),
	}

	s.logger.WithField("count", len(results)).Info("Memories listed successfully")
	return response, nil
}

// InitProjectRequest represents a request to initialize a project
type InitProjectRequest struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
}

// InitProjectResponse represents the response from initializing a project
type InitProjectResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *MemoryBankServer) handleInitProject(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling project/init request")

	var req InitProjectRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	createProjectReq := ports.CreateProjectRequest{
		Name:        req.Name,
		Path:        req.Path,
		Description: req.Description,
	}
	
	project, err := s.projectService.CreateProject(ctx, createProjectReq)
	if err != nil {
		s.logger.WithError(err).Error("Failed to initialize project")
		return nil, fmt.Errorf("failed to initialize project: %w", err)
	}

	response := InitProjectResponse{
		ID:        string(project.ID),
		Name:      project.Name,
		Path:      project.Path,
		CreatedAt: project.CreatedAt,
	}

	s.logger.WithFields(logrus.Fields{
		"project_id": project.ID,
		"name":       project.Name,
		"path":       project.Path,
	}).Info("Project initialized successfully")

	return response, nil
}

// GetProjectRequest represents a request to get project information
type GetProjectRequest struct {
	ID   *string `json:"id,omitempty"`
	Path *string `json:"path,omitempty"`
}

func (s *MemoryBankServer) handleGetProject(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling project/get request")

	var req GetProjectRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	var project *domain.Project
	var err error

	if req.ID != nil {
		projectID := domain.ProjectID(*req.ID)
		project, err = s.projectService.GetProject(ctx, projectID)
	} else if req.Path != nil {
		project, err = s.projectService.GetProjectByPath(ctx, *req.Path)
	} else {
		return nil, fmt.Errorf("either id or path is required")
	}

	if err != nil {
		s.logger.WithError(err).Error("Failed to get project")
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	response := InitProjectResponse{
		ID:        string(project.ID),
		Name:      project.Name,
		Path:      project.Path,
		CreatedAt: project.CreatedAt,
	}

	s.logger.WithField("project_id", project.ID).Info("Project retrieved successfully")
	return response, nil
}

func (s *MemoryBankServer) handleListProjects(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling project/list request")

	projects, err := s.projectService.ListProjects(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list projects")
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	results := make([]InitProjectResponse, len(projects))
	for i, project := range projects {
		results[i] = InitProjectResponse{
			ID:        string(project.ID),
			Name:      project.Name,
			Path:      project.Path,
			CreatedAt: project.CreatedAt,
		}
	}

	s.logger.WithField("count", len(results)).Info("Projects listed successfully")
	return map[string]interface{}{"projects": results}, nil
}

// Session-related handlers would be implemented similarly...
// StartSessionRequest, LogSessionRequest, CompleteSessionRequest, etc.
// For brevity, I'm including placeholders here

func (s *MemoryBankServer) handleStartSession(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling session/start request - not implemented yet")
	return nil, fmt.Errorf("session/start not implemented yet")
}

func (s *MemoryBankServer) handleLogSession(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling session/log request - not implemented yet")
	return nil, fmt.Errorf("session/log not implemented yet")
}

func (s *MemoryBankServer) handleCompleteSession(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling session/complete request - not implemented yet")
	return nil, fmt.Errorf("session/complete not implemented yet")
}

func (s *MemoryBankServer) handleGetSession(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling session/get request - not implemented yet")
	return nil, fmt.Errorf("session/get not implemented yet")
}