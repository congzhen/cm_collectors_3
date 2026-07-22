package processors

import "sync"

// cronJobExecutionRegistry records jobs that are currently executing in this
// server process. Manual execution and cron callbacks share this registry, so
// the same plan cannot enter its task body twice at the same time.
type cronJobExecutionRegistry struct {
	mu      sync.RWMutex
	running map[string]struct{}
}

var cronJobExecutions = newCronJobExecutionRegistry()

func newCronJobExecutionRegistry() *cronJobExecutionRegistry {
	return &cronJobExecutionRegistry{running: make(map[string]struct{})}
}

func (r *cronJobExecutionRegistry) tryStart(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.running[id]; exists {
		return false
	}
	r.running[id] = struct{}{}
	return true
}

func (r *cronJobExecutionRegistry) finish(id string) {
	r.mu.Lock()
	delete(r.running, id)
	r.mu.Unlock()
}

func (r *cronJobExecutionRegistry) isRunning(id string) bool {
	r.mu.RLock()
	_, exists := r.running[id]
	r.mu.RUnlock()
	return exists
}
