package models

import (
	"cm_collectors_server/datatype"
	"encoding/json"
	"testing"

	"gorm.io/gorm"
)

func TestSharedConfigQuickStarsRemainsLocal(t *testing.T) {
	db := sharedTestDB(t)
	for _, field := range SharedConfigFields("display") {
		if field == "resourceQuickStars" {
			t.Fatal("quick stars must remain a library-specific preference")
		}
	}
	if err := SaveSharedConfig(db, "display", 0, displayTestConfig(t, 32)); err != nil {
		t.Fatal(err)
	}
	if err := db.Transaction(func(tx *gorm.DB) error { return SetLibraryConfigFollow(tx, "A", "display", true, 1) }); err != nil {
		t.Fatal(err)
	}
	for _, local := range []string{`{}`, `{"resourceQuickStars":false}`, `{"resourceQuickStars":true}`} {
		var before datatype.Config_FilesBases
		if err := json.Unmarshal([]byte(local), &before); err != nil {
			t.Fatal(err)
		}
		effective, err := EffectiveLibraryConfig(db, "A", "display", local)
		if err != nil {
			t.Fatal(err)
		}
		var after datatype.Config_FilesBases
		if err := json.Unmarshal([]byte(effective), &after); err != nil {
			t.Fatal(err)
		}
		if after.ResourceQuickStars != before.ResourceQuickStars {
			t.Fatalf("shared configuration changed local quick stars: %s", effective)
		}
		// 跟随公共配置时仍允许单独切换并保存本库开关。
		// 设置页提交原始配置对象，保留公共字段的缺省值和 JSON 形态。
		value, err := parseConfig(effective)
		if err != nil {
			t.Fatal(err)
		}
		value["resourceQuickStars"] = !after.ResourceQuickStars
		incoming, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if err := GuardSharedConfigWrite(db, "A", "display", string(incoming)); err != nil {
			t.Fatalf("following prevented saving the local quick stars switch: %v", err)
		}
	}
}
