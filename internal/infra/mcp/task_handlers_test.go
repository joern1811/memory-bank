package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/joern1811/memory-bank/internal/app"
	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/infra/database"
	"github.com/joern1811/memory-bank/internal/infra/embedding"
	"github.com/joern1811/memory-bank/internal/infra/vector"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"
)

func TestTaskHandlers_CreateTask(t *testing.T) {
	// Setup test services
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Use in-memory database
	db, err := database.NewSQLiteDatabase(":memory:", logger)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Initialize repositories
	memoryRepo := database.NewSQLiteMemoryRepository(db, logger)
	projectRepo := database.NewSQLiteProjectRepository(db, logger)

	// Use mock providers for testing
	embeddingProvider := embedding.NewMockEmbeddingProvider(768, logger)
	vectorStore := vector.NewMockVectorStore(logger)

	// Initialize services
	memoryService := app.NewMemoryService(memoryRepo, embeddingProvider, vectorStore, logger)
	projectService := app.NewProjectService(projectRepo, logger)
	taskService := app.NewTaskService(memoryService, logger)

	// Create test project
	ctx := context.Background()
	project, err := projectService.InitializeProject(ctx, "/test/path", ports.InitializeProjectRequest{
		Name:        "Test Project",
		Description: "Test project for task handlers",
	})
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Initialize MCP server
	server := NewMemoryBankServer(memoryService, projectService, nil, taskService, logger)

	// Test task creation
	// Create a simple request that matches the expected format
	arguments := map[string]interface{}{
		"project_id":  string(project.ID),
		"title":       "Test Task",
		"description": "Test Description",
		"priority":    "high",
		"assignee":    "john.doe",
		"tags":        []interface{}{"test", "example"},
	}

	request := mcp.CallToolRequest{}
	request.Params.Arguments = arguments

	result, err := server.handleCreateTaskTool(ctx, request)
	if err != nil {
		t.Fatalf("Failed to create task via MCP: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result from task creation")
	}

	// Verify result format
	if len(result.Content) == 0 {
		t.Fatal("Expected content in result")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("Expected text content in result")
	}

	// Parse the JSON response
	var createdMemory domain.Memory
	err = json.Unmarshal([]byte(textContent.Text), &createdMemory)
	if err != nil {
		t.Fatalf("Failed to parse task creation response: %v", err)
	}

	// Verify task properties
	if createdMemory.Type != domain.MemoryTypeTask {
		t.Errorf("Expected memory type %s, got %s", domain.MemoryTypeTask, createdMemory.Type)
	}
	if createdMemory.Title != "Test Task" {
		t.Errorf("Expected title 'Test Task', got %s", createdMemory.Title)
	}
	if createdMemory.Content != "Test Description" {
		t.Errorf("Expected content 'Test Description', got %s", createdMemory.Content)
	}
	if createdMemory.ProjectID != project.ID {
		t.Errorf("Expected project ID %s, got %s", project.ID, createdMemory.ProjectID)
	}
}

func TestTaskHandlers_GetTask(t *testing.T) {
	// Setup test services (same as above)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	db, err := database.NewSQLiteDatabase(":memory:", logger)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	memoryRepo := database.NewSQLiteMemoryRepository(db, logger)
	projectRepo := database.NewSQLiteProjectRepository(db, logger)
	embeddingProvider := embedding.NewMockEmbeddingProvider(768, logger)
	vectorStore := vector.NewMockVectorStore(logger)

	memoryService := app.NewMemoryService(memoryRepo, embeddingProvider, vectorStore, logger)
	projectService := app.NewProjectService(projectRepo, logger)
	taskService := app.NewTaskService(memoryService, logger)

	ctx := context.Background()
	project, err := projectService.InitializeProject(ctx, "/test/path", ports.InitializeProjectRequest{
		Name: "Test Project",
	})
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	server := NewMemoryBankServer(memoryService, projectService, nil, taskService, logger)

	// Create a task first
	createRequest := mcp.CallToolRequest{}
	createRequest.Params.Arguments = map[string]interface{}{
		"project_id":  string(project.ID),
		"title":       "Test Task",
		"description": "Test Description",
	}

	createResult, err := server.handleCreateTaskTool(ctx, createRequest)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Parse the created task to get its ID
	textContent := createResult.Content[0].(mcp.TextContent)
	var createdMemory domain.Memory
	err = json.Unmarshal([]byte(textContent.Text), &createdMemory)
	if err != nil {
		t.Fatalf("Failed to parse created task: %v", err)
	}

	// Test getting the task
	getRequest := mcp.CallToolRequest{}
	getRequest.Params.Arguments = map[string]interface{}{
		"id": string(createdMemory.ID),
	}

	getResult, err := server.handleGetTaskTool(ctx, getRequest)
	if err != nil {
		t.Fatalf("Failed to get task via MCP: %v", err)
	}

	if getResult == nil {
		t.Fatal("Expected result from task retrieval")
	}

	// Verify result
	getTextContent := getResult.Content[0].(mcp.TextContent)
	var retrievedMemory domain.Memory
	err = json.Unmarshal([]byte(getTextContent.Text), &retrievedMemory)
	if err != nil {
		t.Fatalf("Failed to parse task retrieval response: %v", err)
	}

	if retrievedMemory.ID != createdMemory.ID {
		t.Errorf("Expected task ID %s, got %s", createdMemory.ID, retrievedMemory.ID)
	}
	if retrievedMemory.Title != "Test Task" {
		t.Errorf("Expected title 'Test Task', got %s", retrievedMemory.Title)
	}
}

