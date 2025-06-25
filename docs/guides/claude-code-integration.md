# Claude Code Integration Guide

Complete guide for integrating Memory Bank as an MCP server with Claude Code for optimal development workflows.

## Overview

Memory Bank is a semantic knowledge management system that runs as an MCP (Model Context Protocol) server and enhances Claude Code with:

- **Memory Management**: Store and retrieve development knowledge (decisions, patterns, solutions, code snippets)
- **Project Management**: Organize knowledge by project with automatic context switching
- **Session Tracking**: Track development sessions with progress logging and outcomes
- **Semantic Search**: Find relevant information using natural language queries
- **CLI Integration**: Complete command-line interface for direct access

## Installation

### Install Memory Bank MCP Server

```bash
# Option 1: Homebrew (recommended)
brew tap joern1811/tap
brew install --cask joern1811/tap/memory-bank

# Option 2: Build from source
git clone https://github.com/joern1811/memory-bank
cd memory-bank
go build ./cmd/memory-bank

# Verify installation
memory-bank --version
```

## Claude Code Configuration

### Option 1: Project-specific (recommended for teams)

Create `.mcp.json` in your project root:

```json
{
  "mcpServers": {
    "memory-bank": {
      "command": "memory-bank",
      "args": [],
      "env": {
        "MEMORY_BANK_DB_PATH": "./memory_bank.db"
      }
    }
  }
}
```

### Option 2: Global via Claude Code CLI (recommended)

```bash
# Global for all projects
claude mcp add memory-bank \
  -e MEMORY_BANK_DB_PATH=~/memory_bank.db \
  --scope user \
  -- memory-bank

# Or local for current directory only
claude mcp add memory-bank \
  -e MEMORY_BANK_DB_PATH=./memory_bank.db \
  --scope local \
  -- memory-bank

# Verify configuration
claude mcp list
```

### 3. Update CLAUDE.md

Add these specific Memory Bank integration instructions to your project's `CLAUDE.md`:

