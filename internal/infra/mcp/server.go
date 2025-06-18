package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/mark3labs/mcp-go/mcp"
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

// RegisterMethods registers all MCP tools and resources for the Memory Bank server
func (s *MemoryBankServer) RegisterMethods(mcpServer *server.MCPServer) {
	// Register system prompt resource
	systemPromptResource := mcp.NewResource(
		"prompt://memory-bank/system",
		"Memory Bank System Prompt",
		mcp.WithResourceDescription("Dynamic system prompt for optimal Memory Bank integration with MCP clients"),
		mcp.WithMIMEType("text/plain"),
	)

	mcpServer.AddResource(systemPromptResource, s.handleSystemPromptResource)
	
	// Register Memory operations as tools
	mcpServer.AddTool(mcp.NewTool("memory_create",
		mcp.WithDescription("Create a new memory entry"),
		mcp.WithString("project_id", mcp.Description("Project ID"), mcp.Required()),
		mcp.WithString("type", mcp.Description("Memory type"), mcp.Required()),
		mcp.WithString("title", mcp.Description("Memory title"), mcp.Required()),
		mcp.WithString("content", mcp.Description("Memory content"), mcp.Required()),
		mcp.WithArray("tags", mcp.Description("Memory tags")),
		mcp.WithString("session_id", mcp.Description("Session ID")),
	), s.handleCreateMemoryTool)
	
	mcpServer.AddTool(mcp.NewTool("memory_search",
		mcp.WithDescription("Search memories semantically"),
		mcp.WithString("query", mcp.Description("Search query"), mcp.Required()),
		mcp.WithString("project_id", mcp.Description("Project ID to filter by")),
		mcp.WithString("type", mcp.Description("Memory type to filter by")),
		mcp.WithArray("tags", mcp.Description("Tags to filter by")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results")),
		mcp.WithNumber("threshold", mcp.Description("Similarity threshold")),
	), s.handleSearchMemoriesTool)
	
	mcpServer.AddTool(mcp.NewTool("memory_get",
		mcp.WithDescription("Get a specific memory by ID"),
		mcp.WithString("id", mcp.Description("Memory ID"), mcp.Required()),
	), s.handleGetMemoryTool)
	
	mcpServer.AddTool(mcp.NewTool("memory_update",
		mcp.WithDescription("Update an existing memory"),
		mcp.WithString("id", mcp.Description("Memory ID"), mcp.Required()),
		mcp.WithString("title", mcp.Description("New title")),
		mcp.WithString("content", mcp.Description("New content")),
		mcp.WithArray("tags", mcp.Description("New tags")),
	), s.handleUpdateMemoryTool)
	
	mcpServer.AddTool(mcp.NewTool("memory_delete",
		mcp.WithDescription("Delete a memory"),
		mcp.WithString("id", mcp.Description("Memory ID"), mcp.Required()),
	), s.handleDeleteMemoryTool)
	
	mcpServer.AddTool(mcp.NewTool("memory_list",
		mcp.WithDescription("List memories with optional filters"),
		mcp.WithString("project_id", mcp.Description("Project ID to filter by")),
		mcp.WithString("type", mcp.Description("Memory type to filter by")),
		mcp.WithArray("tags", mcp.Description("Tags to filter by")),
	), s.handleListMemoriesTool)
	
	// Register advanced search operations
	mcpServer.AddTool(mcp.NewTool("memory_faceted-search",
		mcp.WithDescription("Advanced search with facets and filters"),
		mcp.WithString("query", mcp.Description("Search query"), mcp.Required()),
		mcp.WithString("project_id", mcp.Description("Project ID to filter by")),
		mcp.WithObject("filters", mcp.Description("Search filters")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results")),
		mcp.WithNumber("threshold", mcp.Description("Similarity threshold")),
		mcp.WithBoolean("include_facets", mcp.Description("Include facets in response")),
		mcp.WithObject("sort_by", mcp.Description("Sort options")),
	), s.handleFacetedSearchTool)
	
	mcpServer.AddTool(mcp.NewTool("memory_enhanced-search",
		mcp.WithDescription("Enhanced search with relevance scoring and highlights"),
		mcp.WithString("query", mcp.Description("Search query"), mcp.Required()),
		mcp.WithString("project_id", mcp.Description("Project ID to filter by")),
		mcp.WithString("type", mcp.Description("Memory type to filter by")),
		mcp.WithArray("tags", mcp.Description("Tags to filter by")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results")),
		mcp.WithNumber("threshold", mcp.Description("Similarity threshold")),
	), s.handleEnhancedSearchTool)
	
	mcpServer.AddTool(mcp.NewTool("memory_search-suggestions",
		mcp.WithDescription("Get intelligent search suggestions"),
		mcp.WithString("partial_query", mcp.Description("Partial query for suggestions"), mcp.Required()),
		mcp.WithString("project_id", mcp.Description("Project ID to filter by")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of suggestions")),
	), s.handleSearchSuggestionsTool)
	
	// Register Project operations
	mcpServer.AddTool(mcp.NewTool("project_init",
		mcp.WithDescription("Initialize a new project"),
		mcp.WithString("name", mcp.Description("Project name"), mcp.Required()),
		mcp.WithString("path", mcp.Description("Project path"), mcp.Required()),
		mcp.WithString("description", mcp.Description("Project description")),
	), s.handleInitProjectTool)
	
	mcpServer.AddTool(mcp.NewTool("project_get",
		mcp.WithDescription("Get project information"),
		mcp.WithString("id", mcp.Description("Project ID")),
		mcp.WithString("path", mcp.Description("Project path")),
	), s.handleGetProjectTool)
	
	mcpServer.AddTool(mcp.NewTool("project_list",
		mcp.WithDescription("List all projects"),
	), s.handleListProjectsTool)
	
	// Register Session operations
	mcpServer.AddTool(mcp.NewTool("session_start",
		mcp.WithDescription("Start a new development session"),
		mcp.WithString("title", mcp.Description("Session title"), mcp.Required()),
		mcp.WithString("project_id", mcp.Description("Project ID"), mcp.Required()),
		mcp.WithString("description", mcp.Description("Session description")),
	), s.handleStartSessionTool)
	
	mcpServer.AddTool(mcp.NewTool("session_log",
		mcp.WithDescription("Log progress to the active session"),
		mcp.WithString("message", mcp.Description("Progress message"), mcp.Required()),
		mcp.WithString("project_id", mcp.Description("Project ID")),
		mcp.WithString("session_id", mcp.Description("Session ID")),
	), s.handleLogSessionTool)
	
	mcpServer.AddTool(mcp.NewTool("session_complete",
		mcp.WithDescription("Complete a development session"),
		mcp.WithString("outcome", mcp.Description("Session outcome"), mcp.Required()),
		mcp.WithString("project_id", mcp.Description("Project ID")),
		mcp.WithString("session_id", mcp.Description("Session ID")),
	), s.handleCompleteSessionTool)
	
	mcpServer.AddTool(mcp.NewTool("session_get",
		mcp.WithDescription("Get session details"),
		mcp.WithString("id", mcp.Description("Session ID"), mcp.Required()),
	), s.handleGetSessionTool)
	
	s.logger.Info("MCP tools and resources registered successfully")
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

// Tool handlers that wrap the existing handlers to match MCP tool interface
func (s *MemoryBankServer) handleCreateMemoryTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Handling memory/create tool request")
	
	// Convert tool arguments to JSON for reuse with existing handler
	params, err := json.Marshal(request.Params.Arguments)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error marshaling arguments: %v", err),
				},
			},
		}, nil
	}
	
	result, err := s.handleCreateMemory(ctx, params)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error: %v", err),
				},
			},
		}, nil
	}
	
	// Convert result to JSON string for MCP response
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error marshaling result: %v", err),
				},
			},
		}, nil
	}
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(resultJSON),
			},
		},
	}, nil
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

