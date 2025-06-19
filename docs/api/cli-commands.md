# Memory Bank CLI Reference

Memory Bank provides a comprehensive command-line interface for managing semantic memories, projects, and development sessions. This document covers all available CLI commands with examples and usage patterns.

## Installation and Setup

```bash
# Build the application
go build ./cmd/memory-bank

# Make it executable and add to PATH (optional)
chmod +x memory-bank
mv memory-bank /usr/local/bin/
```

## Global Options

All commands support these global flags:

- `--help, -h`: Show help information
- `--config`: Path to configuration file (default: ~/.memory-bank/config.yaml)
- `--db-path`: Path to SQLite database (default: ./memory_bank.db)
- `--verbose, -v`: Enable verbose logging

## Command Overview

```bash
memory-bank [command] [subcommand] [flags]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  config      Configuration management
  help        Help about any command
  init        Initialize a new project for memory management
  memory      Manage memory entries
  migrate     Database migration management
  project     Manage projects
  search      Search across all memory entries
  server      Start the MCP server
  session     Manage development sessions
  version     Print the version number of memory-bank
```

## Project Management

### `init` - Initialize Project

Initialize a new project for memory management.

**Usage:**
```bash
memory-bank init [path] [flags]
```

**Flags:**
- `--name`: Project name (required)
- `--description`: Project description
- `--force`: Overwrite existing project

**Examples:**
```bash
# Initialize current directory
memory-bank init . --name "My API Project" --description "RESTful API with authentication"

# Initialize specific path
memory-bank init /path/to/project --name "Frontend App" 

# Force initialization (overwrite existing)
memory-bank init . --name "New Project" --force
```

**Output:**
```
✓ Project initialized successfully
  ID: proj_abc123
  Name: My API Project
  Path: /current/directory
  Database: ./memory_bank.db
```

### `project list` - List Projects

List all initialized projects.

**Usage:**
```bash
memory-bank project list [flags]
```

**Flags:**
- `--limit`: Number of results (default: 20)
- `--offset`: Pagination offset

**Examples:**
```bash
# List all projects
memory-bank project list

# Paginated listing
memory-bank project list --limit 5 --offset 10
```

**Output:**
```
ID          Name              Path                    Created           Memories
proj_abc123 My API Project    /current/directory     2024-01-15 10:30  15
proj_def456 Frontend App      /path/to/frontend      2024-01-14 14:20  8
proj_ghi789 Data Pipeline     /path/to/pipeline      2024-01-13 09:15  22

Total: 3 projects
```

### `project get` - Get Project Details

Retrieve detailed information about a specific project.

**Usage:**
```bash
memory-bank project get [project-id-or-name]
```

**Examples:**
```bash
# Get by project ID
memory-bank project get proj_abc123

# Get by project name
memory-bank project get "My API Project"
```

**Output:**
```
Project: proj_abc123
Name: My API Project
Description: RESTful API with authentication
Path: /current/directory
Created: 2024-01-15 10:30:00
Updated: 2024-01-15 15:45:00

Statistics:
├─ Total Memories: 15
├─ Decisions: 5
├─ Patterns: 6
├─ Error Solutions: 2
├─ Code Snippets: 1
└─ Documentation: 1

Recent Activity:
├─ Last Memory: 2024-01-15 15:45 (JWT Authentication Pattern)
├─ Active Sessions: 1
└─ Completed Sessions: 3

Tags (Top 5):
├─ auth (8)
├─ api (6)
├─ security (5)
├─ middleware (4)
└─ database (3)
```

## Memory Management

### `memory create` - Create Memory Entry

Create a new memory entry with automatic semantic embedding.

**Usage:**
```bash
memory-bank memory create [flags]
```

**Flags:**
- `--type`: Memory type (decision, pattern, error_solution, code, documentation) *required*
- `--title`: Memory title *required*
- `--content`: Memory content *required*
- `--tags`: Comma-separated tags
- `--project`: Project ID or name
- `--session`: Session ID

