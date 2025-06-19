// freecache.cache.go
package cache

import (
	"cm_collectors_server/config"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coocood/freecache"
)

type FreeCache struct {
	cache         *freecache.Cache
	defaultExpire time.Duration // 默认过期时间
	locks         sync.Map      // 新增锁存储 map[string]*sync.Mutex
}

func NewFreeCache(config config.FreeCache) *FreeCache {
	size := config.MaxMemoryMB * 1024 * 1024
	if size == 0 {
		size = 100 * 1024 * 1024 // 默认100MB
	}
	return &FreeCache{
		cache:         freecache.NewCache(size),
		defaultExpire: time.Duration(config.DefaultExpireSec) * time.Second,
	}
}

func (fc *FreeCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	if expiration == 0 {
		expiration = fc.defaultExpire
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal error: %w", err)
	}

	return fc.cache.Set([]byte(key), data, int(expiration.Seconds()))
}

func (fc *FreeCache) Get(ctx context.Context, key string, dest any) error {
	data, err := fc.cache.Get([]byte(key))
	if err != nil {
		if errors.Is(err, freecache.ErrNotFound) {
			return ErrCacheMiss
		}
		return fmt.Errorf("cache get error: %w", err)
	}

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

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("cache unmarshal error: %w", err)
	}
	return nil
}

func (fc *FreeCache) Delete(ctx context.Context, key string) error {
	fc.cache.Del([]byte(key))
	return nil
}

func (fc *FreeCache) Exists(ctx context.Context, key string) (bool, error) {
	_, err := fc.cache.Get([]byte(key))
	if err != nil {
		if errors.Is(err, freecache.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (fc *FreeCache) Clear(ctx context.Context) error {
	fc.cache.Clear()
	return nil
}

func (fc *FreeCache) Ping(ctx context.Context) error {
	// FreeCache无需实际ping操作
	return nil
}

type lockInfo struct {
	mutex    sync.Mutex
	released atomic.Bool // 标记锁是否已被手动释放
}

// 实现 Cache 接口的 Lock/Unlock
func (fc *FreeCache) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	// 使用 LoadOrStore 获取或创建 lockInfo
	lockI, _ := fc.locks.LoadOrStore(key, &lockInfo{})
	lock := lockI.(*lockInfo)

	// 尝试加锁
	if lock.mutex.TryLock() {
		// 启动超时清理协程
		timer := time.AfterFunc(ttl, func() {
			if !lock.released.Load() {
				// 仅在未被手动释放时执行解锁和清理
				lock.mutex.Unlock()
				fc.locks.Delete(key)
			}
		})
		// 如果上下文被取消，提前停止定时器
		go func() {
			select {
			case <-ctx.Done():
				timer.Stop()
				// 如果未被手动释放，手动清理
				if !lock.released.Load() {
					lock.mutex.Unlock()
					fc.locks.Delete(key)
				}
			default:
			}
		}()
		return true, nil
	}
	return false, nil
}

func (fc *FreeCache) Unlock(ctx context.Context, key string) error {
	lockI, ok := fc.locks.Load(key)
	if !ok {
		return errors.New("lock not found")
	}
	lock := lockI.(*lockInfo)

	// 标记为已释放，并解锁
	lock.released.Store(true)
	lock.mutex.Unlock()

	// 删除锁并停止定时器（如果存在）
	fc.locks.Delete(key)
	return nil
}
