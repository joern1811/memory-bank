package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/joern1811/memory-bank/internal/app"
	"github.com/joern1811/memory-bank/internal/infra/config"
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
	Config         *config.Config
}

// NewServiceContainer creates a new service container with all dependencies
func NewServiceContainer() (*ServiceContainer, error) {
	return NewServiceContainerWithConfig("")
}

// NewServiceContainerWithConfig creates a service container with specified config file
func NewServiceContainerWithConfig(configPath string) (*ServiceContainer, error) {
	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Validate configuration
	if err := cfg.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Initialize logger with config
	logger := logrus.New()
	
	// Set log format
	if cfg.Logging.Format == "text" {
		logger.SetFormatter(&logrus.TextFormatter{})
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}
	
	// Set log level
	switch strings.ToLower(cfg.Logging.Level) {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// Initialize database using config
	db, err := database.NewSQLiteDatabase(cfg.Database.Path, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize repositories
	memoryRepo := database.NewSQLiteMemoryRepository(db, logger)
	sessionRepo := database.NewSQLiteSessionRepository(db, logger)
	projectRepo := database.NewSQLiteProjectRepository(db, logger)

	// Initialize embedding provider using config
	ollamaConfig := embedding.OllamaConfig{
		BaseURL: cfg.Ollama.BaseURL,
		Model:   cfg.Ollama.Model,
	}
	ollamaProvider := embedding.NewOllamaProvider(ollamaConfig, logger)

	// Check Ollama health
	ctx := context.Background()
	var embeddingProvider ports.EmbeddingProvider = ollamaProvider
	if err := ollamaProvider.HealthCheck(ctx); err != nil {
		logger.Warn("Ollama is not available, falling back to mock provider")
		embeddingProvider = embedding.NewMockEmbeddingProvider(768, logger)
	}

	// Initialize vector store using config
	chromaConfig := vector.ChromaDBConfig{
		BaseURL:    cfg.ChromaDB.BaseURL,
		Collection: cfg.ChromaDB.Collection,
		Tenant:     cfg.ChromaDB.Tenant,
		Database:   cfg.ChromaDB.Database,
		Timeout:    time.Duration(cfg.ChromaDB.Timeout) * time.Second,
	}
	chromaStore := vector.NewChromaDBVectorStore(chromaConfig, logger)

	var vectorStore ports.VectorStore = chromaStore
	if err := chromaStore.HealthCheck(ctx); err != nil {
		logger.Warn("ChromaDB is not available, falling back to mock vector store")
		vectorStore = vector.NewMockVectorStore(logger)
	}

	// Initialize services
	memoryService := app.NewMemoryService(memoryRepo, embeddingProvider, vectorStore, logger)
	projectService := app.NewProjectService(projectRepo, logger)
	sessionService := app.NewSessionService(sessionRepo, projectRepo, logger)

	return &ServiceContainer{
		MemoryService:  memoryService,
		ProjectService: projectService,
		SessionService: sessionService,
		Logger:         logger,
		Config:         cfg,
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

