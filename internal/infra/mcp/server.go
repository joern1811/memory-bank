package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/infra/embedding"
	"github.com/joern1811/memory-bank/internal/infra/vector"
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

	// Register project integration guide resource
	projectGuideResource := mcp.NewResource(
		"guide://memory-bank/project-setup",
		"Memory Bank Project Integration Guide",
		mcp.WithResourceDescription("Dynamic project-specific CLAUDE.md integration guide with intelligent project analysis"),
		mcp.WithMIMEType("text/plain"),
	)

	mcpServer.AddResource(projectGuideResource, s.handleProjectGuideResource)

	// Register static integration guide resource
	staticGuideResource := mcp.NewResource(
		"guide://memory-bank/claude-integration",
		"Memory Bank CLAUDE.md Integration Templates",
		mcp.WithResourceDescription("Static templates and instructions for integrating Memory Bank with Claude Desktop and Claude Code"),
		mcp.WithMIMEType("text/plain"),
	)

	mcpServer.AddResource(staticGuideResource, s.handleStaticGuideResource)

	// Register debugging session starter prompt
	debuggingSessionPrompt := mcp.NewPrompt(
		"start-debugging-session",
		mcp.WithPromptDescription("Start a structured debugging session with memory-guided problem solving"),
		mcp.WithArgument("error_message", 
			mcp.ArgumentDescription("The error message or problem description you're trying to debug"),
			mcp.RequiredArgument()),
		mcp.WithArgument("context", 
			mcp.ArgumentDescription("Additional context about when/how the error occurs (optional)")),
	)

	mcpServer.AddPrompt(debuggingSessionPrompt, s.handleStartDebuggingSessionPrompt)

	// Register memory pattern creation prompt
	memoryPatternPrompt := mcp.NewPrompt(
		"create-memory-pattern",
		mcp.WithPromptDescription("Store a reusable code pattern or solution in Memory Bank"),
		mcp.WithArgument("pattern_name",
			mcp.ArgumentDescription("Name for the pattern (e.g., 'JWT Authentication Middleware')"),
			mcp.RequiredArgument()),
		mcp.WithArgument("pattern_code",
			mcp.ArgumentDescription("The code or implementation details"),
			mcp.RequiredArgument()),
		mcp.WithArgument("use_case",
			mcp.ArgumentDescription("When and why to use this pattern")),
	)

	mcpServer.AddPrompt(memoryPatternPrompt, s.handleCreateMemoryPatternPrompt)

	// Register solution search prompt
	searchSolutionsPrompt := mcp.NewPrompt(
		"search-solutions",
		mcp.WithPromptDescription("Find similar solutions and patterns in your Memory Bank"),
		mcp.WithArgument("problem_description",
			mcp.ArgumentDescription("Describe the problem or feature you're working on"),
			mcp.RequiredArgument()),
		mcp.WithArgument("technology",
			mcp.ArgumentDescription("Specific technology or framework (optional)")),
	)

	mcpServer.AddPrompt(searchSolutionsPrompt, s.handleSearchSolutionsPrompt)

	// Register session review prompt
	sessionReviewPrompt := mcp.NewPrompt(
		"session-review",
		mcp.WithPromptDescription("Review your current development session and get guidance on next steps"),
	)

	mcpServer.AddPrompt(sessionReviewPrompt, s.handleSessionReviewPrompt)

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

	mcpServer.AddTool(mcp.NewTool("session_list",
		mcp.WithDescription("List sessions with optional filters"),
		mcp.WithString("project_id", mcp.Description("Project ID to filter by")),
		mcp.WithString("status", mcp.Description("Session status to filter by")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results")),
	), s.handleListSessionsTool)

	mcpServer.AddTool(mcp.NewTool("session_abort",
		mcp.WithDescription("Abort active sessions"),
		mcp.WithString("project_id", mcp.Description("Project ID"), mcp.Required()),
		mcp.WithString("session_id", mcp.Description("Specific session ID to abort")),
	), s.handleAbortSessionTool)

	// Version tool
	mcpServer.AddTool(mcp.NewTool("version",
		mcp.WithDescription("Get Memory Bank version information"),
	), s.handleVersionTool)

	// System health tool
	mcpServer.AddTool(mcp.NewTool("system_health",
		mcp.WithDescription("Check system health and service connectivity"),
		mcp.WithBoolean("verbose", mcp.Description("Include detailed service information")),
	), s.handleSystemHealthTool)

	s.logger.Info("MCP tools and resources registered successfully")
}

// CreateMemoryRequest represents a request to create a new memory
type CreateMemoryRequest struct {
	ProjectID string                 `json:"project_id"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Tags      []string               `json:"tags,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	SessionID *string                `json:"session_id,omitempty"`
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
	Query     string   `json:"query"`
	ProjectID *string  `json:"project_id,omitempty"`
	Type      *string  `json:"type,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Limit     *int     `json:"limit,omitempty"`
	Threshold *float32 `json:"threshold,omitempty"`
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

// Session-related request/response structures
type StartSessionRequest struct {
	Title       string  `json:"title"`
	ProjectID   string  `json:"project_id"`
	Description *string `json:"description,omitempty"`
}

type StartSessionResponse struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *MemoryBankServer) handleStartSession(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling session/start request")

	var req StartSessionRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}

	// Build service request
	description := req.Title
	if req.Description != nil {
		description = *req.Description
	}

	serviceReq := ports.StartSessionRequest{
		ProjectID:       domain.ProjectID(req.ProjectID),
		TaskDescription: description,
	}

	// Start session
	session, err := s.sessionService.StartSession(ctx, serviceReq)
	if err != nil {
		s.logger.WithError(err).Error("Failed to start session")
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	response := StartSessionResponse{
		ID:          string(session.ID),
		ProjectID:   string(session.ProjectID),
		Title:       req.Title,
		Description: description,
		Status:      string(session.Status),
		CreatedAt:   session.StartTime,
	}

	s.logger.WithFields(logrus.Fields{
		"session_id": session.ID,
		"project_id": session.ProjectID,
	}).Info("Session started successfully")

	return response, nil
}

type LogSessionRequest struct {
	Message   string  `json:"message"`
	ProjectID *string `json:"project_id,omitempty"`
	SessionID *string `json:"session_id,omitempty"`
}

type LogSessionResponse struct {
	Success   bool   `json:"success"`
	SessionID string `json:"session_id"`
}

func (s *MemoryBankServer) handleLogSession(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling session/log request")

	var req LogSessionRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	if req.Message == "" {
		return nil, fmt.Errorf("message is required")
	}

	// Determine session ID
	var sessionID domain.SessionID
	if req.SessionID != nil {
		sessionID = domain.SessionID(*req.SessionID)
	} else if req.ProjectID != nil {
		// Find active session for project
		projectID := domain.ProjectID(*req.ProjectID)
		activeSession, err := s.sessionService.GetActiveSession(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("no active session found for project: %w", err)
		}
		sessionID = activeSession.ID
	} else {
		return nil, fmt.Errorf("either session_id or project_id is required")
	}

	// Log progress
	if err := s.sessionService.LogProgress(ctx, sessionID, req.Message); err != nil {
		s.logger.WithError(err).Error("Failed to log session progress")
		return nil, fmt.Errorf("failed to log session progress: %w", err)
	}

	response := LogSessionResponse{
		Success:   true,
		SessionID: string(sessionID),
	}

	s.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"message":    req.Message,
	}).Info("Session progress logged")

	return response, nil
}

type CompleteSessionRequest struct {
	Outcome   string  `json:"outcome"`
	ProjectID *string `json:"project_id,omitempty"`
	SessionID *string `json:"session_id,omitempty"`
}

type CompleteSessionResponse struct {
	Success   bool      `json:"success"`
	SessionID string    `json:"session_id"`
	Duration  string    `json:"duration"`
	EndedAt   time.Time `json:"ended_at"`
}

func (s *MemoryBankServer) handleCompleteSession(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling session/complete request")

	var req CompleteSessionRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	if req.Outcome == "" {
		return nil, fmt.Errorf("outcome is required")
	}

	// Determine session ID
	var sessionID domain.SessionID
	if req.SessionID != nil {
		sessionID = domain.SessionID(*req.SessionID)
	} else if req.ProjectID != nil {
		// Find active session for project
		projectID := domain.ProjectID(*req.ProjectID)
		activeSession, err := s.sessionService.GetActiveSession(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("no active session found for project: %w", err)
		}
		sessionID = activeSession.ID
	} else {
		return nil, fmt.Errorf("either session_id or project_id is required")
	}

	// Complete session
	if err := s.sessionService.CompleteSession(ctx, sessionID, req.Outcome); err != nil {
		s.logger.WithError(err).Error("Failed to complete session")
		return nil, fmt.Errorf("failed to complete session: %w", err)
	}

	// Get updated session for response
	session, err := s.sessionService.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated session: %w", err)
	}

	var duration string
	if session.EndTime != nil {
		duration = session.Duration().String()
	}

	response := CompleteSessionResponse{
		Success:   true,
		SessionID: string(sessionID),
		Duration:  duration,
		EndedAt:   *session.EndTime,
	}

	s.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"outcome":    req.Outcome,
		"duration":   duration,
	}).Info("Session completed successfully")

	return response, nil
}

type GetSessionRequest struct {
	ID string `json:"id"`
}

type GetSessionResponse struct {
	ID          string                   `json:"id"`
	ProjectID   string                   `json:"project_id"`
	Title       string                   `json:"title"`
	Description string                   `json:"description"`
	Status      string                   `json:"status"`
	Progress    []map[string]interface{} `json:"progress"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
	EndedAt     *time.Time               `json:"ended_at,omitempty"`
	Duration    *string                  `json:"duration,omitempty"`
}

