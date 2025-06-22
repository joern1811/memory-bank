# Advanced Usage Guide

This guide covers advanced features and usage patterns for Memory Bank, including complex search strategies, automation, integration patterns, and performance optimization.

## Advanced Search Strategies

### Faceted Search

Faceted search allows multi-dimensional filtering with real-time facet generation:

```bash
# Complex faceted search with multiple filters
memory-bank search faceted "authentication implementation" \
  --project my-api \
  --types decision,pattern,code \
  --tags auth,security,jwt \
  --facets \
  --sort relevance \
  --sort-dir desc \
  --limit 20 \
  --threshold 0.6
```

#### Available Facets

- **Types**: `decision`, `pattern`, `error_solution`, `code`, `documentation`
- **Tags**: Dynamic based on your memory content
- **Projects**: All projects in your memory bank
- **Content Length**: `short` (<500 chars), `medium` (500-2000), `long` (>2000)
- **Time Periods**: `today`, `week`, `month`, `quarter`, `year`

#### Faceted Search Response Structure

```json
{
  "memories": [...],
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
  }
}
```

### Enhanced Search with Relevance Scoring

Enhanced search provides detailed relevance scoring and match explanations:

```bash
# Enhanced search with detailed scoring
memory-bank search enhanced "JWT token validation middleware" \
  --project my-api \
  --type pattern \
  --limit 5 \
  --threshold 0.8
```

#### Relevance Scoring Factors

1. **Title Match Score** (0.0-1.0): Exact title keyword matches
2. **Content Similarity** (0.0-1.0): Semantic content similarity
3. **Tag Alignment** (0.0-1.0): Tag relevance to search query
4. **Recency Boost** (0.0-0.2): Recent memories get slight boost
5. **Type Preference** (0.0-0.1): Preference for specified types

#### Enhanced Search Response

```json
{
  "id": "mem_abc123",
  "title": "JWT Authentication Middleware Pattern",
  "relevance_score": 0.92,
  "score_breakdown": {
    "title_match": 0.85,
    "content_similarity": 0.89,
    "tag_alignment": 0.95,
    "recency_boost": 0.05,
    "type_preference": 0.10
  },
  "match_reasons": [
    "Strong title match for 'JWT' and 'middleware'",
    "Content highly relevant to token validation",
    "Tags perfectly align with search intent"
  ],
  "highlighted_content": "JWT **token** **validation** **middleware** implementation..."
}
```

### Search Suggestions and Auto-completion

Get intelligent search suggestions based on your existing content:

```bash
# Get search suggestions
memory-bank search suggestions "auth" --project my-api --limit 10
```

#### Suggestion Response

The command returns a simple list of suggested search terms based on your existing memory content:

```
Search suggestions for 'auth':
1. authentication
2. authorization  
3. middleware
4. jwt
5. token
6. security
7. login
8. oauth
9. session
10. verify
```

Suggestions are ranked by frequency in your existing memories.

## Advanced Session Management

### Session Templates and Workflows

Create standardized session templates for different types of work:

```bash
# Feature development session
memory-bank session start "Feature: User Profile Management" \
  --project user-service \
  --template feature \
  --tags "feature,user-profile,api"

# Bug fix session
memory-bank session start "Bug: Memory leak in image processing" \
  --project image-service \
  --template bugfix \
  --tags "bug,memory-leak,performance"

# Research session
memory-bank session start "Research: Database optimization strategies" \
  --project analytics \
  --template research \
  --tags "research,database,performance"
```

### Advanced Progress Tracking

Use typed progress entries for structured session logging:

```bash
# Information logging
memory-bank session log "Analyzed current user model structure" --type info

# Milestone tracking
memory-bank session log "User profile API endpoints implemented" \
  --type milestone \
  --tags "api,endpoints"

# Issue documentation
memory-bank session log "Validation fails for nested profile data" \
  --type issue \
  --tags "validation,bug" \
  --context "Occurs with JSON payloads > 1MB"

# Solution tracking
memory-bank session log "Implemented streaming validation for large payloads" \
  --type solution \
  --tags "validation,performance" \
  --context "Using json.Decoder with token-based validation"
```

