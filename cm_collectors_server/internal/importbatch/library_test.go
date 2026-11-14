package importbatch

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLibraryLockCancellationAndIsolation(t *testing.T) {
	l := &LibraryLocker{}
	release, err := l.Acquire(context.Background(), "a")
	if err != nil {
		t.Fatal(err)
	}
	other, err := l.Acquire(context.Background(), "b")
	if err != nil {
		t.Fatal(err)
	}
	other()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err = l.Acquire(ctx, "a"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	release()
	release()
	if len(l.locks) != 0 {
		t.Fatal("idle locks leaked")
	}
}

func TestLibraryConcurrentImportsDoNotOverlap(t *testing.T) {
	l := &LibraryLocker{}
	var wg sync.WaitGroup
	var active atomic.Int32
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			release, err := l.Acquire(context.Background(), "library")
			if err != nil {
				t.Error(err)
				return
			}
			if active.Add(1) != 1 {
				t.Error("overlapping imports")
			}
			time.Sleep(time.Millisecond)
			active.Add(-1)
			release()
		}()
	}
	wg.Wait()
	if len(l.locks) != 0 {
		t.Fatal("idle locks leaked")
	}
}
