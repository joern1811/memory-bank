package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/spf13/cobra"
)

// gitIntegrationCmd provides git integration commands
var gitIntegrationCmd = &cobra.Command{
	Use:   "git",
	Short: "Git integration commands",
	Long:  "Integrate Memory Bank with Git for automatic progress tracking",
}

var gitScanCommitsCmd = &cobra.Command{
	Use:   "scan-commits",
	Short: "Scan recent git commits for task progress",
	Long:  "Scan recent git commits and automatically update task progress based on commit messages",
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

		// Get count from flag or default to 10
		count, _ := cmd.Flags().GetInt("count")
		if count <= 0 {
			count = 10
		}

		// Get recent git commits
		commits, err := getRecentGitCommits(count)
		if err != nil {
			return fmt.Errorf("failed to get git commits: %w", err)
		}

		fmt.Printf("Scanning %d recent commits for task progress...\n", len(commits))

		progressCount := 0
		for _, commit := range commits {
			// Use intelligent commit analysis
			actions := extractTaskActionsFromCommit(commit.Message)
			if len(actions) == 0 {
				continue
			}

			// Process each task action
			for _, action := range actions {
				// Try to find task by ID (assuming task IDs are Memory IDs)
				taskID := domain.MemoryID(action.TaskID)
				
				// Check if TaskService is available for intelligent updates
				if services.TaskService != nil {
					// Try to get existing task
					existingTask, err := services.TaskService.GetTask(ctx, taskID)
					if err == nil {
						// Build comprehensive update request
						updateReq := ports.UpdateTaskRequest{
							TaskID: taskID,
						}
						
						// Update status if specified
						if action.NewStatus != nil {
							updateReq.Status = action.NewStatus
						}
						
						// Set assignee based on commit author (with intelligent recognition)
						assignee := normalizeAuthorToAssignee(commit.Author, commit.Email)
						if assignee != "unknown" && assignee != existingTask.Assignee {
							updateReq.Assignee = &assignee
						}
						
						// Time tracking: estimate effort based on commit data
						if commit.FilesCount > 0 {
							// Simple heuristic: 15 minutes per file touched, max 4 hours
							estimatedMinutes := commit.FilesCount * 15
							if estimatedMinutes > 240 { // Cap at 4 hours
								estimatedMinutes = 240
							}
							actualHours := estimatedMinutes / 60
							if actualHours > 0 {
								updateReq.ActualHours = &actualHours
							}
						}
						
						// Update task
						_, err := services.TaskService.UpdateTask(ctx, updateReq)
						if err != nil {
							fmt.Printf("Warning: Failed to update task %s: %v\n", action.TaskID, err)
						} else {
							actionDesc := action.Action
							if action.NewStatus != nil {
								actionDesc = fmt.Sprintf("%s → %s", action.Action, string(*action.NewStatus))
							}
							if updateReq.ActualHours != nil {
								actionDesc += fmt.Sprintf(" (+%dh)", *updateReq.ActualHours)
							}
							fmt.Printf("✅ Updated task %s (%s) from commit %s\n", action.TaskID, actionDesc, commit.Hash[:8])
							progressCount++
						}
						
						// Also create a progress memory for historical tracking
						createProgressMemory(services, project.ID, action, commit)
						continue
					}
				}
				
				// If task not found by direct ID, try branch-based recognition
				branchTaskID := extractTaskIDFromBranch(commit.Branch)
				if branchTaskID != "" && branchTaskID == action.TaskID {
					fmt.Printf("🔗 Recognized task %s from branch %s\n", branchTaskID, commit.Branch)
				}
				
				// Fallback: create progress memory entry (original behavior)
				if createProgressMemory(services, project.ID, action, commit) {
					progressCount++
					fmt.Printf("✅ Recorded progress for task %s from commit %s\n", action.TaskID, commit.Hash[:8])
				}
			}
		}

		if progressCount == 0 {
			fmt.Println("No task references found in recent commits")
		} else {
			fmt.Printf("\n🎉 Successfully recorded %d progress entries from git commits\n", progressCount)
		}

		return nil
	},
}

var gitHookInstallCmd = &cobra.Command{
	Use:   "install-hooks",
	Short: "Install git hooks for automatic task tracking",
	Long:  "Install git hooks that automatically track task progress on commits",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if we're in a git repository
		if !isGitRepository() {
			return fmt.Errorf("not in a git repository")
		}

		// Create enhanced post-commit hook with intelligent task tracking
		hookContent := `#!/bin/sh
# Memory Bank intelligent task progress tracking

# Get the current commit information
COMMIT_HASH=$(git rev-parse HEAD)
COMMIT_MSG=$(git log -1 --pretty=%B)
COMMIT_AUTHOR=$(git log -1 --pretty=%an)
COMMIT_EMAIL=$(git log -1 --pretty=%ae)
BRANCH_NAME=$(git rev-parse --abbrev-ref HEAD)

# Check if memory-bank is available
if command -v memory-bank >/dev/null 2>&1; then
    echo "Memory Bank: Analyzing commit $COMMIT_HASH for task updates..."
    
    # Run intelligent git scan with recent commits
    memory-bank git scan-commits --count 1 >/dev/null 2>&1 || true
    
    # Log commit info for debugging (optional)
    echo "Memory Bank: Processed commit by $COMMIT_AUTHOR on branch $BRANCH_NAME"
else
    echo "Memory Bank: CLI not found, skipping task tracking"
fi
`

		hookPath := filepath.Join(".git", "hooks", "post-commit")
		
		// Write hook file
		err := os.WriteFile(hookPath, []byte(hookContent), 0755)
		if err != nil {
			return fmt.Errorf("failed to write git hook: %w", err)
		}

		fmt.Printf("✅ Installed post-commit hook at %s\n", hookPath)
		fmt.Println("📝 Now, commits with task references (e.g., 'fix #task-123') will automatically track progress")
		
		return nil
	},
}

