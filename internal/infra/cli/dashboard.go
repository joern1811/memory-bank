package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/spf13/cobra"
)

// dashboardCmd provides dashboard functionality
var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Show project dashboard",
	Long:  "Display a comprehensive dashboard with tasks, sessions, and project overview",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Initialize services
		services, err := GetServicesForCLI(cmd)
		if err != nil {
			return err
		}

		// Get current project
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		project, err := services.ProjectService.GetProjectByPath(ctx, wd)
		if err != nil {
			return fmt.Errorf("no project found in current directory")
		}

		return showDashboard(ctx, services, project)
	},
}

func showDashboard(ctx context.Context, services *ServiceContainer, project *domain.Project) error {
	fmt.Printf("📊 Memory Bank Dashboard - %s\n", project.Name)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Project Path: %s\n", project.Path)
	fmt.Printf("Created: %s\n", project.CreatedAt.Format("2006-01-02 15:04"))
	if project.Description != "" {
		fmt.Printf("Description: %s\n", project.Description)
	}
	fmt.Println()

	// Show task overview
	if err := showTaskOverview(ctx, services, project.ID); err != nil {
		fmt.Printf("Warning: Failed to load task overview: %v\n", err)
	}

	// Show recent sessions
	if err := showRecentSessions(ctx, services, project.ID); err != nil {
		fmt.Printf("Warning: Failed to load sessions: %v\n", err)
	}

	// Show memory statistics
	if err := showMemoryStats(ctx, services, project.ID); err != nil {
		fmt.Printf("Warning: Failed to load memory statistics: %v\n", err)
	}

	return nil
}

func showTaskOverview(ctx context.Context, services *ServiceContainer, projectID domain.ProjectID) error {
	fmt.Println("📋 Task Overview")
	fmt.Println(strings.Repeat("-", 40))

	// Get all tasks (memories with type task)
	taskType := domain.MemoryTypeTask
	tasks, err := services.MemoryService.ListMemories(ctx, ports.ListMemoriesRequest{
		ProjectID: &projectID,
		Type:      &taskType,
		Limit:     100,
	})
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found. Create your first task with: memory-bank task create")
		fmt.Println()
		return nil
	}

	// Count tasks by status (parsed from context)
	statusCounts := map[string]int{
		"todo":        0,
		"in_progress": 0,
		"done":        0,
		"blocked":     0,
		"unknown":     0,
	}

	recentTasks := make([]*domain.Memory, 0, 5)
	for i, task := range tasks {
		// Try to extract status from context
		status := extractStatusFromContext(task.Context)
		statusCounts[status]++

		// Keep track of recent tasks (first 5)
		if i < 5 {
			recentTasks = append(recentTasks, task)
		}
	}

	// Display summary
	total := len(tasks)
	completed := statusCounts["done"]
	completionRate := 0.0
	if total > 0 {
		completionRate = float64(completed) / float64(total) * 100
	}

	fmt.Printf("Total Tasks: %d\n", total)
	fmt.Printf("✅ Done: %d  🔄 In Progress: %d  📝 Todo: %d  🚫 Blocked: %d\n",
		statusCounts["done"], statusCounts["in_progress"], statusCounts["todo"], statusCounts["blocked"])
	fmt.Printf("Completion Rate: %.1f%%\n", completionRate)
	fmt.Println()

	// Show recent tasks
	if len(recentTasks) > 0 {
		fmt.Println("Recent Tasks:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "ID\tTitle\tStatus\tCreated")
		_, _ = fmt.Fprintln(w, "--\t-----\t------\t-------")

		for _, task := range recentTasks {
			id := string(task.ID)
			if len(id) > 8 {
				id = id[:8] + "..."
			}
			title := task.Title
			if len(title) > 30 {
				title = title[:27] + "..."
			}
			status := extractStatusFromContext(task.Context)
			statusEmoji := getStatusEmoji(status)

			_, _ = fmt.Fprintf(w, "%s\t%s\t%s %s\t%s\n",
				id, title, statusEmoji, status, task.CreatedAt.Format("01-02"))
		}
		_ = w.Flush()
		fmt.Println()
	}

	return nil
}

