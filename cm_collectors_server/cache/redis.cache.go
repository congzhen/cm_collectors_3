package cache

import (
	"cm_collectors_server/config"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client        *redis.Client
	locker        *redislock.Client // 锁客户端
	defaultExpire time.Duration     // 默认过期时间

	mu          sync.Mutex // 保护activeLocks
	activeLocks map[string]*redislock.Lock
}

func NewRedisCache(config config.RedisCache) *RedisCache {
	redisClient := redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:        config.Password,
		DB:              config.Db,
		MinIdleConns:    config.MinIdleConns,                                 // 最小空闲连接
		PoolSize:        config.PoolSize,                                     // 连接池大小
		ConnMaxIdleTime: time.Duration(config.ConnMaxIdleTime) * time.Second, // 最大空闲时间
	})
	return &RedisCache{
		client:        redisClient,
		locker:        redislock.New(redisClient),                           // 初始化锁客户端
		defaultExpire: time.Duration(config.DefaultExpireSec) * time.Second, // 默认过期时间
	}
}

func (rc *RedisCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	if expiration == 0 {
		expiration = rc.defaultExpire
	}
	// 统一序列化处理
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal error: %w", err)
	}
	return rc.client.Set(ctx, key, data, expiration).Err()
}

func (rc *RedisCache) Get(ctx context.Context, key string, dest any) error {
	data, err := rc.client.Get(ctx, key).Bytes()
	switch {
	case errors.Is(err, redis.Nil):
		return ErrCacheMiss
	case err != nil:
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

func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	return rc.client.Del(ctx, key).Err()
}

func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	exists, err := rc.client.Exists(ctx, key).Result()
	return exists == 1, err
}

func (rc *RedisCache) Clear(ctx context.Context) error {
	return rc.client.FlushDB(ctx).Err()
}

func (rc *RedisCache) Ping(ctx context.Context) error {
	return rc.client.Ping(ctx).Err()
}

func (rc *RedisCache) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	lockKey := "lock:" + key
	lock, err := rc.locker.Obtain(ctx, lockKey, ttl, &redislock.Options{
		RetryStrategy: redislock.LinearBackoff(100 * time.Millisecond),
	})

	switch {
	case errors.Is(err, redislock.ErrNotObtained):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("lock acquisition failed: %w", err)
	}

	if rc.activeLocks == nil {
		rc.activeLocks = make(map[string]*redislock.Lock)
	}
	rc.activeLocks[lockKey] = lock
	return true, nil
}

func (rc *RedisCache) Unlock(ctx context.Context, key string) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	lockKey := "lock:" + key
	lock, exists := rc.activeLocks[lockKey]
	if !exists {
		return errors.New("lock not found in active locks")
	}

	if err := lock.Release(ctx); err != nil {
		return fmt.Errorf("lock release failed: %w", err)
	}

	delete(rc.activeLocks, lockKey)
	return nil
}
