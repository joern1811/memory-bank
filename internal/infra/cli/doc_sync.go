package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/spf13/cobra"
)

// DocAnalysisResult represents the result of documentation analysis
type DocAnalysisResult struct {
	ChangeTypes  []string
	AffectedDocs []string
	Suggestions  []string
}

// DocSuggestionsResult represents documentation update suggestions
type DocSuggestionsResult struct {
	GeneralSuggestions []string
	SpecificFiles      []string
	Checklist          []string
}

// DocValidationResult represents documentation validation results
type DocValidationResult struct {
	ValidItems  []string
	Issues      []string
	Suggestions []string
}

// classifyFileChanges analyzes file paths and determines change types
func classifyFileChanges(files []string) []string {
	changeTypes := make(map[string]bool)

	for _, file := range files {
		switch {
		// API related changes
		case strings.Contains(file, "api") || strings.Contains(file, "endpoint") ||
			strings.Contains(file, "route") || strings.Contains(file, "controller") ||
			strings.Contains(file, "handler"):
			changeTypes["api"] = true

		// CLI related changes
		case strings.Contains(file, "cli") || strings.Contains(file, "cmd") ||
			strings.Contains(file, "command") || strings.HasSuffix(file, "main.go") ||
			strings.HasSuffix(file, "main.py") || strings.HasSuffix(file, "main.js"):
			changeTypes["cli"] = true

		// Configuration changes
		case strings.Contains(file, "config") || strings.HasSuffix(file, ".json") ||
			strings.HasSuffix(file, ".yaml") || strings.HasSuffix(file, ".yml") ||
			strings.HasSuffix(file, ".toml") || strings.Contains(file, ".env") ||
			strings.HasPrefix(filepath.Base(file), "Dockerfile") ||
			strings.Contains(file, "docker-compose"):
			changeTypes["config"] = true

		// Database/Schema changes
		case strings.Contains(file, "migration") || strings.Contains(file, "schema") ||
			strings.Contains(file, "model") || strings.Contains(file, "entity"):
			changeTypes["database"] = true

		// Build/Deploy changes
		case strings.HasPrefix(filepath.Base(file), "Makefile") ||
			strings.Contains(file, "build") || strings.Contains(file, "deploy") ||
			strings.HasSuffix(file, ".sh") || strings.HasSuffix(file, ".ps1") ||
			strings.Contains(file, "ci") || strings.Contains(file, "cd") ||
			strings.Contains(file, ".github/"):
			changeTypes["build"] = true

		// Default: code changes
		case strings.HasSuffix(file, ".go") || strings.HasSuffix(file, ".py") ||
			strings.HasSuffix(file, ".js") || strings.HasSuffix(file, ".ts") ||
			strings.HasSuffix(file, ".java") || strings.HasSuffix(file, ".cpp") ||
			strings.HasSuffix(file, ".rs") || strings.HasSuffix(file, ".rb"):
			changeTypes["code"] = true
		}
	}

	// Convert map to slice
	var result []string
	for changeType := range changeTypes {
		result = append(result, changeType)
	}

	return result
}

