package importbatch

import (
	"fmt"
	"testing"
	"time"
)

func TestImportBatchLifecycle(t *testing.T) {
	g := New()
	now := time.Unix(100, 0)
	g.now = func() time.Time { return now }
	runs := 0
	start := func() error { runs++; return nil }
	g.Begin("one", "library")
	g.Begin("two", "library")
	for i := 0; i < 320; i++ {
		if !g.Touch("one", "library") {
			t.Fatal("live batch expired")
		}
		release := g.Hold()
		if ok, _ := g.AdmitBackup(start); ok {
			t.Fatal("backup admitted during import")
		}
		release()
		release() // release must be idempotent
		now = now.Add(time.Millisecond)
	}
	g.End("one")
	if ok, _ := g.AdmitBackup(start); ok {
		t.Fatal("second batch still active")
	}
	g.End("two")
	now = now.Add(QuietWindow - time.Second)
	if ok, _ := g.AdmitBackup(start); ok {
		t.Fatal("quiet window ignored")
	}
	now = now.Add(time.Second)
	if ok, err := g.AdmitBackup(start); !ok || err != nil || runs != 1 {
		t.Fatalf("backup not resumed: %v %v %d", ok, err, runs)
	}
}

func TestAbandonedBatchAndInflightRequest(t *testing.T) {
	g := New()
	now := time.Unix(100, 0)
	g.now = func() time.Time { return now }
	g.Begin("id", "library")
	if g.Touch("id", "other") {
		t.Fatal("accepted other library")
	}
	release := g.Hold()
	now = now.Add(Lease + time.Second)
	if g.Touch("id", "library") {
		t.Fatal("revived expired lease")
	}
	if ok, _ := g.AdmitBackup(func() error { return nil }); ok {
		t.Fatal("inflight request unprotected")
	}
	release()
	now = now.Add(QuietWindow)
	if ok, _ := g.AdmitBackup(func() error { return nil }); !ok {
		t.Fatal("abandoned batch did not release")
	}
}

func TestAdmissionPropagatesFailure(t *testing.T) {
	g := New()
	want := fmt.Errorf("database unavailable")
	if ok, err := g.AdmitBackup(func() error { return want }); !ok || err != want {
		t.Fatal(ok, err)
	}
}
