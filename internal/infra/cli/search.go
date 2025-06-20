package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/joern1811/memory-bank/internal/domain"
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
		services, err := GetServicesForCLI(cmd)
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		ctx := context.Background()

		// Quick health check for external services
		QuickHealthCheck(ctx, services)

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

// Advanced search commands

var facetedSearchCmd = &cobra.Command{
	Use:   "faceted [query]",
	Short: "Advanced search with facets and filters",
	Long: `Perform advanced search with faceted results and comprehensive filtering.
Supports filtering by types, tags, sessions, content length, and more.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		limit, _ := cmd.Flags().GetInt("limit")
		threshold, _ := cmd.Flags().GetFloat32("threshold")
		projectFlag, _ := cmd.Flags().GetString("project")
		typesFlag, _ := cmd.Flags().GetStringSlice("types")
		tagsFlag, _ := cmd.Flags().GetStringSlice("tags")
		includeFacets, _ := cmd.Flags().GetBool("facets")
		sortBy, _ := cmd.Flags().GetString("sort")
		sortDir, _ := cmd.Flags().GetString("sort-dir")
		minLength, _ := cmd.Flags().GetInt("min-length")
		maxLength, _ := cmd.Flags().GetInt("max-length")
		hasContent, _ := cmd.Flags().GetBool("has-content")

		// Get services
		services, err := GetServicesForCLI(cmd)
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		ctx := context.Background()

		fmt.Printf("Faceted search for: %s\n", query)

		// Build search request
		searchReq := ports.FacetedSearchRequest{
			Query:         query,
			Limit:         limit,
			Threshold:     threshold,
			IncludeFacets: includeFacets,
		}

		// Set project filter
		if projectFlag != "" {
			projectID := domain.ProjectID(projectFlag)
			searchReq.ProjectID = &projectID
		}

		// Build filters
		if len(typesFlag) > 0 || len(tagsFlag) > 0 || minLength > 0 || maxLength > 0 || hasContent {
			filters := &ports.SearchFilters{}

			// Convert types
			if len(typesFlag) > 0 {
				for _, t := range typesFlag {
					filters.Types = append(filters.Types, domain.MemoryType(t))
				}
			}

			// Set tags
			if len(tagsFlag) > 0 {
				filters.Tags = domain.Tags(tagsFlag)
			}

			// Set content filters
			if minLength > 0 {
				filters.MinLength = &minLength
			}
			if maxLength > 0 {
				filters.MaxLength = &maxLength
			}
			if hasContent {
				filters.HasContent = hasContent
			}

			searchReq.Filters = filters
		}

		// Set sort options
		if sortBy != "" {
			sortOption := &ports.SortOption{
				Field:     ports.SortField(sortBy),
				Direction: ports.SortDirection(sortDir),
			}
			searchReq.SortBy = sortOption
		}

		// Perform faceted search
		response, err := services.MemoryService.FacetedSearch(ctx, searchReq)
		if err != nil {
			return fmt.Errorf("failed to perform faceted search: %w", err)
		}

		// Display results
		fmt.Printf("\nFaceted Search Results (%d found):\n", response.Total)
		if len(response.Results) == 0 {
			fmt.Println("No memories found matching your query and filters.")
		} else {
			for i, result := range response.Results {
				fmt.Printf("\n%d. %s (Score: %.3f)\n", i+1, result.Memory.Title, result.Similarity)
				fmt.Printf("   Type: %s, Project: %s\n", result.Memory.Type, result.Memory.ProjectID)
				fmt.Printf("   Content: %s\n", truncateString(result.Memory.Content, 100))

				if len(result.Memory.Tags) > 0 {
					fmt.Printf("   Tags: %s\n", strings.Join(result.Memory.Tags, ", "))
				}
				fmt.Printf("   Created: %s\n", result.Memory.CreatedAt.Format("2006-01-02 15:04:05"))
			}
		}

		// Display facets if requested
		if includeFacets && response.Facets != nil {
			fmt.Printf("\nFacets:\n")

			if len(response.Facets.Types) > 0 {
				fmt.Printf("\nTypes:\n")
				for _, facet := range response.Facets.Types {
					fmt.Printf("  %s (%d)\n", facet.Type, facet.Count)
				}
			}

			if len(response.Facets.Tags) > 0 {
				fmt.Printf("\nTags:\n")
				for _, facet := range response.Facets.Tags {
					fmt.Printf("  %s (%d)\n", facet.Tag, facet.Count)
				}
			}

			if len(response.Facets.TimePeriods) > 0 {
				fmt.Printf("\nTime Periods:\n")
				for _, facet := range response.Facets.TimePeriods {
					fmt.Printf("  %s (%d)\n", facet.Period, facet.Count)
				}
			}
		}

		return nil
	},
}

var enhancedSearchCmd = &cobra.Command{
	Use:   "enhanced [query]",
	Short: "Enhanced search with relevance scoring and highlights",
	Long: `Perform enhanced search with improved relevance scoring, match reasons, and content highlights.
Provides detailed insights into why each result was matched.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		limit, _ := cmd.Flags().GetInt("limit")
		threshold, _ := cmd.Flags().GetFloat32("threshold")
		projectFlag, _ := cmd.Flags().GetString("project")
		typeFlag, _ := cmd.Flags().GetString("type")
		tagsFlag, _ := cmd.Flags().GetStringSlice("tags")

		// Get services
		services, err := GetServicesForCLI(cmd)
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		ctx := context.Background()

		fmt.Printf("Enhanced search for: %s\n", query)

		// Build search request
		searchReq := ports.SemanticSearchRequest{
			Query:     query,
			Limit:     limit,
			Threshold: threshold,
		}

		// Set project filter
		if projectFlag != "" {
			projectID := domain.ProjectID(projectFlag)
			searchReq.ProjectID = &projectID
		}

		// Set type filter
		if typeFlag != "" {
			memoryType := domain.MemoryType(typeFlag)
			searchReq.Type = &memoryType
		}

		// Set tags filter
		if len(tagsFlag) > 0 {
			searchReq.Tags = domain.Tags(tagsFlag)
		}

		// Perform enhanced search
		results, err := services.MemoryService.SearchWithRelevanceScoring(ctx, searchReq)
		if err != nil {
			return fmt.Errorf("failed to perform enhanced search: %w", err)
		}

		// Display results
		fmt.Printf("\nEnhanced Search Results (%d found):\n", len(results))
		if len(results) == 0 {
			fmt.Println("No memories found matching your query.")
		} else {
			for i, result := range results {
				fmt.Printf("\n%d. %s\n", i+1, result.Memory.Title)
				fmt.Printf("   Similarity: %.3f | Relevance: %.3f\n", result.Similarity, result.RelevanceScore)
				fmt.Printf("   Type: %s | Project: %s\n", result.Memory.Type, result.Memory.ProjectID)

				if len(result.Memory.Tags) > 0 {
					fmt.Printf("   Tags: %s\n", strings.Join(result.Memory.Tags, ", "))
				}

				// Show match reasons
				if len(result.MatchReasons) > 0 {
					fmt.Printf("   Match Reasons:\n")
					for _, reason := range result.MatchReasons {
						fmt.Printf("     • %s\n", reason)
					}
				}

				// Show highlights
				if len(result.Highlights) > 0 {
					fmt.Printf("   Highlights:\n")
					for _, highlight := range result.Highlights {
						fmt.Printf("     %s\n", highlight)
					}
				}

				fmt.Printf("   Created: %s\n", result.Memory.CreatedAt.Format("2006-01-02 15:04:05"))
			}
		}

		return nil
	},
}