func TestTaskHandlers_UpdateTask(t *testing.T) {
	// Setup test services
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	db, err := database.NewSQLiteDatabase(":memory:", logger)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	memoryRepo := database.NewSQLiteMemoryRepository(db, logger)
	projectRepo := database.NewSQLiteProjectRepository(db, logger)
	embeddingProvider := embedding.NewMockEmbeddingProvider(768, logger)
	vectorStore := vector.NewMockVectorStore(logger)

	memoryService := app.NewMemoryService(memoryRepo, embeddingProvider, vectorStore, logger)
	projectService := app.NewProjectService(projectRepo, logger)
	taskService := app.NewTaskService(memoryService, logger)

	ctx := context.Background()
	project, err := projectService.InitializeProject(ctx, "/test/path", ports.InitializeProjectRequest{
		Name: "Test Project",
	})
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	server := NewMemoryBankServer(memoryService, projectService, nil, taskService, logger)

	// Create a task first
	createRequest := mcp.CallToolRequest{}
	createRequest.Params.Arguments = map[string]interface{}{
		"project_id":  string(project.ID),
		"title":       "Original Task",
		"description": "Original Description",
	}

	createResult, err := server.handleCreateTaskTool(ctx, createRequest)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Parse the created task to get its ID
	textContent := createResult.Content[0].(mcp.TextContent)
	var createdMemory domain.Memory
	err = json.Unmarshal([]byte(textContent.Text), &createdMemory)
	if err != nil {
		t.Fatalf("Failed to parse created task: %v", err)
	}

	// Test updating the task
	updateRequest := mcp.CallToolRequest{}
	updateRequest.Params.Arguments = map[string]interface{}{
		"id":          string(createdMemory.ID),
		"title":       "Updated Task",
		"description": "Updated Description",
		"status":      "in_progress",
		"priority":    "urgent",
		"assignee":    "jane.doe",
	}

	updateResult, err := server.handleUpdateTaskTool(ctx, updateRequest)
	if err != nil {
		t.Fatalf("Failed to update task via MCP: %v", err)
	}

	if updateResult == nil {
		t.Fatal("Expected result from task update")
	}

	// Verify update result
	updateTextContent := updateResult.Content[0].(mcp.TextContent)

	// Try to parse as Task first (TaskService response), then fallback to Memory (MemoryService response)
	var updatedTask domain.Task
	err = json.Unmarshal([]byte(updateTextContent.Text), &updatedTask)
	if err == nil {
		// TaskService response - verify Task fields
		if updatedTask.Memory.Title != "Updated Task" {
			t.Errorf("Expected updated title 'Updated Task', got %s", updatedTask.Memory.Title)
		}
		if updatedTask.Memory.Content != "Updated Description" {
			t.Errorf("Expected updated content 'Updated Description', got %s", updatedTask.Memory.Content)
		}
		if updatedTask.Status != domain.TaskStatusInProgress {
			t.Errorf("Expected status 'in_progress', got %s", updatedTask.Status)
		}
		if updatedTask.Priority != domain.PriorityUrgent {
			t.Errorf("Expected priority 'urgent', got %s", updatedTask.Priority)
		}
		if updatedTask.Assignee != "jane.doe" {
			t.Errorf("Expected assignee 'jane.doe', got %s", updatedTask.Assignee)
		}
	} else {
		// Fallback to Memory response format
		var updatedMemory domain.Memory
		err = json.Unmarshal([]byte(updateTextContent.Text), &updatedMemory)
		if err != nil {
			t.Fatalf("Failed to parse task update response as Task or Memory: %v", err)
		}

		if updatedMemory.Title != "Updated Task" {
			t.Errorf("Expected updated title 'Updated Task', got %s", updatedMemory.Title)
		}
		if updatedMemory.Content != "Updated Description" {
			t.Errorf("Expected updated content 'Updated Description', got %s", updatedMemory.Content)
		}

		// Verify context contains status and other updates (fallback mode)
		expectedContextParts := []string{"Status: in_progress", "Priority: urgent", "Assignee: jane.doe"}
		for _, part := range expectedContextParts {
			if !contains(updatedMemory.Context, part) {
				t.Errorf("Expected context to contain '%s', got: %s", part, updatedMemory.Context)
			}
		}
	}
}

