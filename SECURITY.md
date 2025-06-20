# Security Policy

Memory Bank is designed as a local development tool with a focus on simplicity and local-only operation. This document describes the actual security measures implemented and provides guidance for secure usage.

## Table of Contents

- [Security Model](#security-model)
- [Reporting Vulnerabilities](#reporting-vulnerabilities)
- [Current Security Features](#current-security-features)
- [Security Limitations](#security-limitations)
- [Best Practices](#best-practices)
- [Threat Model](#threat-model)

## Security Model

### Design Principles

Memory Bank follows a **"trusted local environment"** security model:

1. **Local-Only Operation**: All services run on localhost by default
2. **Stdio Isolation**: MCP protocol uses stdio transport (no network exposure)
3. **Minimal Attack Surface**: Simple architecture with few external dependencies
4. **Defense Through Simplicity**: Fewer components mean fewer potential vulnerabilities

### Architecture Overview

```
┌─────────────────┐
│   MCP Client    │ ← stdio transport only
│  (Claude Code)  │
└─────────┬───────┘
          │ stdin/stdout
    ┌─────▼─────┐
    │ Memory    │ ← HTTP to localhost services
    │ Bank      │
    └─────┬─────┘
          │
    ┌─────▼─────┐    ┌─────────────┐    ┌─────────────┐
    │  SQLite   │    │   Ollama    │    │  ChromaDB   │
    │ Database  │    │ (localhost) │    │ (localhost) │
    └───────────┘    └─────────────┘    └─────────────┘
```

## Reporting Vulnerabilities

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report security vulnerabilities by creating a private security advisory on GitHub or by contacting the maintainers directly.

When reporting a vulnerability, please include:
1. Clear description of the vulnerability
2. Steps to reproduce the issue  
3. Memory Bank version and environment details
4. Potential impact assessment

## Current Security Features

### ✅ Implemented Security Measures

#### SQL Injection Protection
- **Parameterized Queries**: All database operations use parameterized queries
- **Type Safety**: Go's type system prevents many injection vulnerabilities

#### Network Isolation
- **Localhost-Only Defaults**: External services default to localhost addresses
- **No Network Binding**: MCP server uses stdio transport only
- **Connection Timeouts**: 30-second timeouts prevent resource exhaustion

#### Input Validation
- **Required Field Validation**: MCP handlers check for required parameters
- **JSON Schema Validation**: Type checking through Go's JSON unmarshaling
- **Path Normalization**: File paths are normalized using `filepath.Abs()`

#### Service Reliability
- **Health Monitoring**: External services (Ollama, ChromaDB) include health checks
- **Graceful Fallbacks**: Automatic fallback to mock providers when services unavailable
- **Connection Pooling**: HTTP clients use connection pooling

## Security Limitations

### ❌ NOT Implemented

Memory Bank currently **does not** implement:

- **No Encryption**: Database, network communication, and configuration files are unencrypted
- **No Authentication**: External services (Ollama, ChromaDB) require no authentication
- **No Authorization**: No access controls beyond filesystem permissions
- **No Audit Logging**: Only basic application logging, no security event tracking
- **No Input Sanitization**: User content is stored as-provided
- **No Rate Limiting**: No protection against abuse or DDoS
- **No Secrets Management**: Configuration stored in plaintext

### Security Assumptions

Memory Bank assumes a **trusted local environment**:
- Single user access to the system
- Trusted local network environment
- No malicious processes on the same system
- User has appropriate filesystem permissions

## Best Practices

### Secure Installation

```bash
# Install to user directory (recommended)
mkdir -p ~/.local/bin
cp memory-bank ~/.local/bin/
chmod 755 ~/.local/bin/memory-bank

# Secure database location
mkdir -p ~/.memory-bank
chmod 700 ~/.memory-bank
```

### Configuration Security

```bash
# Secure configuration files
chmod 600 ~/.memory-bank/config.yaml

# Secure database files  
chmod 600 ~/.memory-bank/memory_bank.db
```

### Environment Variables

```bash
# Use environment variables for service URLs if needed
export OLLAMA_BASE_URL="http://127.0.0.1:11434"
export CHROMADB_BASE_URL="http://127.0.0.1:8000"

# Keep services local-only
# Do NOT expose Ollama or ChromaDB to external networks
```

### File System Security

```bash
# Ensure proper ownership
chown -R $USER:$USER ~/.memory-bank

# Prevent other users from accessing data
chmod -R go-rwx ~/.memory-bank
```

## Threat Model

### In Scope (Mitigated)

- **SQL Injection**: Protected through parameterized queries
- **Path Traversal**: Mitigated through path normalization
- **Network Exposure**: MCP protocol uses stdio only
- **Service Disruption**: Health checks and fallbacks provide resilience

### Out of Scope (Not Mitigated)

- **Local Privilege Escalation**: Relies on OS security
- **Data at Rest Protection**: No encryption implemented
- **Network Eavesdropping**: HTTP communication to localhost services
- **Malicious Input Processing**: No content sanitization
- **Resource Exhaustion**: No rate limiting or resource controls

### Risk Assessment

| Threat Category | Risk Level | Mitigation |
|----------------|------------|------------|
| Remote Network Attack | **Low** | Local-only operation, stdio transport |
| SQL Injection | **Low** | Parameterized queries |
| Path Traversal | **Low** | Path normalization |
| Data Breach (at rest) | **Medium** | File permissions only |
| Service Disruption | **Low** | Health checks, fallbacks |
| Privilege Escalation | **Medium** | OS-dependent |

## Development Security

### For Contributors

- Always use parameterized queries for database operations
- Validate all external inputs appropriately
- Keep external service communication local-only by default
- Add health checks for new external dependencies
- Write security-focused unit tests

### Code Review Checklist

- [ ] No hardcoded credentials or secrets
- [ ] Input validation for all external inputs
- [ ] Parameterized database queries
- [ ] Appropriate error handling (no information leakage)
- [ ] Local-only service connections

## Future Security Enhancements

Memory Bank may implement these security features in future versions:

- Database encryption at rest
- Authentication for external services
- Comprehensive audit logging
- Input sanitization and content filtering
- Resource usage limits
- Configuration encryption

## Contact Information

For security-related inquiries:
- **GitHub Security Advisories**: Preferred method for vulnerability reports
- **GitHub Issues**: For general security questions (non-sensitive)

---

**Last Updated**: January 2024  
**Security Model**: Trusted Local Environment  
**Threat Level**: Low (local development tool)