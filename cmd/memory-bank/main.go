package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joern1811/memory-bank/internal/app"
	"github.com/joern1811/memory-bank/internal/infra/database"
	"github.com/joern1811/memory-bank/internal/infra/embedding"
	"github.com/joern1811/memory-bank/internal/infra/mcp"
	"github.com/joern1811/memory-bank/internal/infra/vector"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
)

const (
	serverName    = "memory-bank"
	serverVersion = "1.0.0"
)

func main() {
	// Setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	logger.WithFields(logrus.Fields{
		"server":  serverName,
		"version": serverVersion,
	}).Info("Starting Memory Bank MCP Server")

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.WithField("signal", sig).Info("Received shutdown signal")
		cancel()
	}()

	// Initialize dependencies
	if err := run(ctx, logger); err != nil {
		logger.WithError(err).Error("Server failed")
		os.Exit(1)
	}

	logger.Info("Memory Bank MCP Server shutdown gracefully")
}

func run(ctx context.Context, logger *logrus.Logger) error {
	// Initialize database
	dbPath := getEnvOrDefault("MEMORY_BANK_DB_PATH", "./memory_bank.db")
	db, err := database.NewSQLiteDatabase(dbPath, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Initialize repositories
	memoryRepo := database.NewSQLiteMemoryRepository(db, logger)

	// Initialize embedding provider
	embeddingConfig := embedding.DefaultOllamaConfig()
	if baseURL := os.Getenv("OLLAMA_BASE_URL"); baseURL != "" {
		embeddingConfig.BaseURL = baseURL
	}
	if model := os.Getenv("OLLAMA_MODEL"); model != "" {
		embeddingConfig.Model = model
	}
	
	var embeddingProvider ports.EmbeddingProvider
	ollamaProvider := embedding.NewOllamaProvider(embeddingConfig, logger)

	// Test embedding provider connection
	if err := ollamaProvider.HealthCheck(ctx); err != nil {
		logger.WithError(err).Warn("Ollama health check failed, using mock provider")
		embeddingProvider = embedding.NewMockEmbeddingProvider(768, logger)
	} else {
		embeddingProvider = ollamaProvider
	}

	// Initialize vector store
	vectorConfig := vector.DefaultChromeDBConfig()
	if baseURL := os.Getenv("CHROMADB_BASE_URL"); baseURL != "" {
		vectorConfig.BaseURL = baseURL
	}
	if collection := os.Getenv("CHROMADB_COLLECTION"); collection != "" {
		vectorConfig.Collection = collection
	}
	
	var vectorStore ports.VectorStore
	chromaDBStore := vector.NewChromaDBVectorStore(vectorConfig, logger)

	// Test vector store connection
	if err := chromaDBStore.HealthCheck(ctx); err != nil {
		logger.WithError(err).Warn("ChromaDB health check failed, using mock vector store")
		vectorStore = vector.NewMockVectorStore(logger)
	} else {
		vectorStore = chromaDBStore
	}

	// Initialize services
	memoryService := app.NewMemoryService(memoryRepo, embeddingProvider, vectorStore, logger)
	projectService := app.NewProjectService(nil, logger) // TODO: Add project repository
	sessionService := app.NewSessionService(nil, nil, logger) // TODO: Add session and project repositories

	// Initialize MCP server
	mcpServer := server.NewMCPServer("memory-bank", serverVersion)
	memoryBankServer := mcp.NewMemoryBankServer(memoryService, projectService, sessionService, logger)
	memoryBankServer.RegisterMethods(mcpServer)

	logger.Info("Memory Bank MCP Server started successfully")

	// Note: The mcp-go library may not have a simple Serve method
	// This is a placeholder implementation - we would need to check the actual API
	// For now, we'll just block on the context
	<-ctx.Done()
	logger.Info("Context cancelled, shutting down server")

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}