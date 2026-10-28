package processors

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Completed segments are reused only briefly, in memory. The first request
// still streams FFmpeg output immediately rather than waiting for the full TS.
// Bound both retained data and concurrent transcodes during rapid seeking.
var playbackSegments = newVideoSegmentCache(64*1024*1024, 8*1024*1024, 2*time.Minute, 3)

type videoSegmentEntry struct {
	data []byte
	used time.Time
}

type videoSegmentCache struct {
	mu       sync.Mutex
	entries  map[string]videoSegmentEntry
	flights  map[string]chan struct{}
	bytes    int
	maxBytes int
	maxEntry int
	ttl      time.Duration
	slots    chan struct{}
}

func newVideoSegmentCache(maxBytes, maxEntry int, ttl time.Duration, concurrency int) *videoSegmentCache {
	return &videoSegmentCache{
		entries: make(map[string]videoSegmentEntry), flights: make(map[string]chan struct{}),
		maxBytes: maxBytes, maxEntry: maxEntry, ttl: ttl, slots: make(chan struct{}, concurrency),
	}
}

func videoSegmentKey(src string, start, duration float64) (string, error) {
	stat, err := os.Stat(src)
	if err != nil {
		return "", err
	}
	// Check the current file under the existing media read lock. A replacement
	// must never receive a segment generated from the previous file snapshot.
	return fmt.Sprintf("%s:%d:%d:%g:%g", filepath.Clean(src), stat.Size(), stat.ModTime().UnixNano(), start, duration), nil
}

func (s *videoSegmentCache) prune(now time.Time) {
	for key, entry := range s.entries {
		if now.Sub(entry.used) >= s.ttl {
			delete(s.entries, key)
			s.bytes -= len(entry.data)
		}
	}
}

// stream coalesces requests for the same segment. Waiters can cancel without
// cancelling the owner; if the owner disconnects, the next waiter retries.
func (s *videoSegmentCache) stream(ctx context.Context, key string, out io.Writer, produce func(io.Writer) error) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		s.mu.Lock()
		now := time.Now()
		s.prune(now)
		if entry, ok := s.entries[key]; ok {
			entry.used = now
			s.entries[key] = entry
			s.mu.Unlock()
			_, err := io.Copy(out, bytes.NewReader(entry.data))
			return err
		}
		if done, ok := s.flights[key]; ok {
			s.mu.Unlock()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-done:
				continue
			}
		}
		done := make(chan struct{})
		s.flights[key] = done
		s.mu.Unlock()
		return s.generate(ctx, key, out, produce, done)
	}
}

func (s *videoSegmentCache) generate(ctx context.Context, key string, out io.Writer, produce func(io.Writer) error, done chan struct{}) error {
	defer func() {
		s.mu.Lock()
		delete(s.flights, key)
		close(done)
		s.mu.Unlock()
	}()
	select {
	case s.slots <- struct{}{}:
		defer func() { <-s.slots }()
	case <-ctx.Done():
		return ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	buffer := &segmentCapture{limit: s.maxEntry}
	err := produce(io.MultiWriter(out, buffer))
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if buffer.overflow || buffer.Len() == 0 || buffer.Len() > s.maxBytes {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prune(time.Now())
	for s.bytes+buffer.Len() > s.maxBytes {
		var oldest string
		var used time.Time
		for k, entry := range s.entries {
			if used.IsZero() || entry.used.Before(used) {
				oldest, used = k, entry.used
			}
		}
		s.bytes -= len(s.entries[oldest].data)
		delete(s.entries, oldest)
	}
	s.entries[key] = videoSegmentEntry{data: buffer.Bytes(), used: time.Now()}
	s.bytes += buffer.Len()
	return nil
}

type segmentCapture struct {
	bytes.Buffer
	limit    int
	overflow bool
}

func (b *segmentCapture) Write(data []byte) (int, error) {
	if !b.overflow {
		if b.Len()+len(data) > b.limit {
			b.overflow = true
			b.Buffer = bytes.Buffer{}
		} else {
			return b.Buffer.Write(data)
		}
	}
	// Large segments still stream normally; they simply aren't retained.
	return len(data), nil
}
