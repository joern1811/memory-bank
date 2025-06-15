# Memory Bank

A semantic memory management system for Claude Code using hexagonal architecture.

## Architecture

```
memory-bank/
├── cmd/           # Main applications (CLI, MCP Server)
├── internal/      
│   ├── domain/    # Business logic, entities, value objects
│   ├── app/       # Application services, use cases
│   ├── infra/     # Infrastructure (DB, embeddings, etc.)
│   └── ports/     # Interfaces for external communication
└── pkg/           # Public API
```

## Hexagonal Architecture Layers

### Domain Layer (Core Business Logic)
- **Entities**: Memory, Decision, Pattern, ErrorSolution, Session, Project
- **Value Objects**: MemoryID, EmbeddingVector, Tags, etc.
- **Domain Services**: Pure business logic

### Ports (Interfaces)
- **Primary Ports**: Use cases that the outside world calls
- **Secondary Ports**: Interfaces for external dependencies (DB, embeddings, etc.)

### Application Layer
- **Use Cases**: Orchestrate domain objects and call secondary ports
- **Application Services**: Coordinate between domain and infrastructure

### Infrastructure Layer
- **Adapters**: Concrete implementations of secondary ports
- **Embedding Providers**: Ollama, OpenAI (pluggable)
- **Vector Stores**: ChromaDB, Qdrant (pluggable)
- **Databases**: SQLite, PostgreSQL (pluggable)

### External Layer
- **MCP Server**: Model Context Protocol interface
- **CLI**: Command-line interface
- **HTTP API**: REST API (future)

## Getting Started

```bash
go mod tidy
go build ./cmd/memory-bank
./memory-bank --help
```

## Dependencies

- **MCP**: github.com/mark3labs/mcp-go
- **Database**: SQLite (github.com/mattn/go-sqlite3)
- **Embeddings**: Ollama API
- **Vector Store**: ChromaDB
- **CLI**: Cobra + Viper
- **Testing**: Testify + Testcontainers