func (s *MemoryBankServer) handleGetSession(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling session/get request")

	var req GetSessionRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	if req.ID == "" {
		return nil, fmt.Errorf("id is required")
	}

	sessionID := domain.SessionID(req.ID)
	session, err := s.sessionService.GetSession(ctx, sessionID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get session")
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	response := GetSessionResponse{
		ID:          string(session.ID),
		ProjectID:   string(session.ProjectID),
		Title:       session.TaskDescription,
		Description: session.Summary,
		Status:      string(session.Status),
		Progress:    make([]map[string]interface{}, len(session.Progress)),
		CreatedAt:   session.StartTime,
		UpdatedAt:   session.StartTime, // Using StartTime as fallback
	}

	// Convert progress entries
	for i, entry := range session.Progress {
		response.Progress[i] = map[string]interface{}{
			"timestamp": entry.Timestamp,
			"type":      entry.Type,
			"content":   entry.Message,
		}
	}

	// Add optional fields
	if session.EndTime != nil {
		response.EndedAt = session.EndTime
		duration := session.Duration().String()
		response.Duration = &duration
	}

	s.logger.WithField("session_id", sessionID).Info("Session retrieved successfully")

	return response, nil
}

// Advanced Search Features

// FacetedSearchRequest represents a request for faceted search
type FacetedSearchRequest struct {
	Query         string         `json:"query"`
	ProjectID     *string        `json:"project_id,omitempty"`
	Filters       *SearchFilters `json:"filters,omitempty"`
	Limit         *int           `json:"limit,omitempty"`
	Threshold     *float32       `json:"threshold,omitempty"`
	IncludeFacets *bool          `json:"include_facets,omitempty"`
	SortBy        *SortOption    `json:"sort_by,omitempty"`
}

// SearchFilters represents comprehensive search filters
type SearchFilters struct {
	Types      []string    `json:"types,omitempty"`
	Tags       []string    `json:"tags,omitempty"`
	SessionIDs []string    `json:"session_ids,omitempty"`
	TimeFilter *TimeFilter `json:"time_filter,omitempty"`
	HasContent *bool       `json:"has_content,omitempty"`
	MinLength  *int        `json:"min_length,omitempty"`
	MaxLength  *int        `json:"max_length,omitempty"`
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

// ListSessionsRequest represents a request to list sessions
type ListSessionsRequest struct {
	ProjectID *string `json:"project_id,omitempty"`
	Status    *string `json:"status,omitempty"`
	Limit     *int    `json:"limit,omitempty"`
}

// ListSessionsResponse represents the response from listing sessions
type ListSessionsResponse struct {
	Sessions []GetSessionResponse `json:"sessions"`
	Total    int                  `json:"total"`
}

func (s *MemoryBankServer) handleListSessions(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling session/list request")

	var req ListSessionsRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	// Set default limit
	limit := 20
	if req.Limit != nil {
		limit = *req.Limit
	}

	// Build filters
	filters := ports.SessionFilters{
		Limit: limit,
	}

	if req.ProjectID != nil {
		projectID := domain.ProjectID(*req.ProjectID)
		filters.ProjectID = &projectID
	}

	if req.Status != nil {
		status := domain.SessionStatus(*req.Status)
		filters.Status = &status
	}

	// List sessions
	sessions, err := s.sessionService.ListSessions(ctx, filters)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list sessions")
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	// Convert to response format
	sessionResponses := make([]GetSessionResponse, len(sessions))
	for i, session := range sessions {
		sessionResponse := GetSessionResponse{
			ID:          string(session.ID),
			ProjectID:   string(session.ProjectID),
			Title:       session.TaskDescription,
			Description: session.Summary,
			Status:      string(session.Status),
			Progress:    make([]map[string]interface{}, len(session.Progress)),
			CreatedAt:   session.StartTime,
			UpdatedAt:   session.StartTime,
		}

		// Convert progress entries
		for j, entry := range session.Progress {
			sessionResponse.Progress[j] = map[string]interface{}{
				"timestamp": entry.Timestamp,
				"type":      entry.Type,
				"content":   entry.Message,
			}
		}

		// Add optional fields
		if session.EndTime != nil {
			sessionResponse.EndedAt = session.EndTime
			duration := session.Duration().String()
			sessionResponse.Duration = &duration
		}

		sessionResponses[i] = sessionResponse
	}

	response := ListSessionsResponse{
		Sessions: sessionResponses,
		Total:    len(sessionResponses),
	}

	s.logger.WithFields(logrus.Fields{
		"count":      len(sessions),
		"project_id": req.ProjectID,
		"status":     req.Status,
	}).Info("Sessions listed successfully")

	return response, nil
}

func (s *MemoryBankServer) handleListSessionsTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.wrapHandler(ctx, request, s.handleListSessions)
}

// AbortSessionRequest represents a request to abort sessions
type AbortSessionRequest struct {
	ProjectID string  `json:"project_id"`
	SessionID *string `json:"session_id,omitempty"`
}

// AbortSessionResponse represents the response from aborting sessions
type AbortSessionResponse struct {
	Success    bool     `json:"success"`
	AbortedIDs []string `json:"aborted_session_ids"`
	Message    string   `json:"message"`
}

func (s *MemoryBankServer) handleAbortSession(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling session/abort request")

	var req AbortSessionRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request parameters: %w", err)
	}

	if req.ProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}

	projectID := domain.ProjectID(req.ProjectID)
	var abortedIDs []string

	if req.SessionID != nil {
		// Abort specific session
		sessionID := domain.SessionID(*req.SessionID)
		err := s.sessionService.AbortSession(ctx, sessionID)
		if err != nil {
			s.logger.WithError(err).Error("Failed to abort session")
			return nil, fmt.Errorf("failed to abort session: %w", err)
		}
		abortedIDs = []string{string(sessionID)}
	} else {
		// Abort all active sessions for project
		abortedSessionIDs, err := s.sessionService.AbortActiveSessionsForProject(ctx, projectID)
		if err != nil {
			s.logger.WithError(err).Error("Failed to abort active sessions")
			return nil, fmt.Errorf("failed to abort active sessions: %w", err)
		}

		for _, id := range abortedSessionIDs {
			abortedIDs = append(abortedIDs, string(id))
		}
	}

	var message string
	if len(abortedIDs) == 0 {
		message = "No active sessions found to abort"
	} else if len(abortedIDs) == 1 {
		message = "Successfully aborted 1 session"
	} else {
		message = fmt.Sprintf("Successfully aborted %d sessions", len(abortedIDs))
	}

	response := AbortSessionResponse{
		Success:    true,
		AbortedIDs: abortedIDs,
		Message:    message,
	}

	s.logger.WithFields(logrus.Fields{
		"project_id":    req.ProjectID,
		"session_id":    req.SessionID,
		"aborted_count": len(abortedIDs),
	}).Info("Session abort completed")

	return response, nil
}

