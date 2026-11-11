// Package coverpreview stores a bounded, disposable cache of cover candidates.
package coverpreview

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Cache struct {
	Root string
	mu   sync.Mutex
}
type entry struct {
	Size     int64    `json:"size"`
	Modified int64    `json:"modified"`
	Frames   int      `json:"frames"`
	Images   []string `json:"images"`
}

// Get keys by source path: changing frame count replaces, rather than adds, a resource.
// The lock also coalesces repeated opens while the first extraction is still running.
func (c *Cache) Get(source string, frames int, refresh bool, generate func() ([]string, error)) ([]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	setting, settingErr := c.settingsLocked()
	if settingErr != nil {
		setting = Settings{MaxResources: 5}
	}
	path, err := filepath.Abs(source)
	if err != nil {
		return generate()
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return generate()
	}
	hash := sha256.Sum256([]byte(filepath.Clean(path)))
	name := filepath.Join(c.Root, hex.EncodeToString(hash[:])+".json")
	if !refresh {
		data, err := os.ReadFile(name)
		var saved entry
		if err == nil && json.Unmarshal(data, &saved) == nil && saved.Size == info.Size() && saved.Modified == info.ModTime().UnixNano() && saved.Frames == frames && len(saved.Images) > 0 {
			now := time.Now()
			_ = os.Chtimes(name, now, now)
			c.trim(name, setting.MaxResources-1)
			return saved.Images, nil
		}
	}
	images, err := generate()
	if err != nil || len(images) == 0 {
		return images, err
	}
	// A changed source must not acquire a cache entry for the old version.
	after, err := os.Stat(path)
	if err != nil || after.Size() != info.Size() || !after.ModTime().Equal(info.ModTime()) {
		return images, nil
	}
	if os.MkdirAll(c.Root, 0755) != nil {
		return images, nil
	}
	data, err := json.Marshal(entry{info.Size(), info.ModTime().UnixNano(), frames, images})
	if err != nil {
		return images, nil
	}
	tmp, err := os.CreateTemp(c.Root, ".preview-*")
	if err != nil {
		return images, nil
	}
	defer os.Remove(tmp.Name())
	_, writeErr := tmp.Write(data)
	closeErr := tmp.Close()
	if writeErr != nil || closeErr != nil {
		return images, nil
	}
	// Make room for this resource before publishing the entry.
	if !c.trim(name, setting.MaxResources-1) {
		return images, nil
	}
	// Replace directly so a failed publication does not first delete the old cache.
	_ = os.Rename(tmp.Name(), name)
	return images, nil
}

// Only touch our hash-named entries, never unrelated cache files or subdirectories.
func (c *Cache) trim(keep string, limit int) bool {
	entries, err := os.ReadDir(c.Root)
	if err != nil {
		return false
	}
	type item struct {
		name     string
		modified time.Time
	}
	var files []item
	for _, e := range entries {
		if !isCacheFile(e) {
			continue
		}
		name := filepath.Join(c.Root, e.Name())
		if name == keep {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return false
		}
		files = append(files, item{name, info.ModTime()})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].modified.Before(files[j].modified) })
	for len(files) > limit {
		if os.Remove(files[0].name) != nil {
			return false
		}
		files = files[1:]
	}
	return true
}

type Settings struct {
	MaxResources int `json:"maxResources"`
}
type Status struct {
	Count int   `json:"count"`
	Bytes int64 `json:"bytes"`
}

func (c *Cache) settingsLocked() (Settings, error) {
	value := Settings{MaxResources: 5}
	data, err := os.ReadFile(filepath.Join(c.Root, "settings.json"))
	if os.IsNotExist(err) {
		return value, nil
	}
	if err != nil {
		return value, err
	}
	if err = json.Unmarshal(data, &value); err != nil {
		return value, err
	}
	if value.MaxResources < 1 || value.MaxResources > 100 {
		return value, fmt.Errorf("封面预览缓存数量应为 1 至 100")
	}
	return value, nil
}
func (c *Cache) Settings() (Settings, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.settingsLocked()
}
func (c *Cache) SaveSettings(value Settings) error {
	if value.MaxResources < 1 || value.MaxResources > 100 {
		return fmt.Errorf("封面预览缓存数量应为 1 至 100")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := os.MkdirAll(c.Root, 0755); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(c.Root, ".settings-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	_, writeErr := f.Write(data)
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	if err = os.Rename(f.Name(), filepath.Join(c.Root, "settings.json")); err != nil {
		return err
	}
	if !c.trim("", value.MaxResources) {
		return fmt.Errorf("设置已保存，但清理超额封面预览缓存失败，请重试清理")
	}
	return nil
}
func isCacheFile(e os.DirEntry) bool {
	if e.IsDir() || len(e.Name()) != 69 || !strings.HasSuffix(e.Name(), ".json") {
		return false
	}
	_, err := hex.DecodeString(strings.TrimSuffix(e.Name(), ".json"))
	return err == nil
}
func (c *Cache) Status() (Status, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var value Status
	entries, err := os.ReadDir(c.Root)
	if os.IsNotExist(err) {
		return value, nil
	}
	if err != nil {
		return value, err
	}
	for _, e := range entries {
		if !isCacheFile(e) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return value, err
		}
		value.Count++
		value.Bytes += info.Size()
	}
	return value, nil
}
func (c *Cache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := os.Stat(c.Root); os.IsNotExist(err) {
		return nil
	}
	if !c.trim("", 0) {
		return fmt.Errorf("清理封面预览缓存失败，请检查目录权限")
	}
	return nil
}
