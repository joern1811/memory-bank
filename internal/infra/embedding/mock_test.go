package embedding

import (
	"context"
	"fmt"
	"testing"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/sirupsen/logrus"
)

func setupTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests
	return logger
}

func TestNewMockEmbeddingProvider(t *testing.T) {
	logger := setupTestLogger()
	dimensions := 384

	provider := NewMockEmbeddingProvider(dimensions, logger)

	if provider == nil {
		t.Fatal("Expected non-nil provider")
	}
	if provider.dimensions != dimensions {
		t.Errorf("Expected dimensions %d, got %d", dimensions, provider.dimensions)
	}
	if provider.logger != logger {
		t.Error("Expected logger to be set")
	}
}

func TestMockEmbeddingProvider_GenerateEmbedding(t *testing.T) {
	logger := setupTestLogger()
	dimensions := 100
	provider := NewMockEmbeddingProvider(dimensions, logger)

	ctx := context.Background()
	text := "test text"

	embedding, err := provider.GenerateEmbedding(ctx, text)
	if err != nil {
		t.Fatalf("Failed to generate mock embedding: %v", err)
	}

	// Verify embedding properties
	if len(embedding) != dimensions {
		t.Errorf("Expected embedding length %d, got %d", dimensions, len(embedding))
	}

	// Verify all values are in valid range [0, 1]
	for i, value := range embedding {
		if value < 0 || value > 1 {
			t.Errorf("Expected embedding[%d] to be in range [0, 1], got %f", i, value)
		}
	}

	// Test deterministic behavior - same text should produce same embedding
	embedding2, err := provider.GenerateEmbedding(ctx, text)
	if err != nil {
		t.Fatalf("Failed to generate second mock embedding: %v", err)
	}

	if len(embedding2) != len(embedding) {
		t.Errorf("Expected same embedding length for same text")
	}

	for i := range embedding {
		if embedding[i] != embedding2[i] {
			t.Errorf("Expected deterministic embedding, but embedding[%d] differs: %f != %f", i, embedding[i], embedding2[i])
		}
	}
}

func TestMockEmbeddingProvider_GenerateEmbedding_DifferentTexts(t *testing.T) {
	logger := setupTestLogger()
	dimensions := 50
	provider := NewMockEmbeddingProvider(dimensions, logger)

	ctx := context.Background()
	text1 := "hello world"
	text2 := "goodbye world"

	embedding1, err := provider.GenerateEmbedding(ctx, text1)
	if err != nil {
		t.Fatalf("Failed to generate embedding for text1: %v", err)
	}

	embedding2, err := provider.GenerateEmbedding(ctx, text2)
	if err != nil {
		t.Fatalf("Failed to generate embedding for text2: %v", err)
	}

	// Different texts should produce different embeddings
	identical := true
	for i := range embedding1 {
		if embedding1[i] != embedding2[i] {
			identical = false
			break
		}
	}

	if identical {
		t.Error("Expected different embeddings for different texts")
	}
}

func TestMockEmbeddingProvider_GenerateEmbedding_EmptyText(t *testing.T) {
	logger := setupTestLogger()
	dimensions := 10
	provider := NewMockEmbeddingProvider(dimensions, logger)

	ctx := context.Background()
	embedding, err := provider.GenerateEmbedding(ctx, "")
	if err != nil {
		t.Fatalf("Failed to generate embedding for empty text: %v", err)
	}

	if len(embedding) != dimensions {
		t.Errorf("Expected embedding length %d for empty text, got %d", dimensions, len(embedding))
	}

	// Empty text should still produce valid values
	for i, value := range embedding {
		if value < 0 || value > 1 {
			t.Errorf("Expected embedding[%d] to be in range [0, 1] for empty text, got %f", i, value)
		}
	}
}

func TestMockEmbeddingProvider_GenerateBatchEmbeddings(t *testing.T) {
	logger := setupTestLogger()
	dimensions := 25
	provider := NewMockEmbeddingProvider(dimensions, logger)

	ctx := context.Background()
	texts := []string{
		"first text",
		"second text",
		"third text",
	}

	embeddings, err := provider.GenerateBatchEmbeddings(ctx, texts)
	if err != nil {
		t.Fatalf("Failed to generate batch embeddings: %v", err)
	}

	// Verify batch properties
	if len(embeddings) != len(texts) {
		t.Errorf("Expected %d embeddings, got %d", len(texts), len(embeddings))
	}

	// Verify each embedding
	for i, embedding := range embeddings {
		if len(embedding) != dimensions {
			t.Errorf("Expected embedding %d length %d, got %d", i, dimensions, len(embedding))
		}

		// Verify values are in valid range
		for j, value := range embedding {
			if value < 0 || value > 1 {
				t.Errorf("Expected embedding[%d][%d] to be in range [0, 1], got %f", i, j, value)
			}
		}
	}

	// Verify consistency - generating individual embeddings should match batch
	for i, text := range texts {
		individualEmbedding, err := provider.GenerateEmbedding(ctx, text)
		if err != nil {
			t.Fatalf("Failed to generate individual embedding for text %d: %v", i, err)
		}

		batchEmbedding := embeddings[i]
		if len(individualEmbedding) != len(batchEmbedding) {
			t.Errorf("Embedding %d length mismatch: individual=%d, batch=%d", i, len(individualEmbedding), len(batchEmbedding))
			continue
		}

		for j := range individualEmbedding {
			if individualEmbedding[j] != batchEmbedding[j] {
				t.Errorf("Embedding %d mismatch at position %d: individual=%f, batch=%f", i, j, individualEmbedding[j], batchEmbedding[j])
			}
		}
	}
}