func (s *MemoryBankServer) handleAbortSessionTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.wrapHandler(ctx, request, s.handleAbortSession)
}

// VersionResponse represents the version information
type VersionResponse struct {
	Version     string `json:"version"`
	BuildTime   string `json:"build_time"`
	GoVersion   string `json:"go_version"`
	GitCommit   string `json:"git_commit"`
	Application string `json:"application"`
}

// HealthServiceStatus represents the health status of a service
type HealthServiceStatus struct {
	Service      string                 `json:"service"`
	Status       string                 `json:"status"`
	Available    bool                   `json:"available"`
	ResponseTime string                 `json:"response_time"`
	Details      map[string]interface{} `json:"details,omitempty"`
	Error        string                 `json:"error,omitempty"`
}

// SystemHealthRequest represents the request for system health
type SystemHealthRequest struct {
	Verbose bool `json:"verbose,omitempty"`
}

// SystemHealthResponse represents the overall system health
type SystemHealthResponse struct {
	Overall       string                 `json:"overall"`
	Timestamp     string                 `json:"timestamp"`
	Services      []HealthServiceStatus  `json:"services"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
}

func (s *MemoryBankServer) handleVersion(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling version request")

	// Get version information (these would typically be set at build time)
	response := VersionResponse{
		Version:     "1.0.0", // This would be injected at build time
		BuildTime:   "dev",   // This would be injected at build time
		GoVersion:   "go1.21+",
		GitCommit:   "dev", // This would be injected at build time
		Application: "Memory Bank MCP Server",
	}

	s.logger.Info("Version information retrieved")
	return response, nil
}

func (s *MemoryBankServer) handleVersionTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.wrapHandler(ctx, request, s.handleVersion)
}

func (s *MemoryBankServer) handleSystemHealth(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.logger.Debug("Handling system health check request")

	var req SystemHealthRequest
	if len(params) > 0 {
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid request parameters: %w", err)
		}
	}

	// Perform health checks
	health := s.checkSystemHealth(ctx, req.Verbose)

	s.logger.WithField("overall_status", health.Overall).Info("System health check completed")
	return health, nil
}

func (s *MemoryBankServer) handleSystemHealthTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.wrapHandler(ctx, request, s.handleSystemHealth)
}

func (s *MemoryBankServer) checkSystemHealth(ctx context.Context, verbose bool) *SystemHealthResponse {
	health := &SystemHealthResponse{
		Timestamp: time.Now().Format(time.RFC3339),
		Services:  make([]HealthServiceStatus, 0),
	}

	// Add configuration if verbose
	if verbose {
		health.Configuration = map[string]interface{}{
			"ollama_base_url":      getEnvOrDefault("OLLAMA_BASE_URL", "http://localhost:11434"),
			"ollama_model":         getEnvOrDefault("OLLAMA_MODEL", "nomic-embed-text"),
			"chromadb_base_url":    getEnvOrDefault("CHROMADB_BASE_URL", "http://localhost:8000"),
			"chromadb_collection":  getEnvOrDefault("CHROMADB_COLLECTION", "memory_bank"),
			"chromadb_data_path":   getEnvOrDefault("CHROMADB_DATA_PATH", "./chromadb_data"),
			"chromadb_auto_start":  getEnvOrDefault("CHROMADB_AUTO_START", "false"),
			"database_path":        getEnvOrDefault("MEMORY_BANK_DB_PATH", "./memory_bank.db"),
		}
	}

	// Check Ollama health
	ollamaStatus := s.checkOllamaHealth(ctx, verbose)
	health.Services = append(health.Services, ollamaStatus)

	// Check ChromaDB health
	chromaStatus := s.checkChromaDBHealth(ctx, verbose)
	health.Services = append(health.Services, chromaStatus)

	// Check database health (simplified check)
	dbStatus := s.checkDatabaseHealth(ctx, verbose)
	health.Services = append(health.Services, dbStatus)

	// Determine overall status
	allHealthy := true
	for _, service := range health.Services {
		if !service.Available {
			allHealthy = false
			break
		}
	}

	if allHealthy {
		health.Overall = "healthy"
	} else {
		health.Overall = "degraded"
	}

	return health
}

func (s *MemoryBankServer) checkOllamaHealth(ctx context.Context, verbose bool) HealthServiceStatus {
	status := HealthServiceStatus{
		Service: "ollama",
		Status:  "unknown",
	}

	// Create Ollama provider for health checking
	ollamaConfig := embedding.DefaultOllamaConfig()
	if baseURL := os.Getenv("OLLAMA_BASE_URL"); baseURL != "" {
		ollamaConfig.BaseURL = baseURL
	}
	if model := os.Getenv("OLLAMA_MODEL"); model != "" {
		ollamaConfig.Model = model
	}

	ollamaProvider := embedding.NewOllamaProvider(ollamaConfig, s.logger)

	// Measure response time
	start := time.Now()
	err := ollamaProvider.HealthCheck(ctx)
	responseTime := time.Since(start)

	status.ResponseTime = responseTime.String()

	if err != nil {
		status.Status = "unhealthy"
		status.Available = false
		status.Error = err.Error()
		if verbose {
			status.Details = map[string]interface{}{
				"base_url": ollamaConfig.BaseURL,
				"model":    ollamaConfig.Model,
				"fallback": "mock provider",
			}
		}
	} else {
		status.Status = "healthy"
		status.Available = true
		if verbose {
			status.Details = map[string]interface{}{
				"base_url":   ollamaConfig.BaseURL,
				"model":      ollamaConfig.Model,
				"dimensions": ollamaProvider.GetDimensions(),
			}
		}
	}

	return status
}

func (s *MemoryBankServer) checkChromaDBHealth(ctx context.Context, verbose bool) HealthServiceStatus {
	status := HealthServiceStatus{
		Service: "chromadb",
		Status:  "unknown",
	}

	// Create ChromaDB store for health checking
	chromaConfig := vector.DefaultChromeDBConfig()
	if baseURL := os.Getenv("CHROMADB_BASE_URL"); baseURL != "" {
		chromaConfig.BaseURL = baseURL
	}
	if collection := os.Getenv("CHROMADB_COLLECTION"); collection != "" {
		chromaConfig.Collection = collection
	}
	if dataPath := os.Getenv("CHROMADB_DATA_PATH"); dataPath != "" {
		chromaConfig.DataPath = dataPath
	}
	if autoStart := os.Getenv("CHROMADB_AUTO_START"); autoStart == "true" {
		chromaConfig.AutoStart = true
	}

	chromaStore := vector.NewChromaDBVectorStore(chromaConfig, s.logger)

	// Measure response time
	start := time.Now()
	err := chromaStore.HealthCheck(ctx)
	responseTime := time.Since(start)

	status.ResponseTime = responseTime.String()

	if err != nil {
		status.Status = "unhealthy"
		status.Available = false
		status.Error = err.Error()
		if verbose {
			status.Details = map[string]interface{}{
				"base_url":   chromaConfig.BaseURL,
				"collection": chromaConfig.Collection,
				"tenant":     chromaConfig.Tenant,
				"database":   chromaConfig.Database,
				"data_path":  chromaConfig.DataPath,
				"auto_start": chromaConfig.AutoStart,
				"fallback":   "mock vector store",
			}
		}
	} else {
		status.Status = "healthy"
		status.Available = true
		if verbose {
			details := map[string]interface{}{
				"base_url":   chromaConfig.BaseURL,
				"collection": chromaConfig.Collection,
				"tenant":     chromaConfig.Tenant,
				"database":   chromaConfig.Database,
				"data_path":  chromaConfig.DataPath,
				"auto_start": chromaConfig.AutoStart,
			}

			// Try to get additional details
			if collections, err := chromaStore.ListCollections(ctx); err == nil {
				details["available_collections"] = collections
				details["collections_count"] = len(collections)
			}

			status.Details = details
		}
	}

	return status
}

func (s *MemoryBankServer) checkDatabaseHealth(ctx context.Context, verbose bool) HealthServiceStatus {
	status := HealthServiceStatus{
		Service: "database",
		Status:  "unknown",
	}

	// Simple database health check by trying to list projects
	start := time.Now()
	_, err := s.projectService.ListProjects(ctx)
	responseTime := time.Since(start)

	status.ResponseTime = responseTime.String()

	if err != nil {
		status.Status = "unhealthy"
		status.Available = false
		status.Error = err.Error()
		if verbose {
			status.Details = map[string]interface{}{
				"path": getEnvOrDefault("MEMORY_BANK_DB_PATH", "./memory_bank.db"),
				"type": "sqlite",
			}
		}
	} else {
		status.Status = "healthy"
		status.Available = true
		if verbose {
			status.Details = map[string]interface{}{
				"path": getEnvOrDefault("MEMORY_BANK_DB_PATH", "./memory_bank.db"),
				"type": "sqlite",
			}
		}
	}

	return status
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// handleProjectGuideResource handles requests for the dynamic project integration guide resource
func (s *MemoryBankServer) handleProjectGuideResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	s.logger.Debug("Handling project integration guide resource request")

	// Generate dynamic project-specific integration guide
	guide, err := s.generateProjectIntegrationGuide(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate project integration guide")
		return nil, fmt.Errorf("failed to generate project integration guide: %w", err)
	}

	// Create text resource contents
	textContent := mcp.TextResourceContents{
		URI:      request.Params.URI,
		MIMEType: "text/plain",
		Text:     guide,
	}

	s.logger.Info("Project integration guide resource generated successfully")
	return []mcp.ResourceContents{textContent}, nil
}

// handleStaticGuideResource handles requests for the static integration guide resource
func (s *MemoryBankServer) handleStaticGuideResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	s.logger.Debug("Handling static integration guide resource request")

	// Generate static integration guide with templates
	guide := s.generateStaticIntegrationGuide()

	// Create text resource contents
	textContent := mcp.TextResourceContents{
		URI:      request.Params.URI,
		MIMEType: "text/plain",
		Text:     guide,
	}

	s.logger.Info("Static integration guide resource generated successfully")
	return []mcp.ResourceContents{textContent}, nil
}

// generateProjectIntegrationGuide creates a dynamic, project-specific integration guide
func (s *MemoryBankServer) generateProjectIntegrationGuide(ctx context.Context) (string, error) {
	// Detect current working directory and analyze project
	currentDir, err := os.Getwd()
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get current directory")
		currentDir = "unknown"
	}

	// Analyze project structure and technology stack
	projectAnalysis := s.analyzeProject(currentDir)
	
	// Get existing project and memories for context
	var existingProject *domain.Project
	var existingMemories []ports.MemorySearchResult
	
	// Try to find existing project by path
	if projects, err := s.projectService.ListProjects(ctx); err == nil {
		for _, project := range projects {
			if project.Path == currentDir {
				existingProject = project
				break
			}
		}
	}

	// Get existing memories if project exists
	if existingProject != nil {
		searchReq := ports.SemanticSearchRequest{
			Query:     "", // Empty query to get all memories
			ProjectID: &existingProject.ID,
			Limit:     100,
			Threshold: 0.0,
		}
		if memories, err := s.memoryService.SearchMemories(ctx, searchReq); err == nil {
			existingMemories = memories
		}
	}

	// Generate customized CLAUDE.md content
	guide := s.buildProjectGuide(currentDir, projectAnalysis, existingProject, existingMemories)
	
	return guide, nil
}

// ProjectAnalysis contains detected project information
type ProjectAnalysis struct {
	ProjectType    string
	TechStack      []string
	HasPackageJSON bool
	HasGoMod       bool
	HasPyProject   bool
	HasDockerfile  bool
	HasMakefile    bool
	SuggestedTags  []string
	MemoryTypes    []string
}

// analyzeProject analyzes the current project directory to detect technology stack
func (s *MemoryBankServer) analyzeProject(projectPath string) ProjectAnalysis {
	analysis := ProjectAnalysis{
		ProjectType:   "unknown",
		TechStack:     []string{},
		SuggestedTags: []string{},
		MemoryTypes:   []string{"decision", "pattern", "error_solution", "code", "documentation"},
	}

	// Check for common project files
	checkFile := func(filename string) bool {
		_, err := os.Stat(filepath.Join(projectPath, filename))
		return err == nil
	}

	// Detect technology stack
	if checkFile("package.json") {
		analysis.HasPackageJSON = true
		analysis.TechStack = append(analysis.TechStack, "javascript", "node")
		analysis.SuggestedTags = append(analysis.SuggestedTags, "javascript", "node", "npm")
		
		// Check for specific frameworks
		if checkFile("next.config.js") || checkFile("next.config.ts") {
			analysis.TechStack = append(analysis.TechStack, "nextjs")
			analysis.SuggestedTags = append(analysis.SuggestedTags, "nextjs", "react", "frontend")
			analysis.ProjectType = "web-app"
		} else if checkFile("src/App.js") || checkFile("src/App.tsx") {
			analysis.TechStack = append(analysis.TechStack, "react")
			analysis.SuggestedTags = append(analysis.SuggestedTags, "react", "frontend")
			analysis.ProjectType = "web-app"
		}
	}

	if checkFile("go.mod") {
		analysis.HasGoMod = true
		analysis.TechStack = append(analysis.TechStack, "go")
		analysis.SuggestedTags = append(analysis.SuggestedTags, "go", "backend")
		
		// Check for specific patterns
		if checkFile("cmd") {
			analysis.ProjectType = "cli-tool"
			analysis.SuggestedTags = append(analysis.SuggestedTags, "cli")
		} else if checkFile("internal") || checkFile("pkg") {
			analysis.ProjectType = "api"
			analysis.SuggestedTags = append(analysis.SuggestedTags, "api", "microservice")
		}
	}

	if checkFile("requirements.txt") || checkFile("pyproject.toml") || checkFile("setup.py") {
		analysis.HasPyProject = true
		analysis.TechStack = append(analysis.TechStack, "python")
		analysis.SuggestedTags = append(analysis.SuggestedTags, "python")
		
		// Check for data science patterns
		if checkFile("notebooks") || checkFile("data") {
			analysis.ProjectType = "data-science"
			analysis.SuggestedTags = append(analysis.SuggestedTags, "data", "ml", "analysis")
		}
	}

	if checkFile("Dockerfile") {
		analysis.HasDockerfile = true
		analysis.TechStack = append(analysis.TechStack, "docker")
		analysis.SuggestedTags = append(analysis.SuggestedTags, "docker", "deployment")
	}

	if checkFile("Makefile") {
		analysis.HasMakefile = true
		analysis.SuggestedTags = append(analysis.SuggestedTags, "build", "automation")
	}

	// Additional infrastructure detection
	if checkFile("terraform") {
		analysis.TechStack = append(analysis.TechStack, "terraform")
		analysis.SuggestedTags = append(analysis.SuggestedTags, "terraform", "infrastructure")
	}

	if checkFile("k8s") || checkFile("kubernetes") {
		analysis.TechStack = append(analysis.TechStack, "kubernetes")
		analysis.SuggestedTags = append(analysis.SuggestedTags, "kubernetes", "k8s", "deployment")
	}

	// Default project type if not detected
	if analysis.ProjectType == "unknown" {
		if len(analysis.TechStack) > 0 {
			analysis.ProjectType = "library"
		} else {
			analysis.ProjectType = "general"
		}
	}

	return analysis
}

// buildProjectGuide constructs the complete project-specific guide
func (s *MemoryBankServer) buildProjectGuide(projectPath string, analysis ProjectAnalysis, existingProject *domain.Project, existingMemories []ports.MemorySearchResult) string {
	var guide strings.Builder

	// Header
	guide.WriteString("# Memory Bank - Project Integration Guide\n\n")
	guide.WriteString("*This guide has been automatically generated based on your current project structure and existing memories.*\n\n")

	// Project Analysis Section
	guide.WriteString("## 📊 Project Analysis\n\n")
	guide.WriteString(fmt.Sprintf("- **Project Path**: `%s`\n", projectPath))
	guide.WriteString(fmt.Sprintf("- **Project Type**: %s\n", analysis.ProjectType))
	guide.WriteString(fmt.Sprintf("- **Technology Stack**: %s\n", strings.Join(analysis.TechStack, ", ")))
	
	if existingProject != nil {
		guide.WriteString(fmt.Sprintf("- **Memory Bank Status**: ✅ Initialized (Project ID: `%s`)\n", string(existingProject.ID)))
		guide.WriteString(fmt.Sprintf("- **Existing Memories**: %d entries\n", len(existingMemories)))
	} else {
		guide.WriteString("- **Memory Bank Status**: ❌ Not initialized\n")
	}
	guide.WriteString("\n")

	// Current Status Section
	if existingProject != nil {
		guide.WriteString("## 📈 Current Memory Status\n\n")
		
		// Analyze existing memories by type
		typeCount := make(map[string]int)
		tagUsage := make(map[string]int)
		
		for _, memory := range existingMemories {
			typeCount[string(memory.Memory.Type)]++
			for _, tag := range memory.Memory.Tags {
				tagUsage[tag]++
			}
		}
		
		if len(typeCount) > 0 {
			guide.WriteString("### Memory Types:\n")
			for memType, count := range typeCount {
				guide.WriteString(fmt.Sprintf("- **%s**: %d entries\n", memType, count))
			}
			guide.WriteString("\n")
		}
		
		if len(tagUsage) > 0 {
			guide.WriteString("### Most Used Tags:\n")
			// Show top 10 tags
			count := 0
			for tag, usage := range tagUsage {
				if count >= 10 {
					break
				}
				guide.WriteString(fmt.Sprintf("- `%s` (%d uses)\n", tag, usage))
				count++
			}
			guide.WriteString("\n")
		}
	}

	// Setup Instructions
	if existingProject == nil {
		guide.WriteString("## 🚀 Quick Setup\n\n")
		guide.WriteString("Since this project is not yet initialized in Memory Bank, start with:\n\n")
		guide.WriteString("```markdown\n")
		guide.WriteString("Initialize this project in Memory Bank:\n")
		guide.WriteString("- Use `project/init` to register this project\n")
		guide.WriteString("- Copy the CLAUDE.md template below to your project root\n")
		guide.WriteString("- Start documenting your first architectural decisions\n")
		guide.WriteString("```\n\n")
	}

	// Customized CLAUDE.md Template
	guide.WriteString("## 📝 Customized CLAUDE.md Template\n\n")
	guide.WriteString("Copy this template to your project's `CLAUDE.md` file:\n\n")
	guide.WriteString("```markdown\n")
	guide.WriteString(s.generateProjectSpecificCLAUDEmd(analysis, existingProject))
	guide.WriteString("```\n\n")

	// Project-Specific Recommendations
	guide.WriteString("## 💡 Project-Specific Recommendations\n\n")
	guide.WriteString(s.generateProjectRecommendations(analysis, existingMemories))

	// Next Steps
	guide.WriteString("## ✅ Next Steps\n\n")
	if existingProject == nil {
		guide.WriteString("1. **Initialize Project**: Use Memory Bank to register this project\n")
		guide.WriteString("2. **Add CLAUDE.md**: Copy the template above to your project root\n")
		guide.WriteString("3. **Document Initial Decisions**: Store your first architectural decisions\n")
		guide.WriteString("4. **Start Development Session**: Use session management for complex tasks\n")
	} else {
		guide.WriteString("1. **Update CLAUDE.md**: Enhance your existing CLAUDE.md with the template above\n")
		guide.WriteString("2. **Review Existing Memories**: Check if any memories need updating\n")
		guide.WriteString("3. **Improve Tagging**: Consider adopting the suggested tagging strategy\n")
		guide.WriteString("4. **Expand Documentation**: Add missing patterns and decisions\n")
	}
	guide.WriteString("\n")

	// Footer
	guide.WriteString("---\n\n")
	guide.WriteString("*This guide was generated at " + time.Now().Format("2006-01-02 15:04:05") + " based on your current project structure.*\n")
	guide.WriteString("*Refresh this resource to get updated recommendations as your project evolves.*\n")

	return guide.String()
}

