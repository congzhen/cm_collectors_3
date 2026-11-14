package importbatch

import (
	"context"
	"sync"
)

type libraryLock struct {
	token chan struct{}
	refs  int
}
type LibraryLocker struct {
	mu    sync.Mutex
	locks map[string]*libraryLock
}

var LibraryLocks = &LibraryLocker{locks: make(map[string]*libraryLock)}

// Serialize check-and-write for imports into the same library, including episode
// grouping. Waiting is cancellable and unused locks are removed.
func (l *LibraryLocker) Acquire(ctx context.Context, library string) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	l.mu.Lock()
	if l.locks == nil {
		l.locks = make(map[string]*libraryLock)
	}
	entry := l.locks[library]
	if entry == nil {
		entry = &libraryLock{token: make(chan struct{}, 1)}
		entry.token <- struct{}{}
		l.locks[library] = entry
	}
	entry.refs++
	l.mu.Unlock()
	drop := func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		entry.refs--
		if entry.refs == 0 {
			delete(l.locks, library)
		}
	}
	select {
	case <-ctx.Done():
		drop()
		return nil, ctx.Err()
	case <-entry.token:
	}
	var once sync.Once
	release := func() { once.Do(func() { entry.token <- struct{}{}; drop() }) }
	if err := ctx.Err(); err != nil {
		release()
		return nil, err
	}
	return release, nil
}