func TestMockEmbeddingProvider_GenerateBatchEmbeddings_Empty(t *testing.T) {
	logger := setupTestLogger()
	provider := NewMockEmbeddingProvider(100, logger)

	ctx := context.Background()
	embeddings, err := provider.GenerateBatchEmbeddings(ctx, []string{})
	if err != nil {
		t.Errorf("Expected no error for empty batch, got: %v", err)
	}
	if len(embeddings) != 0 {
		t.Errorf("Expected empty result for empty batch, got %d embeddings", len(embeddings))
	}
}

func TestMockEmbeddingProvider_GetDimensions(t *testing.T) {
	logger := setupTestLogger()
	
	testCases := []int{1, 50, 384, 768, 1536}
	
	for _, expectedDim := range testCases {
		provider := NewMockEmbeddingProvider(expectedDim, logger)
		actualDim := provider.GetDimensions()
		
		if actualDim != expectedDim {
			t.Errorf("Expected dimensions %d, got %d", expectedDim, actualDim)
		}
	}
}

func TestMockEmbeddingProvider_GetModelName(t *testing.T) {
	logger := setupTestLogger()
	provider := NewMockEmbeddingProvider(100, logger)

	modelName := provider.GetModelName()
	expectedName := "mock-embedding-model"
	
	if modelName != expectedName {
		t.Errorf("Expected model name '%s', got '%s'", expectedName, modelName)
	}
}

func TestSimpleHash_Consistency(t *testing.T) {
	testCases := []struct {
		text string
		expectedHash int
	}{
		{"", 0},
		{"a", 97},
		{"hello", 99162322},
		{"Hello", 69609650}, // Different case should produce different hash
	}

	for _, tc := range testCases {
		hash := simpleHash(tc.text)
		if hash != tc.expectedHash {
			t.Errorf("For text '%s', expected hash %d, got %d", tc.text, tc.expectedHash, hash)
		}
	}
}

func TestSimpleHash_Deterministic(t *testing.T) {
	text := "test string for hash consistency"
	
	hash1 := simpleHash(text)
	hash2 := simpleHash(text)
	
	if hash1 != hash2 {
		t.Errorf("Expected consistent hash for same text, got %d and %d", hash1, hash2)
	}
}

func TestSimpleHash_Different(t *testing.T) {
	texts := []string{
		"text1",
		"text2", 
		"different text",
		"yet another text",
		"completely different content",
	}
	
	hashes := make(map[int]string)
	
	for _, text := range texts {
		hash := simpleHash(text)
		
		// Check for collisions (unlikely but possible)
		if existingText, exists := hashes[hash]; exists {
			t.Logf("Hash collision detected: '%s' and '%s' both hash to %d", text, existingText, hash)
		} else {
			hashes[hash] = text
		}
	}
	
	// We should have at least some different hashes
	if len(hashes) < 3 {
		t.Errorf("Expected at least 3 different hashes, got %d", len(hashes))
	}
}

// Test integration with domain types
func TestMockEmbeddingProvider_DomainIntegration(t *testing.T) {
	logger := setupTestLogger()
	provider := NewMockEmbeddingProvider(768, logger)

	ctx := context.Background()
	
	// Test with different memory types content
	testTexts := []string{
		"Authentication decision: Use JWT tokens for API security",
		"Repository pattern: Implement data access layer abstraction", 
		"Error solution: NullPointerException when accessing user.getName()",
		"Code snippet: func validateUser(user *User) error { return nil }",
	}

	for _, text := range testTexts {
		embedding, err := provider.GenerateEmbedding(ctx, text)
		if err != nil {
			t.Errorf("Failed to generate embedding for text '%s': %v", text, err)
			continue
		}

		// Verify embedding is valid domain.EmbeddingVector
		var _ domain.EmbeddingVector = embedding
		
		if len(embedding) != 768 {
			t.Errorf("Expected 768-dimensional embedding, got %d", len(embedding))
		}

		// Verify embedding values are reasonable
		for i, value := range embedding {
			if value < 0 || value > 1 {
				t.Errorf("Embedding value out of range at position %d: %f", i, value)
				break
			}
		}
	}
}

// Benchmark tests for performance characteristics
func BenchmarkMockEmbeddingProvider_GenerateEmbedding(b *testing.B) {
	logger := setupTestLogger()
	provider := NewMockEmbeddingProvider(768, logger)

	ctx := context.Background()
	text := "This is a benchmark test for mock embedding generation performance."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.GenerateEmbedding(ctx, text)
		if err != nil {
			b.Fatalf("Failed to generate embedding: %v", err)
		}
	}
}

func BenchmarkMockEmbeddingProvider_GenerateBatchEmbeddings(b *testing.B) {
	logger := setupTestLogger()
	provider := NewMockEmbeddingProvider(768, logger)

	ctx := context.Background()
	texts := make([]string, 100)
	for i := range texts {
		texts[i] = fmt.Sprintf("Benchmark text number %d for batch embedding generation.", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.GenerateBatchEmbeddings(ctx, texts)
		if err != nil {
			b.Fatalf("Failed to generate batch embeddings: %v", err)
		}
	}
}

func BenchmarkSimpleHash(b *testing.B) {
	text := "This is a test string for benchmarking the simple hash function performance."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = simpleHash(text)
	}
}