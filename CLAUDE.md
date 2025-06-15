# Memory Bank - Development Documentation

## Project Overview

Memory Bank is a semantic memory management system for Claude Code using hexagonal architecture. It provides intelligent storage and retrieval of development knowledge including decisions, patterns, error solutions, and session context.

## Architecture

### Hexagonal Architecture
- **Domain Layer**: Core business logic (entities, value objects)
- **Application Layer**: Use cases and application services
- **Infrastructure Layer**: External dependencies (database, embeddings, vector store)
- **Ports**: Interfaces between layers
- **Adapters**: Concrete implementations

### Technology Stack
- **Language**: Go 1.21+
- **Database**: SQLite (with potential PostgreSQL support)
- **Embeddings**: Ollama (nomic-embed-text model)
- **Vector Store**: ChromaDB (planned)
- **MCP**: github.com/mark3labs/mcp-go
- **CLI**: Cobra + Viper
- **Logging**: Logrus

## Project Structure

```
memory-bank/
├── cmd/memory-bank/     # Main CLI application
├── internal/
│   ├── domain/          # Business entities and value objects
│   ├── app/             # Application services and use cases
│   ├── infra/           # Infrastructure implementations
│   │   ├── embedding/   # Ollama embedding provider
│   │   ├── database/    # SQLite repositories
│   │   └── vector/      # Vector store implementations (TODO)
│   └── ports/           # Interface definitions
└── pkg/                 # Public API (TODO)
```

## Core Entities

### Memory Types
- `Decision`: Architectural decisions with rationale and options
- `Pattern`: Code/design patterns with implementations
- `ErrorSolution`: Error signatures with solutions
- `Code`: Code snippets with context
- `Documentation`: Project documentation
- `Session`: Development session tracking

### Value Objects
- `MemoryID`, `ProjectID`, `SessionID`: Unique identifiers
- `EmbeddingVector`: Float32 slice for vector embeddings
- `Tags`: String slice for categorization
- `Similarity`: Float32 for similarity scores

## Key Features (Implemented)

### ✅ Domain Layer
- Complete entity definitions (Memory, Project, Session)
- Value objects with business logic
- Type-safe identifiers

### ✅ Application Layer
- MemoryService: CRUD operations + semantic search
- ProjectService: Project management with auto-detection
- SessionService: Development session tracking

### ✅ Infrastructure Layer
- OllamaProvider: Local embedding generation
- SQLiteMemoryRepository: Persistent storage
- MockEmbeddingProvider: Testing support

### ✅ Ports Layer
- Repository interfaces
- Service interfaces
- Request/response DTOs

## Key Features (TODO)

### 🔲 Vector Store Integration
- ChromaDB adapter implementation
- Similarity search functionality
- Vector storage management

### 🔲 MCP Server Implementation
- JSON-RPC handlers
- Claude Code integration
- Request validation

### 🔲 CLI Commands
- Project initialization
- Memory management
- Search functionality
- Session management

### 🔲 Database Migrations
- Schema versioning
- Migration scripts
- Data integrity

### 🔲 Configuration
- YAML/JSON config files
- Environment variables
- Provider selection

## Getting Started

### Prerequisites
```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull embedding model
ollama pull nomic-embed-text
```

### Development Setup
```bash
cd /Users/jdombrowski/git/claude/projects/memory-bank

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Build application
go build ./cmd/memory-bank
```

### Basic Usage (Planned)
```bash
# Initialize project
./memory-bank init /path/to/project

# Start session
./memory-bank session start "Implementing authentication"

# Add decision
./memory-bank memory add decision "Use JWT for authentication" \
  --rationale "Stateless, scalable, industry standard" \
  --options "JWT,Sessions,OAuth2"

# Search memories
./memory-bank search "authentication patterns"

# Complete session
./memory-bank session complete "JWT authentication implemented successfully"
```

## Testing Strategy

### Unit Tests
- Domain logic testing
- Service behavior testing
- Repository testing with mocks

### Integration Tests
- Database integration
- Ollama integration
- End-to-end workflows

### Test Utilities
- Mock implementations
- Test data builders
- Integration test helpers

## Development Guidelines

### Code Quality
- Follow Go conventions
- Use meaningful variable names
- Write comprehensive tests
- Document public APIs

### Error Handling
- Use wrapped errors with context
- Log errors at appropriate levels
- Provide user-friendly error messages
- Handle edge cases gracefully

### Performance Considerations
- Batch embedding operations when possible
- Use connection pooling for database
- Implement caching for frequent queries
- Monitor memory usage for large embeddings

## Security Considerations

### Data Privacy
- All embeddings generated locally (Ollama)
- No external API calls for sensitive data
- Local SQLite storage

### Input Validation
- Sanitize all user inputs
- Validate file paths
- Check embedding dimensions

## Deployment

### Local Development
- SQLite database
- Local Ollama instance
- File-based configuration

### Production (Future)
- PostgreSQL database
- Shared Ollama instance
- Environment-based configuration
- Docker containerization

## Monitoring and Logging

### Logging Levels
- DEBUG: Detailed operation tracing
- INFO: Normal operation events
- WARN: Recoverable issues
- ERROR: Critical failures

### Metrics (Future)
- Embedding generation times
- Search response times
- Database query performance
- Memory usage patterns

## Contributing

### Development Workflow
1. Create feature branch
2. Implement with tests
3. Run full test suite
4. Update documentation
5. Submit for review

### Code Review Checklist
- [ ] Tests cover new functionality
- [ ] Documentation updated
- [ ] Error handling implemented
- [ ] Logging appropriately placed
- [ ] Performance considerations addressed

## API Reference (Planned)

### MCP Methods
- `memory/create`: Create new memory entry
- `memory/search`: Semantic search across memories
- `memory/get`: Retrieve specific memory
- `project/init`: Initialize project memory
- `session/start`: Start development session
- `session/log`: Log session progress

### CLI Commands
- `init`: Initialize project
- `memory`: Memory management
- `search`: Search operations
- `session`: Session management
- `config`: Configuration management

## Known Issues

### Current Limitations
- No vector store implementation yet
- CLI application incomplete
- No MCP server implementation
- Limited configuration options

### Future Improvements
- Batch processing optimization
- Advanced search filters
- Cross-project intelligence
- Performance monitoring
- Configuration hot-reload

## Resources

### Dependencies
- [Ollama](https://ollama.com/): Local LLM and embedding models
- [ChromaDB](https://www.trychroma.com/): Vector database
- [Cobra](https://cobra.dev/): CLI framework
- [Logrus](https://github.com/sirupsen/logrus): Structured logging

### References
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
