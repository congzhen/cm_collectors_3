package cache

import (
	"container/list"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"reflect"
	"sync"
	"time"
)

const (
	defaultShards     = 256
	defaultMaxSize    = 10000
	cleanupInterval   = 1 * time.Minute
	expirationBuckets = 60
)

var (
	ErrCacheMiss   = fmt.Errorf("cache: key not found")
	ErrInvalidType = fmt.Errorf("cache: invalid value type")
	ErrLockNotHeld = errors.New("lock not held by client")
)

// 时钟接口用于测试
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (c realClock) Now() time.Time { return time.Now() }

type LocalCache struct {
	shards    []*cacheShard
	clock     Clock
	timeWheel *timeWheel
	locks     sync.Map // 保存所有互斥锁
}

type cacheShard struct {
	store     map[string]localCacheItem
	evictList *list.List
	mu        sync.RWMutex
	maxSize   int
}

type localCacheItem struct {
	value      []byte
	expiration time.Time
	element    *list.Element
}

type timeWheel struct {
	buckets []map[string]struct{}
	current int
	mu      sync.RWMutex
}

func NewLocalCache() *LocalCache {
	lc := &LocalCache{
		shards:    make([]*cacheShard, defaultShards),
		clock:     realClock{},
		timeWheel: newTimeWheel(expirationBuckets),
	}

	for i := range lc.shards {
		lc.shards[i] = &cacheShard{
			store:     make(map[string]localCacheItem),
			evictList: list.New(),
			maxSize:   defaultMaxSize / defaultShards,
		}
	}

	go lc.backgroundCleanup()
	return lc
}

func newTimeWheel(buckets int) *timeWheel {
	tw := &timeWheel{
		buckets: make([]map[string]struct{}, buckets),
		current: 0,
	}
	for i := range tw.buckets {
		tw.buckets[i] = make(map[string]struct{})
	}
	return tw
}

func (lc *LocalCache) getShard(key string) *cacheShard {
	h := fnv.New64a()
	h.Write([]byte(key))
	return lc.shards[h.Sum64()%uint64(len(lc.shards))]
}

func (tw *timeWheel) add(key string, expireTime time.Time) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	// 计算过期时间所在的时间槽
	minutes := expireTime.Unix() / 60
	slot := int(minutes % int64(len(tw.buckets)))
	tw.buckets[slot][key] = struct{}{}
}

func (tw *timeWheel) advance(now time.Time) []string {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	currentSlot := int(now.Unix()/60) % len(tw.buckets)
	if currentSlot == tw.current {
		return nil
	}

	expiredKeys := make([]string, 0, len(tw.buckets[tw.current]))
	for k := range tw.buckets[tw.current] {
		expiredKeys = append(expiredKeys, k)
		delete(tw.buckets[tw.current], k)
	}
	tw.current = currentSlot
	return expiredKeys
}

func (lc *LocalCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal error: %w", err)
	}

	shard := lc.getShard(key)
	expireTime := lc.clock.Now().Add(expiration)

	shard.mu.Lock()
	defer shard.mu.Unlock()

	// LRU淘汰
	if shard.evictList.Len() >= shard.maxSize {
		if elem := shard.evictList.Back(); elem != nil {
			evictKey := elem.Value.(string)
			delete(shard.store, evictKey)
			shard.evictList.Remove(elem)
		}
	}

	// 更新时间轮
	lc.timeWheel.add(key, expireTime)

	// 写入缓存
	elem := shard.evictList.PushFront(key)
	shard.store[key] = localCacheItem{
		value:      data,
		expiration: expireTime,
		element:    elem,
	}
	return nil
}

func (lc *LocalCache) Get(ctx context.Context, key string, dest any) error {
	shard := lc.getShard(key)

	shard.mu.RLock()
	item, exists := shard.store[key]
	shard.mu.RUnlock()

	if !exists {
		return ErrCacheMiss
	}

	now := lc.clock.Now()
	if now.After(item.expiration) {
		shard.mu.Lock()
		defer shard.mu.Unlock()
		delete(shard.store, key)
		shard.evictList.Remove(item.element)
		return ErrCacheMiss
	}

	// 更新LRU位置
	shard.mu.Lock()
	shard.evictList.MoveToFront(item.element)
	shard.mu.Unlock()

	if dest == nil {
		return ErrInvalidType
	}
	val := reflect.ValueOf(dest)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("dest requires pointer type, got %T", dest)
	}

	if val.IsNil() {
		return errors.New("dest pointer cannot be nil")
	}

	if err := json.Unmarshal(item.value, dest); err != nil {
		return fmt.Errorf("cache unmarshal error: %w", err)
	}
	return nil
}

func (lc *LocalCache) Delete(ctx context.Context, key string) error {
	shard := lc.getShard(key)

	shard.mu.Lock()
	defer shard.mu.Unlock()

	if item, exists := shard.store[key]; exists {
		delete(shard.store, key)
		shard.evictList.Remove(item.element)
	}
	return nil
}

func (lc *LocalCache) Exists(ctx context.Context, key string) (bool, error) {
	shard := lc.getShard(key)

	shard.mu.RLock()
	defer shard.mu.RUnlock()

	item, exists := shard.store[key]
	if exists && lc.clock.Now().Before(item.expiration) {
		return true, nil
	}
	return false, nil
}

func (lc *LocalCache) Clear(ctx context.Context) error {
	for _, shard := range lc.shards {
		shard.mu.Lock()
		shard.store = make(map[string]localCacheItem)
		shard.evictList.Init()
		shard.mu.Unlock()
	}
	return nil
}

func (lc *LocalCache) Ping(ctx context.Context) error {
	return nil
}

func (lc *LocalCache) backgroundCleanup() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		expiredKeys := lc.timeWheel.advance(lc.clock.Now())
		for _, key := range expiredKeys {
			shard := lc.getShard(key)
			shard.mu.Lock()
			if item, exists := shard.store[key]; exists {
				delete(shard.store, key)
				shard.evictList.Remove(item.element)
			}
			shard.mu.Unlock()
		}
	}
}

func (lc *LocalCache) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	mu, _ := lc.locks.LoadOrStore(key, &sync.Mutex{})
	lock := mu.(*sync.Mutex)

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		if lock.TryLock() {
			time.AfterFunc(ttl, func() { lock.Unlock() }) // 自动过期
			return true, nil
		}
		return false, nil
	}
}

func (lc *LocalCache) Unlock(ctx context.Context, key string) error {
	mu, exists := lc.locks.Load(key)
	if !exists {
		return ErrLockNotHeld
	}
	mu.(*sync.Mutex).Unlock()
	return nil
}