**Examples:**
```bash
# Create a decision memory
memory-bank memory create \
  --type decision \
  --title "Use PostgreSQL for primary database" \
  --content "After evaluating PostgreSQL, MySQL, and MongoDB, we chose PostgreSQL for its ACID compliance, JSON support, and strong ecosystem." \
  --tags "database,architecture,postgresql" \
  --project "my-project"

# Create a code pattern memory
memory-bank memory create \
  --type pattern \
  --title "JWT Authentication Middleware" \
  --content "func AuthMiddleware(next http.Handler) http.Handler { ... }" \
  --tags "go,auth,middleware,jwt"

# Create error solution memory
memory-bank memory create \
  --type error_solution \
  --title "Fix CORS preflight issues" \
  --content "Add OPTIONS handler and proper CORS headers for preflight requests" \
  --tags "cors,http,api"
```

### `memory list` - List Memory Entries

List memory entries with optional filtering.

**Usage:**
```bash
memory-bank memory list [flags]
```

**Flags:**
- `--project`: Filter by project ID or name
- `--type`: Filter by memory type
- `--tags`: Filter by tags (comma-separated)
- `--limit`: Number of results (default: 20)
- `--offset`: Pagination offset

**Examples:**
```bash
# List all memories
memory-bank memory list

# List decisions for specific project
memory-bank memory list --project "my-project" --type decision

# List memories with specific tags
memory-bank memory list --tags "auth,security" --limit 10

# Paginated listing
memory-bank memory list --limit 5 --offset 10
```

**Output:**
```
ID          Type        Title                           Tags           Created
mem_abc123  decision    Use PostgreSQL for database     database,arch  2024-01-15 10:30
mem_def456  pattern     JWT Authentication Middleware   go,auth        2024-01-15 11:15
mem_ghi789  error_sol   Fix CORS preflight issues      cors,api       2024-01-15 12:00

Total: 3 memories
```

### `memory search` - Search Memory Entries

Perform semantic search within memory entries for a specific project.

**Usage:**
```bash
memory-bank memory search [query] [flags]
```

**Flags:**
- `--project`: Project ID or name *required*
- `--limit`: Number of results (default: 10)
- `--threshold`: Similarity threshold 0.0-1.0 (default: 0.5)
- `--type`: Filter by memory type

**Examples:**
```bash
# Search within project
memory-bank memory search "authentication patterns" --project "my-project"

# Search with high similarity threshold
memory-bank memory search "database design" --project "my-project" --threshold 0.8

# Search specific type only
memory-bank memory search "error handling" --project "my-project" --type error_solution
```

**Output:**
```
Found 2 results for "authentication patterns":

1. JWT Authentication Middleware (similarity: 0.87)
   Type: pattern
   Tags: go, auth, middleware, jwt
   Created: 2024-01-15 11:15
   
   func AuthMiddleware(next http.Handler) http.Handler { ... }

2. OAuth2 vs JWT Decision (similarity: 0.73)
   Type: decision  
   Tags: auth, security, oauth2, jwt
   Created: 2024-01-15 09:45
   
   After comparing OAuth2 and JWT approaches...
```

### `memory get` - Get Memory Entry

Retrieve a specific memory entry by ID.

**Usage:**
```bash
memory-bank memory get [memory-id]
```

**Examples:**
```bash
memory-bank memory get mem_abc123
```

**Output:**
```
Memory: mem_abc123
Title: Use PostgreSQL for database
Type: decision
Project: my-project (proj_def456)
Tags: database, architecture, postgresql
Created: 2024-01-15 10:30:00
Updated: 2024-01-15 10:30:00

Content:
After evaluating PostgreSQL, MySQL, and MongoDB, we chose PostgreSQL 
for its ACID compliance, JSON support, and strong ecosystem.
```

### `memory update` - Update Memory Entry

Update an existing memory entry.

**Usage:**
```bash
memory-bank memory update [memory-id] [flags]
```

**Flags:**
- `--title`: New title
- `--content`: New content
- `--tags`: New tags (comma-separated)
- `--type`: New type

