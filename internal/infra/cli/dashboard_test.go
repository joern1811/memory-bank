package cli

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/joern1811/memory-bank/internal/app"
	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/infra/database"
	"github.com/joern1811/memory-bank/internal/infra/embedding"
	"github.com/joern1811/memory-bank/internal/infra/vector"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/sirupsen/logrus"
)

func setupTestDashboard(t *testing.T) (*ServiceContainer, *domain.Project) {
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
	sessionRepo := database.NewSQLiteSessionRepository(db, logger)

	// Use mock providers for testing
	embeddingProvider := embedding.NewMockEmbeddingProvider(768, logger)
	vectorStore := vector.NewMockVectorStore(logger)

	// Initialize services
	memoryService := app.NewMemoryService(memoryRepo, embeddingProvider, vectorStore, logger)
	projectService := app.NewProjectService(projectRepo, logger)
	sessionService := app.NewSessionService(sessionRepo, projectRepo, logger)
	taskService := app.NewTaskService(memoryService, logger)

	services := &ServiceContainer{
		MemoryService:  memoryService,
		ProjectService: projectService,
		SessionService: sessionService,
		TaskService:    taskService,
	}

	// Create test project
	ctx := context.Background()
	project, err := projectService.InitializeProject(ctx, "/test/dashboard", ports.InitializeProjectRequest{
		Name:        "Dashboard Test Project",
		Description: "Test project for dashboard tests",
	})
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	return services, project
}

func createTestData(t *testing.T, services *ServiceContainer, project *domain.Project) {
	ctx := context.Background()

	// Create various memory types
	memories := []ports.CreateMemoryRequest{
		{
			ProjectID: project.ID,
			Type:      domain.MemoryTypeDecision,
			Title:     "Architecture Decision",
			Content:   "Use microservices architecture",
			Tags:      domain.Tags{"architecture", "decision"},
		},
		{
			ProjectID: project.ID,
			Type:      domain.MemoryTypePattern,
			Title:     "Repository Pattern",
			Content:   "Use repository pattern for data access",
			Tags:      domain.Tags{"pattern", "data"},
		},
		{
			ProjectID: project.ID,
			Type:      domain.MemoryTypeErrorSolution,
			Title:     "Database Connection Issue",
			Content:   "Fixed timeout error by increasing connection pool",
			Tags:      domain.Tags{"database", "error"},
		},
	}

	for _, req := range memories {
		_, err := services.MemoryService.CreateMemory(ctx, req)
		if err != nil {
			t.Fatalf("Failed to create test memory: %v", err)
		}
	}

	// Create tasks with different statuses
	tasks := []struct {
		title       string
		priority    domain.Priority
		status      string
		assignee    string
		description string
	}{
		{"Complete authentication", domain.PriorityHigh, "done", "john.doe", "Implement JWT authentication"},
		{"Add user dashboard", domain.PriorityMedium, "in_progress", "jane.doe", "Create user dashboard UI"},
		{"Setup monitoring", domain.PriorityLow, "todo", "bob.smith", "Setup application monitoring"},
		{"Fix performance bug", domain.PriorityUrgent, "blocked", "alice.jones", "Investigate slow query"},
	}

	for _, task := range tasks {
		// Create task memory with status in context
		context := "Status: " + task.status + ", Priority: " + string(task.priority) + ", Assignee: " + task.assignee

		_, err := services.MemoryService.CreateMemory(ctx, ports.CreateMemoryRequest{
			ProjectID: project.ID,
			Type:      domain.MemoryTypeTask,
			Title:     task.title,
			Content:   task.description,
			Context:   context,
			Tags:      domain.Tags{"task", string(task.priority)},
		})
		if err != nil {
			t.Fatalf("Failed to create test task: %v", err)
		}
	}

	// Create sessions
	sessions := []struct {
		name        string
		description string
		status      domain.SessionStatus
		progress    []string
	}{
		{
			"Authentication Implementation",
			"Implementing JWT authentication system",
			domain.SessionStatusCompleted,
			[]string{"Created JWT middleware", "Added login endpoint", "Implemented token validation"},
		},
		{
			"User Dashboard Development",
			"Building user dashboard interface",
			domain.SessionStatusActive,
			[]string{"Created dashboard layout", "Added user profile section"},
		},
		{
			"Performance Optimization",
			"Optimizing application performance",
			domain.SessionStatusPaused,
			[]string{"Identified slow queries", "Added database indexes"},
		},
	}

	for _, sessionData := range sessions {
		session, err := services.SessionService.StartSession(ctx, ports.StartSessionRequest{
			ProjectID:       project.ID,
			TaskDescription: sessionData.name + ": " + sessionData.description,
		})
		if err != nil {
			t.Fatalf("Failed to create test session: %v", err)
		}

		// Add progress entries
		for _, progress := range sessionData.progress {
			err = services.SessionService.LogProgress(ctx, session.ID, progress)
			if err != nil {
				t.Fatalf("Failed to log session progress: %v", err)
			}
		}

		// Update session status
		if sessionData.status != domain.SessionStatusActive {
			switch sessionData.status {
			case domain.SessionStatusCompleted:
				err = services.SessionService.CompleteSession(ctx, session.ID, "Successfully completed")
			case domain.SessionStatusPaused:
				// For paused, we'd need to implement pause functionality
				// For now, we'll leave it as active in tests
			}
			if err != nil {
				t.Fatalf("Failed to update session status: %v", err)
			}
		}
	}
}

