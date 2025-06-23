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
)

// Full-featured task handlers using TaskService for complete functionality

func (s *MemoryBankServer) handleCreateTaskTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Handling task_create tool request")
	
	argsMap, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Invalid arguments"}},
		}, nil
	}

	// Extract required fields
	projectID, _ := argsMap["project_id"].(string)
	title, _ := argsMap["title"].(string)
	description, _ := argsMap["description"].(string)
	
	if projectID == "" || title == "" || description == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "project_id, title, and description are required"}},
		}, nil
	}

	// Build create request with all optional fields
	req := ports.CreateTaskRequest{
		ProjectID:   domain.ProjectID(projectID),
		Title:       title,
		Description: description,
	}

	// Optional priority
	if priority, ok := argsMap["priority"].(string); ok && priority != "" {
		req.Priority = domain.Priority(priority)
	} else {
		req.Priority = domain.PriorityMedium // default
	}

	// Optional assignee
	if assignee, ok := argsMap["assignee"].(string); ok && assignee != "" {
		req.Assignee = assignee
	}

	// Optional due date (expecting YYYY-MM-DD format)
	if dueDateStr, ok := argsMap["due_date"].(string); ok && dueDateStr != "" {
		if dueDate, err := time.Parse("2006-01-02", dueDateStr); err == nil {
			req.DueDate = &dueDate
		}
	}

	// Optional estimated hours
	if estimatedHours, ok := argsMap["estimated_hours"].(float64); ok {
		hours := int(estimatedHours)
		req.EstimatedHours = &hours
	}

	// Optional tags
	if tagsInterface, ok := argsMap["tags"]; ok {
		if tagsList, ok := tagsInterface.([]interface{}); ok {
			var tags domain.Tags
			for _, tag := range tagsList {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				}
			}
			req.Tags = tags
		}
	}

	// Use TaskService for full functionality
	var task *domain.Task
	var err error
	
	if s.taskService != nil {
		task, err = s.taskService.CreateTask(ctx, req)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Failed to create task: " + err.Error()}},
			}, nil
		}
	} else {
		// Graceful fallback to MemoryService if TaskService not available
		memory, fallbackErr := s.memoryService.CreateMemory(ctx, ports.CreateMemoryRequest{
			ProjectID: req.ProjectID,
			Type:      domain.MemoryTypeTask,
			Title:     req.Title,
			Content:   req.Description,
			Context:   fmt.Sprintf("Priority: %s, Assignee: %s", req.Priority, req.Assignee),
			Tags:      req.Tags,
		})
		if fallbackErr != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Failed to create task: " + fallbackErr.Error()}},
			}, nil
		}
		
		response, _ := json.MarshalIndent(memory, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(response)}},
		}, nil
	}

	response, _ := json.MarshalIndent(task, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(response)}},
	}, nil
}

func (s *MemoryBankServer) handleGetTaskTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Handling task_get tool request")
	
	argsMap, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Invalid arguments"}},
		}, nil
	}

	id, ok := argsMap["id"].(string)
	if !ok || id == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Task ID is required"}},
		}, nil
	}

	// Use TaskService for full functionality
	if s.taskService != nil {
		task, err := s.taskService.GetTask(ctx, domain.MemoryID(id))
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Failed to get task: " + err.Error()}},
			}, nil
		}

		response, _ := json.MarshalIndent(task, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(response)}},
		}, nil
	}

	// Graceful fallback to MemoryService
	memory, err := s.memoryService.GetMemory(ctx, domain.MemoryID(id))
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Failed to get task: " + err.Error()}},
		}, nil
	}

	response, _ := json.MarshalIndent(memory, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(response)}},
	}, nil
}

