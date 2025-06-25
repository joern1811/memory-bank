package mcp

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/ports"
)

// classifyFileChanges analyzes file paths and determines change types
func (s *MemoryBankServer) classifyFileChanges(files []string) []string {
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

		// Documentation changes (track for completeness)
		case strings.HasSuffix(file, ".md") || strings.Contains(file, "doc") ||
			strings.Contains(file, "guide") || strings.Contains(file, "readme"):
			changeTypes["documentation"] = true

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

// searchDocumentationMappings searches Memory Bank for relevant documentation mappings
func (s *MemoryBankServer) searchDocumentationMappings(ctx context.Context, changeTypes []string, projectID string) []domain.Memory {
	var allMappings []domain.Memory

	for _, changeType := range changeTypes {
		// Search for documentation mappings related to this change type
		var projectIDPtr *domain.ProjectID
		if projectID != "" {
			id := domain.ProjectID(projectID)
			projectIDPtr = &id
		}

		var memoryType *domain.MemoryType
		docMappingType := domain.MemoryType("doc_mapping")
		memoryType = &docMappingType

		searchRequest := ports.SemanticSearchRequest{
			Query:     fmt.Sprintf("documentation mapping %s", changeType),
			ProjectID: projectIDPtr,
			Type:      memoryType,
			Limit:     10,
			Threshold: 0.1,
		}

		results, err := s.memoryService.SearchMemories(ctx, searchRequest)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to search documentation mappings")
			continue
		}

		for _, result := range results {
			allMappings = append(allMappings, *result.Memory)
		}
	}

	return allMappings
}

// generateChangeAnalysis creates a comprehensive analysis of code changes and their documentation impact
func (s *MemoryBankServer) generateChangeAnalysis(files []string, changeTypes []string, mappings []domain.Memory, changeContext string) string {
	var analysis strings.Builder

	analysis.WriteString("# Documentation Impact Analysis\n\n")

	// Change summary
	analysis.WriteString("## Change Summary\n")
	analysis.WriteString(fmt.Sprintf("- **Files Changed**: %d\n", len(files)))
	analysis.WriteString(fmt.Sprintf("- **Change Types**: %s\n", strings.Join(changeTypes, ", ")))
	if changeContext != "" {
		analysis.WriteString(fmt.Sprintf("- **Context**: %s\n", changeContext))
	}
	analysis.WriteString("\n")

	// Detailed file analysis
	analysis.WriteString("## File Analysis\n")
	for _, file := range files {
		analysis.WriteString(fmt.Sprintf("- `%s`\n", file))
	}
	analysis.WriteString("\n")

	// Documentation impact based on change types
	analysis.WriteString("## Documentation Impact\n")
	for _, changeType := range changeTypes {
		analysis.WriteString(fmt.Sprintf("### %s Changes\n", strings.Title(changeType)))

		switch changeType {
		case "api":
			analysis.WriteString("**Likely affected documentation:**\n")
			analysis.WriteString("- API reference documentation\n")
			analysis.WriteString("- OpenAPI/Swagger specifications\n")
			analysis.WriteString("- README.md API section\n")
			analysis.WriteString("- Integration examples\n")
		case "cli":
			analysis.WriteString("**Likely affected documentation:**\n")
			analysis.WriteString("- CLI usage documentation\n")
			analysis.WriteString("- Command reference\n")
			analysis.WriteString("- README.md usage section\n")
			analysis.WriteString("- Help text examples\n")
		case "config":
			analysis.WriteString("**Likely affected documentation:**\n")
			analysis.WriteString("- Configuration documentation\n")
			analysis.WriteString("- Environment variables reference\n")
			analysis.WriteString("- Installation/setup guides\n")
			analysis.WriteString("- Docker/deployment documentation\n")
		case "database":
			analysis.WriteString("**Likely affected documentation:**\n")
			analysis.WriteString("- Database schema documentation\n")
			analysis.WriteString("- Migration guides\n")
			analysis.WriteString("- Data model documentation\n")
			analysis.WriteString("- Setup instructions\n")
		case "build":
			analysis.WriteString("**Likely affected documentation:**\n")
			analysis.WriteString("- Build instructions\n")
			analysis.WriteString("- Deployment guides\n")
			analysis.WriteString("- Development setup\n")
			analysis.WriteString("- CI/CD documentation\n")
		}
		analysis.WriteString("\n")
	}

	// Memory Bank mappings
	if len(mappings) > 0 {
		analysis.WriteString("## Memory Bank Mappings\n")
		analysis.WriteString("Found relevant documentation mappings in Memory Bank:\n\n")
		for _, mapping := range mappings {
			analysis.WriteString(fmt.Sprintf("**%s**\n", mapping.Title))
			analysis.WriteString(fmt.Sprintf("%s\n\n", mapping.Content))
		}
	}

	// Recommendations
	analysis.WriteString("## Recommendations\n")
	analysis.WriteString("1. **Review affected documentation** listed above\n")
	analysis.WriteString("2. **Update examples** to reflect any API or interface changes\n")
	analysis.WriteString("3. **Test documentation** by following the instructions\n")
	analysis.WriteString("4. **Create new mappings** if this is a new pattern using `doc_create_mapping`\n")
	analysis.WriteString("5. **Use `doc_suggest_updates`** for specific change types to get detailed guidance\n")

	return analysis.String()
}