// generateProjectSpecificCLAUDEmd creates a customized CLAUDE.md template
func (s *MemoryBankServer) generateProjectSpecificCLAUDEmd(analysis ProjectAnalysis, existingProject *domain.Project) string {
	var template strings.Builder

	template.WriteString("# Memory Bank Integration\n\n")
	template.WriteString("## Project Configuration\n\n")
	template.WriteString(fmt.Sprintf("- **Project Type**: %s\n", analysis.ProjectType))
	template.WriteString(fmt.Sprintf("- **Tech Stack**: %s\n", strings.Join(analysis.TechStack, ", ")))
	
	if existingProject != nil {
		template.WriteString(fmt.Sprintf("- **Project ID**: %s\n", string(existingProject.ID)))
	} else {
		template.WriteString("- **Project ID**: [TO BE SET AFTER INITIALIZATION]\n")
	}
	template.WriteString("\n")

	template.WriteString("## Memory Bank Protocol (MANDATORY)\n\n")
	template.WriteString("When Memory Bank MCP tools are available, Claude MUST:\n\n")
	template.WriteString("1. **ALWAYS START** with `memory/search` using project context\n")
	template.WriteString("2. **SEARCH BEFORE IMPLEMENTING** - Check existing patterns and solutions\n")
	template.WriteString("3. **DOCUMENT DECISIONS** - Store all architectural choices with rationale\n")
	template.WriteString("4. **STORE PATTERNS** - Save reusable code patterns and best practices\n")
	template.WriteString("5. **LOG ERRORS** - Document error solutions for future reference\n")
	template.WriteString("6. **USE SESSIONS** - Track progress on complex, multi-step tasks\n\n")

	template.WriteString("## Project-Specific Memory Types\n\n")
	for _, memType := range analysis.MemoryTypes {
		switch memType {
		case "decision":
			template.WriteString("- **Decision**: Architectural choices, technology selections, design patterns\n")
		case "pattern":
			template.WriteString("- **Pattern**: Reusable code templates, design patterns, configuration examples\n")
		case "error_solution":
			template.WriteString("- **ErrorSolution**: Bug fixes, troubleshooting guides, common issues\n")
		case "code":
			template.WriteString("- **Code**: Utility functions, snippets, examples, boilerplate\n")
		case "documentation":
			template.WriteString("- **Documentation**: API docs, setup guides, deployment instructions\n")
		}
	}
	template.WriteString("\n")

	template.WriteString("## Tagging Strategy\n\n")
	template.WriteString("Use these tags consistently:\n\n")
	template.WriteString("### Technology Tags\n")
	for _, tech := range analysis.TechStack {
		template.WriteString(fmt.Sprintf("- `%s`\n", tech))
	}
	template.WriteString("\n")

	template.WriteString("### Domain Tags\n")
	for _, tag := range analysis.SuggestedTags {
		template.WriteString(fmt.Sprintf("- `%s`\n", tag))
	}
	template.WriteString("\n")

	template.WriteString("### Component Tags\n")
	switch analysis.ProjectType {
	case "web-app":
		template.WriteString("- `frontend`, `backend`, `database`, `api`, `ui`, `auth`\n")
	case "cli-tool":
		template.WriteString("- `cli`, `commands`, `config`, `output`, `parsing`\n")
	case "api":
		template.WriteString("- `endpoint`, `middleware`, `validation`, `auth`, `database`\n")
	case "data-science":
		template.WriteString("- `preprocessing`, `model`, `analysis`, `visualization`, `pipeline`\n")
	default:
		template.WriteString("- `core`, `utils`, `config`, `tests`, `docs`\n")
	}
	template.WriteString("\n")

	template.WriteString("## Automatic Behaviors\n\n")
	template.WriteString("Claude MUST automatically:\n\n")
	template.WriteString("1. Search for existing knowledge before implementing features\n")
	template.WriteString("2. Store architectural decisions with alternatives considered\n")
	template.WriteString("3. Document reusable patterns with usage examples\n")
	template.WriteString("4. Save error solutions with prevention strategies\n")
	template.WriteString("5. Use consistent tagging as defined above\n")
	template.WriteString("6. Reference existing memories in explanations\n")
	template.WriteString("7. Ask permission before storing new memories\n\n")

	template.WriteString("## Session Management\n\n")
	template.WriteString("Use sessions for:\n")
	switch analysis.ProjectType {
	case "web-app":
		template.WriteString("- New feature implementation (frontend + backend)\n")
		template.WriteString("- UI/UX improvements and redesigns\n")
		template.WriteString("- Performance optimization tasks\n")
		template.WriteString("- Security enhancements and audits\n")
	case "cli-tool":
		template.WriteString("- New command implementation\n")
		template.WriteString("- Configuration system changes\n")
		template.WriteString("- Output format modifications\n")
		template.WriteString("- Cross-platform compatibility work\n")
	case "api":
		template.WriteString("- New endpoint development\n")
		template.WriteString("- Database schema changes\n")
		template.WriteString("- Authentication/authorization updates\n")
		template.WriteString("- API versioning and migration\n")
	case "data-science":
		template.WriteString("- Model development and training\n")
		template.WriteString("- Data pipeline construction\n")
		template.WriteString("- Analysis and reporting tasks\n")
		template.WriteString("- Visualization and dashboard creation\n")
	default:
		template.WriteString("- Multi-file feature implementations\n")
		template.WriteString("- Complex refactoring tasks\n")
		template.WriteString("- Architecture changes\n")
		template.WriteString("- Integration projects\n")
	}
	template.WriteString("\n")

	template.WriteString("## Search Keywords\n\n")
	template.WriteString("Common search terms for this project:\n")
	keywords := s.generateSearchKeywords(analysis)
	for _, keyword := range keywords {
		template.WriteString(fmt.Sprintf("- \"%s\"\n", keyword))
	}

	return template.String()
}