// Advanced Search Features

// FacetedSearchRequest represents a request for faceted search
type FacetedSearchRequest struct {
	Query         string                 `json:"query"`
	ProjectID     *string                `json:"project_id,omitempty"`
	Filters       *SearchFilters         `json:"filters,omitempty"`
	Limit         *int                   `json:"limit,omitempty"`
	Threshold     *float32               `json:"threshold,omitempty"`
	IncludeFacets *bool                  `json:"include_facets,omitempty"`
	SortBy        *SortOption            `json:"sort_by,omitempty"`
}

// SearchFilters represents comprehensive search filters
type SearchFilters struct {
	Types      []string     `json:"types,omitempty"`
	Tags       []string     `json:"tags,omitempty"`
	SessionIDs []string     `json:"session_ids,omitempty"`
	TimeFilter *TimeFilter  `json:"time_filter,omitempty"`
	HasContent *bool        `json:"has_content,omitempty"`
	MinLength  *int         `json:"min_length,omitempty"`
	MaxLength  *int         `json:"max_length,omitempty"`
}

// TimeFilter represents time-based filtering options
type TimeFilter struct {
	After  *string `json:"after,omitempty"`  // ISO 8601 format
	Before *string `json:"before,omitempty"` // ISO 8601 format
}

// SortOption represents sorting options for search results
type SortOption struct {
	Field     string `json:"field"`     // "relevance", "created_at", "updated_at", "title", "type"
	Direction string `json:"direction"` // "asc", "desc"
}

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
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// TagFacet represents a tag facet  
type TagFacet struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// ProjectFacet represents a project facet
type ProjectFacet struct {
	ProjectID   string `json:"project_id"`
	ProjectName string `json:"project_name"`
	Count       int    `json:"count"`
}