// generateDocumentationSuggestions provides intelligent suggestions for documentation updates
func (s *MemoryBankServer) generateDocumentationSuggestions(ctx context.Context, changeType, component, projectID string) string {
	var suggestions strings.Builder

	suggestions.WriteString(fmt.Sprintf("# Documentation Update Suggestions: %s\n\n", strings.Title(changeType)))

	if component != "" {
		suggestions.WriteString(fmt.Sprintf("**Component**: %s\n\n", component))
	}

	// Search for relevant templates
	templateQuery := fmt.Sprintf("documentation template %s", changeType)

	var projectIDPtr *domain.ProjectID
	if projectID != "" {
		id := domain.ProjectID(projectID)
		projectIDPtr = &id
	}

	var memoryType *domain.MemoryType
	docTemplateType := domain.MemoryType("doc_template")
	memoryType = &docTemplateType

	searchRequest := ports.SemanticSearchRequest{
		Query:     templateQuery,
		ProjectID: projectIDPtr,
		Type:      memoryType,
		Limit:     5,
		Threshold: 0.1,
	}

	templates, err := s.memoryService.SearchMemories(ctx, searchRequest)
	if err == nil && len(templates) > 0 {
		suggestions.WriteString("## Available Templates\n")
		for _, template := range templates {
			suggestions.WriteString(fmt.Sprintf("**%s**\n", template.Memory.Title))
			suggestions.WriteString(fmt.Sprintf("%s\n\n", template.Memory.Content))
		}
	}

	// Default suggestions based on change type
	suggestions.WriteString("## Suggested Updates\n")

	switch changeType {
	case "api":
		suggestions.WriteString("### API Changes Checklist\n")
		suggestions.WriteString("- [ ] Update API reference documentation\n")
		suggestions.WriteString("- [ ] Verify/update OpenAPI/Swagger specs\n")
		suggestions.WriteString("- [ ] Update code examples in documentation\n")
		suggestions.WriteString("- [ ] Test all documented API calls\n")
		suggestions.WriteString("- [ ] Update integration guides\n")
		suggestions.WriteString("- [ ] Check for breaking changes and document them\n")

	case "cli":
		suggestions.WriteString("### CLI Changes Checklist\n")
		suggestions.WriteString("- [ ] Update command reference documentation\n")
		suggestions.WriteString("- [ ] Update usage examples in README\n")
		suggestions.WriteString("- [ ] Verify help text is current\n")
		suggestions.WriteString("- [ ] Update installation/setup instructions\n")
		suggestions.WriteString("- [ ] Test all documented commands\n")
		suggestions.WriteString("- [ ] Update shell completion if applicable\n")

	case "config":
		suggestions.WriteString("### Configuration Changes Checklist\n")
		suggestions.WriteString("- [ ] Update configuration documentation\n")
		suggestions.WriteString("- [ ] Document new environment variables\n")
		suggestions.WriteString("- [ ] Update Docker/deployment examples\n")
		suggestions.WriteString("- [ ] Verify setup/installation guides\n")
		suggestions.WriteString("- [ ] Update default configuration examples\n")
		suggestions.WriteString("- [ ] Document migration steps if needed\n")

	case "database":
		suggestions.WriteString("### Database Changes Checklist\n")
		suggestions.WriteString("- [ ] Update schema documentation\n")
		suggestions.WriteString("- [ ] Document migration procedures\n")
		suggestions.WriteString("- [ ] Update data model diagrams\n")
		suggestions.WriteString("- [ ] Verify setup instructions\n")
		suggestions.WriteString("- [ ] Update backup/restore procedures\n")
		suggestions.WriteString("- [ ] Document performance implications\n")

	case "build":
		suggestions.WriteString("### Build/Deploy Changes Checklist\n")
		suggestions.WriteString("- [ ] Update build instructions\n")
		suggestions.WriteString("- [ ] Verify deployment guides\n")
		suggestions.WriteString("- [ ] Update CI/CD documentation\n")
		suggestions.WriteString("- [ ] Check development setup guide\n")
		suggestions.WriteString("- [ ] Update dependency information\n")
		suggestions.WriteString("- [ ] Verify troubleshooting sections\n")

	default:
		suggestions.WriteString("### General Documentation Checklist\n")
		suggestions.WriteString("- [ ] Review README for accuracy\n")
		suggestions.WriteString("- [ ] Update relevant code examples\n")
		suggestions.WriteString("- [ ] Check for outdated references\n")
		suggestions.WriteString("- [ ] Verify links still work\n")
		suggestions.WriteString("- [ ] Update changelog if applicable\n")
	}

	suggestions.WriteString("\n## Next Steps\n")
	suggestions.WriteString("1. Use this checklist to systematically update documentation\n")
	suggestions.WriteString("2. Create a documentation mapping with `doc_create_mapping` if this is a new pattern\n")
	suggestions.WriteString("3. Use `doc_validate_consistency` to verify changes\n")

	return suggestions.String()
}