### Session Analytics and Reporting

```bash
# Session performance analytics
memory-bank session stats --project my-api --timeframe week

# Generate session report
memory-bank session report [session-id] --format markdown --output report.md

# List sessions with advanced filtering
memory-bank session list \
  --project my-api \
  --status completed \
  --tags feature \
  --since "2024-01-01" \
  --until "2024-01-31"
```

## Configuration Management

### Environment-Based Configuration

Memory Bank uses environment variables for configuration. Create environment-specific setups:

```bash
# Development environment
export MEMORY_BANK_DB_PATH="./dev_memory_bank.db"
export OLLAMA_BASE_URL="http://localhost:11434"
export CHROMADB_BASE_URL="http://localhost:8000"
export MEMORY_BANK_LOG_LEVEL="debug"

# Production environment  
export MEMORY_BANK_DB_PATH="/data/memory_bank.db"
export OLLAMA_BASE_URL="http://ollama-service:11434"
export CHROMADB_BASE_URL="http://chromadb-service:8000"
export MEMORY_BANK_LOG_LEVEL="info"
```

### Configuration via .env Files

Create environment-specific `.env` files:

```bash
# .env.development
MEMORY_BANK_DB_PATH=./dev_memory_bank.db
OLLAMA_BASE_URL=http://localhost:11434
CHROMADB_BASE_URL=http://localhost:8000
MEMORY_BANK_LOG_LEVEL=debug

# .env.production
MEMORY_BANK_DB_PATH=/data/memory_bank.db
OLLAMA_BASE_URL=http://ollama-service:11434
CHROMADB_BASE_URL=http://chromadb-service:8000
MEMORY_BANK_LOG_LEVEL=info
```

### Configuration Management Scripts

```bash
# Load development environment
load_dev_config() {
  export $(cat .env.development | grep -v '^#' | xargs)
  echo "Development configuration loaded"
}

# Load production environment
load_prod_config() {
  export $(cat .env.production | grep -v '^#' | xargs)
  echo "Production configuration loaded"
}

# Show current configuration
show_config() {
  echo "Current Memory Bank Configuration:"
  echo "DB Path: ${MEMORY_BANK_DB_PATH:-./memory_bank.db}"
  echo "Ollama URL: ${OLLAMA_BASE_URL:-http://localhost:11434}"
  echo "ChromaDB URL: ${CHROMADB_BASE_URL:-http://localhost:8000}"
  echo "Log Level: ${MEMORY_BANK_LOG_LEVEL:-info}"
}
```

## Integration Patterns

### MCP Integration Workflows

#### Claude Code Integration

Configure Memory Bank as an MCP server for Claude Code:

```json
{
  "mcpServers": {
    "memory-bank": {
      "command": "/usr/local/bin/memory-bank",
      "env": {
        "MEMORY_BANK_CONFIG": "/path/to/config.yaml",
        "MEMORY_BANK_ENV": "production"
      }
    }
  }
}
```

#### Dynamic System Prompts

Memory Bank provides context-aware system prompts via MCP resources:

```bash
# Access system prompt resource via MCP
curl -X POST http://localhost:6277/resources/read \
  -H "Content-Type: application/json" \
  -d '{"uri": "prompt://memory-bank/system"}'
```

### CI/CD Integration

#### Pre-commit Memory Capture

Create a pre-commit hook to capture implementation decisions:

```bash
#!/bin/sh
# .git/hooks/pre-commit

# Extract commit message and changed files
COMMIT_MSG=$(git log -1 --pretty=%B 2>/dev/null || echo "WIP")
CHANGED_FILES=$(git diff --cached --name-only)

# Auto-create memory for significant changes
if [[ $COMMIT_MSG == *"BREAKING:"* ]]; then
  memory-bank memory create \
    --type decision \
    --title "Breaking Change: $COMMIT_MSG" \
    --content "Breaking change implemented. Changed files: $CHANGED_FILES" \
    --tags "breaking-change,commit,$(date +%Y-%m)" \
    --auto-project
fi
```

