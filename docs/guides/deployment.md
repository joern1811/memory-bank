# Production Deployment Guide

Guide for deploying Memory Bank as an MCP server with Claude Code in production environments.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Configuration](#configuration)
- [MCP Integration](#mcp-integration)
- [Development Setup](#development-setup)
- [Troubleshooting](#troubleshooting)

## Overview

Memory Bank is a semantic memory management system that runs as an MCP (Model Context Protocol) server. It provides intelligent storage and retrieval of development knowledge.

### Architecture

```
┌─────────────────────────────────────┐
│           Claude Desktop            │
│              or                     │
│           Claude Code               │
└─────────────┬───────────────────────┘
              │ MCP Protocol
              │
┌─────────────▼───────────────────────┐
│        Memory Bank MCP Server       │
│  ┌─────────────────────────────────┐ │
│  │         Go Application          │ │
│  │  • Memory CRUD operations       │ │
│  │  • Semantic search             │ │
│  │  • Session management          │ │
│  │  • Project management          │ │
│  └─────────────────────────────────┘ │
│  ┌─────────────────────────────────┐ │
│  │         Data Storage            │ │
│  │  • SQLite database             │ │
│  │  • Ollama embeddings (opt.)    │ │
│  │  • ChromaDB vectors (opt.)     │ │
│  └─────────────────────────────────┘ │
└─────────────────────────────────────┘
```

## Installation

### Option 1: Homebrew (Recommended)

```bash
# Install via Homebrew
brew tap joern1811/tap
brew install --cask joern1811/tap/memory-bank
```

### Option 2: Build from Source

```bash
# Clone and build
git clone https://github.com/your-repo/memory-bank.git
cd memory-bank
go build ./cmd/memory-bank

# Move binary to PATH (optional)
sudo mv memory-bank /usr/local/bin/
```

### Option 3: Download Release

```bash
# Download latest release
wget https://github.com/your-repo/memory-bank/releases/latest/download/memory-bank-linux-amd64.tar.gz
tar -xzf memory-bank-linux-amd64.tar.gz
sudo mv memory-bank /usr/local/bin/
```

## Configuration

### Quick Start (Mock Providers)

For immediate use without external dependencies:

```bash
# Test the installation
memory-bank --help

# Initialize a project (uses mock providers by default)
memory-bank init . --name "My Project"

# Test basic functionality
memory-bank memory create --type decision --title "Test" --content "Test memory"
memory-bank search "test"
```

### Full Setup with External Services

#### 1. Install Ollama (for embeddings)

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull embedding model
ollama pull nomic-embed-text
```

#### 2. Install ChromaDB (for vector search)

```bash
# Option A: Using uvx (recommended)
uvx --from "chromadb[server]" chroma run --host localhost --port 8000 --path ./chromadb_data &

# Option B: Using Docker
docker run -d --name chroma -p 8000:8000 -v chroma-data:/chroma/chroma chromadb/chroma
```

#### 3. Environment Variables

```bash
# Optional: Configure external services
export OLLAMA_BASE_URL="http://localhost:11434"
export OLLAMA_MODEL="nomic-embed-text"
export CHROMADB_BASE_URL="http://localhost:8000"
export MEMORY_BANK_DB_PATH="./memory_bank.db"
export MEMORY_BANK_LOG_LEVEL="info"
```

## MCP Integration

Memory Bank integrates with Claude via the Model Context Protocol (MCP).

### Claude Code Configuration (Primary Integration)

#### Option 1: Claude Code CLI (Recommended)

```bash
# Global for all projects
claude mcp add memory-bank \
  -e MEMORY_BANK_DB_PATH=~/memory_bank.db \
  --scope user \
  -- memory-bank

# Project-specific (for teams)
claude mcp add memory-bank \
  -e MEMORY_BANK_DB_PATH=./memory_bank.db \
  --scope local \
  -- memory-bank

# Verify configuration
claude mcp list

# Add user-scoped server (personal use)
claude mcp add memory-bank -s user memory-bank

# Verify installation
claude mcp list
```

#### Option 2: Manual Configuration Files

**Project Scope**: Create `.mcp.json` in your project root:

```json
{
  "mcpServers": {
    "memory-bank": {
      "command": "memory-bank",
      "args": ["mcp", "serve"],
      "env": {
        "MEMORY_BANK_DB_PATH": "./memory_bank.db"
      }
    }
  }
}
```

**User Scope**: Edit `~/.claude.json`:

```json
{
  "mcpServers": {
    "memory-bank": {
      "command": "memory-bank",
      "args": ["mcp", "serve"],
      "env": {
        "MEMORY_BANK_DB_PATH": "~/memory_bank.db"
      }
    }
  }
}
```

### Testing MCP Integration

#### Using MCP Inspector

```bash
# Install MCP Inspector for debugging
npx @modelcontextprotocol/inspector memory-bank
```

#### Manual Testing

```bash
# Test MCP server directly
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}' | memory-bank

# Test initialization
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "project_init", "arguments": {"name": "Test", "path": "."}}}' | memory-bank
```

### Verification

After configuration, restart Claude Desktop/Code and verify:

1. Memory Bank appears in available tools
2. You can create and search memories
3. Project initialization works
4. Session management functions properly

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MEMORY_BANK_DB_PATH` | `./memory_bank.db` | SQLite database location |
| `MEMORY_BANK_LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |
| `OLLAMA_BASE_URL` | `http://localhost:11434` | Ollama server URL |
| `OLLAMA_MODEL` | `nomic-embed-text` | Embedding model name |
| `CHROMADB_BASE_URL` | `http://localhost:8000` | ChromaDB server URL |

## Development Setup

### CLI Development

For development and testing:

```bash
# Clone repository
git clone https://github.com/your-repo/memory-bank.git
cd memory-bank

# Install dependencies
go mod tidy

# Build and test
go build ./cmd/memory-bank
go test ./...

# Run tests with coverage
go test -v -cover ./...

# Initialize development project
./memory-bank init . --name "Development"

# Create test memories
./memory-bank memory create --type decision --title "Use Go" --content "Go is excellent for this project"
./memory-bank memory create --type pattern --title "Error handling" --content "Always wrap errors with context"

# Test search functionality
./memory-bank search "error handling"
./memory-bank search enhanced "Go patterns"
```

### MCP Development

```bash
# Test MCP server in isolation
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}' | ./memory-bank

# Test specific MCP tools
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "memory_create", "arguments": {"type": "test", "title": "Test", "content": "Test memory"}}}' | ./memory-bank

# Debug with verbose logging
MEMORY_BANK_LOG_LEVEL=debug ./memory-bank
```

### Docker Development (Optional)

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o memory-bank ./cmd/memory-bank

FROM alpine:3.18
RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/
COPY --from=builder /app/memory-bank .
CMD ["./memory-bank"]
```

```bash
# Build Docker image
docker build -t memory-bank:dev .

# Run as MCP server
docker run -v $(pwd):/data memory-bank:dev
```

## Troubleshooting

### Common Issues

#### Installation Problems

```bash
# Binary not found
which memory-bank
echo $PATH

# Permission denied
ls -la /usr/local/bin/memory-bank
sudo chmod +x /usr/local/bin/memory-bank

# Build issues (if building from source)
go version  # Ensure Go 1.21+
go mod tidy
go clean -cache
```

#### MCP Integration Issues

```bash
# Test MCP server
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}' | memory-bank

