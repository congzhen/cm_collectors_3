package models

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"reflect"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func sidebarTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { sqlDB.Close() })
	if err := db.AutoMigrate(&Performer{}, &Resources{}, &ResourcesPerformers{}, &ResourcesDirectors{}, &FilesRelatedPerformerBases{}, &ResourcePlayDaily{}, &VideoMoment{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestSidebarPerformerModesFilterBeforeLimitAndPreservePreferred(t *testing.T) {
	db := sidebarTestDB(t)
	now := core.TimeNow()
	for i, id := range []string{"pinned", "count", "hot", "recent", "empty", "hidden", "deleted"} {
		created := datatype.CustomTime(now.Add(time.Duration(i) * time.Hour))
		photo := "face.jpg"
		if id == "hidden" {
			photo = ""
		}
		p := Performer{ID: id, PerformerBasesID: "base", Photo: photo, CreatedAt: &created, Status: true}
		if err := db.Create(&p).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Model(&Performer{}).Where("id = ?", "deleted").Update("status", false).Error; err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		id, actor, lib string
		hot            int
		active         bool
	}{
		{"pin", "pinned", "lib", 999, true},
		{"c1", "count", "lib", 1, true}, {"c2", "count", "lib", 1, true},
		{"h", "hot", "lib", 20, true}, {"r", "recent", "lib", 3, true},
		{"hidden", "hidden", "lib", 9999, true},
		{"deleted", "deleted", "lib", 99999, true},
		{"outside", "count", "other-lib", 9999, true},
		{"disabled", "count", "lib", 9999, false},
	} {
		r := Resources{ID: item.id, FilesBasesID: item.lib, Hot: item.hot, Status: true}
		if err := db.Create(&r).Error; err != nil {
			t.Fatal(err)
		}
		if !item.active {
			db.Model(&Resources{}).Where("id = ?", item.id).Update("status", false)
		}
		if err := db.Create(&ResourcesPerformers{ID: item.id, ResourcesID: item.id, PerformerID: item.actor}).Error; err != nil {
			t.Fatal(err)
		}
	}
	// 同一资源同时关联演员和导演只能计一次；否则 count 会超过 recent 的近期热度。
	if err := db.Create(&ResourcesDirectors{ID: "double", ResourcesID: "c1", PerformerID: "count"}).Error; err != nil {
		t.Fatal(err)
	}
	for _, row := range []ResourcePlayDaily{
		{ResourcesID: "c1", PlayDate: now.Format("2006-01-02"), PlayCount: 2},
		{ResourcesID: "r", PlayDate: now.AddDate(0, 0, -29).Format("2006-01-02"), PlayCount: 3},
		{ResourcesID: "h", PlayDate: now.AddDate(0, 0, -30).Format("2006-01-02"), PlayCount: 100},
		{ResourcesID: "h", PlayDate: now.AddDate(0, 0, 1).Format("2006-01-02"), PlayCount: 100},
		{ResourcesID: "outside", PlayDate: now.Format("2006-01-02"), PlayCount: 1000},
		{ResourcesID: "disabled", PlayDate: now.Format("2006-01-02"), PlayCount: 1000},
		{ResourcesID: "hidden", PlayDate: now.Format("2006-01-02"), PlayCount: 1000},
	} {
		if err := db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, test := range []struct {
		mode string
		want []string
	}{
		{"default", []string{"pinned", "empty", "recent"}},
		{PerformerSortResourceCountDesc, []string{"pinned", "count", "recent"}},
		{"hotDesc", []string{"pinned", "hot", "recent"}},
		{"recentDesc", []string{"pinned", "recent", "count"}},
	} {
		t.Run(test.mode, func(t *testing.T) {
			preferred := []string{"hidden", "pinned", "pinned", "deleted", "missing"}
			rows, err := (Performer{}).ListTopPreferredPerformers(db, preferred, "base", true, 3, "lib", SidebarPerformerOptions{SortMode: test.mode, RecentDays: 30})
			if err != nil {
				t.Fatal(err)
			}
			ids := []string{}
			for _, p := range *rows {
				ids = append(ids, p.ID)
			}
			if !reflect.DeepEqual(ids, test.want) {
				t.Fatalf("got %v want %v", ids, test.want)
			}
			rows, err = (Performer{}).ListTopPreferredPerformers(db, preferred, "base", false, 3, "lib", SidebarPerformerOptions{SortMode: test.mode})
			if err != nil {
				t.Fatal(err)
			}
			if len(*rows) != 3 || (*rows)[0].ID != "hidden" || (*rows)[1].ID != "pinned" {
				t.Fatalf("preferred selection not restored: %+v", rows)
			}
			rows, err = (Performer{}).ListTopPreferredPerformers(db, nil, "base", false, 20, "lib", SidebarPerformerOptions{SortMode: test.mode})
			if err != nil {
				t.Fatal(err)
			}
			if len(*rows) != 6 {
				t.Fatalf("insufficient candidates should return six active actors, got %d", len(*rows))
			}
		})
	}
	rows, err := (Performer{}).ListTopPreferredPerformers(db, []string{"pinned", "count"}, "base", true, 1, "lib")
	if err != nil || len(*rows) != 1 || (*rows)[0].ID != "pinned" {
		t.Fatalf("limit not honored: %v %v", rows, err)
	}
	rows, err = (Performer{}).ListTopPreferredPerformers(db, []string{"pinned"}, "base", true, 0, "lib")
	if err != nil || len(*rows) != 0 {
		t.Fatalf("zero limit: %v %v", rows, err)
	}
	// 清空近期统计时按默认顺序补齐；设置七天则排除29天前记录。
	rows, err = (Performer{}).ListTopPreferredPerformers(db, nil, "base", true, 1, "lib", SidebarPerformerOptions{SortMode: "recentDesc", RecentDays: 7})
	if err != nil || (*rows)[0].ID != "count" {
		t.Fatalf("seven-day window: %v %v", rows, err)
	}
	db.Where("1=1").Delete(&ResourcePlayDaily{})
	rows, err = (Performer{}).ListTopPreferredPerformers(db, nil, "base", true, 1, "lib", SidebarPerformerOptions{SortMode: "recentDesc"})
	if err != nil || (*rows)[0].ID != "empty" {
		t.Fatalf("empty history fallback: %v %v", rows, err)
	}
}

func TestRecordResourcePlayDailyAndRollback(t *testing.T) {
	db := sidebarTestDB(t)
	if err := db.Create(&Resources{ID: "r", FilesBasesID: "lib", Hot: 10}).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 26, 12, 0, 0, 0, time.Local)
	for _, when := range []time.Time{now, now, now.AddDate(0, 0, 1)} {
		if err := RecordResourcePlay(db, "r", "video.mp4", when); err != nil {
			t.Fatal(err)
		}
	}
	var r Resources
	db.First(&r, "id = ?", "r")
	var days []ResourcePlayDaily
	db.Order("play_date").Find(&days)
	if r.Hot != 13 || r.LastPlayFile != "video.mp4" || len(days) != 2 || days[0].PlayCount != 2 || days[1].PlayCount != 1 {
		t.Fatalf("unexpected totals: %+v %+v", r, days)
	}
	if r.LastPlayTime == nil || time.Time(*r.LastPlayTime).Format("2006-01-02 15:04:05") != now.AddDate(0, 0, 1).Format("2006-01-02 15:04:05") {
		t.Fatal("last play time not updated")
	}
	if err := RecordResourcePlay(db, "missing", "", now); err != gorm.ErrRecordNotFound {
		t.Fatalf("missing resource: %v", err)
	}
	if err := db.Migrator().DropTable(&ResourcePlayDaily{}); err != nil {
		t.Fatal(err)
	}
	if err := RecordResourcePlay(db, "r", "failed.mp4", now); err == nil {
		t.Fatal("expected daily write failure")
	}
	db.First(&r, "id = ?", "r")
	if r.Hot != 13 || r.LastPlayFile != "video.mp4" {
		t.Fatal("failed daily write must roll back resource changes")
	}
}

func TestDeleteResourceCleansDailyStats(t *testing.T) {
	db := sidebarTestDB(t)
	for _, id := range []string{"one", "two"} {
		db.Create(&Resources{ID: id, FilesBasesID: "lib"})
		if err := RecordResourcePlay(db, id, "", core.TimeNow()); err != nil {
			t.Fatal(err)
		}
	}
	if err := (Resources{}).DeleteById(db, "one"); err != nil {
		t.Fatal(err)
	}
	var n int64
	db.Model(&ResourcePlayDaily{}).Count(&n)
	if n != 1 {
		t.Fatalf("want one remaining daily record, got %d", n)
	}
	if err := (Resources{}).DeleteByFilesBasesID(db, "lib"); err != nil {
		t.Fatal(err)
	}
	db.Model(&ResourcePlayDaily{}).Count(&n)
	if n != 0 {
		t.Fatalf("library deletion left %d daily records", n)
	}
}