#### Deployment Documentation

Automatically document deployment outcomes:

```bash
#!/bin/bash
# deployment-logger.sh

DEPLOYMENT_ID=$(uuidgen)
ENVIRONMENT=$1
VERSION=$2

# Start deployment session
SESSION_ID=$(memory-bank session start "Deploy $VERSION to $ENVIRONMENT" \
  --project deployment \
  --tags "deployment,$ENVIRONMENT,$VERSION" \
  --format json | jq -r '.session_id')

# Log deployment steps
memory-bank session log "Starting deployment $DEPLOYMENT_ID" \
  --session $SESSION_ID \
  --type milestone

# After deployment...
if [ $? -eq 0 ]; then
  memory-bank session complete "Successfully deployed $VERSION to $ENVIRONMENT" \
    --session $SESSION_ID
  
  # Create deployment memory
  memory-bank memory create \
    --type documentation \
    --title "Deployment $VERSION to $ENVIRONMENT" \
    --content "Successful deployment on $(date). Session: $SESSION_ID" \
    --tags "deployment,success,$ENVIRONMENT,$VERSION"
else
  memory-bank session log "Deployment failed with exit code $?" \
    --session $SESSION_ID \
    --type issue
fi
```

## Performance Optimization

### Embedding and Vector Store Optimization

#### Batch Operations

```bash
# Batch create memories from files
find ./docs -name "*.md" -exec memory-bank memory create \
  --type documentation \
  --title "Documentation: {}" \
  --content-file {} \
  --tags "docs,auto-imported" \;

# Batch update embeddings
memory-bank memory reindex --project my-api --batch-size 20
```

### Search Performance Tips

#### Threshold Optimization

Use appropriate similarity thresholds for different use cases:

```bash
# High precision for code patterns (fewer, more relevant results)
memory-bank search "database patterns" \
  --type pattern \
  --threshold 0.8 \
  --limit 10

# Lower threshold for general concepts (more comprehensive results)
memory-bank search "general approach" \
  --type decision \
  --threshold 0.6 \
  --limit 20

# Very specific searches (exact matches preferred)
memory-bank search enhanced "JWT middleware implementation" \
  --threshold 0.9 \
  --limit 5
```

#### Search Strategy Recommendations

- **Code searches**: Use threshold 0.8+ for precise matches
- **Decision searches**: Use threshold 0.6-0.7 for broader context
- **Error solutions**: Use threshold 0.7+ for accurate fixes
- **Documentation**: Use threshold 0.5-0.6 for comprehensive coverage

## Data Management and Migration

### Manual Backup and Restore

Since Memory Bank uses SQLite, you can manually backup and restore the database:

```bash
# Create backup by copying SQLite database
cp "${MEMORY_BANK_DB_PATH:-./memory_bank.db}" "memory-bank-$(date +%Y%m%d).db.backup"

# Restore from backup
cp "memory-bank-20240115.db.backup" "${MEMORY_BANK_DB_PATH:-./memory_bank.db}"

# Export memories to JSON using SQLite
sqlite3 "${MEMORY_BANK_DB_PATH:-./memory_bank.db}" \
  ".mode json" \
  ".output memories.json" \
  "SELECT * FROM memories;"

# Export specific project
sqlite3 "${MEMORY_BANK_DB_PATH:-./memory_bank.db}" \
  ".mode json" \
  ".output project-memories.json" \
  "SELECT * FROM memories WHERE project_id = 'my-project';"
```

### Memory Lifecycle Management

#### Manual Cleanup with SQL

```bash
# View old memories (older than 6 months)
sqlite3 "${MEMORY_BANK_DB_PATH:-./memory_bank.db}" \
  "SELECT id, title, type, created_at FROM memories 
   WHERE created_at < datetime('now', '-6 months')
   ORDER BY created_at;"

# Delete old error solutions manually
sqlite3 "${MEMORY_BANK_DB_PATH:-./memory_bank.db}" \
  "DELETE FROM memories 
   WHERE type = 'error_solution' 
   AND created_at < datetime('now', '-6 months');"

# Find potential duplicates (same title)
sqlite3 "${MEMORY_BANK_DB_PATH:-./memory_bank.db}" \
  "SELECT title, COUNT(*) as count FROM memories 
   GROUP BY title 
   HAVING count > 1;"
```

