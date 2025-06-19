package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/spf13/cobra"
)

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

var memoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "Manage memory entries",
	Long:  `Create, search, update, and delete memory entries in the knowledge base.`,
}

var memoryCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new memory entry",
	Long: `Create a new memory entry with specified type, title, and content.
Supported types: decision, pattern, error-solution, code, documentation`,
	RunE: func(cmd *cobra.Command, args []string) error {
		memoryType, _ := cmd.Flags().GetString("type")
		title, _ := cmd.Flags().GetString("title")
		content, _ := cmd.Flags().GetString("content")
		tagsStr, _ := cmd.Flags().GetString("tags")
		projectID, _ := cmd.Flags().GetString("project")

		if memoryType == "" || title == "" || content == "" {
			return fmt.Errorf("type, title, and content are required")
		}

		var tags []string
		if tagsStr != "" {
			tags = strings.Split(tagsStr, ",")
			for i, tag := range tags {
				tags[i] = strings.TrimSpace(tag)
			}
		}

		// Get services
		services, err := GetServicesForCLI(cmd)
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		ctx := context.Background()
		
		fmt.Printf("Creating memory entry:\n")
		fmt.Printf("  Type: %s\n", memoryType)
		fmt.Printf("  Title: %s\n", title)
		fmt.Printf("  Content: %s\n", content)
		if len(tags) > 0 {
			fmt.Printf("  Tags: %s\n", strings.Join(tags, ", "))
		}
		if projectID != "" {
			fmt.Printf("  Project: %s\n", projectID)
		}

		// Create memory request
		req := ports.CreateMemoryRequest{
			Type:    domain.MemoryType(memoryType),
			Title:   title,
			Content: content,
			Tags:    tags,
		}

		// Set project ID if provided
		if projectID != "" {
			pid := domain.ProjectID(projectID)
			req.ProjectID = pid
		} else {
			// Use default project if none specified
			req.ProjectID = domain.ProjectID("default")
		}

		// Create memory
		memory, err := services.MemoryService.CreateMemory(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to create memory: %w", err)
		}
		
		fmt.Printf("✓ Memory entry created successfully (ID: %s)\n", memory.ID)
		return nil
	},
}

var memorySearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search memory entries",
	Long:  `Search memory entries using semantic search based on content similarity.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		projectID, _ := cmd.Flags().GetString("project")
		limit, _ := cmd.Flags().GetInt("limit")
		threshold, _ := cmd.Flags().GetFloat32("threshold")

		// Get services
		services, err := GetServicesForCLI(cmd)
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		ctx := context.Background()
		
		fmt.Printf("Searching for: %s\n", query)
		if projectID != "" {
			fmt.Printf("Project filter: %s\n", projectID)
		}
		fmt.Printf("Limit: %d, Threshold: %.2f\n", limit, threshold)

		// Create search request
		searchReq := ports.SemanticSearchRequest{
			Query:     query,
			Limit:     limit,
			Threshold: threshold,
		}

		// Set project filter if provided
		if projectID != "" {
			pid := domain.ProjectID(projectID)
			searchReq.ProjectID = &pid
		}

		// Search memories
		results, err := services.MemoryService.SearchMemories(ctx, searchReq)
		if err != nil {
			return fmt.Errorf("failed to search memories: %w", err)
		}
		
		fmt.Printf("\nSearch Results (%d found):\n", len(results))
		if len(results) == 0 {
			fmt.Println("No memories found matching your query.")
		} else {
			for i, result := range results {
				fmt.Printf("\n%d. %s (Score: %.3f)\n", i+1, result.Memory.Title, result.Similarity)
				fmt.Printf("   Type: %s, Project: %s\n", result.Memory.Type, result.Memory.ProjectID)
				fmt.Printf("   Content: %s\n", truncateString(result.Memory.Content, 100))
				if len(result.Memory.Tags) > 0 {
					fmt.Printf("   Tags: %s\n", strings.Join(result.Memory.Tags, ", "))
				}
			}
		}
		
		return nil
	},
}

var memoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List memory entries",
	Long:  `List all memory entries, optionally filtered by project or type.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, _ := cmd.Flags().GetString("project")
		memoryType, _ := cmd.Flags().GetString("type")
		limit, _ := cmd.Flags().GetInt("limit")

		// Get services
		services, err := GetServicesForCLI(cmd)
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		ctx := context.Background()
		
		fmt.Printf("Listing memory entries")
		if projectID != "" {
			fmt.Printf(" for project: %s", projectID)
		}
		if memoryType != "" {
			fmt.Printf(" of type: %s", memoryType)
		}
		fmt.Printf(" (limit: %d)\n", limit)

		// Create list request
		listReq := ports.ListMemoriesRequest{
			Limit: limit,
		}

		// Set project filter if provided
		if projectID != "" {
			pid := domain.ProjectID(projectID)
			listReq.ProjectID = &pid
		}

		// Set type filter if provided
		if memoryType != "" {
			mtype := domain.MemoryType(memoryType)
			listReq.Type = &mtype
		}

		// List memories
		memories, err := services.MemoryService.ListMemories(ctx, listReq)
		if err != nil {
			return fmt.Errorf("failed to list memories: %w", err)
		}
		
		fmt.Printf("\nMemory Entries (%d found):\n", len(memories))
		if len(memories) == 0 {
			if projectID == "" {
				fmt.Println("No memories found. Try specifying a project with --project flag.")
			} else {
				fmt.Println("No memories found for the specified filters.")
			}
		} else {
			for i, memory := range memories {
				fmt.Printf("\n%d. %s\n", i+1, memory.Title)
				fmt.Printf("   ID: %s\n", memory.ID)
				fmt.Printf("   Type: %s, Project: %s\n", memory.Type, memory.ProjectID)
				fmt.Printf("   Content: %s\n", truncateString(memory.Content, 100))
				if len(memory.Tags) > 0 {
					fmt.Printf("   Tags: %s\n", strings.Join(memory.Tags, ", "))
				}
				fmt.Printf("   Created: %s\n", memory.CreatedAt.Format("2006-01-02 15:04:05"))
			}
		}
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(memoryCmd)
	
	// Add subcommands
	memoryCmd.AddCommand(memoryCreateCmd)
	memoryCmd.AddCommand(memorySearchCmd)
	memoryCmd.AddCommand(memoryListCmd)
	
	// Flags for create command
	memoryCreateCmd.Flags().StringP("type", "t", "", "memory type (decision, pattern, error-solution, code, documentation)")
	memoryCreateCmd.Flags().StringP("title", "", "", "memory title")
	memoryCreateCmd.Flags().StringP("content", "", "", "memory content")
	memoryCreateCmd.Flags().StringP("tags", "", "", "comma-separated tags")
	memoryCreateCmd.Flags().StringP("project", "p", "", "project ID")
	
	// Flags for search command
	memorySearchCmd.Flags().StringP("project", "p", "", "filter by project ID")
	memorySearchCmd.Flags().IntP("limit", "l", 10, "maximum number of results")
	memorySearchCmd.Flags().Float32P("threshold", "", 0.5, "similarity threshold")
	
	// Flags for list command
	memoryListCmd.Flags().StringP("project", "p", "", "filter by project ID")
	memoryListCmd.Flags().StringP("type", "t", "", "filter by memory type")
	memoryListCmd.Flags().IntP("limit", "l", 50, "maximum number of results")
}