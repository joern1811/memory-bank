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
Add to `~/.claude/settings.json`:

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

Add this to your project's `CLAUDE.md` file:

```markdown
## Memory Bank Integration

### Required Behaviors
When Memory Bank MCP tools are available, Claude MUST:

1. **START SESSIONS**: Begin with `memory_search` to understand context
2. **BEFORE TASKS**: Search for existing implementations
3. **DURING DEVELOPMENT**: Document decisions and patterns
4. **AFTER SOLVING**: Store solutions with prevention strategies
5. **COMPLEX WORK**: Use sessions for multi-step tasks

### Session Management Protocol
For complex tasks (3+ files, new features, debugging), Claude MUST:

1. **Start Session**: Use `session_start` with descriptive title and project context
2. **Log Progress**: Use `session_log` at key milestones with specific progress updates
3. **Complete Session**: Use `session_complete` with summary of outcomes and lessons learned

**Session Usage Examples:**
```
# Start complex task
session_start "Implementing JWT authentication system" --project my_project

# Log key progress points
session_log "Created JWT middleware with token validation"
session_log "Added login/logout endpoints with proper error handling" 
session_log "Implemented protected route middleware"

# Complete with summary
session_complete "Successfully implemented JWT authentication with middleware, endpoints, and route protection. All tests passing."
```

### Memory Types
- `decision`: Architectural choices with rationale
- `pattern`: Reusable solutions with examples
- `error_solution`: Bug fixes with prevention
- `code`: Useful snippets with context

### Tagging Strategy
- Technology: `go`, `javascript`, `react`, `api`
- Domain: `auth`, `database`, `frontend`, `backend`
- Type: `pattern`, `decision`, `error`, `optimization`
- Complexity: `simple`, `medium`, `complex`
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

### v1.10 - ChromaDB Distance Metric Fix
- Fixed semantic search returning 0 results
- Configured ChromaDB to use cosine distance metric
- Enhanced metadata handling for ChromaDB compatibility

### v1.9 - ChromaDB v2 API Migration
- Migrated from deprecated v1 API to v2 API
- Updated all vector operations to use v2 endpoints
- Enhanced health checks with dedicated heartbeat endpoint

### v1.8 - Complete MCP Server
- Fixed MCP protocol integration bugs
- All 16 MCP tools properly registered
- Complete Memory, Project, and Session operations

### Key Features Completed
- Complete MCP server with 16 tools
- CLI interface with session management
- Advanced search (faceted, enhanced, suggestions)
- Performance optimizations (caching, batching)
- Database migrations and configuration management

## Dependencies

- [Ollama](https://ollama.com/): Local embedding models
- [ChromaDB](https://www.trychroma.com/): Vector database
- [Cobra](https://cobra.dev/): CLI framework
- [Logrus](https://github.com/sirupsen/logrus): Structured logging