// SessionFacet represents a session facet
type SessionFacet struct {
	SessionID    string `json:"session_id"`
	SessionTitle string `json:"session_title"`
	Count        int    `json:"count"`
}

// TimePeriodFacet represents a time period facet
type TimePeriodFacet struct {
	Period string `json:"period"`
	Count  int    `json:"count"`
}

func (s *MemoryBankServer) handleFacetedSearch(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling memory/faceted-search request")

	var req FacetedSearchRequest
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

	includeFacets := false
	if req.IncludeFacets != nil {
		includeFacets = *req.IncludeFacets
	}

	// Build search request
	searchReq := ports.FacetedSearchRequest{
		Query:         req.Query,
		Limit:         limit,
		Threshold:     threshold,
		IncludeFacets: includeFacets,
	}

	// Set project ID
	if req.ProjectID != nil {
		projectID := domain.ProjectID(*req.ProjectID)
		searchReq.ProjectID = &projectID
	}

	// Convert filters
	if req.Filters != nil {
		filters := &ports.SearchFilters{}

		// Convert types
		if len(req.Filters.Types) > 0 {
			for _, t := range req.Filters.Types {
				filters.Types = append(filters.Types, domain.MemoryType(t))
			}
		}

		// Convert tags
		if len(req.Filters.Tags) > 0 {
			filters.Tags = domain.Tags(req.Filters.Tags)
		}

		// Convert session IDs
		if len(req.Filters.SessionIDs) > 0 {
			for _, sid := range req.Filters.SessionIDs {
				sessionID := domain.SessionID(sid)
				filters.SessionIDs = append(filters.SessionIDs, sessionID)
			}
		}

		// Convert time filter
		if req.Filters.TimeFilter != nil {
			filters.TimeFilter = &ports.TimeFilter{
				After:  req.Filters.TimeFilter.After,
				Before: req.Filters.TimeFilter.Before,
			}
		}

		// Convert other filters
		if req.Filters.HasContent != nil {
			filters.HasContent = *req.Filters.HasContent
		}
		filters.MinLength = req.Filters.MinLength
		filters.MaxLength = req.Filters.MaxLength

		searchReq.Filters = filters
	}

	// Convert sort option
	if req.SortBy != nil {
		sortOption := &ports.SortOption{
			Direction: ports.SortDirection(req.SortBy.Direction),
		}

		switch req.SortBy.Field {
		case "relevance":
			sortOption.Field = ports.SortByRelevance
		case "created_at":
			sortOption.Field = ports.SortByCreatedAt
		case "updated_at":
			sortOption.Field = ports.SortByUpdatedAt
		case "title":
			sortOption.Field = ports.SortByTitle
		case "type":
			sortOption.Field = ports.SortByType
		default:
			sortOption.Field = ports.SortByRelevance
		}

		searchReq.SortBy = sortOption
	}

	// Perform faceted search
	searchResponse, err := s.memoryService.FacetedSearch(ctx, searchReq)
	if err != nil {
		s.logger.WithError(err).Error("Failed to perform faceted search")
		return nil, fmt.Errorf("failed to perform faceted search: %w", err)
	}

	// Convert results
	results := make([]MemorySearchResult, len(searchResponse.Results))
	for i, result := range searchResponse.Results {
		results[i] = MemorySearchResult{
			ID:         string(result.Memory.ID),
			ProjectID:  string(result.Memory.ProjectID),
			Type:       string(result.Memory.Type),
			Title:      result.Memory.Title,
			Content:    result.Memory.Content,
			Tags:       []string(result.Memory.Tags),
			Metadata:   map[string]interface{}{"context": result.Memory.Context},
			Similarity: float32(result.Similarity),
			CreatedAt:  result.Memory.CreatedAt,
			UpdatedAt:  result.Memory.UpdatedAt,
		}
	}

	response := FacetedSearchResponse{
		Results: results,
		Total:   searchResponse.Total,
	}

	// Convert facets if included
	if searchResponse.Facets != nil {
		facets := &SearchFacets{}

		// Convert type facets
		for _, typeFacet := range searchResponse.Facets.Types {
			facets.Types = append(facets.Types, TypeFacet{
				Type:  string(typeFacet.Type),
				Count: typeFacet.Count,
			})
		}

		// Convert tag facets
		for _, tagFacet := range searchResponse.Facets.Tags {
			facets.Tags = append(facets.Tags, TagFacet{
				Tag:   tagFacet.Tag,
				Count: tagFacet.Count,
			})
		}

		// Convert project facets
		for _, projectFacet := range searchResponse.Facets.Projects {
			facets.Projects = append(facets.Projects, ProjectFacet{
				ProjectID:   string(projectFacet.ProjectID),
				ProjectName: projectFacet.ProjectName,
				Count:       projectFacet.Count,
			})
		}

		// Convert session facets
		for _, sessionFacet := range searchResponse.Facets.Sessions {
			facets.Sessions = append(facets.Sessions, SessionFacet{
				SessionID:    string(sessionFacet.SessionID),
				SessionTitle: sessionFacet.SessionTitle,
				Count:        sessionFacet.Count,
			})
		}

		// Convert time period facets
		for _, timeFacet := range searchResponse.Facets.TimePeriods {
			facets.TimePeriods = append(facets.TimePeriods, TimePeriodFacet{
				Period: timeFacet.Period,
				Count:  timeFacet.Count,
			})
		}

		response.Facets = facets
	}

	s.logger.WithFields(logrus.Fields{
		"query":         req.Query,
		"results_count": len(results),
		"has_facets":    response.Facets != nil,
	}).Info("Faceted search completed")

	return response, nil
}