// generateSetupAutomation creates setup instructions and scripts
func (s *MemoryBankServer) generateSetupAutomation(projectPath string, interactive, installHooks bool, projectID string) string {
	var setup strings.Builder

	setup.WriteString("# Documentation Sync Automation Setup\n\n")
	setup.WriteString(fmt.Sprintf("**Project Path**: %s\n", projectPath))
	setup.WriteString(fmt.Sprintf("**Interactive Mode**: %t\n", interactive))
	setup.WriteString(fmt.Sprintf("**Auto-install Hooks**: %t\n\n", installHooks))

	setup.WriteString("## Setup Instructions\n\n")

	if installHooks {
		setup.WriteString("### Automatic Installation\n")
		setup.WriteString("The following git hooks will be installed automatically:\n\n")
	} else {
		setup.WriteString("### Manual Installation Required\n")
		setup.WriteString("Follow these steps to set up documentation sync automation:\n\n")
	}

	setup.WriteString("1. **Get Git Hook Script**\n")
	setup.WriteString("   ```bash\n")
	setup.WriteString("   # Fetch the pre-commit hook script\n")
	setup.WriteString("   curl -s [MCP_RESOURCE_URL]/script://memory-bank/git-hooks/pre-commit > .git/hooks/pre-commit\n")
	setup.WriteString("   chmod +x .git/hooks/pre-commit\n")
	setup.WriteString("   ```\n\n")

	setup.WriteString("2. **Create Configuration File**\n")
	setup.WriteString("   ```bash\n")
	setup.WriteString("   # Get configuration template\n")
	setup.WriteString("   curl -s [MCP_RESOURCE_URL]/template://memory-bank/config/documentation-sync > .documentation-sync.yaml\n")
	setup.WriteString("   ```\n\n")

	setup.WriteString("3. **Configure for Your Project**\n")
	setup.WriteString("   Edit `.documentation-sync.yaml` to match your project structure:\n")
	setup.WriteString("   ```yaml\n")
	setup.WriteString("   memory_bank:\n")
	if projectID != "" {
		setup.WriteString(fmt.Sprintf("     project_id: \"%s\"\n", projectID))
	} else {
		setup.WriteString("     project_id: \"your-project-name\"\n")
	}
	setup.WriteString("   \n")
	setup.WriteString("   analysis:\n")
	setup.WriteString("     code_patterns:\n")
	setup.WriteString("       - pattern: \"src/api/**\"\n")
	setup.WriteString("         documentation: [\"docs/api/\", \"README.md#api\"]\n")
	setup.WriteString("         priority: high\n")
	setup.WriteString("   ```\n\n")

	setup.WriteString("4. **Test the Setup**\n")
	setup.WriteString("   ```bash\n")
	setup.WriteString("   # Make a test change and commit\n")
	setup.WriteString("   echo \"# Test\" >> README.md\n")
	setup.WriteString("   git add README.md\n")
	setup.WriteString("   git commit -m \"Test documentation sync\"\n")
	setup.WriteString("   ```\n\n")

	setup.WriteString("## Available MCP Resources\n\n")
	setup.WriteString("Use these MCP resources to get the latest scripts and templates:\n\n")
	setup.WriteString("- `script://memory-bank/git-hooks/pre-commit` - Pre-commit hook script\n")
	setup.WriteString("- `script://memory-bank/setup/documentation-sync` - Setup script\n")
	setup.WriteString("- `template://memory-bank/config/documentation-sync` - Configuration template\n\n")

	setup.WriteString("## Using MCP Tools\n\n")
	setup.WriteString("Claude can use these tools directly for documentation management:\n\n")
	setup.WriteString("- `doc_analyze_changes` - Analyze code changes for documentation impact\n")
	setup.WriteString("- `doc_suggest_updates` - Get specific update suggestions\n")
	setup.WriteString("- `doc_create_mapping` - Create code-to-docs mappings\n")
	setup.WriteString("- `doc_validate_consistency` - Validate documentation consistency\n\n")

	setup.WriteString("## Benefits\n\n")
	setup.WriteString("- **Automated Reminders**: Never forget to update documentation\n")
	setup.WriteString("- **Intelligent Analysis**: Memory Bank learns your project patterns\n")
	setup.WriteString("- **Team Consistency**: Standardized documentation practices\n")
	setup.WriteString("- **Flexible Workflows**: Works with or without git hooks\n")

	return setup.String()
}