// generateSearchKeywords creates relevant search keywords based on project analysis
func (s *MemoryBankServer) generateSearchKeywords(analysis ProjectAnalysis) []string {
	keywords := []string{}

	// Base keywords for all projects
	keywords = append(keywords, "architecture", "configuration", "deployment", "testing", "error handling")

	// Project-type specific keywords
	switch analysis.ProjectType {
	case "web-app":
		keywords = append(keywords, "authentication", "routing", "state management", "API integration", "responsive design")
	case "cli-tool":
		keywords = append(keywords, "command parsing", "configuration files", "output formatting", "argument validation")
	case "api":
		keywords = append(keywords, "endpoint design", "middleware", "request validation", "response formatting", "rate limiting")
	case "data-science":
		keywords = append(keywords, "data preprocessing", "model training", "feature engineering", "visualization", "pipeline")
	}

	// Technology-specific keywords
	for _, tech := range analysis.TechStack {
		switch tech {
		case "react":
			keywords = append(keywords, "component design", "hooks", "state management")
		case "go":
			keywords = append(keywords, "goroutines", "channels", "error handling")
		case "python":
			keywords = append(keywords, "virtual environment", "dependencies", "packaging")
		case "docker":
			keywords = append(keywords, "containerization", "image optimization", "multi-stage builds")
		}
	}

	return keywords
}