**Examples:**
```bash
# Update title and tags
memory-bank memory update mem_abc123 \
  --title "Use PostgreSQL as Primary Database" \
  --tags "database,architecture,postgresql,primary"

# Update content only
memory-bank memory update mem_abc123 \
  --content "Updated rationale with performance benchmarks..."
```

### `memory delete` - Delete Memory Entry

Delete a memory entry and its embeddings.

**Usage:**
```bash
memory-bank memory delete [memory-id]
```

**Examples:**
```bash
memory-bank memory delete mem_abc123
```

## Global Search

### `search` - Search All Memories

Perform semantic search across all memories in all projects.

**Usage:**
```bash
memory-bank search [query] [flags]
```

**Flags:**
- `--limit`: Number of results (default: 10)
- `--threshold`: Similarity threshold 0.0-1.0 (default: 0.5)
- `--type`: Filter by memory type
- `--project`: Filter by project

**Examples:**
```bash
# Global search across all projects
memory-bank search "error handling patterns"

# Search with filters
memory-bank search "authentication" --type decision --limit 5

# High-precision search
memory-bank search "JWT implementation" --threshold 0.8
```

**Output:**
```
Global search results for "error handling patterns":

1. Error Handling Middleware (similarity: 0.91)
   Project: api-service
   Type: pattern
   Created: 2024-01-15 14:20
   
2. Database Error Recovery (similarity: 0.78)
   Project: backend-core
   Type: error_solution
   Created: 2024-01-14 16:30

3. Client-Side Error Boundaries (similarity: 0.72)
   Project: frontend-app
   Type: pattern
   Created: 2024-01-13 11:15
```

### Advanced Search Features (Future Enhancement)

Memory Bank's search command provides semantic search with various filters. Advanced search features like faceted search, enhanced relevance scoring, and search suggestions are planned for future releases.

**Current Search Capabilities:**
The basic `search` command provides powerful semantic search across all memories with the following features:

- **Semantic Search**: Vector-based similarity matching
- **Project Filtering**: Scope search to specific projects  
- **Type Filtering**: Filter by memory type (decision, pattern, etc.)
- **Similarity Thresholds**: Control result precision
- **Result Limiting**: Pagination and result count control

**Planned Advanced Features:**
- **Faceted Search**: Multi-dimensional filtering with real-time statistics
- **Enhanced Relevance Scoring**: Detailed match explanations and scoring breakdown
- **Search Suggestions**: AI-powered query suggestions based on content patterns
- **Content Highlighting**: Automatic highlighting of matched terms
- **Search Analytics**: Track search patterns and popular queries

*Note: The above advanced search features are documented for future implementation. Current users should use the basic `search` command with available flags.*

## Session Management

### `session start` - Start Development Session

Start a new development session for progress tracking.

**Usage:**
```bash
memory-bank session start [title] [flags]
```

**Flags:**
- `--project`: Project ID or name *required*
- `--description`: Session description
- `--tags`: Session tags (comma-separated)

**Examples:**
```bash
# Start simple session
memory-bank session start "Implement user authentication" --project "my-project"

# Start with description and tags
memory-bank session start "API Rate Limiting" \
  --project "api-service" \
  --description "Add rate limiting middleware with Redis backend" \
  --tags "api,middleware,redis,security"
```

**Output:**
```
✓ Development session started
  ID: sess_abc123
  Title: Implement user authentication
  Project: my-project
  Started: 2024-01-15 15:30:00
```

### `session log` - Log Session Progress

Add progress entries to the active session.

**Usage:**
```bash
memory-bank session log [content] [flags]
```

**Flags:**
- `--session`: Session ID (uses active session if not specified)
- `--type`: Entry type (info, milestone, issue, solution) default: info
- `--tags`: Entry tags (comma-separated)

**Examples:**
```bash
# Log general progress
memory-bank session log "Implemented JWT token generation"

# Log milestone with tags
memory-bank session log "Authentication middleware completed" \
  --type milestone \
  --tags "auth,middleware,complete"

# Log issue encountered
memory-bank session log "CORS preflight failing in production" \
  --type issue \
  --tags "cors,production,bug"

# Log solution found
memory-bank session log "Fixed CORS by adding OPTIONS handler" \
  --type solution \
  --tags "cors,fix,options"
```