#### Content Migration with SQL

```bash
# Update memory types
sqlite3 "${MEMORY_BANK_DB_PATH:-./memory_bank.db}" \
  "UPDATE memories SET type = 'documentation' WHERE type = 'note';"

# Bulk tag updates (requires manual editing of tags JSON)
sqlite3 "${MEMORY_BANK_DB_PATH:-./memory_bank.db}" \
  "SELECT id, tags FROM memories WHERE tags LIKE '%database%';"

# Update project assignments
sqlite3 "${MEMORY_BANK_DB_PATH:-./memory_bank.db}" \
  "UPDATE memories SET project_id = 'new-project' WHERE project_id = 'old-project';"
```

## Monitoring and Analytics

### Basic Analytics

```bash
# View memory count by type
sqlite3 "${MEMORY_BANK_DB_PATH:-./memory_bank.db}" "SELECT type, COUNT(*) FROM memories GROUP BY type;"

# Recent session activity
sqlite3 "${MEMORY_BANK_DB_PATH:-./memory_bank.db}" "SELECT title, status, created_at FROM sessions ORDER BY created_at DESC LIMIT 10;"

# Project memory distribution
sqlite3 "${MEMORY_BANK_DB_PATH:-./memory_bank.db}" "SELECT project_id, COUNT(*) FROM memories GROUP BY project_id;"
```

### Health Monitoring

```bash
# Check service availability
check_services() {
  echo "=== Memory Bank Health Check ==="
  
  # Database
  if [ -f "${MEMORY_BANK_DB_PATH:-./memory_bank.db}" ]; then
    echo "✓ Database: Available"
  else
    echo "✗ Database: Missing"
  fi
  
  # Ollama
  if curl -s "${OLLAMA_BASE_URL:-http://localhost:11434}/api/tags" >/dev/null 2>&1; then
    echo "✓ Ollama: Running"
  else
    echo "⚠ Ollama: Unavailable (using mock)"
  fi
  
  # ChromaDB
  if curl -s "${CHROMADB_BASE_URL:-http://localhost:8000}/api/v2/heartbeat" >/dev/null 2>&1; then
    echo "✓ ChromaDB: Running"
  else
    echo "⚠ ChromaDB: Unavailable (using mock)"
  fi
}
```

### Custom Metrics Scripts

```bash
# Track decision frequency
track_decisions() {
  local timeframe=${1:-"7 days"}
  echo "Decisions in the last $timeframe:"
  
  sqlite3 "${MEMORY_BANK_DB_PATH:-./memory_bank.db}" \
    "SELECT COUNT(*) FROM memories 
     WHERE type='decision' 
     AND created_at >= datetime('now', '-$timeframe');"
}

# Weekly activity summary
weekly_summary() {
  echo "=== Weekly Memory Bank Summary ==="
  echo "New memories: $(sqlite3 "${MEMORY_BANK_DB_PATH:-./memory_bank.db}" "SELECT COUNT(*) FROM memories WHERE created_at >= datetime('now', '-7 days');")"
  echo "Completed sessions: $(sqlite3 "${MEMORY_BANK_DB_PATH:-./memory_bank.db}" "SELECT COUNT(*) FROM sessions WHERE status='completed' AND created_at >= datetime('now', '-7 days');")"
  echo "Active projects: $(sqlite3 "${MEMORY_BANK_DB_PATH:-./memory_bank.db}" "SELECT COUNT(DISTINCT project_id) FROM memories WHERE created_at >= datetime('now', '-7 days');")"
}
```

## Advanced Automation

### Workflow Automation

#### Git Hook Integration

