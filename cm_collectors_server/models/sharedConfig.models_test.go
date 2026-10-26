package models

import (
	"cm_collectors_server/datatype"
	"encoding/json"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func sharedTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { sqlDB.Close() })
	if err := db.AutoMigrate(&FilesBases{}, &FilesBasesSetting{}, &SharedLibraryConfig{}, &LibraryConfigFollow{}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"A", "B", "C"} {
		if err := db.Create(&FilesBases{ID: id, Name: id, Status: true}).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Create(&FilesBasesSetting{FilesBasesID: id, ConfigJsonData: `{"pageLimit":12,"coverDisplayTag":["local-tag"],"sampleFolder":"local-folder","coverPosterData":[{"name":"local"}]}`}).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func TestSharedConfigUsesProductionTableNames(t *testing.T) {
	db := sharedTestDB(t)
	for _, name := range []string{"shared_library_config", "library_config_follow"} {
		if !db.Migrator().HasTable(name) {
			t.Fatalf("missing production table %s", name)
		}
	}
	for _, name := range []string{"shared_library_configs", "library_config_follows"} {
		if db.Migrator().HasTable(name) {
			t.Fatalf("unexpected plural table %s", name)
		}
	}
	// 即使没有公共配置或跟随库，设置页也必须能正常查询状态。
	state, err := SharedConfigStatus(db, "A", "display")
	if err != nil {
		t.Fatal(err)
	}
	if state.Available || state.Following || len(state.Libraries) != 0 {
		t.Fatalf("unexpected initial state: %#v", state)
	}
	if err := SaveSharedConfig(db, "display", 0, displayTestConfig(t, 32)); err != nil {
		t.Fatal(err)
	}
	if err := db.Transaction(func(tx *gorm.DB) error { return SetLibraryConfigFollow(tx, "A", "display", true, 1) }); err != nil {
		t.Fatal(err)
	}
	// 再次迁移保持已保存数据，不能通过换表名掩盖查询错误。
	if err := db.AutoMigrate(&SharedLibraryConfig{}, &LibraryConfigFollow{}); err != nil {
		t.Fatal(err)
	}
	state, err = SharedConfigStatus(db, "A", "display")
	if err != nil {
		t.Fatal(err)
	}
	if !state.Following || state.Revision != 1 || len(state.Libraries) != 1 || state.Libraries[0].ID != "A" {
		t.Fatalf("migration lost shared configuration: %#v", state)
	}
}

func displayTestConfig(t *testing.T, limit int) string {
	t.Helper()
	data, _ := json.Marshal(datatype.Config_FilesBases{})
	value, _ := parseConfig(string(data))
	value["casualViewModule"] = false
	value["casualViewNumber"] = 10
	value["pageLimit"] = limit
	value["videoPreviewGeneration"] = "playback"
	value["videoPreviewInterval"] = 30
	// 空数组是有效设置，不能作为空值拒绝。
	for _, key := range SharedConfigFields("display") {
		if value[key] == nil {
			value[key] = []string{}
		}
	}
	value["coverDisplayTag"] = []string{"must-not-copy"}
	value["sampleFolder"] = "must-not-copy"
	data, _ = json.Marshal(value)
	return string(data)
}
func TestSharedPreviewFieldsFollowRestoreAndValidate(t *testing.T) {
	db := sharedTestDB(t)
	local := `{"videoPreviewGeneration":"manual","videoPreviewInterval":120,"sampleFolder":"keep"}`
	if err := db.Model(&FilesBasesSetting{}).Where("filesBases_id = ?", "A").Update("config_json_data", local).Error; err != nil {
		t.Fatal(err)
	}
	if err := SaveSharedConfig(db, "display", 0, displayTestConfig(t, 32)); err != nil {
		t.Fatal(err)
	}
	if err := db.Transaction(func(tx *gorm.DB) error { return SetLibraryConfigFollow(tx, "A", "display", true, 1) }); err != nil {
		t.Fatal(err)
	}
	effective, err := EffectiveLibraryConfig(db, "A", "display", local)
	if err != nil {
		t.Fatal(err)
	}
	config, err := parseConfig(effective)
	if err != nil || config["videoPreviewGeneration"] != "playback" || config["videoPreviewInterval"] != float64(30) || config["sampleFolder"] != "keep" {
		t.Fatalf("unexpected effective config: %s, %v", effective, err)
	}
	if err := GuardSharedConfigWrite(db, "A", "display", local); err == nil {
		t.Fatal("following library changed shared preview fields")
	}
	if err := db.Transaction(func(tx *gorm.DB) error { return SetLibraryConfigFollow(tx, "A", "display", false, 1, "restore") }); err != nil {
		t.Fatal(err)
	}
	var saved FilesBasesSetting
	db.First(&saved, "filesBases_id = ?", "A")
	config, err = parseConfig(saved.ConfigJsonData)
	if err != nil || config["videoPreviewGeneration"] != "manual" || config["videoPreviewInterval"] != float64(120) {
		t.Fatalf("preview snapshot lost: %+v %v", config, err)
	}
	for key, value := range map[string]interface{}{"videoPreviewGeneration": "invalid", "videoPreviewInterval": 15} {
		invalid, _ := parseConfig(displayTestConfig(t, 32))
		invalid[key] = value
		raw, _ := json.Marshal(invalid)
		if err := SaveSharedConfig(db, "display", 1, string(raw)); err == nil {
			t.Fatalf("accepted invalid %s", key)
		}
	}
}

func TestSharedDisplayIsolationDetachAndRevision(t *testing.T) {
	db := sharedTestDB(t)
	if err := SaveSharedConfig(db, "display", 0, displayTestConfig(t, 32)); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"A", "B"} {
		if err := db.Transaction(func(tx *gorm.DB) error { return SetLibraryConfigFollow(tx, id, "display", true, 1) }); err != nil {
			t.Fatal(err)
		}
	}
	read := func(id string) map[string]interface{} {
		var setting FilesBasesSetting
		if err := db.Where("filesBases_id = ?", id).Take(&setting).Error; err != nil {
			t.Fatal(err)
		}
		raw, err := EffectiveLibraryConfig(db, id, "display", setting.ConfigJsonData)
		if err != nil {
			t.Fatal(err)
		}
		value, err := parseConfig(raw)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	if read("A")["pageLimit"] != float64(32) || read("C")["pageLimit"] != float64(12) {
		t.Fatal("library isolation failed")
	}
	if read("A")["sampleFolder"] != "local-folder" || read("A")["coverDisplayTag"].([]interface{})[0] != "local-tag" {
		t.Fatal("local references overwritten")
	}
	if err := SaveSharedConfig(db, "display", 1, displayTestConfig(t, 64)); err != nil {
		t.Fatal(err)
	}
	if err := SaveSharedConfig(db, "display", 1, displayTestConfig(t, 16)); err == nil {
		t.Fatal("stale revision accepted")
	}
	if read("B")["pageLimit"] != float64(64) {
		t.Fatal("new public config not effective")
	}
	if err := GuardSharedConfigWrite(db, "A", "display", displayTestConfig(t, 99)); err == nil {
		t.Fatal("ordinary save bypassed following")
	}
	if err := db.Transaction(func(tx *gorm.DB) error { return SetLibraryConfigFollow(tx, "B", "display", false, 2) }); err != nil {
		t.Fatal(err)
	}
	if err := SaveSharedConfig(db, "display", 2, displayTestConfig(t, 128)); err != nil {
		t.Fatal(err)
	}
	if read("B")["pageLimit"] != float64(64) || read("A")["pageLimit"] != float64(128) {
		t.Fatal("detach did not preserve effective values")
	}
	state, err := SharedConfigStatus(db, "B", "display")
	if err != nil {
		t.Fatal(err)
	}
	if state.Following || len(state.Libraries) != 1 || state.Libraries[0].ID != "A" {
		t.Fatalf("incorrect follower list: %#v", state)
	}
}

func TestSharedConfigLegacyAndFailedTransactions(t *testing.T) {
	db := sharedTestDB(t)
	raw, err := EffectiveLibraryConfig(db, "C", "display", "")
	if err != nil || raw != "" {
		t.Fatal("empty legacy config changed")
	}
	if err := SaveSharedConfig(db, "display", 0, `{"pageLimit":32}`); err == nil {
		t.Fatal("incomplete config accepted")
	}
	if err := SaveSharedConfig(db, "display", 0, displayTestConfig(t, 32)); err != nil {
		t.Fatal(err)
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := SetLibraryConfigFollow(tx, "A", "display", true, 1); err != nil {
			return err
		}
		return SetLibraryConfigFollow(tx, "A", "unknown", true, 1)
	})
	if err == nil {
		t.Fatal("invalid module accepted")
	}
	state, err := SharedConfigStatus(db, "A", "display")
	if err != nil || state.Following {
		t.Fatal("failed transaction left follow state")
	}
	if err := GuardSharedConfigWrite(db, "C", "display", `{"pageLimit":17}`); err != nil {
		t.Fatal("independent legacy save rejected", err)
	}
}

