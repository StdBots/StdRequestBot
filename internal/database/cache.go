package database

import (
	"fmt"
	"sync"
	"time"
)

// DedupeCache manages thread-safe request deduplication with a sliding time window
type DedupeCache struct {
	handled map[string]time.Time
	mu      sync.RWMutex
	window  time.Duration
}

// NewDedupeCache creates a deduplication cache with a custom time window (e.g. 180s)
func NewDedupeCache(window time.Duration) *DedupeCache {
	c := &DedupeCache{
		handled: make(map[string]time.Time),
		window:  window,
	}
	go c.startCleanup(30 * time.Second)
	return c
}

// AlreadyHandled checks if (chatID, userID) was processed within the dedupe window
func (c *DedupeCache) AlreadyHandled(chatID int64, userID int64) bool {
	key := fmt.Sprintf("%d:%d", chatID, userID)
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	lastTime, exists := c.handled[key]
	if exists && now.Sub(lastTime) < c.window {
		return true // Duplicate within window
	}

	c.handled[key] = now
	return false
}

func (c *DedupeCache) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for k, t := range c.handled {
			if now.Sub(t) >= c.window {
				delete(c.handled, k)
			}
		}
		c.mu.Unlock()
	}
}