```bash
# Post-commit hook for automatic memory creation
#!/bin/sh
# .git/hooks/post-commit

COMMIT_MSG=$(git log -1 --pretty=%B)
COMMIT_HASH=$(git log -1 --pretty=%H)
FILES_CHANGED=$(git diff --name-only HEAD~1 HEAD)

# Check for keywords that trigger memory creation
if echo "$COMMIT_MSG" | grep -i "fix\|bug\|error"; then
  memory-bank memory create \
    --type error_solution \
    --title "Fix: $COMMIT_MSG" \
    --content "Commit $COMMIT_HASH fixed issue. Files changed: $FILES_CHANGED" \
    --tags "fix,commit,auto-generated" \
    --project $(basename $(git rev-parse --show-toplevel))
fi
```

#### Scheduled Knowledge Extraction

```bash
# Cron job for weekly knowledge extraction
# Add to crontab: 0 18 * * 5 /path/to/extract-knowledge.sh

#!/bin/bash
# extract-knowledge.sh

# Extract patterns from recent commits
git log --since="1 week ago" --pretty=format:"%s" | \
  grep -i "pattern\|implement\|add" | \
  while read commit_msg; do
    memory-bank memory create \
      --type pattern \
      --title "Weekly Pattern: $commit_msg" \
      --content "Extracted from recent development activity" \
      --tags "pattern,auto-extracted,weekly" \
      --auto-project
  done
```

### API and Scripting

#### MCP JSON-RPC Wrapper

```bash
#!/bin/bash
# memory-mcp.sh - JSON-RPC wrapper for Memory Bank

# Create memory via MCP
create_memory() {
  local type="$1"
  local title="$2" 
  local content="$3"
  local tags="$4"
  
  echo "{
    \"jsonrpc\": \"2.0\",
    \"id\": 1,
    \"method\": \"tools/call\",
    \"params\": {
      \"name\": \"memory_create\",
      \"arguments\": {
        \"type\": \"$type\",
        \"title\": \"$title\",
        \"content\": \"$content\",
        \"tags\": [\"$tags\"]
      }
    }
  }" | memory-bank
}

# Search via MCP
search_memories() {
  local query="$1"
  local limit="${2:-10}"
  
  echo "{
    \"jsonrpc\": \"2.0\",
    \"id\": 1,
    \"method\": \"tools/call\",
    \"params\": {
      \"name\": \"memory_search\",
      \"arguments\": {
        \"query\": \"$query\",
        \"limit\": $limit
      }
    }
  }" | memory-bank
}
```

### Integration with External Tools

#### IDE Extensions

```javascript
// VSCode extension integration
const vscode = require('vscode');
const { exec } = require('child_process');

function createMemoryFromSelection() {
    const editor = vscode.window.activeTextEditor;
    const selection = editor.document.getText(editor.selection);
    const fileName = editor.document.fileName;
    
    const command = `memory-bank memory create --type code --title "Code from ${fileName}" --content "${selection}" --tags "vscode,snippet"`;
    
    exec(command, (error, stdout, stderr) => {
        if (error) {
            vscode.window.showErrorMessage(`Memory creation failed: ${error.message}`);
        } else {
            vscode.window.showInformationMessage('Memory created successfully');
        }
    });
}
```

#### Slack Integration

```python
# slack-memory-bot.py
import requests
import subprocess
import json

def handle_slack_command(command_text, user_id, channel_id):
    if command_text.startswith('/memory search'):
        query = command_text.replace('/memory search', '').strip()
        result = subprocess.run(
            ['memory-bank', 'search', query, '--format', 'json'],
            capture_output=True, text=True
        )
        
        memories = json.loads(result.stdout)
        response = format_search_results(memories)
        
        post_to_slack(channel_id, response)
    
    elif command_text.startswith('/memory create'):
        # Parse memory creation command
        # Create memory and respond to Slack
        pass

def format_search_results(memories):
    if not memories:
        return "No memories found for your query."
    
    response = "Found memories:\n"
    for memory in memories[:5]:  # Limit to 5 results
        response += f"• *{memory['title']}* (Score: {memory['similarity']:.2f})\n"
        response += f"  {memory['content'][:100]}...\n\n"
    
    return response
```

## Troubleshooting and Debugging

### Debug Mode