### `session complete` - Complete Session

Complete the active development session with outcome.

**Usage:**
```bash
memory-bank session complete [outcome] [flags]
```

**Flags:**
- `--session`: Session ID (uses active session if not specified)
- `--summary`: Detailed completion summary

**Examples:**
```bash
# Simple completion
memory-bank session complete "Successfully implemented JWT authentication"

# Detailed completion
memory-bank session complete "Authentication system implemented" \
  --summary "Completed JWT-based auth with login/register endpoints, middleware, and comprehensive tests. Performance tests show <10ms token validation."
```

**Output:**
```
✓ Development session completed
  ID: sess_abc123
  Duration: 2h 45m
  Progress Entries: 8
  Outcome: Successfully implemented JWT authentication
```

### `session list` - List Sessions

List development sessions with optional filtering.

**Usage:**
```bash
memory-bank session list [flags]
```

**Flags:**
- `--project`: Filter by project ID or name
- `--status`: Filter by status (active, completed, aborted)
- `--limit`: Number of results (default: 20)
- `--offset`: Pagination offset

**Examples:**
```bash
# List all sessions
memory-bank session list

# List active sessions for project
memory-bank session list --project "my-project" --status active

# List recent completed sessions
memory-bank session list --status completed --limit 5
```

**Output:**
```
ID          Title                     Project     Status     Started          Duration
sess_abc123 Implement authentication  my-project  completed  2024-01-15 15:30  2h 45m
sess_def456 Add rate limiting         api-service active     2024-01-15 18:15  -
sess_ghi789 Fix CORS issues          frontend    completed  2024-01-15 09:00  1h 30m

Total: 3 sessions
```

### `session get` - Get Session Details

Retrieve detailed session information including progress log.

**Usage:**
```bash
memory-bank session get [session-id]
```

**Examples:**
```bash
memory-bank session get sess_abc123
```

**Output:**
```
Session: sess_abc123
Title: Implement user authentication
Project: my-project (proj_def456)
Status: completed
Started: 2024-01-15 15:30:00
Completed: 2024-01-15 18:15:00
Duration: 2h 45m
Tags: auth, api, security

Progress Log:
┌──────────────────────┬───────────┬─────────────────────────────────────┬──────────────────┐
│ Time                 │ Type      │ Content                             │ Tags             │
├──────────────────────┼───────────┼─────────────────────────────────────┼──────────────────┤
│ 2024-01-15 15:45:00  │ milestone │ JWT token generation implemented    │ jwt, auth        │
│ 2024-01-15 16:30:00  │ issue     │ CORS preflight failing             │ cors, bug        │
│ 2024-01-15 16:45:00  │ solution  │ Added OPTIONS handler for CORS     │ cors, fix        │
│ 2024-01-15 17:30:00  │ milestone │ Authentication middleware complete  │ middleware, auth │
└──────────────────────┴───────────┴─────────────────────────────────────┴──────────────────┘

Outcome: Successfully implemented JWT authentication
Summary: Completed JWT-based auth with login/register endpoints...
```

### `session abort` - Abort Sessions

Abort active sessions for a project.

**Usage:**
```bash
memory-bank session abort [flags]
```

**Flags:**
- `--project`: Project ID or name *required*

**Examples:**
```bash
memory-bank session abort --project "my-project"
```

**Output:**
```
✓ Aborted 1 active session(s) for project "my-project"
  Session: sess_def456 - "Add rate limiting"
```

## Additional Commands

### `completion` - Shell Autocompletion

Generate autocompletion scripts for various shells.

**Usage:**
```bash
memory-bank completion [shell]
```

**Supported Shells:**
- `bash`: Bash shell completion
- `zsh`: Zsh shell completion  
- `fish`: Fish shell completion
- `powershell`: PowerShell completion

**Examples:**
```bash
# Generate bash completion
memory-bank completion bash > /etc/bash_completion.d/memory-bank

# Generate zsh completion
memory-bank completion zsh > "${fpath[1]}/_memory-bank"

# Install completion for current shell
source <(memory-bank completion bash)

# Fish shell completion
memory-bank completion fish > ~/.config/fish/completions/memory-bank.fish
```