func importTestConfig() string {
	return `{"videoSuffixName":["mp4"],"autoGetVideoDefinition":true,"resourceNamingMode":"fileName","importMode":"append","coverPosterMatchName":["poster"],"coverPosterFuzzyMatch":true,"coverPosterUseRandomImageIfNoMatch":false,"coverPosterSuffixName":["jpg"],"autoCreatePoster":true,"folderToSeries":false,"similarNameToSeries":true,"folderToSeriesSortMode":"nameAsc","enableNfoFuzzyMatch":true,"useRandomNfoIfNoneMatch":false,"nfo":{"nfoStatus":true,"roots":["movie"],"titles":["title"],"issueNumbers":[],"issuingDates":[],"score":[],"abstracts":[],"tags":["genre"],"tagAutoCreate":false,"performerNames":[],"performerMatchAliasName":false,"performerAutoCreate":false,"performerThumbs":[]},"scanDiskPaths":["source-library"],"coverPosterType":5,"coverPosterWidth":999}`
}

func scraperTestConfig(timeout int) string {
	value := map[string]interface{}{"videoSuffixName": []string{"mp4"}, "scraperConfigs": []string{"example"}, "concurrency": 3, "retryCount": 3, "timeout": timeout, "skipIfNfoExists": true, "saveNfo": true, "enableDownloadImages": true, "useTagAsImageName": true, "enableUserSimulation": false, "scanDiskPaths": []string{"source-library"}}
	data, _ := json.Marshal(value)
	return string(data)
}

