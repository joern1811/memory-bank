package cli

import (
	"fmt"

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

		// TODO: Call memory service to search across all memories
		// ctx := context.Background()
		
		fmt.Printf("Global search for: %s\n", query)
		fmt.Printf("Limit: %d, Threshold: %.2f\n", limit, threshold)
		if showContent {
			fmt.Println("Including content in results")
		}

		// TODO: Call memory service to search across all memories
		// results := memoryService.SearchAllMemories(ctx, query, limit, threshold)
		
		fmt.Println("\nSearch Results:")
		fmt.Println("(No results - service integration pending)")
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	
	searchCmd.Flags().IntP("limit", "l", 10, "maximum number of results")
	searchCmd.Flags().Float32P("threshold", "", 0.5, "similarity threshold")
	searchCmd.Flags().Bool("content", false, "show content in results")
}