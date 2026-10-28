package controllers

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHLSFailureAfterMediaCleanupIsReported(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest("GET", "/api/video/m3u8/stream/d/0/10/1.ts", nil)
	client := c.Request.Context()
	media, release := context.WithCancel(client)
	c.Request = c.Request.WithContext(media)
	c.Header("Content-Type", "video/MP2T")
	release() // Matches the processor's normal deferred media-reader cleanup.
	finishVideoM3u8Stream(c, client, errors.New("FFmpeg unavailable"))
	if !strings.Contains(recorder.Body.String(), `"status":false`) || !strings.Contains(recorder.Body.String(), "FFmpeg unavailable") {
		t.Fatalf("内部上下文正常清理不能吞掉真实错误: %s", recorder.Body.String())
	}
	if !strings.HasPrefix(recorder.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("流开始前的错误应返回 JSON 类型: %s", recorder.Header().Get("Content-Type"))
	}
}

func TestHLSCanceledAndStartedResponsesAreNotAppended(t *testing.T) {
	for _, mode := range []string{"client-cancel", "media-cancel", "started", "success"} {
		t.Run(mode, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest("GET", "/segment.ts", nil)
			ctx, cancel := context.WithCancel(c.Request.Context())
			defer cancel()
			var err error
			switch mode {
			case "client-cancel":
				cancel()
				err = errors.New("write failed")
			case "media-cancel":
				err = context.Canceled
			case "started":
				_, _ = c.Writer.Write([]byte("TS bytes"))
				err = errors.New("encoder failed")
			}
			finishVideoM3u8Stream(c, ctx, err)
			want := ""
			if mode == "started" {
				want = "TS bytes"
			}
			if recorder.Body.String() != want {
				t.Fatalf("不能追加错误响应: %q", recorder.Body.String())
			}
		})
	}
}