// analyzeCodeChanges performs documentation impact analysis
func analyzeCodeChanges(ctx context.Context, memoryService ports.MemoryService, files []string) (*DocAnalysisResult, error) {
	changeTypes := classifyFileChanges(files)

	result := &DocAnalysisResult{
		ChangeTypes: changeTypes,
	}

	// Search for existing mappings
	for _, changeType := range changeTypes {
		searchReq := ports.SemanticSearchRequest{
			Query:     fmt.Sprintf("documentation mapping %s", changeType),
			Limit:     5,
			Threshold: 0.3,
		}

		memories, err := memoryService.SearchMemories(ctx, searchReq)
		if err == nil {
			for _, memory := range memories {
				if strings.Contains(memory.Memory.Content, "Documentation files:") {
					lines := strings.Split(memory.Memory.Content, "\n")
					for _, line := range lines {
						if strings.Contains(line, "Documentation files:") {
							docs := strings.TrimPrefix(line, "Documentation files:")
							docs = strings.TrimSpace(docs)
							if docs != "" {
								docFiles := strings.Split(docs, ",")
								for _, doc := range docFiles {
									doc = strings.TrimSpace(doc)
									if doc != "" {
										result.AffectedDocs = append(result.AffectedDocs, doc)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Generate default suggestions based on change types
	for _, changeType := range changeTypes {
		switch changeType {
		case "api":
			result.AffectedDocs = append(result.AffectedDocs, "README.md", "docs/api/", "*.openapi.yaml")
			result.Suggestions = append(result.Suggestions,
				"Update API documentation",
				"Verify endpoint examples",
				"Update OpenAPI/Swagger specs")
		case "cli":
			result.AffectedDocs = append(result.AffectedDocs, "README.md", "docs/cli/", "docs/usage/")
			result.Suggestions = append(result.Suggestions,
				"Update CLI usage examples",
				"Verify command documentation",
				"Update help text examples")
		case "config":
			result.AffectedDocs = append(result.AffectedDocs, "docs/configuration.md", "README.md")
			result.Suggestions = append(result.Suggestions,
				"Update configuration documentation",
				"Verify environment variables",
				"Update setup/installation guides")
		}
	}

	return result, nil
}

// generateUpdateSuggestions creates specific documentation update suggestions
func generateUpdateSuggestions(ctx context.Context, memoryService ports.MemoryService, changeType, component string) (*DocSuggestionsResult, error) {
	result := &DocSuggestionsResult{}

	// Search for similar patterns
	searchQuery := fmt.Sprintf("%s documentation", changeType)
	if component != "" {
		searchQuery = fmt.Sprintf("%s %s documentation", changeType, component)
	}

	searchReq := ports.SemanticSearchRequest{
		Query:     searchQuery,
		Limit:     10,
		Threshold: 0.3,
	}

	memories, err := memoryService.SearchMemories(ctx, searchReq)
	if err == nil {
		for _, memory := range memories {
			if strings.Contains(memory.Memory.Content, "checklist") ||
				strings.Contains(memory.Memory.Content, "update") {
				lines := strings.Split(memory.Memory.Content, "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
						item := strings.TrimPrefix(line, "- ")
						item = strings.TrimPrefix(item, "* ")
						result.Checklist = append(result.Checklist, item)
					}
				}
			}
		}
	}

	// Generate default suggestions
	switch changeType {
	case "api":
		result.GeneralSuggestions = append(result.GeneralSuggestions,
			"Update API reference documentation",
			"Verify all endpoint examples are current",
			"Test documented API calls")
		result.SpecificFiles = append(result.SpecificFiles,
			"docs/api/endpoints.md",
			"README.md#api-reference",
			"openapi.yaml")
	case "cli":
		result.GeneralSuggestions = append(result.GeneralSuggestions,
			"Update command usage examples",
			"Verify help text output",
			"Update installation instructions")
		result.SpecificFiles = append(result.SpecificFiles,
			"docs/cli/commands.md",
			"README.md#usage")
	}

	return result, nil
}

// validateDocumentationConsistency checks documentation consistency
func validateDocumentationConsistency(ctx context.Context, projectPath string, focusAreas []string) (*DocValidationResult, error) {
	result := &DocValidationResult{}

	// Check common documentation files
	commonDocs := []string{
		"README.md",
		"docs/",
		"CHANGELOG.md",
		"CONTRIBUTING.md",
	}

	for _, doc := range commonDocs {
		fullPath := filepath.Join(projectPath, doc)
		if _, err := os.Stat(fullPath); err == nil {
			result.ValidItems = append(result.ValidItems, doc)
		} else {
			result.Issues = append(result.Issues, fmt.Sprintf("Missing: %s", doc))
		}
	}

	// Focus area specific checks
	for _, area := range focusAreas {
		switch area {
		case "api":
			apiDocs := []string{"docs/api/", "openapi.yaml", "swagger.json"}
			for _, doc := range apiDocs {
				fullPath := filepath.Join(projectPath, doc)
				if _, err := os.Stat(fullPath); err != nil {
					result.Issues = append(result.Issues, fmt.Sprintf("Missing API documentation: %s", doc))
				}
			}
		case "cli":
			cliDocs := []string{"docs/cli/", "docs/commands.md"}
			for _, doc := range cliDocs {
				fullPath := filepath.Join(projectPath, doc)
				if _, err := os.Stat(fullPath); err != nil {
					result.Issues = append(result.Issues, fmt.Sprintf("Missing CLI documentation: %s", doc))
				}
			}
		}
	}

	// Generate suggestions
	if len(result.Issues) > 0 {
		result.Suggestions = append(result.Suggestions,
			"Create missing documentation files",
			"Set up documentation templates",
			"Add documentation to CI/CD pipeline")
	}

	return result, nil
}

var docSyncCmd = &cobra.Command{
	Use:   "doc",
	Short: "Documentation synchronization tools",
	Long:  `Tools for analyzing code changes and managing documentation synchronization.`,
}

var docAnalyzeCmd = &cobra.Command{
	Use:   "analyze-changes [files...]",
	Short: "Analyze code changes for documentation impact",
	Long: `Analyze specified files or changed files to determine what documentation 
might need to be updated. Provides intelligent suggestions based on code patterns.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get files to analyze
		files := args
		if len(files) == 0 {
			// No files specified, get help
			return cmd.Help()
		}

		// Get services
		services, err := GetServicesForCLI(cmd)
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		ctx := context.Background()

		// Analyze changes
		analysis, err := analyzeCodeChanges(ctx, services.MemoryService, files)
		if err != nil {
			return fmt.Errorf("failed to analyze changes: %w", err)
		}

		// Display results
		fmt.Printf("Documentation Impact Analysis\n")
		fmt.Printf("=============================\n\n")

		fmt.Printf("Analyzed files: %d\n", len(files))
		for _, file := range files {
			fmt.Printf("  - %s\n", file)
		}

		if len(analysis.ChangeTypes) > 0 {
			fmt.Printf("\nDetected change types:\n")
			for _, changeType := range analysis.ChangeTypes {
				fmt.Printf("  - %s\n", changeType)
			}
		}

		if len(analysis.AffectedDocs) > 0 {
			fmt.Printf("\nPotentially affected documentation:\n")
			for _, doc := range analysis.AffectedDocs {
				fmt.Printf("  - %s\n", doc)
			}
		}

		if len(analysis.Suggestions) > 0 {
			fmt.Printf("\nSuggestions:\n")
			for _, suggestion := range analysis.Suggestions {
				fmt.Printf("  • %s\n", suggestion)
			}
		}

		return nil
	},
}

var docSuggestCmd = &cobra.Command{
	Use:   "suggest-updates",
	Short: "Get intelligent documentation update suggestions",
	Long: `Get specific suggestions for documentation updates based on change type,
component, and historical patterns stored in Memory Bank.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		changeType, _ := cmd.Flags().GetString("change-type")
		component, _ := cmd.Flags().GetString("component")

		// Get services
		services, err := GetServicesForCLI(cmd)
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		ctx := context.Background()

		// Get suggestions
		suggestions, err := generateUpdateSuggestions(ctx, services.MemoryService, changeType, component)
		if err != nil {
			return fmt.Errorf("failed to generate suggestions: %w", err)
		}

		// Display results
		fmt.Printf("Documentation Update Suggestions\n")
		fmt.Printf("================================\n\n")

		if changeType != "" {
			fmt.Printf("Change type: %s\n", changeType)
		}
		if component != "" {
			fmt.Printf("Component: %s\n", component)
		}
		fmt.Println()

		if len(suggestions.GeneralSuggestions) > 0 {
			fmt.Printf("General recommendations:\n")
			for _, suggestion := range suggestions.GeneralSuggestions {
				fmt.Printf("  • %s\n", suggestion)
			}
			fmt.Println()
		}

		if len(suggestions.SpecificFiles) > 0 {
			fmt.Printf("Specific files to update:\n")
			for _, file := range suggestions.SpecificFiles {
				fmt.Printf("  - %s\n", file)
			}
			fmt.Println()
		}

		if len(suggestions.Checklist) > 0 {
			fmt.Printf("Documentation checklist:\n")
			for _, item := range suggestions.Checklist {
				fmt.Printf("  ☐ %s\n", item)
			}
		}

		return nil
	},
}

var docCreateMappingCmd = &cobra.Command{
	Use:   "create-mapping",
	Short: "Create code-to-documentation mapping",
	Long: `Create a mapping between code patterns and documentation files.
This helps the system understand what documentation needs updating when code changes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		codePattern, _ := cmd.Flags().GetString("code-pattern")
		docFiles, _ := cmd.Flags().GetStringSlice("documentation-files")
		changeType, _ := cmd.Flags().GetString("change-type")
		priority, _ := cmd.Flags().GetString("priority")
		projectID, _ := cmd.Flags().GetString("project")

		if codePattern == "" || len(docFiles) == 0 {
			return fmt.Errorf("code-pattern and documentation-files are required")
		}

		// Get services
		services, err := GetServicesForCLI(cmd)
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}

		ctx := context.Background()

		// Create mapping content
		content := fmt.Sprintf("Code pattern: %s\nDocumentation files: %s\nChange type: %s\nPriority: %s",
			codePattern, strings.Join(docFiles, ", "), changeType, priority)

		// Create memory entry
		var pid domain.ProjectID
		if projectID != "" {
			pid = domain.ProjectID(projectID)
		} else {
			// Try to get current project
			cwd, _ := os.Getwd()
			project, err := services.ProjectService.GetProjectByPath(ctx, cwd)
			if err == nil && project != nil {
				pid = project.ID
			} else {
				return fmt.Errorf("project ID required (use --project flag or run from initialized project)")
			}
		}

		req := ports.CreateMemoryRequest{
			ProjectID: pid,
			Type:      domain.MemoryTypeDocumentation,
			Title:     fmt.Sprintf("Documentation mapping: %s", codePattern),
			Content:   content,
			Tags:      domain.Tags{"mapping", changeType, priority},
		}

		memory, err := services.MemoryService.CreateMemory(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to create mapping: %w", err)
		}

		fmt.Printf("Created documentation mapping:\n")
		fmt.Printf("  ID: %s\n", memory.ID)
		fmt.Printf("  Pattern: %s\n", codePattern)
		fmt.Printf("  Documentation: %s\n", strings.Join(docFiles, ", "))
		fmt.Printf("  Change type: %s\n", changeType)
		fmt.Printf("  Priority: %s\n", priority)

		return nil
	},
}

var docValidateCmd = &cobra.Command{
	Use:   "validate-consistency",
	Short: "Validate documentation consistency",
	Long: `Validate that documentation is consistent with the current codebase.
Checks for outdated documentation and missing documentation for new features.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := cmd.Flags().GetString("project-path")
		focusAreas, _ := cmd.Flags().GetStringSlice("focus-areas")

		if projectPath == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			projectPath = cwd
		}

		ctx := context.Background()

		// Validate consistency
		result, err := validateDocumentationConsistency(ctx, projectPath, focusAreas)
		if err != nil {
			return fmt.Errorf("failed to validate consistency: %w", err)
		}

		// Display results
		fmt.Printf("Documentation Consistency Report\n")
		fmt.Printf("===============================\n\n")

		fmt.Printf("Project: %s\n", projectPath)
		if len(focusAreas) > 0 {
			fmt.Printf("Focus areas: %s\n", strings.Join(focusAreas, ", "))
		}
		fmt.Println()

		if len(result.ValidItems) > 0 {
			fmt.Printf("✅ Valid documentation:\n")
			for _, item := range result.ValidItems {
				fmt.Printf("  - %s\n", item)
			}
			fmt.Println()
		}

		if len(result.Issues) > 0 {
			fmt.Printf("⚠️  Issues found:\n")
			for _, issue := range result.Issues {
				fmt.Printf("  - %s\n", issue)
			}
			fmt.Println()
		}

		if len(result.Suggestions) > 0 {
			fmt.Printf("💡 Suggestions:\n")
			for _, suggestion := range result.Suggestions {
				fmt.Printf("  • %s\n", suggestion)
			}
		}

		return nil
	},
}

var docSetupCmd = &cobra.Command{
	Use:   "setup-automation",
	Short: "Set up documentation sync automation",
	Long: `Set up git hooks and automation scripts for documentation synchronization.
Generates and optionally installs pre-commit hooks and configuration files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := cmd.Flags().GetString("project-path")
		interactive, _ := cmd.Flags().GetBool("interactive")
		installHooks, _ := cmd.Flags().GetBool("install-hooks")

		if projectPath == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			projectPath = cwd
		}

		fmt.Printf("Setting up documentation sync automation\n")
		fmt.Printf("=======================================\n\n")

		fmt.Printf("Project path: %s\n", projectPath)
		fmt.Printf("Interactive mode: %t\n", interactive)
		fmt.Printf("Install hooks: %t\n", installHooks)
		fmt.Println()

		// Check if git repository
		gitDir := filepath.Join(projectPath, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			return fmt.Errorf("not a git repository: %s", projectPath)
		}

		// Copy hook script
		hookSource := filepath.Join(projectPath, "scripts", "git-hooks", "pre-commit-doc-sync")
		hookTarget := filepath.Join(gitDir, "hooks", "pre-commit")

		if installHooks {
			fmt.Printf("Installing pre-commit hook...\n")

			// Check if hook already exists
			if _, err := os.Stat(hookTarget); err == nil {
				fmt.Printf("Warning: Pre-commit hook already exists at %s\n", hookTarget)
				if interactive {
					fmt.Printf("Overwrite? (y/n): ")
					var response string
					_, _ = fmt.Scanln(&response)
					if response != "y" && response != "yes" {
						fmt.Println("Skipping hook installation.")
						return nil
					}
				} else {
					fmt.Println("Use --interactive to confirm overwrite.")
					return nil
				}
			}

			// Read source hook
			hookContent, err := os.ReadFile(hookSource)
			if err != nil {
				return fmt.Errorf("failed to read hook source: %w", err)
			}

			// Write target hook
			err = os.WriteFile(hookTarget, hookContent, 0755)
			if err != nil {
				return fmt.Errorf("failed to write hook: %w", err)
			}

			fmt.Printf("✅ Pre-commit hook installed at %s\n", hookTarget)
		} else {
			fmt.Printf("Hook script available at: %s\n", hookSource)
			fmt.Printf("To install manually, run:\n")
			fmt.Printf("  cp %s %s\n", hookSource, hookTarget)
			fmt.Printf("  chmod +x %s\n", hookTarget)
		}

		// Create config file template
		configPath := filepath.Join(projectPath, ".documentation-sync.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			fmt.Printf("\nCreating configuration template...\n")

			configContent := `# Documentation Sync Configuration
memory_bank:
  project_id: "` + filepath.Base(projectPath) + `"

analysis:
  code_patterns:
    - pattern: "src/api/**"
      documentation: ["docs/api/", "README.md#api"]
      priority: high
    
    - pattern: "src/cli/**"
      documentation: ["docs/cli/", "README.md#usage"]
      priority: high
    
    - pattern: "config/**"
      documentation: ["docs/configuration.md", "README.md#configuration"]
      priority: medium

workflows:
  pre_commit:
    enabled: true
    interactive: true
`

			err = os.WriteFile(configPath, []byte(configContent), 0644)
			if err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}

			fmt.Printf("✅ Configuration template created at %s\n", configPath)
		} else {
			fmt.Printf("Configuration file already exists at %s\n", configPath)
		}

		fmt.Printf("\n🎉 Documentation sync automation setup complete!\n")
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("1. Review and customize %s\n", configPath)
		fmt.Printf("2. Test the setup: git add . && git commit -m \"Test doc sync\"\n")
		fmt.Printf("3. Create documentation mappings using: memory-bank doc create-mapping\n")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(docSyncCmd)

	// Add subcommands
	docSyncCmd.AddCommand(docAnalyzeCmd)
	docSyncCmd.AddCommand(docSuggestCmd)
	docSyncCmd.AddCommand(docCreateMappingCmd)
	docSyncCmd.AddCommand(docValidateCmd)
	docSyncCmd.AddCommand(docSetupCmd)

	// Flags for suggest-updates
	docSuggestCmd.Flags().String("change-type", "", "type of change (api, cli, config, database, build)")
	docSuggestCmd.Flags().String("component", "", "specific component that changed")

	// Flags for create-mapping
	docCreateMappingCmd.Flags().String("code-pattern", "", "code pattern (e.g., 'src/api/**')")
	docCreateMappingCmd.Flags().StringSlice("documentation-files", []string{}, "documentation files affected")
	docCreateMappingCmd.Flags().String("change-type", "", "type of change this mapping applies to")
	docCreateMappingCmd.Flags().String("priority", "medium", "priority level (low, medium, high)")
	docCreateMappingCmd.Flags().StringP("project", "p", "", "project ID")

	// Flags for validate-consistency
	docValidateCmd.Flags().String("project-path", "", "project path to analyze (default: current directory)")
	docValidateCmd.Flags().StringSlice("focus-areas", []string{}, "specific areas to focus on (api, cli, config, etc.)")

	// Flags for setup-automation
	docSetupCmd.Flags().String("project-path", "", "project path (default: current directory)")
	docSetupCmd.Flags().Bool("interactive", false, "interactive mode for confirmations")
	docSetupCmd.Flags().Bool("install-hooks", false, "automatically install git hooks")
}
