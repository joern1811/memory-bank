package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage development sessions",
	Long:  `Start, log progress, complete, and manage development sessions to track work context.`,
}

var sessionStartCmd = &cobra.Command{
	Use:   "start [task-description]",
	Short: "Start a new development session",
	Long: `Start a new development session with a task description.
The session will be created and set as active for the project.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskDescription := args[0]
		projectID, _ := cmd.Flags().GetString("project")

		if taskDescription == "" {
			return fmt.Errorf("task description is required")
		}

		// Get services
		services, err := GetServices()
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		ctx := context.Background()
		
		fmt.Printf("Starting new development session:\n")
		fmt.Printf("  Task: %s\n", taskDescription)
		if projectID != "" {
			fmt.Printf("  Project: %s\n", projectID)
		}

		// Create session request
		req := ports.StartSessionRequest{
			TaskDescription: taskDescription,
		}

		// Set project ID if provided
		if projectID != "" {
			pid := domain.ProjectID(projectID)
			req.ProjectID = pid
		} else {
			// Use default project if none specified
			req.ProjectID = domain.ProjectID("default")
		}

		// Start session
		session, err := services.SessionService.StartSession(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to start session: %w", err)
		}
		
		fmt.Printf("✓ Development session started successfully (ID: %s)\n", session.ID)
		fmt.Printf("  Status: %s\n", session.Status)
		fmt.Printf("  Started: %s\n", session.StartTime.Format("2006-01-02 15:04:05"))
		return nil
	},
}

var sessionLogCmd = &cobra.Command{
	Use:   "log [progress-entry]",
	Short: "Log progress to current session",
	Long: `Log progress to the currently active session for the project.
If no active session exists, you will be prompted to start one.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		progressEntry := args[0]
		projectID, _ := cmd.Flags().GetString("project")
		sessionID, _ := cmd.Flags().GetString("session")

		if progressEntry == "" {
			return fmt.Errorf("progress entry is required")
		}

		// Get services
		services, err := GetServices()
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		ctx := context.Background()
		
		var targetSessionID domain.SessionID

		// If session ID provided, use it directly
		if sessionID != "" {
			targetSessionID = domain.SessionID(sessionID)
		} else {
			// Find active session for project
			var pid domain.ProjectID
			if projectID != "" {
				pid = domain.ProjectID(projectID)
			} else {
				pid = domain.ProjectID("default")
			}

			activeSession, err := services.SessionService.GetActiveSession(ctx, pid)
			if err != nil {
				return fmt.Errorf("no active session found for project %s. Start a session first with 'session start'", pid)
			}
			targetSessionID = activeSession.ID
		}

		fmt.Printf("Logging progress to session %s:\n", targetSessionID)
		fmt.Printf("  Entry: %s\n", progressEntry)

		// Log progress
		err = services.SessionService.LogProgress(ctx, targetSessionID, progressEntry)
		if err != nil {
			return fmt.Errorf("failed to log progress: %w", err)
		}
		
		fmt.Printf("✓ Progress logged successfully\n")
		return nil
	},
}