func (s *MemoryBankServer) handleUpdateTaskTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Handling task_update tool request")

	argsMap, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Invalid arguments"}},
		}, nil
	}

	// Get task ID
	id, ok := argsMap["id"].(string)
	if !ok || id == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Task ID is required"}},
		}, nil
	}

	// Use TaskService for full functionality
	if s.taskService != nil {
		// Build update request
		req := ports.UpdateTaskRequest{
			TaskID: domain.MemoryID(id),
		}

		// Update fields if provided
		if title, ok := argsMap["title"].(string); ok && title != "" {
			req.Title = &title
		}
		
		if description, ok := argsMap["description"].(string); ok && description != "" {
			req.Description = &description
		}
		
		if status, ok := argsMap["status"].(string); ok && status != "" {
			taskStatus := domain.TaskStatus(status)
			req.Status = &taskStatus
		}
		
		if priority, ok := argsMap["priority"].(string); ok && priority != "" {
			taskPriority := domain.Priority(priority)
			req.Priority = &taskPriority
		}
		
		if assignee, ok := argsMap["assignee"].(string); ok && assignee != "" {
			req.Assignee = &assignee
		}
		
		if dueDateStr, ok := argsMap["due_date"].(string); ok && dueDateStr != "" {
			if dueDate, err := time.Parse("2006-01-02", dueDateStr); err == nil {
				req.DueDate = &dueDate
			}
		}
		
		if clearDueDate, ok := argsMap["clear_due_date"].(bool); ok && clearDueDate {
			req.ClearDueDate = true
		}
		
		if estimatedHours, ok := argsMap["estimated_hours"].(float64); ok {
			hours := int(estimatedHours)
			req.EstimatedHours = &hours
		}
		
		if actualHours, ok := argsMap["actual_hours"].(float64); ok {
			hours := int(actualHours)
			req.ActualHours = &hours
		}
		
		if tagsInterface, ok := argsMap["tags"]; ok {
			if tagsList, ok := tagsInterface.([]interface{}); ok {
				var tags domain.Tags
				for _, tag := range tagsList {
					if tagStr, ok := tag.(string); ok {
						tags = append(tags, tagStr)
					}
				}
				req.Tags = tags
			}
		}

		task, err := s.taskService.UpdateTask(ctx, req)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Failed to update task: " + err.Error()}},
			}, nil
		}

		response, _ := json.MarshalIndent(task, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(response)}},
		}, nil
	}

	// Graceful fallback to MemoryService (simplified functionality)
	memory, err := s.memoryService.GetMemory(ctx, domain.MemoryID(id))
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Failed to get task: " + err.Error()}},
		}, nil
	}

	// Update basic fields
	updated := false
	
	if title, ok := argsMap["title"].(string); ok && title != "" {
		memory.Title = title
		updated = true
	}
	
	if description, ok := argsMap["description"].(string); ok && description != "" {
		memory.Content = description
		updated = true
	}
	
	// Build context with task information
	contextParts := []string{}
	
	if status, ok := argsMap["status"].(string); ok && status != "" {
		contextParts = append(contextParts, "Status: "+status)
		updated = true
	}
	
	if priority, ok := argsMap["priority"].(string); ok && priority != "" {
		contextParts = append(contextParts, "Priority: "+priority)
		updated = true
	}
	
	if assignee, ok := argsMap["assignee"].(string); ok && assignee != "" {
		contextParts = append(contextParts, "Assignee: "+assignee)
		updated = true
	}
	
	if estimatedHours, ok := argsMap["estimated_hours"].(float64); ok && estimatedHours > 0 {
		contextParts = append(contextParts, fmt.Sprintf("Estimated: %.0fh", estimatedHours))
		updated = true
	}
	
	if actualHours, ok := argsMap["actual_hours"].(float64); ok && actualHours > 0 {
		contextParts = append(contextParts, fmt.Sprintf("Actual: %.0fh", actualHours))
		updated = true
	}
	
	if dueDate, ok := argsMap["due_date"].(string); ok && dueDate != "" {
		contextParts = append(contextParts, "Due: "+dueDate)
		updated = true
	}
	
	if clearDueDate, ok := argsMap["clear_due_date"].(bool); ok && clearDueDate {
		contextParts = append(contextParts, "Due date cleared")
		updated = true
	}

	// Update context if any changes were made
	if updated && len(contextParts) > 0 {
		if memory.Context != "" {
			memory.Context = memory.Context + " | " + strings.Join(contextParts, ", ")
		} else {
			memory.Context = strings.Join(contextParts, ", ")
		}
	}
	
	// Update tags if provided
	if tags, ok := argsMap["tags"].([]interface{}); ok {
		memory.Tags = make(domain.Tags, 0, len(tags))
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				memory.Tags = append(memory.Tags, tagStr)
			}
		}
		updated = true
	}

	// Save changes if any were made
	if updated {
		err = s.memoryService.UpdateMemory(ctx, memory)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Failed to update task: " + err.Error()}},
			}, nil
		}
	}

	response, _ := json.MarshalIndent(memory, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(response)}},
	}, nil
}

func (s *MemoryBankServer) handleDeleteTaskTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Handling task_delete tool request")
	
	argsMap, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Invalid arguments"}},
		}, nil
	}

	id, ok := argsMap["id"].(string)
	if !ok || id == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Task ID is required"}},
		}, nil
	}

	// Use TaskService for full functionality
	if s.taskService != nil {
		err := s.taskService.DeleteTask(ctx, domain.MemoryID(id))
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Failed to delete task: " + err.Error()}},
			}, nil
		}
	} else {
		// Graceful fallback to MemoryService
		err := s.memoryService.DeleteMemory(ctx, domain.MemoryID(id))
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Failed to delete task: " + err.Error()}},
			}, nil
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Task deleted successfully"}},
	}, nil
}

