package processorscache

import (
	"testing"
	"time"
)

type cacheTestCloser struct {
	closed bool
}

func (c *cacheTestCloser) Close() error {
	c.closed = true
	return nil
}

func TestGenericLastUseCacheTakeRemovesValue(t *testing.T) {
	value := &cacheTestCloser{}
	cache := &GenericLastUseCache[*cacheTestCloser]{
		prefixKey:  "test",
		expiration: time.Minute,
		entries: map[string]*genericCacheEntry[*cacheTestCloser]{
			"test:path": {value: value, lastUsed: time.Now().UnixMilli()},
		},
	}
	taken, ok := cache.Take("path")
	if !ok || taken != value {
		t.Fatal("Take should return the stored value")
	}
	if _, exists := cache.Get("path"); exists {
		t.Fatal("Take should remove the stored value")
	}
	if value.closed {
		t.Fatal("Take leaves cleanup ownership to the caller")
	}
}

func TestGenericLastUseCacheCleanupClosesExpiredValue(t *testing.T) {
	value := &cacheTestCloser{}
	cache := &GenericLastUseCache[*cacheTestCloser]{
		prefixKey:  "test",
		expiration: time.Millisecond,
		entries: map[string]*genericCacheEntry[*cacheTestCloser]{
			"test:path": {value: value, lastUsed: 0},
		},
	}
	cache.cleanupExpiredEntries()
	if !value.closed {
		t.Fatal("expired closeable values must be closed after removal")
	}
	if cache.Size() != 0 {
		t.Fatal("expired value should be removed")
	}
}
