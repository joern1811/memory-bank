# Memory Bank - Development Documentation

## Overview

Memory Bank is a semantic memory management system for Claude Code using hexagonal architecture. It provides intelligent storage and retrieval of development knowledge including decisions, patterns, error solutions, and session context.

### Technology Stack
- **Language**: Go 1.21+
- **Database**: SQLite with automatic initialization
- **Embeddings**: Ollama (nomic-embed-text) with Mock fallback
- **Vector Store**: ChromaDB (v2 API) with Mock fallback
- **MCP**: github.com/mark3labs/mcp-go v0.32.0
- **Logging**: Logrus with structured JSON logging

### Memory Types
- `Decision`: Architectural decisions with rationale
- `Pattern`: Reusable code/design patterns
- `ErrorSolution`: Error fixes with prevention strategies
- `Code`: Code snippets with context
- `Documentation`: Project documentation
- `Session`: Development session tracking

## Current Status 🚀

✅ **FULLY FUNCTIONAL**: Memory Bank is a complete MCP server with CLI interface!

### Core Features
- **MCP Server**: Complete JSON-RPC protocol with 16 tools
- **CLI Interface**: Full command-line interface with Cobra framework
- **Semantic Search**: Vector-based similarity search with ChromaDB
- **Automatic Fallbacks**: Works without external dependencies (mock providers)
- **Memory Operations**: Full CRUD operations via MCP protocol and CLI
- **SQLite Storage**: Persistent memory storage with automatic schema setup
- **Health Monitoring**: Automatic health checks for Ollama and ChromaDB
- **Session Management**: Development session tracking and progress logging


## Quick Start

### Installation
```bash
# Homebrew (recommended)
brew tap joern1811/tap
brew install --cask joern1811/tap/memory-bank

# Or build from source
go build ./cmd/memory-bank
```

### Basic Usage
```bash
# Start MCP server
memory-bank

# Or use CLI commands
./memory-bank init /path/to/project --name "My Project"
./memory-bank memory create --type decision --title "Use JWT" --content "..."
./memory-bank search "authentication patterns"
```

### Production Setup

```bash
# Install Ollama for embeddings
curl -fsSL https://ollama.com/install.sh | sh
ollama pull nomic-embed-text

# Start ChromaDB for vector search
uvx --from "chromadb[server]" chroma run --host localhost --port 8000 --path ./chromadb_data &
```

## MCP Integration

### Claude Desktop Configuration

Add to your Claude Desktop configuration file:
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "memory-bank": {
      "command": "memory-bank",
      "args": [],
      "env": {
        "MEMORY_BANK_DB_PATH": "~/memory_bank.db"
      }
    }
  }
}
```

### Claude Code Configuration

**Option 1 - Project scope** (recommended for team projects):
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

**Option 2 - User scope** (for all your projects):
Use Claude Code CLI:

```bash
claude mcp add memory-bank \
  -e MEMORY_BANK_DB_PATH=~/memory_bank.db \
  --scope user \
  -- memory-bank
```

### Testing & Development

```bash
# MCP Inspector for debugging
npx @modelcontextprotocol/inspector ./memory-bank

# Manual JSON-RPC testing
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}' | ./memory-bank
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OLLAMA_BASE_URL` | `http://localhost:11434` | Ollama server URL |
| `OLLAMA_MODEL` | `nomic-embed-text` | Embedding model name |
| `CHROMADB_BASE_URL` | `http://localhost:8000` | ChromaDB server URL |
| `MEMORY_BANK_DB_PATH` | `./memory_bank.db` | SQLite database path |
| `MEMORY_BANK_LOG_LEVEL` | `info` | Logging level |

### Troubleshooting

1. **Server fails to start**: Check binary path and permissions
2. **Ollama connection fails**: Verify Ollama is running on port 11434
3. **ChromaDB connection fails**: Ensure ChromaDB is running on port 8000
4. **Debug mode**: Set `MEMORY_BANK_LOG_LEVEL=debug`

### Development

```bash
# Build and test
go mod tidy
go test ./...
go build ./cmd/memory-bank

# Start server
./memory-bank
```

## Usage Best Practices

### Memory Type Guidelines

- **Decision**: Architectural choices with rationale and alternatives
- **Pattern**: Reusable code/design patterns with examples
- **ErrorSolution**: Bug fixes with root cause and prevention
- **Code**: Useful snippets with context and usage
- **Session**: Track complex development tasks

### Search Strategies

- Use semantic terms: "authentication middleware" not "authMiddleware"
- Try `memory_enhanced-search` for better relevance
- Use `memory_faceted-search` for complex filtering
- Use `memory_search-suggestions` for discovery

### Workflow Integration

1. **Before coding**: Search for existing patterns
2. **During development**: Document decisions as made
3. **After solving problems**: Store solutions immediately
4. **Complex tasks**: Use sessions for tracking progress

### Common Use Cases

- Document architectural decisions with rationale
- Store reusable patterns and code snippets
- Track error solutions and prevention strategies
- Manage development sessions for complex features

## CLAUDE.md Integration

### Memory Bank Protocol

Use these MANDATORY protocols in your project's `CLAUDE.md` file for maximum Memory Bank effectiveness:

