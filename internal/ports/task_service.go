package ports

import (
	"context"
	"time"

	"github.com/joern1811/memory-bank/internal/domain"
)

// TaskService defines the interface for task management operations
type TaskService interface {
	// Task CRUD operations
	CreateTask(ctx context.Context, req CreateTaskRequest) (*domain.Task, error)
	GetTask(ctx context.Context, taskID domain.MemoryID) (*domain.Task, error)
	UpdateTask(ctx context.Context, req UpdateTaskRequest) (*domain.Task, error)
	DeleteTask(ctx context.Context, taskID domain.MemoryID) error

	// Task status management
	UpdateTaskStatus(ctx context.Context, taskID domain.MemoryID, status domain.TaskStatus) error

	// Task filtering and search
	ListTasks(ctx context.Context, filters TaskFilters) ([]*domain.Task, error)
	GetTasksByProject(ctx context.Context, projectID domain.ProjectID) ([]*domain.Task, error)
	GetTasksByAssignee(ctx context.Context, assignee string) ([]*domain.Task, error)
	GetTasksByStatus(ctx context.Context, status domain.TaskStatus) ([]*domain.Task, error)
	GetTasksByPriority(ctx context.Context, priority domain.Priority) ([]*domain.Task, error)
	GetOverdueTasks(ctx context.Context) ([]*domain.Task, error)

	// Task dependency management
	AddTaskDependency(ctx context.Context, taskID, dependencyID domain.MemoryID) error
	RemoveTaskDependency(ctx context.Context, taskID, dependencyID domain.MemoryID) error
	GetTaskDependencies(ctx context.Context, taskID domain.MemoryID) ([]*domain.Task, error)
	GetTaskDependents(ctx context.Context, taskID domain.MemoryID) ([]*domain.Task, error)

	// Task hierarchy management
	AddSubtask(ctx context.Context, parentID, subtaskID domain.MemoryID) error
	RemoveSubtask(ctx context.Context, parentID, subtaskID domain.MemoryID) error
	GetSubtasks(ctx context.Context, parentID domain.MemoryID) ([]*domain.Task, error)
	GetParentTask(ctx context.Context, taskID domain.MemoryID) (*domain.Task, error)

	// Task analytics
	GetTaskStatistics(ctx context.Context, projectID domain.ProjectID) (*TaskStatistics, error)
	GetTaskEfficiencyReport(ctx context.Context, projectID domain.ProjectID) (*EfficiencyReport, error)
}

// CreateTaskRequest represents a request to create a new task
type CreateTaskRequest struct {
	ProjectID      domain.ProjectID
	Title          string
	Description    string
	Priority       domain.Priority
	DueDate        *time.Time
	Assignee       string
	EstimatedHours *int
	ParentTask     *domain.MemoryID
	Dependencies   []domain.MemoryID
	Tags           domain.Tags
}

// UpdateTaskRequest represents a request to update a task
type UpdateTaskRequest struct {
	TaskID         domain.MemoryID
	Title          *string
	Description    *string
	Status         *domain.TaskStatus
	Priority       *domain.Priority
	DueDate        *time.Time
	ClearDueDate   bool
	Assignee       *string
	EstimatedHours *int
	ActualHours    *int
	Tags           domain.Tags
}

// TaskFilters represents filters for task queries
type TaskFilters struct {
	ProjectID     *domain.ProjectID
	Status        *domain.TaskStatus
	Priority      *domain.Priority
	Assignee      *string
	DueBefore     *time.Time
	DueAfter      *time.Time
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	Tags          domain.Tags
	HasDueDate    *bool
	IsOverdue     *bool
	ParentTask    *domain.MemoryID
	HasSubtasks   *bool
	Limit         int
	Offset        int
	SortBy        string // "priority", "due_date", "created_at", "updated_at", "title"
	SortOrder     string // "asc", "desc"
}

// TaskStatistics represents task statistics for a project
type TaskStatistics struct {
	TotalTasks      int                     `json:"total_tasks"`
	CompletedTasks  int                     `json:"completed_tasks"`
	InProgressTasks int                     `json:"in_progress_tasks"`
	TodoTasks       int                     `json:"todo_tasks"`
	BlockedTasks    int                     `json:"blocked_tasks"`
	OverdueTasks    int                     `json:"overdue_tasks"`
	TasksByPriority map[domain.Priority]int `json:"tasks_by_priority"`
	TasksByAssignee map[string]int          `json:"tasks_by_assignee"`
	AverageHours    float64                 `json:"average_hours"`
	TotalHours      int                     `json:"total_hours"`
	CompletionRate  float64                 `json:"completion_rate"`
}

// EfficiencyReport represents efficiency metrics for tasks
type EfficiencyReport struct {
	TotalTasksWithEstimates int     `json:"total_tasks_with_estimates"`
	AverageEfficiencyRatio  float64 `json:"average_efficiency_ratio"`
	TasksOnTime             int     `json:"tasks_on_time"`
	TasksOverTime           int     `json:"tasks_over_time"`
	TasksUnderTime          int     `json:"tasks_under_time"`
	MostEfficientAssignee   string  `json:"most_efficient_assignee"`
	LeastEfficientAssignee  string  `json:"least_efficient_assignee"`
}
