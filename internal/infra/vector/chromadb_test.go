package vector

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/sirupsen/logrus"
)

func setupTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Reduce noise in tests
	return logger
}

func TestNewChromaDBVectorStore(t *testing.T) {
	logger := setupTestLogger()
	
	// Test with default config
	store := NewChromaDBVectorStore(ChromaDBConfig{}, logger)
	if store == nil {
		t.Fatal("Expected non-nil store")
	}
	
	// Test with custom config
	config := ChromaDBConfig{
		BaseURL:    "http://localhost:9000",
		Collection: "test_collection",
	}
	store = NewChromaDBVectorStore(config, logger)
	
	if store.baseURL != config.BaseURL {
		t.Errorf("Expected baseURL %s, got %s", config.BaseURL, store.baseURL)
	}
	if store.collection != config.Collection {
		t.Errorf("Expected collection %s, got %s", config.Collection, store.collection)
	}
}

func TestChromaDBVectorStore_Store(t *testing.T) {
	// Mock collections response for getCollectionID
	mockCollections := []chromaDBCollection{
		{Name: "test_collection", ID: "test_col_id"},
	}
	
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/tenants/default_tenant/databases/default_database/collections":
			// Handle collection listing for getCollectionID
			if r.Method != "GET" {
				t.Errorf("Expected GET request for collections, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockCollections)
			
		case "/api/v2/tenants/default_tenant/databases/default_database/collections/test_col_id/add":
			// Handle actual store operation
			if r.Method != "POST" {
				t.Errorf("Expected POST request for add, got %s", r.Method)
			}
			
			// Verify request body
			var doc chromaDBDocument
			if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}
			
			if len(doc.IDs) != 1 || doc.IDs[0] != "test_id" {
				t.Errorf("Expected ID 'test_id', got %v", doc.IDs)
			}
			
			w.WriteHeader(http.StatusOK)
			
		default:
			t.Errorf("Unexpected request path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	
	// Create store with default tenant/database
	config := ChromaDBConfig{
		BaseURL:    server.URL,
		Collection: "test_collection",
		Tenant:     "default_tenant",
		Database:   "default_database",
	}
	store := NewChromaDBVectorStore(config, setupTestLogger())
	
	// Test store operation
	vector := domain.EmbeddingVector{0.1, 0.2, 0.3}
	metadata := map[string]interface{}{
		"type":  "test",
		"title": "Test Memory",
	}
	
	err := store.Store(context.Background(), "test_id", vector, metadata)
	if err != nil {
		t.Errorf("Store failed: %v", err)
	}
}

func TestChromaDBVectorStore_Store_Error(t *testing.T) {
	// Mock server that returns an error for collection listing
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}))
	defer server.Close()
	
	config := ChromaDBConfig{
		BaseURL:    server.URL,
		Collection: "test_collection",
		Tenant:     "default_tenant",
		Database:   "default_database",
	}
	store := NewChromaDBVectorStore(config, setupTestLogger())
	
	vector := domain.EmbeddingVector{0.1, 0.2, 0.3}
	metadata := map[string]interface{}{"type": "test"}
	
	err := store.Store(context.Background(), "test_id", vector, metadata)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	
	if !strings.Contains(err.Error(), "ChromaDB API error") {
		t.Errorf("Expected ChromaDB API error, got: %v", err)
	}
}