// EnhancedSearchRequest represents a request for enhanced search with relevance scoring
type EnhancedSearchRequest struct {
	Query     string   `json:"query"`
	ProjectID *string  `json:"project_id,omitempty"`
	Type      *string  `json:"type,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Limit     *int     `json:"limit,omitempty"`
	Threshold *float32 `json:"threshold,omitempty"`
}

// EnhancedSearchResponse represents enhanced search results
type EnhancedSearchResponse struct {
	Results []EnhancedMemorySearchResult `json:"results"`
	Total   int                          `json:"total"`
}

// EnhancedMemorySearchResult represents a memory with enhanced relevance scoring
type EnhancedMemorySearchResult struct {
	ID             string                 `json:"id"`
	ProjectID      string                 `json:"project_id"`
	Type           string                 `json:"type"`
	Title          string                 `json:"title"`
	Content        string                 `json:"content"`
	Tags           []string               `json:"tags"`
	Metadata       map[string]interface{} `json:"metadata"`
	Similarity     float32                `json:"similarity"`
	RelevanceScore float64                `json:"relevance_score"`
	MatchReasons   []string               `json:"match_reasons"`
	Highlights     []string               `json:"highlights"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

func (s *MemoryBankServer) handleEnhancedSearch(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling memory/enhanced-search request")

	var req EnhancedSearchRequest
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

	// Build search request
	searchQuery := ports.SemanticSearchRequest{
		Query:     req.Query,
		Limit:     limit,
		Threshold: threshold,
	}

	// Set optional fields
	if req.ProjectID != nil {
		projectID := domain.ProjectID(*req.ProjectID)
		searchQuery.ProjectID = &projectID
	}
	if req.Type != nil {
		memoryType := domain.MemoryType(*req.Type)
		searchQuery.Type = &memoryType
	}
	if len(req.Tags) > 0 {
		searchQuery.Tags = domain.Tags(req.Tags)
	}

	// Perform enhanced search
	searchResults, err := s.memoryService.SearchWithRelevanceScoring(ctx, searchQuery)
	if err != nil {
		s.logger.WithError(err).Error("Failed to perform enhanced search")
		return nil, fmt.Errorf("failed to perform enhanced search: %w", err)
	}

	// Convert results
	results := make([]EnhancedMemorySearchResult, len(searchResults))
	for i, result := range searchResults {
		results[i] = EnhancedMemorySearchResult{
			ID:             string(result.Memory.ID),
			ProjectID:      string(result.Memory.ProjectID),
			Type:           string(result.Memory.Type),
			Title:          result.Memory.Title,
			Content:        result.Memory.Content,
			Tags:           []string(result.Memory.Tags),
			Metadata:       map[string]interface{}{"context": result.Memory.Context},
			Similarity:     float32(result.Similarity),
			RelevanceScore: result.RelevanceScore,
			MatchReasons:   result.MatchReasons,
			Highlights:     result.Highlights,
			CreatedAt:      result.Memory.CreatedAt,
			UpdatedAt:      result.Memory.UpdatedAt,
		}
	}

	response := EnhancedSearchResponse{
		Results: results,
		Total:   len(results),
	}

	s.logger.WithFields(logrus.Fields{
		"query":         req.Query,
		"results_count": len(results),
	}).Info("Enhanced search completed")

	return response, nil
}

// SearchSuggestionsRequest represents a request for search suggestions
type SearchSuggestionsRequest struct {
	PartialQuery string  `json:"partial_query"`
	ProjectID    *string `json:"project_id,omitempty"`
	Limit        *int    `json:"limit,omitempty"`
}

// SearchSuggestionsResponse represents search suggestions
type SearchSuggestionsResponse struct {
	Suggestions []string `json:"suggestions"`
}

func (s *MemoryBankServer) handleSearchSuggestions(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling memory/search-suggestions request")

	var req SearchSuggestionsRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	if req.PartialQuery == "" {
		return nil, fmt.Errorf("partial_query is required")
	}

	// Convert project ID
	var projectID *domain.ProjectID
	if req.ProjectID != nil {
		pid := domain.ProjectID(*req.ProjectID)
		projectID = &pid
	}

	// Get suggestions
	suggestions, err := s.memoryService.GetSearchSuggestions(ctx, req.PartialQuery, projectID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get search suggestions")
		return nil, fmt.Errorf("failed to get search suggestions: %w", err)
	}

	// Apply limit if specified
	if req.Limit != nil && len(suggestions) > *req.Limit {
		suggestions = suggestions[:*req.Limit]
	}

	response := SearchSuggestionsResponse{
		Suggestions: suggestions,
	}

	s.logger.WithFields(logrus.Fields{
		"partial_query":     req.PartialQuery,
		"suggestions_count": len(suggestions),
	}).Info("Search suggestions generated")

	return response, nil
}

// handleSystemPromptResource handles requests for the system prompt resource
func (s *MemoryBankServer) handleSystemPromptResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	s.logger.Debug("Handling system prompt resource request")

	// Generate dynamic system prompt based on current context
	prompt, err := s.generateSystemPrompt(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate system prompt")
		return nil, fmt.Errorf("failed to generate system prompt: %w", err)
	}

	// Create text resource contents
	textContent := mcp.TextResourceContents{
		URI:      request.Params.URI,
		MIMEType: "text/plain",
		Text:     prompt,
	}

	s.logger.Info("System prompt resource generated successfully")
	return []mcp.ResourceContents{textContent}, nil
}