// GitCommit represents a git commit with enhanced information
type GitCommit struct {
	Hash       string
	Author     string
	Email      string
	Date       string
	Message    string
	Branch     string
	FilesCount int
}

// getRecentGitCommits gets recent git commits with enhanced information
func getRecentGitCommits(count int) ([]GitCommit, error) {
	// Enhanced git log format: Hash|Author|Email|Date|Subject
	cmd := exec.Command("git", "log", fmt.Sprintf("-%d", count), "--pretty=format:%H|%an|%ae|%ad|%s", "--date=iso")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	commits := make([]GitCommit, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) != 5 {
			continue
		}

		// Get current branch (for context)
		branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		branchOutput, _ := branchCmd.Output()
		branch := strings.TrimSpace(string(branchOutput))

		// Get file count for this commit
		filesCmd := exec.Command("git", "show", "--name-only", "--pretty=format:", parts[0])
		filesOutput, _ := filesCmd.Output()
		filesCount := len(strings.Split(strings.TrimSpace(string(filesOutput)), "\n"))
		if strings.TrimSpace(string(filesOutput)) == "" {
			filesCount = 0
		}

		commits = append(commits, GitCommit{
			Hash:       parts[0],
			Author:     parts[1],
			Email:      parts[2],
			Date:       parts[3],
			Message:    parts[4],
			Branch:     branch,
			FilesCount: filesCount,
		})
	}

	return commits, nil
}

// normalizeAuthorToAssignee converts git author name/email to a potential assignee
func normalizeAuthorToAssignee(author, email string) string {
	// Try to extract username from email (common pattern: user@domain.com -> user)
	if email != "" {
		emailParts := strings.Split(email, "@")
		if len(emailParts) > 0 && emailParts[0] != "" {
			return emailParts[0]
		}
	}

	// If no valid email, use author name (convert to lowercase, replace spaces with dots)
	if author != "" {
		normalized := strings.ToLower(author)
		normalized = strings.ReplaceAll(normalized, " ", ".")
		return normalized
	}

	return "unknown"
}