func TestChromaDBVectorStore_Delete(t *testing.T) {
	// Mock collections response for getCollectionID
	mockCollections := []chromaDBCollection{
		{Name: "test_collection", ID: "test_col_id"},
	}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/tenants/default_tenant/databases/default_database/collections":
			// Handle collection listing for getCollectionID
			if r.Method != "GET" {
				t.Errorf("Expected GET request for collections, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockCollections)
			
		case "/api/v2/tenants/default_tenant/databases/default_database/collections/test_col_id/delete":
			// Handle actual delete operation
			if r.Method != "POST" {
				t.Errorf("Expected POST request for delete, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			
		default:
			t.Errorf("Unexpected request path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	
	config := ChromaDBConfig{
		BaseURL:    server.URL,
		Collection: "test_collection",
		Tenant:     "default_tenant",
		Database:   "default_database",
	}
	store := NewChromaDBVectorStore(config, setupTestLogger())
	
	err := store.Delete(context.Background(), "test_id")
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
}

func TestChromaDBVectorStore_Search(t *testing.T) {
	// Mock collections response for getCollectionID
	mockCollections := []chromaDBCollection{
		{Name: "test_collection", ID: "test_col_id"},
	}
	
	// Mock response data
	mockResponse := chromaDBQueryResponse{
		IDs:       [][]string{{"id1", "id2"}},
		Distances: [][]float32{{0.1, 0.3}}, // ChromaDB returns distances
		Metadatas: [][]map[string]interface{}{
			{
				{"type": "decision", "title": "Memory 1"},
				{"type": "pattern", "title": "Memory 2"},
			},
		},
	}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/tenants/default_tenant/databases/default_database/collections":
			// Handle collection listing for getCollectionID
			if r.Method != "GET" {
				t.Errorf("Expected GET request for collections, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockCollections)
			
		case "/api/v2/tenants/default_tenant/databases/default_database/collections/test_col_id/query":
			// Handle actual search operation
			if r.Method != "POST" {
				t.Errorf("Expected POST request for query, got %s", r.Method)
			}
			
			// Verify request body
			var queryReq chromaDBQueryRequest
			if err := json.NewDecoder(r.Body).Decode(&queryReq); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}
			
			if queryReq.NResults != 5 {
				t.Errorf("Expected NResults 5, got %d", queryReq.NResults)
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
			
		default:
			t.Errorf("Unexpected request path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	
	config := ChromaDBConfig{
		BaseURL:    server.URL,
		Collection: "test_collection",
		Tenant:     "default_tenant",
		Database:   "default_database",
	}
	store := NewChromaDBVectorStore(config, setupTestLogger())
	
	vector := domain.EmbeddingVector{0.1, 0.2, 0.3}
	results, err := store.Search(context.Background(), vector, 5, 0.0)
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}
	
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	
	// Check first result
	if results[0].ID != "id1" {
		t.Errorf("Expected ID 'id1', got %s", results[0].ID)
	}
	
	// Check similarity conversion (1.0 - distance)
	expectedSimilarity := domain.Similarity(1.0 - 0.1)
	if results[0].Similarity != expectedSimilarity {
		t.Errorf("Expected similarity %f, got %f", expectedSimilarity, results[0].Similarity)
	}
}

func TestChromaDBVectorStore_Search_WithThreshold(t *testing.T) {
	// Mock collections response for getCollectionID
	mockCollections := []chromaDBCollection{
		{Name: "test_collection", ID: "test_col_id"},
	}
	
	// Mock response with one result above threshold and one below
	mockResponse := chromaDBQueryResponse{
		IDs:       [][]string{{"id1", "id2"}},
		Distances: [][]float32{{0.1, 0.8}}, // 0.9 and 0.2 similarity
		Metadatas: [][]map[string]interface{}{
			{
				{"type": "decision"},
				{"type": "pattern"},
			},
		},
	}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/tenants/default_tenant/databases/default_database/collections":
			// Handle collection listing for getCollectionID
			if r.Method != "GET" {
				t.Errorf("Expected GET request for collections, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockCollections)
			
		case "/api/v2/tenants/default_tenant/databases/default_database/collections/test_col_id/query":
			// Handle actual search operation
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
			
		default:
			t.Errorf("Unexpected request path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	
	config := ChromaDBConfig{
		BaseURL:    server.URL,
		Collection: "test_collection",
		Tenant:     "default_tenant",
		Database:   "default_database",
	}
	store := NewChromaDBVectorStore(config, setupTestLogger())
	
	vector := domain.EmbeddingVector{0.1, 0.2, 0.3}
	results, err := store.Search(context.Background(), vector, 10, 0.5) // threshold 0.5
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}
	
	// Only the first result (similarity 0.9) should pass the threshold
	if len(results) != 1 {
		t.Errorf("Expected 1 result above threshold, got %d", len(results))
	}
	
	if results[0].ID != "id1" {
		t.Errorf("Expected ID 'id1', got %s", results[0].ID)
	}
}

func TestChromaDBVectorStore_CreateCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		
		expectedPath := "/api/v2/tenants/default_tenant/databases/default_database/collections"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}
		
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()
	
	config := ChromaDBConfig{
		BaseURL:  server.URL,
		Tenant:   "default_tenant",
		Database: "default_database",
	}
	store := NewChromaDBVectorStore(config, setupTestLogger())
	
	err := store.CreateCollection(context.Background(), "new_collection")
	if err != nil {
		t.Errorf("CreateCollection failed: %v", err)
	}
}

func TestChromaDBVectorStore_ListCollections(t *testing.T) {
	mockCollections := []chromaDBCollection{
		{Name: "collection1", ID: "id1"},
		{Name: "collection2", ID: "id2"},
	}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		
		expectedPath := "/api/v2/tenants/default_tenant/databases/default_database/collections"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockCollections)
	}))
	defer server.Close()
	
	config := ChromaDBConfig{
		BaseURL:  server.URL,
		Tenant:   "default_tenant",
		Database: "default_database",
	}
	store := NewChromaDBVectorStore(config, setupTestLogger())
	
	collections, err := store.ListCollections(context.Background())
	if err != nil {
		t.Errorf("ListCollections failed: %v", err)
	}
	
	if len(collections) != 2 {
		t.Errorf("Expected 2 collections, got %d", len(collections))
	}
	
	if collections[0] != "collection1" || collections[1] != "collection2" {
		t.Errorf("Unexpected collection names: %v", collections)
	}
}

func TestMockVectorStore(t *testing.T) {
	logger := setupTestLogger()
	store := NewMockVectorStore(logger)
	
	// Test Store
	vector1 := domain.EmbeddingVector{1.0, 0.0, 0.0}
	vector2 := domain.EmbeddingVector{0.0, 1.0, 0.0}
	metadata1 := map[string]interface{}{"type": "decision"}
	metadata2 := map[string]interface{}{"type": "pattern"}
	
	err := store.Store(context.Background(), "id1", vector1, metadata1)
	if err != nil {
		t.Errorf("Store failed: %v", err)
	}
	
	err = store.Store(context.Background(), "id2", vector2, metadata2)
	if err != nil {
		t.Errorf("Store failed: %v", err)
	}
	
	// Test Search
	queryVector := domain.EmbeddingVector{0.9, 0.1, 0.0} // Similar to vector1
	results, err := store.Search(context.Background(), queryVector, 10, 0.0)
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}
	
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	
	// Results should be sorted by similarity (highest first)
	// queryVector is more similar to vector1 than vector2
	if results[0].ID != "id1" {
		t.Errorf("Expected first result to be 'id1', got %s", results[0].ID)
	}
	
	// Test Update
	newVector := domain.EmbeddingVector{0.5, 0.5, 0.0}
	newMetadata := map[string]interface{}{"type": "updated"}
	
	err = store.Update(context.Background(), "id1", newVector, newMetadata)
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}
	
	// Verify update worked by checking the stored entry
	if entry, exists := store.vectors["id1"]; !exists {
		t.Error("Entry id1 should exist after update")
	} else {
		if entry.Metadata["type"] != "updated" {
			t.Errorf("Expected updated metadata, got %v", entry.Metadata)
		}
	}
	
	// Test Delete
	err = store.Delete(context.Background(), "id1")
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
	
	if _, exists := store.vectors["id1"]; exists {
		t.Error("Entry id1 should not exist after delete")
	}
	
	// Test Collection operations
	err = store.CreateCollection(context.Background(), "test_collection")
	if err != nil {
		t.Errorf("CreateCollection failed: %v", err)
	}
	
	collections, err := store.ListCollections(context.Background())
	if err != nil {
		t.Errorf("ListCollections failed: %v", err)
	}
	
	if len(collections) != 1 || collections[0] != "test_collection" {
		t.Errorf("Expected ['test_collection'], got %v", collections)
	}
	
	err = store.DeleteCollection(context.Background(), "test_collection")
	if err != nil {
		t.Errorf("DeleteCollection failed: %v", err)
	}
	
	collections, err = store.ListCollections(context.Background())
	if err != nil {
		t.Errorf("ListCollections failed: %v", err)
	}
	
	if len(collections) != 0 {
		t.Errorf("Expected empty collections list, got %v", collections)
	}
}

func TestCalculateDotProduct(t *testing.T) {
	// Test normal vectors
	a := domain.EmbeddingVector{1.0, 2.0, 3.0}
	b := domain.EmbeddingVector{4.0, 5.0, 6.0}
	
	expected := domain.Similarity(1.0*4.0 + 2.0*5.0 + 3.0*6.0) // 32.0
	result := calculateDotProduct(a, b)
	
	if result != expected {
		t.Errorf("Expected %f, got %f", expected, result)
	}
	
	// Test different length vectors
	c := domain.EmbeddingVector{1.0, 2.0}
	d := domain.EmbeddingVector{1.0, 2.0, 3.0}
	
	result = calculateDotProduct(c, d)
	if result != 0 {
		t.Errorf("Expected 0 for different length vectors, got %f", result)
	}
	
	// Test zero vectors
	zero1 := domain.EmbeddingVector{0.0, 0.0, 0.0}
	zero2 := domain.EmbeddingVector{0.0, 0.0, 0.0}
	
	result = calculateDotProduct(zero1, zero2)
	if result != 0 {
		t.Errorf("Expected 0 for zero vectors, got %f", result)
	}
}

func TestSearchByText_NotImplemented(t *testing.T) {
	logger := setupTestLogger()
	
	// Test ChromaDB store
	config := ChromaDBConfig{BaseURL: "http://localhost:8000"}
	store := NewChromaDBVectorStore(config, logger)
	
	_, err := store.SearchByText(context.Background(), "test", 10, 0.5)
	if err == nil {
		t.Error("Expected error for SearchByText, got nil")
	}
	
	// Test Mock store
	mockStore := NewMockVectorStore(logger)
	_, err = mockStore.SearchByText(context.Background(), "test", 10, 0.5)
	if err == nil {
		t.Error("Expected error for SearchByText on mock store, got nil")
	}
}