package embedding

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/joern1811/memory-bank/internal/domain"
)

func TestDefaultOllamaConfig(t *testing.T) {
	config := DefaultOllamaConfig()

	if config.BaseURL != "http://localhost:11434" {
		t.Errorf("Expected BaseURL 'http://localhost:11434', got %s", config.BaseURL)
	}
	if config.Model != "nomic-embed-text" {
		t.Errorf("Expected Model 'nomic-embed-text', got %s", config.Model)
	}
	if config.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout 30s, got %v", config.Timeout)
	}
	if config.MaxConcurrentRequests != 5 {
		t.Errorf("Expected MaxConcurrentRequests 5, got %d", config.MaxConcurrentRequests)
	}
}

func TestNewOllamaProvider(t *testing.T) {
	logger := setupTestLogger()
	config := OllamaConfig{
		BaseURL:               "http://test.example.com",
		Model:                 "test-model",
		Timeout:               10 * time.Second,
		MaxConcurrentRequests: 3,
	}

	provider := NewOllamaProvider(config, logger)

	if provider == nil {
		t.Fatal("Expected non-nil provider")
	}
	if provider.baseURL != config.BaseURL {
		t.Errorf("Expected BaseURL %s, got %s", config.BaseURL, provider.baseURL)
	}
	if provider.model != config.Model {
		t.Errorf("Expected Model %s, got %s", config.Model, provider.model)
	}
	if provider.client == nil {
		t.Error("Expected non-nil HTTP client")
	}
	if provider.maxConcurrentRequests != config.MaxConcurrentRequests {
		t.Errorf("Expected MaxConcurrentRequests %d, got %d", config.MaxConcurrentRequests, provider.maxConcurrentRequests)
	}
	if len(provider.semaphore) != 0 || cap(provider.semaphore) != config.MaxConcurrentRequests {
		t.Errorf("Expected semaphore capacity %d, got %d", config.MaxConcurrentRequests, cap(provider.semaphore))
	}
}

func TestNewOllamaProvider_DefaultConfig(t *testing.T) {
	logger := setupTestLogger()
	config := OllamaConfig{} // Empty config should use defaults

	provider := NewOllamaProvider(config, logger)

	defaultConfig := DefaultOllamaConfig()
	if provider.baseURL != defaultConfig.BaseURL {
		t.Errorf("Expected default BaseURL %s, got %s", defaultConfig.BaseURL, provider.baseURL)
	}
	if provider.model != defaultConfig.Model {
		t.Errorf("Expected default Model %s, got %s", defaultConfig.Model, provider.model)
	}
}

func TestOllamaProvider_GenerateEmbedding_Success(t *testing.T) {
	// Create mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/embeddings" {
			t.Errorf("Expected path /api/embeddings, got %s", r.URL.Path)
		}

		// Verify Content-Type header
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Parse request body
		var req ollamaEmbeddingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		// Verify request content
		if req.Model != "test-model" {
			t.Errorf("Expected model 'test-model', got %s", req.Model)
		}
		if req.Prompt != "test text" {
			t.Errorf("Expected prompt 'test text', got %s", req.Prompt)
		}

		// Send mock response
		response := ollamaEmbeddingResponse{
			Embedding: []float64{0.1, 0.2, 0.3, 0.4, 0.5},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create provider with mock server URL
	logger := setupTestLogger()
	config := OllamaConfig{
		BaseURL:               server.URL,
		Model:                 "test-model",
		Timeout:               5 * time.Second,
		MaxConcurrentRequests: 1,
	}
	provider := NewOllamaProvider(config, logger)

	// Test embedding generation
	ctx := context.Background()
	embedding, err := provider.GenerateEmbedding(ctx, "test text")
	if err != nil {
		t.Fatalf("Failed to generate embedding: %v", err)
	}

	// Verify embedding
	expected := domain.EmbeddingVector{0.1, 0.2, 0.3, 0.4, 0.5}
	if len(embedding) != len(expected) {
		t.Errorf("Expected embedding length %d, got %d", len(expected), len(embedding))
	}
	for i, v := range expected {
		if embedding[i] != v {
			t.Errorf("Expected embedding[%d] = %f, got %f", i, v, embedding[i])
		}
	}
}