func TestTaskHandlers_ListTasks(t *testing.T) {
	// Setup test services
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	db, err := database.NewSQLiteDatabase(":memory:", logger)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	memoryRepo := database.NewSQLiteMemoryRepository(db, logger)
	projectRepo := database.NewSQLiteProjectRepository(db, logger)
	embeddingProvider := embedding.NewMockEmbeddingProvider(768, logger)
	vectorStore := vector.NewMockVectorStore(logger)

	memoryService := app.NewMemoryService(memoryRepo, embeddingProvider, vectorStore, logger)
	projectService := app.NewProjectService(projectRepo, logger)
	taskService := app.NewTaskService(memoryService, logger)

	ctx := context.Background()
	project, err := projectService.InitializeProject(ctx, "/test/path", ports.InitializeProjectRequest{
		Name: "Test Project",
	})
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	server := NewMemoryBankServer(memoryService, projectService, nil, taskService, logger)

	// Create multiple tasks
	taskTitles := []string{"Task 1", "Task 2", "Task 3"}
	for _, title := range taskTitles {
		createRequest := mcp.CallToolRequest{}
		createRequest.Params.Arguments = map[string]interface{}{
			"project_id":  string(project.ID),
			"title":       title,
			"description": "Test Description",
		}

		_, err := server.handleCreateTaskTool(ctx, createRequest)
		if err != nil {
			t.Fatalf("Failed to create task %s: %v", title, err)
		}
	}

	// Test listing tasks
	listRequest := mcp.CallToolRequest{}
	listRequest.Params.Arguments = map[string]interface{}{
		"project_id": string(project.ID),
	}

	listResult, err := server.handleListTasksTool(ctx, listRequest)
	if err != nil {
		t.Fatalf("Failed to list tasks via MCP: %v", err)
	}

	if listResult == nil {
		t.Fatal("Expected result from task listing")
	}

	// Verify list result
	listTextContent := listResult.Content[0].(mcp.TextContent)
	var tasks []domain.Memory
	err = json.Unmarshal([]byte(listTextContent.Text), &tasks)
	if err != nil {
		t.Fatalf("Failed to parse task list response: %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(tasks))
	}

	// Verify all tasks are of correct type
	for _, task := range tasks {
		if task.Type != domain.MemoryTypeTask {
			t.Errorf("Expected task type %s, got %s", domain.MemoryTypeTask, task.Type)
		}
		if task.ProjectID != project.ID {
			t.Errorf("Expected project ID %s, got %s", project.ID, task.ProjectID)
		}
	}
}

func TestTaskHandlers_DeleteTask(t *testing.T) {
	// Setup test services
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	db, err := database.NewSQLiteDatabase(":memory:", logger)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	memoryRepo := database.NewSQLiteMemoryRepository(db, logger)
	projectRepo := database.NewSQLiteProjectRepository(db, logger)
	embeddingProvider := embedding.NewMockEmbeddingProvider(768, logger)
	vectorStore := vector.NewMockVectorStore(logger)

	memoryService := app.NewMemoryService(memoryRepo, embeddingProvider, vectorStore, logger)
	projectService := app.NewProjectService(projectRepo, logger)
	taskService := app.NewTaskService(memoryService, logger)

	ctx := context.Background()
	project, err := projectService.InitializeProject(ctx, "/test/path", ports.InitializeProjectRequest{
		Name: "Test Project",
	})
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	server := NewMemoryBankServer(memoryService, projectService, nil, taskService, logger)

	// Create a task first
	createRequest := mcp.CallToolRequest{}
	createRequest.Params.Arguments = map[string]interface{}{
		"project_id":  string(project.ID),
		"title":       "Task to Delete",
		"description": "This task will be deleted",
	}

	createResult, err := server.handleCreateTaskTool(ctx, createRequest)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Parse the created task to get its ID
	textContent := createResult.Content[0].(mcp.TextContent)
	var createdMemory domain.Memory
	err = json.Unmarshal([]byte(textContent.Text), &createdMemory)
	if err != nil {
		t.Fatalf("Failed to parse created task: %v", err)
	}

	// Test deleting the task
	deleteRequest := mcp.CallToolRequest{}
	deleteRequest.Params.Arguments = map[string]interface{}{
		"id": string(createdMemory.ID),
	}

	deleteResult, err := server.handleDeleteTaskTool(ctx, deleteRequest)
	if err != nil {
		t.Fatalf("Failed to delete task via MCP: %v", err)
	}

	if deleteResult == nil {
		t.Fatal("Expected result from task deletion")
	}

	// Verify deletion result
	deleteTextContent := deleteResult.Content[0].(mcp.TextContent)
	if deleteTextContent.Text != "Task deleted successfully" {
		t.Errorf("Expected success message, got: %s", deleteTextContent.Text)
	}

	// Verify task is actually deleted by trying to get it
	getRequest := mcp.CallToolRequest{}
	getRequest.Params.Arguments = map[string]interface{}{
		"id": string(createdMemory.ID),
	}

	getResult, err := server.handleGetTaskTool(ctx, getRequest)
	if err != nil {
		t.Fatalf("Error when trying to get deleted task: %v", err)
	}

	// Should get an error message about task not found
	getTextContent := getResult.Content[0].(mcp.TextContent)
	if !contains(getTextContent.Text, "Failed to get task") {
		t.Errorf("Expected error message about task not found, got: %s", getTextContent.Text)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
