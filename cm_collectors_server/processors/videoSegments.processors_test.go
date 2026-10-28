package processors

import (
	"bytes"
	ff "cm_collectors_server/processorsFFmpeg"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestVideoSegmentFFmpegCacheProducesPlayableOutput(t *testing.T) {
	tools, err := filepath.Abs("../ffmpeg")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", tools+string(os.PathListSeparator)+os.Getenv("PATH"))
	tool, err := (ff.FFmpeg{}).IsFFmpegAvailable()
	if err != nil {
		t.Skipf("FFmpeg unavailable: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	src := filepath.Join(t.TempDir(), "source.mp4")
	fixture := ff.CreateCommandContext(ctx, tool, "-nostdin", "-v", "error",
		"-f", "lavfi", "-i", "testsrc2=size=160x90:rate=24",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=48000",
		"-t", "3", "-c:v", "libx264", "-c:a", "aac", "-y", src)
	if data, err := fixture.CombinedOutput(); err != nil {
		t.Fatalf("fixture: %v %s", err, data)
	}
	cache := newVideoSegmentCache(1024*1024, 1024*1024, time.Minute, 1)
	calls := 0
	produce := func(w io.Writer) error {
		calls++
		cmd, err := (ff.M3U8{}).PlayVideoM3u8Context(ctx, src, 1, 1, true, true)
		if err != nil {
			return err
		}
		cmd.Stdout = w
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("segment: %w %s", err, stderr.String())
		}
		return nil
	}
	var first, second bytes.Buffer
	key, err := videoSegmentKey(src, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, out := range []*bytes.Buffer{&first, &second} {
		if err := cache.stream(ctx, key, out, produce); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 1 || first.Len() == 0 || !bytes.Equal(first.Bytes(), second.Bytes()) {
		t.Fatal("重复请求应复用完整的真实 TS 输出")
	}
	decode := ff.CreateCommandContext(ctx, tool, "-nostdin", "-v", "error", "-i", "pipe:0", "-f", "null", "-")
	decode.Stdin = bytes.NewReader(second.Bytes())
	if output, err := decode.CombinedOutput(); err != nil || len(output) != 0 {
		t.Fatalf("缓存分片应能解码音视频: %v %s", err, output)
	}
}

func TestVideoSegmentStreamsThenReusesCompletedOutput(t *testing.T) {
	cache := newVideoSegmentCache(64, 32, time.Minute, 2)
	var calls atomic.Int32
	firstFrame := make(chan struct{})
	release := make(chan struct{})
	produce := func(w io.Writer) error {
		calls.Add(1)
		if _, err := w.Write([]byte("first")); err != nil {
			return err
		}
		close(firstFrame)
		<-release
		_, err := w.Write([]byte("last"))
		return err
	}
	var first bytes.Buffer
	owner := make(chan error, 1)
	go func() { owner <- cache.stream(context.Background(), "same", &first, produce) }()
	<-firstFrame
	// The producer is still running; the client already received the first bytes.
	if first.String() != "first" {
		t.Fatal("首个请求不应等待整个分片转码完成")
	}
	const waiters = 5
	var group sync.WaitGroup
	outputs := make([]bytes.Buffer, waiters)
	errs := make([]error, waiters)
	for i := range outputs {
		group.Add(1)
		go func(i int) {
			defer group.Done()
			errs[i] = cache.stream(context.Background(), "same", &outputs[i], produce)
		}(i)
	}
	close(release)
	if err := <-owner; err != nil {
		t.Fatal(err)
	}
	group.Wait()
	for i := range outputs {
		if errs[i] != nil || outputs[i].String() != "firstlast" {
			t.Fatalf("重复请求结果不完整: %q, %v", outputs[i].String(), errs[i])
		}
	}
	if calls.Load() != 1 {
		t.Fatal("同一分片不应重复转码")
	}
}

func TestVideoSegmentFailureAndOversizeAreNotCached(t *testing.T) {
	for _, mode := range []string{"error", "oversize", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			cache := newVideoSegmentCache(32, 8, time.Minute, 1)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var first bytes.Buffer
			_ = cache.stream(ctx, "key", &first, func(w io.Writer) error {
				_, _ = w.Write([]byte("partial"))
				switch mode {
				case "error":
					return errors.New("转码中断")
				case "cancel":
					cancel()
				default:
					_, _ = w.Write([]byte("too large"))
				}
				return nil
			})
			var next bytes.Buffer
			err := cache.stream(context.Background(), "key", &next, func(w io.Writer) error {
				_, err := w.Write([]byte("new"))
				return err
			})
			if err != nil || next.String() != "new" {
				t.Fatalf("失败、中断或超限数据不应作为完整缓存返回: %q %v", next.String(), err)
			}
		})
	}
}

func TestVideoSegmentBudgetExpiryAndFileVersion(t *testing.T) {
	cache := newVideoSegmentCache(6, 6, time.Minute, 1)
	write := func(key string) {
		t.Helper()
		if err := cache.stream(context.Background(), key, io.Discard, func(w io.Writer) error {
			_, err := w.Write([]byte("1234"))
			return err
		}); err != nil {
			t.Fatal(err)
		}
	}
	write("old")
	write("new")
	if cache.bytes != 4 || len(cache.entries) != 1 || cache.entries["new"].data == nil {
		t.Fatal("缓存应按容量淘汰旧分片")
	}
	cache.prune(time.Now().Add(2 * time.Minute))
	if cache.bytes != 0 || len(cache.entries) != 0 {
		t.Fatal("超时分片应被清理")
	}
	src := filepath.Join(t.TempDir(), "video.mp4")
	if err := os.WriteFile(src, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	oldKey, _ := videoSegmentKey(src, 10, 10)
	if err := os.WriteFile(src, []byte("new content"), 0600); err != nil {
		t.Fatal(err)
	}
	newKey, _ := videoSegmentKey(src, 10, 10)
	otherKey, _ := videoSegmentKey(src, 20, 10)
	if oldKey == newKey || newKey == otherKey {
		t.Fatal("视频替换或定位变化必须隔离缓存")
	}
}

func TestVideoSegmentWaitingRequestsCanCancel(t *testing.T) {
	cache := newVideoSegmentCache(64, 32, time.Minute, 1)
	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- cache.stream(context.Background(), "busy", io.Discard, func(w io.Writer) error {
			close(started)
			<-release
			_, err := w.Write([]byte("ok"))
			return err
		})
	}()
	<-started
	for _, key := range []string{"busy", "different"} {
		ctx, cancel := context.WithCancel(context.Background())
		result := make(chan error, 1)
		go func() {
			result <- cache.stream(ctx, key, io.Discard, func(io.Writer) error {
				t.Error("已取消的等待请求不应执行转码")
				return nil
			})
		}()
		cancel()
		select {
		case err := <-result:
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("应返回取消错误: %v", err)
			}
		case <-time.After(time.Second):
			t.Fatal("等待分片或转码额度时取消未生效")
		}
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if len(cache.flights) != 0 || len(cache.slots) != 0 {
		t.Fatal("完成或取消后应释放任务与额度")
	}
}

func TestVideoSegmentCanceledOwnerAllowsWaitingClientToRetry(t *testing.T) {
	cache := newVideoSegmentCache(64, 32, time.Minute, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{})
	owner := make(chan error, 1)
	go func() {
		owner <- cache.stream(ctx, "same", io.Discard, func(w io.Writer) error {
			_, _ = w.Write([]byte("partial"))
			close(started)
			<-ctx.Done()
			return errors.New("encoder process killed")
		})
	}()
	<-started
	var output bytes.Buffer
	waiter := make(chan error, 1)
	go func() {
		waiter <- cache.stream(context.Background(), "same", &output, func(w io.Writer) error {
			_, err := w.Write([]byte("complete"))
			return err
		})
	}()
	cancel()
	if err := <-owner; !errors.Is(err, context.Canceled) {
		t.Fatalf("编码器终止错误应识别为取消: %v", err)
	}
	select {
	case err := <-waiter:
		if err != nil || output.String() != "complete" {
			t.Fatalf("等待者应重新生成完整数据: %q %v", output.String(), err)
		}
	case <-time.After(time.Second):
		t.Fatal("首请求取消后等待者未恢复")
	}
}