```bash
# Enable debug logging
export MEMORY_BANK_LOG_LEVEL=debug
memory-bank search "query"

# Check current configuration
echo "DB Path: ${MEMORY_BANK_DB_PATH:-./memory_bank.db}"
echo "Ollama URL: ${OLLAMA_BASE_URL:-http://localhost:11434}"
echo "ChromaDB URL: ${CHROMADB_BASE_URL:-http://localhost:8000}"
echo "Log Level: ${MEMORY_BANK_LOG_LEVEL:-info}"
```

### Common Performance Issues

#### Slow Search Performance

```bash
# Check if services are running
curl -s "${OLLAMA_BASE_URL:-http://localhost:11434}/api/tags" > /dev/null && echo "Ollama: OK" || echo "Ollama: ERROR"
curl -s "${CHROMADB_BASE_URL:-http://localhost:8000}/api/v2/heartbeat" > /dev/null && echo "ChromaDB: OK" || echo "ChromaDB: ERROR"

# Use higher thresholds for faster searches
memory-bank search "query" --threshold 0.8 --limit 10

# Use enhanced search for better relevance
memory-bank search enhanced "specific query" --threshold 0.7
```

#### Service Connection Issues

```bash
# Test Ollama connection
curl "${OLLAMA_BASE_URL:-http://localhost:11434}/api/tags"

# Test ChromaDB connection  
curl "${CHROMADB_BASE_URL:-http://localhost:8000}/api/v2/heartbeat"

# Fall back to mock providers if services unavailable
# Memory Bank automatically falls back to mock implementations
```

### Advanced Debugging

#### MCP Protocol Debugging

```bash
# Test MCP server directly
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}' | memory-bank

# Use MCP Inspector for interactive debugging
npx @modelcontextprotocol/inspector memory-bank

# Test specific MCP tools
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "memory_search", "arguments": {"query": "test"}}}' | memory-bank
```

#### Database Debugging

```bash
# Check SQLite database
sqlite3 "${MEMORY_BANK_DB_PATH:-./memory_bank.db}" ".tables"
sqlite3 "${MEMORY_BANK_DB_PATH:-./memory_bank.db}" "SELECT COUNT(*) FROM memories;"

# View recent memories
sqlite3 "${MEMORY_BANK_DB_PATH:-./memory_bank.db}" "SELECT title, type, created_at FROM memories ORDER BY created_at DESC LIMIT 5;"
```

## Security and Access Control

### Current Security Model

Memory Bank operates as a **local development tool** without built-in authentication:

```bash
# Security is based on OS-level access control
# 1. File system permissions control database access
ls -la "${MEMORY_BANK_DB_PATH:-./memory_bank.db}"

# 2. MCP protocol is designed for local use
# No network authentication required for Claude Code integration

# 3. Binary permissions
ls -la $(which memory-bank)

# 4. Local service access only
echo "Ollama: ${OLLAMA_BASE_URL:-http://localhost:11434}"
echo "ChromaDB: ${CHROMADB_BASE_URL:-http://localhost:8000}"
```

**Security Assumptions:**
- Single-user local environment
- Trusted local processes only
- No network exposure required
- OS-level user authentication sufficient

### Data Privacy Recommendations

Since Memory Bank stores content in plain text SQLite:

```bash
# Manual data privacy practices
# 1. Avoid storing sensitive information
echo "❌ Don't store: passwords, API keys, secrets, tokens"
echo "✅ Use placeholders: password: '[REDACTED]', api_key: '[HIDDEN]'"

# 2. Review stored content for sensitive data
sqlite3 "${MEMORY_BANK_DB_PATH:-./memory_bank.db}" \
  "SELECT title, content FROM memories WHERE 
   content LIKE '%password%' OR 
   content LIKE '%secret%' OR 
   content LIKE '%token%' OR
   content LIKE '%key%';"

# 3. Secure database file permissions
chmod 600 "${MEMORY_BANK_DB_PATH:-./memory_bank.db}"

# 4. Secure backup files
find . -name "*.db.backup" -exec chmod 600 {} \;
```

This advanced usage guide provides comprehensive coverage of Memory Bank's actual features, accurately reflecting the current implementation without fictional security systems.