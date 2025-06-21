package vector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/sirupsen/logrus"
)

// ChromaDBVectorStore implements the VectorStore interface using ChromaDB
type ChromaDBVectorStore struct {
	baseURL      string
	collection   string
	collectionID string
	tenant       string
	database     string
	dataPath     string
	autoStart    bool
	client       *http.Client
	logger       *logrus.Logger
}

// ChromaDBConfig holds configuration for ChromaDB provider
type ChromaDBConfig struct {
	BaseURL    string        `json:"base_url"`
	Collection string        `json:"collection"`
	Tenant     string        `json:"tenant"`
	Database   string        `json:"database"`
	Timeout    time.Duration `json:"timeout"`
	DataPath   string        `json:"data_path"`   // Expected ChromaDB data directory
	AutoStart  bool          `json:"auto_start"`  // Whether to auto-start ChromaDB if not running
}

// DefaultChromeDBConfig returns default configuration for ChromaDB
func DefaultChromeDBConfig() ChromaDBConfig {
	return ChromaDBConfig{
		BaseURL:    "http://localhost:8000",
		Collection: "memory_bank",
		Tenant:     "default_tenant",
		Database:   "default_database",
		Timeout:    30 * time.Second,
		DataPath:   "./chromadb_data",  // Default data directory
		AutoStart:  false,              // Don't auto-start by default
	}
}

// NewChromaDBVectorStore creates a new ChromaDB vector store
func NewChromaDBVectorStore(config ChromaDBConfig, logger *logrus.Logger) *ChromaDBVectorStore {
	if config.BaseURL == "" {
		config = DefaultChromeDBConfig()
	}

	// Configure HTTP client with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	return &ChromaDBVectorStore{
		baseURL:    config.BaseURL,
		collection: config.Collection,
		tenant:     config.Tenant,
		database:   config.Database,
		dataPath:   config.DataPath,
		autoStart:  config.AutoStart,
		client: &http.Client{
			Timeout:   config.Timeout,
			Transport: transport,
		},
		logger: logger,
	}
}

// getCollectionID retrieves the UUID for a collection by name
func (c *ChromaDBVectorStore) getCollectionID(ctx context.Context, name string) (string, error) {
	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections",
		c.baseURL, c.tenant, c.database)

	c.logger.WithFields(logrus.Fields{
		"url":  url,
		"name": name,
	}).Info("Fetching collection ID")

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to list collections: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.WithError(err).Warn("Failed to close response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"body":   string(body),
		}).Error("ChromaDB collections API failed")
		return "", fmt.Errorf("ChromaDB API error (status %d): %s", resp.StatusCode, string(body))
	}

	var collections []chromaDBCollection
	if err := json.NewDecoder(resp.Body).Decode(&collections); err != nil {
		return "", fmt.Errorf("failed to decode collections response: %w", err)
	}

	c.logger.WithField("collections_count", len(collections)).Debug("Retrieved collections")

	for _, col := range collections {
		c.logger.WithFields(logrus.Fields{
			"collection_name": col.Name,
			"collection_id":   col.ID,
		}).Debug("Checking collection")

		if col.Name == name {
			c.logger.WithField("found_id", col.ID).Debug("Found collection ID")
			return col.ID, nil
		}
	}

	return "", fmt.Errorf("collection %s not found", name)
}

// ensureCollectionID ensures the collection ID is loaded
func (c *ChromaDBVectorStore) ensureCollectionID(ctx context.Context) error {
	c.logger.WithFields(logrus.Fields{
		"current_collection_id": c.collectionID,
		"collection_name":       c.collection,
	}).Debug("Ensuring collection ID is loaded")

	if c.collectionID != "" {
		c.logger.Debug("Collection ID already loaded")
		return nil
	}

	c.logger.Debug("Loading collection ID")
	id, err := c.getCollectionID(ctx, c.collection)
	if err != nil {
		// If collection doesn't exist, create it
		if strings.Contains(err.Error(), "not found") {
			c.logger.WithField("collection", c.collection).Info("Collection not found, creating it")
			if createErr := c.CreateCollection(ctx, c.collection); createErr != nil {
				c.logger.WithError(createErr).Error("Failed to create collection")
				return fmt.Errorf("failed to create collection: %w", createErr)
			}
			// Try to get the collection ID again after creation
			id, err = c.getCollectionID(ctx, c.collection)
			if err != nil {
				c.logger.WithError(err).Error("Failed to get collection ID after creation")
				return err
			}
		} else {
			c.logger.WithError(err).Error("Failed to get collection ID")
			return err
		}
	}

	c.collectionID = id
	c.logger.WithField("loaded_id", id).Debug("Collection ID loaded successfully")
	return nil
}