```markdown
## Memory Bank Integration - MANDATORY PROTOCOLS

### MANDATORY: Pre-Task Search Protocol
Before ANY implementation task, Claude MUST execute this exact sequence:

1. **IMMEDIATE SEARCH**: Execute memory_search "relevant keywords" using 2-3 semantic terms
2. **ENHANCED SEARCH**: If <3 relevant results, use memory_enhanced-search with broader terms  
3. **FACETED SEARCH**: For complex features, use memory_faceted-search with filters
4. **PATTERN CHECK**: Search specifically for patterns: memory_search "pattern-type patterns" --type pattern

**Search Response Evaluation:**
- If 0 results: Proceed with implementation but document as new pattern
- If 1-2 results: Review and adapt existing approach, document differences
- If 3+ results: Synthesize existing approaches, document consolidation decision

### MANDATORY: Session Management Protocol
Claude MUST start sessions for these triggers:

**Session Start Triggers (REQUIRED):**
- Modifying 3+ files in single task
- New feature implementation (>2 hour estimated work)
- Complex debugging (multiple investigation steps)
- Architectural decisions affecting multiple components
- Error resolution requiring code changes

**Required Session Commands:**
    session_start "Specific descriptive title" --project PROJECT_NAME
    
    # At key milestones
    session_log "COMPLETED: [specific accomplishment with technical details]"
    session_log "DECISION POINT: [decision context] - evaluating [specific options]"  
    session_log "RESOLVED: [specific problem] - solution: [specific solution]"
    
    # At completion
    session_complete "[summary] - [key learnings] - [time taken] - [follow-up needed]"

### MANDATORY: Decision Documentation Protocol
Execute memory_create IMMEDIATELY when making these decisions:

**Decision Triggers (AUTOMATIC EXECUTION REQUIRED):**
- Technology/library choice (>30 seconds consideration)
- Architecture pattern selection
- API design decisions
- Database schema choices
- Security approach decisions
- Performance optimization strategies

**Decision Template (EXACT FORMAT):**
    memory_create --type decision --title "[Decision]: [Concise description]" --content "
    CONTEXT: [Why this decision was needed]
    EVALUATED OPTIONS: 
    1. [Option 1] - [pros/cons]
    2. [Option 2] - [pros/cons] 
    CHOSEN: [Selected option]
    RATIONALE: [Why this option was selected]
    TRADE-OFFS: [What we're giving up]
    VALIDATION: [How we'll verify this was right]
    " --tags "decision,[relevant-tech],[component]" --project PROJECT_NAME

### MANDATORY: Pattern Storage Protocol
After successfully implementing ANY reusable solution, execute:

**Pattern Creation Triggers (AUTOMATIC):**
- Code that will be reused in this project
- Solutions to common problems
- Architectural patterns successfully applied
- Configuration patterns that work
- Testing patterns that prove effective

**Pattern Template (EXACT FORMAT):**
    memory_create --type pattern --title "[Pattern]: [Pattern name]" --content "
    PURPOSE: [What problem this solves]
    IMPLEMENTATION: [Key implementation details]
    USAGE: [When and how to use this]
    EXAMPLE: [Code example or reference]
    DEPENDENCIES: [Required libraries/setup]
    GOTCHAS: [Things to watch out for]
    " --tags "pattern,[tech-stack],[domain]" --project PROJECT_NAME

### MANDATORY: Error Solution Documentation
When resolving ANY error/bug, execute IMMEDIATELY after fix:

    memory_create --type error_solution --title "[ERROR]: [Brief error description]" --content "
    ERROR SYMPTOM: [What user observed]
    ROOT CAUSE: [Actual technical cause]
    SOLUTION: [Exact steps taken to fix]
    PREVENTION: [How to prevent recurrence]
    INVESTIGATION STEPS: [How we found the cause]
    TIME SPENT: [Actual investigation + fix time]
    " --tags "error,[error-type],[component]" --project PROJECT_NAME

### MANDATORY: Task Management Integration
For complex work, Claude MUST use task management:

**Task Creation (REQUIRED for estimated >4 hour work):**
    task_create --project PROJECT_NAME --title "[TASK]: [Specific title]" 
    --description "[Detailed description with acceptance criteria]" 
    --estimated-hours [realistic estimate] --priority [high/medium/low] 
    --tags "[relevant-tags]"

**Task Updates (REQUIRED at completion):**
    task_update TASK_ID --status done --actual-hours [actual time]

### MANDATORY: Documentation Sync
Before commits affecting user-facing features:

    doc_analyze_changes --changed-files "[file1,file2,file3]" 
    --change-context "[brief description of changes]"
    
    doc_suggest_updates --change-type [api|cli|config] --component "[component name]"

### Execution Rules
1. **NO ASSUMPTIONS**: Always search before implementing
2. **IMMEDIATE DOCUMENTATION**: Document decisions when made, not later
3. **SPECIFIC TAGGING**: Use consistent, specific tags (not generic ones)
4. **ACCURATE TIME TRACKING**: Record actual time, not estimates
5. **ERROR CAPTURE**: Document ALL errors encountered, even minor ones

### Verification Commands
Periodically verify Memory Bank usage:
    memory_list --project PROJECT_NAME --type decision   # Check decision capture
    memory_list --project PROJECT_NAME --type pattern    # Check pattern capture  
    task_list --project PROJECT_NAME --status done       # Check task completion
    session_list --project PROJECT_NAME                  # Check session history
```

## CLI Client Usage

### Basic Commands

```bash
# Initialize project
memory-bank init /path/to/project --name "My Project"

# Create memory
memory-bank memory create \
  --type decision \
  --title "Use JWT for Authentication" \
  --content "Decision to use JWT for stateless authentication..."

# Search
memory-bank search "authentication patterns"
memory-bank search enhanced "JWT implementation" --type decision

# Manage sessions
memory-bank session start "Implement Auth System" --project my-project
memory-bank session log "Created JWT middleware"
memory-bank session complete "Auth system fully implemented"

# Check system health
memory-bank health
```

### Advanced CLI Features

```bash
# Faceted search with filters
memory-bank search faceted "auth" --types decision,pattern --tags backend,security

# Memory management
memory-bank memory list --project my-project --type pattern
memory-bank memory get <memory-id>
memory-bank memory update <memory-id> --title "New Title"

# Project management
memory-bank project list
memory-bank project get --path /path/to/project

# Session tracking
memory-bank session list --project my-project --status active
memory-bank session get <session-id>
```

## Available MCP Tools

Memory Bank provides 30+ MCP tools organized into functional categories:

### Memory Management

| Tool | Purpose | Example Usage |
|------|---------|---------------|
| `memory_create` | Create new memory entries | Store decisions, patterns, solutions |
| `memory_search` | Basic semantic search | Find existing solutions quickly |
| `memory_enhanced-search` | Advanced search with relevance scoring | Get detailed match insights |
| `memory_faceted-search` | Complex filtering and faceted results | Search with multiple criteria |
| `memory_search-suggestions` | Get intelligent search suggestions | Discover related knowledge |
| `memory_get` | Retrieve specific memory by ID | Access exact memory entries |
| `memory_update` | Update existing memories | Keep information current |
| `memory_delete` | Remove outdated memories | Maintain clean knowledge base |
| `memory_list` | List memories with filters | Browse knowledge by type/project |

### Project Management

| Tool | Purpose | Example Usage |
|------|---------|---------------|
| `project_init` | Initialize new project | Set up project-specific knowledge base |
| `project_get` | Get project information | Retrieve project context and settings |
| `project_list` | List all projects | Browse available projects |
| `project_update` | Update project information | Change project name or description |
| `project_delete` | Delete project and data | Clean up obsolete projects |

### Session Management

| Tool | Purpose | Example Usage |
|------|---------|---------------|
| `session_start` | Begin development session | Track complex features or debugging |
| `session_log` | Log progress during session | Record milestones and decisions |
| `session_complete` | Complete session with outcome | Summarize results and learnings |
| `session_get` | Retrieve session details | Review past work and context |
| `session_list` | List sessions with filters | Browse development history |
| `session_abort` | Abort incomplete sessions | Clean up interrupted work |

### Task Management

| Tool | Purpose | Example Usage |
|------|---------|---------------|
| `task_create` | Create new development tasks | Plan features and track work |
| `task_get` | Retrieve task details | Check task status and progress |
| `task_update` | Update task information | Change status, add time, update details |
| `task_delete` | Remove tasks | Clean up completed or cancelled work |
| `task_list` | List tasks with filtering | View work by status, assignee, project |
| `task_add_dependency` | Add task dependencies | Link prerequisite tasks |
| `task_remove_dependency` | Remove task dependencies | Update task relationships |
| `task_add_subtask` | Create subtask relationships | Break down complex work |
| `task_remove_subtask` | Remove subtask relationships | Restructure task hierarchy |
| `task_statistics` | Get project task statistics | Track completion rates and metrics |
| `task_efficiency_report` | Generate efficiency reports | Analyze time estimates vs actual |

### Documentation Sync

| Tool | Purpose | Example Usage |
|------|---------|---------------|
| `doc_analyze_changes` | Analyze code changes for doc impact | Check what docs need updating |
| `doc_suggest_updates` | Get documentation update suggestions | Get specific guidance for changes |
| `doc_create_mapping` | Create code-to-docs relationships | Map components to documentation |
| `doc_validate_consistency` | Validate documentation state | Check for outdated documentation |
| `doc_setup_automation` | Set up automated workflows | Install git hooks and scripts |

### System Tools

| Tool | Purpose | Example Usage |
|------|---------|---------------|
| `system_health` | Check system connectivity | Verify Ollama and ChromaDB status |
| `version` | Get Memory Bank version info | Check installed version |

## Available MCP Prompts

Memory Bank provides intelligent prompts for common development workflows:

| Prompt | Purpose | Required Arguments | Optional Arguments |
|--------|---------|-------------------|-------------------|
| `start-debugging-session` | Start structured debugging session with memory-guided problem solving | `error_message` | `context` |
| `create-memory-pattern` | Store a reusable code pattern or solution in Memory Bank | `pattern_name`, `pattern_code` | `use_case` |
| `search-solutions` | Find similar solutions and patterns in your Memory Bank | `problem_description` | `technology` |
| `session-review` | Review current development session and get guidance on next steps | None | None |

### Using Prompts

Prompts provide intelligent, context-aware assistance for common development scenarios:

```bash
# Start a debugging session with guided problem solving
Claude: Use the "start-debugging-session" prompt to begin structured debugging.
# Automatically searches for similar errors and suggests investigation steps

# Create and store reusable patterns
Claude: Use "create-memory-pattern" to store this authentication middleware pattern.
# Guides you through documenting the pattern with proper context and usage

# Find relevant solutions
Claude: Use "search-solutions" to find existing authentication implementations.
# Performs intelligent search and presents relevant memories with context

# Review development progress
Claude: Use "session-review" to assess current work and plan next steps.
# Analyzes current session, suggests priorities, and identifies knowledge gaps
```

## Dynamic MCP Resources

Memory Bank provides dynamic resources that generate content based on your project:

| Resource | Purpose | Usage |
|----------|---------|-------|
| `prompt://memory-bank/system` | Dynamic system prompt with project context | Automatic Claude optimization |
| `guide://memory-bank/project-setup` | Project-specific integration guide | Tailored setup instructions |
| `guide://memory-bank/claude-integration` | General integration templates | Standard integration patterns |

## Workflow Examples

### Starting a New Feature

```bash
# 1. Search for existing patterns
Claude: "I'm implementing user authentication. Let me check existing patterns."
> memory_search "authentication implementation patterns" --type pattern

# 2. Start development session
> session_start "Implement JWT authentication system" --project my-app

# 3. Document architectural decision
> memory_create --type decision --title "Choose JWT for authentication" 
  --content "Decided on JWT tokens for stateless authentication. Alternatives considered: sessions, OAuth. JWT chosen for API scalability and mobile app support."

# 4. Log progress during development
> session_log "Created JWT middleware with token validation and refresh logic"
> session_log "Added login/logout endpoints with proper error handling"
> session_log "Implemented protected route middleware"

# 5. Store reusable pattern
> memory_create --type pattern --title "JWT Authentication Middleware"
  --content "Standard JWT middleware pattern with token validation, refresh, and error handling. Used across all protected routes."

# 6. Complete session
> session_complete "Successfully implemented JWT authentication. All tests passing. Pattern documented for reuse."
```

### Debugging Complex Issues

```bash
# 1. Search for similar errors
Claude: "Getting database connection timeouts. Let me check for similar issues."
> memory_search "database connection timeout postgres" --type error_solution

# 2. Start debugging session
> session_start "Debug database connection timeouts in production" --project my-app

# 3. Log investigation steps
> session_log "Checked connection pool settings - found max connections set too low"
> session_log "Analyzed slow query logs - identified N+1 query in user loading"
> session_log "Implemented connection pooling optimization and query improvements"

# 4. Document solution
> memory_create --type error_solution --title "Database connection timeout fix"
  --content "Root cause: Connection pool exhaustion from N+1 queries. Solution: Increase pool size and add query optimization. Prevention: Monitor connection metrics."

# 5. Complete with learnings
> session_complete "Fixed connection timeouts by optimizing queries and pool configuration. Added monitoring to prevent future issues."
```

### Code Review and Knowledge Sharing

```bash
# 1. Search for architectural patterns
Claude: "Reviewing this PR. Let me check our architectural guidelines."
> memory_search "microservice communication patterns" --type decision

# 2. Document new patterns found in PR
> memory_create --type pattern --title "Event-driven service communication"
  --content "New pattern using event bus for loose coupling between services. Improves scalability and maintainability."

# 3. Store review insights
> memory_create --type decision --title "API versioning strategy review"
  --content "Confirmed v1 API versioning approach is working well. Continue using semantic versioning for breaking changes."
```

### Project Onboarding

```bash
# 1. Initialize project knowledge
> project_init "/path/to/new-project" --name "E-commerce Platform" 
  --description "Multi-tenant e-commerce platform with microservices architecture"

# 2. Search for onboarding information
> memory_search "project setup development environment" --project new-project

# 3. Create onboarding documentation
> memory_create --type documentation --title "Development Environment Setup"
  --content "Complete setup guide for local development including Docker, database, and service dependencies."
```

## Advanced Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MEMORY_BANK_DB_PATH` | `./memory_bank.db` | SQLite database location |
| `MEMORY_BANK_LOG_LEVEL` | `info` | Logging verbosity |
| `OLLAMA_BASE_URL` | `http://localhost:11434` | Ollama server for embeddings |
| `OLLAMA_MODEL` | `nomic-embed-text` | Embedding model name |
| `CHROMADB_BASE_URL` | `http://localhost:8000` | ChromaDB server for vector search |

### Production Setup

For enhanced performance with semantic search:

```bash
# 1. Install Ollama for embeddings
curl -fsSL https://ollama.com/install.sh | sh
ollama pull nomic-embed-text

# 2. Start ChromaDB for vector search
uvx --from "chromadb[server]" chroma run --host localhost --port 8000 --path ./chromadb_data &

# 3. Configure Memory Bank to use external services
export OLLAMA_BASE_URL="http://localhost:11434"
export CHROMADB_BASE_URL="http://localhost:8000"
```

### Team Configuration

For team environments, use shared database with project-specific configuration in `.mcp.json`:

```json
{
  "mcpServers": {
    "memory-bank": {
      "command": "memory-bank",
      "args": [],
      "env": {
        "MEMORY_BANK_DB_PATH": "/shared/team/memory_bank.db",
        "MEMORY_BANK_LOG_LEVEL": "warn"
      }
    }
  }
}
```

Or use Claude Code CLI for team setup:

```bash
# Add Memory Bank with shared database for the team
claude mcp add memory-bank \
  -e MEMORY_BANK_DB_PATH=/shared/team/memory_bank.db \
  -e MEMORY_BANK_LOG_LEVEL=warn \
  --scope local \
  -- memory-bank
```

## Best Practices

### Memory Management

1. **Consistent Tagging**: Use standardized tags (auth, api, frontend, backend, database)
2. **Clear Titles**: Make memory titles searchable and descriptive
3. **Rich Context**: Include rationale and alternatives in decisions
4. **Regular Cleanup**: Remove outdated memories and update evolving patterns

### Session Management

1. **Meaningful Sessions**: Use sessions for complex work spanning multiple commits
2. **Progress Logging**: Log key milestones and decision points
3. **Complete Sessions**: Always complete sessions with outcomes and learnings
4. **Session Scope**: Keep sessions focused on specific features or problems

### Search Strategy

1. **Search First**: Always search before implementing new solutions
2. **Use Multiple Approaches**: Try different search terms and filters
3. **Leverage Suggestions**: Use search suggestions to discover related knowledge
4. **Filter Effectively**: Use project and type filters to narrow results

### Team Collaboration

1. **Shared Standards**: Establish team-wide tagging and naming conventions
2. **Knowledge Reviews**: Regularly review and update shared knowledge
3. **Onboarding Integration**: Include Memory Bank in new developer onboarding
4. **Cross-Project Learning**: Share patterns and solutions across projects

## Troubleshooting

### Common Issues

**MCP Server Not Starting**
```bash
# Check Memory Bank installation
memory-bank --version

# Test server manually
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}' | memory-bank

# Check Claude Code MCP configuration
claude-code --list-mcp-servers
```

**Search Not Working**
```bash
# Check external services
memory-bank health

# Test basic search
memory-bank search "test query" --limit 5

# Check database
ls -la memory_bank.db
```

**Performance Issues**
```bash
# Enable debug logging
export MEMORY_BANK_LOG_LEVEL=debug

# Check service connectivity
curl http://localhost:11434/api/version  # Ollama
curl http://localhost:8000/api/v1/version  # ChromaDB
```

### Getting Help

1. **System Health**: Use `system_health` MCP tool or `memory-bank health` CLI
2. **Logs**: Enable debug logging with `MEMORY_BANK_LOG_LEVEL=debug`
3. **MCP Inspector**: Use `npx @modelcontextprotocol/inspector ./memory-bank` for debugging
4. **Documentation**: Check [Memory Bank documentation](../README.md) for detailed guides

## Migration from Other Tools

### From Linear Knowledge Systems

1. **Export Existing Knowledge**: Extract decisions, patterns, and solutions
2. **Create Initial Memories**: Import key knowledge as Memory Bank entries
3. **Establish Workflow**: Integrate Memory Bank into existing development process
4. **Train Team**: Ensure team understands search and documentation practices

### From Documentation-Only Systems

1. **Dynamic Knowledge**: Transition from static docs to searchable, contextual knowledge
2. **Session Tracking**: Add progress tracking to development workflows
3. **Pattern Recognition**: Identify and codify reusable patterns from existing code
4. **Continuous Learning**: Establish practices for ongoing knowledge capture

---

Memory Bank transforms Claude Code into an intelligent development partner that learns from your team's experience and provides contextual assistance throughout the development lifecycle.