// generateProjectRecommendations creates specific recommendations based on project analysis
func (s *MemoryBankServer) generateProjectRecommendations(analysis ProjectAnalysis, existingMemories []ports.MemorySearchResult) string {
	var recommendations strings.Builder

	// Technology-specific recommendations
	switch analysis.ProjectType {
	case "web-app":
		recommendations.WriteString("### Web Application Best Practices\n")
		recommendations.WriteString("- Document component architecture and state management patterns\n")
		recommendations.WriteString("- Store API integration patterns and error handling strategies\n")
		recommendations.WriteString("- Save performance optimization techniques and metrics\n")
		recommendations.WriteString("- Record security implementations (auth, validation, sanitization)\n\n")

	case "cli-tool":
		recommendations.WriteString("### CLI Tool Best Practices\n")
		recommendations.WriteString("- Document command structure and argument parsing patterns\n")
		recommendations.WriteString("- Store configuration file formats and validation logic\n")
		recommendations.WriteString("- Save cross-platform compatibility solutions\n")
		recommendations.WriteString("- Record user experience patterns (help text, error messages)\n\n")

	case "api":
		recommendations.WriteString("### API Development Best Practices\n")
		recommendations.WriteString("- Document endpoint design patterns and RESTful conventions\n")
		recommendations.WriteString("- Store middleware implementations and request/response patterns\n")
		recommendations.WriteString("- Save authentication and authorization strategies\n")
		recommendations.WriteString("- Record database integration patterns and query optimizations\n\n")

	case "data-science":
		recommendations.WriteString("### Data Science Best Practices\n")
		recommendations.WriteString("- Document data preprocessing pipelines and transformations\n")
		recommendations.WriteString("- Store model architectures and hyperparameter configurations\n")
		recommendations.WriteString("- Save visualization patterns and reporting templates\n")
		recommendations.WriteString("- Record experiment tracking and reproducibility practices\n\n")
	}

	// Memory gap analysis
	if len(existingMemories) > 0 {
		recommendations.WriteString("### Memory Gap Analysis\n")
		
		typeCount := make(map[string]int)
		for _, memory := range existingMemories {
			typeCount[string(memory.Memory.Type)]++
		}
		
		// Check for missing important memory types
		if typeCount["decision"] == 0 {
			recommendations.WriteString("- ⚠️ **Missing Decisions**: Start documenting architectural decisions\n")
		}
		if typeCount["pattern"] == 0 {
			recommendations.WriteString("- ⚠️ **Missing Patterns**: Begin storing reusable code patterns\n")
		}
		if typeCount["error_solution"] == 0 {
			recommendations.WriteString("- ⚠️ **Missing Error Solutions**: Document troubleshooting knowledge\n")
		}
		recommendations.WriteString("\n")
	}

	// Technology-specific memory suggestions
	for _, tech := range analysis.TechStack {
		switch tech {
		case "go":
			recommendations.WriteString("### Go-Specific Recommendations\n")
			recommendations.WriteString("- Store common error handling patterns\n")
			recommendations.WriteString("- Document goroutine and channel usage patterns\n")
			recommendations.WriteString("- Save interface design decisions\n")
			recommendations.WriteString("- Record testing strategies and mock patterns\n\n")

		case "react":
			recommendations.WriteString("### React-Specific Recommendations\n")
			recommendations.WriteString("- Document component composition patterns\n")
			recommendations.WriteString("- Store custom hook implementations\n")
			recommendations.WriteString("- Save state management decisions (Context, Redux, Zustand)\n")
			recommendations.WriteString("- Record performance optimization techniques\n\n")

		case "python":
			recommendations.WriteString("### Python-Specific Recommendations\n")
			recommendations.WriteString("- Document dependency management strategies\n")
			recommendations.WriteString("- Store common utility functions and decorators\n")
			recommendations.WriteString("- Save testing patterns and fixtures\n")
			recommendations.WriteString("- Record packaging and deployment configurations\n\n")
		}
	}

	return recommendations.String()
}