// buildCollectionOperationURL builds the full URL for collection operations using collection ID
func (c *ChromaDBVectorStore) buildCollectionOperationURL(ctx context.Context, operation string) (string, error) {
	// Ensure collection ID is loaded
	if err := c.ensureCollectionID(ctx); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections/%s/%s",
		c.baseURL, c.tenant, c.database, c.collectionID, operation), nil
}

// chromaDBDocument represents a document in ChromaDB
type chromaDBDocument struct {
	IDs        []string                 `json:"ids"`
	Documents  []string                 `json:"documents,omitempty"`
	Metadatas  []map[string]interface{} `json:"metadatas,omitempty"`
	Embeddings [][]float32              `json:"embeddings,omitempty"`
}

// chromaDBQueryRequest represents a query request to ChromaDB
type chromaDBQueryRequest struct {
	QueryEmbeddings [][]float32            `json:"query_embeddings"`
	NResults        int                    `json:"n_results"`
	Include         []string               `json:"include,omitempty"`
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

// normalizeMetadata converts metadata to ChromaDB-compatible format
func (c *ChromaDBVectorStore) normalizeMetadata(metadata map[string]interface{}) map[string]interface{} {
	normalized := make(map[string]interface{})

	for key, value := range metadata {
		c.logger.WithFields(logrus.Fields{
			"key":   key,
			"value": value,
			"type":  fmt.Sprintf("%T", value),
		}).Debug("Processing metadata field")

		switch v := value.(type) {
		case []string:
			// Convert string slices to comma-separated strings
			normalized[key] = strings.Join(v, ",")
		case []interface{}:
			// Convert interface slices to comma-separated strings
			var strSlice []string
			for _, item := range v {
				strSlice = append(strSlice, fmt.Sprintf("%v", item))
			}
			normalized[key] = strings.Join(strSlice, ",")
		case time.Time:
			// Convert time to string
			normalized[key] = v.Format(time.RFC3339)
		default:
			// Try to detect if it's a slice that wasn't caught above
			valueStr := fmt.Sprintf("%v", value)
			if strings.HasPrefix(valueStr, "[") && strings.HasSuffix(valueStr, "]") {
				// Likely a slice, convert to string representation
				normalized[key] = strings.Trim(valueStr, "[]")
			} else {
				// Keep other types as-is (string, int, float, bool)
				normalized[key] = value
			}
		}
	}

	return normalized
}

// Store stores a vector with metadata in ChromaDB
func (c *ChromaDBVectorStore) Store(ctx context.Context, id string, vector domain.EmbeddingVector, metadata map[string]interface{}) error {
	c.logger.WithFields(logrus.Fields{
		"collection":    c.collection,
		"id":            id,
		"vector_length": len(vector),
		"metadata_keys": len(metadata),
	}).Debug("Storing vector in ChromaDB")

	// Normalize metadata for ChromaDB compatibility
	normalizedMetadata := c.normalizeMetadata(metadata)

	// Prepare document
	doc := chromaDBDocument{
		IDs:        []string{id},
		Embeddings: [][]float32{vector},
		Metadatas:  []map[string]interface{}{normalizedMetadata},
	}

	// Convert to JSON
	jsonBody, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	// Create HTTP request
	url, err := c.buildCollectionOperationURL(ctx, "add")
	if err != nil {
		return fmt.Errorf("failed to build collection URL: %w", err)
	}
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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.WithError(err).Warn("Failed to close response body")
		}
	}()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)

		// If collection doesn't exist (404), try to create it and retry
		if resp.StatusCode == http.StatusNotFound {
			c.logger.WithField("collection", c.collection).Info("Collection not found, creating it")
			if err := c.CreateCollection(ctx, c.collection); err != nil {
				return fmt.Errorf("failed to create collection: %w", err)
			}

			// Reset collection ID so it gets loaded again with the new collection
			c.collectionID = ""

			// Retry the store operation
			return c.Store(ctx, id, vector, metadata)
		}

		return fmt.Errorf("chromadb API error (status %d): %s", resp.StatusCode, string(body))
	}

	c.logger.WithField("id", id).Debug("Vector stored successfully in ChromaDB")
	return nil
}

