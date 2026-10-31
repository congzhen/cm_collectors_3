package models

import (
	"encoding/json"
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"reflect"
	"testing"
)

func resourceGroupsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Resources{}, &VideoMoment{}, &ResourcePlayDaily{}, &FilesBasesSetting{}, &LibraryConfigFollow{}); err != nil {
		t.Fatal(err)
	}
	for _, r := range []Resources{{ID: "a", FilesBasesID: "lib", Title: "first"}, {ID: "b", FilesBasesID: "lib", Title: "second"}, {ID: "c", FilesBasesID: "lib", Title: "third"}, {ID: "d", FilesBasesID: "lib"}, {ID: "x", FilesBasesID: "other"}} {
		if err := db.Create(&r).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}
func resourceGroupIDs(t *testing.T, db *gorm.DB, group string) []string {
	t.Helper()
	ids := []string{}
	if err := db.Model(&Resources{}).Where("association_group_id = ?", group).Order("association_group_sort, id").Pluck("id", &ids).Error; err != nil {
		t.Fatal(err)
	}
	return ids
}
func TestResourceGroupsLifecycle(t *testing.T) {
	db := resourceGroupsTestDB(t)
	m := ResourceGroups{}
	group, err := m.Create(db, "lib", []string{"b", "a"})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Add(db, "lib", group, []string{"a", "c"}); err != nil {
		t.Fatal(err)
	}
	if got := resourceGroupIDs(t, db, group); !reflect.DeepEqual(got, []string{"b", "a", "c"}) {
		t.Fatal(got)
	}
	if err := m.Reorder(db, "lib", group, []string{"c", "b", "a"}); err != nil {
		t.Fatal(err)
	}
	if got := resourceGroupIDs(t, db, group); !reflect.DeepEqual(got, []string{"c", "b", "a"}) {
		t.Fatal(got)
	}
	summaries, total, err := m.Summaries(db, "lib", nil, "third", 1, 20)
	if err != nil || total != 1 || len(summaries) != 1 || summaries[0].Count != 3 || summaries[0].Preview[0].ID != "c" {
		t.Fatalf("%+v %d %v", summaries, total, err)
	}
	if err := m.Remove(db, "lib", group, []string{"c"}); err != nil {
		t.Fatal(err)
	}
	if len(resourceGroupIDs(t, db, group)) != 2 {
		t.Fatal("remaining pair lost")
	}
	if err := m.Remove(db, "lib", group, []string{"a"}); err != nil {
		t.Fatal(err)
	}
	if len(resourceGroupIDs(t, db, group)) != 0 {
		t.Fatal("singleton group retained")
	}
	var survivor Resources
	db.First(&survivor, "id = ?", "b")
	if survivor.ID != "b" || survivor.AssociationGroupID != "" {
		t.Fatal("resource deleted when group dissolved")
	}
}
func TestResourceGroupsRejectCrossLibraryAndConflictingMembership(t *testing.T) {
	db := resourceGroupsTestDB(t)
	m := ResourceGroups{}
	if _, err := m.Create(db, "lib", []string{"a", "x"}); err == nil {
		t.Fatal("cross-library create")
	}
	group, err := m.Create(db, "lib", []string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := m.Create(db, "lib", []string{"c", "d"})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Add(db, "lib", group, []string{"c"}); err == nil {
		t.Fatal("silently moved another group's resource")
	}
	if err := m.Add(db, "other", group, []string{"x"}); err == nil {
		t.Fatal("cross-library add")
	}
	if err := m.Remove(db, "lib", group, []string{"c"}); err == nil {
		t.Fatal("removed wrong group's member")
	}
	if err := m.Reorder(db, "lib", group, []string{"a", "c"}); err == nil {
		t.Fatal("invalid reorder")
	}
	if _, err := m.Create(db, "lib", []string{"a", "d"}); err == nil {
		t.Fatal("duplicate group membership")
	}
	if len(resourceGroupIDs(t, db, group)) != 2 || len(resourceGroupIDs(t, db, other)) != 2 {
		t.Fatal("failed operation changed groups")
	}
	summaries, total, err := m.Summaries(db, "other", []string{group}, "", 1, 20)
	if err != nil || total != 0 || len(summaries) != 0 {
		t.Fatal("group read escaped library", err)
	}
}
func TestResourceGroupsDeleteAndMoveCleanup(t *testing.T) {
	db := resourceGroupsTestDB(t)
	m := ResourceGroups{}
	group, err := m.Create(db, "lib", []string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if err := (Resources{}).DeleteById(db, "a"); err != nil {
		t.Fatal(err)
	}
	if len(resourceGroupIDs(t, db, group)) != 0 {
		t.Fatal("deletion left singleton")
	}
	group, err = m.Create(db, "lib", []string{"b", "c"})
	if err != nil {
		t.Fatal(err)
	}
	if err := (Resources{}).Update(db, &Resources{ID: "b", FilesBasesID: "other"}, []string{"filesBases_id"}); err != nil {
		t.Fatal(err)
	}
	if len(resourceGroupIDs(t, db, group)) != 0 {
		t.Fatal("move left cross-library group")
	}
	var moved Resources
	db.First(&moved, "id = ?", "b")
	if moved.FilesBasesID != "other" || moved.AssociationGroupID != "" {
		t.Fatal(moved)
	}
}
func TestResourceGroupDisplayIsIndependent(t *testing.T) {
	db := resourceGroupsTestDB(t)
	settings := FilesBasesSetting{FilesBasesID: "lib", ConfigJsonData: `{"pageLimit":32,"mergeResourceGroups":true,"resourceGroupExpandStyle":"strip","resourceGroupCoverStyle":"book"}`}
	if err := db.Create(&settings).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&LibraryConfigFollow{FilesBasesID: "lib", Module: "display"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := SetResourceGroupDisplay(db, "lib", false); err != nil {
		t.Fatal(err)
	}
	db.First(&settings, "filesBases_id = ?", "lib")
	var config map[string]interface{}
	json.Unmarshal([]byte(settings.ConfigJsonData), &config)
	if config["mergeResourceGroups"] != false || config["pageLimit"] != float64(32) || config["resourceGroupExpandStyle"] != "strip" || config["resourceGroupCoverStyle"] != "book" {
		t.Fatal(config)
	}
	var count int64
	db.Model(&LibraryConfigFollow{}).Count(&count)
	if count != 1 {
		t.Fatal("changed public follow state")
	}
	for _, field := range SharedConfigFields("display") {
		if field == "mergeResourceGroups" || field == "resourceGroupExpandStyle" || field == "resourceGroupCoverStyle" {
			t.Fatal("group display must remain independent")
		}
	}
}

func TestResourceGroupMigrationPreservesExistingResources(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("CREATE TABLE resources (id TEXT PRIMARY KEY, filesBases_id TEXT, title TEXT)").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO resources (id, filesBases_id, title) VALUES ('legacy', 'lib', 'existing title')").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Resources{}); err != nil {
		t.Fatal(err)
	}
	var resource Resources
	if err := db.First(&resource, "id = ?", "legacy").Error; err != nil {
		t.Fatal(err)
	}
	if resource.Title != "existing title" || resource.AssociationGroupID != "" || resource.AssociationGroupSort != 0 {
		t.Fatalf("migration changed existing data: %+v", resource)
	}
	if err := db.AutoMigrate(&Resources{}); err != nil {
		t.Fatal(err)
	}
}

func TestResourceGroupsCapacityRejectsAddWithoutChangingMembership(t *testing.T) {
	db := resourceGroupsTestDB(t)
	members := make([]Resources, maxResourceGroupMembers)
	for i := range members {
		members[i] = Resources{ID: fmt.Sprintf("member%04d", i), FilesBasesID: "lib", AssociationGroupID: "full-group", AssociationGroupSort: i}
	}
	if err := db.CreateInBatches(members, 20).Error; err != nil {
		t.Fatal(err)
	}
	if err := (ResourceGroups{}).Add(db, "lib", "full-group", []string{"c"}); err == nil {
		t.Fatal("accepted an over-capacity group")
	}
	var candidate Resources
	if err := db.First(&candidate, "id = ?", "c").Error; err != nil {
		t.Fatal(err)
	}
	if candidate.AssociationGroupID != "" || len(resourceGroupIDs(t, db, "full-group")) != maxResourceGroupMembers {
		t.Fatal("capacity check changed membership")
	}
}

func TestResourceGroupsInfoMatchesResourceDetailData(t *testing.T) {
	db := resourceGroupsTestDB(t)
	if err := db.SetupJoinTable(&Resources{}, "Performers", &ResourcesPerformers{}); err != nil {
		t.Fatal(err)
	}
	if err := db.SetupJoinTable(&Resources{}, "Directors", &ResourcesDirectors{}); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&ResourcesDramaSeries{}, &ResourcesVideoMetadata{}, &TagClass{}, &ResourcesPerformers{}, &ResourcesDirectors{}); err != nil {
		t.Fatal(err)
	}
	group, err := (ResourceGroups{}).Create(db, "lib", []string{"b", "a"})
	if err != nil {
		t.Fatal(err)
	}
	performer := Performer{ID: "actor"}
	if err := db.Create(&performer).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&ResourcesPerformers{ID: "actor-a", ResourcesID: "a", PerformerID: performer.ID}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&ResourcesDirectors{ID: "actor-b", ResourcesID: "b", PerformerID: performer.ID}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&ResourcesPerformers{ID: "actor-x", ResourcesID: "x", PerformerID: performer.ID}).Error; err != nil {
		t.Fatal(err)
	}
	for _, video := range []ResourcesDramaSeries{{ID: "later", ResourcesID: "a", Sort: 2, Src: "/later.mp4"}, {ID: "first", ResourcesID: "a", Sort: 1, Src: "/first.mp4"}} {
		if err := db.Create(&video).Error; err != nil {
			t.Fatal(err)
		}
	}
	members, err := (ResourceGroups{}).Info(db, "lib", group)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 2 || members[0].ID != "b" || members[1].ID != "a" {
		t.Fatalf("unexpected order: %+v", members)
	}
	if len(members[0].Directors) != 1 || members[0].Directors[0].ResourceCount != 2 || len(members[1].Performers) != 1 || members[1].Performers[0].ResourceCount != 2 {
		t.Fatalf("actor counts differ from ordinary resource detail: %+v", members)
	}
	if len(members[1].ResourcesDramaSeries) != 2 || members[1].ResourcesDramaSeries[0].ID != "first" {
		t.Fatal("video data or ordering lost")
	}
	for _, member := range members {
		if member.Tags == nil || member.Performers == nil || member.Directors == nil || member.ResourcesDramaSeries == nil {
			t.Fatal("nil resource associations")
		}
	}
	if _, err := (ResourceGroups{}).Info(db, "other", group); err == nil {
		t.Fatal("group detail escaped library")
	}
}