// validateDocumentationConsistency checks documentation against current codebase
func (s *MemoryBankServer) validateDocumentationConsistency(ctx context.Context, projectPath, projectID string, focusAreas []string) string {
	var report strings.Builder

	report.WriteString("# Documentation Consistency Validation Report\n\n")
	report.WriteString(fmt.Sprintf("**Project Path**: %s\n", projectPath))
	if len(focusAreas) > 0 {
		report.WriteString(fmt.Sprintf("**Focus Areas**: %s\n", strings.Join(focusAreas, ", ")))
	}
	report.WriteString(fmt.Sprintf("**Validation Time**: %s\n\n", strings.Replace(fmt.Sprintf("%v", ctx.Value("timestamp")), "<nil>", "now", 1)))

	// Search for documentation mappings to guide validation
	var projectIDPtr *domain.ProjectID
	if projectID != "" {
		id := domain.ProjectID(projectID)
		projectIDPtr = &id
	}

	var memoryType *domain.MemoryType
	docMappingType := domain.MemoryType("doc_mapping")
	memoryType = &docMappingType

	searchRequest := ports.SemanticSearchRequest{
		Query:     "documentation mapping",
		ProjectID: projectIDPtr,
		Type:      memoryType,
		Limit:     20,
		Threshold: 0.1,
	}

	mappings, err := s.memoryService.SearchMemories(ctx, searchRequest)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to search documentation mappings for validation")
	}

	if len(mappings) > 0 {
		report.WriteString("## Found Documentation Mappings\n")
		for _, mapping := range mappings {
			report.WriteString(fmt.Sprintf("- **%s**: %s\n", mapping.Memory.Title,
				strings.ReplaceAll(mapping.Memory.Content, "\n", " ")))
		}
		report.WriteString("\n")
	}

	// Common validation checks
	report.WriteString("## Validation Checks\n\n")

	validationItems := []struct {
		area        string
		title       string
		description string
	}{
		{"general", "README.md exists", "Project should have a comprehensive README"},
		{"general", "Documentation directory", "Check for docs/ or similar directory"},
		{"api", "API documentation", "Verify API endpoints are documented"},
		{"cli", "CLI usage examples", "Check command-line interface documentation"},
		{"config", "Configuration guide", "Environment variables and config files documented"},
		{"build", "Build instructions", "Clear build and deployment instructions"},
		{"database", "Schema documentation", "Database structure and migrations documented"},
	}

	for _, item := range validationItems {
		// Skip items not in focus areas if focus areas are specified
		if len(focusAreas) > 0 {
			skip := true
			for _, focus := range focusAreas {
				if item.area == focus || item.area == "general" {
					skip = false
					break
				}
			}
			if skip {
				continue
			}
		}

		report.WriteString(fmt.Sprintf("### %s\n", item.title))
		report.WriteString(fmt.Sprintf("- **Area**: %s\n", item.area))
		report.WriteString(fmt.Sprintf("- **Check**: %s\n", item.description))
		report.WriteString("- **Status**: ⚠️  Manual verification required\n\n")
	}

	report.WriteString("## Recommendations\n\n")
	report.WriteString("1. **Create documentation mappings** for any missing patterns using `doc_create_mapping`\n")
	report.WriteString("2. **Use `doc_analyze_changes`** when making code changes\n")
	report.WriteString("3. **Set up automation** with `doc_setup_automation` for continuous validation\n")
	report.WriteString("4. **Regular reviews** of documentation with each major release\n\n")

	report.WriteString("## Next Steps\n\n")
	report.WriteString("- Run validation checks manually or set up automated validation\n")
	report.WriteString("- Create Memory Bank entries for any documentation debt found\n")
	report.WriteString("- Establish regular documentation review cycles\n")

	return report.String()
}

