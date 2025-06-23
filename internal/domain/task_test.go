package domain

import (
	"testing"
	"time"
)

func TestNewTask(t *testing.T) {
	projectID := ProjectID("test-project")
	title := "Test Task"
	description := "Test Description"
	priority := PriorityHigh

	task := NewTask(projectID, title, description, priority)

	// Verify basic properties
	if task.Memory == nil {
		t.Error("Expected task to have a memory object")
	}
	if task.Memory.ProjectID != projectID {
		t.Errorf("Expected project ID %s, got %s", projectID, task.Memory.ProjectID)
	}
	if task.Memory.Type != MemoryTypeTask {
		t.Errorf("Expected memory type %s, got %s", MemoryTypeTask, task.Memory.Type)
	}
	if task.Memory.Title != title {
		t.Errorf("Expected title %s, got %s", title, task.Memory.Title)
	}
	if task.Memory.Content != description {
		t.Errorf("Expected content %s, got %s", description, task.Memory.Content)
	}

	// Verify task-specific properties
	if task.Status != TaskStatusTodo {
		t.Errorf("Expected status %s, got %s", TaskStatusTodo, task.Status)
	}
	if task.Priority != priority {
		t.Errorf("Expected priority %s, got %s", priority, task.Priority)
	}
	if len(task.Dependencies) != 0 {
		t.Errorf("Expected empty dependencies, got %d", len(task.Dependencies))
	}
	if len(task.Subtasks) != 0 {
		t.Errorf("Expected empty subtasks, got %d", len(task.Subtasks))
	}
}

func TestNewTaskWithDetails(t *testing.T) {
	projectID := ProjectID("test-project")
	title := "Test Task"
	description := "Test Description"
	priority := PriorityMedium
	dueDate := time.Now().Add(24 * time.Hour)
	assignee := "john.doe"
	estimatedHours := 8

	task := NewTaskWithDetails(projectID, title, description, priority, &dueDate, assignee, &estimatedHours)

	if task.DueDate == nil {
		t.Error("Expected due date to be set")
	}
	if !task.DueDate.Equal(dueDate) {
		t.Errorf("Expected due date %v, got %v", dueDate, *task.DueDate)
	}
	if task.Assignee != assignee {
		t.Errorf("Expected assignee %s, got %s", assignee, task.Assignee)
	}
	if task.EstimatedHours == nil {
		t.Error("Expected estimated hours to be set")
	}
	if *task.EstimatedHours != estimatedHours {
		t.Errorf("Expected estimated hours %d, got %d", estimatedHours, *task.EstimatedHours)
	}
}

func TestTaskUpdateStatus(t *testing.T) {
	task := NewTask(ProjectID("test"), "Test", "Description", PriorityMedium)
	oldUpdatedAt := task.Memory.UpdatedAt

	// Wait a small amount to ensure time difference
	time.Sleep(1 * time.Millisecond)

	task.UpdateStatus(TaskStatusInProgress)

	if task.Status != TaskStatusInProgress {
		t.Errorf("Expected status %s, got %s", TaskStatusInProgress, task.Status)
	}
	if !task.Memory.UpdatedAt.After(oldUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestTaskAddDependency(t *testing.T) {
	task := NewTask(ProjectID("test"), "Test", "Description", PriorityMedium)
	depID := MemoryID("dep-123")

	task.AddDependency(depID)

	if len(task.Dependencies) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(task.Dependencies))
	}
	if task.Dependencies[0] != depID {
		t.Errorf("Expected dependency %s, got %s", depID, task.Dependencies[0])
	}

	// Test adding same dependency again (should not duplicate)
	task.AddDependency(depID)
	if len(task.Dependencies) != 1 {
		t.Errorf("Expected 1 dependency after duplicate add, got %d", len(task.Dependencies))
	}
}

func TestTaskRemoveDependency(t *testing.T) {
	task := NewTask(ProjectID("test"), "Test", "Description", PriorityMedium)
	depID1 := MemoryID("dep-1")
	depID2 := MemoryID("dep-2")

	task.AddDependency(depID1)
	task.AddDependency(depID2)

	if len(task.Dependencies) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(task.Dependencies))
	}

	task.RemoveDependency(depID1)

	if len(task.Dependencies) != 1 {
		t.Errorf("Expected 1 dependency after removal, got %d", len(task.Dependencies))
	}
	if task.Dependencies[0] != depID2 {
		t.Errorf("Expected remaining dependency %s, got %s", depID2, task.Dependencies[0])
	}
}