func TestSharedTaskModulesKeepPathsAndSnapshots(t *testing.T) {
	db := sharedTestDB(t)
	if err := db.Model(&FilesBasesSetting{}).Where("filesBases_id = ?", "A").Updates(map[string]interface{}{
		"scan_disk_json_data": `{"scanDiskPaths":["A-import"],"coverPosterType":2,"coverPosterWidth":400,"legacyLocal":true}`,
		"scraper_json_data":   `{"scanDiskPaths":["A-scrape"]}`,
	}).Error; err != nil {
		t.Fatal(err)
	}
	if err := SaveSharedConfig(db, "import", 0, importTestConfig()); err != nil {
		t.Fatal(err)
	}
	if err := SaveSharedConfig(db, "scraper", 0, scraperTestConfig(30)); err != nil {
		t.Fatal(err)
	}
	for _, module := range []string{"import", "scraper"} {
		if err := db.Transaction(func(tx *gorm.DB) error { return SetLibraryConfigFollow(tx, "A", module, true, 1) }); err != nil {
			t.Fatal(err)
		}
	}
	var local FilesBasesSetting
	db.Where("filesBases_id = ?", "A").Take(&local)
	importRaw, err := EffectiveLibraryConfig(db, "A", "import", local.ScanDiskJsonData)
	if err != nil {
		t.Fatal(err)
	}
	var importConfig datatype.Config_ScanDisk
	if err := json.Unmarshal([]byte(importRaw), &importConfig); err != nil {
		t.Fatal(err)
	}
	if importConfig.ScanDiskPaths[0] != "A-import" || importConfig.CoverPosterType != 2 || importConfig.CoverPosterWidth != 400 {
		t.Fatal("import copied another library's local fields")
	}
	encoded, _ := json.Marshal(importConfig)
	if err := GuardSharedConfigWrite(db, "A", "import", string(encoded)); err != nil {
		t.Fatal("normal import execution save rejected", err)
	}
	scraperRaw, err := EffectiveLibraryConfig(db, "A", "scraper", local.ScraperJsonData)
	if err != nil {
		t.Fatal(err)
	}
	var snapshot datatype.Config_Scraper
	if err := json.Unmarshal([]byte(scraperRaw), &snapshot); err != nil {
		t.Fatal(err)
	}
	encoded, _ = json.Marshal(snapshot)
	if err := GuardSharedConfigWrite(db, "A", "scraper", string(encoded)); err != nil {
		t.Fatal("scraper save lost concurrency", err)
	}
	if snapshot.ScanDiskPaths[0] != "A-scrape" || snapshot.Concurrency != 3 {
		t.Fatal("scraper local values changed")
	}
	if err := SaveSharedConfig(db, "scraper", 1, scraperTestConfig(90)); err != nil {
		t.Fatal(err)
	}
	if snapshot.Timeout != 30 {
		t.Fatal("running task snapshot changed")
	}
	if err := GuardSharedConfigWrite(db, "A", "scraper", string(encoded)); err == nil {
		t.Fatal("stale execution setup should request reload")
	}
	if err := db.Transaction(func(tx *gorm.DB) error { return SetLibraryConfigFollow(tx, "A", "scraper", false, 2) }); err != nil {
		t.Fatal(err)
	}
	db.Where("filesBases_id = ?", "A").Take(&local)
	var detached datatype.Config_Scraper
	json.Unmarshal([]byte(local.ScraperJsonData), &detached)
	if detached.Timeout != 90 || detached.ScanDiskPaths[0] != "A-scrape" {
		t.Fatal("detach did not retain task parameters and paths")
	}
	state, _ := SharedConfigStatus(db, "A", "import")
	if !state.Following {
		t.Fatal("detaching scraper also detached import")
	}
	state, _ = SharedConfigStatus(db, "A", "display")
	if state.Following {
		t.Fatal("task following changed display following")
	}
	if err := SaveSharedConfig(db, "scraper", 2, scraperTestConfig(0)); err == nil {
		t.Fatal("invalid timeout accepted")
	}
}