// BatchStore stores multiple vectors with metadata in ChromaDB in a single request
func (c *ChromaDBVectorStore) BatchStore(ctx context.Context, items []ports.BatchStoreItem) error {
	if len(items) == 0 {
		return nil
	}

	c.logger.WithFields(logrus.Fields{
		"collection": c.collection,
		"batch_size": len(items),
	}).Debug("Batch storing vectors in ChromaDB")

	// Prepare batch document
	doc := chromaDBDocument{
		IDs:        make([]string, len(items)),
		Embeddings: make([][]float32, len(items)),
		Metadatas:  make([]map[string]interface{}, len(items)),
	}

	for i, item := range items {
		doc.IDs[i] = item.ID
		doc.Embeddings[i] = item.Vector
		doc.Metadatas[i] = c.normalizeMetadata(item.Metadata)
	}

	// Convert to JSON
	jsonBody, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal batch document: %w", err)
	}

	// Create HTTP request
	url, err := c.buildCollectionOperationURL(ctx, "add")
	if err != nil {
		return fmt.Errorf("failed to build collection URL: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute batch request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.WithError(err).Warn("Failed to close response body")
		}
	}()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chromadb batch API error (status %d): %s", resp.StatusCode, string(body))
	}

	c.logger.WithField("batch_size", len(items)).Debug("Batch vectors stored successfully in ChromaDB")
	return nil
}