var suggestionsCmd = &cobra.Command{
	Use:   "suggestions [partial-query]",
	Short: "Get search suggestions based on existing content",
	Long: `Get intelligent search suggestions based on your existing memory content.
Suggestions are generated from titles, tags, and frequently used terms.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		partialQuery := args[0]
		projectFlag, _ := cmd.Flags().GetString("project")
		limit, _ := cmd.Flags().GetInt("limit")

		// Get services
		services, err := GetServicesForCLI(cmd)
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		ctx := context.Background()

		// Set project filter
		var projectID *domain.ProjectID
		if projectFlag != "" {
			pid := domain.ProjectID(projectFlag)
			projectID = &pid
		}

		// Get suggestions
		suggestions, err := services.MemoryService.GetSearchSuggestions(ctx, partialQuery, projectID)
		if err != nil {
			return fmt.Errorf("failed to get search suggestions: %w", err)
		}

		// Apply limit
		if limit > 0 && len(suggestions) > limit {
			suggestions = suggestions[:limit]
		}

		// Display suggestions
		fmt.Printf("Search suggestions for '%s':\n", partialQuery)
		if len(suggestions) == 0 {
			fmt.Println("No suggestions found.")
		} else {
			for i, suggestion := range suggestions {
				fmt.Printf("%d. %s\n", i+1, suggestion)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// Basic search flags
	searchCmd.Flags().IntP("limit", "l", 10, "maximum number of results")
	searchCmd.Flags().Float32P("threshold", "", 0.5, "similarity threshold")
	searchCmd.Flags().Bool("content", false, "show content in results")

	// Add advanced search commands
	searchCmd.AddCommand(facetedSearchCmd)
	searchCmd.AddCommand(enhancedSearchCmd)
	searchCmd.AddCommand(suggestionsCmd)

	// Faceted search flags
	facetedSearchCmd.Flags().IntP("limit", "l", 10, "maximum number of results")
	facetedSearchCmd.Flags().Float32P("threshold", "", 0.5, "similarity threshold")
	facetedSearchCmd.Flags().StringP("project", "p", "", "filter by project ID")
	facetedSearchCmd.Flags().StringSlice("types", []string{}, "filter by memory types")
	facetedSearchCmd.Flags().StringSlice("tags", []string{}, "filter by tags")
	facetedSearchCmd.Flags().Bool("facets", false, "include faceted results")
	facetedSearchCmd.Flags().String("sort", "relevance", "sort by: relevance, created_at, updated_at, title, type")
	facetedSearchCmd.Flags().String("sort-dir", "desc", "sort direction: asc, desc")
	facetedSearchCmd.Flags().Int("min-length", 0, "minimum content length")
	facetedSearchCmd.Flags().Int("max-length", 0, "maximum content length")
	facetedSearchCmd.Flags().Bool("has-content", false, "filter memories with content")

	// Enhanced search flags
	enhancedSearchCmd.Flags().IntP("limit", "l", 10, "maximum number of results")
	enhancedSearchCmd.Flags().Float32P("threshold", "", 0.5, "similarity threshold")
	enhancedSearchCmd.Flags().StringP("project", "p", "", "filter by project ID")
	enhancedSearchCmd.Flags().StringP("type", "t", "", "filter by memory type")
	enhancedSearchCmd.Flags().StringSlice("tags", []string{}, "filter by tags")

	// Suggestions flags
	suggestionsCmd.Flags().StringP("project", "p", "", "filter by project ID")
	suggestionsCmd.Flags().IntP("limit", "l", 10, "maximum number of suggestions")
}