var sessionCompleteCmd = &cobra.Command{
	Use:   "complete [outcome]",
	Short: "Complete current session",
	Long: `Complete the currently active session with an optional outcome description.
The session will be marked as completed and no longer active.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var outcome string
		if len(args) > 0 {
			outcome = args[0]
		}
		
		projectID, _ := cmd.Flags().GetString("project")
		sessionID, _ := cmd.Flags().GetString("session")

		// Get services
		services, err := GetServices()
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		ctx := context.Background()
		
		var targetSessionID domain.SessionID

		// If session ID provided, use it directly
		if sessionID != "" {
			targetSessionID = domain.SessionID(sessionID)
		} else {
			// Find active session for project
			var pid domain.ProjectID
			if projectID != "" {
				pid = domain.ProjectID(projectID)
			} else {
				pid = domain.ProjectID("default")
			}

			activeSession, err := services.SessionService.GetActiveSession(ctx, pid)
			if err != nil {
				return fmt.Errorf("no active session found for project %s", pid)
			}
			targetSessionID = activeSession.ID
		}

		fmt.Printf("Completing session %s\n", targetSessionID)
		if outcome != "" {
			fmt.Printf("  Outcome: %s\n", outcome)
		}

		// Complete session
		err = services.SessionService.CompleteSession(ctx, targetSessionID, outcome)
		if err != nil {
			return fmt.Errorf("failed to complete session: %w", err)
		}
		
		fmt.Printf("✓ Session completed successfully\n")
		return nil
	},
}

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sessions",
	Long:  `List development sessions, optionally filtered by project or status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, _ := cmd.Flags().GetString("project")
		status, _ := cmd.Flags().GetString("status")
		limit, _ := cmd.Flags().GetInt("limit")

		// Get services
		services, err := GetServices()
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		ctx := context.Background()
		
		fmt.Printf("Listing sessions")
		if projectID != "" {
			fmt.Printf(" for project: %s", projectID)
		}
		if status != "" {
			fmt.Printf(" with status: %s", status)
		}
		fmt.Printf(" (limit: %d)\n", limit)

		// Set project ID
		var pid domain.ProjectID
		if projectID != "" {
			pid = domain.ProjectID(projectID)
		} else {
			pid = domain.ProjectID("default")
		}

		// List sessions
		sessions, err := services.SessionService.ListSessions(ctx, pid)
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}
		
		// Filter by status if provided
		if status != "" {
			filtered := make([]*domain.Session, 0)
			for _, session := range sessions {
				if strings.EqualFold(string(session.Status), status) {
					filtered = append(filtered, session)
				}
			}
			sessions = filtered
		}

		// Apply limit
		if limit > 0 && len(sessions) > limit {
			sessions = sessions[:limit]
		}

		fmt.Printf("\nSessions (%d found):\n", len(sessions))
		if len(sessions) == 0 {
			if projectID == "" {
				fmt.Println("No sessions found. Try specifying a project with --project flag.")
			} else {
				fmt.Println("No sessions found for the specified filters.")
			}
		} else {
			for i, session := range sessions {
				fmt.Printf("\n%d. %s\n", i+1, session.TaskDescription)
				fmt.Printf("   ID: %s\n", session.ID)
				fmt.Printf("   Project: %s, Status: %s\n", session.ProjectID, session.Status)
				fmt.Printf("   Started: %s\n", session.StartTime.Format("2006-01-02 15:04:05"))
				
				if session.EndTime != nil {
					fmt.Printf("   Ended: %s\n", session.EndTime.Format("2006-01-02 15:04:05"))
					fmt.Printf("   Duration: %s\n", session.Duration().String())
				} else if session.IsActive() {
					fmt.Printf("   Duration: %s (ongoing)\n", time.Since(session.StartTime).Truncate(time.Second).String())
				}
				
				if session.Outcome != "" {
					fmt.Printf("   Outcome: %s\n", truncateString(session.Outcome, 100))
				}
				
				if len(session.ProgressLog) > 0 {
					fmt.Printf("   Progress entries: %d\n", len(session.ProgressLog))
					// Show last progress entry
					lastEntry := session.ProgressLog[len(session.ProgressLog)-1]
					fmt.Printf("   Latest: %s\n", truncateString(lastEntry, 80))
				}
			}
		}
		
		return nil
	},
}

