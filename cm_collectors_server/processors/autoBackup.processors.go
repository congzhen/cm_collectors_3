package processors

import (
	"archive/zip"
	"cm_collectors_server/config"
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

const (
	autoBackupTickerInterval = 10 * time.Minute
	autoBackupQuietWindow    = 30 * time.Second
	autoBackupRunningTTL     = 30 * time.Minute
	autoBackupCopyBufferSize = 64 * 1024
	autoBackupCopyPause      = 10 * time.Millisecond
)

var autoBackupQueue = struct {
	sync.Mutex
	timer  *time.Timer
	queued bool
	reason string
}{}

var autoBackupRunMu sync.Mutex

type AutoBackup struct{}

type AutoBackupStateData struct {
	State  *models.AutoBackupState `json:"state"`
	Config config.AutoBackup       `json:"config"`
}

type slowWriter struct {
	w io.Writer
}

func InitAutoBackup() {
	_, err := AutoBackup{}.ensureState()
	if err != nil {
		core.LogErr(err)
		return
	}
	AutoBackup{}.CheckTimeTrigger("startup")
	go AutoBackup{}.ticker()
}

func (AutoBackup) ticker() {
	ticker := time.NewTicker(autoBackupTickerInterval)
	defer ticker.Stop()
	for range ticker.C {
		AutoBackup{}.CheckTimeTrigger("time")
	}
}

func (t AutoBackup) StateData() (*AutoBackupStateData, error) {
	state, err := t.ensureState()
	if err != nil {
		return nil, err
	}
	return &AutoBackupStateData{
		State:  state,
		Config: t.config(),
	}, nil
}

func (t AutoBackup) AutoBackupList() ([]string, error) {
	savePath, err := t.backupSavePath()
	if err != nil {
		return nil, err
	}
	return t.zipFiles(savePath)
}

func (t AutoBackup) DeleteAutoBackup(fileName string) error {
	savePath, err := t.backupSavePath()
	if err != nil {
		return err
	}
	fileName = filepath.Base(filepath.Clean(fileName))
	if filepath.Ext(fileName) != ".zip" {
		fileName += ".zip"
	}
	return os.Remove(filepath.Join(savePath, fileName))
}

func (t AutoBackup) ManualRun() error {
	return t.run("manual")
}

func (t AutoBackup) CheckTimeTrigger(reason string) {
	cfg := t.config()
	if !cfg.Enabled || cfg.IntervalHours <= 0 {
		return
	}
	state, err := t.ensureState()
	if err != nil {
		core.LogErr(err)
		return
	}
	now := datatype.CustomTime(core.TimeNow())
	core.DBS().Model(&models.AutoBackupState{}).
		Where("id = ?", models.AutoBackupStateID).
		Updates(map[string]any{
			"last_time_check_at": &now,
			"updated_at":         &now,
		})
	if state.LastSuccessBackupAt == nil || state.LastSuccessBackupAt.IsZero() {
		t.QueueRun(reason)
		return
	}
	last := time.Time(*state.LastSuccessBackupAt)
	if time.Since(last) >= time.Duration(cfg.IntervalHours)*time.Hour {
		t.QueueRun(reason)
	}
}

func (t AutoBackup) RecordResourceChanges(count int) {
	if count <= 0 {
		return
	}
	cfg := t.config()
	if !cfg.Enabled {
		return
	}
	if _, err := t.ensureState(); err != nil {
		core.LogErr(err)
		return
	}
	now := datatype.CustomTime(core.TimeNow())
	db := core.DBS()
	if err := db.Model(&models.AutoBackupState{}).
		Where("id = ?", models.AutoBackupStateID).
		Updates(map[string]any{
			"pending_resource_change_count": gorm.Expr("pending_resource_change_count + ?", count),
			"last_resource_change_at":       &now,
			"updated_at":                    &now,
		}).Error; err != nil {
		core.LogErr(err)
		return
	}
	if cfg.ResourceChangeThreshold <= 0 {
		return
	}
	state, err := t.ensureState()
	if err != nil {
		core.LogErr(err)
		return
	}
	if state.PendingResourceChangeCount >= cfg.ResourceChangeThreshold {
		t.QueueRun("resourceChange")
	}
}

func (t AutoBackup) QueueRun(reason string) {
	if reason == "" {
		reason = "auto"
	}
	autoBackupQueue.Lock()
	defer autoBackupQueue.Unlock()
	autoBackupQueue.reason = mergeBackupReason(autoBackupQueue.reason, reason)
	if autoBackupQueue.queued {
		return
	}
	autoBackupQueue.queued = true
	autoBackupQueue.timer = time.AfterFunc(autoBackupQuietWindow, func() {
		autoBackupQueue.Lock()
		runReason := autoBackupQueue.reason
		autoBackupQueue.queued = false
		autoBackupQueue.reason = ""
		autoBackupQueue.timer = nil
		autoBackupQueue.Unlock()
		if err := t.run(runReason); err != nil {
			core.LogErr(err)
		}
	})
}

func (t AutoBackup) run(reason string) error {
	autoBackupRunMu.Lock()
	defer autoBackupRunMu.Unlock()

	cfg := t.config()
	if reason != "manual" && !cfg.Enabled {
		return nil
	}
	if err := t.tryMarkRunning(); err != nil {
		return err
	}
	zipPath, err := t.createBackupZip()
	if err != nil {
		t.markFailed(err)
		return err
	}
	if err := t.cleanupOldBackups(); err != nil {
		core.LogErr(err)
	}
	t.markSuccess(zipPath, reason)
	return nil
}

func (t AutoBackup) config() config.AutoBackup {
	cfg := core.Config.AutoBackup
	if cfg.BackupPath == "" {
		cfg.BackupPath = "./auto_backup"
	}
	if cfg.MaxBackups <= 0 {
		cfg.MaxBackups = 5
	}
	return cfg
}

func (AutoBackup) ensureState() (*models.AutoBackupState, error) {
	db := core.DBS()
	if err := db.AutoMigrate(&models.AutoBackupState{}); err != nil {
		return nil, err
	}
	return models.AutoBackupState{}.Ensure(db)
}

func (t AutoBackup) tryMarkRunning() error {
	state, err := t.ensureState()
	if err != nil {
		return err
	}
	if state.Running && state.RunningStartedAt != nil && !state.RunningStartedAt.IsZero() {
		started := time.Time(*state.RunningStartedAt)
		if time.Since(started) < autoBackupRunningTTL {
			return errors.New("auto backup is already running")
		}
	}
	now := datatype.CustomTime(core.TimeNow())
	return core.DBS().Model(&models.AutoBackupState{}).
		Where("id = ?", models.AutoBackupStateID).
		Updates(map[string]any{
			"running":            true,
			"running_started_at": &now,
			"updated_at":         &now,
		}).Error
}

func (t AutoBackup) markSuccess(zipPath, reason string) {
	now := datatype.CustomTime(core.TimeNow())
	err := core.DBS().Model(&models.AutoBackupState{}).
		Where("id = ?", models.AutoBackupStateID).
		Updates(map[string]any{
			"last_success_backup_at":        &now,
			"pending_resource_change_count": 0,
			"running":                       false,
			"running_started_at":            nil,
			"last_backup_path":              filepath.ToSlash(zipPath),
			"last_backup_reason":            reason,
			"last_error":                    "",
			"updated_at":                    &now,
		}).Error
	if err != nil {
		core.LogErr(err)
	}
}

func (t AutoBackup) markFailed(runErr error) {
	now := datatype.CustomTime(core.TimeNow())
	err := core.DBS().Model(&models.AutoBackupState{}).
		Where("id = ?", models.AutoBackupStateID).
		Updates(map[string]any{
			"running":            false,
			"running_started_at": nil,
			"last_error":         runErr.Error(),
			"updated_at":         &now,
		}).Error
	if err != nil {
		core.LogErr(err)
	}
}

func (t AutoBackup) createBackupZip() (string, error) {
	sourcePath, err := filepath.Abs(core.Config.System.FilePath)
	if err != nil {
		return "", err
	}
	savePath, err := t.backupSavePath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(savePath, 0755); err != nil {
		return "", err
	}
	zipPath := t.nextBackupZipPath(savePath)
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", err
	}
	zipWriter := zip.NewWriter(zipFile)

	var sqliteSnapshot string
	if core.Config.System.Database == "sqlite3" {
		sqliteSnapshot, err = t.createSqliteSnapshot(savePath)
		if err != nil {
			zipWriter.Close()
			zipFile.Close()
			os.Remove(zipPath)
			return "", err
		}
		defer os.Remove(sqliteSnapshot)
	}

	if err := t.zipDirectory(zipWriter, sourcePath, savePath, sqliteSnapshot); err != nil {
		zipWriter.Close()
		zipFile.Close()
		os.Remove(zipPath)
		return "", err
	}
	if err := zipWriter.Close(); err != nil {
		zipFile.Close()
		os.Remove(zipPath)
		return "", err
	}
	if err := zipFile.Close(); err != nil {
		os.Remove(zipPath)
		return "", err
	}
	return zipPath, nil
}