// generateSystemPrompt creates a dynamic system prompt with current context
func (s *MemoryBankServer) generateSystemPrompt(ctx context.Context) (string, error) {
	// Get current project information and existing memories for context
	projects, err := s.projectService.ListProjects(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to load projects for system prompt")
		projects = []*domain.Project{} // Continue with empty projects
	}

	// Build contextual information
	var contextInfo strings.Builder
	contextInfo.WriteString("# Memory Bank Integration Context\n\n")
	
	if len(projects) > 0 {
		contextInfo.WriteString("## Current Projects\n")
		for _, project := range projects {
			contextInfo.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", 
				project.Name, 
				string(project.ID), 
				project.Description))
		}
		contextInfo.WriteString("\n")
	}

	// Get recent memories for each project to provide context
	memoryContext := s.buildMemoryContext(ctx, projects)
	if memoryContext != "" {
		contextInfo.WriteString("## Available Memory Types\n")
		contextInfo.WriteString(memoryContext)
		contextInfo.WriteString("\n")
	}

	// Generate the complete system prompt
	systemPrompt := fmt.Sprintf(`# Memory Bank - MCP Integration System Prompt

You are working with Memory Bank, a semantic memory management system that helps store and retrieve development knowledge. Use this system to maintain context across development sessions and build institutional knowledge.

%s## How to Use Memory Bank Effectively

### When to Store Memories
- **Decisions**: Store architectural decisions with rationale and alternatives considered
- **Patterns**: Save code patterns, design patterns, and best practices you discover
- **Error Solutions**: Document errors encountered and their solutions for future reference
- **Code Snippets**: Store reusable code examples with context
- **Documentation**: Keep project-specific documentation and notes
- **Session Progress**: Track development session progress and outcomes

### Memory Creation Best Practices
- **Use descriptive titles**: Clear, searchable titles help retrieval
- **Add relevant tags**: Use consistent tagging for better organization (e.g., "auth", "api", "frontend")
- **Include context**: Add enough context so the memory is useful later
- **Structure content**: Use markdown formatting for readability
- **Reference related memories**: Link to related decisions or patterns when relevant

### Effective Search Strategies
- **Semantic search**: Use natural language queries to find related content
- **Faceted search**: Filter by type, tags, or project for precise results
- **Enhanced search**: Get relevance scoring and match explanations
- **Use search suggestions**: Get intelligent suggestions based on existing content

### Memory Types Guide
- **Decision**: type "decision" - For architectural and design decisions
- **Pattern**: type "pattern" - For reusable code or design patterns  
- **ErrorSolution**: type "error_solution" - For documented error fixes
- **Code**: type "code" - For code snippets and examples
- **Documentation**: type "documentation" - For project documentation

### Available MCP Methods
- memory/create: Create new memory entries
- memory/search: Semantic search across memories
- memory/faceted-search: Advanced search with filters and facets
- memory/enhanced-search: Search with relevance scoring and highlights
- memory/search-suggestions: Get intelligent search suggestions
- memory/get: Retrieve specific memory by ID
- memory/update: Update existing memories
- memory/delete: Remove memories
- memory/list: List memories with optional filters
- project/init: Initialize new project
- project/get: Get project information
- project/list: List all projects
- session/start: Start development session
- session/log: Log session progress
- session/complete: Complete session with outcome

### Integration Tips
1. **Start sessions**: Use session/start when beginning major development work
2. **Document decisions**: Store important decisions as they're made
3. **Search before implementing**: Check existing patterns and solutions first
4. **Tag consistently**: Use consistent tags across related memories
5. **Update memories**: Keep memories current as code evolves
6. **Complete sessions**: Document outcomes when finishing work

### Example Usage Patterns
- Before starting a new feature, search for related patterns: memory/search "authentication patterns"
- When encountering an error, check for solutions: memory/search "JWT token validation error"
- After making a decision, document it: memory/create with type "decision"
- Store useful code patterns: memory/create with type "pattern"
- Track progress on complex tasks: session/start, session/log, session/complete

Remember: The more you use Memory Bank consistently, the more valuable it becomes as your development knowledge base grows and evolves.
`, contextInfo.String())

	return systemPrompt, nil
}

