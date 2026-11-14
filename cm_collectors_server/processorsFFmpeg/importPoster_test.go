package processorsffmpeg

import (
	"bytes"
	"context"
	"errors"
	"image"
	_ "image/jpeg"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestPosterMJPEGTruncatedTail(t *testing.T) {
	data := []byte{0xff, 0xd8, 1, 2, 3, 4, 0xff, 0xd9, 0xff, 0xd8, 1}
	frames := (Thumbnail{}).splitMJPEGStream(data)
	if len(frames) != 1 || len(frames[0]) != 8 {
		t.Fatalf("unexpected frames: %v", frames)
	}
	if len((Thumbnail{}).splitMJPEGStream([]byte{0xff, 0xd8})) != 0 {
		t.Fatal("accepted incomplete JPEG")
	}
}

func TestPosterRealFFmpeg(t *testing.T) {
	name := "ffmpeg"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	tool, err := exec.LookPath(name)
	if err != nil {
		tool, err = filepath.Abs(filepath.Join("..", "ffmpeg", name))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = os.Stat(tool); err != nil {
			t.Skip("FFmpeg unavailable")
		}
	}
	t.Setenv("PATH", filepath.Dir(tool)+string(os.PathListSeparator)+os.Getenv("PATH"))
	video := filepath.Join(t.TempDir(), "sample.mp4")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := createCommandContext(ctx, tool, "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc2=size=1280x720:rate=25", "-t", "5", "-c:v", "mpeg4", "-threads", "1", video)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create fixture: %v %s", err, out)
	}
	data, err := (Thumbnail{}).ExtractThumbnailPosterContext(ctx, video)
	if err != nil {
		t.Fatal(err)
	}
	c, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || c.Width != 400 || c.Height <= 0 {
		t.Fatalf("poster geometry: %+v %v", c, err)
	}
	deadline, stop := context.WithTimeout(context.Background(), time.Nanosecond)
	defer stop()
	if _, err = (Thumbnail{}).ExtractThumbnailPosterContext(deadline, video); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline not propagated: %v", err)
	}
}
func TestPosterCancellationBeforeLaunch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	posterSlots <- struct{}{}
	defer func() { <-posterSlots }()
	if _, err := (Thumbnail{}).ExtractThumbnailPosterContext(ctx, "unused"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
func TestBoundedPosterOutput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b := &boundedOutput{limit: 4, cancel: cancel}
	if n, err := b.Write([]byte("123456")); n != 6 || err != nil {
		t.Fatal(n, err)
	}
	if b.String() != "1234" || !b.exceeded || ctx.Err() == nil {
		t.Fatal("unbounded output or no cancellation")
	}
	b.Write([]byte("more"))
	if b.Len() != 4 {
		t.Fatal("buffer grew")
	}
}

func TestBoundedPosterOutputThroughIOCopy(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out := &boundedOutput{limit: 4, cancel: cancel}
	// os/exec forwards pipe data through io.Copy, which may use ReaderFrom.
	n, err := io.Copy(out, struct{ io.Reader }{strings.NewReader("123456789")})
	if err != nil || n != 9 || out.Len() != 4 || !out.exceeded || ctx.Err() == nil {
		t.Fatalf("limit bypassed: n=%d len=%d exceeded=%v canceled=%v err=%v", n, out.Len(), out.exceeded, ctx.Err(), err)
	}
}

func TestPosterToolAvailabilityHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := (FFmpeg{}).isToolAvailableContext(ctx, "unused-tool"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
