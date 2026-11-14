// Package importbatch tracks live import batches without persisting abandoned UI tasks.
package importbatch

import (
	"sync"
	"time"
)

const Lease = 90 * time.Second
const QuietWindow = 30 * time.Second

type entry struct {
	library string
	expires time.Time
}
type Gate struct {
	sync.Mutex
	entries    map[string]entry
	inflight   int
	quietUntil time.Time
	now        func() time.Time
}

func New() *Gate { return &Gate{entries: make(map[string]entry), now: time.Now} }

var Global = New()

func (g *Gate) Begin(id, library string) {
	g.Lock()
	defer g.Unlock()
	g.expireLocked(g.now())
	g.entries[id] = entry{library, g.now().Add(Lease)}
}
func (g *Gate) Touch(id, library string) bool {
	g.Lock()
	defer g.Unlock()
	e, ok := g.entries[id]
	if !ok || e.library != library || !e.expires.After(g.now()) {
		return false
	}
	e.expires = g.now().Add(Lease)
	g.entries[id] = e
	return true
}
func (g *Gate) End(id string) {
	g.Lock()
	defer g.Unlock()
	if _, ok := g.entries[id]; ok {
		delete(g.entries, id)
		g.quietUntil = g.now().Add(QuietWindow)
	}
}

// Hold protects in-flight requests and scheduled imports even if a UI lease expires.
func (g *Gate) Hold() func() {
	g.Lock()
	g.inflight++
	g.Unlock()
	var once sync.Once
	return func() {
		once.Do(func() { g.Lock(); defer g.Unlock(); g.inflight--; g.quietUntil = g.now().Add(QuietWindow) })
	}
}

// AdmitBackup serializes backup admission with batch registration. Already admitted
// backups can finish normally; manual backups do not use this gate.
func (g *Gate) AdmitBackup(start func() error) (bool, error) {
	g.Lock()
	defer g.Unlock()
	now := g.now()
	g.expireLocked(now)
	if len(g.entries) > 0 || g.inflight > 0 || now.Before(g.quietUntil) {
		return false, nil
	}
	return true, start()
}

func (g *Gate) expireLocked(now time.Time) {
	for id, e := range g.entries {
		if !e.expires.After(now) {
			delete(g.entries, id)
			g.quietUntil = now.Add(QuietWindow)
		}
	}
}
