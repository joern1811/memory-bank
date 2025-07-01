# Troubleshooting Guide

This guide helps you diagnose and resolve common issues with Memory Bank. Use the table of contents to quickly find solutions to specific problems.

## Table of Contents

- [Quick Diagnostics](#quick-diagnostics)
- [Installation Issues](#installation-issues)
- [Configuration Problems](#configuration-problems)
- [Database Issues](#database-issues)
- [External Service Problems](#external-service-problems)
- [Search and Embedding Issues](#search-and-embedding-issues)
- [Performance Problems](#performance-problems)
- [MCP Integration Issues](#mcp-integration-issues)
- [Command Line Interface Problems](#command-line-interface-problems)
- [Session Management Issues](#session-management-issues)
- [Error Messages Reference](#error-messages-reference)
- [Debug Tools and Logging](#debug-tools-and-logging)
- [Recovery Procedures](#recovery-procedures)
- [Getting Help](#getting-help)

## Quick Diagnostics

Start here for rapid problem identification:

```bash
# Check overall system health
memory-bank health check --verbose

# Verify basic functionality
memory-bank memory create --type test --title "Health Check" --content "Testing basic functionality"
memory-bank search "Health Check" --limit 1
memory-bank memory delete [returned-id]

# Check configuration
memory-bank config show

# View recent logs
memory-bank logs --tail 50
```

### System Requirements Check

```bash
# Check Go version (requires 1.21+)
go version

# Check available disk space
df -h $(dirname $(memory-bank config get database.path))

# Check memory usage
memory-bank stats system --memory --cpu

# Verify external services
curl -f http://localhost:11434/api/tags 2>/dev/null && echo "Ollama: OK" || echo "Ollama: Unavailable"
curl -f http://localhost:8000/api/v2/heartbeat 2>/dev/null && echo "ChromaDB: OK" || echo "ChromaDB: Unavailable"
```

## Installation Issues

### Build Failures

**Problem**: `go build` fails with dependency errors
```bash
go build ./cmd/memory-bank
# Error: module not found, version conflicts, etc.
```

**Solutions**:
```bash
# Clean module cache and rebuild
go clean -modcache
go mod tidy
go mod download
go build ./cmd/memory-bank

# Force module refresh
rm go.sum
go mod tidy
go build ./cmd/memory-bank

# Check Go version compatibility
go version  # Ensure 1.21+
```

### Permission Issues

**Problem**: Binary execution permission denied

**Solutions**:
```bash
# Make binary executable
chmod +x memory-bank

# If installed system-wide
sudo chmod +x /usr/local/bin/memory-bank

# Check file ownership
ls -la memory-bank
```

### PATH Configuration

**Problem**: `memory-bank: command not found`

**Solutions**:
```bash
# Add to PATH temporarily
export PATH=$PATH:/path/to/memory-bank/directory

# Add to shell profile permanently
echo 'export PATH=$PATH:/path/to/memory-bank' >> ~/.bashrc
source ~/.bashrc

# System-wide installation
sudo cp memory-bank /usr/local/bin/
sudo chmod +x /usr/local/bin/memory-bank
```

## Configuration Problems

### Configuration File Not Found

**Problem**: Configuration not loading, default values used

**Solutions**:
```bash
# Check configuration file locations
memory-bank config locations

# Initialize default configuration
memory-bank config init --force

# Verify configuration loading
memory-bank config validate

# Set configuration path explicitly
export MEMORY_BANK_CONFIG=/path/to/config.yaml
```

### Invalid Configuration

**Problem**: Configuration validation errors

**Diagnosis**:
```bash
# Validate configuration syntax
memory-bank config validate --verbose

# Check specific configuration sections
memory-bank config get database
memory-bank config get embedding
memory-bank config get vector
```

**Solutions**:
```bash
# Reset to defaults
memory-bank config reset

# Fix specific configuration values
memory-bank config set database.path "/valid/path/memory_bank.db"
memory-bank config set embedding.provider "ollama"

# Use environment variables instead
export MEMORY_BANK_DB_PATH="/valid/path/memory_bank.db"
export OLLAMA_BASE_URL="http://localhost:11434"
```

### Environment Variable Conflicts

**Problem**: Environment variables override configuration file unexpectedly

**Diagnosis**:
```bash
# List Memory Bank related environment variables
env | grep MEMORY_BANK
env | grep OLLAMA
env | grep CHROMADB

# Check configuration precedence
memory-bank config explain --show-sources
```

**Solutions**:
```bash
# Clear conflicting environment variables
unset MEMORY_BANK_DB_PATH
unset OLLAMA_BASE_URL

# Use configuration file exclusively
memory-bank config set use_environment false
```

## Database Issues

### Database Locked Error

**Problem**: `database is locked` error

**Diagnosis**:
```bash
# Check for running Memory Bank processes
ps aux | grep memory-bank

# Check database file permissions
ls -la /path/to/memory_bank.db

# Check if file is being accessed by other processes
lsof /path/to/memory_bank.db
```

**Solutions**:
```bash
# Kill other Memory Bank processes
pkill memory-bank

# Check and fix file permissions
chmod 644 /path/to/memory_bank.db
chown $USER:$USER /path/to/memory_bank.db

# Move database to accessible location
memory-bank config set database.path "$HOME/memory_bank.db"

# Use temporary database for testing
memory-bank --db-path /tmp/test_memory.db memory list
```

### Database Corruption

**Problem**: SQLite database corruption errors

**Diagnosis**:
```bash
# Check database integrity
memory-bank database check --integrity

# Attempt to read database structure
sqlite3 memory_bank.db ".schema"
```

**Solutions**:
```bash
# Backup current database
cp memory_bank.db memory_bank_backup.db

# Attempt database repair
memory-bank database repair --backup

# Recover from backup if available
memory-bank database restore --from-backup memory_bank_backup.db

# Start fresh (last resort)
rm memory_bank.db
memory-bank init --force
```

### Schema Migration Issues

**Problem**: Database schema version mismatch

**Diagnosis**:
```bash
# Check current schema version
memory-bank database version

# List available migrations
memory-bank database migrations --list
```

**Solutions**:
```bash
# Run pending migrations
memory-bank database migrate --up

# Rollback problematic migration
memory-bank database migrate --down --steps 1

# Force schema recreation
memory-bank database recreate --backup-first
```

### Performance Issues with Database

**Problem**: Slow database operations

**Diagnosis**:
```bash
# Check database size and statistics
memory-bank database stats --verbose

# Analyze slow queries
memory-bank database analyze --slow-queries

# Check disk space
df -h $(dirname $(memory-bank config get database.path))
```

**Solutions**:
```bash
# Optimize database
memory-bank database optimize --vacuum --reindex

# Clean old data
memory-bank memory cleanup --older-than "6 months"

# Move database to faster storage
memory-bank database move --to /path/to/ssd/memory_bank.db
```

## External Service Problems

### Ollama Connection Issues

**Problem**: Cannot connect to Ollama service

**Diagnosis**:
```bash
# Check if Ollama is running
curl http://localhost:11434/api/tags

# Check Ollama process
ps aux | grep ollama

# Test different endpoints
curl http://localhost:11434/api/version
```

**Solutions**:
```bash
# Start Ollama service
ollama serve

# Check if model is available
ollama list
ollama pull nomic-embed-text

# Use different port or host
memory-bank config set embedding.ollama.base_url "http://localhost:11435"

# Use mock provider as fallback
memory-bank config set embedding.provider "mock"

# Test with different model
memory-bank config set embedding.ollama.model "all-minilm"
```

### Ollama Model Issues

**Problem**: Embedding model not working or unavailable

**Diagnosis**:
```bash
# List available models
ollama list

# Test model directly
ollama run nomic-embed-text "test embedding"

# Check model pull status
ollama ps
```

**Solutions**:
```bash
# Pull required model
ollama pull nomic-embed-text

# Try alternative models
ollama pull all-minilm
memory-bank config set embedding.ollama.model "all-minilm"

# Check disk space for model storage
df -h ~/.ollama

# Restart Ollama service
ollama stop
ollama serve
```

### ChromaDB Connection Issues

**Problem**: Cannot connect to ChromaDB

**Diagnosis**:
```bash
# Check ChromaDB availability
curl http://localhost:8000/api/v2/heartbeat

# Check if ChromaDB is running
docker ps | grep chroma
# OR for native installation
ps aux | grep chroma
```

**Solutions**:
```bash
# Start ChromaDB with Docker
docker run -p 8000:8000 chromadb/chroma

# Start ChromaDB with uvx
uvx --from "chromadb[server]" chroma run --host 0.0.0.0 --port 8000

# Use different port
memory-bank config set vector.chromadb.base_url "http://localhost:8001"

# Use mock vector store
memory-bank config set vector.provider "mock"

# Check network connectivity
nc -zv localhost 8000
```

### ChromaDB API Version Issues

**Problem**: API version compatibility errors

**Diagnosis**:
```bash
# Check ChromaDB version
curl http://localhost:8000/api/v2/version

# Test different API endpoints
curl http://localhost:8000/api/v2/collections
```

**Solutions**:
```bash
# Update ChromaDB to latest version
docker pull chromadb/chroma:latest

# Use specific API version in configuration
memory-bank config set vector.chromadb.api_version "v2"

# Check Memory Bank compatibility
memory-bank version --compatibility
```

## Search and Embedding Issues

### No Search Results

**Problem**: Search returns no results despite having memories

**Diagnosis**:
```bash
# Check if memories exist
memory-bank memory list --limit 10

# Test with lower threshold
memory-bank search "query" --threshold 0.1

# Check embedding generation
memory-bank search "query" --debug --verbose

# Verify vector store connectivity
memory-bank debug vector-store --test-search
```

**Solutions**:
```bash
# Regenerate embeddings
memory-bank memory reindex --all

# Lower search threshold
memory-bank search "query" --threshold 0.3

# Check content indexing
memory-bank debug embeddings --check-indexed

# Use exact text search instead
memory-bank memory list | grep "search term"
```

### Poor Search Quality

**Problem**: Search returns irrelevant results

**Diagnosis**:
```bash
# Test with different similarity thresholds
memory-bank search "query" --threshold 0.8
memory-bank search "query" --threshold 0.5

# Check embedding model quality
memory-bank benchmark embedding --test-similarity

# Analyze search result relevance
memory-bank search "query" --explain --debug
```

**Solutions**:
```bash
# Use enhanced search for better relevance
memory-bank search enhanced "query" --limit 10

# Improve search query specificity
memory-bank search "specific technical term" --type pattern

# Use faceted search for filtering
memory-bank search faceted "query" --types decision,pattern

# Retrain with better content
memory-bank memory create --type pattern --title "Specific Pattern" --content "Detailed implementation"
```

### Embedding Generation Failures

**Problem**: Embeddings not being generated

**Diagnosis**:
```bash
# Test embedding generation directly
memory-bank debug embeddings --test-generation --text "test content"

# Check Ollama model availability
ollama list | grep nomic-embed-text

# Verify embedding configuration
memory-bank config get embedding
```

**Solutions**:
```bash
# Switch to mock provider temporarily
memory-bank config set embedding.provider "mock"

# Pull embedding model again
ollama pull nomic-embed-text

# Test with different model
memory-bank config set embedding.ollama.model "all-minilm"

# Check available memory for model
free -h
```

## Performance Problems

### Slow Search Performance

**Problem**: Search queries take too long

**Diagnosis**:
```bash
# Benchmark search performance
memory-bank benchmark search --queries 10

# Check vector store performance
memory-bank debug vector-store --benchmark

# Monitor resource usage during search
top -p $(pgrep memory-bank)
```

**Solutions**:
```bash
# Enable search result caching
memory-bank config set performance.search.cache_enabled true

# Reduce search result limit
memory-bank search "query" --limit 5

# Optimize database indices
memory-bank database optimize --reindex

# Use higher similarity threshold
memory-bank search "query" --threshold 0.8
```

### High Memory Usage

**Problem**: Memory Bank consumes excessive RAM

**Diagnosis**:
```bash
# Check memory usage statistics
memory-bank stats system --memory --detailed

# Monitor memory during operations
memory-bank memory list --profile-memory

# Check cache sizes
memory-bank config get performance.embedding.cache_size
```

**Solutions**:
```bash
# Reduce cache sizes
memory-bank config set performance.embedding.cache_size 100

# Clear caches
memory-bank cache clear --embedding --search

# Optimize memory usage
memory-bank config set performance.memory.limit "512MB"

# Process memories in smaller batches
memory-bank memory reindex --batch-size 5
```

### Slow Startup Time

**Problem**: Memory Bank takes long time to start

**Diagnosis**:
```bash
# Profile startup time
time memory-bank --help

# Check initialization steps
memory-bank --debug --verbose config show

# Monitor external service connections
memory-bank health check --startup-time
```

**Solutions**:
```bash
# Disable health checks on startup
memory-bank config set startup.health_check_enabled false

# Use mock providers for faster startup
memory-bank config set embedding.provider "mock"
memory-bank config set vector.provider "mock"

# Optimize database connection
memory-bank config set database.connection_timeout "5s"
```

## MCP Integration Issues

### MCP Server Not Starting

**Problem**: MCP server fails to start or respond

**Diagnosis**:
```bash
# Test MCP server directly
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}' | memory-bank mcp serve
```

**Solutions**:
```bash
# Test with MCP Inspector
npx @modelcontextprotocol/inspector memory-bank mcp serve

# Enable MCP debug logging
MEMORY_BANK_LOG_LEVEL=debug memory-bank mcp serve

# Enable verbose output
memory-bank mcp serve --verbose
```

### Claude Code Integration Problems

**Problem**: Claude Code cannot connect to Memory Bank

**Diagnosis**:
```bash
# Check Claude Code configuration
cat ~/.config/claude-desktop/config.json

# Verify binary path in configuration
which memory-bank

# Test MCP connection manually
echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"capabilities": {}}}' | memory-bank
```

**Solutions**:
```bash
# Update Claude Code configuration with correct path
# Edit ~/.config/claude-desktop/config.json
{
  "mcpServers": {
    "memory-bank": {
      "command": "/usr/local/bin/memory-bank",
      "args": ["mcp", "serve"],
      "env": {
        "MEMORY_BANK_LOG_LEVEL": "info"
      }
    }
  }
}

# Restart Claude Code
# Verify configuration syntax
cat ~/.config/claude-desktop/config.json | jq .
```

### MCP Tool Registration Issues

**Problem**: MCP tools not appearing in Claude Code

**Diagnosis**:
```bash
# List available MCP tools
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}' | memory-bank

# Check tool registration
memory-bank mcp list-tools --verbose

# Test specific tool
echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "memory/search", "arguments": {"query": "test"}}}' | memory-bank
```

**Solutions**:
```bash
# Force tool re-registration
memory-bank mcp register-tools --force

# Check for tool naming conflicts
memory-bank mcp validate-tools

# Update MCP protocol version
memory-bank config set mcp.protocol_version "2024-11-05"
```

## Command Line Interface Problems

### Command Not Found

**Problem**: Specific Memory Bank commands not recognized

**Diagnosis**:
```bash
# List available commands
memory-bank --help

# Check command spelling and syntax
memory-bank memory --help
memory-bank session --help

# Verify binary version
memory-bank version
```

**Solutions**:
```bash
# Use correct command syntax
memory-bank memory create --type decision --title "Title" --content "Content"

# Check for command aliases
memory-bank search --help | grep aliases

# Update to latest version if commands are missing
go build ./cmd/memory-bank
```

### Argument Parsing Issues

**Problem**: Command arguments not parsed correctly

**Diagnosis**:
```bash
# Test with simple arguments
memory-bank memory list

# Check argument quoting
memory-bank memory create --title "Title with spaces" --content "Content"

# Use debug mode to see parsed arguments
memory-bank --debug memory create --title "Test"
```

**Solutions**:
```bash
# Use proper quoting for arguments with spaces
memory-bank memory create --title "My Decision" --content "Detailed explanation"

# Escape special characters
memory-bank memory create --content "Code with \"quotes\""

# Use configuration file for complex arguments
memory-bank memory create --config memory-config.yaml
```

### Output Formatting Issues

**Problem**: Command output not formatted as expected

**Diagnosis**:
```bash
# Test different output formats
memory-bank memory list --format json
memory-bank memory list --format table
memory-bank memory list --format yaml

# Check terminal capabilities
echo $TERM
tput colors
```

**Solutions**:
```bash
# Force specific output format
memory-bank memory list --format json --no-color

# Use raw output for scripting
memory-bank memory list --format raw

# Adjust terminal settings
export TERM=xterm-256color
```

## Session Management Issues

### Session State Problems

**Problem**: Sessions not tracking state correctly

**Diagnosis**:
```bash
# Check active sessions
memory-bank session list --status active

# Verify session data integrity
memory-bank session get [session-id] --verbose

# Check session storage
memory-bank debug session-store --validate
```

**Solutions**:
```bash
# Complete orphaned sessions
memory-bank session list --status active | xargs -I {} memory-bank session complete {} "Session auto-completed"

# Reset session state
memory-bank session reset [session-id]

# Clear session cache
memory-bank cache clear --sessions
```

### Session Creation Failures

**Problem**: Cannot create new sessions

**Diagnosis**:
```bash
# Test session creation with minimal data
memory-bank session start "Test Session" --project test

# Check project existence
memory-bank project list

# Verify session storage permissions
ls -la ~/.memory-bank/sessions/
```

**Solutions**:
```bash
# Create project first
memory-bank project init . --name "Test Project"

# Fix session storage permissions
mkdir -p ~/.memory-bank/sessions
chmod 755 ~/.memory-bank/sessions

# Use different session storage location
memory-bank config set session.storage_path "/tmp/memory-bank-sessions"
```

## Error Messages Reference

### Common Error Messages and Solutions

#### `failed to connect to database: database is locked`
**Cause**: Another process is using the database
**Solution**: Kill other Memory Bank processes or change database path

#### `embedding generation failed: connection refused`
**Cause**: Ollama service not running
**Solution**: Start Ollama with `ollama serve`

#### `vector search failed: ChromaDB unavailable`
**Cause**: ChromaDB service not accessible
**Solution**: Start ChromaDB or switch to mock provider

#### `configuration validation failed: invalid file path`
**Cause**: Configuration file syntax error or invalid path
**Solution**: Run `memory-bank config validate` and fix errors

#### `memory creation failed: content too large`
**Cause**: Memory content exceeds size limits
**Solution**: Reduce content size or increase limits in configuration

#### `search failed: no embeddings found`
**Cause**: Memories haven't been indexed yet
**Solution**: Run `memory-bank memory reindex --all`

#### `session operation failed: no active session`
**Cause**: No active session for current project
**Solution**: Start a session with `memory-bank session start`

#### `MCP protocol error: method not found`
**Cause**: MCP client requesting unsupported method
**Solution**: Update Memory Bank or check MCP client compatibility

## Debug Tools and Logging

### Enabling Debug Mode

```bash
# Enable debug logging
export MEMORY_BANK_LOG_LEVEL=debug

# Use verbose output
memory-bank --verbose search "query"

# Profile operations
memory-bank --profile memory create --type test --title "Test" --content "Test"
```

### Log File Analysis

```bash
# View recent logs
memory-bank logs --tail 100

# Search logs for errors
memory-bank logs --grep "ERROR" --since "1 hour ago"

# Export logs for analysis
memory-bank logs --export /tmp/memory-bank-logs.json

# Monitor logs in real-time
memory-bank logs --follow
```

### Performance Profiling

```bash
# Profile memory usage
memory-bank profile memory --duration 30s

# Profile CPU usage
memory-bank profile cpu --output cpu-profile.prof

# Profile search operations
memory-bank profile search --queries 100 --output search-profile.json

# Generate performance report
memory-bank profile report --output performance-report.html
```

## Recovery Procedures

### Disaster Recovery

#### Complete System Recovery
```bash
# 1. Backup current state
memory-bank backup create --full --output disaster-backup.tar.gz

# 2. Verify backup integrity
memory-bank backup verify disaster-backup.tar.gz

# 3. Clean installation
rm -rf ~/.memory-bank
rm memory_bank.db

# 4. Reinstall Memory Bank
go build ./cmd/memory-bank

# 5. Restore from backup
memory-bank backup restore disaster-backup.tar.gz

# 6. Verify restoration
memory-bank health check --full
```

#### Data Recovery from Corruption
```bash
# 1. Stop all Memory Bank processes
pkill memory-bank

# 2. Backup corrupted database
cp memory_bank.db memory_bank_corrupted.db

# 3. Attempt SQLite recovery
sqlite3 memory_bank_corrupted.db ".recover" | sqlite3 memory_bank_recovered.db

# 4. Verify recovered data
memory-bank --db-path memory_bank_recovered.db memory list

# 5. Replace original database
mv memory_bank_recovered.db memory_bank.db

# 6. Rebuild indices
memory-bank database optimize --reindex
```

### Partial Recovery Procedures

#### Recover Missing Embeddings
```bash
# Identify memories without embeddings
memory-bank debug embeddings --list-missing

# Regenerate missing embeddings
memory-bank memory reindex --missing-only

# Verify embedding completeness
memory-bank debug embeddings --verify-all
```

#### Recover Vector Store Data
```bash
# Check vector store consistency
memory-bank debug vector-store --check-consistency

# Rebuild vector store from database
memory-bank vector-store rebuild --from-database

# Verify vector data integrity
memory-bank debug vector-store --verify-data
```

## Getting Help

### Documentation Resources

- **Getting Started Guide**: `docs/guides/getting-started.md`
- **API Documentation**: `docs/api/`
- **Examples**: `docs/guides/examples.md`
- **Advanced Usage**: `docs/guides/advanced-usage.md`

### Community Support

```bash
# Check version and build information
memory-bank version --verbose

# Generate diagnostic report
memory-bank diagnostics --output memory-bank-diagnostic.json

# Export configuration for support
memory-bank config export --sanitized --output config-for-support.yaml
```

### Bug Reporting

When reporting bugs, include:

1. **Version Information**:
   ```bash
   memory-bank version --verbose
   ```

2. **System Information**:
   ```bash
   uname -a
   go version
   memory-bank health check --system
   ```

3. **Configuration**:
   ```bash
   memory-bank config export --sanitized
   ```

4. **Logs**:
   ```bash
   memory-bank logs --since "1 hour ago" --level error
   ```

5. **Reproduction Steps**: Clear steps to reproduce the issue

6. **Expected vs Actual Behavior**: What you expected vs what happened

### Emergency Recovery

If Memory Bank is completely broken:

```bash
# Nuclear option: complete reset
rm -rf ~/.memory-bank
rm memory_bank.db
unset MEMORY_BANK_*
go build ./cmd/memory-bank
./memory-bank init --force
```

Remember to backup your data before performing destructive operations!

---

This troubleshooting guide covers the most common issues and their solutions. For issues not covered here, please check the [project documentation](../README.md) or create a detailed bug report with the information outlined in the [Getting Help](#getting-help) section.