// generateStaticIntegrationGuide creates the static integration guide with templates
func (s *MemoryBankServer) generateStaticIntegrationGuide() string {
	return `# Memory Bank - CLAUDE.md Integration Templates

This resource provides static templates and instructions for integrating Memory Bank with Claude Desktop and Claude Code. Use these templates as starting points and customize them for your specific projects.

## Global CLAUDE.md Configuration

Add this to your global CLAUDE.md file (~/.config/claude/CLAUDE.md):

` + "```markdown" + `
## Memory Bank Protocol (GLOBAL)

### Automatic Memory Bank Usage
When Memory Bank MCP tools are detected, Claude MUST:

1. **ALWAYS START** with memory/search using project context
2. **AUTOMATICALLY STORE** all decisions, patterns, and solutions
3. **PROACTIVELY REFERENCE** existing knowledge in responses
4. **CONTINUOUSLY UPDATE** memory with new learnings

### Required Workflow
1. **Session Start**: Search existing knowledge (memory/search)
2. **Before Implementation**: Check for existing patterns
3. **During Work**: Document decisions and store solutions
4. **After Completion**: Store lessons learned and update patterns

### Memory Integration Rules
- Reference existing memories in explanations
- Always store architectural decisions with rationale
- Document error solutions with prevention strategies
- Tag consistently using project-specific conventions
- Use sessions for complex multi-step work

### Enforcement
- Never implement without first searching existing knowledge
- Always ask permission before storing new memories
- Proactively suggest memory storage for important decisions
- Reference memory IDs in responses for traceability
` + "```" + `

## Project-Specific Template

Copy this template to your project's CLAUDE.md file and customize:

` + "```markdown" + `
## Memory Bank Integration (PROJECT-SPECIFIC)

### Project Context
- Project Type: [web-app/api/cli-tool/library/data-science]
- Tech Stack: [list your technologies]
- Project ID: [set after project/init]

### Required Memory Types
- **Decisions**: [specific to your project]
- **Patterns**: [reusable patterns for your domain]
- **ErrorSolutions**: [common errors in your tech stack]
- **Code**: [useful snippets for your project]
- **Documentation**: [project-specific docs]

### Tagging Strategy
Required tags for this project:
- Technology: [your tech stack tags]
- Domain: [your domain-specific tags]
- Component: [your architecture components]
- Complexity: simple, medium, complex

### Search Keywords
Common search terms for this project:
- [domain-specific terminology]
- [technology-specific patterns]
- [architecture concepts]

### Automatic Behaviors
When working on this project, Claude MUST:
1. Search for existing implementations before coding
2. Store all architectural decisions
3. Document reusable patterns
4. Save error solutions with prevention strategies
5. Use consistent tagging strategy
6. Reference existing memories in explanations

### Session Management
Use sessions for:
- Multi-file feature implementations
- Complex refactoring tasks
- Bug investigation and resolution
- Architecture planning and implementation
` + "```" + `

## Technology-Specific Templates

### Web Application Template

` + "```markdown" + `
### Web App Memory Configuration
- **Decisions**: Framework choice, state management, routing, deployment
- **Patterns**: Component patterns, API integration, authentication flows
- **ErrorSolutions**: CORS issues, state bugs, performance problems
- **Code**: Utility hooks, validation schemas, middleware functions

### Tagging Strategy
- Technology: react, javascript, typescript, html, css
- Domain: frontend, backend, api, database, auth
- Component: component, service, util, config, test

### Session Usage
- Feature development (frontend + backend)
- UI/UX improvements
- Performance optimization
- Security implementation
` + "```" + `

### CLI Tool Template

` + "```markdown" + `
### CLI Tool Memory Configuration
- **Decisions**: Command structure, configuration format, output design
- **Patterns**: Argument parsing, command patterns, plugin architecture
- **ErrorSolutions**: Cross-platform issues, dependency problems, user errors
- **Code**: Command handlers, utility functions, configuration validators

### Tagging Strategy
- Technology: [go/python/rust/etc], cli, terminal
- Domain: commands, config, output, validation
- Component: cmd, util, config, parser, formatter

### Session Usage
- New command implementation
- Configuration system changes
- Cross-platform compatibility
- Plugin development
` + "```" + `

### API Development Template

` + "```markdown" + `
### API Memory Configuration
- **Decisions**: Framework choice, database design, authentication method
- **Patterns**: Endpoint patterns, middleware, validation, error handling
- **ErrorSolutions**: Database connection issues, validation errors, auth problems
- **Code**: Middleware functions, validation schemas, utility functions

### Tagging Strategy
- Technology: [fastapi/express/gin/etc], api, rest, graphql
- Domain: endpoint, middleware, auth, database, validation
- Component: handler, service, model, middleware, util

### Session Usage
- Endpoint development
- Database schema changes
- Authentication implementation
- API versioning
` + "```" + `

### Data Science Template

` + "```markdown" + `
### Data Science Memory Configuration
- **Decisions**: Model architecture, feature selection, evaluation metrics
- **Patterns**: Data pipelines, preprocessing steps, visualization templates
- **ErrorSolutions**: Data quality issues, model performance problems, pipeline failures
- **Code**: Preprocessing functions, model utilities, visualization scripts

### Tagging Strategy
- Technology: python, pandas, sklearn, tensorflow, pytorch
- Domain: data, model, preprocessing, analysis, visualization
- Component: pipeline, model, feature, metric, plot

### Session Usage
- Model development
- Data pipeline creation
- Analysis projects
- Visualization development
` + "```" + `

## Implementation Checklist

### 1. Global Setup
- [ ] Add Memory Bank protocol to global CLAUDE.md
- [ ] Configure MCP server in Claude Desktop/Code
- [ ] Test connection with memory/search

### 2. Project Setup
- [ ] Initialize project with project/init
- [ ] Copy and customize project template
- [ ] Document initial architectural decisions
- [ ] Set up tagging strategy

### 3. Usage Workflow
- [ ] Start sessions with memory search
- [ ] Store decisions as they're made
- [ ] Document patterns when discovered
- [ ] Save error solutions immediately
- [ ] Use consistent tagging

### 4. Maintenance
- [ ] Weekly: Review and update memories
- [ ] Monthly: Clean up duplicate entries
- [ ] Quarterly: Reorganize tag structure
- [ ] Continuously: Improve search strategies

## Best Practices

### Memory Creation
- Use descriptive, searchable titles
- Include sufficient context for future use
- Structure content with markdown formatting
- Link to related memories when relevant
- Tag consistently and meaningfully

### Search Strategies
- Use semantic queries over exact matches
- Start broad, then narrow with filters
- Leverage faceted search for complex queries
- Use suggestions for discovery
- Reference memory IDs in documentation

### Session Management
- Start sessions for multi-step work
- Log progress at meaningful milestones
- Complete sessions with outcome summaries
- Use sessions to maintain context

### Tag Management
- Establish consistent tagging conventions
- Use hierarchical tags when appropriate
- Review and standardize tags regularly
- Include both technical and domain tags
- Maintain a project tag glossary

---

*For dynamic, project-specific integration guides, use the guide://memory-bank/project-setup resource.*`
}