func TestSharedNewLibraryImportDoesNotChooseFirstPosterPreset(t *testing.T) {
	db := sharedTestDB(t)
	if err := SaveSharedConfig(db, "import", 0, importTestConfig()); err != nil {
		t.Fatal(err)
	}
	if err := db.Transaction(func(tx *gorm.DB) error { return SetLibraryConfigFollow(tx, "C", "import", true, 1) }); err != nil {
		t.Fatal(err)
	}
	raw, err := EffectiveLibraryConfig(db, "C", "import", "")
	if err != nil {
		t.Fatal(err)
	}
	value, _ := parseConfig(raw)
	if value["coverPosterType"] != float64(-1) {
		t.Fatal("new following library unexpectedly selected a local preset")
	}
	if _, ok := value["scanDiskPaths"]; ok {
		t.Fatal("source paths leaked into new library")
	}
}

func TestBrokenPublicConfigDoesNotBlockIndependentLibrary(t *testing.T) {
	db := sharedTestDB(t)
	if err := db.Create(&SharedLibraryConfig{Module: "display", Revision: 1, Config: "broken"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := GuardSharedConfigWrite(db, "A", "display", `{"pageLimit":32}`); err != nil {
		t.Fatal("independent save depends on broken public config", err)
	}
	if err := db.Create(&LibraryConfigFollow{FilesBasesID: "B", Module: "display"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := GuardSharedConfigWrite(db, "B", "display", `{"pageLimit":32}`); err == nil {
		t.Fatal("broken public config silently accepted")
	}
}