// BatchDelete removes multiple vectors from ChromaDB in a single request
func (c *ChromaDBVectorStore) BatchDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	c.logger.WithFields(logrus.Fields{
		"collection": c.collection,
		"batch_size": len(ids),
	}).Debug("Batch deleting vectors from ChromaDB")

	// Prepare delete request
	deleteDoc := map[string]interface{}{
		"ids": ids,
	}

	jsonBody, err := json.Marshal(deleteDoc)
	if err != nil {
		return fmt.Errorf("failed to marshal batch delete request: %w", err)
	}

	// Create HTTP request
	url, err := c.buildCollectionOperationURL(ctx, "delete")
	if err != nil {
		return fmt.Errorf("failed to build collection URL: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute batch delete request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.WithError(err).Warn("Failed to close response body")
		}
	}()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chromadb batch delete API error (status %d): %s", resp.StatusCode, string(body))
	}

	c.logger.WithField("batch_size", len(ids)).Debug("Batch vectors deleted successfully from ChromaDB")
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
	url, err := c.buildCollectionOperationURL(ctx, "delete")
	if err != nil {
		return fmt.Errorf("failed to build collection URL: %w", err)
	}
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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.WithError(err).Warn("Failed to close response body")
		}
	}()

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
		"collection":    c.collection,
		"id":            id,
		"vector_length": len(vector),
		"metadata_keys": len(metadata),
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
	url, err := c.buildCollectionOperationURL(ctx, "update")
	if err != nil {
		return fmt.Errorf("failed to build collection URL: %w", err)
	}
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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.WithError(err).Warn("Failed to close response body")
		}
	}()

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
		"collection":    c.collection,
		"vector_length": len(vector),
		"limit":         limit,
		"threshold":     threshold,
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
	url, err := c.buildCollectionOperationURL(ctx, "query")
	if err != nil {
		return nil, fmt.Errorf("failed to build collection URL: %w", err)
	}
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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.WithError(err).Warn("Failed to close response body")
		}
	}()

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

	// Prepare collection request with cosine distance metric
	collectionReq := map[string]interface{}{
		"name": name,
		"metadata": map[string]interface{}{
			"description": "Memory Bank collection",
			"hnsw:space":  "cosine", // Use cosine distance metric
		},
		"configuration": map[string]interface{}{
			"hnsw": map[string]interface{}{
				"space": "cosine", // Use cosine distance metric
			},
		},
	}

	jsonBody, err := json.Marshal(collectionReq)
	if err != nil {
		return fmt.Errorf("failed to marshal collection request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections",
		c.baseURL, c.tenant, c.database)
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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.WithError(err).Warn("Failed to close response body")
		}
	}()

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
	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections/%s", c.baseURL, c.tenant, c.database, name)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.WithError(err).Warn("Failed to close response body")
		}
	}()

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
	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections",
		c.baseURL, c.tenant, c.database)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.WithError(err).Warn("Failed to close response body")
		}
	}()

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

	// First check if ChromaDB is running
	err := c.checkChromaDBConnection(ctx)
	if err != nil {
		// If auto-start is enabled and ChromaDB is not running, try to start it
		if c.autoStart && c.isConnectionError(err) {
			c.logger.Info("ChromaDB not running, attempting to auto-start...")
			if startErr := c.startChromaDB(ctx); startErr != nil {
				return fmt.Errorf("chromadb not running and auto-start failed: %w", startErr)
			}
			// Retry connection after starting
			if retryErr := c.checkChromaDBConnection(ctx); retryErr != nil {
				return fmt.Errorf("chromadb auto-start succeeded but connection still failed: %w", retryErr)
			}
		} else {
			return err
		}
	}

	// Validate ChromaDB configuration
	if err := c.validateChromaDBConfiguration(ctx); err != nil {
		c.logger.WithError(err).Warn("ChromaDB configuration validation failed")
		// Don't fail health check for configuration warnings, just log them
	}

	c.logger.Info("ChromaDB health check passed")
	return nil
}

// checkChromaDBConnection checks if ChromaDB is accessible
func (c *ChromaDBVectorStore) checkChromaDBConnection(ctx context.Context) error {
	// Use the dedicated heartbeat endpoint for health checks
	url := fmt.Sprintf("%s/api/v2/heartbeat", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("chromadb connection failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.WithError(err).Warn("Failed to close response body")
		}
	}()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chromadb health check failed (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// isConnectionError checks if the error is a connection error (vs configuration error)
func (c *ChromaDBVectorStore) isConnectionError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "connection failed") ||
		strings.Contains(errStr, "timeout")
}

// validateChromaDBConfiguration checks if the running ChromaDB instance matches expected configuration
func (c *ChromaDBVectorStore) validateChromaDBConfiguration(ctx context.Context) error {
	// Check if the expected data path exists and is being used
	if c.dataPath != "" {
		absDataPath, err := filepath.Abs(c.dataPath)
		if err != nil {
			return fmt.Errorf("failed to resolve data path %s: %w", c.dataPath, err)
		}

		// Check if data directory exists
		if _, err := os.Stat(absDataPath); os.IsNotExist(err) {
			return fmt.Errorf("expected ChromaDB data directory does not exist: %s", absDataPath)
		}

		// Try to detect if ChromaDB is using the expected path by checking running processes
		if err := c.validateChromaDBProcessPath(absDataPath); err != nil {
			return fmt.Errorf("chromadb process validation failed: %w", err)
		}
	}

	return nil
}

