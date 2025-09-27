package processorscache

import (
	"fmt"
	"sync"
	"time"
)

// GenericLastUseCache 是一个基于最后使用时间的缓存
type GenericLastUseCache[T any] struct {
	prefixKey       string
	entries         map[string]*genericCacheEntry[T]
	mu              sync.RWMutex
	expiration      time.Duration
	cleanupInterval time.Duration
}

// genericCacheEntry 代表缓存中的一个条目
type genericCacheEntry[T any] struct {
	value    T
	lastUsed int64 // 最后使用时间戳（毫秒）
}

// NewGenericLastUseCache 创建一个新的最后使用时间缓存
// expiration: 条目过期时间
// cleanupInterval: 清理过期间隔
func NewGenericLastUseCache[T any](prefixKey string, expiration, cleanupInterval time.Duration) *GenericLastUseCache[T] {
	cache := &GenericLastUseCache[T]{
		prefixKey:       prefixKey,
		entries:         make(map[string]*genericCacheEntry[T]),
		expiration:      expiration,
		cleanupInterval: cleanupInterval,
	}

	// 启动后台清理goroutine
	go cache.startCleanupRoutine()

	return cache
}

// startCleanupRoutine 启动定期清理过期条目的goroutine
func (c *GenericLastUseCache[T]) startCleanupRoutine() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanupExpiredEntries()
	}
}

// Get 获取缓存的值
func (c *GenericLastUseCache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fullKey := c.getFullKey(key)
	entry, exists := c.entries[fullKey]
	if !exists {
		var zero T
		return zero, false
	}

	// 更新最后使用时间
	entry.lastUsed = c.getCurrentTimeMillis()
	return entry.value, true
}

// Set 设置缓存的值
func (c *GenericLastUseCache[T]) Set(key string, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	fullKey := c.getFullKey(key)
	entry := &genericCacheEntry[T]{
		value:    value,
		lastUsed: c.getCurrentTimeMillis(),
	}
	c.entries[fullKey] = entry
}

// Delete 删除指定键的缓存条目
func (c *GenericLastUseCache[T]) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	fullKey := c.getFullKey(key)
	delete(c.entries, fullKey)
}

// cleanupExpiredEntries 清理过期的缓存条目
func (c *GenericLastUseCache[T]) cleanupExpiredEntries() {
	c.mu.Lock()
	defer c.mu.Unlock()

	currentTime := c.getCurrentTimeMillis()
	expirationMillis := int64(c.expiration / time.Millisecond)

	for key, entry := range c.entries {
		if entry.lastUsed < currentTime-expirationMillis {
			fmt.Println("######################### 释放缓存:", key)
			delete(c.entries, key)
		}
	}
}

// getCurrentTimeMillis 获取当前时间的毫秒数
func (c *GenericLastUseCache[T]) getCurrentTimeMillis() int64 {
	return time.Now().UnixNano() / 1e6
}

// getFullKey 获取带前缀的完整键名
func (c *GenericLastUseCache[T]) getFullKey(key string) string {
	if c.prefixKey == "" {
		return key
	}
	return c.prefixKey + ":" + key
}

// Size 返回缓存中条目的数量
func (c *GenericLastUseCache[T]) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}
