package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/spf13/cobra"
)

// NewTaskCmd creates the task command for external use with dependency injection
// This is kept for backward compatibility and testing
func NewTaskCmd(taskService ports.TaskService, projectService ports.ProjectService) *cobra.Command {
	// For now, return the service-lookup version
	// In the future, this could be enhanced to use the injected services directly
	return createTaskCommand()
}

func init() {
	// Add task command to root using service lookup pattern
	rootCmd.AddCommand(createTaskCommand())
}

func createTaskCommand() *cobra.Command {
	taskCmd := &cobra.Command{
		Use:   "task",
		Short: "Manage tasks",
		Long:  "Create, list, update, and manage tasks in Memory Bank",
	}

	// Add subcommands using the improved service-lookup pattern
	taskCmd.AddCommand(
		createTaskCreateCmd(),
		createTaskListCmd(),
		createTaskGetCmd(),
		createTaskUpdateCmd(),
		createTaskDeleteCmd(),
		createTaskStatsCmd(),
	)

	return taskCmd
}

func createTaskCreateCmd() *cobra.Command {
	var (
		projectID      string
		title          string
		description    string
		priority       string
		dueDate        string
		assignee       string
		estimatedHours int
		tags           []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new task",
		Long:  "Create a new task with title, description, and optional metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Initialize services
			services, err := GetServicesForCLI(cmd)
			if err != nil {
				return err
			}

			// Get project ID if not provided
			if projectID == "" {
				wd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get working directory: %w", err)
				}

				project, err := services.ProjectService.GetProjectByPath(ctx, wd)
				if err != nil {
					return fmt.Errorf("no project found in current directory. Use --project-id or run from a project directory")
				}
				projectID = string(project.ID)
			}

			// Use TaskService if available, otherwise fall back to MemoryService
			if services.TaskService != nil {
				// Create request for TaskService
				req := ports.CreateTaskRequest{
					ProjectID:   domain.ProjectID(projectID),
					Title:       title,
					Description: description,
					Priority:    domain.Priority(priority),
					Tags:        domain.Tags(tags),
				}

				// Parse due date if provided
				if dueDate != "" {
					parsedDate, err := time.Parse("2006-01-02", dueDate)
					if err != nil {
						return fmt.Errorf("invalid due date format. Use YYYY-MM-DD: %w", err)
					}
					req.DueDate = &parsedDate
				}

				// Set assignee if provided
				if assignee != "" {
					req.Assignee = assignee
				}

				// Set estimated hours if provided
				if estimatedHours > 0 {
					req.EstimatedHours = &estimatedHours
				}

				task, err := services.TaskService.CreateTask(ctx, req)
				if err != nil {
					return fmt.Errorf("failed to create task: %w", err)
				}

				output, err := json.MarshalIndent(task, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to format output: %w", err)
				}

				fmt.Println(string(output))
				return nil
			}

			// Fallback to MemoryService for basic task creation
			contextStr := fmt.Sprintf("Priority: %s", priority)
			if assignee != "" {
				contextStr += fmt.Sprintf(", Assignee: %s", assignee)
			}
			if dueDate != "" {
				contextStr += fmt.Sprintf(", Due Date: %s", dueDate)
			}

			memory, err := services.MemoryService.CreateMemory(ctx, ports.CreateMemoryRequest{
				ProjectID: domain.ProjectID(projectID),
				Type:      domain.MemoryTypeTask,
				Title:     title,
				Content:   description,
				Context:   contextStr,
				Tags:      domain.Tags(tags),
			})
			if err != nil {
				return fmt.Errorf("failed to create task: %w", err)
			}

			fmt.Println("Task created successfully:")
			output, err := json.MarshalIndent(memory, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format output: %w", err)
			}

			fmt.Println(string(output))
			return nil
		},
	}

	cmd.Flags().StringVar(&projectID, "project-id", "", "Project ID (auto-detected if not provided)")
	cmd.Flags().StringVar(&title, "title", "", "Task title")
	cmd.Flags().StringVar(&description, "description", "", "Task description")
	cmd.Flags().StringVar(&priority, "priority", "medium", "Task priority (low, medium, high, urgent)")
	cmd.Flags().StringVar(&dueDate, "due-date", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Assignee username")
	cmd.Flags().IntVar(&estimatedHours, "estimated-hours", 0, "Estimated hours to complete")
	cmd.Flags().StringSliceVar(&tags, "tags", []string{}, "Task tags")

	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("description")

	return cmd
}

