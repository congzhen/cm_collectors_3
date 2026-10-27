package selfupdate

import (
	"archive/zip"
	"crypto/sha256"
	"debug/elf"
	"debug/pe"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

const maxArchiveSize int64 = 2 << 30
const maxExpandedSize int64 = 6 << 30

func serverName() string {
	if runtime.GOOS == "windows" {
		return "cm_collectors_server.exe"
	}
	return "cm_collectors_server"
}
func writeJSON(name string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	f, e := os.OpenFile(name+".tmp", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	_, e = f.Write(b)
	if e == nil {
		e = f.Sync()
	}
	ce := f.Close()
	if e != nil {
		return e
	}
	if ce != nil {
		return ce
	}
	return replaceFile(name+".tmp", name)
}
func readJSON(name string, v any) error {
	b, e := os.ReadFile(name)
	if e != nil {
		return e
	}
	return json.Unmarshal(b, v)
}
func copyFile(src, dst string, mode os.FileMode) error {
	if e := os.MkdirAll(filepath.Dir(dst), 0755); e != nil {
		return e
	}
	in, e := os.Open(src)
	if e != nil {
		return e
	}
	defer in.Close()
	out, e := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if e != nil {
		return e
	}
	_, e = io.Copy(out, in)
	if e == nil {
		e = out.Sync()
	}
	ce := out.Close()
	if e != nil {
		return e
	}
	return ce
}
func within(root, name string) bool {
	rel, e := filepath.Rel(root, name)
	return e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && !filepath.IsAbs(rel)
}
func safeName(name string) bool {
	if name == "" || strings.ContainsAny(name, "\\:\x00") || path.IsAbs(name) || path.Clean(name) != strings.TrimSuffix(name, "/") {
		return false
	}
	for _, p := range strings.Split(strings.TrimSuffix(name, "/"), "/") {
		if p == ".." || p == "." || strings.TrimRight(p, ". ") != p {
			return false
		}
		base := strings.ToUpper(strings.Split(p, ".")[0])
		if base == "CON" || base == "NUL" || base == "AUX" || base == "PRN" || (len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '0' && base[3] <= '9') {
			return false
		}
	}
	return true
}
func protected(rel string, target string, extra []string) bool {
	p := strings.ToLower(filepath.ToSlash(rel))
	first := strings.Split(p, "/")[0]
	base := path.Base(p)
	if base == "config.json" || base == "config.yaml" || base == "config.yml" {
		if _, e := os.Stat(target); e == nil {
			return true
		}
	}
	if first == ".cm-updates" || first == "db" || first == "cert" || first == "logs" || first == "log" || first == "auto_backup" || first == "backup" || first == "db_backup" || first == "runtime" || first == "config.yaml" || first == "config.json" {
		return true
	}
	if first == "scraper" {
		if _, e := os.Stat(target); e == nil {
			return true
		}
	}
	for _, v := range extra {
		if within(v, target) {
			return true
		}
	}
	return false
}

// No symlinks/reparse points may occur along the destination path.
func safeDestination(root, rel string) (string, error) {
	target := filepath.Join(root, rel)
	if !within(root, target) {
		return "", fmt.Errorf("更新路径越界")
	}
	for p := target; ; p = filepath.Dir(p) {
		st, e := os.Lstat(p)
		if e != nil && !os.IsNotExist(e) {
			return "", e
		}
		if e == nil && st.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("更新路径包含符号链接: %s", p)
		}
		if p == root {
			break
		}
	}
	return target, nil
}
func verifyArchive(file string, a Asset) error {
	f, e := os.Open(file)
	if e != nil {
		return e
	}
	defer f.Close()
	h := sha256.New()
	n, e := io.Copy(h, io.LimitReader(f, maxArchiveSize+1))
	if e != nil {
		return e
	}
	if n != a.Size || !strings.EqualFold(hex.EncodeToString(h.Sum(nil)), a.SHA256) {
		return fmt.Errorf("更新包大小或 SHA-256 校验失败")
	}
	return nil
}

type archiveEntry struct {
	rel  string
	file *zip.File
}

// Inspect paths once; preflight can use sizes without extracting another copy.
func inspectArchive(r *zip.ReadCloser) ([]archiveEntry, uint64, error) {
	if len(r.File) > 20000 {
		return nil, 0, fmt.Errorf("更新包文件过多")
	}
	root := ""
	found := 0
	var total uint64
	for _, f := range r.File {
		if !safeName(f.Name) || f.Mode()&os.ModeSymlink != 0 || (!f.FileInfo().IsDir() && !f.Mode().IsRegular()) {
			return nil, 0, fmt.Errorf("更新包包含不安全路径: %s", f.Name)
		}
		if f.UncompressedSize64 > uint64(maxExpandedSize)-total {
			return nil, 0, fmt.Errorf("解压体积超过限制")
		}
		total += f.UncompressedSize64
		if path.Base(f.Name) == serverName() && !f.FileInfo().IsDir() {
			root = strings.TrimSuffix(f.Name, serverName())
			found++
		}
	}
	if found != 1 {
		return nil, 0, fmt.Errorf("更新包必须包含一个 %s", serverName())
	}
	seen := map[string]bool{}
	entries := []archiveEntry{}
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if !strings.HasPrefix(f.Name, root) {
			return nil, 0, fmt.Errorf("更新包文件不在统一程序目录内")
		}
		rel := strings.TrimPrefix(f.Name, root)
		key := strings.ToLower(rel)
		if seen[key] {
			return nil, 0, fmt.Errorf("更新包包含重复路径")
		}
		seen[key] = true
		entries = append(entries, archiveEntry{filepath.FromSlash(rel), f})
	}
	return entries, total, nil
}
func extract(file, stage string) ([]string, error) {
	r, e := zip.OpenReader(file)
	if e != nil {
		return nil, e
	}
	defer r.Close()
	entries, total, e := inspectArchive(r)
	if e != nil {
		return nil, e
	}
	if e := os.MkdirAll(stage, 0700); e != nil {
		return nil, e
	}
	if e := requireSpace(stage, total+(64<<20)); e != nil {
		return nil, e
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		rel, f := entry.rel, entry.file
		dst := filepath.Join(stage, rel)
		if !within(stage, dst) {
			return nil, fmt.Errorf("更新路径越界")
		}
		if e := os.MkdirAll(filepath.Dir(dst), 0755); e != nil {
			return nil, e
		}
		in, e := f.Open()
		if e != nil {
			return nil, e
		}
		mode := os.FileMode(0644)
		if f.Mode()&0111 != 0 || rel == serverName() {
			mode = 0755
		}
		out, e := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
		if e != nil {
			in.Close()
			return nil, e
		}
		_, e = io.Copy(out, io.LimitReader(in, int64(f.UncompressedSize64)+1))
		ce := out.Close()
		ie := in.Close()
		if e != nil {
			return nil, e
		}
		if ce != nil {
			return nil, ce
		}
		if ie != nil {
			return nil, ie
		}
		files = append(files, rel)
	}
	if e := validateBinary(filepath.Join(stage, serverName())); e != nil {
		return nil, e
	}
	return files, nil
}

func validateBinary(name string) error {
	if runtime.GOOS == "windows" {
		f, e := pe.Open(name)
		if e != nil {
			return e
		}
		defer f.Close()
		if f.Machine != pe.IMAGE_FILE_MACHINE_AMD64 {
			return fmt.Errorf("更新包不是 Windows amd64 程序")
		}
	} else {
		f, e := elf.Open(name)
		if e != nil {
			return e
		}
		defer f.Close()
		if f.Machine != elf.EM_X86_64 {
			return fmt.Errorf("更新包不是 Linux amd64 程序")
		}
		// Both Alpine and Ubuntu images must use a static Go server.
		if os.Getenv("CM_CONTAINER") == "1" {
			for _, p := range f.Progs {
				if p.Type == elf.PT_INTERP {
					return fmt.Errorf("Docker 内更新需要静态链接的 Linux 程序，请使用 CGO_ENABLED=0 构建")
				}
			}
		}
	}
	return nil
}
