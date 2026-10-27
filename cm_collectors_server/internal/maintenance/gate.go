// Package maintenance coordinates update installation with requests and background jobs.
package maintenance

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrPaused = errors.New("正在准备安装更新，请稍后重试")
var gate struct {
	sync.Mutex
	paused bool
	active int
}

func Begin() bool {
	gate.Lock()
	defer gate.Unlock()
	if gate.paused {
		return false
	}
	gate.active++
	return true
}
func End() { gate.Lock(); gate.active--; gate.Unlock() }

// Go transfers work from an already admitted request/job to a child goroutine.
func Go(fn func()) {
	gate.Lock()
	gate.active++
	gate.Unlock()
	go func() { defer End(); fn() }()
}
func Paused() bool { gate.Lock(); defer gate.Unlock(); return gate.paused }
func Resume()      { gate.Lock(); gate.paused = false; gate.Unlock() }
func Drain(ctx context.Context) error {
	gate.Lock()
	gate.paused = true
	gate.Unlock()
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		gate.Lock()
		n := gate.active
		gate.Unlock()
		if n == 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
