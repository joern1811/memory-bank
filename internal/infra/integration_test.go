package infra

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joern1811/memory-bank/internal/infra/config"
	"github.com/joern1811/memory-bank/internal/infra/embedding"
	"github.com/joern1811/memory-bank/internal/infra/vector"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOllamaIntegration tests real Ollama integration
func TestOllamaIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Load configuration
	cfg, err := config.LoadConfig("")
	require.NoError(t, err)

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Reduce noise in tests

	// Create Ollama provider
	ollamaConfig := embedding.OllamaConfig{
		BaseURL: cfg.Ollama.BaseURL,
		Model:   cfg.Ollama.Model,
	}
	provider := embedding.NewOllamaProvider(ollamaConfig, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("HealthCheck", func(t *testing.T) {
		err := provider.HealthCheck(ctx)
		if err != nil {
			t.Skipf("Ollama not available: %v", err)
		}
	})

	t.Run("GenerateEmbedding", func(t *testing.T) {
		if err := provider.HealthCheck(ctx); err != nil {
			t.Skipf("Ollama not available: %v", err)
		}

		// Test generating embedding
		text := "This is a test document for embedding generation"
		embedding, err := provider.GenerateEmbedding(ctx, text)
		require.NoError(t, err)
		
		// Verify embedding properties
		assert.NotNil(t, embedding)
		assert.Greater(t, len(embedding), 0)
		assert.LessOrEqual(t, len(embedding), 2048) // Reasonable upper bound
		
		// Verify embedding values are not all zeros
		hasNonZero := false
		for _, val := range embedding {
			if val != 0 {
				hasNonZero = true
				break
			}
		}
		assert.True(t, hasNonZero, "Embedding should contain non-zero values")
	})

	t.Run("BatchEmbedding", func(t *testing.T) {
		if err := provider.HealthCheck(ctx); err != nil {
			t.Skipf("Ollama not available: %v", err)
		}

		// Test batch embedding generation
		texts := []string{
			"First test document",
			"Second test document",
			"Third test document with more content",
		}
		
		embeddings, err := provider.GenerateBatchEmbeddings(ctx, texts)
		require.NoError(t, err)
		require.Len(t, embeddings, len(texts))

		// Verify all embeddings are valid
		for i, emb := range embeddings {
			assert.Greater(t, len(emb), 0, "Embedding %d should not be empty", i)
			
			// Verify embeddings are different (not all the same)
			if i > 0 {
				different := false
				for j := 0; j < len(emb) && j < len(embeddings[0]); j++ {
					if emb[j] != embeddings[0][j] {
						different = true
						break
					}
				}
				assert.True(t, different, "Embeddings should be different for different texts")
			}
		}
	})
}