func TestDashboardCmd(t *testing.T) {
	services, project := setupTestDashboard(t)

	// Create test data
	createTestData(t, services, project)

	// Test the dashboard functionality directly instead of trying to mock global functions
	ctx := context.Background()

	// Test the showDashboard function directly
	output := captureOutput(func() {
		err := showDashboard(ctx, services, project)
		if err != nil {
			t.Errorf("Failed to show dashboard: %v", err)
		}
	})

	// Verify dashboard content
	expectedContent := []string{
		"Memory Bank Dashboard",
		project.Name,
		"Task Overview",
		"Recent Sessions",
		"Memory Statistics",
	}

	for _, content := range expectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("Expected '%s' in dashboard output, got: %s", content, output)
		}
	}
}

func TestShowTaskOverview(t *testing.T) {
	services, project := setupTestDashboard(t)
	createTestData(t, services, project)

	ctx := context.Background()

	// Test task overview function directly
	output := captureOutput(func() {
		err := showTaskOverview(ctx, services, project.ID)
		if err != nil {
			t.Errorf("Failed to show task overview: %v", err)
		}
	})

	// Verify task overview content
	expectedContent := []string{
		"Task Overview",
		"Total Tasks:",
		"Done:",
		"Todo:",
	}

	for _, content := range expectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("Expected '%s' in task overview output, got: %s", content, output)
		}
	}
}

func TestShowTaskOverview_NoTasks(t *testing.T) {
	services, project := setupTestDashboard(t)
	// Don't create test data - test empty state

	ctx := context.Background()

	output := captureOutput(func() {
		err := showTaskOverview(ctx, services, project.ID)
		if err != nil {
			t.Errorf("Failed to show task overview: %v", err)
		}
	})

	if !strings.Contains(output, "No tasks found") {
		t.Errorf("Expected 'No tasks found' message in output, got: %s", output)
	}
	if !strings.Contains(output, "memory-bank task create") {
		t.Errorf("Expected help message in output, got: %s", output)
	}
}

func TestShowRecentSessions(t *testing.T) {
	services, project := setupTestDashboard(t)
	createTestData(t, services, project)

	ctx := context.Background()

	output := captureOutput(func() {
		err := showRecentSessions(ctx, services, project.ID)
		if err != nil {
			t.Errorf("Failed to show recent sessions: %v", err)
		}
	})

	// Verify sessions content (names may be truncated in display)
	expectedContent := []string{
		"Recent Sessions",
		"Authentication Impleme", // Truncated name
		"User Dashboard Develop", // Truncated name
		"Performance Optimizati", // Truncated name
		"Name",
		"Status",
		"Duration",
		"Progress",
	}

	for _, content := range expectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("Expected '%s' in sessions output, got: %s", content, output)
		}
	}
}

func TestShowMemoryStats(t *testing.T) {
	services, project := setupTestDashboard(t)
	createTestData(t, services, project)

	ctx := context.Background()

	output := captureOutput(func() {
		err := showMemoryStats(ctx, services, project.ID)
		if err != nil {
			t.Errorf("Failed to show memory stats: %v", err)
		}
	})

	// Verify memory statistics content
	expectedContent := []string{
		"Memory Statistics",
		"Total Memories:",
		"🎯 decision:",
		"🔄 pattern:",
		"🐛 error_solution:",
		"📋 task:",
	}

	for _, content := range expectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("Expected '%s' in memory stats output, got: %s", content, output)
		}
	}
}

func TestExtractStatusFromContext(t *testing.T) {
	testCases := []struct {
		context  string
		expected string
	}{
		{"Status: done, Priority: high", "done"},
		{"Status: in_progress, Assignee: john.doe", "in_progress"},
		{"Status: todo, Priority: medium", "todo"},
		{"Status: blocked, Issue: dependency", "blocked"},
		{"Priority: high, Assignee: jane.doe", "unknown"},
		{"Some other context without status", "unknown"},
		{"Status: completed, Priority: low", "done"}, // completed maps to done
	}

	for _, tc := range testCases {
		result := extractStatusFromContext(tc.context)
		if result != tc.expected {
			t.Errorf("For context '%s', expected status '%s', got '%s'",
				tc.context, tc.expected, result)
		}
	}
}

func TestGetStatusEmoji(t *testing.T) {
	testCases := []struct {
		status   string
		expected string
	}{
		{"done", "✅"},
		{"in_progress", "🔄"},
		{"todo", "📝"},
		{"blocked", "🚫"},
		{"unknown", "❓"},
		{"invalid", "❓"},
	}

	for _, tc := range testCases {
		result := getStatusEmoji(tc.status)
		if result != tc.expected {
			t.Errorf("For status '%s', expected emoji '%s', got '%s'",
				tc.status, tc.expected, result)
		}
	}
}

func TestGetSessionStatusEmoji(t *testing.T) {
	testCases := []struct {
		status   domain.SessionStatus
		expected string
	}{
		{domain.SessionStatusActive, "🟢"},
		{domain.SessionStatusCompleted, "✅"},
		{domain.SessionStatusPaused, "⏸️"},
		{domain.SessionStatusAborted, "❌"},
		{"invalid", "❓"},
	}

	for _, tc := range testCases {
		result := getSessionStatusEmoji(tc.status)
		if result != tc.expected {
			t.Errorf("For session status '%s', expected emoji '%s', got '%s'",
				tc.status, tc.expected, result)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	testCases := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Minute, "30m"},
		{90 * time.Minute, "2h30m"},
		{2 * time.Hour, "2h0m"},
		{5 * time.Minute, "5m"},
		{125 * time.Minute, "2h5m"},
	}

	for _, tc := range testCases {
		result := formatDuration(tc.duration)
		if result != tc.expected {
			t.Errorf("For duration %v, expected '%s', got '%s'",
				tc.duration, tc.expected, result)
		}
	}
}
