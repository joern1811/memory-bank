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

### Advanced Configuration

Create environment-specific configurations:

```yaml
# ~/.memory-bank/config.yaml
environments:
  development:
    database:
      path: "./dev_memory_bank.db"
    embedding:
      provider: "ollama"
    vector:
      provider: "mock"  # Faster for development
    
  production:
    database:
      path: "/data/memory_bank.db"
    embedding:
      provider: "ollama"
      ollama:
        base_url: "http://ollama-service:11434"
    vector:
      provider: "chromadb"
      chromadb:
        base_url: "http://chromadb-service:8000"
        tenant: "production"
        database: "memory_bank"

# Performance tuning
performance:
  embedding:
    batch_size: 10
    cache_size: 1000
    cache_ttl: "1h"
  
  vector:
    max_results: 100
    timeout: "30s"
  
  database:
    max_connections: 25
    connection_timeout: "10s"

# Search configuration
search:
  default_threshold: 0.7
  max_results: 50
  highlight_enabled: true
  suggestion_limit: 10
```

### Configuration Validation and Management

```bash
# Validate configuration
memory-bank config validate

# Show current configuration
memory-bank config show

# Set specific configuration values
memory-bank config set embedding.provider ollama
memory-bank config set vector.chromadb.base_url http://localhost:8001

# Environment-specific configuration
memory-bank --env production config show
memory-bank --env development session start "Test feature"
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

#### Connection Pool Configuration

```yaml
performance:
  ollama:
    max_connections: 10
    connection_timeout: "30s"
    request_timeout: "60s"
  
  chromadb:
    max_connections: 15
    connection_timeout: "10s"
    request_timeout: "30s"
    batch_size: 50
```

### Search Performance Tuning

#### Cache Configuration

```yaml
performance:
  embedding:
    cache_enabled: true
    cache_size: 1000        # Number of embeddings to cache
    cache_ttl: "1h"         # Time to live
    
  search:
    result_cache_enabled: true
    result_cache_size: 100  # Number of search results to cache
    result_cache_ttl: "10m"
```

#### Search Optimization

```bash
# Use specific thresholds for different content types
memory-bank search "database patterns" \
  --type pattern \
  --threshold 0.8 \        # Higher precision for code patterns
  --limit 10

memory-bank search "general approach" \
  --type decision \
  --threshold 0.6 \        # Lower threshold for decisions
  --limit 20
```

## Data Management and Migration

### Backup and Restore

```bash
# Create backup
memory-bank backup create --output memory-bank-$(date +%Y%m%d).backup

# Restore from backup
memory-bank backup restore --input memory-bank-20240115.backup

# Export to different formats
memory-bank export --format json --output memories.json
memory-bank export --format csv --output memories.csv --project my-api
```

### Memory Lifecycle Management

#### Automatic Cleanup

```bash
# Archive old memories
memory-bank memory archive \
  --older-than "6 months" \
  --type error_solution \
  --auto-confirm

# Delete draft memories
memory-bank memory delete \
  --tag draft \
  --older-than "30 days" \
  --auto-confirm

# Merge duplicate memories
memory-bank memory deduplicate \
  --threshold 0.95 \
  --interactive
```

#### Content Migration

```bash
# Migrate from legacy format
memory-bank migrate from-legacy \
  --input legacy-memories.json \
  --mapping-file migration-map.yaml

# Update memory types
memory-bank memory update-type \
  --from "note" \
  --to "documentation" \
  --project old-project

# Bulk tag update
memory-bank memory retag \
  --old-tag "database" \
  --new-tag "persistence" \
  --project my-api
```

## Monitoring and Analytics

### Usage Analytics

```bash
# Memory usage statistics
memory-bank stats memory --project my-api --timeframe month

# Search analytics
memory-bank stats search --queries --popular-terms --timeframe week

# Session productivity metrics
memory-bank stats sessions --completion-rate --avg-duration --project my-api
```

### Health Monitoring

```bash
# Check system health
memory-bank health check --verbose

# Service availability
memory-bank health services --ollama --chromadb --database

