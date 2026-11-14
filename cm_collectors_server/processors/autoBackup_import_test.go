package processors

import (
	"cm_collectors_server/config"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestImportChangesSurvivePreviouslyRunningBackup(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	sqlDB.SetMaxOpenConns(1)
	if err = db.AutoMigrate(&models.AutoBackupState{}); err != nil {
		t.Fatal(err)
	}
	state := models.AutoBackupState{ID: models.AutoBackupStateID, PendingResourceChangeCount: 5, Running: true}
	if err = db.Create(&state).Error; err != nil {
		t.Fatal(err)
	}
	// Three new resources arrive while the old backup is running.
	if err = db.Model(&state).Update("pending_resource_change_count", 8).Error; err != nil {
		t.Fatal(err)
	}
	if err = saveAutoBackupSuccess(db, "backup.zip", "resourceChange", 5); err != nil {
		t.Fatal(err)
	}
	if err = db.First(&state).Error; err != nil {
		t.Fatal(err)
	}
	if state.PendingResourceChangeCount != 3 || state.Running {
		t.Fatalf("new import changes lost: %+v", state)
	}
	if err = saveAutoBackupSuccess(db, "next.zip", "manual", 10); err != nil {
		t.Fatal(err)
	}
	if err = db.First(&state).Error; err != nil {
		t.Fatal(err)
	}
	if state.PendingResourceChangeCount != 0 {
		t.Fatal("counter became negative")
	}
}

func TestImportDeferredBackupRechecksConditions(t *testing.T) {
	now := time.Now()
	recent := datatype.CustomTime(now.Add(-time.Minute))
	old := datatype.CustomTime(now.Add(-2 * time.Hour))
	cfg := config.AutoBackup{Enabled: true, IntervalHours: 1, ResourceChangeThreshold: 100}
	state := models.AutoBackupState{LastSuccessBackupAt: &recent, PendingResourceChangeCount: 0}
	if autoBackupDue(cfg, &state, now) {
		t.Fatal("stale queued trigger caused duplicate backup after manual backup")
	}
	state.PendingResourceChangeCount = 100
	if !autoBackupDue(cfg, &state, now) {
		t.Fatal("new import changes ignored")
	}
	state.PendingResourceChangeCount = 0
	state.LastSuccessBackupAt = &old
	if !autoBackupDue(cfg, &state, now) {
		t.Fatal("time trigger ignored")
	}
	cfg.Enabled = false
	if autoBackupDue(cfg, &state, now) {
		t.Fatal("disabled backup admitted")
	}
	cfg.Enabled = true
	cfg.IntervalHours = 0
	cfg.ResourceChangeThreshold = 0
	if autoBackupDue(cfg, &state, now) {
		t.Fatal("removed triggers still executed")
	}
}