var sessionGetCmd = &cobra.Command{
	Use:   "get [session-id]",
	Short: "Get session details",
	Long:  `Retrieve detailed information about a specific session including all progress entries.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionID := args[0]

		if sessionID == "" {
			return fmt.Errorf("session ID is required")
		}

		// Get services
		services, err := GetServices()
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		ctx := context.Background()
		
		fmt.Printf("Retrieving session: %s\n", sessionID)

		// Get session
		session, err := services.SessionService.GetSession(ctx, domain.SessionID(sessionID))
		if err != nil {
			return fmt.Errorf("failed to get session: %w", err)
		}
		
		fmt.Printf("\nSession Details:\n")
		fmt.Printf("  ID: %s\n", session.ID)
		fmt.Printf("  Project: %s\n", session.ProjectID)
		fmt.Printf("  Task: %s\n", session.TaskDescription)
		fmt.Printf("  Status: %s\n", session.Status)
		fmt.Printf("  Started: %s\n", session.StartTime.Format("2006-01-02 15:04:05"))
		
		if session.EndTime != nil {
			fmt.Printf("  Ended: %s\n", session.EndTime.Format("2006-01-02 15:04:05"))
			fmt.Printf("  Duration: %s\n", session.Duration().String())
		} else if session.IsActive() {
			fmt.Printf("  Duration: %s (ongoing)\n", time.Since(session.StartTime).Truncate(time.Second).String())
		}
		
		if session.Outcome != "" {
			fmt.Printf("  Outcome: %s\n", session.Outcome)
		}
		
		if len(session.ProgressLog) > 0 {
			fmt.Printf("\nProgress Log (%d entries):\n", len(session.ProgressLog))
			for i, entry := range session.ProgressLog {
				fmt.Printf("  %d. %s\n", i+1, entry)
			}
		} else {
			fmt.Printf("\nProgress Log: No entries yet\n")
		}
		
		return nil
	},
}

var sessionAbortCmd = &cobra.Command{
	Use:   "abort [session-id]",
	Short: "Abort session",
	Long: `Abort a session without completing it. The session will be marked as aborted.
If no session ID is provided, the active session for the project will be aborted.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var sessionID string
		if len(args) > 0 {
			sessionID = args[0]
		}
		
		projectID, _ := cmd.Flags().GetString("project")

		// Get services
		services, err := GetServices()
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		ctx := context.Background()
		
		var targetSessionID domain.SessionID

		// If session ID provided, use it directly
		if sessionID != "" {
			targetSessionID = domain.SessionID(sessionID)
		} else {
			// Find active session for project
			var pid domain.ProjectID
			if projectID != "" {
				pid = domain.ProjectID(projectID)
			} else {
				pid = domain.ProjectID("default")
			}

			activeSession, err := services.SessionService.GetActiveSession(ctx, pid)
			if err != nil {
				return fmt.Errorf("no active session found for project %s", pid)
			}
			targetSessionID = activeSession.ID
		}

		fmt.Printf("Aborting session %s\n", targetSessionID)

		// Abort session
		err = services.SessionService.AbortSession(ctx, targetSessionID)
		if err != nil {
			return fmt.Errorf("failed to abort session: %w", err)
		}
		
		fmt.Printf("✓ Session aborted successfully\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(sessionCmd)
	
	// Add subcommands
	sessionCmd.AddCommand(sessionStartCmd)
	sessionCmd.AddCommand(sessionLogCmd)
	sessionCmd.AddCommand(sessionCompleteCmd)
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionGetCmd)
	sessionCmd.AddCommand(sessionAbortCmd)
	
	// Flags for start command
	sessionStartCmd.Flags().StringP("project", "p", "", "project ID")
	
	// Flags for log command
	sessionLogCmd.Flags().StringP("project", "p", "", "project ID (if no session ID provided)")
	sessionLogCmd.Flags().StringP("session", "s", "", "specific session ID")
	
	// Flags for complete command
	sessionCompleteCmd.Flags().StringP("project", "p", "", "project ID (if no session ID provided)")
	sessionCompleteCmd.Flags().StringP("session", "s", "", "specific session ID")
	
	// Flags for list command
	sessionListCmd.Flags().StringP("project", "p", "", "filter by project ID")
	sessionListCmd.Flags().StringP("status", "", "", "filter by status (active, completed, aborted)")
	sessionListCmd.Flags().IntP("limit", "l", 20, "maximum number of results")
	
	// Flags for abort command
	sessionAbortCmd.Flags().StringP("project", "p", "", "project ID (if no session ID provided)")
}