//go:build importstress

package processors

import (
	"archive/zip"
	"bytes"
	"cm_collectors_server/config"
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/internal/importbatch"
	"cm_collectors_server/models"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

// Opt-in, isolated integration stress test. Never starts the HTTP server.
// Run alone: go test -tags=importstress -vet=off ./processors -run '^TestImportStressLocalPosters$' -v -count=1 -timeout=15m
func TestImportStressLocalPosters(t *testing.T) {
	root := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)
	// The application logger owns its file for the lifetime of the process.
	// Use the null device so Windows can remove the isolated directory on exit.
	if err = os.WriteFile("config.yaml", []byte("sqlite3:\n  loggerLevel: silent\nsystem:\n  logFilePath: NUL\n"), 0600); err != nil {
		t.Fatal(err)
	}
	core.Init()
	if err = models.DB_Init(core.DBS()); err != nil {
		t.Fatal(err)
	}
	db, _ := core.DBS().DB()
	defer db.Close()
	setting, err := (models.VideoMetadataSetting{}).Ensure(core.DBS())
	if err != nil {
		t.Fatal(err)
	}
	if err = core.DBS().Model(setting).Update("paused", true).Error; err != nil {
		t.Fatal(err)
	}
	core.Config.AutoBackup = config.AutoBackup{Enabled: true, ResourceChangeThreshold: 100, BackupPath: filepath.Join(root, "backups"), MaxBackups: 5}
	defer func() {
		autoBackupQueue.Lock()
		defer autoBackupQueue.Unlock()
		if autoBackupQueue.timer != nil {
			autoBackupQueue.timer.Stop()
		}
	}()
	img := image.NewYCbCr(image.Rect(0, 0, 6000, 4000), image.YCbCrSubsampleRatio420)
	for i := range img.Y {
		img.Y[i] = byte((i/6000 + i%6000) / 17)
	}
	for i := range img.Cb {
		img.Cb[i] = 110
		img.Cr[i] = 145
	}
	var encoded bytes.Buffer
	if err = jpeg.Encode(&encoded, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	img = nil
	t.Logf("fixture: 6000x4000 JPEG, %d bytes; each batch has 320 distinct video paths + local JPEGs", encoded.Len())
	for _, mode := range []struct {
		name  string
		cover int
	}{{"original", -1}, {"crop", 0}} {
		t.Run(mode.name, func(t *testing.T) {
			library := "stress-" + mode.name
			if err := core.DBS().Create(&models.FilesBases{ID: library, Name: library}).Error; err != nil {
				t.Fatal(err)
			}
			paths := make([]string, 320)
			for i := range paths {
				dir := filepath.Join(root, "fixtures", mode.name, fmt.Sprintf("%03d", i))
				if err := os.MkdirAll(dir, 0700); err != nil {
					t.Fatal(err)
				}
				paths[i] = filepath.Join(dir, "movie.mp4")
				if err := os.WriteFile(paths[i], []byte("placeholder: video probing disabled"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "movie.jpg"), encoded.Bytes(), 0600); err != nil {
					t.Fatal(err)
				}
			}
			cfg := datatype.Config_ScanDisk{ImportMode: datatype.ImportMode_append, CoverPosterSuffixName: []string{"jpg"}, CoverPosterType: mode.cover, CoverPosterWidth: 400, CoverPosterHeight: 600, AutoCreatePoster: true}
			runtime.GC()
			var before runtime.MemStats
			runtime.ReadMemStats(&before)
			stop := make(chan struct{})
			var wg sync.WaitGroup
			wg.Add(1)
			var peakHeap, peakWS, peakPrivate uint64
			go func() {
				defer wg.Done()
				tick := time.NewTicker(10 * time.Millisecond)
				defer tick.Stop()
				for {
					select {
					case <-stop:
						return
					case <-tick.C:
						var m runtime.MemStats
						runtime.ReadMemStats(&m)
						if m.HeapAlloc > peakHeap {
							peakHeap = m.HeapAlloc
						}
						ws, p := stressProcessMemory()
						if ws > peakWS {
							peakWS = ws
						}
						if p > peakPrivate {
							peakPrivate = p
						}
					}
				}
			}()
			defer func() { close(stop); wg.Wait() }()
			importbatch.Global.Begin(library, library)
			defer importbatch.Global.End(library)
			started := time.Now()
			for i, p := range paths {
				if !importbatch.Global.Touch(library, library) {
					t.Fatal("lease expired")
				}
				warnings, err := (ImportData{}).ScanDiskImportDataResult(context.Background(), library, p, cfg)
				if err != nil || len(warnings) != 0 {
					t.Fatalf("row %d: %v %v", i, err, warnings)
				}
				if (i+1)%80 == 0 {
					var n int64
					core.DBS().Model(&models.Resources{}).Where("filesBases_id = ?", library).Count(&n)
					if n != int64(i+1) {
						t.Fatalf("progress %d but persisted %d", i+1, n)
					}
					if err := (AutoBackup{}).run("resourceChange"); err != nil {
						t.Fatal(err)
					}
					state, err := (AutoBackup{}).ensureState()
					if err != nil {
						t.Fatal(err)
					}
					if state.Running || state.PendingResourceChangeCount != i+1 {
						t.Fatalf("backup ran during import: %+v", state)
					}
					var m runtime.MemStats
					runtime.ReadMemStats(&m)
					t.Logf("progress %d/320, persisted=%d, pendingBackup=%d, heap=%.1f MiB, elapsed=%s", i+1, n, state.PendingResourceChangeCount, float64(m.HeapAlloc)/(1<<20), time.Since(started).Round(time.Millisecond))
				}
			}
			var resources []models.Resources
			if err := core.DBS().Where("filesBases_id = ?", library).Find(&resources).Error; err != nil {
				t.Fatal(err)
			}
			for _, r := range resources {
				f, err := os.Open(filepath.Join(core.Config.System.FilePath, "resCoverPoster", library, r.CoverPoster))
				if err != nil {
					t.Fatal(err)
				}
				decoded, err := jpeg.DecodeConfig(f)
				f.Close()
				if err != nil {
					t.Fatal(err)
				}
				wantW, wantH := 6000, 4000
				if mode.cover >= 0 {
					wantW, wantH = 400, 600
				}
				if decoded.Width != wantW || decoded.Height != wantH {
					t.Fatal("incorrect saved poster dimensions", decoded)
				}
			}
			runtime.GC()
			var after runtime.MemStats
			runtime.ReadMemStats(&after)
			t.Logf("import completed in %s; heap after GC %.1f -> %.1f MiB; total allocated %.1f MiB", time.Since(started).Round(time.Millisecond), float64(before.HeapAlloc)/(1<<20), float64(after.HeapAlloc)/(1<<20), float64(after.TotalAlloc-before.TotalAlloc)/(1<<20))
			importbatch.Global.End(library)
			if err := (AutoBackup{}).run("resourceChange"); err != nil {
				t.Fatal(err)
			}
			state, _ := (AutoBackup{}).ensureState()
			if state.PendingResourceChangeCount != 320 {
				t.Fatal("quiet window bypassed")
			}
			backupWaitStarted := time.Now()
			deadline := time.Now().Add(5 * time.Minute)
			for time.Now().Before(deadline) {
				state, err = (AutoBackup{}).ensureState()
				if err != nil {
					t.Fatal(err)
				}
				if state.PendingResourceChangeCount == 0 && !state.Running && state.LastBackupPath != "" {
					break
				}
				time.Sleep(time.Second)
			}
			if state.PendingResourceChangeCount != 0 || state.Running || state.LastError != "" {
				t.Fatalf("automatic backup did not recover: %+v", state)
			}
			t.Logf("automatic backup completed %s after batch ended", time.Since(backupWaitStarted).Round(time.Second))
			z, err := zip.OpenReader(state.LastBackupPath)
			if err != nil {
				t.Fatal(err)
			}
			posters := 0
			for _, f := range z.File {
				if filepath.Ext(f.Name) == ".jpg" {
					posters++
				}
			}
			z.Close()
			wantPosters := 320
			if mode.name == "crop" {
				wantPosters = 640
			}
			if posters != wantPosters {
				t.Fatalf("backup poster count %d != %d", posters, wantPosters)
			}
			// Stop sampling before reading the accumulated peaks.
			stop <- struct{}{}
			wg.Wait()
			t.Logf("peak sampled heap=%.1f MiB, process working set=%.1f MiB, private bytes=%.1f MiB (includes backup); automatic backup recovered, ZIP posters=%d", float64(peakHeap)/(1<<20), float64(peakWS)/(1<<20), float64(peakPrivate)/(1<<20), posters)
		})
	}
}

func stressProcessMemory() (uint64, uint64) {
	type counters struct {
		CB, Faults                                                                            uint32
		PeakWS, WS, PeakPaged, Paged, PeakNonPaged, NonPaged, Pagefile, PeakPagefile, Private uintptr
	}
	var c counters
	c.CB = uint32(unsafe.Sizeof(c))
	proc := syscall.NewLazyDLL("psapi.dll").NewProc("GetProcessMemoryInfo")
	ok, _, _ := proc.Call(^uintptr(0), uintptr(unsafe.Pointer(&c)), uintptr(c.CB))
	if ok == 0 {
		return 0, 0
	}
	return uint64(c.WS), uint64(c.Private)
}