func (t AutoBackup) nextBackupZipPath(savePath string) string {
	baseName := time.Now().Format("20060102_150405")
	zipPath := filepath.Join(savePath, baseName+".zip")
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		return zipPath
	}
	for i := 1; ; i++ {
		candidate := filepath.Join(savePath, fmt.Sprintf("%s_%02d.zip", baseName, i))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

func (t AutoBackup) createSqliteSnapshot(savePath string) (string, error) {
	tempFile, err := os.CreateTemp(savePath, ".cm_auto_backup_*.db")
	if err != nil {
		return "", err
	}
	snapshotPath := tempFile.Name()
	tempFile.Close()
	os.Remove(snapshotPath)
	if err := core.DBS().Exec("VACUUM INTO ?", snapshotPath).Error; err != nil {
		return "", err
	}
	return snapshotPath, nil
}

func (t AutoBackup) zipDirectory(zipWriter *zip.Writer, sourcePath string, savePath string, sqliteSnapshot string) error {
	sqlitePath, _ := filepath.Abs(core.Config.Sqlite3.Path)
	sqliteWalPath := sqlitePath + "-wal"
	sqliteShmPath := sqlitePath + "-shm"
	sqliteAdded := false
	savePath = filepath.Clean(savePath)

	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == sourcePath {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if isPathInside(absPath, savePath) {
			return nil
		}
		if sqliteSnapshot != "" && sameCleanPath(absPath, sqlitePath) {
			relPath, err := filepath.Rel(sourcePath, absPath)
			if err != nil {
				return err
			}
			sqliteAdded = true
			return t.addFileToZip(zipWriter, sqliteSnapshot, relPath)
		}
		if sqliteSnapshot != "" && (sameCleanPath(absPath, sqliteWalPath) || sameCleanPath(absPath, sqliteShmPath)) {
			return nil
		}
		relPath, err := filepath.Rel(sourcePath, absPath)
		if err != nil {
			return err
		}
		return t.addFileToZip(zipWriter, absPath, relPath)
	})
	if err != nil {
		return err
	}
	if sqliteSnapshot != "" && !sqliteAdded {
		return fmt.Errorf("sqlite database file not found in backup source")
	}
	return nil
}