func TestTaskAddSubtask(t *testing.T) {
	task := NewTask(ProjectID("test"), "Test", "Description", PriorityMedium)
	subtaskID := MemoryID("subtask-123")

	task.AddSubtask(subtaskID)

	if len(task.Subtasks) != 1 {
		t.Errorf("Expected 1 subtask, got %d", len(task.Subtasks))
	}
	if task.Subtasks[0] != subtaskID {
		t.Errorf("Expected subtask %s, got %s", subtaskID, task.Subtasks[0])
	}

	// Test adding same subtask again (should not duplicate)
	task.AddSubtask(subtaskID)
	if len(task.Subtasks) != 1 {
		t.Errorf("Expected 1 subtask after duplicate add, got %d", len(task.Subtasks))
	}
}

func TestTaskParentTaskManagement(t *testing.T) {
	task := NewTask(ProjectID("test"), "Test", "Description", PriorityMedium)
	parentID := MemoryID("parent-123")

	// Test setting parent task
	task.SetParentTask(parentID)
	if task.ParentTask == nil {
		t.Error("Expected parent task to be set")
	}
	if *task.ParentTask != parentID {
		t.Errorf("Expected parent task %s, got %s", parentID, *task.ParentTask)
	}

	// Test clearing parent task
	task.ClearParentTask()
	if task.ParentTask != nil {
		t.Error("Expected parent task to be cleared")
	}
}

func TestTaskPriorityUpdate(t *testing.T) {
	task := NewTask(ProjectID("test"), "Test", "Description", PriorityLow)
	
	task.UpdatePriority(PriorityUrgent)
	
	if task.Priority != PriorityUrgent {
		t.Errorf("Expected priority %s, got %s", PriorityUrgent, task.Priority)
	}
}

func TestTaskEstimateUpdate(t *testing.T) {
	task := NewTask(ProjectID("test"), "Test", "Description", PriorityMedium)
	
	task.UpdateEstimate(16)
	
	if task.EstimatedHours == nil {
		t.Error("Expected estimated hours to be set")
	}
	if *task.EstimatedHours != 16 {
		t.Errorf("Expected estimated hours 16, got %d", *task.EstimatedHours)
	}
}

func TestTaskActualHoursLogging(t *testing.T) {
	task := NewTask(ProjectID("test"), "Test", "Description", PriorityMedium)
	
	task.LogActualHours(12)
	
	if task.ActualHours == nil {
		t.Error("Expected actual hours to be set")
	}
	if *task.ActualHours != 12 {
		t.Errorf("Expected actual hours 12, got %d", *task.ActualHours)
	}
}

func TestTaskDueDateManagement(t *testing.T) {
	task := NewTask(ProjectID("test"), "Test", "Description", PriorityMedium)
	dueDate := time.Now().Add(48 * time.Hour)

	// Test setting due date
	task.SetDueDate(dueDate)
	if task.DueDate == nil {
		t.Error("Expected due date to be set")
	}
	if !task.DueDate.Equal(dueDate) {
		t.Errorf("Expected due date %v, got %v", dueDate, *task.DueDate)
	}

	// Test clearing due date
	task.ClearDueDate()
	if task.DueDate != nil {
		t.Error("Expected due date to be cleared")
	}
}

func TestTaskAssigneeManagement(t *testing.T) {
	task := NewTask(ProjectID("test"), "Test", "Description", PriorityMedium)
	assignee := "jane.doe"

	// Test assigning
	task.AssignTo(assignee)
	if task.Assignee != assignee {
		t.Errorf("Expected assignee %s, got %s", assignee, task.Assignee)
	}

	// Test unassigning
	task.Unassign()
	if task.Assignee != "" {
		t.Errorf("Expected empty assignee, got %s", task.Assignee)
	}
}