// TestChromaDBIntegration tests real ChromaDB integration
func TestChromaDBIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Load configuration
	cfg, err := config.LoadConfig("")
	require.NoError(t, err)

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Reduce noise in tests

	// Create ChromaDB vector store with test collection
	testCollection := "test_integration_" + generateRandomString(8)
	chromaConfig := vector.ChromaDBConfig{
		BaseURL:    cfg.ChromaDB.BaseURL,
		Collection: testCollection,
	}
	store := vector.NewChromaDBVectorStore(chromaConfig, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Cleanup test collection after test
	defer func() {
		if err := store.DeleteCollection(ctx, testCollection); err != nil {
			t.Logf("Failed to cleanup test collection: %v", err)
		}
	}()

	t.Run("HealthCheck", func(t *testing.T) {
		err := store.HealthCheck(ctx)
		if err != nil {
			t.Skipf("ChromaDB not available: %v", err)
		}
	})

	t.Run("CreateCollection", func(t *testing.T) {
		if err := store.HealthCheck(ctx); err != nil {
			t.Skipf("ChromaDB not available: %v", err)
		}

		err := store.CreateCollection(ctx, testCollection)
		require.NoError(t, err)

		// Verify collection exists
		collections, err := store.ListCollections(ctx)
		require.NoError(t, err)
		assert.Contains(t, collections, testCollection)
	})

	t.Run("StoreAndSearchVectors", func(t *testing.T) {
		if err := store.HealthCheck(ctx); err != nil {
			t.Skipf("ChromaDB not available: %v", err)
		}

		// Create test collection if not exists
		_ = store.CreateCollection(ctx, testCollection)

		// Test data
		testVectors := []struct {
			id       string
			vector   []float32
			metadata map[string]interface{}
		}{
			{
				id:     "test1",
				vector: []float32{1.0, 0.0, 0.0},
				metadata: map[string]interface{}{
					"title": "Test Document 1",
					"type":  "test",
				},
			},
			{
				id:     "test2",
				vector: []float32{0.0, 1.0, 0.0},
				metadata: map[string]interface{}{
					"title": "Test Document 2",
					"type":  "test",
				},
			},
			{
				id:     "test3",
				vector: []float32{0.0, 0.0, 1.0},
				metadata: map[string]interface{}{
					"title": "Test Document 3",
					"type":  "test",
				},
			},
		}

		// Store vectors
		for _, tv := range testVectors {
			err := store.Store(ctx, tv.id, tv.vector, tv.metadata)
			require.NoError(t, err)
		}

		// Search for similar vectors
		queryVector := []float32{1.0, 0.1, 0.1} // Should be closest to test1
		results, err := store.Search(ctx, queryVector, 2, 0.0)
		require.NoError(t, err)
		require.Greater(t, len(results), 0)

		// Verify results are ordered by similarity
		assert.Equal(t, "test1", results[0].ID)
		assert.Greater(t, float32(results[0].Similarity), float32(0.8)) // Should be very similar

		// Test update
		updatedVector := []float32{0.9, 0.1, 0.0}
		updatedMetadata := map[string]interface{}{
			"title":   "Updated Test Document 1",
			"type":    "test",
			"updated": true,
		}
		err = store.Update(ctx, "test1", updatedVector, updatedMetadata)
		require.NoError(t, err)

		// Search again to verify update
		results, err = store.Search(ctx, updatedVector, 1, 0.0)
		require.NoError(t, err)
		require.Greater(t, len(results), 0)
		assert.Equal(t, "test1", results[0].ID)
		assert.Equal(t, "Updated Test Document 1", results[0].Metadata["title"])

		// Test delete
		err = store.Delete(ctx, "test1")
		require.NoError(t, err)

		// Verify deletion
		results, err = store.Search(ctx, updatedVector, 3, 0.0)
		require.NoError(t, err)
		for _, result := range results {
			assert.NotEqual(t, "test1", result.ID)
		}
	})
}

