# Memory Bank MCP API Reference

Memory Bank implements the Model Context Protocol (MCP) to provide semantic memory management for Claude Code. This document describes all available MCP methods and their parameters.

## Overview

The Memory Bank MCP server provides three main categories of operations:
- **Memory Operations**: CRUD operations for semantic memory entries
- **Project Operations**: Project initialization and management  
- **Session Operations**: Development session tracking and management

All methods follow JSON-RPC 2.0 protocol format and return structured responses with proper error handling.

## Memory Operations

### `memory/create`

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
  "method": "memory/create",
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

### `memory/search`

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
  "method": "memory/search", 
  "params": {
    "query": "authentication patterns and JWT implementation",
    "project_id": "proj_456",
    "limit": 5,
    "threshold": 0.7
  }
}
```

### `memory/get`

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

### `memory/update`

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

### `memory/delete`

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

### `memory/list`

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

### `project/init`

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
  "method": "project/init",
  "params": {
    "name": "E-Commerce API",
    "path": "/home/user/projects/ecommerce-api",
    "description": "RESTful API for e-commerce platform with microservices architecture"
  }
}
```

### `project/get`

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

### `project/list`

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

### `session/start`

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
  "method": "session/start",
  "params": {
    "title": "Implement User Authentication",
    "project_id": "proj_456", 
    "description": "Adding JWT-based authentication with login/register endpoints",
    "tags": ["auth", "api", "security"]
  }
}
```

### `session/log`

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

### `session/complete`

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

### `session/get`

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

### `session/list`

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

### `session/abort`

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