### `help` - Command Help

Get help information for any command.

**Usage:**
```bash
memory-bank help [command]
```

**Examples:**
```bash
# General help
memory-bank help

# Command-specific help
memory-bank help memory
memory-bank help search
memory-bank help session

# Subcommand help
memory-bank help memory create
memory-bank help session start
```

### `version` - Version Information

Display version information and build details.

**Usage:**
```bash
memory-bank version [flags]
```

**Flags:**
- `--verbose`: Show detailed build information

**Examples:**
```bash
# Basic version
memory-bank version

# Detailed version info
memory-bank version --verbose
```

**Output:**
```
Memory Bank v1.10.0

# With --verbose:
Memory Bank v1.10.0
  Build Date: 2024-01-15T10:30:00Z
  Git Commit: 29acdd6a
  Go Version: go1.21.5
  OS/Arch: darwin/amd64
  Compiler: gc
```

## Project Management

### `project` - Project Operations

Manage projects and project-specific operations.

**Usage:**
```bash
memory-bank project [subcommand]
```

**Subcommands:**
- `list`: List all projects
- `get`: Get project details
- `delete`: Delete a project (with confirmation)
- `stats`: Show project statistics

**Examples:**
```bash
# Project statistics
memory-bank project stats proj_abc123
memory-bank project stats "My API Project"

# Delete project (requires confirmation)
memory-bank project delete proj_abc123
```

**Project Stats Output:**
```
Project Statistics: My API Project

Memory Distribution:
┌────────────────┬───────┬──────────┐
│ Type           │ Count │ Avg Size │
├────────────────┼───────┼──────────┤
│ decision       │ 5     │ 245 chars│
│ pattern        │ 6     │ 412 chars│
│ error_solution │ 2     │ 189 chars│
│ code           │ 1     │ 567 chars│
│ documentation  │ 1     │ 234 chars│
└────────────────┴───────┴──────────┘

Activity Timeline:
├─ First Memory: 2024-01-10 (Database Choice Decision)
├─ Most Active Day: 2024-01-15 (6 memories created)
├─ Recent Activity: 2024-01-15 15:45
└─ Growth Rate: 2.3 memories/day (last 7 days)

Session Summary:
├─ Total Sessions: 4
├─ Completed: 3 (avg duration: 2h 15m)
├─ Active: 1
└─ Success Rate: 75%

Top Tags:
├─ auth (8) ████████████████
├─ api (6) ████████████
├─ security (5) ██████████
├─ middleware (4) ████████
└─ database (3) ██████
```

## Configuration Management

### `config` - Manage Configuration

Manage Memory Bank configuration settings.

**Usage:**
```bash
memory-bank config [subcommand]
```

**Subcommands:**
- `init`: Initialize default configuration file
- `show`: Display current configuration
- `set`: Set configuration value
- `validate`: Validate configuration file

**Examples:**
```bash
# Initialize default config
memory-bank config init

# Show current configuration
memory-bank config show

# Set configuration value
memory-bank config set embedding.provider ollama
memory-bank config set vector.provider chromadb
memory-bank config set database.path ./custom_memory.db

# Validate configuration
memory-bank config validate

# Get specific configuration value
memory-bank config get embedding.provider
memory-bank config get vector.chromadb.base_url

# Reset configuration to defaults
memory-bank config reset

# Export configuration (sanitized for sharing)
memory-bank config export --sanitized
```

**Additional Config Subcommands:**
```bash
# Show configuration locations checked
memory-bank config locations

# Show configuration with sources (env, file, default)
memory-bank config show --sources

# Set environment-specific configuration
memory-bank config set --env production database.path /var/lib/memory-bank/prod.db

# Validate external service connections
memory-bank config test-connections
```

