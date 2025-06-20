package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/sirupsen/logrus"
)

// OllamaProvider implements the EmbeddingProvider interface using Ollama
type OllamaProvider struct {
	baseURL               string
	model                 string
	client                *http.Client
	logger                *logrus.Logger
	maxConcurrentRequests int
	semaphore             chan struct{}
}

// OllamaConfig holds configuration for Ollama provider
type OllamaConfig struct {
	BaseURL               string        `json:"base_url"`
	Model                 string        `json:"model"`
	Timeout               time.Duration `json:"timeout"`
	MaxConcurrentRequests int           `json:"max_concurrent_requests"`
}

// DefaultOllamaConfig returns default configuration for Ollama
func DefaultOllamaConfig() OllamaConfig {
	return OllamaConfig{
		BaseURL:               "http://localhost:11434",
		Model:                 "nomic-embed-text",
		Timeout:               30 * time.Second,
		MaxConcurrentRequests: 5,
	}
}

// NewOllamaProvider creates a new Ollama embedding provider
func NewOllamaProvider(config OllamaConfig, logger *logrus.Logger) *OllamaProvider {
	if config.BaseURL == "" {
		config = DefaultOllamaConfig()
	}

	// Configure HTTP client with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	return &OllamaProvider{
		baseURL: config.BaseURL,
		model:   config.Model,
		client: &http.Client{
			Timeout:   config.Timeout,
			Transport: transport,
		},
		logger:                logger,
		maxConcurrentRequests: config.MaxConcurrentRequests,
		semaphore:             make(chan struct{}, config.MaxConcurrentRequests),
	}
}

// ollamaEmbeddingRequest represents a request to Ollama's embedding API
type ollamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// ollamaEmbeddingResponse represents a response from Ollama's embedding API
type ollamaEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
	Error     string    `json:"error,omitempty"`
}

// GenerateEmbedding generates an embedding for a single text
func (p *OllamaProvider) GenerateEmbedding(ctx context.Context, text string) (domain.EmbeddingVector, error) {
	p.logger.WithFields(logrus.Fields{
		"model":       p.model,
		"text_length": len(text),
	}).Debug("Generating embedding")

	// Prepare request
	reqBody := ollamaEmbeddingRequest{
		Model:  p.model,
		Prompt: text,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/embeddings", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			p.logger.WithError(err).Warn("Failed to close response body")
		}
	}()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var embeddingResp ollamaEmbeddingResponse
	if err := json.Unmarshal(body, &embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if embeddingResp.Error != "" {
		return nil, fmt.Errorf("ollama error: %s", embeddingResp.Error)
	}

	// Convert float64 to float32
	embedding := make(domain.EmbeddingVector, len(embeddingResp.Embedding))
	for i, v := range embeddingResp.Embedding {
		embedding[i] = float32(v)
	}

	p.logger.WithField("dimensions", len(embedding)).Debug("Embedding generated successfully")
	return embedding, nil
}

// GenerateBatchEmbeddings generates embeddings for multiple texts with concurrent processing
func (p *OllamaProvider) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([]domain.EmbeddingVector, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	p.logger.WithFields(logrus.Fields{
		"model":       p.model,
		"batch_size":  len(texts),
		"concurrency": p.maxConcurrentRequests,
	}).Debug("Generating batch embeddings")

	embeddings := make([]domain.EmbeddingVector, len(texts))
	errors := make([]error, len(texts))

	var wg sync.WaitGroup
	for i, text := range texts {
		wg.Add(1)
		go func(index int, content string) {
			defer wg.Done()

			// Acquire semaphore for concurrency control
			p.semaphore <- struct{}{}
			defer func() { <-p.semaphore }()

			embedding, err := p.GenerateEmbedding(ctx, content)
			embeddings[index] = embedding
			errors[index] = err
		}(i, text)
	}

	wg.Wait()

	// Check for any errors
	for i, err := range errors {
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
		}
	}

	p.logger.WithField("batch_size", len(embeddings)).Debug("Batch embeddings generated successfully")
	return embeddings, nil
}

// GetDimensions returns the dimension size of embeddings from this provider
func (p *OllamaProvider) GetDimensions() int {
	// nomic-embed-text produces 768-dimensional embeddings
	switch p.model {
	case "nomic-embed-text":
		return 768
	case "all-minilm":
		return 384
	default:
		return 768 // Default assumption
	}
}

// GetModelName returns the model name being used
func (p *OllamaProvider) GetModelName() string {
	return p.model
}

// HealthCheck verifies that Ollama is running and the model is available
func (p *OllamaProvider) HealthCheck(ctx context.Context) error {
	p.logger.Debug("Performing Ollama health check")

	// Test with a simple embedding
	_, err := p.GenerateEmbedding(ctx, "test")
	if err != nil {
		return fmt.Errorf("ollama health check failed: %w", err)
	}

	p.logger.Info("Ollama health check passed")
	return nil
}

// MockEmbeddingProvider provides a mock implementation for testing
type MockEmbeddingProvider struct {
	dimensions int
	logger     *logrus.Logger
}

// NewMockEmbeddingProvider creates a new mock embedding provider
func NewMockEmbeddingProvider(dimensions int, logger *logrus.Logger) *MockEmbeddingProvider {
	return &MockEmbeddingProvider{
		dimensions: dimensions,
		logger:     logger,
	}
}

// GenerateEmbedding generates a mock embedding
func (m *MockEmbeddingProvider) GenerateEmbedding(ctx context.Context, text string) (domain.EmbeddingVector, error) {
	m.logger.Debug("Generating mock embedding")

	// Generate deterministic mock embedding based on text hash
	embedding := make(domain.EmbeddingVector, m.dimensions)
	hash := simpleHash(text)

	for i := range embedding {
		// Ensure positive values by using absolute value and proper modulo
		value := ((hash + i) % 100)
		if value < 0 {
			value = -value
		}
		embedding[i] = float32(value) / 100.0
	}

	return embedding, nil
}

// GenerateBatchEmbeddings generates mock embeddings for multiple texts
func (m *MockEmbeddingProvider) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([]domain.EmbeddingVector, error) {
	embeddings := make([]domain.EmbeddingVector, len(texts))

	for i, text := range texts {
		embedding, err := m.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}

// GetDimensions returns the dimension size
func (m *MockEmbeddingProvider) GetDimensions() int {
	return m.dimensions
}

// GetModelName returns the mock model name
func (m *MockEmbeddingProvider) GetModelName() string {
	return "mock-embedding-model"
}

// simpleHash generates a simple hash for deterministic mock embeddings
func simpleHash(s string) int {
	hash := 0
	for _, char := range s {
		hash = hash*31 + int(char)
	}
	return hash
}
