package selfupdate

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"testing"
	"time"
)

type testTransport func(*http.Request) (*http.Response, error)

func (f testTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func mockHTTP(t *testing.T, body string, status int) *int {
	t.Helper()
	old := http.DefaultTransport
	calls := new(int)
	http.DefaultTransport = testTransport(func(r *http.Request) (*http.Response, error) {
		*calls++
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: r}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = old })
	return calls
}
func TestResolvePinsReleaseAndDigest(t *testing.T) {
	name := "cm_collectors_3_linux.zip"
	if runtime.GOOS == "windows" {
		name = "cm_collectors_3_windows.zip"
	}
	fixed := "https://github.com/congzhen/cm_collectors_3/releases/download/v3.4.0/" + name
	body := fmt.Sprintf(`{"tag_name":"v3.4.0","assets":[{"name":%q,"browser_download_url":%q,"size":123,"digest":"sha256:%s"}]}`, name, fixed, strings.Repeat("ab", 32))
	mockHTTP(t, body, 200)
	m := Manifest{Version: "3.4.0", WindowsURL: "https://github.com/congzhen/cm_collectors_3/releases/latest/download/" + name, LinuxURL: "https://github.com/congzhen/cm_collectors_3/releases/latest/download/" + name}
	a, e := Resolve(context.Background(), m)
	if e != nil {
		t.Fatal(e)
	}
	if a.URL != fixed || a.Size != 123 || a.SHA256 != strings.Repeat("ab", 32) {
		t.Fatalf("unexpected asset: %+v", a)
	}
}
func TestResolveRejectsReleaseVersionRace(t *testing.T) {
	mockHTTP(t, `{"tag_name":"v3.5.0","assets":[]}`, 200)
	u := "https://github.com/congzhen/cm_collectors_3/releases/latest/download/app.zip"
	if _, e := Resolve(context.Background(), Manifest{Version: "3.4.0", WindowsURL: u, LinuxURL: u}); e == nil {
		t.Fatal("latest version race accepted")
	}
}
func TestCheckFailureCachedAndRetriedAfterCooldown(t *testing.T) {
	calls := mockHTTP(t, `unavailable`, 503)
	m := &Manager{hooks: Hooks{ConfigURL: func() string {
		return "https://raw.githubusercontent.com/congzhen/cm_collectors_3/master/updateSoftConfig.json"
	}}}
	for i := 0; i < 2; i++ {
		if _, e := m.Check(context.Background(), false); e == nil {
			t.Fatal("expected check failure")
		}
	}
	if *calls != 1 {
		t.Fatalf("no failure cache: %d", *calls)
	}
	m.checked = time.Now().Add(-2 * time.Minute)
	_, _ = m.Check(context.Background(), false)
	if *calls != 2 {
		t.Fatalf("did not retry after cooldown: %d", *calls)
	}
}
func TestCheckRejectsMalformedVersion(t *testing.T) {
	mockHTTP(t, `{"version":"garbage"}`, 200)
	if _, e := Check(context.Background(), "https://raw.githubusercontent.com/congzhen/cm_collectors_3/master/updateSoftConfig.json"); e == nil {
		t.Fatal("invalid version accepted")
	}
}
