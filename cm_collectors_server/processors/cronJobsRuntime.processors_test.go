package processors

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestCronJobExecutionRegistryRejectsConcurrentStart(t *testing.T) {
	registry := newCronJobExecutionRegistry()

	const attempts = 32
	ready := make(chan struct{})
	var started int32
	var wg sync.WaitGroup
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ready
			if registry.tryStart("job-1") {
				atomic.AddInt32(&started, 1)
			}
		}()
	}

	close(ready)
	wg.Wait()

	if started != 1 {
		t.Fatalf("expected exactly one start, got %d", started)
	}
	if !registry.isRunning("job-1") {
		t.Fatal("expected job-1 to be reported as running")
	}
}

func TestCronJobExecutionRegistryReleasesJob(t *testing.T) {
	registry := newCronJobExecutionRegistry()
	if !registry.tryStart("job-1") {
		t.Fatal("expected first start to succeed")
	}

	registry.finish("job-1")

	if registry.isRunning("job-1") {
		t.Fatal("expected job-1 to be released")
	}
	if !registry.tryStart("job-1") {
		t.Fatal("expected job-1 to start again after release")
	}
}

func TestCronJobExecutionRegistryAllowsDifferentJobs(t *testing.T) {
	registry := newCronJobExecutionRegistry()
	if !registry.tryStart("job-1") {
		t.Fatal("expected job-1 to start")
	}
	if !registry.tryStart("job-2") {
		t.Fatal("expected a different job to start concurrently")
	}
}