func TestOllamaProvider_GenerateEmbedding_ServerError(t *testing.T) {
	// Create mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	logger := setupTestLogger()
	config := OllamaConfig{
		BaseURL:               server.URL,
		Model:                 "test-model",
		Timeout:               5 * time.Second,
		MaxConcurrentRequests: 1,
	}
	provider := NewOllamaProvider(config, logger)

	ctx := context.Background()
	_, err := provider.GenerateEmbedding(ctx, "test text")
	if err == nil {
		t.Error("Expected error for server error response")
	}
	if !strings.Contains(err.Error(), "ollama API error") {
		t.Errorf("Expected 'ollama API error' in error message, got: %v", err)
	}
}

func TestOllamaProvider_GenerateEmbedding_OllamaError(t *testing.T) {
	// Create mock server that returns an Ollama error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ollamaEmbeddingResponse{
			Error: "Model not found",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := setupTestLogger()
	config := OllamaConfig{
		BaseURL:               server.URL,
		Model:                 "test-model",
		Timeout:               5 * time.Second,
		MaxConcurrentRequests: 1,
	}
	provider := NewOllamaProvider(config, logger)

	ctx := context.Background()
	_, err := provider.GenerateEmbedding(ctx, "test text")
	if err == nil {
		t.Error("Expected error for Ollama error response")
	}
	if !strings.Contains(err.Error(), "ollama error: Model not found") {
		t.Errorf("Expected 'ollama error: Model not found' in error message, got: %v", err)
	}
}

func TestOllamaProvider_GenerateEmbedding_InvalidJSON(t *testing.T) {
	// Create mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	logger := setupTestLogger()
	config := OllamaConfig{
		BaseURL:               server.URL,
		Model:                 "test-model",
		Timeout:               5 * time.Second,
		MaxConcurrentRequests: 1,
	}
	provider := NewOllamaProvider(config, logger)

	ctx := context.Background()
	_, err := provider.GenerateEmbedding(ctx, "test text")
	if err == nil {
		t.Error("Expected error for invalid JSON response")
	}
	if !strings.Contains(err.Error(), "failed to unmarshal response") {
		t.Errorf("Expected 'failed to unmarshal response' in error message, got: %v", err)
	}
}

func TestOllamaProvider_GenerateEmbedding_ContextCancellation(t *testing.T) {
	// Create mock server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		response := ollamaEmbeddingResponse{
			Embedding: []float64{0.1, 0.2, 0.3},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := setupTestLogger()
	config := OllamaConfig{
		BaseURL:               server.URL,
		Model:                 "test-model",
		Timeout:               5 * time.Second,
		MaxConcurrentRequests: 1,
	}
	provider := NewOllamaProvider(config, logger)

	// Create context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := provider.GenerateEmbedding(ctx, "test text")
	if err == nil {
		t.Error("Expected error for context cancellation")
	}
	if !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("Expected context deadline exceeded error, got: %v", err)
	}
}

