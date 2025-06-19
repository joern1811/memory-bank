# Memory Bank MCP API Reference

Memory Bank implements the Model Context Protocol (MCP) to provide semantic memory management for Claude Code. This document describes all available MCP methods and their parameters.

## Overview

The Memory Bank MCP server provides three main categories of operations:
- **Memory Operations**: CRUD operations for semantic memory entries
- **Project Operations**: Project initialization and management  
- **Session Operations**: Development session tracking and management

All methods follow JSON-RPC 2.0 protocol format and return structured responses with proper error handling.

## Memory Operations

### `memory_create`

Creates a new memory entry with semantic embeddings for search.

**Parameters:**
```json
{
  "project_id": "string (optional)",
  "type": "decision|pattern|error_solution|code|documentation",
  "title": "string (required)",
  "content": "string (required)", 
  "tags": ["string"] (optional),
  "session_id": "string (optional)"
}
```

**Response:**
```json
{
  "id": "mem_abc123",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Example:**
```json
{
  "method": "memory_create",
  "params": {
    "project_id": "proj_456",
    "type": "decision",
    "title": "Use JWT for Authentication",
    "content": "After evaluating OAuth2, session cookies, and JWT tokens, we decided to implement JWT-based authentication for our API. This provides stateless authentication and works well with our microservice architecture.",
    "tags": ["auth", "security", "api", "jwt"],
    "session_id": "sess_789"
  }
}
```

### `memory_search`

Performs semantic search across memory entries using vector embeddings.

**Parameters:**
```json
{
  "query": "string (required)",
  "project_id": "string (optional)",
  "limit": "number (default: 10, max: 100)",
  "threshold": "number (default: 0.5, range: 0.0-1.0)",
  "type": "string (optional)"
}
```

**Response:**
```json
{
  "results": [
    {
      "id": "mem_abc123",
      "title": "Use JWT for Authentication",
      "content": "After evaluating OAuth2...",
      "type": "decision",
      "tags": ["auth", "security"],
      "similarity": 0.87,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1
}
```

**Example:**
```json
{
  "method": "memory_search", 
  "params": {
    "query": "authentication patterns and JWT implementation",
    "project_id": "proj_456",
    "limit": 5,
    "threshold": 0.7
  }
}
```

**Error Response:**
```json
{
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": {
      "field": "threshold",
      "error": "must be between 0.0 and 1.0"
    }
  }
}
```

### `memory_get`

Retrieves a specific memory entry by ID.

**Parameters:**
```json
{
  "id": "string (required)"
}
```

**Response:**
```json
{
  "id": "mem_abc123",
  "title": "Use JWT for Authentication",
  "content": "After evaluating OAuth2...",
  "type": "decision", 
  "tags": ["auth", "security"],
  "project_id": "proj_456",
  "session_id": "sess_789",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### `memory_update`

Updates an existing memory entry. Only provided fields are updated.

**Parameters:**
```json
{
  "id": "string (required)",
  "title": "string (optional)",
  "content": "string (optional)",
  "tags": ["string"] (optional),
  "type": "string (optional)"
}
```

**Response:**
```json
{
  "updated_at": "2024-01-15T11:45:00Z"
}
```

### `memory_delete`

Deletes a memory entry and its associated vector embeddings.

**Parameters:**
```json
{
  "id": "string (required)"
}
```

**Response:**
```json
{
  "deleted": true
}
```

### `memory_list`

Lists memory entries with optional filtering and pagination.

**Parameters:**
```json
{
  "project_id": "string (optional)",
  "type": "string (optional)", 
  "tags": ["string"] (optional),
  "limit": "number (default: 20, max: 100)",
  "offset": "number (default: 0)"
}
```

**Response:**
```json
{
  "memories": [
    {
      "id": "mem_abc123",
      "title": "Use JWT for Authentication", 
      "type": "decision",
      "tags": ["auth", "security"],
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1,
  "offset": 0,
  "limit": 20
}
```

## Project Operations

### `project_init`

Initializes a new project for memory management.

**Parameters:**
```json
{
  "name": "string (required)",
  "path": "string (required)",
  "description": "string (optional)"
}
```

**Response:**
```json
{
  "id": "proj_abc123",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Example:**
```json
{
  "method": "project_init",
  "params": {
    "name": "E-Commerce API",
    "path": "/home/user/projects/ecommerce-api",
    "description": "RESTful API for e-commerce platform with microservices architecture"
  }
}
```

### `project_get`

Retrieves project information by ID or path.

**Parameters:**
```json
{
  "id": "string (optional)",
  "path": "string (optional)"
}
```
*Note: Either `id` or `path` must be provided.*

**Response:**
```json
{
  "id": "proj_abc123",
  "name": "E-Commerce API",
  "path": "/home/user/projects/ecommerce-api", 
  "description": "RESTful API for e-commerce platform",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### `project_list`

Lists all registered projects.

**Parameters:**
```json
{
  "limit": "number (default: 50, max: 100)",
  "offset": "number (default: 0)"
}
```

**Response:**
```json
{
  "projects": [
    {
      "id": "proj_abc123",
      "name": "E-Commerce API",
      "path": "/home/user/projects/ecommerce-api",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1,
  "offset": 0,
  "limit": 50
}
```

## Session Operations

### `session_start`

Starts a new development session for tracking progress.

**Parameters:**
```json
{
  "title": "string (required)",
  "project_id": "string (required)",
  "description": "string (optional)",
  "tags": ["string"] (optional)
}
```

**Response:**
```json
{
  "id": "sess_abc123",
  "started_at": "2024-01-15T10:30:00Z"
}
```

**Example:**
```json
{
  "method": "session_start",
  "params": {
    "title": "Implement User Authentication",
    "project_id": "proj_456", 
    "description": "Adding JWT-based authentication with login/register endpoints",
    "tags": ["auth", "api", "security"]
  }
}
```

### `session_log`

Logs progress or notes to an active session.

**Parameters:**
```json
{
  "session_id": "string (required)",
  "entry_type": "info|milestone|issue|solution (default: info)",
  "content": "string (required)",
  "tags": ["string"] (optional)
}
```

**Response:**
```json
{
  "logged_at": "2024-01-15T10:45:00Z"
}
```

### `session_complete`

Completes an active session with outcome summary.

**Parameters:**
```json
{
  "session_id": "string (required)",
  "outcome": "string (required)",
  "summary": "string (optional)"
}
```

**Response:**
```json
{
  "completed_at": "2024-01-15T11:30:00Z",
  "duration_minutes": 60
}
```

### `session_get`

Retrieves detailed session information including progress logs.

**Parameters:**
```json
{
  "id": "string (required)"
}
```

**Response:**
```json
{
  "id": "sess_abc123",
  "title": "Implement User Authentication",
  "project_id": "proj_456",
  "status": "completed",
  "started_at": "2024-01-15T10:30:00Z",
  "completed_at": "2024-01-15T11:30:00Z",
  "duration_minutes": 60,
  "progress": [
    {
      "timestamp": "2024-01-15T10:45:00Z",
      "type": "milestone",
      "content": "JWT middleware implemented and tested",
      "tags": ["jwt", "middleware"]
    }
  ]
}
```

### `session_list`

Lists sessions with optional filtering.

**Parameters:**
```json
{
  "project_id": "string (optional)",
  "status": "active|completed|aborted (optional)",
  "limit": "number (default: 20, max: 100)",
  "offset": "number (default: 0)"
}
```

**Response:**
```json
{
  "sessions": [
    {
      "id": "sess_abc123",
      "title": "Implement User Authentication",
      "project_id": "proj_456",
      "status": "completed",
      "started_at": "2024-01-15T10:30:00Z",
      "duration_minutes": 60
    }
  ],
  "total": 1
}
```

### `session_abort`

Aborts active sessions for a project.

**Parameters:**
```json
{
  "project_id": "string (required)"
}
```

**Response:**
```json
{
  "aborted_sessions": ["sess_abc123"],
  "count": 1
}
```

## Error Handling

All methods return structured errors following JSON-RPC 2.0 format:

```json
{
  "error": {
    "code": -32000,
    "message": "Memory not found",
    "data": {
      "memory_id": "mem_invalid123"
    }
  }
}
```

### Common Error Codes

- `-32700`: Parse error (invalid JSON)
- `-32600`: Invalid request (missing required fields)
- `-32601`: Method not found
- `-32602`: Invalid params (validation failed)
- `-32000`: Server error (internal errors)

### Validation Errors

- Missing required fields
- Invalid memory types
- Invalid similarity thresholds (must be 0.0-1.0)
- Invalid limits (max 100 for most operations)
- Invalid project paths
- Duplicate project names

## Rate Limits and Performance

- **Embedding Generation**: Batched for performance (up to 10 concurrent)
- **Vector Search**: Optimized with caching (60-80% cache hit rate)
- **Database Operations**: Connection pooling and batch queries
- **Memory Usage**: Lightweight metadata queries for list operations

## Authentication and Security

Currently, Memory Bank MCP server runs locally without authentication. In production deployments:

- All data stored locally in SQLite database
- Embeddings generated locally via Ollama (no external API calls)
- File system access restricted to configured project paths
- Input validation and sanitization on all operations

## Health and Monitoring

The server performs automatic health checks on startup:
- Ollama service availability and model access
- ChromaDB connection and collection management
- SQLite database schema validation and migrations

Fallback to mock providers ensures functionality even when external services are unavailable.

## Advanced Features

### Faceted Search

Memory Bank supports advanced faceted search through the `memory_faceted-search` method:

**Parameters:**
```json
{
  "query": "string (required)",
  "project_id": "string (optional)",
  "filters": {
    "types": ["decision", "pattern"],
    "tags": ["auth", "security"],
    "min_length": 100,
    "has_content": true
  },
  "include_facets": true,
  "sort_by": {"field": "relevance", "direction": "desc"},
  "limit": 20,
  "threshold": 0.6
}
```

**Response with Facets:**
```json
{
  "results": [...],
  "facets": {
    "types": {
      "decision": 15,
      "pattern": 8,
      "code": 12
    },
    "tags": {
      "auth": 10,
      "security": 8,
      "jwt": 6
    },
    "content_length": {
      "short": 5,
      "medium": 18,
      "long": 2
    }
  },
  "total": 25
}
```

### Enhanced Relevance Scoring

Use `memory_enhanced-search` for detailed relevance scoring:

**Response with Scoring Details:**
```json
{
  "results": [
    {
      "id": "mem_abc123",
      "title": "JWT Authentication Middleware",
      "relevance_score": 0.92,
      "score_breakdown": {
        "title_match": 0.85,
        "content_similarity": 0.89,
        "tag_alignment": 0.95,
        "recency_boost": 0.05
      },
      "match_reasons": [
        "Strong title match for 'JWT' and 'middleware'",
        "Content highly relevant to token validation",
        "Tags perfectly align with search intent"
      ],
      "highlighted_content": "JWT **token** **validation** **middleware** implementation..."
    }
  ]
}
```

### Search Suggestions

Get intelligent search suggestions with `memory_search-suggestions`:

**Example:**
```json
{
  "method": "memory_search-suggestions",
  "params": {
    "partial_query": "auth",
    "project_id": "proj_456",
    "limit": 10
  }
}
```

**Response:**
```json
{
  "suggestions": [
    {
      "query": "authentication middleware patterns",
      "reason": "Based on 5 related memories",
      "confidence": 0.9
    },
    {
      "query": "JWT token validation",
      "reason": "Matches existing code patterns",
      "confidence": 0.85
    }
  ]
}
```

## Complete Usage Examples

### Architectural Decision Recording

```json
// 1. Create decision memory
{
  "method": "memory_create",
  "params": {
    "type": "decision",
    "title": "Microservices vs Monolith Architecture",
    "content": "After analyzing our team size (8 developers), expected growth (50+ in 2 years), and system complexity, we decided to adopt microservices architecture. Key factors: better team autonomy, independent deployments, technology diversity, but accepted trade-offs in operational complexity and distributed system challenges.",
    "tags": ["architecture", "microservices", "scaling", "team-organization"],
    "project_id": "proj_ecommerce"
  }
}

// 2. Add related pattern
{
  "method": "memory_create", 
  "params": {
    "type": "pattern",
    "title": "API Gateway Pattern for Microservices",
    "content": "Implemented Kong API Gateway as single entry point for all microservices. Handles authentication, rate limiting, request routing, and response aggregation. Configuration: auth-service for JWT validation, rate limiting at 1000 req/min per user, circuit breaker with 5-second timeout.",
    "tags": ["microservices", "api-gateway", "kong", "authentication"],
    "project_id": "proj_ecommerce"
  }
}

// 3. Search for related decisions
{
  "method": "memory_search",
  "params": {
    "query": "microservices architecture decisions",
    "project_id": "proj_ecommerce",
    "type": "decision", 
    "limit": 5
  }
}
```

### Error Solution Documentation

```json
// 1. Document error solution
{
  "method": "memory_create",
  "params": {
    "type": "error_solution",
    "title": "Fix Docker Build Failing with Go Module Cache",
    "content": "Error: 'go: cannot find module providing package'. Solution: Add GOPROXY and GOSUMDB environment variables in Dockerfile, copy go.mod and go.sum before source code, run 'go mod download' as separate layer for better caching. Final working Dockerfile pattern included.",
    "tags": ["docker", "golang", "build", "modules", "ci-cd"],
    "project_id": "proj_ecommerce"
  }
}

// 2. Search for build-related issues
{
  "method": "memory_enhanced-search",
  "params": {
    "query": "docker build problems golang modules",
    "type": "error_solution",
    "threshold": 0.7,
    "limit": 3
  }
}
```

### Development Session Tracking

```json
// 1. Start development session
{
  "method": "session_start",
  "params": {
    "title": "Implement Payment Processing",
    "project_id": "proj_ecommerce",
    "description": "Add Stripe integration for payment processing with webhook handling"
  }
}

// 2. Log progress throughout development
{
  "method": "session_log",
  "params": {
    "message": "Stripe SDK integrated and configured",
    "type": "info"
  }
}

{
  "method": "session_log", 
  "params": {
    "message": "Payment endpoint implemented with validation",
    "type": "milestone"
  }
}

{
  "method": "session_log",
  "params": {
    "message": "Webhook signature validation failing",
    "type": "issue"
  }
}

{
  "method": "session_log",
  "params": {
    "message": "Fixed webhook validation by updating endpoint_secret configuration",
    "type": "solution"
  }
}

// 3. Complete session
{
  "method": "session_complete",
  "params": {
    "outcome": "Payment processing fully implemented with Stripe integration. Includes webhook handling, error recovery, and comprehensive test coverage."
  }
}
```

### Project Knowledge Management

```json
// 1. Initialize project
{
  "method": "project_init",
  "params": {
    "name": "E-commerce Platform",
    "path": "/path/to/ecommerce-project", 
    "description": "Modern e-commerce platform built with microservices architecture"
  }
}

// 2. Create project overview documentation
{
  "method": "memory_create",
  "params": {
    "type": "documentation",
    "title": "Project Architecture Overview",
    "content": "System consists of 7 microservices: user-service (auth), product-service (catalog), order-service (transactions), payment-service (stripe), inventory-service (stock), notification-service (emails), and api-gateway (kong). Each service has independent database, deployment pipeline, and monitoring.",
    "tags": ["architecture", "overview", "microservices", "documentation"],
    "project_id": "proj_ecommerce"
  }
}

// 3. Search project knowledge
{
  "method": "memory_faceted-search",
  "params": {
    "query": "service architecture patterns",
    "project_id": "proj_ecommerce",
    "include_facets": true,
    "filters": {
      "types": ["decision", "pattern", "documentation"]
    }
  }
}
```

## Error Handling Best Practices

### Validation Error Example

```json
// Request with invalid parameters
{
  "method": "memory_create",
  "params": {
    "type": "invalid_type",
    "title": "",
    "content": "Content without title"
  }
}

// Error response
{
  "error": {
    "code": -32602,
    "message": "Invalid parameters",
    "data": {
      "validation_errors": [
        {
          "field": "type",
          "error": "must be one of: decision, pattern, error_solution, code, documentation"
        },
        {
          "field": "title", 
          "error": "title cannot be empty"
        }
      ]
    }
  }
}
```

### Service Unavailable Example

```json
// When embedding service is down
{
  "error": {
    "code": -32000,
    "message": "Embedding generation failed",
    "data": {
      "service": "ollama",
      "fallback": "Using mock embeddings for basic functionality",
      "retry_suggested": true
    }
  }
}
```

### Database Error Example

```json
// Database constraint violation
{
  "error": {
    "code": -32000,
    "message": "Database operation failed",
    "data": {
      "constraint": "unique_memory_title_per_project",
      "suggestion": "Use a different title or update the existing memory"
    }
  }
}
```

## Integration Patterns

### Batch Operations

For bulk operations, combine multiple method calls:

```json
// Create multiple related memories in sequence
[
  {
    "method": "memory_create",
    "params": {"type": "decision", "title": "Database Choice", "content": "..."}
  },
  {
    "method": "memory_create", 
    "params": {"type": "pattern", "title": "Connection Pool Pattern", "content": "..."}
  },
  {
    "method": "memory_search",
    "params": {"query": "database patterns", "limit": 10}
  }
]
```

### Workflow Integration

Typical development workflow using Memory Bank:

1. **Project Setup**: `project_init` → `memory_create` (documentation)
2. **Decision Making**: `memory_search` → analyze → `memory_create` (decision)
3. **Development**: `session_start` → `session_log` → `memory_create` (patterns/solutions)
4. **Knowledge Retrieval**: `memory_search` → `memory_get` → apply knowledge
5. **Session Completion**: `session_complete` → `memory_create` (lessons learned)

This integration pattern ensures comprehensive knowledge capture throughout the development lifecycle.