package cli

import (
	"context"
	"fmt"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up and regenerate embeddings for all memories",
	Long: `Cleanup command performs the following operations:
1. Deletes the ChromaDB collection
2. Resets all embedding flags in SQLite database
3. Regenerates embeddings for all existing memories

This is useful when the embedding system was broken or collection configuration changed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		// Initialize services
		services, err := GetServicesForCLI(cmd)
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		projectID, _ := cmd.Flags().GetString("project")

		fmt.Println("Starting Memory Bank cleanup...")

		if dryRun {
			fmt.Println("DRY RUN MODE - No changes will be made")
		}

		if !dryRun {
			fmt.Println("🚀 Starting embedding cleanup...")
			
			// Perform cleanup via service layer (hexagonal architecture)
			result, err := services.MemoryService.CleanupEmbeddings(ctx, domain.ProjectID(projectID))
			if err != nil {
				return fmt.Errorf("cleanup failed: %w", err)
			}

			// Display results
			fmt.Printf("\n📊 Cleanup Results:\n")
			fmt.Printf("  - Total memories: %d\n", result.TotalMemories)
			fmt.Printf("  - Memories processed: %d\n", result.MemoriesProcessed)
			fmt.Printf("  - Embeddings generated: %d\n", result.EmbeddingsGenerated)
			fmt.Printf("  - Errors: %d\n", result.Errors)

			if len(result.ErrorMessages) > 0 {
				fmt.Printf("\n❌ Errors encountered:\n")
				for _, errMsg := range result.ErrorMessages {
					fmt.Printf("  - %s\n", errMsg)
				}
			}

			if result.Errors > 0 {
				return fmt.Errorf("cleanup completed with %d errors", result.Errors)
			}

			fmt.Println("\n🎉 Cleanup completed successfully!")
		} else {
			// Dry run - just show what would be done
			memoriesReq := ports.ListMemoriesRequest{
				ProjectID: func() *domain.ProjectID { pid := domain.ProjectID(projectID); return &pid }(),
				Limit:     1000,
			}
			memories, err := services.MemoryService.ListMemories(ctx, memoriesReq)
			if err != nil {
				return fmt.Errorf("failed to list memories: %w", err)
			}

			memoriesWithoutEmbeddings := 0
			memoriesWithEmbeddings := 0

			for _, memory := range memories {
				if memory.HasEmbedding {
					memoriesWithEmbeddings++
				} else {
					memoriesWithoutEmbeddings++
				}
			}

			fmt.Printf("Would process %d memories:\n", len(memories))
			fmt.Printf("  - %d with embeddings (will regenerate)\n", memoriesWithEmbeddings)
			fmt.Printf("  - %d without embeddings (will generate)\n", memoriesWithoutEmbeddings)
		}

		return nil
	},
}

func init() {
	cleanupCmd.Flags().Bool("dry-run", false, "show what would be done without making changes")
	cleanupCmd.Flags().StringP("project", "p", "default", "project ID to clean up (default: default)")
	
	rootCmd.AddCommand(cleanupCmd)
}