**Configuration Export Output:**
```
# Memory Bank Configuration Export
# Generated: 2024-01-15 16:30:00
# Version: 1.10.0

database:
  path: "[REDACTED]"
  max_connections: 25

embedding:
  provider: "ollama"
  ollama:
    base_url: "http://localhost:11434"
    model: "nomic-embed-text"

vector:
  provider: "chromadb"
  chromadb:
    base_url: "http://localhost:8000"
    collection: "memory_bank"

logging:
  level: "info"
  format: "json"
```

## Database Management

### `migrate` - Database Migrations

Manage database schema migrations.

**Usage:**
```bash
memory-bank migrate [subcommand]
```

**Subcommands:**
- `status`: Show migration status
- `up`: Apply pending migrations
- `down`: Rollback migrations
- `reset`: Reset database to initial state

**Examples:**
```bash
# Check migration status
memory-bank migrate status

# Apply all pending migrations
memory-bank migrate up

# Rollback last migration
memory-bank migrate down

# Reset database (WARNING: destroys all data)
memory-bank migrate reset
```

## Future Enhancements

Memory Bank's CLI is designed for extensibility. Future versions may include additional utility commands for enhanced functionality:

### Planned Utility Commands

**Health and Diagnostics:**
- `health`: System health checks and diagnostics
- `benchmark`: Performance benchmarking tools
- `debug`: Debug utilities and troubleshooting tools

**Data Management:**
- `backup/restore`: Backup and restore functionality
- `export/import`: Data export/import in various formats
- `stats`: Usage statistics and analytics

**Advanced Features:**
- `watch`: Real-time monitoring of memory bank changes
- `sync`: Synchronization with remote memory banks
- `analyze`: Content analysis and insights

*Note: These features are planned for future releases. Current functionality is provided through the core commands: init, memory, search, session, config, migrate, project, server, completion, help, and version.*

## Server Mode

### `server` - Start MCP Server

Start Memory Bank in MCP server mode for Claude Code integration.

**Usage:**
```bash
memory-bank server [flags]
```

**Flags:**
- `--port`: Server port (default: auto-assigned)
- `--host`: Server host (default: localhost)

**Examples:**
```bash
# Start MCP server (default mode)
memory-bank server

# Start with specific port
memory-bank server --port 8080
```

## Environment Variables

Memory Bank supports configuration via environment variables:

```bash
# Database configuration
export MEMORY_BANK_DB_PATH="./memory_bank.db"

# Ollama configuration
export OLLAMA_BASE_URL="http://localhost:11434"
export OLLAMA_MODEL="nomic-embed-text"

# ChromaDB configuration  
export CHROMADB_BASE_URL="http://localhost:8000"
export CHROMADB_COLLECTION="memory_bank"

# Logging configuration
export MEMORY_BANK_LOG_LEVEL="info"
export MEMORY_BANK_LOG_FORMAT="json"
```

## Common Workflows

### Development Session Workflow

```bash
# 1. Start a session
memory-bank session start "Add user profiles" --project "my-app"

# 2. Log progress as you work
memory-bank session log "Created user profile model"
memory-bank session log "Added profile endpoints" --type milestone
memory-bank session log "Tests failing on validation" --type issue
memory-bank session log "Fixed validation with proper schema" --type solution

# 3. Complete the session
memory-bank session complete "User profiles implemented with full CRUD"
```

### Knowledge Capture Workflow

```bash
# 1. Capture architectural decisions
memory-bank memory create --type decision \
  --title "Microservices vs Monolith" \
  --content "Chose microservices for better scalability..." \
  --tags "architecture,microservices"

# 2. Document patterns and solutions
memory-bank memory create --type pattern \
  --title "Error Handling Middleware" \
  --content "Standard error middleware pattern..." \
  --tags "go,middleware,errors"

# 3. Search when needed
memory-bank search "error handling" --threshold 0.7
```

### Project Setup Workflow

```bash
# 1. Initialize project
memory-bank init . --name "E-commerce API" --description "Backend API for shop"

# 2. Configure if needed
memory-bank config set embedding.provider ollama
memory-bank config set vector.provider chromadb

# 3. Start capturing knowledge
memory-bank memory create --type decision --title "Initial architecture" ...
```

## Advanced CLI Usage

