# Workflows and Examples Guide

Complete guide covering workflows, examples, and advanced usage patterns for Memory Bank. This includes the new task management and documentation sync features.

## Table of Contents

1. [Development Workflows](#development-workflows)
2. [Task Management](#task-management)
3. [Documentation Sync](#documentation-sync)
4. [Advanced Search Strategies](#advanced-search-strategies)
5. [Real-World Examples](#real-world-examples)
6. [Integration Patterns](#integration-patterns)

## Development Workflows

### Complete Feature Development Workflow

```bash
# 1. Start development session
memory-bank session start "Implement user authentication system" --project my-api

# 2. Search for existing patterns
memory-bank search "JWT authentication middleware" --type pattern

# 3. Create task for feature development
memory-bank task create \
  --project my-api \
  --title "Implement JWT authentication system" \
  --description "Complete authentication system with login, logout, and middleware" \
  --priority high \
  --estimated-hours 16 \
  --tags auth,security,jwt

# 4. Break down into subtasks
memory-bank task create \
  --project my-api \
  --title "Create JWT middleware" \
  --description "Token validation and refresh logic" \
  --parent-task <parent-task-id> \
  --estimated-hours 4

memory-bank task create \
  --project my-api \
  --title "Implement login endpoints" \
  --description "POST /login and POST /logout endpoints" \
  --parent-task <parent-task-id> \
  --estimated-hours 6

# 5. Document architectural decisions
memory-bank memory create --type decision \
  --title "JWT vs Session Authentication" \
  --content "Chose JWT for stateless authentication. Considered sessions but JWT better for API scalability and mobile apps." \
  --tags auth,jwt,decision \
  --project my-api

# 6. Track progress
memory-bank session log "Created JWT middleware with token validation"
memory-bank task update <task-id> --status in_progress --actual-hours 3

# 7. Store reusable patterns
memory-bank memory create --type pattern \
  --title "JWT Middleware Pattern" \
  --content "Standard JWT middleware with validation, refresh, and error handling" \
  --tags auth,middleware,pattern \
  --project my-api

# 8. Complete tasks and session
memory-bank task update <task-id> --status done --actual-hours 15
memory-bank session complete "Successfully implemented JWT auth system. All tests passing."
```

## Task Management

### Task Creation and Management

```bash
# Create a complex task with dependencies
memory-bank task create \
  --project e-commerce \
  --title "Implement order processing pipeline" \
  --description "Complete order processing from cart to fulfillment" \
  --priority high \
  --due-date "2024-01-15T17:00:00Z" \
  --estimated-hours 40 \
  --assignee "john.doe@company.com" \
  --tags backend,orders,pipeline

# Add dependencies
memory-bank task add-dependency <task-id> <payment-service-task-id>
memory-bank task add-dependency <task-id> <inventory-task-id>

# Create subtasks
memory-bank task create \
  --project e-commerce \
  --title "Payment processing integration" \
  --parent-task <parent-task-id> \
  --estimated-hours 12 \
  --assignee "alice@company.com"

# Track progress
memory-bank task update <task-id> --status in_progress --actual-hours 8
memory-bank task list --project e-commerce --status in_progress --assignee "john.doe"

# Generate reports
memory-bank task statistics --project e-commerce
memory-bank task efficiency-report --project e-commerce
```

### Advanced Task Queries

```bash
# Find overdue tasks
memory-bank task list --is-overdue true --priority high

# Find tasks by date range
memory-bank task list \
  --due-after "2024-01-01T00:00:00Z" \
  --due-before "2024-01-31T23:59:59Z" \
  --sort-by due_date

# Complex filtering
memory-bank task list \
  --project my-api \
  --tags auth,security \
  --status todo,in_progress \
  --assignee "team-lead@company.com" \
  --sort-by priority \
  --sort-order desc
```

## Documentation Sync

### Automatic Documentation Management

```bash
# Analyze code changes for documentation impact
memory-bank doc analyze-changes \
  --changed-files "src/api/auth.go,src/middleware/jwt.go,docs/api.md" \
  --change-context "Added JWT authentication system with middleware"

# Get suggestions for documentation updates
memory-bank doc suggest-updates \
  --change-type api \
  --component authentication

# Create mappings between code and documentation
memory-bank doc create-mapping \
  --code-pattern "src/api/*" \
  --documentation-files "docs/api.md,docs/authentication.md" \
  --change-type api \
  --priority high

# Validate documentation consistency
memory-bank doc validate-consistency \
  --project-path . \
  --focus-areas api,cli,config

# Set up automation (git hooks)
memory-bank doc setup-automation \
  --project-path . \
  --install-hooks true \
  --interactive true
```

### Documentation Workflow Integration

```bash
# Before making API changes
memory-bank doc analyze-changes --changed-files "src/api/users.go"
# Output: Suggests updating docs/api.md and docs/user-management.md

# After implementing changes
memory-bank doc validate-consistency --focus-areas api
# Output: Identifies outdated API documentation

# Store documentation decisions
memory-bank memory create --type documentation \
  --title "API Documentation Strategy" \
  --content "Use OpenAPI spec as source of truth. Auto-generate from code comments." \
  --tags documentation,api,strategy
```

## Advanced Search Strategies

### Faceted Search

```bash
# Multi-dimensional search with facets
memory-bank search faceted "authentication patterns" \
  --project my-api \
  --types decision,pattern,code \
  --tags auth,security,jwt \
  --include-facets true \
  --threshold 0.7 \
  --limit 20
```

### Enhanced Search with Relevance

```bash
# Enhanced search with detailed scoring
memory-bank search enhanced "error handling patterns" \
  --type pattern \
  --tags error,handling \
  --threshold 0.6 \
  --limit 10
```

### Search Suggestions and Discovery

```bash
# Get intelligent search suggestions
memory-bank search-suggestions "auth" --project my-api --limit 5

# Example output:
# - "authentication middleware patterns"
# - "JWT token validation"
# - "OAuth integration examples"
# - "session management decisions"
```

## Real-World Examples

### Microservices Architecture Project

```bash
# Initialize microservices project
memory-bank init . --name "E-commerce Microservices" \
  --description "Event-driven microservices architecture for e-commerce platform"

# Document service communication decisions
memory-bank memory create --type decision \
  --title "Event-driven communication between services" \
  --content "Use Apache Kafka for async communication. Considered REST but chose events for better decoupling and scalability." \
  --tags microservices,kafka,events,communication

# Create service development tasks
memory-bank task create \
  --project e-commerce \
  --title "User Service Implementation" \
  --description "User management microservice with CRUD operations" \
  --estimated-hours 32 \
  --tags user-service,microservice

# Track cross-service patterns
memory-bank memory create --type pattern \
  --title "Service Health Check Pattern" \
  --content "Standard health check endpoint with dependencies status" \
  --tags microservices,health-check,monitoring
```

### Bug Investigation Workflow

```bash
# Start debugging session
memory-bank session start "Debug production database timeout" --project my-api

# Search for similar issues
memory-bank search "database timeout connection pool" --type error_solution

# Create investigation task
memory-bank task create \
  --project my-api \
  --title "Investigate database timeouts in production" \
  --description "Connection timeouts causing 500 errors" \
  --priority urgent \
  --tags bug,database,timeout

# Document investigation steps
memory-bank session log "Checked connection pool settings - max connections too low"
memory-bank session log "Analyzed slow query logs - found N+1 query in user loading"

# Store solution
memory-bank memory create --type error_solution \
  --title "Database connection timeout fix" \
  --content "Root cause: Connection pool exhaustion from N+1 queries. Solution: Increase pool size to 50, optimize user loading query. Prevention: Add connection monitoring." \
  --tags database,timeout,n+1-query,solution

# Complete debugging
memory-bank task update <task-id> --status done --actual-hours 6
memory-bank session complete "Fixed timeouts by optimizing queries and pool config. Added monitoring."
```

### Team Knowledge Sharing

```bash
# Code review workflow
memory-bank search "microservice communication patterns" --type decision

# Document new patterns from PR
memory-bank memory create --type pattern \
  --title "Event sourcing for order state management" \
  --content "Pattern for tracking order state changes using event sourcing with Kafka" \
  --tags event-sourcing,orders,kafka,pattern

# Share architectural insights
memory-bank memory create --type decision \
  --title "Service mesh adoption decision" \
  --content "Adopted Istio for service mesh. Considered Linkerd but chose Istio for better observability and security features." \
  --tags service-mesh,istio,architecture
```

## Integration Patterns

### CI/CD Integration

```bash
# In CI pipeline (e.g., GitHub Actions)
- name: Document build artifacts
  run: |
    memory-bank memory create --type documentation \
      --title "Build ${{ github.sha }}" \
      --content "Successful build with test coverage: 87%" \
      --tags build,ci,coverage

# Pre-commit hook for documentation
#!/bin/bash
changed_files=$(git diff --cached --name-only)
memory-bank doc analyze-changes --changed-files "$changed_files"
```

### IDE Integration

```bash
# VS Code task for quick memory creation
{
  "label": "Create Memory Bank Entry",
  "type": "shell",
  "command": "memory-bank memory create --type ${input:memoryType} --title '${input:title}' --content '${input:content}'"
}
```

### Automation Examples

```bash
# Daily knowledge capture
#!/bin/bash
# Capture daily standup notes
memory-bank memory create --type documentation \
  --title "Daily standup $(date +%Y-%m-%d)" \
  --content "$(cat standup-notes.md)" \
  --tags standup,daily,team

# Weekly pattern review
memory-bank search faceted "patterns" \
  --types pattern \
  --created-after "$(date -d '7 days ago' +%Y-%m-%d)" \
  --include-facets
```

### Performance Optimization

```bash
# Batch operations for large datasets
memory-bank memory list --project large-project --limit 1000 | \
  jq '.[] | select(.tags | contains(["deprecated"]))' | \
  xargs memory-bank memory delete

# Optimize search performance
export CHROMADB_BASE_URL="http://high-performance-chromadb:8000"
export OLLAMA_MODEL="nomic-embed-text"  # Optimized embedding model
```

## Best Practices Summary

### Memory Management
- Use consistent tagging strategies across team
- Regular cleanup of outdated memories
- Rich context in decisions with alternatives considered
- Link related memories through consistent tags

### Task Management
- Break large tasks into manageable subtasks
- Use realistic time estimates and track actuals
- Set clear priorities and dependencies
- Regular task reviews and updates

### Documentation Sync
- Set up automation early in project lifecycle
- Create clear code-to-docs mappings
- Regular consistency validation
- Include documentation in definition of done

### Search Strategy
- Start with simple searches, refine with facets
- Use multiple search approaches for discovery
- Leverage suggestions for pattern recognition
- Combine search with browsing for exploration

This workflow-driven approach ensures Memory Bank becomes an integral part of your development process, capturing knowledge as it's created and making it easily accessible when needed.