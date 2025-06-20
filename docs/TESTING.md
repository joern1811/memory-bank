# Memory Bank Testing Guide

This document provides comprehensive guidance for testing the Memory Bank application, including setup, test types, and best practices.

## Table of Contents

- [Overview](#overview)
- [Test Structure](#test-structure)
- [Running Tests](#running-tests)
- [Test Types](#test-types)
- [Integration Testing](#integration-testing)
- [Test Utilities](#test-utilities)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

Memory Bank uses a comprehensive testing strategy with multiple test types:

- **Unit Tests**: Fast, isolated tests using mocks
- **Integration Tests**: Tests with real external services (Ollama, ChromaDB)
- **Benchmark Tests**: Performance testing
- **End-to-End Tests**: Complete workflow testing

### Test Coverage

The test suite covers:
- ✅ Domain layer (entities, value objects)
- ✅ Application layer (services, use cases)
- ✅ Infrastructure layer (repositories, external services)
- ✅ Integration scenarios (real service interactions)

## Test Structure

```
internal/
├── testutil/              # Shared test utilities
│   ├── mocks.go          # Mock implementations
│   ├── fixtures.go       # Test data builders
│   ├── database.go       # Database test helpers
│   ├── assertions.go     # Custom assertions
│   └── integration.go    # Integration test helpers
├── domain/
│   └── *_test.go         # Domain layer unit tests
├── app/
│   └── *_test.go         # Application layer unit tests
└── infra/
    ├── *_test.go         # Infrastructure unit tests
    ├── integration_test.go      # Real service integration tests
    └── enhanced_integration_test.go  # Enhanced scenarios
```

## Running Tests

### Unit Tests (Default)

```bash
# Run all unit tests (fast, no external dependencies)
go test ./...

# Run with verbose output
go test -v ./...

# Run in short mode (skips long-running tests)
go test -short ./...

# Run specific package
go test ./internal/app

# Run with coverage
go test -cover ./...
```

### Integration Tests

Integration tests require build tags and external services:

```bash
# Run integration tests (requires Ollama + ChromaDB)
go test -tags=integration ./...

# Run specific integration test
go test -tags=integration ./internal/infra -v

# Run with timeout for long-running tests
go test -tags=integration -timeout=10m ./...
```

### Benchmark Tests

```bash
# Run benchmark tests
go test -bench=. ./...

# Run specific benchmarks
go test -bench=BenchmarkOllamaProvider ./internal/infra/embedding

# Benchmark with memory allocation stats
go test -bench=. -benchmem ./...
```

## Test Types

### 1. Unit Tests

**Purpose**: Test individual components in isolation
**Speed**: Very fast (< 2 seconds total)
**Dependencies**: Mocks only

**Example**:
```go
func TestMemoryService_CreateMemory(t *testing.T) {
    // Setup mocks
    memoryRepo := testutil.NewMockMemoryRepository(logger)
    embeddingProvider := testutil.NewMockEmbeddingProvider(768, logger)
    vectorStore := testutil.NewMockVectorStore(logger)
    
    // Test service
    service := app.NewMemoryService(memoryRepo, embeddingProvider, vectorStore, logger)
    
    // Execute and verify
    memory, err := service.CreateMemory(ctx, request)
    require.NoError(t, err)
    testutil.AssertMemoryEqual(t, expected, memory)
}
```

### 2. Integration Tests

**Purpose**: Test with real external services
**Speed**: Slower (30s - 5min)
**Dependencies**: Ollama, ChromaDB
**Build Tag**: `// +build integration`

**Example**:
```go
// +build integration

func TestOllamaIntegration(t *testing.T) {
    testutil.SkipIfShort(t, "")
    
    config := testutil.DefaultIntegrationTestConfig()
    config.UseRealServices = true
    
    provider := testutil.SetupEmbeddingProvider(t, config)
    
    embedding, err := provider.GenerateEmbedding(ctx, "test text")
    require.NoError(t, err)
    testutil.AssertEmbeddingVector(t, embedding, 768)
}
```

### 3. End-to-End Tests

**Purpose**: Test complete workflows
**Location**: `enhanced_integration_test.go`
**Features**: Memory lifecycle, concurrent operations, performance

## Integration Testing

### Prerequisites

#### Option 1: Native Setup

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull embedding model
ollama pull nomic-embed-text

# Start ChromaDB (using uvx - no installation needed)
uvx --from "chromadb[server]" chroma run --host localhost --port 8000 --path ./chromadb_data &
```

#### Option 2: Docker Setup

```bash
# Install Ollama (native)
curl -fsSL https://ollama.com/install.sh | sh
ollama pull nomic-embed-text

# Start ChromaDB (Docker)
docker run -p 8000:8000 -v ./chromadb_data:/chroma/chroma chromadb/chroma
```

### Environment Variables

```bash
export MEMORY_BANK_OLLAMA_BASE_URL="http://localhost:11434"
export MEMORY_BANK_OLLAMA_MODEL="nomic-embed-text"
export MEMORY_BANK_CHROMADB_BASE_URL="http://localhost:8000"
export MEMORY_BANK_CHROMADB_COLLECTION="test"
export MEMORY_BANK_LOG_LEVEL="debug"  # For debugging
```

### Running Integration Tests

```bash
# Check service availability first
go test -tags=integration ./internal/infra -run TestIntegrationEnvironment -v

# Run all integration tests
go test -tags=integration ./... -v

# Run specific integration scenarios
go test -tags=integration ./internal/infra -run TestEndToEndMemoryOperations -v
go test -tags=integration ./internal/app -run TestMemorySearchDebug -v
```

## Test Utilities

### Mock Implementations

Located in `internal/testutil/mocks.go`:

- `MockMemoryRepository`: Thread-safe in-memory repository
- `MockProjectRepository`: Project management mock
- `MockSessionRepository`: Session tracking mock
- `MockEmbeddingProvider`: Deterministic embedding generation
- `MockVectorStore`: In-memory vector search

### Test Data Builders

Located in `internal/testutil/fixtures.go`:

```go
// Create test memory with builder pattern
memory := testutil.NewTestMemory().
    WithTitle("Custom Title").
    WithType(domain.MemoryTypeDecision).
    WithTags("tag1", "tag2").
    Build()

// Create multiple test memories
memories := testutil.CreateTestMemories(5, projectID)
```

### Database Helpers

Located in `internal/testutil/database.go`:

```go
// Setup test database with cleanup
db := testutil.SetupTestDatabase(t, nil)

// Setup all repositories
memoryRepo, projectRepo, sessionRepo := testutil.SetupAllTestRepositories(t, nil)

// Populate with test data
projectID, sessionID := testutil.PopulateTestData(t, ctx, memoryRepo, projectRepo, sessionRepo)
```

### Custom Assertions

Located in `internal/testutil/assertions.go`:

```go
// Assert memory equality
testutil.AssertMemoryEqual(t, expected, actual)

// Assert existence in repository
memory := testutil.AssertMemoryExists(t, ctx, repo, memoryID)

// Assert search results
testutil.AssertSearchResults(t, results, 1)

// Assert service health
testutil.AssertServiceHealthy(t, ctx, embeddingProvider)
```

### Integration Helpers

Located in `internal/testutil/integration.go`:

```go
// Setup with automatic fallback to mocks
config := testutil.DefaultIntegrationTestConfig()
embeddingProvider := testutil.SetupEmbeddingProvider(t, config)
vectorStore := testutil.SetupVectorStore(t, config)

// Use integration test runner
runner := testutil.NewIntegrationTestRunner(t, config)
runner.Run(t, "TestName", func(t *testing.T, ctx context.Context, embedding ports.EmbeddingProvider, vector ports.VectorStore) {
    // Test implementation
})
```

## Best Practices

### Test Organization

1. **Use build tags** for integration tests:
   ```go
   // +build integration
   ```

2. **Keep unit tests fast** (< 2 seconds total execution)

3. **Use meaningful test names**:
   ```go
   func TestMemoryService_CreateMemory_WithValidInput_ShouldSucceed(t *testing.T)
   func TestMemoryService_CreateMemory_WithDuplicateID_ShouldReturnError(t *testing.T)
   ```

4. **Group related tests** with subtests:
   ```go
   t.Run("CreateMemory", func(t *testing.T) {
       t.Run("ValidInput", func(t *testing.T) { /* ... */ })
       t.Run("InvalidInput", func(t *testing.T) { /* ... */ })
   })
   ```

### Test Data

1. **Use builders** for complex test data:
   ```go
   memory := testutil.NewTestMemory().WithTitle("Custom").Build()
   ```

2. **Create unique identifiers** to avoid collisions:
   ```go
   id := fmt.Sprintf("test_%d", time.Now().UnixNano())
   ```

3. **Clean up resources** in tests:
   ```go
   t.Cleanup(func() {
       repo.Delete(ctx, memory.ID)
   })
   ```

### Mocking

1. **Use mocks for unit tests**, real services for integration
2. **Make mocks thread-safe** with mutexes
3. **Implement realistic behavior** in mocks
4. **Verify mock interactions** when needed

### Error Testing

1. **Test both success and failure paths**
2. **Use specific error assertions**:
   ```go
   assert.Equal(t, ports.ErrMemoryNotFound, err)
   ```
3. **Test edge cases** (nil inputs, empty strings, etc.)

## Troubleshooting

### Common Issues

#### "Tests are hanging"
- **Cause**: Integration tests trying to connect to unavailable services
- **Solution**: Use build tags to separate integration tests
  ```bash
  go test ./...  # Only unit tests
  go test -tags=integration ./...  # Include integration tests
  ```

#### "UNIQUE constraint failed"
- **Cause**: Deterministic ID generation causing collisions
- **Solution**: Use unique identifiers in test data
  ```go
  id := fmt.Sprintf("test_%d_%d", time.Now().Unix(), time.Now().UnixNano())
  ```

#### "Context deadline exceeded"
- **Cause**: Long-running operations in tests
- **Solution**: Increase timeout or use shorter test data
  ```go
  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
  ```

#### "ChromaDB connection failed"
- **Check**: Is ChromaDB running on correct port?
  ```bash
  curl http://localhost:8000/api/v1/heartbeat
  ```
- **Solution**: Start ChromaDB or run tests with mocks only

#### "Ollama connection failed"  
- **Check**: Is Ollama running and model available?
  ```bash
  curl http://localhost:11434/api/version
  ollama list
  ```
- **Solution**: Start Ollama, pull model, or use mock provider

### Debugging Tests

1. **Enable debug logging**:
   ```bash
   export MEMORY_BANK_LOG_LEVEL=debug
   ```

2. **Run specific test with verbose output**:
   ```bash
   go test -v ./internal/app -run TestMemoryService_CreateMemory
   ```

3. **Use test debugger** in IDE or add print statements

4. **Check integration test environment**:
   ```bash
   go test -tags=integration ./internal/infra -run TestIntegrationEnvironment -v
   ```

### Performance Issues

1. **Profile test execution**:
   ```bash
   go test -cpuprofile=cpu.prof -memprofile=mem.prof ./...
   ```

2. **Run benchmarks** to identify bottlenecks:
   ```bash
   go test -bench=. -benchmem ./...
   ```

3. **Use in-memory databases** for faster tests
4. **Parallelize independent tests**:
   ```go
   t.Parallel()
   ```

## Test Metrics

### Current Coverage

- **Domain Layer**: 100% core functionality
- **Application Layer**: Full CRUD + search operations
- **Infrastructure Layer**: Database, embedding, vector operations
- **Integration Layer**: Real service interactions

### Performance Targets

- **Unit Tests**: < 2 seconds total
- **Integration Tests**: < 5 minutes total
- **Memory Usage**: < 100MB during test execution
- **Embedding Generation**: < 30 seconds per request (real Ollama)

### Quality Gates

All tests must:
- ✅ Pass consistently
- ✅ Clean up resources
- ✅ Use appropriate timeouts
- ✅ Have meaningful assertions
- ✅ Follow naming conventions