func TestOllamaProvider_GenerateBatchEmbeddings_Success(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		var req ollamaEmbeddingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		// Return different embeddings based on prompt
		var embedding []float64
		switch req.Prompt {
		case "text 1":
			embedding = []float64{0.1, 0.2}
		case "text 2":
			embedding = []float64{0.3, 0.4}
		case "text 3":
			embedding = []float64{0.5, 0.6}
		default:
			embedding = []float64{0.0, 0.0}
		}

		response := ollamaEmbeddingResponse{
			Embedding: embedding,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := setupTestLogger()
	config := OllamaConfig{
		BaseURL:               server.URL,
		Model:                 "test-model",
		Timeout:               5 * time.Second,
		MaxConcurrentRequests: 2,
	}
	provider := NewOllamaProvider(config, logger)

	ctx := context.Background()
	texts := []string{"text 1", "text 2", "text 3"}
	embeddings, err := provider.GenerateBatchEmbeddings(ctx, texts)
	if err != nil {
		t.Fatalf("Failed to generate batch embeddings: %v", err)
	}

	// Verify embeddings
	if len(embeddings) != 3 {
		t.Errorf("Expected 3 embeddings, got %d", len(embeddings))
	}

	expectedEmbeddings := []domain.EmbeddingVector{
		{0.1, 0.2},
		{0.3, 0.4},
		{0.5, 0.6},
	}

	for i, expected := range expectedEmbeddings {
		if len(embeddings[i]) != len(expected) {
			t.Errorf("Expected embedding %d length %d, got %d", i, len(expected), len(embeddings[i]))
			continue
		}
		for j, v := range expected {
			if embeddings[i][j] != v {
				t.Errorf("Expected embedding[%d][%d] = %f, got %f", i, j, v, embeddings[i][j])
			}
		}
	}

	// Verify that requests were made (should be 3 for 3 texts)
	if requestCount != 3 {
		t.Errorf("Expected 3 requests, got %d", requestCount)
	}
}

func TestOllamaProvider_GenerateBatchEmbeddings_Empty(t *testing.T) {
	logger := setupTestLogger()
	config := DefaultOllamaConfig()
	provider := NewOllamaProvider(config, logger)

	ctx := context.Background()
	embeddings, err := provider.GenerateBatchEmbeddings(ctx, []string{})
	if err != nil {
		t.Errorf("Expected no error for empty batch, got: %v", err)
	}
	if embeddings != nil {
		t.Errorf("Expected nil embeddings for empty batch, got: %v", embeddings)
	}
}

func TestOllamaProvider_GenerateBatchEmbeddings_PartialFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ollamaEmbeddingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		// Fail on "text 2"
		if req.Prompt == "text 2" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server error"))
			return
		}

		response := ollamaEmbeddingResponse{
			Embedding: []float64{0.1, 0.2},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := setupTestLogger()
	config := OllamaConfig{
		BaseURL:               server.URL,
		Model:                 "test-model",
		Timeout:               5 * time.Second,
		MaxConcurrentRequests: 2,
	}
	provider := NewOllamaProvider(config, logger)

	ctx := context.Background()
	texts := []string{"text 1", "text 2", "text 3"}
	_, err := provider.GenerateBatchEmbeddings(ctx, texts)
	if err == nil {
		t.Error("Expected error for partial batch failure")
	}
	if !strings.Contains(err.Error(), "failed to generate embedding for text 1") {
		t.Errorf("Expected error message about text 1, got: %v", err)
	}
}

func TestOllamaProvider_GetDimensions(t *testing.T) {
	logger := setupTestLogger()

	tests := []struct {
		model       string
		expectedDim int
	}{
		{"nomic-embed-text", 768},
		{"all-minilm", 384},
		{"unknown-model", 768}, // Default
	}

	for _, test := range tests {
		config := OllamaConfig{
			BaseURL: "http://localhost:11434",
			Model:   test.model,
		}
		provider := NewOllamaProvider(config, logger)

		dimensions := provider.GetDimensions()
		if dimensions != test.expectedDim {
			t.Errorf("For model %s, expected dimensions %d, got %d", test.model, test.expectedDim, dimensions)
		}
	}
}

func TestOllamaProvider_GetModelName(t *testing.T) {
	logger := setupTestLogger()
	config := OllamaConfig{
		BaseURL: "http://localhost:11434",
		Model:   "test-model-name",
	}
	provider := NewOllamaProvider(config, logger)

	modelName := provider.GetModelName()
	if modelName != "test-model-name" {
		t.Errorf("Expected model name 'test-model-name', got %s", modelName)
	}
}

func TestOllamaProvider_HealthCheck_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ollamaEmbeddingResponse{
			Embedding: []float64{0.1, 0.2, 0.3},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := setupTestLogger()
	config := OllamaConfig{
		BaseURL:               server.URL,
		Model:                 "test-model",
		Timeout:               5 * time.Second,
		MaxConcurrentRequests: 1,
	}
	provider := NewOllamaProvider(config, logger)

	ctx := context.Background()
	err := provider.HealthCheck(ctx)
	if err != nil {
		t.Errorf("Expected health check to pass, got error: %v", err)
	}
}

