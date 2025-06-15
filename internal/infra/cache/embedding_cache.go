package cache

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/joern1811/memory-bank/internal/domain"
	"github.com/joern1811/memory-bank/internal/ports"
	"github.com/sirupsen/logrus"
)

// EmbeddingCache wraps an embedding provider with caching capabilities
type EmbeddingCache struct {
	cache      map[string]CacheEntry
	mutex      sync.RWMutex
	maxSize    int
	ttl        time.Duration
	underlying ports.EmbeddingProvider
	logger     *logrus.Logger
	
	// Stats for monitoring
	hits   int64
	misses int64
}

// CacheEntry represents a cached embedding with metadata
type CacheEntry struct {
	Embedding domain.EmbeddingVector
	CreatedAt time.Time
	LastUsed  time.Time
}

// CacheConfig holds configuration for the embedding cache
type CacheConfig struct {
	MaxSize int           `json:"max_size"`
	TTL     time.Duration `json:"ttl"`
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		MaxSize: 1000,
		TTL:     1 * time.Hour,
	}
}

// NewEmbeddingCache creates a new embedding cache that wraps an underlying provider
func NewEmbeddingCache(underlying ports.EmbeddingProvider, config CacheConfig, logger *logrus.Logger) *EmbeddingCache {
	if config.MaxSize <= 0 {
		config = DefaultCacheConfig()
	}

	cache := &EmbeddingCache{
		cache:      make(map[string]CacheEntry),
		maxSize:    config.MaxSize,
		ttl:        config.TTL,
		underlying: underlying,
		logger:     logger,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	logger.WithFields(logrus.Fields{
		"max_size": config.MaxSize,
		"ttl":      config.TTL,
	}).Info("Embedding cache initialized")

	return cache
}

// GenerateEmbedding generates an embedding with caching
func (c *EmbeddingCache) GenerateEmbedding(ctx context.Context, text string) (domain.EmbeddingVector, error) {
	key := c.hashText(text)

	// Check cache first
	c.mutex.RLock()
	if entry, exists := c.cache[key]; exists && c.isEntryValid(entry) {
		// Update last used time
		entry.LastUsed = time.Now()
		c.cache[key] = entry
		c.hits++
		c.mutex.RUnlock()
		
		c.logger.WithFields(logrus.Fields{
			"cache_key": key[:8] + "...",
			"text_len":  len(text),
		}).Debug("Cache hit for embedding")
		
		return entry.Embedding, nil
	}
	c.mutex.RUnlock()

	// Cache miss - generate new embedding
	c.logger.WithFields(logrus.Fields{
		"cache_key": key[:8] + "...",
		"text_len":  len(text),
	}).Debug("Cache miss for embedding")

	embedding, err := c.underlying.GenerateEmbedding(ctx, text)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.mutex.Lock()
	c.misses++
	
	// Check if we need to evict entries
	if len(c.cache) >= c.maxSize {
		c.evictOldest()
	}
	
	c.cache[key] = CacheEntry{
		Embedding: embedding,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
	}
	c.mutex.Unlock()

	c.logger.WithField("cache_key", key[:8]+"...").Debug("Embedding cached")
	return embedding, nil
}

// GenerateBatchEmbeddings generates embeddings for multiple texts with caching
func (c *EmbeddingCache) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([]domain.EmbeddingVector, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	c.logger.WithField("batch_size", len(texts)).Debug("Generating batch embeddings with caching")

	embeddings := make([]domain.EmbeddingVector, len(texts))
	missingIndices := make([]int, 0)
	missingTexts := make([]string, 0)

	// Check cache for each text
	for i, text := range texts {
		key := c.hashText(text)
		
		c.mutex.RLock()
		if entry, exists := c.cache[key]; exists && c.isEntryValid(entry) {
			// Cache hit
			entry.LastUsed = time.Now()
			c.cache[key] = entry
			embeddings[i] = entry.Embedding
			c.hits++
			c.mutex.RUnlock()
		} else {
			// Cache miss
			c.mutex.RUnlock()
			missingIndices = append(missingIndices, i)
			missingTexts = append(missingTexts, text)
			c.misses++
		}
	}

	// Generate embeddings for cache misses
	if len(missingTexts) > 0 {
		c.logger.WithFields(logrus.Fields{
			"cache_hits":   len(texts) - len(missingTexts),
			"cache_misses": len(missingTexts),
			"hit_rate":     float64(len(texts)-len(missingTexts)) / float64(len(texts)) * 100,
		}).Debug("Batch cache performance")

		newEmbeddings, err := c.underlying.GenerateBatchEmbeddings(ctx, missingTexts)
		if err != nil {
			return nil, err
		}

		// Store new embeddings in cache
		c.mutex.Lock()
		for i, embedding := range newEmbeddings {
			originalIndex := missingIndices[i]
			text := missingTexts[i]
			key := c.hashText(text)
			
			// Check if we need to evict entries
			if len(c.cache) >= c.maxSize {
				c.evictOldest()
			}
			
			c.cache[key] = CacheEntry{
				Embedding: embedding,
				CreatedAt: time.Now(),
				LastUsed:  time.Now(),
			}
			
			embeddings[originalIndex] = embedding
		}
		c.mutex.Unlock()
	}

	c.logger.WithField("batch_size", len(embeddings)).Debug("Batch embeddings with caching completed")
	return embeddings, nil
}

// GetDimensions returns the dimension size from the underlying provider
func (c *EmbeddingCache) GetDimensions() int {
	return c.underlying.GetDimensions()
}

// GetModelName returns the model name from the underlying provider
func (c *EmbeddingCache) GetModelName() string {
	return c.underlying.GetModelName()
}

// GetStats returns cache statistics
func (c *EmbeddingCache) GetStats() (hits, misses int64, hitRate float64) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	total := c.hits + c.misses
	hitRate = 0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total) * 100
	}
	
	return c.hits, c.misses, hitRate
}

// ClearCache clears all cached embeddings
func (c *EmbeddingCache) ClearCache() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.cache = make(map[string]CacheEntry)
	c.hits = 0
	c.misses = 0
	
	c.logger.Info("Embedding cache cleared")
}

// hashText generates a consistent hash for text caching
func (c *EmbeddingCache) hashText(text string) string {
	hash := sha256.Sum256([]byte(text))
	return fmt.Sprintf("%x", hash)
}

// isEntryValid checks if a cache entry is still valid (not expired)
func (c *EmbeddingCache) isEntryValid(entry CacheEntry) bool {
	return time.Since(entry.CreatedAt) < c.ttl
}

// evictOldest removes the oldest entry from the cache (LRU eviction)
func (c *EmbeddingCache) evictOldest() {
	if len(c.cache) == 0 {
		return
	}
	
	var oldestKey string
	var oldestTime time.Time
	first := true
	
	for key, entry := range c.cache {
		if first || entry.LastUsed.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.LastUsed
			first = false
		}
	}
	
	delete(c.cache, oldestKey)
	c.logger.WithField("evicted_key", oldestKey[:8]+"...").Debug("Evicted oldest cache entry")
}

// cleanup runs periodically to remove expired entries
func (c *EmbeddingCache) cleanup() {
	ticker := time.NewTicker(c.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()
		evicted := 0
		
		for key, entry := range c.cache {
			if now.Sub(entry.CreatedAt) > c.ttl {
				delete(c.cache, key)
				evicted++
			}
		}
		
		if evicted > 0 {
			c.logger.WithField("evicted_count", evicted).Debug("Cleaned up expired cache entries")
		}
		
		c.mutex.Unlock()
	}
}