```markdown
## Memory Bank Integration - MANDATORY PROTOCOLS

### MANDATORY: Pre-Task Search Protocol
Before ANY implementation task, Claude MUST execute this exact sequence:

1. **IMMEDIATE SEARCH**: Execute `memory_search "relevant keywords"` using 2-3 semantic terms
2. **ENHANCED SEARCH**: If <3 relevant results, use `memory_enhanced-search` with broader terms
3. **FACETED SEARCH**: For complex features, use `memory_faceted-search` with filters
4. **PATTERN CHECK**: Search specifically for patterns: `memory_search "pattern-type patterns" --type pattern`

### MANDATORY: Decision Documentation Protocol
Execute `memory_create` IMMEDIATELY when making these decisions:

**Decision Triggers (AUTOMATIC EXECUTION REQUIRED):**
- Technology/library choice (>30 seconds consideration)
- Architecture pattern selection
- API design decisions
- Database schema choices
- Security approach decisions

**Decision Template (EXACT FORMAT):**
```
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
```

### MANDATORY: Session Management Protocol
**Session Start Triggers (REQUIRED):**
- Modifying 3+ files in single task
- New feature implementation (>2 hour estimated work)
- Complex debugging (multiple investigation steps)
- Architectural decisions affecting multiple components

**Required Session Logging:**
```
session_start "Specific descriptive title" --project PROJECT_NAME

# At key milestones
session_log "COMPLETED: [specific accomplishment with technical details]"
session_log "DECISION POINT: [decision context] - evaluating [specific options]"  
session_log "RESOLVED: [specific problem] - solution: [specific solution]"

# At completion
session_complete "[summary] - [key learnings] - [time taken] - [follow-up needed]"
```

### MANDATORY: Pattern Storage Protocol
After successfully implementing ANY reusable solution:

```
memory_create --type pattern --title "[Pattern]: [Pattern name]" --content "
PURPOSE: [What problem this solves]
IMPLEMENTATION: [Key implementation details]
USAGE: [When and how to use this]
EXAMPLE: [Code example or reference]
DEPENDENCIES: [Required libraries/setup]
GOTCHAS: [Things to watch out for]
" --tags "pattern,[tech-stack],[domain]" --project PROJECT_NAME
```

### Execution Rules
1. **NO ASSUMPTIONS**: Always search before implementing
2. **IMMEDIATE DOCUMENTATION**: Document decisions when made, not later
3. **SPECIFIC TAGGING**: Use consistent, specific tags
4. **ACCURATE TIME TRACKING**: Record actual time spent
5. **ERROR CAPTURE**: Document ALL errors encountered
```






### CLI Usage Examples

```bash
# Initialize project
./memory-bank init /path/to/project --name "My Project"

# Create memories
./memory-bank memory create --type decision --title "Use JWT" --content "..."

# Search
./memory-bank search "authentication patterns"
./memory-bank search enhanced "JWT implementation" --type decision
./memory-bank search faceted "auth" --types decision,pattern

# Sessions
./memory-bank session start "Implementing auth" --project my_project
./memory-bank session log "Added JWT middleware"
./memory-bank session complete "Auth fully implemented"
```

## Development

### Testing
- Unit tests for domain logic
- Integration tests for database/vector stores
- Mock implementations for offline development

### Guidelines
- Follow Go conventions
- Write comprehensive tests
- Use structured logging
- Handle errors gracefully

### Security
- All embeddings generated locally (Ollama)
- Local SQLite storage
- Input validation and sanitization

## API Reference

### MCP Resources

- **`prompt://memory-bank/system`**: Dynamic system prompt with project context
- **`guide://memory-bank/project-setup`**: Project-specific integration guide
- **`guide://memory-bank/claude-integration`**: General integration templates

### MCP Methods

#### Memory Operations
- `memory_create`: Create new memory entry
- `memory_search`: Basic semantic search
- `memory_enhanced-search`: Enhanced search with relevance scoring
- `memory_faceted-search`: Advanced search with filters
- `memory_search-suggestions`: Get intelligent suggestions
- `memory_get`, `memory_update`, `memory_delete`, `memory_list`

#### Project Operations
- `project_init`: Initialize new project
- `project_get`: Get project by ID or path
- `project_list`: List all projects

#### Session Operations
- `session_start`: Start development session
- `session_log`: Log session progress
- `session_complete`: Complete session
- `session_get`, `session_list`, `session_abort`

## Recent Updates

### Latest - Documentation & Configuration Updates
- Fixed MCP configuration paths in documentation
- Added comprehensive development and production workflow targets
- Streamlined Memory Bank documentation structure
- Corrected MCP tool naming conventions in docs
- Updated ChromaDB configuration fields documentation

### Production Release Features
- Complete MCP server with 16 tools
- CLI interface with session management  
- Advanced search (faceted, enhanced, suggestions)
- Dynamic MCP resources with project analysis
- Homebrew Cask distribution setup
- Performance optimizations (caching, batching)
- Database migrations and configuration management

## Dependencies

- [Ollama](https://ollama.com/): Local embedding models
- [ChromaDB](https://www.trychroma.com/): Vector database
- [Cobra](https://cobra.dev/): CLI framework
- [Logrus](https://github.com/sirupsen/logrus): Structured logging