func (t AutoBackup) addFileToZip(zipWriter *zip.Writer, filePath string, relPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer, err := zipWriter.Create(filepath.ToSlash(relPath))
	if err != nil {
		return err
	}
	_, err = io.CopyBuffer(slowWriter{w: writer}, file, make([]byte, autoBackupCopyBufferSize))
	time.Sleep(autoBackupCopyPause)
	return err
}

func (w slowWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	if len(p) >= autoBackupCopyBufferSize/2 {
		time.Sleep(autoBackupCopyPause)
	}
	return n, err
}

func (t AutoBackup) backupSavePath() (string, error) {
	savePath := t.config().BackupPath
	if savePath == "" {
		savePath = "./auto_backup"
	}
	return filepath.Abs(savePath)
}

func (t AutoBackup) cleanupOldBackups() error {
	cfg := t.config()
	if cfg.MaxBackups <= 0 {
		return nil
	}
	savePath, err := t.backupSavePath()
	if err != nil {
		return err
	}
	files, err := t.zipFiles(savePath)
	if err != nil {
		return err
	}
	if len(files) <= cfg.MaxBackups {
		return nil
	}
	sort.SliceStable(files, func(i, j int) bool {
		iInfo, iErr := os.Stat(files[i])
		jInfo, jErr := os.Stat(files[j])
		if iErr != nil || jErr != nil {
			return files[i] < files[j]
		}
		return iInfo.ModTime().After(jInfo.ModTime())
	})
	for _, file := range files[cfg.MaxBackups:] {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			core.LogErr(err)
		}
	}
	return nil
}

func (AutoBackup) zipFiles(savePath string) ([]string, error) {
	entries, err := os.ReadDir(savePath)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	files := []string{}
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".zip" {
			continue
		}
		files = append(files, filepath.ToSlash(filepath.Join(savePath, entry.Name())))
	}
	sort.SliceStable(files, func(i, j int) bool {
		iInfo, iErr := os.Stat(files[i])
		jInfo, jErr := os.Stat(files[j])
		if iErr != nil || jErr != nil {
			return files[i] > files[j]
		}
		return iInfo.ModTime().After(jInfo.ModTime())
	})
	return files, nil
}

func sameCleanPath(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}

func isPathInside(path string, dir string) bool {
	path = filepath.Clean(path)
	dir = filepath.Clean(dir)
	if path == dir {
		return true
	}
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	return rel != "." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".."
}

func mergeBackupReason(existing, next string) string {
	if existing == "" || existing == next {
		return next
	}
	if existing == "resourceChange" || next == "resourceChange" {
		if strings.Contains(existing, "time") || next == "time" || next == "startup" || next == "config" {
			return "time+resourceChange"
		}
		return "resourceChange"
	}
	if existing == "time+resourceChange" || next == "time+resourceChange" {
		return "time+resourceChange"
	}
	if existing == "manual" || next == "manual" {
		return "manual"
	}
	return next
}
