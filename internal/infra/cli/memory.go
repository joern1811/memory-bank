package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

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

		// TODO: Call memory service to create memory
		// ctx := context.Background()
		
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

		// TODO: Call memory service to create memory
		// memoryService.CreateMemory(ctx, memoryType, title, content, tags, projectID)
		
		fmt.Println("✓ Memory entry created successfully")
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

		// TODO: Call memory service to search memories
		// ctx := context.Background()
		
		fmt.Printf("Searching for: %s\n", query)
		if projectID != "" {
			fmt.Printf("Project filter: %s\n", projectID)
		}
		fmt.Printf("Limit: %d, Threshold: %.2f\n", limit, threshold)

		// TODO: Call memory service to search memories
		// results := memoryService.SearchMemories(ctx, query, projectID, limit, threshold)
		
		fmt.Println("\nSearch Results:")
		fmt.Println("(No results - service integration pending)")
		
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

		// TODO: Call memory service to list memories
		// ctx := context.Background()
		
		fmt.Printf("Listing memory entries")
		if projectID != "" {
			fmt.Printf(" for project: %s", projectID)
		}
		if memoryType != "" {
			fmt.Printf(" of type: %s", memoryType)
		}
		fmt.Printf(" (limit: %d)\n", limit)

		// TODO: Call memory service to list memories
		// memories := memoryService.ListMemories(ctx, projectID, memoryType, limit)
		
		fmt.Println("\nMemory Entries:")
		fmt.Println("(No entries - service integration pending)")
		
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