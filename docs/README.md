# Memory Bank Documentation

Welcome to the Memory Bank documentation! This comprehensive guide covers everything you need to know about using Memory Bank for semantic memory management in your development projects.

## What is Memory Bank?

Memory Bank is a semantic memory management system designed specifically for developers. It helps you capture, organize, and retrieve development knowledge including architectural decisions, code patterns, error solutions, and session progress using natural language search powered by vector embeddings.

## Quick Navigation

### 🚀 Getting Started
- **[Getting Started Guide](guides/getting-started.md)** - Complete setup and first steps
- **[Examples and Use Cases](guides/examples.md)** - Real-world scenarios and workflows

### 📚 API Reference
- **[MCP Methods](api/mcp-methods.md)** - Complete MCP protocol reference for Claude Code integration
- **[CLI Commands](api/cli-commands.md)** - Comprehensive command-line interface reference

### 🏗️ Architecture
- **[Project README](../README.md)** - Architecture overview and technical details
- **[Development Guide](../CLAUDE.md)** - Detailed development documentation

## Documentation Structure

```
docs/
├── README.md              # This overview document
├── guides/
│   ├── getting-started.md # Setup and basic usage
│   └── examples.md        # Real-world use cases
└── api/
    ├── mcp-methods.md     # MCP protocol reference
    └── cli-commands.md    # CLI commands reference
```

## Key Features

### 🧠 Semantic Memory Management
- **Capture Decisions**: Store architectural decisions with context and rationale
- **Document Patterns**: Save reusable code patterns and design solutions  
- **Track Solutions**: Remember how you solved specific errors or problems
- **Manage Sessions**: Track development progress with structured logging

### 🔍 Intelligent Search
- **Vector-Based Search**: Find relevant knowledge using natural language queries
- **Similarity Matching**: Discover related content even when keywords don't match exactly
- **Contextual Filtering**: Search within specific projects, types, or time periods
- **Configurable Thresholds**: Adjust search precision based on your needs

### 🔧 Dual Interface
- **MCP Protocol**: Seamless integration with Claude Code for AI-assisted development
- **CLI Interface**: Traditional command-line tools for direct usage and automation
- **Automatic Fallbacks**: Works offline with mock providers when external services unavailable

### 💾 Persistent Storage
- **SQLite Database**: Local storage with automatic schema management
- **Vector Embeddings**: Powered by Ollama (local) or ChromaDB for enhanced search
- **Migration Support**: Schema versioning with upgrade and rollback capabilities

## Memory Types

Memory Bank supports five distinct types of knowledge:

| Type | Purpose | Examples |
|------|---------|----------|
| **Decision** | Architectural and technical decisions with rationale | Technology choices, design patterns, process decisions |
| **Pattern** | Reusable code patterns and design solutions | Code templates, utilities, best practices |
| **Error Solution** | Solutions to specific errors and problems | Bug fixes, configuration issues, performance optimizations |
| **Code** | Code snippets with context and explanations | Utility functions, configuration examples, integration samples |
| **Documentation** | Project documentation and explanations | Process docs, API guides, setup instructions |

## Common Workflows

### Knowledge Capture Workflow
1. **Make Decision** → Document with `memory create --type decision`
2. **Solve Problem** → Store solution with `memory create --type error_solution`
3. **Create Pattern** → Save for reuse with `memory create --type pattern`
4. **Need Information** → Search with `search "natural language query"`

### Development Session Workflow
1. **Start Work** → `session start "Feature description"`
2. **Log Progress** → `session log "Progress update"`
3. **Track Issues** → `session log "Problem found" --type issue`
4. **Document Solutions** → `session log "Solution applied" --type solution`
5. **Complete Work** → `session complete "Final outcome"`

### Team Knowledge Sharing
1. **Document Standards** → `memory create --type documentation`
2. **Share Patterns** → `memory create --type pattern` with team tags
3. **Preserve Solutions** → `memory create --type error_solution`
4. **Search History** → `memory search` for past decisions and solutions

## Integration Options

### With Claude Code (MCP Protocol)
```bash
# Start MCP server for Claude Code integration
memory-bank server

# Claude Code can now access your memory bank through MCP protocol
# for storing and retrieving development knowledge during coding sessions
```

### Command Line Usage
```bash
# Initialize project
memory-bank init . --name "My Project"

# Create memories
memory-bank memory create --type decision --title "Use PostgreSQL"

# Search knowledge
memory-bank search "database patterns"

# Track development sessions
memory-bank session start "Add authentication"
memory-bank session log "JWT middleware implemented"
memory-bank session complete "Auth system ready"
```

### Automated Workflows
```bash
# Integrate with CI/CD
memory-bank memory create --type documentation \
  --title "Deployment notes" \
  --content "$(cat deployment_notes.md)"

# Capture test failures
if ! go test ./...; then
  memory-bank memory create --type error_solution \
    --title "Test failure: $(date)" \
    --content "$(go test ./... 2>&1)"
fi
```

## External Dependencies (Optional)

Memory Bank works out-of-the-box with mock providers, but can be enhanced with:

### Ollama (Local Embeddings)
- **Purpose**: Generate vector embeddings locally for semantic search
- **Setup**: `ollama pull nomic-embed-text`
- **Benefit**: Enhanced search accuracy with privacy-first local processing

### ChromaDB (Vector Database)
- **Purpose**: High-performance vector storage and similarity search
- **Setup**: `docker run -p 8000:8000 chromadb/chroma`
- **Benefit**: Improved search performance and advanced vector operations

## Performance Optimizations

Memory Bank includes several performance optimizations:

- **Connection Pooling**: Optimized HTTP clients for external services
- **Concurrent Processing**: Worker pools for batch embedding generation
- **Caching**: LRU cache for duplicate content with 60-80% hit rates
- **Batch Operations**: Bulk database operations for improved throughput
- **Lightweight Queries**: Metadata-only queries for performance-sensitive operations

## Getting Help

### Documentation
- **Start Here**: [Getting Started Guide](guides/getting-started.md)
- **See Examples**: [Examples and Use Cases](guides/examples.md)
- **API Reference**: [MCP Methods](api/mcp-methods.md) and [CLI Commands](api/cli-commands.md)

### Command Line Help
```bash
# General help
memory-bank --help

# Command-specific help
memory-bank memory --help
memory-bank session --help

# Configuration help
memory-bank config --help
```

### Troubleshooting
- **Database Issues**: Check file permissions and available disk space
- **Embedding Problems**: Verify Ollama installation and model availability
- **Search Issues**: Try lower similarity thresholds or different query terms
- **Performance**: Monitor logs for slow queries and optimize as needed

## Contributing

Memory Bank is designed to grow with your needs. Key areas for contribution:

- **Memory Types**: Suggest new types for different knowledge categories
- **Search Improvements**: Enhanced filtering and relevance scoring
- **Integration**: New MCP methods or CLI commands
- **Performance**: Optimization for large knowledge bases
- **Documentation**: Examples, tutorials, and best practices

## Next Steps

1. **[Get Started](guides/getting-started.md)** - Set up Memory Bank for your first project
2. **[View Examples](guides/examples.md)** - See real-world usage patterns
3. **[API Reference](api/)** - Explore all available features
4. **[Development](../CLAUDE.md)** - Learn about the architecture and contribute

---

Memory Bank transforms how you capture and retrieve development knowledge. Start building your semantic knowledge base today!