package cache

import (
	"cm_collectors_server/config"
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrCacheMiss   = fmt.Errorf("cache: key not found")
	ErrInvalidType = fmt.Errorf("cache: invalid value type")
	ErrLockNotHeld = errors.New("lock not held by client")
)

type Cache interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Get(ctx context.Context, key string, dest any) error // 修改返回值为error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error
	Ping(ctx context.Context) error
	Lock(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Unlock(ctx context.Context, key string) error
}

func NewCache(config config.Cache) (Cache, error) {
	switch config.Mode {
	case "redis":
		cache := NewRedisCache(config.Redis)
		if err := cache.Ping(context.Background()); err != nil {
			return nil, err
		}
		return cache, nil
	case "freeCache":
		return NewFreeCache(config.FreeCache), nil
	default:
		return nil, errors.New("unsupported cache type")
	}
}
