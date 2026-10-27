package maintenance

import (
	"context"
	"testing"
	"time"
)

func TestDrainWaitsForTransferredChildren(t *testing.T) {
	Resume()
	defer Resume()
	if !Begin() {
		t.Fatal("request rejected")
	}
	release := make(chan struct{})
	Go(func() { <-release })
	End()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- Drain(ctx) }()
	for !Paused() {
		time.Sleep(time.Millisecond)
	}
	if Begin() {
		End()
		t.Fatal("new request admitted during drain")
	}
	select {
	case <-done:
		t.Fatal("drained before child finished")
	default:
	}
	close(release)
	if e := <-done; e != nil {
		t.Fatal(e)
	}
}
func TestCancelledDrainCanResume(t *testing.T) {
	Resume()
	defer Resume()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if Drain(ctx) == nil {
		t.Fatal("cancelled empty drain succeeded")
	}
	Resume()
	if !Begin() {
		t.Fatal("did not resume")
	}
	End()
}