func TestResourceGroupSummariesBatchAndLibraryPagination(t *testing.T) {
	db := resourceGroupsTestDB(t)
	ids := make([]string, 23)
	rows := make([]Resources, 0, 46)
	for i := range ids {
		ids[i] = fmt.Sprintf("group-%02d", i)
		for j := 0; j < 2; j++ {
			rows = append(rows, Resources{ID: fmt.Sprintf("batch-%02d-%d", i, j), FilesBasesID: "lib", AssociationGroupID: ids[i], AssociationGroupSort: j, Title: "batch member"})
		}
	}
	if err := db.CreateInBatches(rows, 20).Error; err != nil {
		t.Fatal(err)
	}
	batch, total, err := (ResourceGroups{}).Summaries(db, "lib", ids, "", 1, 20)
	if err != nil || total != 23 || len(batch) != 23 {
		t.Fatalf("batch summaries truncated: %d %d %v", len(batch), total, err)
	}
	page, total, err := (ResourceGroups{}).Summaries(db, "lib", nil, "", 2, 20)
	if err != nil || total != 23 || len(page) != 3 || page[0].ID != ids[20] {
		t.Fatalf("invalid library group pagination: %+v %d %v", page, total, err)
	}
	if err := db.Model(&Resources{}).Where("id = ?", "batch-22-0").Update("issueNumber", "UNIQUE-ISSUE").Error; err != nil {
		t.Fatal(err)
	}
	page, total, err = (ResourceGroups{}).Summaries(db, "lib", nil, "UNIQUE-ISSUE", 1, 20)
	if err != nil || total != 1 || len(page) != 1 || page[0].ID != ids[22] {
		t.Fatalf("issue search did not find group: %+v %d %v", page, total, err)
	}
}