// handleStartDebuggingSessionPrompt handles the debugging session starter prompt
func (s *MemoryBankServer) handleStartDebuggingSessionPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	s.logger.Debug("Handling start debugging session prompt request")

	errorMessage := request.Params.Arguments["error_message"]
	context := request.Params.Arguments["context"]

	// Build the conversation starter
	promptText := fmt.Sprintf(`I need help debugging this issue:

**Error:** %s`, errorMessage)

	if context != "" {
		promptText += fmt.Sprintf(`

**Context:** %s`, context)
	}

	promptText += `

Please help me:
1. Search for similar errors in my Memory Bank
2. Start a debugging session to track our progress
3. Guide me through a systematic debugging process
4. Document any solutions we find for future reference

Let's start by searching for similar issues and then create a debugging session.`

	result := &mcp.GetPromptResult{
		Description: "Start a structured debugging session with memory-guided problem solving",
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: promptText,
				},
			},
		},
	}

	s.logger.Info("Start debugging session prompt generated successfully")
	return result, nil
}

// handleCreateMemoryPatternPrompt handles the create memory pattern prompt
func (s *MemoryBankServer) handleCreateMemoryPatternPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	s.logger.Debug("Handling create memory pattern prompt request")

	patternName := request.Params.Arguments["pattern_name"]
	patternCode := request.Params.Arguments["pattern_code"]
	useCase := request.Params.Arguments["use_case"]

	promptText := fmt.Sprintf(`I want to store this code pattern in my Memory Bank:

**Pattern Name:** %s

**Code/Implementation:**
` + "```" + `
%s
` + "```", patternName, patternCode)

	if useCase != "" {
		promptText += fmt.Sprintf(`

**Use Case:** %s`, useCase)
	}

	promptText += `

Please help me:
1. Store this pattern in Memory Bank with appropriate tags
2. Suggest related patterns that might exist
3. Recommend when and how to use this pattern
4. Add it to my project's pattern library

Let's create this memory entry and organize it properly.`

	result := &mcp.GetPromptResult{
		Description: "Store a reusable code pattern or solution in Memory Bank",
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: promptText,
				},
			},
		},
	}

	s.logger.Info("Create memory pattern prompt generated successfully")
	return result, nil
}

// handleSearchSolutionsPrompt handles the search solutions prompt
func (s *MemoryBankServer) handleSearchSolutionsPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	s.logger.Debug("Handling search solutions prompt request")

	problemDescription := request.Params.Arguments["problem_description"]
	technology := request.Params.Arguments["technology"]

	promptText := fmt.Sprintf(`I'm working on this problem and need to find relevant solutions or patterns:

**Problem/Feature:** %s`, problemDescription)

	if technology != "" {
		promptText += fmt.Sprintf(`
**Technology:** %s`, technology)
	}

	promptText += `

Please help me:
1. Search my Memory Bank for similar problems and solutions
2. Find relevant code patterns that might apply
3. Look for architectural decisions that relate to this
4. Suggest implementation approaches based on what we find

Let's start by searching for related memories and patterns.`

	result := &mcp.GetPromptResult{
		Description: "Find similar solutions and patterns in your Memory Bank",
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: promptText,
				},
			},
		},
	}

	s.logger.Info("Search solutions prompt generated successfully")
	return result, nil
}

// handleSessionReviewPrompt handles the session review prompt
func (s *MemoryBankServer) handleSessionReviewPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	s.logger.Debug("Handling session review prompt request")

	promptText := `I'd like to review my current development session and get guidance on next steps.

Please help me:
1. Review my active development session and recent progress
2. Analyze what I've accomplished so far
3. Identify any patterns or decisions I should document
4. Suggest logical next steps based on my current work
5. Recommend what memories I should create from this session

Let's look at my session history and plan the next phase of development.`

	result := &mcp.GetPromptResult{
		Description: "Review your current development session and get guidance on next steps",
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: promptText,
				},
			},
		},
	}

	s.logger.Info("Session review prompt generated successfully")
	return result, nil
}

// Helper function for min operation
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