func createTaskListCmd() *cobra.Command {
	var (
		projectID string
		status    string
		priority  string
		assignee  string
		tags      []string
		limit     int
		overdue   bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		Long:  "List tasks with optional filters",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Initialize services
			services, err := GetServicesForCLI(cmd)
			if err != nil {
				return err
			}

			// Get project ID if not provided
			if projectID == "" {
				wd, err := os.Getwd()
				if err == nil {
					if project, err := services.ProjectService.GetProjectByPath(ctx, wd); err == nil {
						projectID = string(project.ID)
					}
				}
			}

			// Use TaskService if available
			if services.TaskService != nil {
				// Build filters
				filters := ports.TaskFilters{
					Limit: limit,
					Tags:  domain.Tags(tags),
				}

				if projectID != "" {
					pid := domain.ProjectID(projectID)
					filters.ProjectID = &pid
				}

				if status != "" {
					s := domain.TaskStatus(status)
					filters.Status = &s
				}

				if priority != "" {
					p := domain.Priority(priority)
					filters.Priority = &p
				}

				if assignee != "" {
					filters.Assignee = &assignee
				}

				if overdue {
					filters.IsOverdue = &overdue
				}

				tasks, err := services.TaskService.ListTasks(ctx, filters)
				if err != nil {
					return fmt.Errorf("failed to list tasks: %w", err)
				}

				if len(tasks) == 0 {
					fmt.Println("No tasks found")
					return nil
				}

				output, err := json.MarshalIndent(tasks, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to format output: %w", err)
				}

				fmt.Println(string(output))
				return nil
			}

			// Fallback to MemoryService
			var pid *domain.ProjectID
			if projectID != "" {
				p := domain.ProjectID(projectID)
				pid = &p
			}

			taskType := domain.MemoryTypeTask
			memories, err := services.MemoryService.ListMemories(ctx, ports.ListMemoriesRequest{
				ProjectID: pid,
				Type:      &taskType,
				Tags:      domain.Tags(tags),
				Limit:     limit,
			})
			if err != nil {
				return fmt.Errorf("failed to list tasks: %w", err)
			}

			if len(memories) == 0 {
				fmt.Println("No tasks found")
				return nil
			}

			output, err := json.MarshalIndent(memories, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format output: %w", err)
			}

			fmt.Println(string(output))
			return nil
		},
	}

	cmd.Flags().StringVar(&projectID, "project-id", "", "Filter by project ID")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status (todo, in_progress, done, blocked)")
	cmd.Flags().StringVar(&priority, "priority", "", "Filter by priority (low, medium, high, urgent)")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Filter by assignee")
	cmd.Flags().StringSliceVar(&tags, "tags", []string{}, "Filter by tags")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of results")
	cmd.Flags().BoolVar(&overdue, "overdue", false, "Show only overdue tasks")

	return cmd
}

func createTaskGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <task-id>",
		Short: "Get a specific task",
		Long:  "Retrieve detailed information about a specific task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Initialize services
			services, err := GetServicesForCLI(cmd)
			if err != nil {
				return err
			}

			taskID := domain.MemoryID(args[0])

			// Use TaskService if available
			if services.TaskService != nil {
				task, err := services.TaskService.GetTask(ctx, taskID)
				if err != nil {
					return fmt.Errorf("failed to get task: %w", err)
				}

				output, err := json.MarshalIndent(task, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to format output: %w", err)
				}

				fmt.Println(string(output))
				return nil
			}

			// Fallback to MemoryService
			memory, err := services.MemoryService.GetMemory(ctx, taskID)
			if err != nil {
				return fmt.Errorf("failed to get task: %w", err)
			}

			output, err := json.MarshalIndent(memory, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format output: %w", err)
			}

			fmt.Println(string(output))
			return nil
		},
	}

	return cmd
}

func createTaskUpdateCmd() *cobra.Command {
	var (
		title          string
		description    string
		status         string
		priority       string
		assignee       string
		dueDate        string
		clearDueDate   bool
		estimatedHours int
		actualHours    int
		tags           []string
	)

	cmd := &cobra.Command{
		Use:   "update <task-id>",
		Short: "Update a task",
		Long:  "Update an existing task's properties",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Initialize services
			services, err := GetServicesForCLI(cmd)
			if err != nil {
				return err
			}

			taskID := domain.MemoryID(args[0])

			// Use TaskService if available
			if services.TaskService != nil {
				// Build update request
				req := ports.UpdateTaskRequest{
					TaskID:       taskID,
					ClearDueDate: clearDueDate,
				}

				// Set fields if provided
				if title != "" {
					req.Title = &title
				}
				if description != "" {
					req.Description = &description
				}
				if status != "" {
					s := domain.TaskStatus(status)
					req.Status = &s
				}
				if priority != "" {
					p := domain.Priority(priority)
					req.Priority = &p
				}
				if assignee != "" {
					req.Assignee = &assignee
				}
				if dueDate != "" {
					parsedDate, err := time.Parse("2006-01-02", dueDate)
					if err != nil {
						return fmt.Errorf("invalid due date format. Use YYYY-MM-DD: %w", err)
					}
					req.DueDate = &parsedDate
				}
				if estimatedHours > 0 {
					req.EstimatedHours = &estimatedHours
				}
				if actualHours > 0 {
					req.ActualHours = &actualHours
				}
				if len(tags) > 0 {
					req.Tags = domain.Tags(tags)
				}

				task, err := services.TaskService.UpdateTask(ctx, req)
				if err != nil {
					return fmt.Errorf("failed to update task: %w", err)
				}

				fmt.Println("Task updated successfully:")
				output, err := json.MarshalIndent(task, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to format output: %w", err)
				}

				fmt.Println(string(output))
				return nil
			}

			// Basic update via MemoryService (limited functionality)
			fmt.Println("Basic task update functionality - use TaskService for full features")
			return fmt.Errorf("task update requires TaskService - not available in current configuration")
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "New task title")
	cmd.Flags().StringVar(&description, "description", "", "New task description")
	cmd.Flags().StringVar(&status, "status", "", "New task status (todo, in_progress, done, blocked)")
	cmd.Flags().StringVar(&priority, "priority", "", "New task priority (low, medium, high, urgent)")
	cmd.Flags().StringVar(&assignee, "assignee", "", "New assignee")
	cmd.Flags().StringVar(&dueDate, "due-date", "", "New due date (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&clearDueDate, "clear-due-date", false, "Clear the due date")
	cmd.Flags().IntVar(&estimatedHours, "estimated-hours", 0, "New estimated hours")
	cmd.Flags().IntVar(&actualHours, "actual-hours", 0, "Actual hours spent")
	cmd.Flags().StringSliceVar(&tags, "tags", []string{}, "New tags (replaces existing)")

	return cmd
}

func createTaskDeleteCmd() *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <task-id>",
		Short: "Delete a task",
		Long:  "Delete a task permanently",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Initialize services
			services, err := GetServicesForCLI(cmd)
			if err != nil {
				return err
			}

			taskID := domain.MemoryID(args[0])

			if !confirm {
				fmt.Printf("Are you sure you want to delete task %s? This action cannot be undone.\n", taskID)
				fmt.Print("Type 'yes' to confirm: ")
				var response string
				_, _ = fmt.Scanln(&response)
				if strings.ToLower(response) != "yes" {
					fmt.Println("Delete cancelled")
					return nil
				}
			}

			// Use TaskService if available
			if services.TaskService != nil {
				err = services.TaskService.DeleteTask(ctx, taskID)
				if err != nil {
					return fmt.Errorf("failed to delete task: %w", err)
				}
			} else {
				// Fallback to MemoryService
				err = services.MemoryService.DeleteMemory(ctx, taskID)
				if err != nil {
					return fmt.Errorf("failed to delete task: %w", err)
				}
			}

			fmt.Printf("Task %s deleted successfully\n", taskID)
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")

	return cmd
}