func TestOllamaProvider_HealthCheck_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Server error"))
	}))
	defer server.Close()

	logger := setupTestLogger()
	config := OllamaConfig{
		BaseURL:               server.URL,
		Model:                 "test-model",
		Timeout:               5 * time.Second,
		MaxConcurrentRequests: 1,
	}
	provider := NewOllamaProvider(config, logger)

	ctx := context.Background()
	err := provider.HealthCheck(ctx)
	if err == nil {
		t.Error("Expected health check to fail")
	}
	if !strings.Contains(err.Error(), "ollama health check failed") {
		t.Errorf("Expected 'ollama health check failed' in error message, got: %v", err)
	}
}

func TestOllamaProvider_ConcurrencyControl(t *testing.T) {
	requestTimes := make([]time.Time, 0)
	var requestMutex sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestMutex.Lock()
		requestTimes = append(requestTimes, time.Now())
		requestMutex.Unlock()

		// Simulate processing time
		time.Sleep(50 * time.Millisecond)

		response := ollamaEmbeddingResponse{
			Embedding: []float64{0.1, 0.2},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := setupTestLogger()
	config := OllamaConfig{
		BaseURL:               server.URL,
		Model:                 "test-model",
		Timeout:               5 * time.Second,
		MaxConcurrentRequests: 2, // Limit to 2 concurrent requests
	}
	provider := NewOllamaProvider(config, logger)

	ctx := context.Background()
	// Send 4 requests - should be processed in 2 batches
	texts := []string{"text 1", "text 2", "text 3", "text 4"}

	start := time.Now()
	_, err := provider.GenerateBatchEmbeddings(ctx, texts)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to generate batch embeddings: %v", err)
	}

	// With concurrency limit of 2 and 50ms processing time per request,
	// 4 requests should take at least 100ms (2 batches)
	minExpectedDuration := 100 * time.Millisecond
	if duration < minExpectedDuration {
		t.Errorf("Expected duration >= %v, got %v (concurrency control may not be working)", minExpectedDuration, duration)
	}

	// Verify that no more than 2 requests started at the same time
	requestMutex.Lock()
	defer requestMutex.Unlock()

	if len(requestTimes) != 4 {
		t.Errorf("Expected 4 requests, got %d", len(requestTimes))
	}
}

// Benchmark tests
func BenchmarkOllamaProvider_GenerateEmbedding(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ollamaEmbeddingResponse{
			Embedding: make([]float64, 768), // Simulate realistic embedding size
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := setupTestLogger()
	config := OllamaConfig{
		BaseURL:               server.URL,
		Model:                 "test-model",
		Timeout:               30 * time.Second,
		MaxConcurrentRequests: 5,
	}
	provider := NewOllamaProvider(config, logger)

	ctx := context.Background()
	text := "This is a test text for benchmarking embedding generation performance."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.GenerateEmbedding(ctx, text)
		if err != nil {
			b.Fatalf("Failed to generate embedding: %v", err)
		}
	}
}

func BenchmarkOllamaProvider_GenerateBatchEmbeddings(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ollamaEmbeddingResponse{
			Embedding: make([]float64, 768),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := setupTestLogger()
	config := OllamaConfig{
		BaseURL:               server.URL,
		Model:                 "test-model",
		Timeout:               30 * time.Second,
		MaxConcurrentRequests: 5,
	}
	provider := NewOllamaProvider(config, logger)

	ctx := context.Background()
	texts := []string{
		"First test text for batch embedding generation.",
		"Second test text for batch embedding generation.",
		"Third test text for batch embedding generation.",
		"Fourth test text for batch embedding generation.",
		"Fifth test text for batch embedding generation.",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.GenerateBatchEmbeddings(ctx, texts)
		if err != nil {
			b.Fatalf("Failed to generate batch embeddings: %v", err)
		}
	}
}
