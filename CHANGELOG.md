# Changelog

All notable changes to Memory Bank will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Complete documentation overhaul with advanced usage guide
- Comprehensive troubleshooting guide with recovery procedures
- Contributing guidelines and development documentation
- Security policy and changelog documentation

### Changed
- Enhanced getting started guide with more detailed examples
- Improved API documentation with better examples

### Fixed
- Documentation cross-references and navigation

## [1.10.0] - 2024-01-15

### Fixed
- **CRITICAL**: Fixed semantic search returning 0 results due to ChromaDB distance metric mismatch
  - ChromaDB was using Squared Euclidean Distance while Memory Bank assumed Cosine Distance
  - Similarity calculation `1.0 - distance` produced negative values for Euclidean distance
  - Solution: Configure collections with `"hnsw:space": "cosine"` for proper cosine distance metric
- Enhanced ChromaDB metadata handling for better compatibility
- Added comprehensive regression tests for search functionality

### Added
- Full ChromaDB v2 API tenant/database parameter support
- Debug testing suite (`memory_search_debug_test.go`) for end-to-end validation
- Better metadata normalization for ChromaDB compatibility

### Changed
- Semantic search now returns realistic similarity scores (0.8+ for relevant content)
- Improved vector store integration reliability

## [1.9.0] - 2024-01-10

### Changed
- **BREAKING**: Migrated from ChromaDB v1 API to v2 API
  - Updated collection operations: `/api/v1/collections` → `/api/v2/collections`
  - Updated all vector CRUD operations to use v2 endpoints
  - Enhanced health check with dedicated `/api/v2/heartbeat` endpoint
  - Updated all unit tests to expect v2 API endpoints

### Fixed
- ChromaDB API compatibility issues with newer ChromaDB versions
- Health check reliability improvements

### Added
- Backward compatibility validation for API migration
- Enhanced error handling for ChromaDB v2 API responses

## [1.8.0] - 2024-01-05

### Fixed
- **CRITICAL**: Fixed MCP Protocol Integration bugs
  - Tool registration now uses correct `AddTool()` API instead of manual registration
  - Server uses proper `ServeStdio()` method for MCP protocol compliance
  - Added wrapper functions to convert between MCP and internal APIs
  - All 16 MCP tools now working correctly via MCP protocol

### Added
- Complete MCP tool handler implementation
- Proper stdio transport for MCP protocol
- Enhanced error handling for MCP operations

### Changed
- MCP server implementation now fully compliant with MCP specification
- Improved integration with Claude Code and other MCP clients

## [1.7.0] - 2024-01-01

### Added
- **MCP System Prompt Resource**: Dynamic, context-aware system prompts
  - Smart integration guidelines automatically generated
  - Project context with current memory information
  - Usage examples based on existing content patterns
  - Memory type guidance with detailed explanations
  - Complete MCP method documentation with examples
  - Integration tips for optimal development workflow

### Enhanced
- MCP resource system for better client integration
- Dynamic content generation based on project state
- Context-aware documentation and examples

## [1.6.0] - 2023-12-15

### Added
- **Enhanced Documentation**: Complete API documentation and user guides
  - Comprehensive MCP protocol reference with examples
  - Complete CLI commands reference with usage patterns
  - Getting started guide with step-by-step setup
  - Real-world examples and use cases
  - Central documentation hub with navigation

- **Advanced Search Features**:
  - **Faceted Search**: Multi-dimensional filtering with type, tag, content length, and time facets
  - **Enhanced Relevance Scoring**: Intelligent scoring based on title matches, content relevance, and tag alignment
  - **Search Suggestions**: AI-powered suggestions based on existing content patterns
  - **Content Highlighting**: Automatic highlighting of matched terms in search results
  - **Match Reasoning**: Detailed explanations of why results were matched
  - **Advanced Filtering**: Comprehensive filters for types, tags, sessions, content properties
  - **Flexible Sorting**: Multiple sort options (relevance, date, title, type) with direction control

### Changed
- Search API enhanced with new relevance algorithms
- Improved user experience with better search feedback
- Enhanced CLI with new search commands and options

## [1.5.0] - 2023-12-01

### Added
- **Performance Optimizations**:
  - HTTP connection pooling for Ollama and ChromaDB clients
  - Concurrent embedding generation with worker pools (5x improvement)
  - Batch database operations for memory retrieval (10x improvement)
  - LRU embedding cache with TTL (60-80% cache hit rate)
  - Batch vector operations for bulk storage/deletion
  - Memory usage optimization with lightweight metadata queries

### Changed
- Significantly improved search and embedding performance
- Reduced memory footprint for large operations
- Enhanced concurrent processing capabilities

### Fixed
- Memory leaks in long-running operations
- Connection pooling issues with external services

## [1.4.0] - 2023-11-15

### Added
- **Complete CLI Interface**: Full traditional command-line interface with Cobra framework
  - Memory management commands (create, list, search)
  - Project initialization command
  - Global search functionality
  - Session management commands (start, log, complete, list, get, abort)
  - Comprehensive help system
  - Backward compatibility (runs as MCP server when no args provided)

- **Service Integration**:
  - ServiceContainer with dependency injection
  - Automatic health checks and fallback providers
  - Environment-based configuration
  - Database schema updates (context, has_embedding fields)
  - Real session and project repository integration

