package vector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/sirupsen/logrus"
)

// ChromaDBVectorStore implements the VectorStore interface using ChromaDB
type ChromaDBVectorStore struct {
	baseURL    string
	collection string
	client     *http.Client
	logger     *logrus.Logger
}

// ChromaDBConfig holds configuration for ChromaDB provider
type ChromaDBConfig struct {
	BaseURL    string        `json:"base_url"`
	Collection string        `json:"collection"`
	Timeout    time.Duration `json:"timeout"`
}

// DefaultChromeDBConfig returns default configuration for ChromaDB
func DefaultChromeDBConfig() ChromaDBConfig {
	return ChromaDBConfig{
		BaseURL:    "http://localhost:8000",
		Collection: "memory_bank",
		Timeout:    30 * time.Second,
	}
}

// NewChromaDBVectorStore creates a new ChromaDB vector store
func NewChromaDBVectorStore(config ChromaDBConfig, logger *logrus.Logger) *ChromaDBVectorStore {
	if config.BaseURL == "" {
		config = DefaultChromeDBConfig()
	}

	return &ChromaDBVectorStore{
		baseURL:    config.BaseURL,
		collection: config.Collection,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
	}
}

// chromaDBDocument represents a document in ChromaDB
type chromaDBDocument struct {
	IDs       []string                 `json:"ids"`
	Documents []string                 `json:"documents,omitempty"`
	Metadatas []map[string]interface{} `json:"metadatas,omitempty"`
	Embeddings [][]float32             `json:"embeddings,omitempty"`
}

// chromaDBQueryRequest represents a query request to ChromaDB
type chromaDBQueryRequest struct {
	QueryEmbeddings [][]float32 `json:"query_embeddings"`
	NResults        int         `json:"n_results"`
	Include         []string    `json:"include,omitempty"`
	Where           map[string]interface{} `json:"where,omitempty"`
}

// chromaDBQueryResponse represents a query response from ChromaDB
type chromaDBQueryResponse struct {
	IDs       [][]string                 `json:"ids"`
	Distances [][]float32                `json:"distances"`
	Metadatas [][]map[string]interface{} `json:"metadatas"`
	Documents [][]string                 `json:"documents"`
}