func (s *MemoryBankServer) handleListTasksTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Handling task_list tool request")
	
	argsMap, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		argsMap = make(map[string]interface{})
	}

	// Use TaskService for full functionality
	if s.taskService != nil {
		// Build filters
		filters := ports.TaskFilters{
			Limit: 20, // default limit
		}

		if projectID, ok := argsMap["project_id"].(string); ok && projectID != "" {
			pid := domain.ProjectID(projectID)
			filters.ProjectID = &pid
		}

		if status, ok := argsMap["status"].(string); ok && status != "" {
			taskStatus := domain.TaskStatus(status)
			filters.Status = &taskStatus
		}

		if priority, ok := argsMap["priority"].(string); ok && priority != "" {
			taskPriority := domain.Priority(priority)
			filters.Priority = &taskPriority
		}

		if assignee, ok := argsMap["assignee"].(string); ok && assignee != "" {
			filters.Assignee = &assignee
		}

		if limit, ok := argsMap["limit"].(float64); ok && limit > 0 {
			filters.Limit = int(limit)
		}

		if overdue, ok := argsMap["overdue"].(bool); ok {
			filters.IsOverdue = &overdue
		}

		if tagsInterface, ok := argsMap["tags"]; ok {
			if tagsList, ok := tagsInterface.([]interface{}); ok {
				var tags domain.Tags
				for _, tag := range tagsList {
					if tagStr, ok := tag.(string); ok {
						tags = append(tags, tagStr)
					}
				}
				filters.Tags = tags
			}
		}

		tasks, err := s.taskService.ListTasks(ctx, filters)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Failed to list tasks: " + err.Error()}},
			}, nil
		}

		response, _ := json.MarshalIndent(tasks, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(response)}},
		}, nil
	}

	// Graceful fallback to MemoryService
	var projectID *domain.ProjectID
	if pid, ok := argsMap["project_id"].(string); ok && pid != "" {
		p := domain.ProjectID(pid)
		projectID = &p
	}

	limit := 20
	if l, ok := argsMap["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	taskType := domain.MemoryTypeTask
	memories, err := s.memoryService.ListMemories(ctx, ports.ListMemoriesRequest{
		ProjectID: projectID,
		Type:      &taskType,
		Limit:     limit,
	})
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Failed to list tasks: " + err.Error()}},
		}, nil
	}

	response, _ := json.MarshalIndent(memories, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(response)}},
	}, nil
}

// Simplified stub handlers for other task operations
func (s *MemoryBankServer) handleAddTaskDependencyTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Task dependencies not yet implemented"}},
	}, nil
}

func (s *MemoryBankServer) handleRemoveTaskDependencyTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Task dependencies not yet implemented"}},
	}, nil
}

func (s *MemoryBankServer) handleAddSubtaskTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Subtasks not yet implemented"}},
	}, nil
}

func (s *MemoryBankServer) handleRemoveSubtaskTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Subtasks not yet implemented"}},
	}, nil
}

func (s *MemoryBankServer) handleTaskStatisticsTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Debug("Handling task_statistics tool request")
	
	argsMap, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		argsMap = make(map[string]interface{})
	}

	// Use TaskService for full functionality
	if s.taskService != nil {
		var projectID domain.ProjectID
		if pid, ok := argsMap["project_id"].(string); ok && pid != "" {
			projectID = domain.ProjectID(pid)
		} else {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "project_id is required for task statistics"}},
			}, nil
		}

		stats, err := s.taskService.GetTaskStatistics(ctx, projectID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Failed to get task statistics: " + err.Error()}},
			}, nil
		}

		response, _ := json.MarshalIndent(stats, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(response)}},
		}, nil
	}

	// Fallback: basic statistics via MemoryService
	var projectID *domain.ProjectID
	if pid, ok := argsMap["project_id"].(string); ok && pid != "" {
		p := domain.ProjectID(pid)
		projectID = &p
	} else {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "project_id is required for task statistics"}},
		}, nil
	}

	taskType := domain.MemoryTypeTask
	memories, err := s.memoryService.ListMemories(ctx, ports.ListMemoriesRequest{
		ProjectID: projectID,
		Type:      &taskType,
		Limit:     1000, // Get all tasks for stats
	})
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Failed to get tasks for statistics: " + err.Error()}},
		}, nil
	}

	// Basic statistics
	basicStats := map[string]interface{}{
		"total_tasks": len(memories),
		"project_id":  string(*projectID),
		"note":        "Limited statistics available without TaskService",
	}

	response, _ := json.MarshalIndent(basicStats, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(response)}},
	}, nil
}

func (s *MemoryBankServer) handleTaskEfficiencyReportTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Efficiency reports not yet implemented"}},
	}, nil
}