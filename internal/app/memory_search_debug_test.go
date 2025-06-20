// +build integration

// Memory Search Debug Integration Test
// This test performs comprehensive debugging of the Memory Bank search functionality
// to identify and resolve semantic search issues. It creates memories, stores them
// with real embeddings, and verifies the complete search pipeline.
//
// Purpose: Debug and validate end-to-end memory search workflow
// Services: Uses real Ollama + ChromaDB services
// Run with: go test -tags=integration ./internal/app -run TestMemorySearchDebug -v

package app

import (
	"context"
	"testing"
	"time"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/infra/config"
	"github.com/joern1811/memory-bank/internal/infra/database"
	"github.com/joern1811/memory-bank/internal/infra/embedding"
	"github.com/joern1811/memory-bank/internal/infra/vector"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMemorySearchDebug debugs the Memory Bank search issue
func TestMemorySearchDebug(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Load configuration
	cfg, err := config.LoadConfig("")
	require.NoError(t, err)

	// Setup debug logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel) // Enable debug logging
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	t.Run("EndToEndMemorySearch", func(t *testing.T) {
		// Initialize database with temporary file
		dbPath := "/tmp/memory_search_debug_test.db"
		db, err := database.NewSQLiteDatabase(dbPath, logger)
		require.NoError(t, err)
		defer db.Close()

		// Initialize repositories
		memoryRepo := database.NewSQLiteMemoryRepository(db, logger)
		projectRepo := database.NewSQLiteProjectRepository(db, logger)

		// Initialize embedding provider
		ollamaConfig := embedding.OllamaConfig{
			BaseURL: cfg.Ollama.BaseURL,
			Model:   cfg.Ollama.Model,
		}
		embeddingProvider := embedding.NewOllamaProvider(ollamaConfig, logger)

		// Check Ollama availability
		if err := embeddingProvider.HealthCheck(ctx); err != nil {
			t.Skipf("Ollama not available: %v", err)
		}

		// Initialize vector store with test collection
		testCollection := "memory_search_debug_" + generateRandomID()
		chromaConfig := vector.ChromaDBConfig{
			BaseURL:    cfg.ChromaDB.BaseURL,
			Collection: testCollection,
			Tenant:     cfg.ChromaDB.Tenant,
			Database:   cfg.ChromaDB.Database,
		}
		vectorStore := vector.NewChromaDBVectorStore(chromaConfig, logger)

		// Check ChromaDB availability
		if err := vectorStore.HealthCheck(ctx); err != nil {
			t.Skipf("ChromaDB not available: %v", err)
		}

		// Cleanup
		defer func() {
			if err := vectorStore.DeleteCollection(ctx, testCollection); err != nil {
				t.Logf("Failed to cleanup test collection: %v", err)
			}
		}()

		// Create collection
		err = vectorStore.CreateCollection(ctx, testCollection)
		require.NoError(t, err)

		// Initialize Memory Service
		memoryService := NewMemoryService(memoryRepo, embeddingProvider, vectorStore, logger)

		// Create a test project
		randomID := generateRandomID()
		project := &domain.Project{
			ID:          domain.ProjectID("test_project_" + randomID),
			Name:        "Test Project",
			Path:        "/tmp/test_" + randomID,
			Description: "Test project for debugging search",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err = projectRepo.Store(ctx, project)
		require.NoError(t, err)

		// Step 1: Create a memory entry
		t.Log("=== STEP 1: Creating Memory Entry ===")
		
		createReq := ports.CreateMemoryRequest{
			ProjectID: project.ID,
			Type:      domain.MemoryTypeDecision,
			Title:     "Authentication Strategy Decision",
			Content:   "We decided to use JWT tokens for user authentication because they are stateless and can be easily validated. This approach provides better scalability compared to session-based authentication.",
			Tags:      []string{"authentication", "jwt", "security", "api"},
		}

		createdMemory, err := memoryService.CreateMemory(ctx, createReq)
		require.NoError(t, err)
		require.NotNil(t, createdMemory)

		t.Logf("Created memory with ID: %s", createdMemory.ID)
		t.Logf("Memory has embedding: %t", createdMemory.HasEmbedding)

		// Verify memory was stored in database
		retrievedMemory, err := memoryRepo.GetByID(ctx, createdMemory.ID)
		require.NoError(t, err)
		require.NotNil(t, retrievedMemory)
		t.Logf("Memory retrieved from DB - HasEmbedding: %t", retrievedMemory.HasEmbedding)

		// Step 2: Verify embedding was generated and stored in ChromaDB
		t.Log("=== STEP 2: Verifying ChromaDB Storage ===")
		
		// Get the exact embedding that was generated
		embeddingText := createdMemory.Title + " " + createdMemory.Content
		queryEmbedding, err := embeddingProvider.GenerateEmbedding(ctx, embeddingText)
		require.NoError(t, err)
		t.Logf("Generated query embedding with length: %d", len(queryEmbedding))

		// Test direct ChromaDB search to verify the vector was stored
		chromaResults, err := vectorStore.Search(ctx, queryEmbedding, 5, 0.0)
		require.NoError(t, err)
		t.Logf("Direct ChromaDB search returned %d results", len(chromaResults))
		
		for i, result := range chromaResults {
			t.Logf("ChromaDB Result %d: ID=%s, Similarity=%.4f, Metadata=%+v", 
				i+1, result.ID, result.Similarity, result.Metadata)
		}

		// Verify that our memory was found in ChromaDB
		found := false
		for _, result := range chromaResults {
			if result.ID == string(createdMemory.ID) {
				found = true
				t.Logf("✓ Found memory in ChromaDB with similarity: %.4f", result.Similarity)
				break
			}
		}
		assert.True(t, found, "Memory should be found in ChromaDB")

		// Step 3: Test Memory Service Search
		t.Log("=== STEP 3: Testing Memory Service Search ===")
		
		searchReq := ports.SemanticSearchRequest{
			Query:     "authentication jwt tokens",
			ProjectID: &project.ID,
			Limit:     10,
			Threshold: 0.1, // Very low threshold to catch any results
		}

		searchResults, err := memoryService.SearchMemories(ctx, searchReq)
		require.NoError(t, err)
		t.Logf("Memory Service search returned %d results", len(searchResults))

		// Debug: Print all search results
		for i, result := range searchResults {
			t.Logf("Memory Service Result %d: ID=%s, Title=%s, Similarity=%.4f", 
				i+1, result.Memory.ID, result.Memory.Title, result.Similarity)
		}

		// Step 4: Compare results
		t.Log("=== STEP 4: Analysis ===")
		
		if len(searchResults) == 0 {
			t.Log("❌ Memory Service returned no results - this is the bug!")
			
			// Additional debugging - let's trace the search process
			t.Log("=== Additional Debugging ===")
			
			// Test the search query embedding
			searchEmbedding, err := embeddingProvider.GenerateEmbedding(ctx, searchReq.Query)
			require.NoError(t, err)
			t.Logf("Search embedding length: %d", len(searchEmbedding))
			
			// Test if search embedding finds the stored vector
			directSearchResults, err := vectorStore.Search(ctx, searchEmbedding, 5, searchReq.Threshold)
			require.NoError(t, err)
			t.Logf("Direct search with search embedding returned %d results", len(directSearchResults))
			
			for i, result := range directSearchResults {
				t.Logf("Direct Search Result %d: ID=%s, Similarity=%.4f", 
					i+1, result.ID, result.Similarity)
			}
			
			// Check if the memory still exists in the database
			allMemories, err := memoryRepo.ListByProject(ctx, project.ID)
			require.NoError(t, err)
			t.Logf("Database contains %d memories for this project", len(allMemories))
			
			if len(allMemories) > 0 {
				t.Logf("First memory: ID=%s, HasEmbedding=%t", allMemories[0].ID, allMemories[0].HasEmbedding)
			}
			
		} else {
			t.Log("✓ Memory Service returned results - search is working!")
			
			// Verify our created memory is in the results
			foundInResults := false
			for _, result := range searchResults {
				if result.Memory.ID == createdMemory.ID {
					foundInResults = true
					t.Logf("✓ Created memory found in search results with similarity: %.4f", result.Similarity)
					break
				}
			}
			assert.True(t, foundInResults, "Created memory should be found in search results")
		}
	})
}

// generateRandomID generates a random ID for test resources
func generateRandomID() string {
	return time.Now().Format("20060102150405")
}