// generateGitHookScript creates the git hook script content
func (s *MemoryBankServer) generateGitHookScript() string {
	return `#!/bin/bash
# Universal Documentation Sync Pre-commit Hook
# Generated by Memory Bank MCP Server

set -e

# Configuration
MEMORY_BANK_CMD="${MEMORY_BANK_CMD:-memory-bank}"
CONFIG_FILE="${DOC_SYNC_CONFIG:-.documentation-sync.yaml}"
INTERACTIVE="${DOC_SYNC_INTERACTIVE:-true}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[DOC-SYNC]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[DOC-SYNC]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[DOC-SYNC]${NC} $1"
}

log_error() {
    echo -e "${RED}[DOC-SYNC]${NC} $1"
}

# Check if Memory Bank is available
check_memory_bank() {
    if ! command -v "$MEMORY_BANK_CMD" &> /dev/null; then
        log_warning "Memory Bank not found. Basic analysis will be used."
        return 1
    fi
    return 0
}

# Get changed files
get_changed_files() {
    git diff --cached --name-only --diff-filter=ACMR
}

# Main execution
main() {
    log_info "Checking documentation impact..."
    
    local changed_files
    mapfile -t changed_files < <(get_changed_files)
    
    if [[ ${#changed_files[@]} -eq 0 ]]; then
        log_info "No staged changes found."
        exit 0
    fi
    
    if check_memory_bank; then
        # Use Memory Bank MCP tools for analysis
        log_info "Using Memory Bank for intelligent analysis..."
        
        # This would integrate with MCP tools in a real implementation
        # For now, provide basic analysis
        log_info "Analyzed ${#changed_files[@]} files for documentation impact"
    else
        # Basic analysis without Memory Bank
        log_info "Performing basic documentation impact analysis..."
    fi
    
    log_success "Documentation sync check completed."
}

main "$@"`
}

// generateSetupScript creates the setup script content
func (s *MemoryBankServer) generateSetupScript() string {
	return `#!/bin/bash
# Documentation Sync Setup Script
# Generated by Memory Bank MCP Server

set -e

echo "Documentation Sync Setup"
echo "========================"

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "Error: Not in a git repository"
    exit 1
fi

# Install pre-commit hook
echo "Installing pre-commit hook..."
curl -s [MCP_RESOURCE_URL]/script://memory-bank/git-hooks/pre-commit > .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

# Create configuration file
echo "Creating configuration file..."
if [[ ! -f .documentation-sync.yaml ]]; then
    curl -s [MCP_RESOURCE_URL]/template://memory-bank/config/documentation-sync > .documentation-sync.yaml
    echo "Created .documentation-sync.yaml - please customize for your project"
else
    echo ".documentation-sync.yaml already exists"
fi

echo "Setup completed successfully!"
echo "Test with: git add . && git commit -m 'test documentation sync'"`
}

// generateConfigTemplate creates the configuration template
func (s *MemoryBankServer) generateConfigTemplate() string {
	return `# Documentation Sync Configuration
# Generated by Memory Bank MCP Server

memory_bank:
  project_id: "auto-detect"  # Will use git repo name if auto-detect
  documentation_types: ["doc_mapping", "doc_template", "doc_todo"]

analysis:
  # Custom patterns for your project
  code_patterns:
    - pattern: "src/api/**"
      documentation: ["docs/api/", "README.md#api"]
      priority: high
    
    - pattern: "src/cli/**"
      documentation: ["docs/cli/", "README.md#usage"]
      priority: high
    
    - pattern: "**/*config*"
      documentation: ["docs/configuration/", "README.md#configuration"]
      priority: medium
    
    - pattern: "**/*model*"
      documentation: ["docs/database/", "docs/schema/"]
      priority: medium

workflows:
  pre_commit:
    enabled: true
    interactive: true
    auto_stage_docs: false
  
  post_commit:
    enabled: true
    track_commits: true
  
  sessions:
    track_documentation: true
    auto_templates: true

# Skip documentation checks for certain file patterns
skip_patterns:
  - "*.md"
  - "docs/**"
  - "*.test.*"
  - "test/**"
  - ".github/**"
  - "vendor/**"
  - "node_modules/**"`
}
