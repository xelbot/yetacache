package yetacache

import (
	"sync"
	"time"
)

type Cache struct {
	cache  map[string]int64
	locker sync.RWMutex

	cleanUpAfter      int64
	cleanupInterval   time.Duration
	defaultExpiration time.Duration
}

func New(defaultExpiration, cleanupInterval time.Duration) *Cache {
	c := &Cache{
		cache:        make(map[string]int64),
		cleanUpAfter: time.Now().Add(cleanupInterval).Unix(),

		cleanupInterval:   cleanupInterval,
		defaultExpiration: defaultExpiration,
	}

	return c
}

// returns true if the cache item exists and has not expired
func (c *Cache) TestItem(id string) bool {
	c.locker.RLock()
	defer c.locker.RUnlock()

	expiration, found := c.cache[id]
	if found {
		result := time.Now().Unix() < expiration

		return result
	}

	return false
}

func (c *Cache) SetItem(id string) {
	c.locker.Lock()
	defer c.locker.Unlock()

	expiration := time.Now().Add(c.defaultExpiration).Unix()
	c.cache[id] = expiration

	now := time.Now().Unix()
	if now > c.cleanUpAfter {
		for key, expirationItem := range c.cache {
			if now > expirationItem {
				delete(c.cache, key)
			}
		}
		c.cleanUpAfter = time.Now().Add(c.cleanupInterval).Unix()
	}
}
