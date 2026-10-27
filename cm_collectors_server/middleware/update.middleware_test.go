package middleware

import (
	"cm_collectors_server/internal/maintenance"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestUpdateMaintenancePreservesBusinessOutsideInstallation(t *testing.T) {
	maintenance.Resume()
	t.Cleanup(maintenance.Resume)
	router := gin.New()
	router.Use(UpdateMaintenance())
	for _, path := range []string{"/api/resource/dataList", "/api/video/mp4/test/video.mp4", "/api/update/status", "/api/health", "/"} {
		router.GET(path, func(c *gin.Context) { c.String(http.StatusOK, "unchanged") })
	}
	request := func(path string, want int) {
		t.Helper()
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != want {
			t.Fatalf("%s: got %d, want %d", path, response.Code, want)
		}
		if want == http.StatusOK && response.Body.String() != "unchanged" {
			t.Fatalf("%s: response changed", path)
		}
	}
	for _, path := range []string{"/api/resource/dataList", "/api/video/mp4/test/video.mp4"} {
		request(path, http.StatusOK)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := maintenance.Drain(ctx); err != nil {
		t.Fatalf("completed requests leaked task counters: %v", err)
	}
	request("/api/resource/dataList", http.StatusServiceUnavailable)
	for _, path := range []string{"/api/update/status", "/api/health", "/"} {
		request(path, http.StatusOK)
	}
	maintenance.Resume()
	request("/api/resource/dataList", http.StatusOK)
	request("/api/video/mp4/test/video.mp4", http.StatusOK)
}

func TestUpdateMaintenanceLetsAnAdmittedRequestFinish(t *testing.T) {
	maintenance.Resume()
	t.Cleanup(maintenance.Resume)
	entered := make(chan struct{})
	release := make(chan struct{})
	defer close(release)
	router := gin.New()
	router.Use(UpdateMaintenance())
	router.GET("/api/existing-task", func(c *gin.Context) {
		close(entered)
		<-release
		c.String(http.StatusOK, "finished")
	})
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/existing-task", nil))
		done <- response
	}()
	<-entered
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := maintenance.Drain(ctx); err != context.DeadlineExceeded {
		t.Fatalf("installation did not wait for the existing request: %v", err)
	}
	// Cancelling installation resumes admissions without cancelling the existing request.
	maintenance.Resume()
	release <- struct{}{}
	select {
	case response := <-done:
		if response.Code != http.StatusOK || response.Body.String() != "finished" {
			t.Fatal("existing request was interrupted")
		}
	case <-time.After(time.Second):
		t.Fatal("existing request did not complete")
	}
}
