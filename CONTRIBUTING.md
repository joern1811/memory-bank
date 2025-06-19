# Contributing to Memory Bank

Thank you for your interest in contributing to Memory Bank! This document provides guidelines and information for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Environment](#development-environment)
- [Contributing Guidelines](#contributing-guidelines)
- [Development Workflow](#development-workflow)
- [Code Style and Standards](#code-style-and-standards)
- [Testing](#testing)
- [Documentation](#documentation)
- [Pull Request Process](#pull-request-process)
- [Release Process](#release-process)
- [Community](#community)

## Code of Conduct

This project adheres to a code of conduct that promotes a welcoming and inclusive environment. By participating, you are expected to uphold this code.

### Our Standards

- **Be respectful**: Treat all community members with respect and courtesy
- **Be inclusive**: Welcome newcomers and help them get involved
- **Be constructive**: Provide helpful feedback and suggestions
- **Be collaborative**: Work together to improve the project
- **Be patient**: Understand that everyone has different experience levels

### Unacceptable Behavior

- Harassment, discrimination, or hostile behavior
- Personal attacks or inflammatory comments
- Spam, trolling, or disruptive behavior
- Sharing private information without consent
- Any behavior that violates applicable laws

## Getting Started

### Prerequisites

- **Go 1.21+**: Required for building and development
- **Git**: For version control
- **Make**: For build automation (optional but recommended)
- **Docker**: For containerized testing (optional)

### Development Dependencies

```bash
# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install golang.org/x/lint/golint@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install gotest.tools/gotestsum@latest

# Install external services for testing
curl -fsSL https://ollama.com/install.sh | sh
docker pull chromadb/chroma
```

## Development Environment

### Setting Up Your Environment

1. **Fork and Clone**:
   ```bash
   git clone https://github.com/yourusername/memory-bank.git
   cd memory-bank
   ```

2. **Set Up Dependencies**:
   ```bash
   go mod download
   go mod tidy
   ```

3. **Verify Installation**:
   ```bash
   make build
   ./memory-bank --help
   ```

4. **Run Tests**:
   ```bash
   make test
   make test-integration  # Requires external services
   ```

### Environment Variables

```bash
# Development configuration
export MEMORY_BANK_LOG_LEVEL=debug
export MEMORY_BANK_DB_PATH="./dev_memory_bank.db"
export OLLAMA_BASE_URL="http://localhost:11434"
export CHROMADB_BASE_URL="http://localhost:8000"

# Testing configuration
export MEMORY_BANK_TEST_DB=":memory:"
export MEMORY_BANK_TEST_INTEGRATION=true
```

### Development Tools

The project includes several development tools:

```bash
# Build and test
make build          # Build binary
make test           # Run unit tests
make test-race      # Run tests with race detection
make test-coverage  # Generate coverage report

# Code quality
make lint           # Run linters
make fmt            # Format code
make vet            # Run go vet

# Development server
make dev            # Start development server with auto-reload
make debug          # Start with debug logging enabled
```

## Contributing Guidelines

### Types of Contributions

We welcome several types of contributions:

1. **Bug Reports**: Report issues you encounter
2. **Feature Requests**: Suggest new features or improvements
3. **Code Contributions**: Submit bug fixes or new features
4. **Documentation**: Improve or add documentation
5. **Testing**: Add or improve tests
6. **Performance**: Optimize existing code

### Before Contributing

1. **Search Existing Issues**: Check if your issue or feature request already exists
2. **Read Documentation**: Familiarize yourself with the project architecture and goals
3. **Start Small**: Begin with small contributions to understand the codebase
4. **Discuss Large Changes**: Open an issue to discuss significant changes before implementing

### Issue Guidelines

When creating an issue:

1. **Use a Clear Title**: Describe the issue concisely
2. **Provide Context**: Include relevant background information
3. **Steps to Reproduce**: For bugs, provide clear reproduction steps
4. **Expected Behavior**: Describe what you expected to happen
5. **Environment Details**: Include version, OS, and configuration details

#### Bug Report Template

```markdown
**Bug Description**
A clear description of what the bug is.

**Steps to Reproduce**
1. Go to '...'
2. Click on '....'
3. See error

**Expected Behavior**
A clear description of what you expected to happen.

**Actual Behavior**
What actually happened.

**Environment**
- Memory Bank Version: [e.g., v1.0.0]
- Go Version: [e.g., 1.21.0]
- OS: [e.g., macOS 14.0]
- Architecture: [e.g., arm64]

**Additional Context**
Add any other context about the problem here.
```

#### Feature Request Template

```markdown
**Feature Description**
A clear description of the feature you'd like to see.

**Use Case**
Describe the problem this feature would solve.

**Proposed Solution**
A clear description of what you want to happen.

**Alternatives Considered**
A clear description of alternative solutions you've considered.

**Additional Context**
Add any other context or screenshots about the feature request here.
```

## Development Workflow

### Git Workflow

We use a standard Git workflow:

1. **Create a Branch**: Create a feature branch from `main`
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make Changes**: Implement your changes with appropriate tests
3. **Commit Changes**: Use conventional commit messages
4. **Push Branch**: Push your branch to your fork
5. **Create Pull Request**: Submit a pull request for review

### Commit Message Format

We use [Conventional Commits](https://www.conventionalcommits.org/) format:

```
type(scope): description

body (optional)

footer (optional)
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `perf`: Performance improvements

**Examples**:
```
feat(search): add faceted search functionality

Implements multi-dimensional filtering with real-time facet generation.
Includes support for type, tag, content length, and time period facets.

Closes #123
```

```
fix(database): resolve connection pool exhaustion

The connection pool was not properly releasing connections after failed
queries, leading to pool exhaustion under high load.

Fixes #456
```

### Branch Naming

Use descriptive branch names:
- `feature/add-session-templates`
- `fix/database-connection-leak`
- `docs/api-documentation-update`
- `refactor/simplify-search-interface`

## Code Style and Standards

### Go Code Style

We follow standard Go conventions:

1. **gofmt**: All code must be formatted with `gofmt`
2. **goimports**: Use `goimports` for import organization
3. **golint**: Follow `golint` recommendations
4. **go vet**: Code must pass `go vet` checks

### Specific Guidelines

#### Code Organization

```go
// Package documentation
package main

// Import groups: standard library, third-party, local
import (
    "context"
    "fmt"
    
    "github.com/sirupsen/logrus"
    
    "github.com/your-org/memory-bank/internal/domain"
)

// Constants
const (
    DefaultTimeout = 30 * time.Second
)

// Types
type Service struct {
    // Exported fields first
    Config *Config
    
    // Unexported fields
    logger *logrus.Logger
}

// Constructor functions
func NewService(config *Config) *Service {
    return &Service{
        Config: config,
        logger: logrus.New(),
    }
}

// Methods: receiver, name, parameters, returns
func (s *Service) ProcessMemory(ctx context.Context, memory *domain.Memory) error {
    // Implementation
}
```

#### Error Handling

```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to process memory %s: %w", memory.ID, err)
}

// Use custom error types for business logic
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}
```

#### Testing Patterns

```go
func TestMemoryService_Create(t *testing.T) {
    // Arrange
    service := setupTestService(t)
    memory := &domain.Memory{
        Title:   "Test Memory",
        Content: "Test content",
    }

    // Act
    result, err := service.Create(context.Background(), memory)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, "Test Memory", result.Title)
    assert.NotEmpty(t, result.ID)
}
```

### Architecture Guidelines

#### Hexagonal Architecture

Follow hexagonal architecture principles:

1. **Domain Layer**: Business logic, entities, value objects
2. **Application Layer**: Use cases, application services
3. **Infrastructure Layer**: External dependencies, repositories
4. **Ports**: Interfaces between layers

```go
// Domain entity
type Memory struct {
    ID      MemoryID
    Title   string
    Content string
    Tags    []string
}

// Port interface
type MemoryRepository interface {
    Create(ctx context.Context, memory *Memory) error
    FindByID(ctx context.Context, id MemoryID) (*Memory, error)
}

// Application service
type MemoryService struct {
    repo MemoryRepository
}

// Infrastructure adapter
type SQLiteMemoryRepository struct {
    db *sql.DB
}
```

#### Dependency Injection

Use constructor injection for dependencies:

```go
type ServiceContainer struct {
    MemoryService  *app.MemoryService
    ProjectService *app.ProjectService
}

func NewServiceContainer(
    memoryRepo ports.MemoryRepository,
    projectRepo ports.ProjectRepository,
    embeddingProvider ports.EmbeddingProvider,
) *ServiceContainer {
    return &ServiceContainer{
        MemoryService:  app.NewMemoryService(memoryRepo, embeddingProvider),
        ProjectService: app.NewProjectService(projectRepo),
    }
}
```

## Testing

### Testing Strategy

We use a comprehensive testing approach:

1. **Unit Tests**: Test individual components in isolation
2. **Integration Tests**: Test component interactions
3. **End-to-End Tests**: Test complete workflows
4. **Benchmark Tests**: Performance testing

### Writing Tests

#### Unit Tests

```go
func TestMemoryValidation(t *testing.T) {
    tests := []struct {
        name    string
        memory  *domain.Memory
        wantErr bool
    }{
        {
            name: "valid memory",
            memory: &domain.Memory{
                Title:   "Valid Title",
                Content: "Valid content",
                Type:    domain.MemoryTypeDecision,
            },
            wantErr: false,
        },
        {
            name: "empty title",
            memory: &domain.Memory{
                Content: "Valid content",
                Type:    domain.MemoryTypeDecision,
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.memory.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

#### Integration Tests

```go
func TestMemoryService_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // Create service with real dependencies
    service := setupRealService(t, db)

    // Test complete workflow
    memory := &domain.Memory{
        Title:   "Integration Test",
        Content: "Test content",
    }

    created, err := service.Create(context.Background(), memory)
    require.NoError(t, err)

    found, err := service.FindByID(context.Background(), created.ID)
    require.NoError(t, err)
    assert.Equal(t, created.Title, found.Title)
}
```

#### Benchmark Tests

```go
func BenchmarkSearchPerformance(b *testing.B) {
    service := setupBenchmarkService(b)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.Search(context.Background(), "test query")
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Test Utilities

Use test helpers for common setup:

```go
func setupTestService(t *testing.T) *app.MemoryService {
    repo := &mocks.MockMemoryRepository{}
    embeddingProvider := &mocks.MockEmbeddingProvider{}
    return app.NewMemoryService(repo, embeddingProvider)
}

func createTestMemory(t *testing.T, title string) *domain.Memory {
    memory := &domain.Memory{
        ID:      domain.NewMemoryID(),
        Title:   title,
        Content: "Test content for " + title,
        Type:    domain.MemoryTypeDecision,
        Tags:    []string{"test"},
    }
    require.NoError(t, memory.Validate())
    return memory
}
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run integration tests
make test-integration

# Run specific test
go test -run TestMemoryService_Create ./internal/app

# Run tests with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./...
```

## Documentation

### Documentation Standards

1. **Code Documentation**: All public APIs must have godoc comments
2. **Architecture Documentation**: Document design decisions and patterns
3. **User Documentation**: Provide clear usage examples
4. **API Documentation**: Document all endpoints and parameters

### Writing Documentation

#### Godoc Comments

```go
// MemoryService provides operations for managing memories in the system.
// It handles creation, retrieval, and semantic search of memory entries.
type MemoryService struct {
    repo              ports.MemoryRepository
    embeddingProvider ports.EmbeddingProvider
}

// Create creates a new memory entry in the system.
// It validates the memory, generates embeddings for the content,
// and stores both the memory and its vector representation.
//
// Returns the created memory with generated ID and metadata,
// or an error if validation or storage fails.
func (s *MemoryService) Create(ctx context.Context, memory *domain.Memory) (*domain.Memory, error) {
    // Implementation
}
```

#### Architecture Decision Records

Document significant decisions in `docs/adr/`:

```markdown
# ADR-001: Use Hexagonal Architecture

## Status
Accepted

## Context
Need clean separation between business logic and external dependencies.

## Decision
Implement hexagonal architecture with ports and adapters pattern.

## Consequences
- Clear separation of concerns
- Improved testability
- Easier to swap implementations
- Additional complexity for simple operations
```

### Documentation Updates

When making changes:

1. **Update Godoc**: Keep public API documentation current
2. **Update User Docs**: Reflect changes in user-facing documentation
3. **Update Examples**: Ensure examples work with changes
4. **Update Architecture Docs**: Document architectural changes

## Pull Request Process

### Before Submitting

1. **Run Tests**: Ensure all tests pass
   ```bash
   make test
   make test-integration
   make lint
   ```

2. **Update Documentation**: Update relevant documentation
3. **Write Tests**: Add tests for new functionality
4. **Check Coverage**: Maintain or improve test coverage

### Pull Request Template

```markdown
## Description
Brief description of changes made.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## How Has This Been Tested?
Describe the tests that you ran to verify your changes.

## Checklist
- [ ] My code follows the project's style guidelines
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
```

### Review Process

1. **Automated Checks**: All CI checks must pass
2. **Code Review**: At least one maintainer review required
3. **Testing**: Changes must include appropriate tests
4. **Documentation**: Documentation must be updated if needed

### Merge Requirements

- All CI checks passing
- At least one approving review from maintainer
- No unresolved conversations
- Branch up to date with main

## Release Process

### Version Management

We use [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Workflow

1. **Update Version**: Update version in relevant files
2. **Update Changelog**: Document changes in CHANGELOG.md
3. **Create Tag**: Create and push version tag
4. **Create Release**: Create GitHub release with notes
5. **Publish Artifacts**: Publish binaries and documentation

### Release Checklist

- [ ] All tests pass
- [ ] Documentation updated
- [ ] Changelog updated
- [ ] Version numbers updated
- [ ] Migration scripts ready (if needed)
- [ ] Backward compatibility verified
- [ ] Performance regression tests pass

## Community

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and discussions
- **Pull Requests**: Code contributions and reviews

### Getting Help

1. **Documentation**: Check existing documentation first
2. **Search Issues**: Look for similar issues or questions
3. **Create Issue**: Create a new issue with detailed information
4. **Be Patient**: Maintainers are volunteers and may take time to respond

### Becoming a Maintainer

Regular contributors who demonstrate commitment to the project may be invited to become maintainers. Maintainers have additional responsibilities:

- **Code Review**: Review and approve pull requests
- **Issue Triage**: Help categorize and prioritize issues
- **Release Management**: Participate in release planning and execution
- **Community Support**: Help answer questions and guide new contributors

### Recognition

We value all contributions and recognize contributors in:
- Release notes
- Contributors file
- Project documentation

Thank you for contributing to Memory Bank! Your efforts help make this project better for everyone.