# Performance metrics
memory-bank health performance --embedding-speed --search-speed
```

### Custom Metrics and Reporting

```yaml
# ~/.memory-bank/metrics.yaml
custom_metrics:
  decision_tracking:
    query: "type:decision created:last_month"
    alert_threshold: 5
    description: "Track architectural decisions per month"
  
  error_resolution_rate:
    query: "type:error_solution tags:resolved"
    timeframe: "week"
    description: "Weekly error resolution tracking"

reporting:
  daily_summary:
    enabled: true
    time: "18:00"
    include: ["new_memories", "session_completions", "search_activity"]
  
  weekly_digest:
    enabled: true
    day: "friday"
    format: "markdown"
    email: "team@company.com"
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

#### REST API Wrapper

```bash
#!/bin/bash
# memory-api.sh - REST API wrapper around Memory Bank

start_api_server() {
  memory-bank api server --port 8080 --auth-token $API_TOKEN &
  echo $! > /tmp/memory-api.pid
}

# Create memory via API
create_memory() {
  curl -X POST http://localhost:8080/api/v1/memories \
    -H "Authorization: Bearer $API_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"type\": \"$1\",
      \"title\": \"$2\",
      \"content\": \"$3\",
      \"tags\": [\"$4\"]
    }"
}

# Search via API
search_memories() {
  curl -X GET "http://localhost:8080/api/v1/search?q=$1&limit=$2" \
    -H "Authorization: Bearer $API_TOKEN"
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
memory-bank search "query" --debug

# Verbose output for all operations
memory-bank --verbose memory create --type decision --title "Test"

# Performance profiling
memory-bank --profile search "complex query" --limit 100
```

### Common Performance Issues

#### Slow Search Performance

```bash
# Check embedding generation time
memory-bank benchmark embedding --count 10

# Check vector search performance  
memory-bank benchmark search --queries 50

# Optimize search configuration
memory-bank config set search.default_threshold 0.8  # Higher threshold = faster
memory-bank config set vector.chromadb.timeout 10s   # Lower timeout
```

#### Memory Usage Optimization

```bash
# Monitor memory usage
memory-bank stats memory-usage --verbose

# Clear caches
memory-bank cache clear --all

# Optimize database
memory-bank database optimize --vacuum --reindex
```

### Advanced Debugging

#### Vector Store Debugging

```bash
# Check vector store health
memory-bank debug vector-store --test-connection --test-operations

# Validate embeddings
memory-bank debug embeddings --check-dimensions --check-validity

# Inspect search pipeline
memory-bank debug search "query" --explain --trace
```

#### Database Debugging

```bash
# Check database integrity
memory-bank debug database --check-integrity --check-indexes

# Analyze query performance
memory-bank debug database --explain-queries --slow-query-log

# Database statistics
memory-bank debug database --stats --table-sizes
```

## Security and Access Control

### Authentication and Authorization

```yaml
# security.yaml
authentication:
  enabled: true
  method: "jwt"  # or "api_key", "basic"
  
  jwt:
    secret: "${JWT_SECRET}"
    expiry: "24h"
    
  api_key:
    header: "X-API-Key"
    keys:
      - name: "development"
        key: "${DEV_API_KEY}"
        permissions: ["read", "write"]
      - name: "production"
        key: "${PROD_API_KEY}"
        permissions: ["read"]

authorization:
  project_based: true
  default_permissions: ["read"]
  
  roles:
    admin:
      permissions: ["read", "write", "delete", "admin"]
    developer:
      permissions: ["read", "write"]
    readonly:
      permissions: ["read"]
```

### Data Privacy and Encryption

```yaml
privacy:
  encryption:
    enabled: true
    algorithm: "AES-256-GCM"
    key_source: "environment"  # or "file", "vault"
    
  data_masking:
    enabled: true
    patterns:
      - type: "email"
        replacement: "[EMAIL]"
      - type: "api_key"
        pattern: "(sk-[a-zA-Z0-9]{32,})"
        replacement: "[API_KEY]"
      - type: "password"
        pattern: "(password|passwd|pwd)[:=]\\s*([^\\s]+)"
        replacement: "$1: [REDACTED]"

audit:
  enabled: true
  log_file: "/var/log/memory-bank/audit.log"
  events: ["create", "update", "delete", "search", "access"]
```

This advanced usage guide provides comprehensive coverage of Memory Bank's sophisticated features, enabling power users to leverage the full potential of the semantic memory management system.