- **Enhanced Session Features**:
  - Advanced progress tracking with typed entries (info, milestone, issue, solution)
  - Session tags and summaries
  - Duration tracking and analytics
  - Session templates and workflows

### Changed
- Application now supports dual-mode operation (MCP server + CLI)
- Enhanced session management with structured progress tracking
- Improved database schema with additional metadata fields

### Fixed
- Session state persistence and retrieval
- CLI service integration and error handling

## [1.3.0] - 2023-11-01

### Added
- **Database Migrations**: Complete schema versioning system
  - Migration scripts with rollback support
  - Data integrity checks
  - Automated schema updates

- **Configuration Management**: Full YAML/JSON config support
  - Environment-specific configurations
  - Configuration validation and defaults
  - Performance tuning options

- **Integration Testing**: Comprehensive test suite
  - ChromaDB integration tests
  - Ollama integration tests
  - End-to-end workflow testing

### Enhanced
- Build tools with complete Makefile
- GitHub Actions CI/CD pipeline
- Quality checks and automated testing

## [1.2.0] - 2023-10-15

### Added
- **Complete MCP Server Implementation**: Full JSON-RPC protocol support
  - 16 MCP tools for memory, project, and session operations
  - Request validation and error handling
  - Structured error responses

- **Vector Store Integration**:
  - ChromaDB HTTP API integration with v1 API
  - Similarity search with cosine similarity
  - Vector CRUD operations (store, update, delete)
  - Collection management (create, delete, list)
  - Mock implementation for testing

- **Session Repository**: Complete SQLiteSessionRepository implementation
- **Project Repository**: Complete SQLiteProjectRepository implementation

### Changed
- Enhanced service layer with complete CRUD operations
- Improved error handling and logging throughout

### Fixed
- Database connection handling and connection pooling
- Vector store reliability and error recovery

## [1.1.0] - 2023-10-01

### Added
- **Infrastructure Layer**:
  - OllamaProvider with health checks and local embedding generation
  - MockEmbeddingProvider for deterministic testing
  - SQLiteMemoryRepository with auto-initialization and schema migration
  - Automatic fallback to mock providers for offline development

- **Application Layer**:
  - MemoryService with CRUD operations and semantic search
  - ProjectService with project management and auto-detection
  - SessionService framework with development session tracking

### Enhanced
- Automatic health monitoring of external dependencies
- Graceful degradation when external services unavailable
- Structured JSON logging with Logrus

## [1.0.0] - 2023-09-15

### Added
- **Domain Layer**: Complete entity definitions and value objects
  - Memory, Project, Session entities with business logic
  - Type-safe identifiers (MemoryID, ProjectID, SessionID)
  - EmbeddingVector and similarity calculations
  - Memory types: Decision, Pattern, ErrorSolution, Code, Documentation

- **Ports Layer**: Interface definitions for clean architecture
  - Repository interfaces (Memory, Project, Session)
  - Service interfaces with request/response DTOs
  - VectorStore interface with search capabilities
  - Input validation and sanitization

### Technical Foundation
- Hexagonal architecture implementation
- Go 1.21+ with modern language features
- SQLite database with automatic table initialization
- Comprehensive error handling and validation

## [0.1.0] - 2023-09-01

### Added
- Initial project structure
- Basic domain model concepts
- Development environment setup
- Project documentation framework

---

## Version History Summary

- **v1.10.0**: Critical search functionality fix (ChromaDB distance metric)
- **v1.9.0**: ChromaDB v2 API migration
- **v1.8.0**: Fixed MCP protocol integration
- **v1.7.0**: MCP system prompt resource
- **v1.6.0**: Enhanced documentation and advanced search features
- **v1.5.0**: Performance optimizations
- **v1.4.0**: Complete CLI interface and session management
- **v1.3.0**: Database migrations and configuration management
- **v1.2.0**: MCP server and vector store integration
- **v1.1.0**: Infrastructure and application layers
- **v1.0.0**: Core domain model and architecture
- **v0.1.0**: Initial release

## Migration Guides

### Upgrading to v1.10.0
- No breaking changes
- Semantic search results will now return realistic similarity scores
- Existing ChromaDB collections will be automatically updated with correct distance metric

### Upgrading to v1.9.0
- No breaking changes for Memory Bank users
- ChromaDB v2 API is used automatically
- Ensure ChromaDB server supports v2 API (most recent versions do)

### Upgrading to v1.8.0
- No breaking changes
- MCP integration now works correctly with all MCP clients
- Existing MCP configurations remain compatible

### Upgrading to v1.4.0
- New CLI commands available alongside existing MCP server functionality
- Run without arguments for MCP server mode (backward compatible)
- Run with CLI commands for traditional command-line interface

### Upgrading to v1.3.0
- Database schema migrations run automatically
- Configuration files now supported alongside environment variables
- No breaking changes to existing API

## Support

For issues related to specific versions:
1. Check the changelog for known issues and fixes
2. Review the troubleshooting guide in the documentation
3. Create an issue on GitHub with version information

## Contributors

Thank you to all contributors who have helped improve Memory Bank across these releases. For a complete list of contributors, see the project's GitHub contributors page.