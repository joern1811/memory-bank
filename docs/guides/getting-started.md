# Getting Started with Memory Bank

Memory Bank is a semantic memory management system designed to help developers capture, organize, and retrieve development knowledge. This guide will walk you through setting up and using Memory Bank for your first project.

## What is Memory Bank?

Memory Bank helps you:
- **Capture decisions**: Store architectural decisions with context and rationale
- **Document patterns**: Save reusable code patterns and design solutions
- **Track solutions**: Remember how you solved specific errors or problems
- **Manage sessions**: Track development progress with structured logging
- **Search semantically**: Find relevant knowledge using natural language queries

## Prerequisites

- **Go 1.21+**: Required for building from source
- **Ollama** (optional): For local embedding generation
- **ChromaDB** (optional): For enhanced vector search performance

Memory Bank works out-of-the-box with mock providers, so external dependencies are optional.

## Installation

### Option 1: Homebrew (Recommended)

```bash
# Add the tap and install
brew tap joern1811/tap
brew install --cask joern1811/tap/memory-bank

# Verify installation
memory-bank --help
```

### Option 2: Pre-built Binaries

Download the latest binary for your platform from [GitHub Releases](https://github.com/joern1811/memory-bank/releases).

### Option 3: Build from Source

```bash
# Clone the repository
git clone https://github.com/joern1811/memory-bank.git
cd memory-bank

# Build the application
go build ./cmd/memory-bank

# Make it executable and add to PATH (optional)
chmod +x memory-bank
sudo mv memory-bank /usr/local/bin/
```

### Option 4: Direct Usage (Development)

```bash
# Run directly with Go
go run ./cmd/memory-bank [commands...]
```

## Quick Start (5 Minutes)

Let's set up Memory Bank for a sample project and create your first memories.

### Step 1: Initialize Your Project

```bash
# Navigate to your project directory
cd /path/to/your/project

# Initialize Memory Bank for this project
memory-bank init . --name "My Web API" --description "RESTful API with authentication"
```

Expected output:
```
✓ Project initialized successfully
  ID: proj_abc123
  Name: My Web API
  Path: /path/to/your/project
  Database: ./memory_bank.db
```

### Step 2: Create Your First Memory

Let's capture an architectural decision:

```bash
memory-bank memory create \
  --type decision \
  --title "Use PostgreSQL for primary database" \
  --content "After evaluating PostgreSQL, MySQL, and MongoDB, we chose PostgreSQL for its ACID compliance, excellent JSON support, and mature ecosystem. It fits our requirements for both relational and document data." \
  --tags "database,architecture,postgresql" \
  --project "My Web API"
```

### Step 3: Search Your Knowledge

Now let's search for database-related decisions:

```bash
memory-bank search "database choices" --limit 5
```

You should see your PostgreSQL decision with a high similarity score.

### Step 4: Start a Development Session

Track your development progress:

```bash
memory-bank session start "Implement user authentication" --project "My Web API"
```

Log some progress:

```bash
memory-bank session log "Created user model with validation"
memory-bank session log "JWT middleware implemented" --type milestone
memory-bank session log "Tests passing for login endpoint" --type milestone
```

Complete the session:

```bash
memory-bank session complete "Authentication system fully implemented with tests"
```

### Step 5: Build Your Knowledge Base

Create different types of memories:

```bash
# Document a code pattern
memory-bank memory create \
  --type pattern \
  --title "JWT Authentication Middleware" \
  --content "func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // JWT validation logic
        token := r.Header.Get(\"Authorization\")
        // ... validation code
        next.ServeHTTP(w, r)
    })
}" \
  --tags "go,middleware,jwt,auth"

# Store an error solution
memory-bank memory create \
  --type error_solution \
  --title "Fix CORS preflight requests" \
  --content "Add explicit OPTIONS handler and ensure proper CORS headers are set for preflight requests. Use cors.New() middleware with AllowedOrigins configuration." \
  --tags "cors,http,api,preflight"

# Add documentation
memory-bank memory create \
  --type documentation \
  --title "API Authentication Flow" \
  --content "1. Client sends credentials to /auth/login
2. Server validates and returns JWT token
3. Client includes token in Authorization header
4. Middleware validates token on protected routes" \
  --tags "api,auth,flow,documentation"
```

Congratulations! You now have a semantic knowledge base for your project.

## Understanding Memory Types

Memory Bank supports five types of memories:

### 1. Decision (`decision`)
Architectural and technical decisions with rationale.

**When to use:**
- Technology choices (database, framework, library selection)
- Architecture decisions (microservices vs monolith)
- Design patterns adoption
- Process decisions

**Example:**
```bash
memory-bank memory create --type decision \
  --title "Adopt microservices architecture" \
  --content "Decision to split monolith into microservices for better scalability and team autonomy. Considered complexity trade-offs but benefits outweigh costs for our team size and growth plans."
```

### 2. Pattern (`pattern`)
Reusable code patterns and design solutions.

**When to use:**
- Code templates and boilerplate
- Design patterns implementation
- Common solutions and utilities
- Best practices examples

**Example:**
```bash
memory-bank memory create --type pattern \
  --title "Repository pattern with interfaces" \
  --content "type UserRepository interface { GetByID(id string) (*User, error) }
type userRepo struct { db *sql.DB }
func (r *userRepo) GetByID(id string) (*User, error) { /* implementation */ }"
```

### 3. Error Solution (`error_solution`)
Solutions to specific errors and problems.

**When to use:**
- Bug fixes and debugging solutions
- Configuration issues and resolutions
- Performance problems and optimizations
- Deployment and environment issues

**Example:**
```bash
memory-bank memory create --type error_solution \
  --title "Docker build fails with module not found" \
  --content "Ensure COPY go.mod go.sum ./ comes before COPY . ./ in Dockerfile. Run go mod download before copying source code to leverage Docker layer caching."
```

### 4. Code (`code`)
Code snippets with context and explanations.

**When to use:**
- Utility functions and helpers
- Configuration examples
- Test patterns and examples
- Integration code samples

**Example:**
```bash
memory-bank memory create --type code \
  --title "Database connection with retry logic" \
  --content "func connectWithRetry(dsn string, maxRetries int) (*sql.DB, error) {
    for i := 0; i < maxRetries; i++ {
        db, err := sql.Open(\"postgres\", dsn)
        if err == nil { return db, nil }
        time.Sleep(time.Second * 2)
    }
    return nil, errors.New(\"max retries exceeded\")
}"
```

### 5. Documentation (`documentation`)
Project documentation and explanations.

**When to use:**
- Process documentation
- API documentation
- Architecture overviews
- Setup and deployment guides

**Example:**
```bash
memory-bank memory create --type documentation \
  --title "Deployment checklist" \
  --content "1. Run tests: go test ./...
2. Build: docker build -t app:latest .
3. Push: docker push registry/app:latest
4. Deploy: kubectl apply -f k8s/
5. Verify: kubectl get pods"
```

## Tagging Strategy

Effective tagging improves search and organization:

### Recommended Tag Categories

1. **Technology**: `go`, `javascript`, `python`, `react`, `postgresql`
2. **Domain**: `auth`, `api`, `database`, `frontend`, `backend`
3. **Type**: `middleware`, `util`, `config`, `test`, `deployment`
4. **Status**: `draft`, `reviewed`, `deprecated`, `active`
5. **Project**: `user-service`, `payment-api`, `admin-dashboard`

### Tagging Examples

```bash
# Good tagging
--tags "go,middleware,auth,jwt,security"
--tags "database,postgresql,migration,schema"
--tags "api,rest,validation,error-handling"

# Less effective
--tags "code,stuff,important"
```

## Advanced Features

### Semantic Search

Memory Bank uses vector embeddings for semantic search, meaning you can find relevant information even when your search terms don't exactly match the stored content.

```bash
# These queries might find the same JWT middleware:
memory-bank search "authentication middleware"
memory-bank search "JWT token validation"
memory-bank search "secure API endpoints"
memory-bank search "user verification code"
```

### Session-Based Development

Track your development progress with structured sessions:

```bash
# Start session
memory-bank session start "Add user profiles" --project "my-app"

# Log different types of progress
memory-bank session log "Created profile model" --type info
memory-bank session log "Profile CRUD endpoints complete" --type milestone
memory-bank session log "Validation failing on empty names" --type issue
memory-bank session log "Added validation middleware" --type solution

# View session progress
memory-bank session get [session-id]

# Complete session
memory-bank session complete "User profiles implemented with validation"
```

### Project-Scoped Operations

Most operations can be scoped to specific projects:

```bash
# Search within project
memory-bank memory search "authentication" --project "my-api"

# List project memories
memory-bank memory list --project "my-api" --type decision

# Project-specific sessions  
memory-bank session list --project "my-api" --status active
```

## External Services Setup (Optional)

For enhanced performance, you can set up external services:

### Ollama Setup (Local Embeddings)

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull embedding model
ollama pull nomic-embed-text

# Verify it's running
curl http://localhost:11434/api/tags
```

### ChromaDB Setup (Vector Database)

```bash
# Using Docker
docker run -p 8000:8000 chromadb/chroma

# Verify it's running
curl http://localhost:8000/api/v2/heartbeat
```

Memory Bank will automatically detect and use these services when available, with graceful fallback to mock providers.

## Configuration

### Configuration File

Create `~/.memory-bank/config.yaml`:

```yaml
database:
  path: "./memory_bank.db"

embedding:
  provider: "ollama"  # or "mock"
  ollama:
    base_url: "http://localhost:11434"
    model: "nomic-embed-text"

vector:
  provider: "chromadb"  # or "mock"
  chromadb:
    base_url: "http://localhost:8000"
    collection: "memory_bank"

logging:
  level: "info"
  format: "json"
```

### Environment Variables

```bash
export MEMORY_BANK_DB_PATH="./memory_bank.db"
export OLLAMA_BASE_URL="http://localhost:11434" 
export OLLAMA_MODEL="nomic-embed-text"
export CHROMADB_BASE_URL="http://localhost:8000"
export CHROMADB_COLLECTION="memory_bank"
```

## Integration with Claude Code

Memory Bank can run as an MCP (Model Context Protocol) server for integration with Claude Code:

```bash
# Start MCP server mode
memory-bank mcp serve

# Or simply (default behavior with no arguments)
memory-bank
```

This allows Claude Code to directly access your memory bank for storing and retrieving development knowledge during coding sessions.

## Best Practices

### 1. Consistent Memory Creation

- **Be specific**: Clear, descriptive titles help with search
- **Include context**: Explain why decisions were made, not just what
- **Tag systematically**: Use consistent tagging conventions
- **Update regularly**: Keep memories current as your understanding evolves

### 2. Effective Search

- **Use natural language**: "How to handle database errors" works better than "db error"
- **Adjust thresholds**: Use `--threshold 0.7` for high-precision results
- **Combine filters**: Use `--type` and `--project` to narrow results
- **Try variations**: Different search terms may surface different relevant memories

### 3. Session Management

- **Start sessions for focused work**: Each feature or bug fix gets its own session
- **Log regularly**: Capture progress, issues, and solutions as they happen
- **Use entry types**: Mark milestones, issues, and solutions appropriately
- **Complete with outcomes**: Summarize what was accomplished

### 4. Project Organization

- **One project per codebase**: Keep related memories together
- **Meaningful project names**: Use consistent naming across your projects
- **Regular cleanup**: Archive or delete outdated memories
- **Cross-reference**: Use tags to link related memories across projects

## Troubleshooting

### Common Issues

**Database locked error:**
```bash
# Ensure no other Memory Bank processes are running
ps aux | grep memory-bank
kill [process-id]
```

**Embeddings not working:**
```bash
# Check Ollama status
ollama list
curl http://localhost:11434/api/tags

# Test with mock provider
memory-bank config set embedding.provider mock
```

**Vector search returning no results:**
```bash
# Check ChromaDB connection
curl http://localhost:8000/api/v2/heartbeat

# Lower search threshold
memory-bank search "query" --threshold 0.3
```

**Configuration not loaded:**
```bash
# Check config file location
memory-bank config show

# Initialize default config
memory-bank config init
```

### Getting Help

```bash
# General help
memory-bank --help

# Command-specific help
memory-bank memory --help
memory-bank session --help

# Configuration validation
memory-bank config validate
```

## Next Steps

Now that you have Memory Bank set up:

1. **Build your knowledge base**: Start capturing decisions, patterns, and solutions
2. **Establish workflows**: Integrate memory creation into your development process
3. **Use sessions**: Track your development progress for better project visibility
4. **Refine your search**: Experiment with different queries and thresholds
5. **Set up external services**: Consider Ollama and ChromaDB for enhanced performance

## Examples and Use Cases

Check out the [Examples Guide](examples.md) for detailed use cases and real-world scenarios showing how to effectively use Memory Bank in different development contexts.

For advanced features and configuration options, refer to the [Advanced Usage Guide](advanced-usage.md).