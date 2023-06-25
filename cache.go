package yetacache

import (
	"sync"
	"time"
)

const DefaultTTL time.Duration = 0

type cacheItem[V any] struct {
	item      V
	expiresAt int64
}

type Cache[K comparable, V any] struct {
	items  map[K]cacheItem[V]
	locker sync.RWMutex
	wg     sync.WaitGroup
	ttl    time.Duration
	stop   chan bool

	cleanupInterval time.Duration
}

// New creates a new instance of cache.
func New[K comparable, V any](defaultExpiration, cleanupInterval time.Duration) *Cache[K, V] {
	c := &Cache[K, V]{
		items:           make(map[K]cacheItem[V]),
		ttl:             defaultExpiration,
		cleanupInterval: cleanupInterval,
		stop:            make(chan bool),
	}

	c.wg.Add(1)
	go func(cleanupInterval time.Duration) {
		defer c.wg.Done()
		c.cleanupLoop(cleanupInterval)
	}(cleanupInterval)

	return c
}

// Has returns true if the cache item exists and has not expired.
func (c *Cache[K, V]) Has(id K) bool {
	c.locker.RLock()
	defer c.locker.RUnlock()

	item, found := c.items[id]
	if found {
		return time.Now().Unix() < item.expiresAt
	}

	return false
}

// Get retrieves an item from the cache by the provided key.
// The second return value indicates if the cache item exists and
// has not expired.
func (c *Cache[K, V]) Get(id K) (V, bool) {
	c.locker.RLock()
	defer c.locker.RUnlock()

	var val V
	item, found := c.items[id]
	if found {
		val = item.item

		return val, time.Now().Unix() < item.expiresAt
	}

	return val, false
}

// Set creates a new item from the provided key and value, adds
// it to the cache. If an item associated with the provided key already
// exists, the new item overwrites the existing one.
func (c *Cache[K, V]) Set(id K, value V, ttl time.Duration) {
	c.locker.Lock()
	defer c.locker.Unlock()

	if ttl == DefaultTTL {
		ttl = c.ttl
	}

	c.items[id] = cacheItem[V]{
		item:      value,
		expiresAt: time.Now().Add(ttl).Unix(),
	}
}

// Delete deletes an item from the cache.
func (c *Cache[K, V]) Delete(id K) {
	c.locker.Lock()
	defer c.locker.Unlock()

	delete(c.items, id)
}

// Clear deletes all items from the cache.
func (c *Cache[K, V]) Clear() {
	c.locker.Lock()
	defer c.locker.Unlock()

	if len(c.items) > 0 {
		for key := range c.items {
			delete(c.items, key)
		}
	}
}

// StopCleanup stops the internal cleanup cycle
// that removes expired items.
func (c *Cache[K, V]) StopCleanup() {
	close(c.stop)
	c.wg.Wait()
}

func (c *Cache[K, V]) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var now int64
	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			c.locker.Lock()
			now = time.Now().Unix()
			for key, item := range c.items {
				if now > item.expiresAt {
					delete(c.items, key)
				}
			}
			c.locker.Unlock()
		}
	}
}
