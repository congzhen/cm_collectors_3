package selfupdate

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Version is set by release builds with -ldflags -X.
var Version = "3.3.0"
var versionPattern = regexp.MustCompile(`^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

type Manifest struct {
	Version    string   `json:"version"`
	Changelog  []string `json:"changelog"`
	WindowsURL string   `json:"windowsDownloadUrl"`
	LinuxURL   string   `json:"linuxDownloadUrl"`
	DockerURL  string   `json:"dockerDownloadUrl"`
}
type CheckResult struct {
	Manifest
	CurrentVersion string `json:"currentVersion"`
	Platform       string `json:"platform"`
	HasUpdate      bool   `json:"hasUpdate"`
}
type Asset struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

func Compare(a, b string) (int, error) {
	if !versionPattern.MatchString(a) || !versionPattern.MatchString(b) {
		return 0, fmt.Errorf("无效的稳定版本号")
	}
	x, y := strings.Split(strings.TrimPrefix(a, "v"), "."), strings.Split(strings.TrimPrefix(b, "v"), ".")
	for i := range x {
		n, e := strconv.ParseUint(x[i], 10, 64)
		if e != nil {
			return 0, e
		}
		m, e := strconv.ParseUint(y[i], 10, 64)
		if e != nil {
			return 0, e
		}
		if n > m {
			return 1, nil
		}
		if n < m {
			return -1, nil
		}
	}
	return 0, nil
}
func githubURL(raw string, download bool) error {
	u, e := url.Parse(raw)
	if e != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" {
		return fmt.Errorf("更新地址必须是 GitHub HTTPS 地址")
	}
	allowed := u.Host == "github.com" || u.Host == "raw.githubusercontent.com" || u.Host == "api.github.com"
	if download {
		allowed = allowed || u.Host == "release-assets.githubusercontent.com" || u.Host == "objects.githubusercontent.com"
	}
	if !allowed {
		return fmt.Errorf("不支持的更新域名: %s", u.Host)
	}
	return nil
}
func client(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout, CheckRedirect: func(r *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("更新地址重定向过多")
		}
		return githubURL(r.URL.String(), true)
	}}
}
func getJSON(ctx context.Context, raw string, target any) error {
	if e := githubURL(raw, false); e != nil {
		return e
	}
	r, e := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if e != nil {
		return e
	}
	r.Header.Set("User-Agent", "CM-Collectors-Updater")
	r.Header.Set("Accept", "application/json")
	resp, e := client(8 * time.Second).Do(r)
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("GitHub 返回 %d", resp.StatusCode)
	}
	b, e := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024+1))
	if e != nil {
		return e
	}
	if len(b) > 2*1024*1024 {
		return fmt.Errorf("更新信息过大")
	}
	return json.Unmarshal(b, target)
}
func Check(ctx context.Context, raw string) (CheckResult, error) {
	r := CheckResult{CurrentVersion: Version, Platform: runtime.GOOS + "/" + runtime.GOARCH}
	if e := getJSON(ctx, raw, &r.Manifest); e != nil {
		return r, e
	}
	n, e := Compare(r.Version, Version)
	r.HasUpdate = n > 0
	return r, e
}

// Resolve pins a legacy /latest/download URL to a release and GitHub's asset digest.
func Resolve(ctx context.Context, m Manifest) (Asset, error) {
	var a Asset
	if runtime.GOARCH != "amd64" || (runtime.GOOS != "windows" && runtime.GOOS != "linux") {
		return a, fmt.Errorf("仅支持 Windows/Linux amd64")
	}
	raw := m.LinuxURL
	if runtime.GOOS == "windows" {
		raw = m.WindowsURL
	}
	if e := githubURL(raw, false); e != nil {
		return a, e
	}
	u, _ := url.Parse(raw)
	p := strings.Split(strings.Trim(u.Path, "/"), "/")
	if u.Host != "github.com" || len(p) != 6 || p[2] != "releases" {
		return a, fmt.Errorf("下载地址须指向 GitHub Release ZIP 附件")
	}
	endpoint := "https://api.github.com/repos/" + p[0] + "/" + p[1] + "/releases/"
	if p[3] == "latest" && p[4] == "download" {
		endpoint += "latest"
	} else if p[3] == "download" {
		endpoint += "tags/" + url.PathEscape(p[4])
	} else {
		return a, fmt.Errorf("无效的 Release 下载地址")
	}
	var release struct {
		Tag        string `json:"tag_name"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
		Assets     []struct {
			Name   string `json:"name"`
			URL    string `json:"browser_download_url"`
			Digest string `json:"digest"`
			Size   int64  `json:"size"`
		} `json:"assets"`
	}
	if e := getJSON(ctx, endpoint, &release); e != nil {
		return a, e
	}
	n, e := Compare(release.Tag, m.Version)
	if e != nil || n != 0 || release.Draft || release.Prerelease {
		return a, fmt.Errorf("更新清单与 GitHub 正式发行版本不一致，请稍后重新检测")
	}
	for _, v := range release.Assets {
		if v.Name == p[5] {
			hash := strings.TrimPrefix(v.Digest, "sha256:")
			b, e := hex.DecodeString(hash)
			if e != nil || len(b) != 32 || !strings.HasPrefix(v.Digest, "sha256:") {
				return a, fmt.Errorf("该附件没有 GitHub SHA-256 校验值，请重新发布附件")
			}
			if v.Size <= 0 || v.Size > maxArchiveSize {
				return a, fmt.Errorf("更新包大小不受支持")
			}
			if e := githubURL(v.URL, false); e != nil {
				return a, e
			}
			if strings.Contains(v.URL, "/latest/") {
				return a, fmt.Errorf("附件地址未固定版本")
			}
			return Asset{URL: v.URL, SHA256: hash, Size: v.Size}, nil
		}
	}
	return a, fmt.Errorf("未找到当前平台的更新附件")
}