// TestFullIntegration tests the complete pipeline with both Ollama and ChromaDB
func TestFullIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Load configuration
	cfg, err := config.LoadConfig("")
	require.NoError(t, err)

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	// Initialize Ollama provider
	ollamaConfig := embedding.OllamaConfig{
		BaseURL: cfg.Ollama.BaseURL,
		Model:   cfg.Ollama.Model,
	}
	embeddingProvider := embedding.NewOllamaProvider(ollamaConfig, logger)

	// Initialize ChromaDB store
	testCollection := "test_full_integration_" + generateRandomString(8)
	chromaConfig := vector.ChromaDBConfig{
		BaseURL:    cfg.ChromaDB.BaseURL,
		Collection: testCollection,
	}
	vectorStore := vector.NewChromaDBVectorStore(chromaConfig, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Cleanup
	defer func() {
		if err := vectorStore.DeleteCollection(ctx, testCollection); err != nil {
			t.Logf("Failed to cleanup test collection: %v", err)
		}
	}()

	// Check health of both services
	if err := embeddingProvider.HealthCheck(ctx); err != nil {
		t.Skipf("Ollama not available: %v", err)
	}
	if err := vectorStore.HealthCheck(ctx); err != nil {
		t.Skipf("ChromaDB not available: %v", err)
	}

	t.Run("EndToEndPipeline", func(t *testing.T) {
		// Create collection
		err := vectorStore.CreateCollection(ctx, testCollection)
		require.NoError(t, err)

		// Test documents
		documents := []struct {
			id      string
			content string
			title   string
		}{
			{"doc1", "Machine learning algorithms for data analysis", "ML Algorithms"},
			{"doc2", "Deep learning neural networks and training", "Deep Learning"},
			{"doc3", "Natural language processing techniques", "NLP Techniques"},
			{"doc4", "Computer vision and image recognition", "Computer Vision"},
		}

		// Generate embeddings and store vectors
		for _, doc := range documents {
			// Generate embedding
			embedding, err := embeddingProvider.GenerateEmbedding(ctx, doc.content)
			require.NoError(t, err)

			// Store in vector database
			metadata := map[string]interface{}{
				"title":   doc.title,
				"content": doc.content,
			}
			err = vectorStore.Store(ctx, doc.id, embedding, metadata)
			require.NoError(t, err)
		}

		// Test semantic search
		queryText := "deep neural networks for machine learning"
		queryEmbedding, err := embeddingProvider.GenerateEmbedding(ctx, queryText)
		require.NoError(t, err)

		// Search for similar documents
		results, err := vectorStore.Search(ctx, queryEmbedding, 3, 0.3)
		require.NoError(t, err)
		require.Greater(t, len(results), 0)

		// Verify results make semantic sense
		// The query should match "Deep Learning" and "ML Algorithms" documents
		foundDeepLearning := false
		foundMLAlgorithms := false
		
		for _, result := range results {
			if result.ID == "doc2" { // Deep Learning
				foundDeepLearning = true
			}
			if result.ID == "doc1" { // ML Algorithms
				foundMLAlgorithms = true
			}
			
			// All results should have reasonable similarity scores
			assert.Greater(t, float32(result.Similarity), float32(0.3))
			
			// Metadata should be preserved
			assert.Contains(t, result.Metadata, "title")
			assert.Contains(t, result.Metadata, "content")
		}

		// At least one of the semantically similar documents should be found
		assert.True(t, foundDeepLearning || foundMLAlgorithms, 
			"Should find semantically similar documents")
	})
}

// generateRandomString generates a random string for test collection names
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}

// TestIntegrationEnvironment verifies the test environment setup
func TestIntegrationEnvironment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("ConfigLoading", func(t *testing.T) {
		cfg, err := config.LoadConfig("")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		// Validate configuration
		err = cfg.ValidateConfig()
		assert.NoError(t, err)
	})

	t.Run("ServiceAvailability", func(t *testing.T) {
		cfg, err := config.LoadConfig("")
		require.NoError(t, err)

		logger := logrus.New()
		logger.SetLevel(logrus.WarnLevel)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Test Ollama availability
		ollamaConfig := embedding.OllamaConfig{
			BaseURL: cfg.Ollama.BaseURL,
			Model:   cfg.Ollama.Model,
		}
		ollamaProvider := embedding.NewOllamaProvider(ollamaConfig, logger)
		ollamaAvailable := ollamaProvider.HealthCheck(ctx) == nil

		// Test ChromaDB availability
		chromaConfig := vector.ChromaDBConfig{
			BaseURL:    cfg.ChromaDB.BaseURL,
			Collection: "test",
		}
		chromaStore := vector.NewChromaDBVectorStore(chromaConfig, logger)
		chromaAvailable := chromaStore.HealthCheck(ctx) == nil

		t.Logf("Ollama available: %v", ollamaAvailable)
		t.Logf("ChromaDB available: %v", chromaAvailable)

		if !ollamaAvailable && !chromaAvailable {
			t.Skip("Neither Ollama nor ChromaDB is available for integration testing")
		}
	})
}

// init sets up integration test environment
func init() {
	// Set test environment variables if not already set
	if os.Getenv("MEMORY_BANK_OLLAMA_BASE_URL") == "" {
		os.Setenv("MEMORY_BANK_OLLAMA_BASE_URL", "http://localhost:11434")
	}
	if os.Getenv("MEMORY_BANK_CHROMADB_BASE_URL") == "" {
		os.Setenv("MEMORY_BANK_CHROMADB_BASE_URL", "http://localhost:8000")
	}
}