### Shell Integration

Memory Bank CLI integrates well with shell workflows:

```bash
# Use with shell pipes and filters
memory-bank memory list | grep "authentication"
memory-bank search "database" | head -5

# Combine with other CLI tools
memory-bank memory list --project "my-api" | wc -l  # Count memories
memory-bank project list | grep -E "(api|service)"  # Filter projects

# Shell aliases for common operations
alias mb='memory-bank'
alias mb-search='memory-bank search'
alias mb-create='memory-bank memory create'
```

### Configuration Best Practices

Optimize your Memory Bank configuration for daily use:

```bash
# Create project-specific configurations
cd /path/to/project
memory-bank init . --name "Project Name"

# Use environment variables for common settings
export MEMORY_BANK_LOG_LEVEL=info
export MEMORY_BANK_DB_PATH="./project_memory.db"

# Set up shell completion
memory-bank completion bash > ~/.memory-bank-completion
echo "source ~/.memory-bank-completion" >> ~/.bashrc
```

### Scripting Integration

Memory Bank is designed for integration with scripts and automation:

```bash
#!/bin/bash
# Example: Daily knowledge capture script

PROJECT="my-current-project"
DATE=$(date +%Y-%m-%d)

# Start a daily session
memory-bank session start "Daily work - $DATE" --project "$PROJECT"

# Your development work happens here...

# End session with summary
memory-bank session complete "Completed daily development tasks"

# Quick health check
memory-bank config validate
```

### Common Integration Patterns

#### Git Hooks Integration

Automatically capture significant commits as memories:

```bash
#!/bin/sh
# .git/hooks/post-commit
# Auto-capture significant commits as memories

COMMIT_MSG=$(git log -1 --pretty=%B)
PROJECT=$(basename $(git rev-parse --show-toplevel))

# Only capture commits with specific prefixes
if echo "$COMMIT_MSG" | grep -E "^(feat|fix|docs):" > /dev/null; then
  TYPE=$(echo "$COMMIT_MSG" | cut -d: -f1)
  
  case $TYPE in
    feat) MEMORY_TYPE="pattern" ;;
    fix) MEMORY_TYPE="error_solution" ;;
    docs) MEMORY_TYPE="documentation" ;;
  esac
  
  memory-bank memory create \
    --type "$MEMORY_TYPE" \
    --title "Git: $COMMIT_MSG" \
    --content "Auto-captured from commit" \
    --tags "git,auto-capture,$TYPE" \
    --project "$PROJECT"
fi
```

#### Development Workflow Integration

Example daily development workflow:

```bash
#!/bin/bash
# daily-workflow.sh

PROJECT="my-current-project"

# Start daily session
memory-bank session start "Daily development $(date +%Y-%m-%d)" --project "$PROJECT"

# Your development work happens here...
echo "Development session started. Work on your project..."

# At the end of the day, complete the session
echo "Ready to complete session? Press enter..."
read
memory-bank session complete "Completed daily development tasks" --project "$PROJECT"
```

## Troubleshooting CLI Issues

### Common CLI Problems

```bash
# Command not found
export PATH=$PATH:/usr/local/bin
which memory-bank

# Permission denied
chmod +x /usr/local/bin/memory-bank

# Configuration not loading
memory-bank config show
memory-bank config locations

# Database locked
ps aux | grep memory-bank
memory-bank health database
```

### Debug Mode

Enable debug output for troubleshooting:

```bash
# Enable debug logging
export MEMORY_BANK_LOG_LEVEL=debug
memory-bank --verbose memory create --type test --title "Debug test"

# Trace command execution
memory-bank --trace search "test query"

# Dry run mode (preview actions)
memory-bank memory delete mem_123 --dry-run
```

### Getting Help

```bash
# Command-specific help
memory-bank memory --help
memory-bank search enhanced --help

# Show available subcommands
memory-bank memory
memory-bank session

# Show version and build info
memory-bank version --verbose

# Check system compatibility
memory-bank health check --system
```

This comprehensive CLI reference covers all available commands, advanced features, and integration patterns for Memory Bank's command-line interface.