# Check Claude configuration
# macOS: ~/Library/Application Support/Claude/claude_desktop_config.json
# Windows: %APPDATA%\Claude\claude_desktop_config.json

# Verify binary path in config
which memory-bank

# Check MCP inspector
npx @modelcontextprotocol/inspector memory-bank
```

#### Database Issues

```bash
# Check database file
ls -la ./memory_bank.db
sqlite3 ./memory_bank.db ".tables"

# Fix permissions
chmod 644 ./memory_bank.db

# Test database access
sqlite3 ./memory_bank.db "SELECT COUNT(*) FROM memories;"
```

#### External Service Issues

```bash
# Test Ollama connection
curl http://localhost:11434/api/tags
ollama list
ollama ps

# Test ChromaDB connection
curl http://localhost:8000/api/v2/heartbeat
curl http://localhost:8000/api/v2/collections

# Check if services are running
ps aux | grep ollama
ps aux | grep chroma
```

#### Performance Issues

```bash
# Check system resources
top
df -h
free -h

# Enable debug logging
MEMORY_BANK_LOG_LEVEL=debug memory-bank

# Check database size
ls -lh ./memory_bank.db
```

### Health Checks

```bash
# Basic functionality test
memory-bank memory create --type test --title "Health Check" --content "Testing functionality"
memory-bank search "health"
memory-bank memory list

# Check system health (if implemented)
memory-bank health

# Verify MCP tools
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}' | memory-bank | jq '.result[].name'
```

### Logging and Debugging

```bash
# Enable verbose logging
export MEMORY_BANK_LOG_LEVEL=debug

# Check for error patterns
grep -i error ~/.claude/logs/*
grep -i memory-bank ~/.claude/logs/*

# Monitor real-time activity
tail -f ./memory-bank.log
```

### Support Resources

- **GitHub Issues**: Report bugs and request features
- **Documentation**: Check the main README and guides
- **MCP Specification**: Model Context Protocol documentation
- **Claude Code Documentation**: Integration guides






---

This deployment guide covers the essential steps for installing and configuring Memory Bank as an MCP server. For additional support, check the troubleshooting section above or refer to the project documentation.