func showRecentSessions(ctx context.Context, services *ServiceContainer, projectID domain.ProjectID) error {
	fmt.Println("🔄 Recent Sessions")
	fmt.Println(strings.Repeat("-", 40))

	// Get recent sessions
	sessions, err := services.SessionService.ListSessions(ctx, ports.SessionFilters{
		ProjectID: &projectID,
		Limit:     5,
	})
	if err != nil {
		return err
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found. Start a session with: memory-bank session start")
		fmt.Println()
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "Name\tStatus\tDuration\tProgress")
	_, _ = fmt.Fprintln(w, "----\t------\t--------\t--------")

	for _, session := range sessions {
		name := session.Name
		if len(name) > 25 {
			name = name[:22] + "..."
		}

		statusEmoji := getSessionStatusEmoji(session.Status)
		duration := formatDuration(session.Duration())
		progressCount := len(session.Progress)

		_, _ = fmt.Fprintf(w, "%s\t%s %s\t%s\t%d entries\n",
			name, statusEmoji, session.Status, duration, progressCount)
	}
	_ = w.Flush()
	fmt.Println()

	return nil
}

func showMemoryStats(ctx context.Context, services *ServiceContainer, projectID domain.ProjectID) error {
	fmt.Println("🧠 Memory Statistics")
	fmt.Println(strings.Repeat("-", 40))

	// Get all memories for the project
	memories, err := services.MemoryService.ListMemories(ctx, ports.ListMemoriesRequest{
		ProjectID: &projectID,
		Limit:     1000, // Get a large number for statistics
	})
	if err != nil {
		return err
	}

	if len(memories) == 0 {
		fmt.Println("No memories found.")
		fmt.Println()
		return nil
	}

	// Count by type
	typeCounts := make(map[domain.MemoryType]int)
	for _, memory := range memories {
		typeCounts[memory.Type]++
	}

	// Display statistics
	fmt.Printf("Total Memories: %d\n", len(memories))

	typeEmojis := map[domain.MemoryType]string{
		domain.MemoryTypeDecision:      "🎯",
		domain.MemoryTypePattern:       "🔄",
		domain.MemoryTypeErrorSolution: "🐛",
		domain.MemoryTypeCode:          "💻",
		domain.MemoryTypeDocumentation: "📚",
		domain.MemoryTypeSession:       "🔄",
		domain.MemoryTypeTask:          "📋",
	}

	for memType, count := range typeCounts {
		emoji := typeEmojis[memType]
		if emoji == "" {
			emoji = "📝"
		}
		fmt.Printf("%s %s: %d\n", emoji, memType, count)
	}
	fmt.Println()

	return nil
}

// Helper functions

func extractStatusFromContext(context string) string {
	context = strings.ToLower(context)
	if strings.Contains(context, "status: done") || strings.Contains(context, "status: completed") {
		return "done"
	}
	if strings.Contains(context, "status: in_progress") || strings.Contains(context, "status: progress") {
		return "in_progress"
	}
	if strings.Contains(context, "status: todo") || strings.Contains(context, "status: pending") {
		return "todo"
	}
	if strings.Contains(context, "status: blocked") {
		return "blocked"
	}
	return "unknown"
}

func getStatusEmoji(status string) string {
	switch status {
	case "done":
		return "✅"
	case "in_progress":
		return "🔄"
	case "todo":
		return "📝"
	case "blocked":
		return "🚫"
	default:
		return "❓"
	}
}

func getSessionStatusEmoji(status domain.SessionStatus) string {
	switch status {
	case domain.SessionStatusActive:
		return "🟢"
	case domain.SessionStatusCompleted:
		return "✅"
	case domain.SessionStatusPaused:
		return "⏸️"
	case domain.SessionStatusAborted:
		return "❌"
	default:
		return "❓"
	}
}

func formatDuration(duration time.Duration) string {
	if duration < time.Hour {
		return fmt.Sprintf("%.0fm", duration.Minutes())
	}
	hours := duration.Hours()
	minutes := int(duration.Minutes()) % 60
	return fmt.Sprintf("%.0fh%dm", hours, minutes)
}

func init() {
	// Add dashboard command to root
	rootCmd.AddCommand(dashboardCmd)
}
