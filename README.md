# Memory Bank

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Build Status](https://github.com/joern1811/memory-bank/workflows/CI%2FCD%20Pipeline/badge.svg)](https://github.com/joern1811/memory-bank/actions)

A semantic memory management system for developers that integrates with Claude Code via MCP (Model Context Protocol). Memory Bank helps you capture, organize, and retrieve development knowledge using AI-powered semantic search.

## 🚀 Features

- **🧠 Semantic Memory Management**: Store decisions, patterns, code snippets, and documentation with AI-powered search
- **🔌 Claude Code Integration**: Native MCP (Model Context Protocol) server for seamless Claude Code integration
- **⚡ Dual Interface**: Both MCP server mode and traditional CLI for flexible usage
- **🎯 Smart Search**: Vector-based semantic search with faceted filtering and relevance scoring
- **📊 Session Tracking**: Structured development session logging with progress tracking
- **🏗️ Clean Architecture**: Hexagonal architecture with pluggable components
- **🔒 Privacy-First**: Local-only processing with optional external services
- **📈 Performance Optimized**: Concurrent processing, caching, and batch operations

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [Installation](#installation)
- [Usage](#usage)
- [Architecture](#architecture)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [Support](#support)

## ⚡ Quick Start

### 1. Install Memory Bank

```bash
# Clone or download Memory Bank
cd /path/to/memory-bank

# Build the application
go build ./cmd/memory-bank

# Make it executable and add to PATH (optional)
chmod +x memory-bank
sudo mv memory-bank /usr/local/bin/
```

### 2. Initialize Your First Project

```bash
# Navigate to your project directory
cd /path/to/your/project

# Initialize Memory Bank for this project
memory-bank init . --name "My Project" --description "My awesome project"
```

### 3. Create Your First Memory

```bash
# Store an architectural decision
memory-bank memory create \
  --type decision \
  --title "Use PostgreSQL for primary database" \
  --content "After evaluating options, chose PostgreSQL for ACID compliance and JSON support" \
  --tags "database,architecture,postgresql"
```

### 4. Search Your Knowledge

```bash
# Find database-related decisions
memory-bank search "database choices" --limit 5

# Use enhanced search with detailed scoring
memory-bank search enhanced "authentication" --type decision
```

### 5. Start Using with Claude Code

```bash
# Run as MCP server (integrates with Claude Code)
memory-bank

# The server will provide semantic memory access to Claude Code
```

## 📦 Installation

### Option 1: Homebrew (macOS/Linux) ⭐ Recommended

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
# Prerequisites: Go 1.21+
go version

# Clone and build
git clone https://github.com/joern1811/memory-bank.git
cd memory-bank
go mod tidy
go build ./cmd/memory-bank

# Verify installation
./memory-bank --help
```

### Option 4: Direct Go Installation

```bash
go install github.com/joern1811/memory-bank/cmd/memory-bank@latest
```

### External Services (Optional)

For enhanced performance, install optional services:

```bash
# Install Ollama for local embeddings
curl -fsSL https://ollama.com/install.sh | sh
ollama pull nomic-embed-text

# Install ChromaDB for vector search
docker run -p 8000:8000 chromadb/chroma
# OR with uvx
uvx --from "chromadb[server]" chroma run --host 0.0.0.0 --port 8000
```

Memory Bank works out-of-the-box with mock providers when external services are unavailable.

## 🎯 Usage

### MCP Server Mode (Claude Code Integration)

```bash
# Start MCP server
memory-bank mcp serve

# Configure in Claude Code
# Add to ~/.config/claude-desktop/config.json:
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
```

### CLI Mode (Direct Usage)

```bash
# Project management
memory-bank init /path/to/project --name "Project Name"
memory-bank project list

# Memory operations
memory-bank memory create --type decision --title "Title" --content "Content"
memory-bank memory list --project "my-project" --type decision
memory-bank search "authentication patterns" --limit 10

# Session tracking
memory-bank session start "Implement user auth" --project "my-api"
memory-bank session log "Added JWT middleware" --type milestone
memory-bank session complete "Authentication implemented successfully"

# Advanced search
memory-bank search faceted "database" --types decision,pattern --facets
memory-bank search enhanced "JWT implementation" --threshold 0.8
memory-bank search suggestions "auth" --limit 10
```

### Memory Types

Memory Bank supports five types of development knowledge:

1. **Decision** (`decision`): Architectural and technical decisions with rationale
2. **Pattern** (`pattern`): Reusable code patterns and design solutions  
3. **Error Solution** (`error_solution`): Solutions to specific errors and problems
4. **Code** (`code`): Code snippets with context and explanations
5. **Documentation** (`documentation`): Project documentation and explanations

### Example Workflows

#### Capturing Architectural Decisions

```bash
memory-bank memory create --type decision \
  --title "Adopt microservices architecture" \
  --content "Decision to split monolith into microservices for better scalability. Considered complexity trade-offs but benefits outweigh costs for our team size." \
  --tags "architecture,microservices,scalability"
```

#### Storing Code Patterns

```bash
memory-bank memory create --type pattern \
  --title "Repository pattern with interfaces" \
  --content "type UserRepository interface { GetByID(id string) (*User, error) }
type userRepo struct { db *sql.DB }
func (r *userRepo) GetByID(id string) (*User, error) { /* implementation */ }" \
  --tags "go,repository,pattern,database"
```

#### Development Session Tracking

```bash
# Start a focused development session
memory-bank session start "Add user authentication" --project "my-api"

# Log progress with different entry types
memory-bank session log "Created user model" --type info
memory-bank session log "JWT middleware implemented" --type milestone
memory-bank session log "Tests failing on edge cases" --type issue
memory-bank session log "Added proper error handling" --type solution

# Complete the session
memory-bank session complete "User authentication fully implemented with tests"
```

## 🏗️ Architecture

Memory Bank uses hexagonal architecture for clean separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                    External Layer                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ MCP Server  │  │ CLI Client  │  │ HTTP API (Future)   │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                   Application Layer                         │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │ Memory Service │ Project Service │ Session Service     │ │
│  │ - CRUD Ops     │ - Mgmt & Init   │ - Progress Tracking │ │
│  │ - Search       │ - Auto-detect   │ - Session Lifecycle │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                     Domain Layer                            │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │ Entities: Memory, Project, Session                      │ │
│  │ Value Objects: MemoryID, EmbeddingVector, Tags         │ │
│  │ Business Logic: Validation, Search, Similarity         │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                 Infrastructure Layer                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ SQLite DB   │  │ Ollama      │  │ ChromaDB           │  │
│  │ + Mock      │  │ + Mock      │  │ + Mock             │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Core Principles

- **Domain-Driven Design**: Business logic isolated from technical concerns
- **Dependency Inversion**: High-level modules don't depend on low-level modules
- **Interface Segregation**: Small, focused interfaces for different concerns
- **Single Responsibility**: Each component has one reason to change
- **Open/Closed**: Open for extension, closed for modification

### Technology Stack

- **Language**: Go 1.21+ with modern language features
- **Database**: SQLite with automatic schema initialization and migrations
- **Embeddings**: Ollama (nomic-embed-text model) with Mock fallback
- **Vector Store**: ChromaDB (v2 API) with Mock fallback  
- **MCP**: github.com/mark3labs/mcp-go v0.32.0
- **CLI**: Cobra framework with comprehensive command structure
- **Logging**: Logrus with structured JSON logging

## 📚 Documentation

### User Documentation

- **[Getting Started Guide](docs/guides/getting-started.md)**: Complete setup and first steps
- **[Examples and Use Cases](docs/guides/examples.md)**: Real-world usage scenarios  
- **[Advanced Usage Guide](docs/guides/advanced-usage.md)**: Power user features and automation
- **[Troubleshooting Guide](docs/guides/troubleshooting.md)**: Common issues and solutions

### API Documentation

- **[MCP Methods Reference](docs/api/mcp-methods.md)**: Complete MCP protocol documentation
- **[CLI Commands Reference](docs/api/cli-commands.md)**: Full command-line interface guide

### Development Documentation

- **[Contributing Guidelines](CONTRIBUTING.md)**: How to contribute to the project
- **[Security Policy](SECURITY.md)**: Security practices and vulnerability reporting
- **[Changelog](CHANGELOG.md)**: Version history and migration guides
- **[Architecture Documentation](CLAUDE.md)**: Detailed technical architecture

### Quick Links

| Topic | Documentation | Description |
|-------|---------------|-------------|
| **First Time Setup** | [Getting Started](docs/guides/getting-started.md) | Complete installation and setup guide |
| **Memory Types** | [Getting Started - Memory Types](docs/guides/getting-started.md#understanding-memory-types) | When and how to use different memory types |
| **Claude Code Integration** | [Getting Started - MCP Integration](docs/guides/getting-started.md#integration-with-claude-code) | Set up Memory Bank with Claude Code |
| **Advanced Search** | [Advanced Usage - Search](docs/guides/advanced-usage.md#advanced-search-strategies) | Faceted search, relevance scoring, suggestions |
| **Automation** | [Advanced Usage - Automation](docs/guides/advanced-usage.md#advanced-automation) | Git hooks, CI/CD integration, workflows |
| **Troubleshooting** | [Troubleshooting Guide](docs/guides/troubleshooting.md) | Common issues and solutions |
| **API Reference** | [MCP Methods](docs/api/mcp-methods.md) | Complete MCP protocol reference |

## 🤝 Contributing

We welcome contributions! Memory Bank is built with clean architecture principles and comprehensive testing.

### Quick Contributing Guide

1. **Fork and Clone**: Fork the repository and clone locally
2. **Set Up Environment**: Install Go 1.21+ and run `go mod tidy`
3. **Make Changes**: Follow our coding standards and architecture principles
4. **Add Tests**: Write tests for new functionality
5. **Submit PR**: Create a pull request with clear description

### Development Setup

```bash
# Clone your fork
git clone https://github.com/yourusername/memory-bank.git
cd memory-bank

# Install dependencies and dev tools
go mod tidy
make dev-setup

# Run tests
make test
make test-integration

# Build and test locally
make build
./memory-bank --help
```

For detailed contributing guidelines, see [CONTRIBUTING.md](CONTRIBUTING.md).

## 🆘 Support

### Getting Help

1. **📖 Documentation**: Check our comprehensive documentation first
2. **🔍 Search Issues**: Look for similar issues in GitHub Issues
3. **❓ Ask Questions**: Create a GitHub Issue for questions
4. **🐛 Report Bugs**: Use our bug report template
5. **💡 Feature Requests**: Use our feature request template

### Community

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and community discussions
- **Pull Requests**: Code contributions and reviews

### Status and Compatibility

- **Latest Version**: v1.10.0
- **Go Compatibility**: Go 1.21+
- **Platform Support**: Linux, macOS, Windows
- **MCP Protocol**: 2024-11-05 specification
- **External Services**: Ollama v0.1.0+, ChromaDB v0.4.0+

---

## 📄 License

Memory Bank is released under the MIT License. See [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- **Hexagonal Architecture**: Inspired by Alistair Cockburn's ports and adapters pattern
- **MCP Protocol**: Built on the Model Context Protocol specification
- **Go Community**: Leveraging excellent Go libraries and patterns
- **Contributors**: Thank you to all contributors who help improve Memory Bank

---

**Built with ❤️ for developers who value knowledge management and semantic search.**