// chromaDBCollection represents a collection in ChromaDB
type chromaDBCollection struct {
	Name     string                 `json:"name"`
	ID       string                 `json:"id"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Store stores a vector with metadata in ChromaDB
func (c *ChromaDBVectorStore) Store(ctx context.Context, id string, vector domain.EmbeddingVector, metadata map[string]interface{}) error {
	c.logger.WithFields(logrus.Fields{
		"collection":      c.collection,
		"id":              id,
		"vector_length":   len(vector),
		"metadata_keys":   len(metadata),
	}).Debug("Storing vector in ChromaDB")

	// Prepare document
	doc := chromaDBDocument{
		IDs:        []string{id},
		Embeddings: [][]float32{vector},
		Metadatas:  []map[string]interface{}{metadata},
	}

	// Convert to JSON
	jsonBody, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/v1/collections/%s/add", c.baseURL, c.collection)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chromadb API error (status %d): %s", resp.StatusCode, string(body))
	}

	c.logger.WithField("id", id).Debug("Vector stored successfully in ChromaDB")
	return nil
}

// Delete removes a vector from ChromaDB
func (c *ChromaDBVectorStore) Delete(ctx context.Context, id string) error {
	c.logger.WithFields(logrus.Fields{
		"collection": c.collection,
		"id":         id,
	}).Debug("Deleting vector from ChromaDB")

	// Prepare delete request
	deleteDoc := map[string]interface{}{
		"ids": []string{id},
	}

	jsonBody, err := json.Marshal(deleteDoc)
	if err != nil {
		return fmt.Errorf("failed to marshal delete request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/v1/collections/%s/delete", c.baseURL, c.collection)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chromadb API error (status %d): %s", resp.StatusCode, string(body))
	}

	c.logger.WithField("id", id).Debug("Vector deleted successfully from ChromaDB")
	return nil
}

// Update updates a vector and its metadata in ChromaDB
func (c *ChromaDBVectorStore) Update(ctx context.Context, id string, vector domain.EmbeddingVector, metadata map[string]interface{}) error {
	c.logger.WithFields(logrus.Fields{
		"collection":      c.collection,
		"id":              id,
		"vector_length":   len(vector),
		"metadata_keys":   len(metadata),
	}).Debug("Updating vector in ChromaDB")

	// Prepare update document
	doc := chromaDBDocument{
		IDs:        []string{id},
		Embeddings: [][]float32{vector},
		Metadatas:  []map[string]interface{}{metadata},
	}

	jsonBody, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/v1/collections/%s/update", c.baseURL, c.collection)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chromadb API error (status %d): %s", resp.StatusCode, string(body))
	}

	c.logger.WithField("id", id).Debug("Vector updated successfully in ChromaDB")
	return nil
}

// Search performs vector similarity search
func (c *ChromaDBVectorStore) Search(ctx context.Context, vector domain.EmbeddingVector, limit int, threshold float32) ([]ports.SearchResult, error) {
	c.logger.WithFields(logrus.Fields{
		"collection":     c.collection,
		"vector_length":  len(vector),
		"limit":          limit,
		"threshold":      threshold,
	}).Debug("Searching vectors in ChromaDB")

	// Prepare query request
	queryReq := chromaDBQueryRequest{
		QueryEmbeddings: [][]float32{vector},
		NResults:        limit,
		Include:         []string{"metadatas", "distances"},
	}

	jsonBody, err := json.Marshal(queryReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/v1/collections/%s/query", c.baseURL, c.collection)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chromadb API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var queryResp chromaDBQueryResponse
	if err := json.Unmarshal(body, &queryResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert to search results
	results := make([]ports.SearchResult, 0)
	if len(queryResp.IDs) > 0 && len(queryResp.IDs[0]) > 0 {
		ids := queryResp.IDs[0]
		distances := queryResp.Distances[0]
		metadatas := queryResp.Metadatas[0]

		for i, id := range ids {
			// Convert distance to similarity (ChromaDB uses cosine distance)
			similarity := domain.Similarity(1.0 - distances[i])
			
			// Apply threshold filter
			if similarity.IsRelevant(threshold) {
				results = append(results, ports.SearchResult{
					ID:         id,
					Similarity: similarity,
					Metadata:   metadatas[i],
				})
			}
		}
	}

	c.logger.WithFields(logrus.Fields{
		"results_count": len(results),
		"threshold":     threshold,
	}).Debug("Vector search completed")

	return results, nil
}

// SearchByText performs text-based search (not implemented as it requires embedding generation)
func (c *ChromaDBVectorStore) SearchByText(ctx context.Context, text string, limit int, threshold float32) ([]ports.SearchResult, error) {
	return nil, fmt.Errorf("SearchByText not implemented - use Search with pre-generated embeddings")
}

// CreateCollection creates a new collection in ChromaDB
func (c *ChromaDBVectorStore) CreateCollection(ctx context.Context, name string) error {
	c.logger.WithField("collection", name).Debug("Creating collection in ChromaDB")

	// Prepare collection request
	collectionReq := map[string]interface{}{
		"name": name,
		"metadata": map[string]interface{}{
			"description": "Memory Bank collection",
		},
	}

	jsonBody, err := json.Marshal(collectionReq)
	if err != nil {
		return fmt.Errorf("failed to marshal collection request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/v1/collections", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chromadb API error (status %d): %s", resp.StatusCode, string(body))
	}

	c.logger.WithField("collection", name).Debug("Collection created successfully")
	return nil
}

// DeleteCollection deletes a collection from ChromaDB
func (c *ChromaDBVectorStore) DeleteCollection(ctx context.Context, name string) error {
	c.logger.WithField("collection", name).Debug("Deleting collection from ChromaDB")

	// Create HTTP request
	url := fmt.Sprintf("%s/api/v1/collections/%s", c.baseURL, name)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chromadb API error (status %d): %s", resp.StatusCode, string(body))
	}

	c.logger.WithField("collection", name).Debug("Collection deleted successfully")
	return nil
}

// ListCollections lists all collections in ChromaDB
func (c *ChromaDBVectorStore) ListCollections(ctx context.Context) ([]string, error) {
	c.logger.Debug("Listing collections from ChromaDB")

	// Create HTTP request
	url := fmt.Sprintf("%s/api/v1/collections", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chromadb API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var collections []chromaDBCollection
	if err := json.Unmarshal(body, &collections); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract collection names
	names := make([]string, len(collections))
	for i, col := range collections {
		names[i] = col.Name
	}

	c.logger.WithField("count", len(names)).Debug("Collections listed successfully")
	return names, nil
}

// HealthCheck verifies that ChromaDB is running and accessible
func (c *ChromaDBVectorStore) HealthCheck(ctx context.Context) error {
	c.logger.Debug("Performing ChromaDB health check")

	// Try to list collections as a health check
	_, err := c.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("chromadb health check failed: %w", err)
	}

	c.logger.Info("ChromaDB health check passed")
	return nil
}

// MockVectorStore provides a mock implementation for testing
type MockVectorStore struct {
	vectors     map[string]mockVectorEntry
	collections map[string]bool
	logger      *logrus.Logger
}

type mockVectorEntry struct {
	Vector   domain.EmbeddingVector
	Metadata map[string]interface{}
}

// NewMockVectorStore creates a new mock vector store
func NewMockVectorStore(logger *logrus.Logger) *MockVectorStore {
	return &MockVectorStore{
		vectors:     make(map[string]mockVectorEntry),
		collections: make(map[string]bool),
		logger:      logger,
	}
}

// Store stores a vector with metadata in the mock store
func (m *MockVectorStore) Store(ctx context.Context, id string, vector domain.EmbeddingVector, metadata map[string]interface{}) error {
	m.logger.WithFields(logrus.Fields{
		"id":             id,
		"vector_length":  len(vector),
		"metadata_keys":  len(metadata),
	}).Debug("Storing vector in mock store")

	m.vectors[id] = mockVectorEntry{
		Vector:   vector,
		Metadata: metadata,
	}

	return nil
}

// Delete removes a vector from the mock store
func (m *MockVectorStore) Delete(ctx context.Context, id string) error {
	m.logger.WithField("id", id).Debug("Deleting vector from mock store")

	delete(m.vectors, id)
	return nil
}

// Update updates a vector and its metadata in the mock store
func (m *MockVectorStore) Update(ctx context.Context, id string, vector domain.EmbeddingVector, metadata map[string]interface{}) error {
	m.logger.WithFields(logrus.Fields{
		"id":             id,
		"vector_length":  len(vector),
		"metadata_keys":  len(metadata),
	}).Debug("Updating vector in mock store")

	m.vectors[id] = mockVectorEntry{
		Vector:   vector,
		Metadata: metadata,
	}

	return nil
}

// Search performs mock vector similarity search using simple dot product
func (m *MockVectorStore) Search(ctx context.Context, vector domain.EmbeddingVector, limit int, threshold float32) ([]ports.SearchResult, error) {
	m.logger.WithFields(logrus.Fields{
		"vector_length": len(vector),
		"limit":         limit,
		"threshold":     threshold,
	}).Debug("Searching vectors in mock store")

	var results []ports.SearchResult

	// Calculate similarity with all stored vectors using dot product
	for id, entry := range m.vectors {
		similarity := calculateDotProduct(vector, entry.Vector)
		
		if similarity.IsRelevant(threshold) {
			results = append(results, ports.SearchResult{
				ID:         id,
				Similarity: similarity,
				Metadata:   entry.Metadata,
			})
		}
	}

	// Sort by similarity (highest first) and limit results
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Similarity < results[j].Similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if len(results) > limit {
		results = results[:limit]
	}

	m.logger.WithField("results_count", len(results)).Debug("Mock vector search completed")
	return results, nil
}

// SearchByText is not implemented in the mock store
func (m *MockVectorStore) SearchByText(ctx context.Context, text string, limit int, threshold float32) ([]ports.SearchResult, error) {
	return nil, fmt.Errorf("SearchByText not implemented in mock store")
}

// CreateCollection creates a mock collection
func (m *MockVectorStore) CreateCollection(ctx context.Context, name string) error {
	m.logger.WithField("collection", name).Debug("Creating mock collection")

	m.collections[name] = true
	return nil
}

// DeleteCollection deletes a mock collection
func (m *MockVectorStore) DeleteCollection(ctx context.Context, name string) error {
	m.logger.WithField("collection", name).Debug("Deleting mock collection")

	delete(m.collections, name)
	return nil
}

// ListCollections lists all mock collections
func (m *MockVectorStore) ListCollections(ctx context.Context) ([]string, error) {
	m.logger.Debug("Listing mock collections")

	names := make([]string, 0, len(m.collections))
	for name := range m.collections {
		names = append(names, name)
	}

	return names, nil
}

// calculateDotProduct calculates the dot product similarity between two vectors
func calculateDotProduct(a, b domain.EmbeddingVector) domain.Similarity {
	if len(a) != len(b) {
		return 0
	}

	var sum float32
	for i := range a {
		sum += a[i] * b[i]
	}

	return domain.Similarity(sum)
}