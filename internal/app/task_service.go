package app

import (
	"context"
	"fmt"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/sirupsen/logrus"
)

// taskService implements the TaskService interface
type taskService struct {
	memoryService ports.MemoryService
	logger        *logrus.Logger
}

// NewTaskService creates a new task service
func NewTaskService(memoryService ports.MemoryService, logger *logrus.Logger) ports.TaskService {
	return &taskService{
		memoryService: memoryService,
		logger:        logger,
	}
}

// CreateTask creates a new task
func (s *taskService) CreateTask(ctx context.Context, req ports.CreateTaskRequest) (*domain.Task, error) {
	s.logger.WithFields(logrus.Fields{
		"project_id": req.ProjectID,
		"title":      req.Title,
		"priority":   req.Priority,
	}).Info("Creating new task")

	// Create the task domain object
	var task *domain.Task
	if req.DueDate != nil || req.Assignee != "" || req.EstimatedHours != nil {
		task = domain.NewTaskWithDetails(req.ProjectID, req.Title, req.Description, req.Priority, req.DueDate, req.Assignee, req.EstimatedHours)
	} else {
		task = domain.NewTask(req.ProjectID, req.Title, req.Description, req.Priority)
	}

	// Set parent task if specified
	if req.ParentTask != nil {
		task.SetParentTask(*req.ParentTask)
	}

	// Add dependencies
	for _, depID := range req.Dependencies {
		task.AddDependency(depID)
	}

	// Add tags
	for _, tag := range req.Tags {
		task.AddTag(tag)
	}

	// Store as memory
	memory, err := s.memoryService.CreateMemory(ctx, ports.CreateMemoryRequest{
		ProjectID: req.ProjectID,
		Type:      domain.MemoryTypeTask,
		Title:     req.Title,
		Content:   req.Description,
		Context:   fmt.Sprintf("Task Status: %s, Priority: %s", task.Status, task.Priority),
		Tags:      req.Tags,
	})
	if err != nil {
		s.logger.WithError(err).Error("Failed to create task memory")
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Update task with the created memory
	task.Memory = memory

	s.logger.WithField("task_id", task.ID).Info("Task created successfully")
	return task, nil
}

// GetTask retrieves a task by ID
func (s *taskService) GetTask(ctx context.Context, taskID domain.MemoryID) (*domain.Task, error) {
	memory, err := s.memoryService.GetMemory(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if memory.Type != domain.MemoryTypeTask {
		return nil, fmt.Errorf("memory %s is not a task", taskID)
	}

	return s.memoryToTask(memory), nil
}

// UpdateTask updates an existing task
func (s *taskService) UpdateTask(ctx context.Context, req ports.UpdateTaskRequest) (*domain.Task, error) {
	s.logger.WithField("task_id", req.TaskID).Info("Updating task")

	// Get existing task
	task, err := s.GetTask(ctx, req.TaskID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	updated := false

	// Update title and description
	if req.Title != nil {
		task.Title = *req.Title
		updated = true
	}
	if req.Description != nil {
		task.Content = *req.Description
		updated = true
	}

	// Update task-specific fields
	if req.Status != nil {
		task.UpdateStatus(*req.Status)
		updated = true
	}
	if req.Priority != nil {
		task.UpdatePriority(*req.Priority)
		updated = true
	}
	if req.DueDate != nil {
		task.SetDueDate(*req.DueDate)
		updated = true
	}
	if req.ClearDueDate {
		task.ClearDueDate()
		updated = true
	}
	if req.Assignee != nil {
		task.AssignTo(*req.Assignee)
		updated = true
	}
	if req.EstimatedHours != nil {
		task.UpdateEstimate(*req.EstimatedHours)
		updated = true
	}
	if req.ActualHours != nil {
		task.LogActualHours(*req.ActualHours)
		updated = true
	}
	if req.Tags != nil {
		task.Tags = req.Tags
		updated = true
	}

	// Update the memory if any changes were made
	if updated {
		// Update the context with current task state
		task.Context = fmt.Sprintf("Task Status: %s, Priority: %s", task.Status, task.Priority)
		
		err := s.memoryService.UpdateMemory(ctx, task.Memory)
		if err != nil {
			s.logger.WithError(err).Error("Failed to update task memory")
			return nil, fmt.Errorf("failed to update task: %w", err)
		}
	}
	s.logger.WithField("task_id", req.TaskID).Info("Task updated successfully")
	return task, nil
}

// DeleteTask deletes a task
func (s *taskService) DeleteTask(ctx context.Context, taskID domain.MemoryID) error {
	s.logger.WithField("task_id", taskID).Info("Deleting task")

	err := s.memoryService.DeleteMemory(ctx, taskID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to delete task")
		return fmt.Errorf("failed to delete task: %w", err)
	}

	s.logger.WithField("task_id", taskID).Info("Task deleted successfully")
	return nil
}

// UpdateTaskStatus updates the status of a task
func (s *taskService) UpdateTaskStatus(ctx context.Context, taskID domain.MemoryID, status domain.TaskStatus) error {
	return s.updateTaskField(ctx, taskID, "status", status)
}

// ListTasks lists tasks with filters
func (s *taskService) ListTasks(ctx context.Context, filters ports.TaskFilters) ([]*domain.Task, error) {
	// Convert task filters to memory filters
	taskType := domain.MemoryTypeTask
	memoryFilters := ports.ListMemoriesRequest{
		ProjectID: filters.ProjectID,
		Type:      &taskType,
		Tags:      filters.Tags,
		Limit:     filters.Limit,
	}

	memories, err := s.memoryService.ListMemories(ctx, memoryFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	tasks := make([]*domain.Task, 0, len(memories))
	for _, memory := range memories {
		task := s.memoryToTask(memory)
		
		// Apply task-specific filters
		if s.matchesTaskFilters(task, filters) {
			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}

// GetTasksByProject gets all tasks for a project
func (s *taskService) GetTasksByProject(ctx context.Context, projectID domain.ProjectID) ([]*domain.Task, error) {
	return s.ListTasks(ctx, ports.TaskFilters{ProjectID: &projectID})
}

// GetTasksByAssignee gets tasks assigned to a specific assignee
func (s *taskService) GetTasksByAssignee(ctx context.Context, assignee string) ([]*domain.Task, error) {
	return s.ListTasks(ctx, ports.TaskFilters{Assignee: &assignee})
}

// GetTasksByStatus gets tasks with a specific status
func (s *taskService) GetTasksByStatus(ctx context.Context, status domain.TaskStatus) ([]*domain.Task, error) {
	return s.ListTasks(ctx, ports.TaskFilters{Status: &status})
}

// GetTasksByPriority gets tasks with a specific priority
func (s *taskService) GetTasksByPriority(ctx context.Context, priority domain.Priority) ([]*domain.Task, error) {
	return s.ListTasks(ctx, ports.TaskFilters{Priority: &priority})
}

// GetOverdueTasks gets all overdue tasks
func (s *taskService) GetOverdueTasks(ctx context.Context) ([]*domain.Task, error) {
	isOverdue := true
	return s.ListTasks(ctx, ports.TaskFilters{IsOverdue: &isOverdue})
}

// Task dependency management methods
func (s *taskService) AddTaskDependency(ctx context.Context, taskID, dependencyID domain.MemoryID) error {
	task, err := s.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	task.AddDependency(dependencyID)
	
	_, err = s.UpdateTask(ctx, ports.UpdateTaskRequest{
		TaskID: taskID,
	})
	return err
}

func (s *taskService) RemoveTaskDependency(ctx context.Context, taskID, dependencyID domain.MemoryID) error {
	task, err := s.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	task.RemoveDependency(dependencyID)
	
	_, err = s.UpdateTask(ctx, ports.UpdateTaskRequest{
		TaskID: taskID,
	})
	return err
}

func (s *taskService) GetTaskDependencies(ctx context.Context, taskID domain.MemoryID) ([]*domain.Task, error) {
	task, err := s.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	dependencies := make([]*domain.Task, 0, len(task.Dependencies))
	for _, depID := range task.Dependencies {
		dep, err := s.GetTask(ctx, depID)
		if err != nil {
			s.logger.WithError(err).WithField("dependency_id", depID).Warn("Failed to get task dependency")
			continue
		}
		dependencies = append(dependencies, dep)
	}

	return dependencies, nil
}

func (s *taskService) GetTaskDependents(ctx context.Context, taskID domain.MemoryID) ([]*domain.Task, error) {
	// This would require a more complex query to find all tasks that depend on this task
	// For now, we'll implement a simple approach by listing all tasks and checking dependencies
	allTasks, err := s.ListTasks(ctx, ports.TaskFilters{})
	if err != nil {
		return nil, err
	}

	dependents := make([]*domain.Task, 0)
	for _, task := range allTasks {
		for _, depID := range task.Dependencies {
			if depID == taskID {
				dependents = append(dependents, task)
				break
			}
		}
	}

	return dependents, nil
}

// Task hierarchy management
func (s *taskService) AddSubtask(ctx context.Context, parentID, subtaskID domain.MemoryID) error {
	// Add subtask to parent
	parent, err := s.GetTask(ctx, parentID)
	if err != nil {
		return err
	}
	parent.AddSubtask(subtaskID)

	// Set parent on subtask
	subtask, err := s.GetTask(ctx, subtaskID)
	if err != nil {
		return err
	}
	subtask.SetParentTask(parentID)

	// Update both tasks
	_, err = s.UpdateTask(ctx, ports.UpdateTaskRequest{TaskID: parentID})
	if err != nil {
		return err
	}

	_, err = s.UpdateTask(ctx, ports.UpdateTaskRequest{TaskID: subtaskID})
	return err
}

func (s *taskService) RemoveSubtask(ctx context.Context, parentID, subtaskID domain.MemoryID) error {
	// Remove subtask from parent
	parent, err := s.GetTask(ctx, parentID)
	if err != nil {
		return err
	}
	parent.RemoveSubtask(subtaskID)

	// Clear parent on subtask
	subtask, err := s.GetTask(ctx, subtaskID)
	if err != nil {
		return err
	}
	subtask.ClearParentTask()

	// Update both tasks
	_, err = s.UpdateTask(ctx, ports.UpdateTaskRequest{TaskID: parentID})
	if err != nil {
		return err
	}

	_, err = s.UpdateTask(ctx, ports.UpdateTaskRequest{TaskID: subtaskID})
	return err
}

func (s *taskService) GetSubtasks(ctx context.Context, parentID domain.MemoryID) ([]*domain.Task, error) {
	parent, err := s.GetTask(ctx, parentID)
	if err != nil {
		return nil, err
	}

	subtasks := make([]*domain.Task, 0, len(parent.Subtasks))
	for _, subtaskID := range parent.Subtasks {
		subtask, err := s.GetTask(ctx, subtaskID)
		if err != nil {
			s.logger.WithError(err).WithField("subtask_id", subtaskID).Warn("Failed to get subtask")
			continue
		}
		subtasks = append(subtasks, subtask)
	}

	return subtasks, nil
}

func (s *taskService) GetParentTask(ctx context.Context, taskID domain.MemoryID) (*domain.Task, error) {
	task, err := s.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if task.ParentTask == nil {
		return nil, nil
	}

	return s.GetTask(ctx, *task.ParentTask)
}

// Task analytics
func (s *taskService) GetTaskStatistics(ctx context.Context, projectID domain.ProjectID) (*ports.TaskStatistics, error) {
	tasks, err := s.GetTasksByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	stats := &ports.TaskStatistics{
		TotalTasks:      len(tasks),
		TasksByPriority: make(map[domain.Priority]int),
		TasksByAssignee: make(map[string]int),
	}

	totalHours := 0
	estimatedTasksCount := 0

	for _, task := range tasks {
		// Count by status
		switch task.Status {
		case domain.TaskStatusDone:
			stats.CompletedTasks++
		case domain.TaskStatusInProgress:
			stats.InProgressTasks++
		case domain.TaskStatusTodo:
			stats.TodoTasks++
		case domain.TaskStatusBlocked:
			stats.BlockedTasks++
		}

		// Count overdue tasks
		if task.IsOverdue() {
			stats.OverdueTasks++
		}

		// Count by priority
		stats.TasksByPriority[task.Priority]++

		// Count by assignee
		if task.Assignee != "" {
			stats.TasksByAssignee[task.Assignee]++
		}

		// Calculate hours
		if task.EstimatedHours != nil {
			totalHours += *task.EstimatedHours
			estimatedTasksCount++
		}
	}

	stats.TotalHours = totalHours
	if estimatedTasksCount > 0 {
		stats.AverageHours = float64(totalHours) / float64(estimatedTasksCount)
	}

	if stats.TotalTasks > 0 {
		stats.CompletionRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}

	return stats, nil
}

func (s *taskService) GetTaskEfficiencyReport(ctx context.Context, projectID domain.ProjectID) (*ports.EfficiencyReport, error) {
	tasks, err := s.GetTasksByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	report := &ports.EfficiencyReport{}
	efficiencySum := 0.0
	assigneeEfficiency := make(map[string][]float64)

	for _, task := range tasks {
		if task.EstimatedHours != nil && task.ActualHours != nil && *task.EstimatedHours > 0 {
			report.TotalTasksWithEstimates++
			
			efficiency := float64(*task.EstimatedHours) / float64(*task.ActualHours)
			efficiencySum += efficiency

			if task.Assignee != "" {
				assigneeEfficiency[task.Assignee] = append(assigneeEfficiency[task.Assignee], efficiency)
			}

			if efficiency > 1.0 {
				report.TasksUnderTime++
			} else if efficiency < 1.0 {
				report.TasksOverTime++
			} else {
				report.TasksOnTime++
			}
		}
	}

	if report.TotalTasksWithEstimates > 0 {
		report.AverageEfficiencyRatio = efficiencySum / float64(report.TotalTasksWithEstimates)
	}

	// Find most and least efficient assignees
	bestEfficiency := 0.0
	worstEfficiency := 999.0

	for assignee, efficiencies := range assigneeEfficiency {
		if len(efficiencies) == 0 {
			continue
		}

		sum := 0.0
		for _, eff := range efficiencies {
			sum += eff
		}
		avgEfficiency := sum / float64(len(efficiencies))

		if avgEfficiency > bestEfficiency {
			bestEfficiency = avgEfficiency
			report.MostEfficientAssignee = assignee
		}
		if avgEfficiency < worstEfficiency {
			worstEfficiency = avgEfficiency
			report.LeastEfficientAssignee = assignee
		}
	}

	return report, nil
}

// Helper methods

func (s *taskService) memoryToTask(memory *domain.Memory) *domain.Task {
	task := &domain.Task{
		Memory:       memory,
		Status:       domain.TaskStatusTodo, // Default
		Priority:     domain.PriorityMedium, // Default
		Dependencies: make([]domain.MemoryID, 0),
		Subtasks:     make([]domain.MemoryID, 0),
	}

	// For now, we'll use simple defaults
	// In a full implementation, we'd need to store task metadata 
	// in a separate table or extend the Memory model
	// We can parse some info from the context field if needed
	
	return task
}

func (s *taskService) matchesTaskFilters(task *domain.Task, filters ports.TaskFilters) bool {
	if filters.Status != nil && task.Status != *filters.Status {
		return false
	}
	if filters.Priority != nil && task.Priority != *filters.Priority {
		return false
	}
	if filters.Assignee != nil && task.Assignee != *filters.Assignee {
		return false
	}
	if filters.DueBefore != nil && (task.DueDate == nil || task.DueDate.After(*filters.DueBefore)) {
		return false
	}
	if filters.DueAfter != nil && (task.DueDate == nil || task.DueDate.Before(*filters.DueAfter)) {
		return false
	}
	if filters.HasDueDate != nil && (*filters.HasDueDate != (task.DueDate != nil)) {
		return false
	}
	if filters.IsOverdue != nil && (*filters.IsOverdue != task.IsOverdue()) {
		return false
	}
	if filters.ParentTask != nil && (task.ParentTask == nil || *task.ParentTask != *filters.ParentTask) {
		return false
	}
	if filters.HasSubtasks != nil && (*filters.HasSubtasks != task.HasSubtasks()) {
		return false
	}

	return true
}

func (s *taskService) updateTaskField(ctx context.Context, taskID domain.MemoryID, field string, value interface{}) error {
	task, err := s.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	// Update the context with the field information
	task.Context = fmt.Sprintf("%s: %v", field, value)
	
	// Update the memory
	err = s.memoryService.UpdateMemory(ctx, task.Memory)
	return err
}