func TestResourceGroupMoveFailureRollsBackMembership(t *testing.T) {
	db := resourceGroupsTestDB(t)
	group, err := (ResourceGroups{}).Create(db, "lib", []string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TRIGGER reject_group_move BEFORE UPDATE OF filesBases_id ON resources WHEN NEW.filesBases_id = 'other' BEGIN SELECT RAISE(ABORT, 'move blocked'); END`).Error; err != nil {
		t.Fatal(err)
	}
	if err := (Resources{}).Update(db, &Resources{ID: "a", FilesBasesID: "other"}, []string{"filesBases_id"}); err == nil {
		t.Fatal("move unexpectedly succeeded")
	}
	if got := resourceGroupIDs(t, db, group); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("failed move damaged group: %v", got)
	}
	var resource Resources
	if err := db.First(&resource, "id = ?", "a").Error; err != nil {
		t.Fatal(err)
	}
	if resource.FilesBasesID != "lib" {
		t.Fatal("failed move changed library")
	}
}

func TestResourceGroupSummaryCoverLimit(t *testing.T) {
	db := resourceGroupsTestDB(t)
	ids := []string{}
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("preview-%02d", i)
		if err := db.Create(&Resources{ID: id, FilesBasesID: "lib"}).Error; err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	model := ResourceGroups{}
	if _, err := model.Create(db, "lib", ids); err != nil {
		t.Fatal(err)
	}
	summaries, total, err := model.Summaries(db, "lib", nil, "", 1, 20)
	if err != nil || total != 1 || len(summaries) != 1 {
		t.Fatalf("%+v %d %v", summaries, total, err)
	}
	if summaries[0].Count != 10 || len(summaries[0].Preview) != 7 {
		t.Fatalf("%+v", summaries[0])
	}
	for i, resource := range summaries[0].Preview {
		if resource.ID != ids[i] {
			t.Fatal("preview order changed")
		}
	}
}

func TestResourceGroupPreviewOnlySerializesCoverFields(t *testing.T) {
	db := resourceGroupsTestDB(t)
	if err := db.Model(&Resources{}).Where("id = ?", "a").Updates(map[string]interface{}{"coverPoster": "cover.jpg", "coverPosterWidth": 316, "coverPosterHeight": 218}).Error; err != nil {
		t.Fatal(err)
	}
	group, err := (ResourceGroups{}).Create(db, "lib", []string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	summaries, _, err := (ResourceGroups{}).Summaries(db, "lib", []string{group}, "", 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(summaries[0].Preview[0])
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]interface{}
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	if fields["coverPoster"] != "cover.jpg" || fields["coverPosterWidth"] != float64(316) || fields["filesBases_id"] != "lib" {
		t.Fatal(fields)
	}
	for _, key := range []string{"dramaSeries", "performers", "directors", "tags", "abstract", "lastPlayFile"} {
		if _, exists := fields[key]; exists {
			t.Fatalf("summary contains unrelated resource field %s", key)
		}
	}
}

func TestResourceGroupDeleteFailureRollsBackAllRelations(t *testing.T) {
	for _, members := range [][]string{{"a", "b"}, {"a", "b", "c"}} {
		t.Run(fmt.Sprintf("members-%d", len(members)), func(t *testing.T) {
			db := resourceGroupsTestDB(t)
			group, err := (ResourceGroups{}).Create(db, "lib", members)
			if err != nil {
				t.Fatal(err)
			}
			if err := db.Create(&VideoMoment{ID: "keep-moment", ResourceID: "a", DramaSeriesID: "video-a", Title: "keep"}).Error; err != nil {
				t.Fatal(err)
			}
			if err := db.Exec(`CREATE TRIGGER reject_group_delete BEFORE DELETE ON resources WHEN OLD.id = 'a' BEGIN SELECT RAISE(ABORT, 'delete blocked'); END`).Error; err != nil {
				t.Fatal(err)
			}
			if err := (Resources{}).DeleteById(db, "a"); err == nil {
				t.Fatal("delete unexpectedly succeeded")
			}
			if got := resourceGroupIDs(t, db, group); !reflect.DeepEqual(got, members) {
				t.Fatalf("failed delete damaged group: %v", got)
			}
			var resource Resources
			if err := db.First(&resource, "id = ?", "a").Error; err != nil {
				t.Fatal(err)
			}
			if resource.AssociationGroupID != group || resource.FilesBasesID != "lib" {
				t.Fatal("resource relation not rolled back")
			}
			var count int64
			if err := db.Model(&VideoMoment{}).Where("id = ?", "keep-moment").Count(&count).Error; err != nil || count != 1 {
				t.Fatal("failed delete removed moment", err)
			}
		})
	}
}