// buildMemoryContext analyzes existing memories to provide context about available content
func (s *MemoryBankServer) buildMemoryContext(ctx context.Context, projects []*domain.Project) string {
	var context strings.Builder
	
	// Get sample of memories from each project to understand what's available
	for _, project := range projects {
		// Search for recent memories in this project
		searchReq := ports.SemanticSearchRequest{
			Query:     "", // Empty query to get all memories
			ProjectID: &project.ID,
			Limit:     10,
			Threshold: 0.0,
		}
		
		memories, err := s.memoryService.SearchMemories(ctx, searchReq)
		if err != nil {
			s.logger.WithError(err).Debug("Failed to load memories for context")
			continue
		}
		
		if len(memories) > 0 {
			context.WriteString(fmt.Sprintf("### %s Project Memories (%d total)\n", project.Name, len(memories)))
			
			// Group by type to show what kinds of memories exist
			typeCount := make(map[domain.MemoryType]int)
			for _, memory := range memories {
				typeCount[memory.Memory.Type]++
			}
			
			for memType, count := range typeCount {
				context.WriteString(fmt.Sprintf("- %s: %d entries\n", string(memType), count))
			}
			context.WriteString("\n")
		}
	}
	
	return context.String()
}

// Additional tool wrapper functions
func (s *MemoryBankServer) handleSearchMemoriesTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	params, err := json.Marshal(request.Params.Arguments)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error marshaling arguments: %v", err)},
			},
		}, nil
	}
	
	result, err := s.handleSearchMemories(ctx, params)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: %v", err)},
			},
		}, nil
	}
	
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error marshaling result: %v", err)},
			},
		}, nil
	}
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: string(resultJSON)},
		},
	}, nil
}