func createTaskStatsCmd() *cobra.Command {
	var projectID string

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show task statistics",
		Long:  "Display task statistics and analytics for a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Initialize services
			services, err := GetServicesForCLI(cmd)
			if err != nil {
				return err
			}

			// Get project ID if not provided
			if projectID == "" {
				wd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get working directory: %w", err)
				}

				project, err := services.ProjectService.GetProjectByPath(ctx, wd)
				if err != nil {
					return fmt.Errorf("no project found in current directory. Use --project-id or run from a project directory")
				}
				projectID = string(project.ID)
			}

			// Use TaskService if available
			if services.TaskService != nil {
				stats, err := services.TaskService.GetTaskStatistics(ctx, domain.ProjectID(projectID))
				if err != nil {
					return fmt.Errorf("failed to get task statistics: %w", err)
				}

				// Display formatted statistics
				fmt.Printf("Task Statistics for Project: %s\n", projectID)
				fmt.Println(strings.Repeat("=", 50))
				fmt.Printf("Total Tasks: %d\n", stats.TotalTasks)
				fmt.Printf("Completed: %d\n", stats.CompletedTasks)
				fmt.Printf("In Progress: %d\n", stats.InProgressTasks)
				fmt.Printf("Todo: %d\n", stats.TodoTasks)
				fmt.Printf("Blocked: %d\n", stats.BlockedTasks)
				fmt.Printf("Overdue: %d\n", stats.OverdueTasks)
				fmt.Printf("Completion Rate: %.1f%%\n", stats.CompletionRate)

				if stats.TotalHours > 0 {
					fmt.Printf("Total Hours: %d\n", stats.TotalHours)
					fmt.Printf("Average Hours per Task: %.1f\n", stats.AverageHours)
				}

				if len(stats.TasksByPriority) > 0 {
					fmt.Println("\nTasks by Priority:")
					for priority, count := range stats.TasksByPriority {
						fmt.Printf("  %s: %d\n", priority, count)
					}
				}

				if len(stats.TasksByAssignee) > 0 {
					fmt.Println("\nTasks by Assignee:")
					for assignee, count := range stats.TasksByAssignee {
						fmt.Printf("  %s: %d\n", assignee, count)
					}
				}

				return nil
			}

			// Basic stats via MemoryService
			taskType := domain.MemoryTypeTask
			pid := domain.ProjectID(projectID)
			memories, err := services.MemoryService.ListMemories(ctx, ports.ListMemoriesRequest{
				ProjectID: &pid,
				Type:      &taskType,
				Limit:     1000, // Get all tasks for stats
			})
			if err != nil {
				return fmt.Errorf("failed to get task memories: %w", err)
			}

			fmt.Printf("Basic Task Statistics for Project: %s\n", projectID)
			fmt.Println(strings.Repeat("=", 50))
			fmt.Printf("Total Tasks: %d\n", len(memories))
			fmt.Println("(For detailed statistics, TaskService is required)")

			return nil
		},
	}

	cmd.Flags().StringVar(&projectID, "project-id", "", "Project ID (auto-detected if not provided)")

	return cmd
}
