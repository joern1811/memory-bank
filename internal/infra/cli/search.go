package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search across all memory entries",
	Long: `Search across all memory entries using semantic search.
This is a convenience command that searches all projects and memory types.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		limit, _ := cmd.Flags().GetInt("limit")
		threshold, _ := cmd.Flags().GetFloat32("threshold")
		showContent, _ := cmd.Flags().GetBool("content")

		// Get services
		services, err := GetServices()
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		ctx := context.Background()
		
		fmt.Printf("Global search for: %s\n", query)
		fmt.Printf("Limit: %d, Threshold: %.2f\n", limit, threshold)
		if showContent {
			fmt.Println("Including content in results")
		}

		// Create search request (no project filter for global search)
		searchReq := ports.SemanticSearchRequest{
			Query:     query,
			Limit:     limit,
			Threshold: threshold,
		}

		// Search memories
		results, err := services.MemoryService.SearchMemories(ctx, searchReq)
		if err != nil {
			return fmt.Errorf("failed to search memories: %w", err)
		}
		
		fmt.Printf("\nGlobal Search Results (%d found):\n", len(results))
		if len(results) == 0 {
			fmt.Println("No memories found matching your query.")
		} else {
			for i, result := range results {
				fmt.Printf("\n%d. %s (Score: %.3f)\n", i+1, result.Memory.Title, result.Similarity)
				fmt.Printf("   Type: %s, Project: %s\n", result.Memory.Type, result.Memory.ProjectID)
				
				if showContent {
					fmt.Printf("   Content: %s\n", result.Memory.Content)
				} else {
					fmt.Printf("   Content: %s\n", truncateString(result.Memory.Content, 100))
				}
				
				if len(result.Memory.Tags) > 0 {
					fmt.Printf("   Tags: %s\n", strings.Join(result.Memory.Tags, ", "))
				}
				fmt.Printf("   Created: %s\n", result.Memory.CreatedAt.Format("2006-01-02 15:04:05"))
			}
		}
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	
	searchCmd.Flags().IntP("limit", "l", 10, "maximum number of results")
	searchCmd.Flags().Float32P("threshold", "", 0.5, "similarity threshold")
	searchCmd.Flags().Bool("content", false, "show content in results")
}