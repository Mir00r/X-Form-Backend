package secrets

import (
	"sync"
	"time"
)

// cacheEntry represents a cached secret with its expiration time
type cacheEntry struct {
	value     string
	expiresAt time.Time
}

// secretCache implements a simple in-memory cache for secrets
type secretCache struct {
	entries    map[string]cacheEntry
	ttl        time.Duration
	maxEntries int
	mu         sync.RWMutex
	hits       int64
	misses     int64
}

// newSecretCache creates a new secret cache with the given configuration
func newSecretCache(config CacheConfig) *secretCache {
	return &secretCache{
		entries:    make(map[string]cacheEntry),
		ttl:        config.TTL,
		maxEntries: config.MaxEntries,
	}
}

// Get retrieves a value from the cache
func (c *secretCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		c.misses++
		return "", false
	}

	// Check if expired
	if time.Now().After(entry.expiresAt) {
		c.misses++
		delete(c.entries, key)
		return "", false
	}

	c.hits++
	return entry.value, true
}

// Set stores a value in the cache
func (c *secretCache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict entries
	if c.maxEntries > 0 && len(c.entries) >= c.maxEntries {
		c.evictOldest()
	}

	c.entries[key] = cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Delete removes a value from the cache
func (c *secretCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

// Clear removes all entries from the cache
func (c *secretCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]cacheEntry)
	c.hits = 0
	c.misses = 0
}

// Stats returns cache statistics
func (c *secretCache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return map[string]interface{}{
		"entries":  len(c.entries),
		"hits":     c.hits,
		"misses":   c.misses,
		"hit_rate": hitRate,
		"ttl":      c.ttl.String(),
	}
}

// evictOldest removes the oldest entry from the cache
func (c *secretCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, entry := range c.entries {
		if first || entry.expiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.expiresAt
			first = false
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}

// cleanup removes expired entries (should be called periodically)
func (c *secretCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, key)
		}
	}
}

// startCleanupWorker starts a background worker to clean up expired entries
func (c *secretCache) startCleanupWorker() {
	go func() {
		ticker := time.NewTicker(c.ttl / 2) // Clean up twice per TTL period
		defer ticker.Stop()

		for range ticker.C {
			c.cleanup()
		}
	}()
}