func TestTaskStatusChecks(t *testing.T) {
	task := NewTask(ProjectID("test"), "Test", "Description", PriorityMedium)

	// Test initial status (todo)
	if !task.IsTodo() {
		t.Error("Expected task to be todo initially")
	}
	if task.IsCompleted() || task.IsInProgress() || task.IsBlocked() {
		t.Error("Expected task to only be todo initially")
	}

	// Test in progress
	task.UpdateStatus(TaskStatusInProgress)
	if !task.IsInProgress() {
		t.Error("Expected task to be in progress")
	}
	if task.IsTodo() || task.IsCompleted() || task.IsBlocked() {
		t.Error("Expected task to only be in progress")
	}

	// Test completed
	task.UpdateStatus(TaskStatusDone)
	if !task.IsCompleted() {
		t.Error("Expected task to be completed")
	}
	if task.IsTodo() || task.IsInProgress() || task.IsBlocked() {
		t.Error("Expected task to only be completed")
	}

	// Test blocked
	task.UpdateStatus(TaskStatusBlocked)
	if !task.IsBlocked() {
		t.Error("Expected task to be blocked")
	}
	if task.IsTodo() || task.IsInProgress() || task.IsCompleted() {
		t.Error("Expected task to only be blocked")
	}
}

func TestTaskOverdueCheck(t *testing.T) {
	task := NewTask(ProjectID("test"), "Test", "Description", PriorityMedium)

	// Task without due date should not be overdue
	if task.IsOverdue() {
		t.Error("Expected task without due date to not be overdue")
	}

	// Task with future due date should not be overdue
	futureDate := time.Now().Add(24 * time.Hour)
	task.SetDueDate(futureDate)
	if task.IsOverdue() {
		t.Error("Expected task with future due date to not be overdue")
	}

	// Task with past due date should be overdue (unless completed)
	pastDate := time.Now().Add(-24 * time.Hour)
	task.SetDueDate(pastDate)
	if !task.IsOverdue() {
		t.Error("Expected task with past due date to be overdue")
	}

	// Completed task with past due date should not be overdue
	task.UpdateStatus(TaskStatusDone)
	if task.IsOverdue() {
		t.Error("Expected completed task to not be overdue even with past due date")
	}
}

func TestTaskEfficiencyRatio(t *testing.T) {
	task := NewTask(ProjectID("test"), "Test", "Description", PriorityMedium)

	// Task without hours should have no efficiency ratio
	ratio := task.GetEfficiencyRatio()
	if ratio != nil {
		t.Error("Expected no efficiency ratio without hours")
	}

	// Task with only estimated hours should have no ratio
	task.UpdateEstimate(8)
	ratio = task.GetEfficiencyRatio()
	if ratio != nil {
		t.Error("Expected no efficiency ratio with only estimated hours")
	}

	// Task with both estimated and actual hours should have ratio
	task.LogActualHours(10)
	ratio = task.GetEfficiencyRatio()
	if ratio == nil {
		t.Error("Expected efficiency ratio with both estimated and actual hours")
	}
	expectedRatio := 8.0 / 10.0 // estimated / actual
	if *ratio != expectedRatio {
		t.Errorf("Expected efficiency ratio %f, got %f", expectedRatio, *ratio)
	}

	// Task with zero estimated hours should have no ratio
	task.UpdateEstimate(0)
	ratio = task.GetEfficiencyRatio()
	if ratio != nil {
		t.Error("Expected no efficiency ratio with zero estimated hours")
	}
}

func TestTaskDependencyAndSubtaskChecks(t *testing.T) {
	task := NewTask(ProjectID("test"), "Test", "Description", PriorityMedium)

	// Initially should have no dependencies or subtasks
	if task.HasDependencies() {
		t.Error("Expected task to have no dependencies initially")
	}
	if task.HasSubtasks() {
		t.Error("Expected task to have no subtasks initially")
	}

	// Add dependency and check
	task.AddDependency(MemoryID("dep-1"))
	if !task.HasDependencies() {
		t.Error("Expected task to have dependencies after adding one")
	}

	// Add subtask and check
	task.AddSubtask(MemoryID("subtask-1"))
	if !task.HasSubtasks() {
		t.Error("Expected task to have subtasks after adding one")
	}
}