// validateChromaDBProcessPath checks if any running ChromaDB process uses the expected data path
func (c *ChromaDBVectorStore) validateChromaDBProcessPath(expectedPath string) error {
	// Use ps to find ChromaDB processes
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list processes: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var chromaProcesses []string
	
	for _, line := range lines {
		if strings.Contains(line, "chroma run") && strings.Contains(line, "--path") {
			chromaProcesses = append(chromaProcesses, line)
		}
	}

	if len(chromaProcesses) == 0 {
		return fmt.Errorf("no running ChromaDB processes found with --path argument")
	}

	// Check if any process uses the expected path
	for _, process := range chromaProcesses {
		if strings.Contains(process, expectedPath) {
			c.logger.WithFields(logrus.Fields{
				"expected_path": expectedPath,
				"process":       process,
			}).Debug("Found ChromaDB process with matching data path")
			return nil
		}
	}

	// Log all found processes for debugging
	c.logger.WithFields(logrus.Fields{
		"expected_path":     expectedPath,
		"found_processes":   chromaProcesses,
	}).Warn("ChromaDB running with different data path than expected")

	return fmt.Errorf("chromadb process found but using different data path (expected: %s)", expectedPath)
}

// startChromaDB attempts to start ChromaDB with the correct configuration
func (c *ChromaDBVectorStore) startChromaDB(ctx context.Context) error {
	if c.dataPath == "" {
		return fmt.Errorf("cannot auto-start ChromaDB: no data path configured")
	}

	// Ensure data directory exists
	absDataPath, err := filepath.Abs(c.dataPath)
	if err != nil {
		return fmt.Errorf("failed to resolve data path: %w", err)
	}

	if err := os.MkdirAll(absDataPath, 0755); err != nil {
		return fmt.Errorf("failed to create data directory %s: %w", absDataPath, err)
	}

	c.logger.WithField("data_path", absDataPath).Info("Starting ChromaDB server...")

	// Extract host and port from baseURL
	host := "localhost"
	port := "8000"
	if strings.Contains(c.baseURL, "://") {
		parts := strings.Split(c.baseURL, "://")
		if len(parts) > 1 {
			hostPort := parts[1]
			if strings.Contains(hostPort, ":") {
				hostPortParts := strings.Split(hostPort, ":")
				host = hostPortParts[0]
				port = hostPortParts[1]
			}
		}
	}

	// Try to start ChromaDB using uvx (preferred) or direct chroma command
	cmd := exec.CommandContext(ctx, "uvx", "--from", "chromadb[server]", "chroma", "run", 
		"--host", host, "--port", port, "--path", absDataPath)
	
	// Start process in background
	if err := cmd.Start(); err != nil {
		// Fallback to direct chroma command if uvx is not available
		cmd = exec.CommandContext(ctx, "chroma", "run", 
			"--host", host, "--port", port, "--path", absDataPath)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start ChromaDB (tried uvx and direct chroma): %w", err)
		}
	}

	c.logger.WithFields(logrus.Fields{
		"host":      host,
		"port":      port,
		"data_path": absDataPath,
		"pid":       cmd.Process.Pid,
	}).Info("ChromaDB server started successfully")

	// Wait a moment for the server to start
	time.Sleep(2 * time.Second)

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
		"id":            id,
		"vector_length": len(vector),
		"metadata_keys": len(metadata),
	}).Debug("Storing vector in mock store")

	m.vectors[id] = mockVectorEntry{
		Vector:   vector,
		Metadata: metadata,
	}

	return nil
}

// BatchStore stores multiple vectors with metadata in the mock store
func (m *MockVectorStore) BatchStore(ctx context.Context, items []ports.BatchStoreItem) error {
	m.logger.WithField("batch_size", len(items)).Debug("Batch storing vectors in mock store")

	for _, item := range items {
		m.vectors[item.ID] = mockVectorEntry{
			Vector:   item.Vector,
			Metadata: item.Metadata,
		}
	}

	return nil
}

// BatchDelete removes multiple vectors from the mock store
func (m *MockVectorStore) BatchDelete(ctx context.Context, ids []string) error {
	m.logger.WithField("batch_size", len(ids)).Debug("Batch deleting vectors from mock store")

	for _, id := range ids {
		delete(m.vectors, id)
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
		"id":            id,
		"vector_length": len(vector),
		"metadata_keys": len(metadata),
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

// HealthCheck always returns success for the mock vector store
func (m *MockVectorStore) HealthCheck(ctx context.Context) error {
	m.logger.Debug("Mock vector store health check (always passes)")
	return nil
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
