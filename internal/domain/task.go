package domain

import (
	"time"
)

// Task represents a task or to-do item stored as a memory
type Task struct {
	*Memory
	Status         TaskStatus  `json:"status"`
	Priority       Priority    `json:"priority"`
	DueDate        *time.Time  `json:"due_date,omitempty"`
	Assignee       string      `json:"assignee,omitempty"`
	Dependencies   []MemoryID  `json:"dependencies,omitempty"`
	EstimatedHours *int        `json:"estimated_hours,omitempty"`
	ActualHours    *int        `json:"actual_hours,omitempty"`
	ParentTask     *MemoryID   `json:"parent_task,omitempty"`
	Subtasks       []MemoryID  `json:"subtasks,omitempty"`
}

// NewTask creates a new task memory
func NewTask(projectID ProjectID, title, description string, priority Priority) *Task {
	memory := NewMemory(projectID, MemoryTypeTask, title, description, "")
	return &Task{
		Memory:       memory,
		Status:       TaskStatusTodo,
		Priority:     priority,
		Dependencies: make([]MemoryID, 0),
		Subtasks:     make([]MemoryID, 0),
	}
}

// NewTaskWithDetails creates a new task with detailed information
func NewTaskWithDetails(projectID ProjectID, title, description string, priority Priority, dueDate *time.Time, assignee string, estimatedHours *int) *Task {
	task := NewTask(projectID, title, description, priority)
	task.DueDate = dueDate
	task.Assignee = assignee
	task.EstimatedHours = estimatedHours
	return task
}

// UpdateStatus updates the task status
func (t *Task) UpdateStatus(status TaskStatus) {
	t.Status = status
	t.UpdatedAt = time.Now()
}

// AddDependency adds a dependency to the task
func (t *Task) AddDependency(dependencyID MemoryID) {
	if !t.hasDependency(dependencyID) {
		t.Dependencies = append(t.Dependencies, dependencyID)
		t.UpdatedAt = time.Now()
	}
}

// RemoveDependency removes a dependency from the task
func (t *Task) RemoveDependency(dependencyID MemoryID) {
	for i, dep := range t.Dependencies {
		if dep == dependencyID {
			t.Dependencies = append(t.Dependencies[:i], t.Dependencies[i+1:]...)
			t.UpdatedAt = time.Now()
			break
		}
	}
}

// AddSubtask adds a subtask to the task
func (t *Task) AddSubtask(subtaskID MemoryID) {
	if !t.hasSubtask(subtaskID) {
		t.Subtasks = append(t.Subtasks, subtaskID)
		t.UpdatedAt = time.Now()
	}
}

// RemoveSubtask removes a subtask from the task
func (t *Task) RemoveSubtask(subtaskID MemoryID) {
	for i, subtask := range t.Subtasks {
		if subtask == subtaskID {
			t.Subtasks = append(t.Subtasks[:i], t.Subtasks[i+1:]...)
			t.UpdatedAt = time.Now()
			break
		}
	}
}

// SetParentTask sets the parent task
func (t *Task) SetParentTask(parentID MemoryID) {
	t.ParentTask = &parentID
	t.UpdatedAt = time.Now()
}

// ClearParentTask clears the parent task
func (t *Task) ClearParentTask() {
	t.ParentTask = nil
	t.UpdatedAt = time.Now()
}

// UpdatePriority updates the task priority
func (t *Task) UpdatePriority(priority Priority) {
	t.Priority = priority
	t.UpdatedAt = time.Now()
}

// UpdateEstimate updates the estimated hours
func (t *Task) UpdateEstimate(hours int) {
	t.EstimatedHours = &hours
	t.UpdatedAt = time.Now()
}

// LogActualHours logs the actual hours spent
func (t *Task) LogActualHours(hours int) {
	t.ActualHours = &hours
	t.UpdatedAt = time.Now()
}

// SetDueDate sets the due date
func (t *Task) SetDueDate(dueDate time.Time) {
	t.DueDate = &dueDate
	t.UpdatedAt = time.Now()
}

// ClearDueDate clears the due date
func (t *Task) ClearDueDate() {
	t.DueDate = nil
	t.UpdatedAt = time.Now()
}

// AssignTo assigns the task to someone
func (t *Task) AssignTo(assignee string) {
	t.Assignee = assignee
	t.UpdatedAt = time.Now()
}

// Unassign removes the assignee
func (t *Task) Unassign() {
	t.Assignee = ""
	t.UpdatedAt = time.Now()
}

// IsCompleted checks if the task is completed
func (t *Task) IsCompleted() bool {
	return t.Status == TaskStatusDone
}

// IsBlocked checks if the task is blocked
func (t *Task) IsBlocked() bool {
	return t.Status == TaskStatusBlocked
}

// IsInProgress checks if the task is in progress
func (t *Task) IsInProgress() bool {
	return t.Status == TaskStatusInProgress
}

// IsTodo checks if the task is in todo status
func (t *Task) IsTodo() bool {
	return t.Status == TaskStatusTodo
}

// IsOverdue checks if the task is overdue
func (t *Task) IsOverdue() bool {
	if t.DueDate == nil {
		return false
	}
	return time.Now().After(*t.DueDate) && !t.IsCompleted()
}

// GetEfficiencyRatio returns the efficiency ratio (estimated vs actual hours)
func (t *Task) GetEfficiencyRatio() *float64 {
	if t.EstimatedHours == nil || t.ActualHours == nil || *t.EstimatedHours == 0 {
		return nil
	}
	ratio := float64(*t.EstimatedHours) / float64(*t.ActualHours)
	return &ratio
}

// HasDependencies checks if the task has dependencies
func (t *Task) HasDependencies() bool {
	return len(t.Dependencies) > 0
}

// HasSubtasks checks if the task has subtasks
func (t *Task) HasSubtasks() bool {
	return len(t.Subtasks) > 0
}

// Helper methods

func (t *Task) hasDependency(dependencyID MemoryID) bool {
	for _, dep := range t.Dependencies {
		if dep == dependencyID {
			return true
		}
	}
	return false
}

func (t *Task) hasSubtask(subtaskID MemoryID) bool {
	for _, subtask := range t.Subtasks {
		if subtask == subtaskID {
			return true
		}
	}
	return false
}