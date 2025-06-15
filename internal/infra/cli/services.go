package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joern1811/memory-bank/internal/app"
	"github.com/joern1811/memory-bank/internal/infra/database"
	"github.com/joern1811/memory-bank/internal/infra/embedding"
	"github.com/joern1811/memory-bank/internal/infra/vector"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/sirupsen/logrus"
)

// ServiceContainer holds all application services
type ServiceContainer struct {
	MemoryService  *app.MemoryService
	ProjectService *app.ProjectService
	SessionService *app.SessionService
	Logger         *logrus.Logger
}

// NewServiceContainer creates a new service container with all dependencies
func NewServiceContainer() (*ServiceContainer, error) {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Get database path from environment or use default
	dbPath := os.Getenv("MEMORY_BANK_DB_PATH")
	if dbPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		dbPath = filepath.Join(homeDir, ".memory_bank.db")
	}

	// Initialize database
	db, err := database.NewSQLiteDatabase(dbPath, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize memory repository
	memoryRepo := database.NewSQLiteMemoryRepository(db, logger)

	// Initialize embedding provider (Ollama with Mock fallback)
	ollamaConfig := embedding.OllamaConfig{
		BaseURL: getEnvOrDefault("OLLAMA_BASE_URL", "http://localhost:11434"),
		Model:   getEnvOrDefault("OLLAMA_MODEL", "nomic-embed-text"),
	}
	ollamaProvider := embedding.NewOllamaProvider(ollamaConfig, logger)

	// Check Ollama health
	ctx := context.Background()
	var embeddingProvider ports.EmbeddingProvider = ollamaProvider
	if err := ollamaProvider.HealthCheck(ctx); err != nil {
		logger.Warn("Ollama is not available, falling back to mock provider")
		embeddingProvider = embedding.NewMockEmbeddingProvider(768, logger)
	}

	// Initialize vector store (ChromaDB with Mock fallback)
	chromaConfig := vector.ChromaDBConfig{
		BaseURL:    getEnvOrDefault("CHROMADB_BASE_URL", "http://localhost:8000"),
		Collection: getEnvOrDefault("CHROMADB_COLLECTION", "memory_bank"),
	}
	chromaStore := vector.NewChromaDBVectorStore(chromaConfig, logger)

	var vectorStore ports.VectorStore = chromaStore
	if err := chromaStore.HealthCheck(ctx); err != nil {
		logger.Warn("ChromaDB is not available, falling back to mock vector store")
		vectorStore = vector.NewMockVectorStore(logger)
	}

	// Initialize services
	memoryService := app.NewMemoryService(memoryRepo, embeddingProvider, vectorStore, logger)
	projectService := app.NewProjectService(nil, logger) // TODO: Implement project repository
	sessionService := app.NewSessionService(nil, nil, logger) // TODO: Implement repositories

	return &ServiceContainer{
		MemoryService:  memoryService,
		ProjectService: projectService,
		SessionService: sessionService,
		Logger:         logger,
	}, nil
}

// Global service container instance
var services *ServiceContainer

// GetServices returns the global service container instance
func GetServices() (*ServiceContainer, error) {
	if services == nil {
		var err error
		services, err = NewServiceContainer()
		if err != nil {
			return nil, err
		}
	}
	return services, nil
}

