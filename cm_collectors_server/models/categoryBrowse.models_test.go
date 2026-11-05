package models

import (
	"cm_collectors_server/datatype"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func categoryBrowseTestDB(t testing.TB) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	connection, _ := db.DB()
	connection.SetMaxOpenConns(1)
	t.Cleanup(func() { connection.Close() })
	for name, model := range map[string]interface{}{"Performers": &ResourcesPerformers{}, "Directors": &ResourcesDirectors{}, "Tags": &ResourcesTags{}} {
		if err = db.SetupJoinTable(&Resources{}, name, model); err != nil {
			t.Fatal(err)
		}
	}
	if err = db.AutoMigrate(&Resources{}, &ResourcesDramaSeries{}, &ResourcesPerformers{}, &ResourcesDirectors{}, &ResourcesTags{}, &TagClass{}, &FilesBasesSetting{}, &LibraryConfigFollow{}, &FilesRelatedPerformerBases{}); err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`INSERT INTO performer(id,name,status) VALUES ('a','演员甲',1),('b','演员乙',1),('off','停用演员',0)`,
		`UPDATE performer SET performerBases_id = 'actors'`,
		`INSERT INTO filesRelatedPerformerBases(id,filesBases_id,performerBases_id) VALUES ('rel-lib','lib','actors')`,
		`INSERT INTO tagClass(id,name,filesBases_id,status) VALUES ('publisher','厂牌','lib',1),('foreign','其他库','other',1)`,
		`INSERT INTO tag(id,name,tagClass_id,status) VALUES ('t1','厂牌甲','publisher',1),('t2','厂牌乙','publisher',1),('t3','停用标签','publisher',0),('tx','外库标签','foreign',1)`,
	} {
		if err = db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 17; i++ {
		id := fmt.Sprintf("r%02d", i)
		stars := 1
		if i%2 == 0 {
			stars = 5
		}
		if err = db.Create(&Resources{ID: id, FilesBasesID: "lib", Title: id, Stars: stars}).Error; err != nil {
			t.Fatal(err)
		}
		if i < 13 {
			if err = db.Exec(`INSERT INTO resourcesPerformers(resources_id,performer_id) VALUES (?, 'a')`, id).Error; err != nil {
				t.Fatal(err)
			}
			if err = db.Exec(`INSERT INTO resourcesTags(resources_id,tag_id) VALUES (?, 't1')`, id).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, s := range []string{
		`INSERT INTO resourcesPerformers(resources_id,performer_id) VALUES ('r00','b'),('r13','b'),('r14','off'),('r15','missing')`,
		`INSERT INTO resourcesTags(resources_id,tag_id) VALUES ('r00','t2'),('r13','t2'),('r14','t3'),('r15','tx')`,
		`INSERT INTO resources(id,filesBases_id,title) VALUES ('foreign','other','外库资源')`,
		`INSERT INTO resourcesPerformers(resources_id,performer_id) VALUES ('foreign','a')`,
	} {
		if err = db.Exec(s).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}
func TestCategoryBrowsePartitionsAndFilters(t *testing.T) {
	db := categoryBrowseTestDB(t)
	for _, source := range []string{"performer", "tag:publisher"} {
		par := CategoryBrowseRequest{FilesBasesID: "lib", Source: source, GroupLimit: 4, Page: 1, Limit: 100}
		result, err := (CategoryBrowse{}).Summaries(db, par)
		if err != nil {
			t.Fatal(err)
		}
		if result.Total != 6 || result.ResourceTotal != 17 || result.CategoryTotal != 3 {
			t.Fatalf("unexpected summary: %+v", result)
		}
		counts := []int64{}
		for _, g := range result.DataList {
			counts = append(counts, g.Count)
			if len(g.Preview) != int(g.Count) {
				t.Fatalf("preview should not exceed part: %+v", g)
			}
		}
		if !reflect.DeepEqual(counts, []int64{4, 4, 4, 1, 2, 3}) {
			t.Fatal(counts)
		}
		last := result.DataList[len(result.DataList)-1]
		if last.CategoryID != "" {
			t.Fatal("unclassified must be last")
		}
		par.SearchData.Star = datatype.I_searchGroup{Logic: datatype.E_searchLogic_single, Options: []string{"5"}}
		filtered, err := (CategoryBrowse{}).Summaries(db, par)
		if err != nil {
			t.Fatal(err)
		}
		if filtered.ResourceTotal != 9 || filtered.Total != 4 {
			t.Fatalf("filter not followed: %+v", filtered)
		}
		par.SearchData.SearchTextSlc = []string{"not-found"}
		empty, err := (CategoryBrowse{}).Summaries(db, par)
		if err != nil || empty.Total != 0 || len(empty.DataList) != 0 {
			t.Fatalf("empty: %+v %v", empty, err)
		}
	}
}

func TestCategoryBrowseCategorySort(t *testing.T) {
	db := categoryBrowseTestDB(t)
	for _, table := range []string{"performer", "tag"} {
		if err := db.Exec("UPDATE " + table + " SET name = 'B', addTime = '2024-01-01 00:00:00' WHERE id IN ('a', 't1')").Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Exec("UPDATE " + table + " SET name = 'A', addTime = '2025-01-01 00:00:00' WHERE id IN ('b', 't2')").Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, source := range []string{"performer", "tag:publisher"} {
		large, small := "a", "b"
		if source != "performer" {
			large, small = "t1", "t2"
		}
		for _, mode := range []string{"", "count", "name", "default"} {
			par := CategoryBrowseRequest{FilesBasesID: "lib", Source: source, CategorySort: mode, GroupLimit: 4, Limit: 100}
			result, err := (CategoryBrowse{}).Summaries(db, par)
			if err != nil {
				t.Fatal(err)
			}
			want := []string{large, large, large, large, small, ""}
			if mode == "name" || mode == "default" {
				want = []string{small, large, large, large, large, ""}
			}
			got := []string{}
			for _, group := range result.DataList {
				got = append(got, group.CategoryID)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("%s %s: got %v want %v", source, mode, got, want)
			}
			// Paging must preserve the same category and adjacent-part order.
			paged := []string{}
			par.Limit = 2
			for page := 1; page <= 3; page++ {
				par.Page = page
				result, err = (CategoryBrowse{}).Summaries(db, par)
				if err != nil {
					t.Fatal(err)
				}
				for _, group := range result.DataList {
					paged = append(paged, group.CategoryID)
				}
			}
			if !reflect.DeepEqual(paged, want) {
				t.Fatal(paged, want)
			}
		}
		// Filtering reduces card membership but does not change lifetime ranking.
		par := CategoryBrowseRequest{FilesBasesID: "lib", Source: source, GroupLimit: 0, SearchData: datatype.ReqParam_SearchData{SearchTextSlc: []string{"r00"}}}
		result, err := (CategoryBrowse{}).Summaries(db, par)
		if err != nil || len(result.DataList) != 2 || result.DataList[0].CategoryID != large || result.DataList[0].Total != 1 {
			t.Fatalf("filtered sort: %+v %v", result, err)
		}
		par.CategorySort = "invalid"
		if _, err := (CategoryBrowse{}).Summaries(db, par); err == nil {
			t.Fatal("accepted invalid category sort")
		}
	}
	if err := db.Create(&FilesBasesSetting{FilesBasesID: "lib", ConfigJsonData: `{}`}).Error; err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"", "count", "name", "default", "invalid"} {
		err := SetCategoryBrowseSettings(db, "lib", CategoryBrowseSettings{Source: "performer", GroupLimit: 12, Sort: mode})
		if mode == "invalid" {
			if err == nil {
				t.Fatal("saved invalid category sort")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		var setting FilesBasesSetting
		db.First(&setting, "filesBases_id = ?", "lib")
		var config map[string]interface{}
		json.Unmarshal([]byte(setting.ConfigJsonData), &config)
		want := mode
		if want == "" {
			want = "count"
		}
		if config["categoryBrowseSort"] != want {
			t.Fatal(config)
		}
	}
}
func TestCategoryBrowseLifetimeCountRanking(t *testing.T) {
	db := categoryBrowseTestDB(t)
	if err := db.Exec(`INSERT INTO filesRelatedPerformerBases(id,filesBases_id,performerBases_id) VALUES ('rel-other','other','actors')`).Error; err != nil {
		t.Fatal(err)
	}
	// The second category has fewer members here, but more overall applications.
	for i := 0; i < 15; i++ {
		id := fmt.Sprintf("other-%02d", i)
		for _, statement := range []struct {
			sql  string
			args []interface{}
		}{
			{`INSERT INTO resources(id,filesBases_id,title,status) VALUES (?,'other',?,1)`, []interface{}{id, id}},
			{`INSERT INTO resourcesPerformers(resources_id,performer_id) VALUES (?,'b')`, []interface{}{id}},
			{`INSERT INTO resourcesDirectors(resources_id,performer_id) VALUES (?,'b')`, []interface{}{id}},
			{`INSERT INTO resourcesTags(resources_id,tag_id) VALUES (?,'t2')`, []interface{}{id}},
		} {
			if err := db.Exec(statement.sql, statement.args...).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, source := range []string{"performer", "tag:publisher"} {
		first := "b"
		if source != "performer" {
			first = "t2"
		}
		par := CategoryBrowseRequest{FilesBasesID: "lib", Source: source, GroupLimit: 4, Limit: 100}
		result, err := (CategoryBrowse{}).Summaries(db, par)
		if err != nil || result.DataList[0].CategoryID != first || result.DataList[0].Total != 2 || result.ResourceTotal != 17 {
			t.Fatalf("lifetime rank must keep current members: %+v %v", result, err)
		}
		par.SearchData.SearchTextSlc = []string{"r00"}
		result, err = (CategoryBrowse{}).Summaries(db, par)
		if err != nil || len(result.DataList) != 2 || result.DataList[0].CategoryID != first || result.DataList[0].Total != 1 {
			t.Fatalf("filter changed lifetime ranking: %+v %v", result, err)
		}
	}
	// Reuse the libraries' effective-resource counts: duplicate actor/director
	// links are counted once and disabled resources excluded.
	counts, err := (Performer{}).resourceCountMap(db, []string{"b"}, "")
	if err != nil || counts["b"] != 17 {
		t.Fatal(counts, err)
	}
	for _, source := range []string{"performer", "tag:publisher"} {
		ranked, err := categoryBrowseRankCounts(db, &CategoryBrowseRequest{FilesBasesID: "lib", Source: source})
		key := "b"
		if source != "performer" {
			key = "t2"
		}
		if err != nil || ranked[key] != 17 {
			t.Fatal("aggregate totals must match library counts", source, ranked, err)
		}
	}
	if err := db.Exec(`UPDATE resources SET status = 0 WHERE filesBases_id = 'other'`).Error; err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{"performer", "tag:publisher"} {
		first := "a"
		if source != "performer" {
			first = "t1"
		}
		result, err := (CategoryBrowse{}).Summaries(db, CategoryBrowseRequest{FilesBasesID: "lib", Source: source, GroupLimit: 0})
		if err != nil || result.DataList[0].CategoryID != first {
			t.Fatalf("disabled resources contributed to ranking: %+v %v", result, err)
		}
	}
}

func categoryBrowseLargeTestDB(t testing.TB) *gorm.DB {
	t.Helper()
	db := categoryBrowseTestDB(t)
	for _, statement := range []string{
		`WITH RECURSIVE n(x) AS (VALUES(1) UNION ALL SELECT x+1 FROM n WHERE x<17000)
		 INSERT INTO performer(id,name,performerBases_id,status)
		 SELECT printf('bulk-%05d',x), 'bulk', 'actors', 1 FROM n`,
		`INSERT INTO resourcesPerformers(resources_id,performer_id) SELECT 'r00',id FROM performer WHERE id LIKE 'bulk-%'`,
		`INSERT INTO tag(id,name,tagClass_id,status) SELECT id,name,'publisher',1 FROM performer WHERE id LIKE 'bulk-%'`,
		`INSERT INTO resourcesTags(resources_id,tag_id) SELECT 'r00',id FROM tag WHERE id LIKE 'bulk-%'`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func BenchmarkCategoryBrowseSummaries(b *testing.B) {
	db := categoryBrowseLargeTestDB(b)
	for _, source := range []string{"performer", "tag:publisher"} {
		b.Run(source, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := (CategoryBrowse{}).Summaries(db, CategoryBrowseRequest{FilesBasesID: "lib", Source: source, GroupLimit: 0, Limit: 24}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func categoryBrowseSplitTestDB(t testing.TB) *gorm.DB {
	t.Helper()
	db := categoryBrowseTestDB(t)
	for _, statement := range []string{
		`WITH RECURSIVE n(x) AS (VALUES(1) UNION ALL SELECT x+1 FROM n WHERE x<1000)
		 INSERT INTO resources(id,filesBases_id,title,status) SELECT printf('dense-%04d',x),'lib','dense',1 FROM n`,
		`INSERT INTO resourcesPerformers(resources_id,performer_id) SELECT id,'a' FROM resources WHERE id LIKE 'dense-%'`,
		`INSERT INTO resourcesTags(resources_id,tag_id) SELECT id,'t1' FROM resources WHERE id LIKE 'dense-%'`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func BenchmarkCategoryBrowseSplitPreviews(b *testing.B) {
	db := categoryBrowseSplitTestDB(b)
	for _, source := range []string{"performer", "tag:publisher"} {
		b.Run(source, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := (CategoryBrowse{}).Summaries(db, CategoryBrowseRequest{FilesBasesID: "lib", Source: source, GroupLimit: 12, Limit: 100}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func TestCategoryBrowseBatchedPreviews(t *testing.T) {
	db := categoryBrowseSplitTestDB(t)
	queries := 0
	if err := db.Callback().Query().After("gorm:query").Register("test:previewQueries", func(tx *gorm.DB) {
		if !tx.DryRun {
			queries++
		}
	}); err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{"performer", "tag:publisher"} {
		category := "a"
		if source != "performer" {
			category = "t1"
		}
		par := CategoryBrowseRequest{FilesBasesID: "lib", Source: source, Limit: 100, SearchData: datatype.ReqParam_SearchData{Sort: datatype.E_searchSort_starDesc}}
		var reference []Resources
		if err := categoryBrowseOrder(categoryBrowseMembers(db, &par, category), &par).Select("resources.id").Find(&reference).Error; err != nil {
			t.Fatal(err)
		}
		for _, cap := range []int{0, 4, 12, 16, 17, 100} {
			par.GroupLimit = cap
			for _, page := range []int{1, 2} {
				par.Page = page
				queries = 0
				result, err := (CategoryBrowse{}).Summaries(db, par)
				if err != nil {
					t.Fatal(err)
				}
				if cap == 12 && page == 1 && queries > 15 {
					t.Fatalf("repeated preview queries: %d", queries)
				}
				for _, group := range result.DataList {
					if len(group.Preview) > 7 {
						t.Fatal("preview overfetch leaked to response")
					}
					if group.CategoryID != category {
						continue
					}
					start := (group.Part - 1) * cap
					want := reference[start:min(start+7, start+int(group.Count))]
					if len(group.Preview) != len(want) {
						t.Fatal("preview size differs", cap, page, group.Part)
					}
					for i, cover := range group.Preview {
						if cover.ID != want[i].ID {
							t.Fatal("preview order/part differs", source, cap, page, group.Part, cover.ID, want[i].ID)
						}
					}
				}
			}
		}
	}
}

func TestCategoryBrowseLargeCategoryCountRanking(t *testing.T) {
	db := categoryBrowseLargeTestDB(t)
	for _, source := range []string{"performer", "tag:publisher"} {
		result, err := (CategoryBrowse{}).Summaries(db, CategoryBrowseRequest{FilesBasesID: "lib", Source: source, GroupLimit: 0, Limit: 2})
		if err != nil {
			t.Fatal(source, err)
		}
		if result.Total != 17003 || result.DataList[0].Total != 13 || result.DataList[1].Total != 2 {
			t.Fatalf("large category ranking: %+v", result)
		}
		result, err = (CategoryBrowse{}).Summaries(db, CategoryBrowseRequest{FilesBasesID: "lib", Source: source, GroupLimit: 0, Limit: 2, Page: 2})
		if err != nil || result.DataList[0].CategoryID != "bulk-17000" || result.DataList[1].CategoryID != "bulk-16999" {
			t.Fatalf("equal totals must retain stable ordering across batches: %+v %v", result, err)
		}
	}
}

// Compare complete optimized pages with independently partitioned member lists,
// including page boundaries that start inside one category and end in another.
func TestCategoryBrowseOptimizedPagesMatchMembers(t *testing.T) {
	db := categoryBrowseSplitTestDB(t)
	if err := db.Exec(`UPDATE resources SET coverPoster = id || '.jpg', coverPosterWidth = 320, coverPosterHeight = 480, abstract = 'not a cover field'`).Error; err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{"performer", "tag:publisher"} {
		large, small := "a", "b"
		if source != "performer" {
			large, small = "t1", "t2"
		}
		for _, filter := range []datatype.ReqParam_SearchData{
			{},
			{Star: datatype.I_searchGroup{Logic: datatype.E_searchLogic_single, Options: []string{"5"}}},
			{Performer: datatype.I_searchGroup{Logic: datatype.E_searchLogic_single, Options: []string{"a"}}},
			{Tag: map[string]datatype.I_searchGroup{"publisher": {Logic: datatype.E_searchLogic_single, Options: []string{"t1"}}}},
		} {
			filter.Sort = datatype.E_searchSort_scoreDesc
			par := CategoryBrowseRequest{FilesBasesID: "lib", Source: source, SearchData: filter, Limit: 7}
			references := map[string][]Resources{}
			for _, category := range []string{large, small, ""} {
				var members []Resources
				if err := categoryBrowseOrder(categoryBrowseMembers(db, &par, category), &par).Find(&members).Error; err != nil {
					t.Fatal(err)
				}
				references[category] = members
			}
			for _, mode := range []string{"count", "name", "default"} {
				caps := []int{12}
				if mode == "count" {
					caps = []int{0, 1, 7, 8, 12, 16, 17, 100}
				}
				order := []string{large, small, ""}
				if mode != "count" {
					order = []string{small, large, ""}
				}
				for _, cap := range caps {
					par.GroupLimit, par.CategorySort = cap, mode
					var expected []CategoryBrowseGroup
					for _, category := range order {
						members := references[category]
						if len(members) == 0 {
							continue
						}
						size := cap
						if size == 0 {
							size = len(members)
						}
						parts := (len(members) + size - 1) / size
						for part := 0; part < parts; part++ {
							count := min(size, len(members)-part*size)
							expected = append(expected, CategoryBrowseGroup{CategoryID: category, Part: part + 1, Parts: parts, Count: int64(count), Total: int64(len(members)), Preview: members[part*size : part*size+min(7, count)]})
						}
					}
					last := (len(expected) + par.Limit - 1) / par.Limit
					for _, page := range []int{1, 2, last, last + 1} {
						par.Page = page
						result, err := (CategoryBrowse{}).Summaries(db, par)
						if err != nil {
							t.Fatal(source, mode, cap, page, err)
						}
						start := min((page-1)*par.Limit, len(expected))
						want := expected[start:min(start+par.Limit, len(expected))]
						if result.Total != int64(len(expected)) || len(result.DataList) != len(want) {
							t.Fatal("page size/total mismatch", source, mode, cap, page)
						}
						for i, group := range result.DataList {
							ref := want[i]
							if group.CategoryID != ref.CategoryID || group.Part != ref.Part || group.Parts != ref.Parts || group.Count != ref.Count || group.Total != ref.Total || len(group.Preview) != len(ref.Preview) {
								t.Fatal("category/part mismatch", source, mode, cap, page, i)
							}
							for j, cover := range group.Preview {
								r := ref.Preview[j]
								if cover.ID != r.ID || cover.CoverPoster != r.CoverPoster || cover.CoverPosterWidth != r.CoverPosterWidth || cover.CoverPosterHeight != r.CoverPosterHeight || cover.Abstract != "" {
									t.Fatal("cover mismatch", source, mode, cap, page, i, j)
								}
							}
						}
					}
				}
			}
		}
	}
}

func TestCategoryBrowsePagingAndMembers(t *testing.T) {
	db := categoryBrowseTestDB(t)
	par := CategoryBrowseRequest{FilesBasesID: "lib", Source: "performer", GroupLimit: 4, Page: 1, Limit: 2, SearchData: datatype.ReqParam_SearchData{Sort: datatype.E_searchSort_random}}
	allIDs := []string{}
	for page := 1; page <= 3; page++ {
		par.Page = page
		result, err := (CategoryBrowse{}).Summaries(db, par)
		if err != nil {
			t.Fatal(err)
		}
		for _, g := range result.DataList {
			request := par
			request.CategoryID = g.CategoryID
			request.Part = g.Part
			members, total, err := (CategoryBrowse{}).Members(db, request)
			if err != nil {
				t.Fatal(err)
			}
			if len(*members) != int(g.Count) || total != g.Total {
				t.Fatal("part membership differs from summary")
			}
			if g.CategoryID == "a" {
				for _, resource := range *members {
					allIDs = append(allIDs, resource.ID)
				}
			}
		}
	}
	if len(allIDs) != 13 {
		t.Fatal(allIDs)
	}
	unique := map[string]bool{}
	for _, id := range allIDs {
		if unique[id] {
			t.Fatal("duplicate part resource", id)
		}
		unique[id] = true
	}
	par.Page = 1
	par.CategoryID = "a"
	par.Part = 0
	par.Limit = 5
	members, total, err := (CategoryBrowse{}).Members(db, par)
	if err != nil || len(*members) != 5 || total != 13 {
		t.Fatalf("all: %v %d %v", members, total, err)
	}
	par.Source = "tag:foreign"
	if _, err = (CategoryBrowse{}).Summaries(db, par); err == nil {
		t.Fatal("accepted cross-library tag class")
	}
	par.Source = "performer"
	par.GroupLimit = 101
	if _, err = (CategoryBrowse{}).Summaries(db, par); err == nil {
		t.Fatal("accepted oversized parts")
	}
}
func TestCategoryBrowseSettingsIsolation(t *testing.T) {
	db := categoryBrowseTestDB(t)
	if err := db.Create(&FilesBasesSetting{FilesBasesID: "lib", ConfigJsonData: `{"pageLimit":33,"mergeResourceGroups":false,"resourceGroupCoverStyle":"book"}`}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&LibraryConfigFollow{FilesBasesID: "lib", Module: "display"}).Error; err != nil {
		t.Fatal(err)
	}
	s := CategoryBrowseSettings{Enabled: true, Source: "performer", GroupLimit: 12, CoverStyle: "adaptive", ExpandStyle: "fold"}
	if err := SetCategoryBrowseSettings(db, "lib", s); err != nil {
		t.Fatal(err)
	}
	var setting FilesBasesSetting
	db.First(&setting, "filesBases_id = ?", "lib")
	var values map[string]interface{}
	json.Unmarshal([]byte(setting.ConfigJsonData), &values)
	if values["pageLimit"] != float64(33) || values["resourceGroupCoverStyle"] != "book" || values["categoryBrowseEnabled"] != true {
		t.Fatal(values)
	}
	for _, field := range SharedConfigFields("display") {
		if len(field) >= 14 && field[:14] == "categoryBrowse" {
			t.Fatal("category settings must stay independent")
		}
	}
	s.Enabled = false
	s.Source = "tag:deleted"
	if err := SetCategoryBrowseSettings(db, "lib", s); err != nil {
		t.Fatal("must allow disabling missing category", err)
	}
	s.CoverStyle = "stack|triptych"
	if err := SetCategoryBrowseSettings(db, "lib", s); err == nil {
		t.Fatal("accepted invalid combined style")
	}
}

func TestCategoryBrowseUnlimitedAndNewBounds(t *testing.T) {
	db := categoryBrowseTestDB(t)
	for i := 0; i < 105; i++ {
		id := fmt.Sprintf("extra%03d", i)
		if err := db.Exec(`INSERT INTO resources(id,filesBases_id,title) VALUES (?,'lib',?)`, id, id).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(`INSERT INTO resourcesPerformers(resources_id,performer_id) VALUES (?,'a')`, id).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, cap := range []int{0, 1, 100} {
		par := CategoryBrowseRequest{FilesBasesID: "lib", Source: "performer", GroupLimit: cap, Page: 1, Limit: 100}
		result, err := (CategoryBrowse{}).Summaries(db, par)
		if err != nil {
			t.Fatal(err)
		}
		var actor *CategoryBrowseGroup
		for i := range result.DataList {
			if result.DataList[i].CategoryID == "a" {
				actor = &result.DataList[i]
				break
			}
		}
		if actor == nil {
			t.Fatal("missing actor group")
		}
		expectedParts, expectedCount := 1, 118
		if cap > 0 {
			expectedParts = (118 + cap - 1) / cap
			expectedCount = min(cap, 118)
		}
		if actor.Parts != expectedParts || actor.Count != int64(expectedCount) || len(actor.Preview) != min(7, expectedCount) {
			t.Fatalf("cap %d: %+v", cap, actor)
		}
		par.CategoryID = "a"
		par.Part = 1
		members, total, err := (CategoryBrowse{}).Members(db, par)
		if err != nil || total != 118 || len(*members) != expectedCount {
			t.Fatalf("cap %d: total %d, err %v", cap, total, err)
		}
		if cap == 0 {
			par.Part = 0
			par.Page = 2
			par.Limit = 7
			members, total, err = (CategoryBrowse{}).Members(db, par)
			if err != nil || total != 118 || len(*members) != 7 {
				t.Fatal("unlimited category resource list must still paginate", err)
			}
			par.Part = 2
			if _, _, err = (CategoryBrowse{}).Members(db, par); err == nil {
				t.Fatal("unlimited category accepted a second part")
			}
		}
	}
	par := CategoryBrowseRequest{FilesBasesID: "lib", Source: "tag:publisher", GroupLimit: 0, Limit: 100}
	tags, err := (CategoryBrowse{}).Summaries(db, par)
	if err != nil || tags.Total != 3 {
		t.Fatalf("unlimited tags: %+v %v", tags, err)
	}
	for _, group := range tags.DataList {
		if group.Parts != 1 || group.Count != group.Total {
			t.Fatalf("unlimited category was partitioned: %+v", group)
		}
	}
	par.Part = 1
	unclassified, total, err := (CategoryBrowse{}).Members(db, par)
	if err != nil || total != 108 || len(*unclassified) != 108 {
		t.Fatal("unlimited unclassified members were truncated", err)
	}
	par.SearchData.SearchTextSlc = []string{"no-category-browse-match"}
	empty, err := (CategoryBrowse{}).Summaries(db, par)
	if err != nil || empty.Total != 0 || empty.ResourceTotal != 0 || len(empty.DataList) != 0 {
		t.Fatalf("unlimited empty filter failed: %+v %v", empty, err)
	}
	if err = db.Create(&FilesBasesSetting{FilesBasesID: "lib", ConfigJsonData: `{"pageLimit":33}`}).Error; err != nil {
		t.Fatal(err)
	}
	settings := CategoryBrowseSettings{Source: "performer", Enabled: true, CoverStyle: "stack", ExpandStyle: "fan"}
	for _, cap := range []int{0, 1, 100} {
		settings.GroupLimit = cap
		if err = SetCategoryBrowseSettings(db, "lib", settings); err != nil {
			t.Fatal(err)
		}
		var saved FilesBasesSetting
		db.First(&saved, "filesBases_id = ?", "lib")
		var fields map[string]interface{}
		json.Unmarshal([]byte(saved.ConfigJsonData), &fields)
		if fields["categoryBrowseGroupLimit"] != float64(cap) {
			t.Fatalf("cap was not preserved: %+v", fields)
		}
	}
	for _, cap := range []int{-1, 101} {
		settings.GroupLimit = cap
		if err = SetCategoryBrowseSettings(db, "lib", settings); err == nil {
			t.Fatal("invalid cap accepted", cap)
		}
	}
}