// getCurrentGitBranch gets the current git branch name
func getCurrentGitBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// extractTaskIDFromBranch extracts task ID from branch name (e.g., feature/task-123 -> 123)
func extractTaskIDFromBranch(branch string) string {
	// Common branch patterns:
	// - feature/task-123
	// - fix/task-456  
	// - task/123
	// - 123-some-description
	patterns := []string{
		`(?:feature|fix|task|bugfix|hotfix)/task-([a-zA-Z0-9-]+)`,
		`(?:feature|fix|task|bugfix|hotfix)/([a-zA-Z0-9-]+)`,
		`^([a-zA-Z0-9]+)-`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(strings.ToLower(branch))
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// TaskAction represents an action to perform on a task based on commit message
type TaskAction struct {
	TaskID    string
	Action    string // "start", "progress", "complete", "close", "fix", "update"
	NewStatus *domain.TaskStatus
	Note      string
}

// extractTaskActionsFromCommit extracts task IDs and determines actions from commit messages
func extractTaskActionsFromCommit(message string) []TaskAction {
	// Intelligent keyword detection for status changes
	statusKeywords := map[string]domain.TaskStatus{
		// Completion keywords
		"complete":   domain.TaskStatusDone,
		"completed":  domain.TaskStatusDone,
		"finish":     domain.TaskStatusDone,
		"finished":   domain.TaskStatusDone,
		"done":       domain.TaskStatusDone,
		"close":      domain.TaskStatusDone,
		"closed":     domain.TaskStatusDone,
		"closes":     domain.TaskStatusDone,
		"resolve":    domain.TaskStatusDone,
		"resolved":   domain.TaskStatusDone,
		"resolves":   domain.TaskStatusDone,
		"implement":  domain.TaskStatusDone,
		"implemented": domain.TaskStatusDone,

		// Start/Progress keywords  
		"start":      domain.TaskStatusInProgress,
		"started":    domain.TaskStatusInProgress,
		"begin":      domain.TaskStatusInProgress,
		"working":    domain.TaskStatusInProgress,
		"wip":        domain.TaskStatusInProgress,
		"progress":   domain.TaskStatusInProgress,
		"update":     domain.TaskStatusInProgress,
		"updated":    domain.TaskStatusInProgress,
		"fix":        domain.TaskStatusInProgress,
		"fixing":     domain.TaskStatusInProgress,
		"improve":    domain.TaskStatusInProgress,
		"improving":  domain.TaskStatusInProgress,
		"refactor":   domain.TaskStatusInProgress,
		"refactoring": domain.TaskStatusInProgress,
	}

	// Enhanced patterns for task detection
	taskPatterns := []string{
		`#task-([a-zA-Z0-9-]+)`,
		`task-([a-zA-Z0-9-]+)`,
		`#([a-fA-F0-9]{8}[a-fA-F0-9]*)`, // Memory IDs (hex strings)
		`(?:ref|refs|references?)\s+#?([a-zA-Z0-9-]+)`,
	}

	var actions []TaskAction
	lowerMessage := strings.ToLower(message)

	// Extract all task IDs first
	var allTaskIDs []string
	for _, pattern := range taskPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(lowerMessage, -1)
		for _, match := range matches {
			if len(match) > 1 {
				allTaskIDs = append(allTaskIDs, match[1])
			}
		}
	}

	// Remove duplicate task IDs
	seen := make(map[string]bool)
	uniqueTaskIDs := make([]string, 0)
	for _, id := range allTaskIDs {
		if !seen[id] {
			seen[id] = true
			uniqueTaskIDs = append(uniqueTaskIDs, id)
		}
	}

	// Determine action for each task based on keywords
	for _, taskID := range uniqueTaskIDs {
		action := TaskAction{
			TaskID: taskID,
			Action: "progress", // default action
			Note:   message,
		}

		// Check for status-changing keywords in the message
		foundStatusKeyword := false
		for keyword, status := range statusKeywords {
			// Look for keyword near the task reference
			keywordPattern := fmt.Sprintf(`\b%s\b.*(?:task-?%s|#%s)\b|\b(?:task-?%s|#%s)\b.*\b%s\b`, 
				keyword, taskID, taskID, taskID, taskID, keyword)
			matched, _ := regexp.MatchString(keywordPattern, lowerMessage)
			
			if matched || strings.Contains(lowerMessage, keyword) {
				action.Action = keyword
				action.NewStatus = &status
				foundStatusKeyword = true
				break
			}
		}

		// If no specific status keyword found, infer from general context
		if !foundStatusKeyword {
			// Default to in_progress for any commit that references a task
			progressStatus := domain.TaskStatusInProgress
			action.NewStatus = &progressStatus
		}

		actions = append(actions, action)
	}

	return actions
}

// Legacy function for backward compatibility
func extractTaskIDsFromCommit(message string) []string {
	actions := extractTaskActionsFromCommit(message)
	var taskIDs []string
	for _, action := range actions {
		taskIDs = append(taskIDs, action.TaskID)
	}
	return taskIDs
}

// createProgressMemory creates a memory entry for git progress tracking
func createProgressMemory(services *ServiceContainer, projectID domain.ProjectID, action TaskAction, commit GitCommit) bool {
	ctx := context.Background()
	
	progressMemory := fmt.Sprintf("Git commit %s: %s", commit.Hash[:8], commit.Message)
	
	actionDesc := action.Action
	if action.NewStatus != nil {
		actionDesc = fmt.Sprintf("%s → %s", action.Action, string(*action.NewStatus))
	}
	
	contextInfo := fmt.Sprintf("Commit: %s, Author: %s (%s), Date: %s, Action: %s, Files: %d", 
		commit.Hash, commit.Author, commit.Email, commit.Date, actionDesc, commit.FilesCount)
	
	// Create a memory entry for this progress
	_, err := services.MemoryService.CreateMemory(ctx, ports.CreateMemoryRequest{
		ProjectID: projectID,
		Type:      domain.MemoryTypeSession, // Using session type for progress tracking
		Title:     fmt.Sprintf("Git Progress: Task %s (%s)", action.TaskID, actionDesc),
		Content:   progressMemory,
		Context:   contextInfo,
		Tags:      domain.Tags{"git-progress", "task-" + action.TaskID, "action-" + action.Action},
	})
	
	if err != nil {
		fmt.Printf("Warning: Failed to create progress memory for task %s: %v\n", action.TaskID, err)
		return false
	}
	
	return true
}

// isGitRepository checks if current directory is a git repository
func isGitRepository() bool {
	_, err := os.Stat(".git")
	return err == nil
}

func init() {
	// Add flags to git scan-commits command
	gitScanCommitsCmd.Flags().IntP("count", "c", 10, "Number of recent commits to scan")

	// Add subcommands to git integration command
	gitIntegrationCmd.AddCommand(gitScanCommitsCmd)
	gitIntegrationCmd.AddCommand(gitHookInstallCmd)

	// Add git integration command to root
	rootCmd.AddCommand(gitIntegrationCmd)
}