// Utility function to wrap existing handlers for MCP tool interface
func (s *MemoryBankServer) wrapHandler(ctx context.Context, request mcp.CallToolRequest, handler func(context.Context, json.RawMessage) (interface{}, error)) (*mcp.CallToolResult, error) {
	params, err := json.Marshal(request.Params.Arguments)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error marshaling arguments: %v", err)},
			},
		}, nil
	}
	
	result, err := handler(ctx, params)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: %v", err)},
			},
		}, nil
	}
	
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error marshaling result: %v", err)},
			},
		}, nil
	}
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: string(resultJSON)},
		},
	}, nil
}

func (s *MemoryBankServer) handleGetMemoryTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.wrapHandler(ctx, request, s.handleGetMemory)
}

func (s *MemoryBankServer) handleUpdateMemoryTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.wrapHandler(ctx, request, s.handleUpdateMemory)
}

func (s *MemoryBankServer) handleDeleteMemoryTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.wrapHandler(ctx, request, s.handleDeleteMemory)
}

func (s *MemoryBankServer) handleListMemoriesTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.wrapHandler(ctx, request, s.handleListMemories)
}

func (s *MemoryBankServer) handleFacetedSearchTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.wrapHandler(ctx, request, s.handleFacetedSearch)
}

func (s *MemoryBankServer) handleEnhancedSearchTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.wrapHandler(ctx, request, s.handleEnhancedSearch)
}

func (s *MemoryBankServer) handleSearchSuggestionsTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.wrapHandler(ctx, request, s.handleSearchSuggestions)
}

func (s *MemoryBankServer) handleInitProjectTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.wrapHandler(ctx, request, s.handleInitProject)
}

func (s *MemoryBankServer) handleGetProjectTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.wrapHandler(ctx, request, s.handleGetProject)
}

func (s *MemoryBankServer) handleListProjectsTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.wrapHandler(ctx, request, s.handleListProjects)
}

func (s *MemoryBankServer) handleStartSessionTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.wrapHandler(ctx, request, s.handleStartSession)
}

func (s *MemoryBankServer) handleLogSessionTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.wrapHandler(ctx, request, s.handleLogSession)
}

func (s *MemoryBankServer) handleCompleteSessionTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.wrapHandler(ctx, request, s.handleCompleteSession)
}

func (s *MemoryBankServer) handleGetSessionTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.wrapHandler(ctx, request, s.handleGetSession)
}