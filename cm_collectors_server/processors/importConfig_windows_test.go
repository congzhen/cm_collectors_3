//go:build importstress

package processors

import (
	"bytes"
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/errorMessage"
	"cm_collectors_server/models"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

// Run alone: this integration test initializes process-global configuration.
func TestImportConfigCombinations(t *testing.T) {
	root := t.TempDir()
	cwd, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)
	if err := os.WriteFile("config.yaml", []byte("sqlite3:\n  loggerLevel: silent\nsystem:\n  logFilePath: NUL\n"), 0600); err != nil {
		t.Fatal(err)
	}
	core.Init()
	if err := models.DB_Init(core.DBS()); err != nil {
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
	var poster bytes.Buffer
	jpeg.Encode(&poster, image.NewRGBA(image.Rect(0, 0, 80, 60)), nil)
	base := datatype.Config_ScanDisk{ImportMode: datatype.ImportMode_append, CoverPosterType: -1, CoverPosterSuffixName: []string{"jpg"}}
	fixture := func(t *testing.T, name string) (string, string) {
		t.Helper()
		lib := core.GenerateUniqueID()
		if err := core.DBS().Create(&models.FilesBases{ID: lib, Name: lib}).Error; err != nil {
			t.Fatal(err)
		}
		dir := filepath.Join(root, lib)
		os.MkdirAll(dir, 0700)
		p := filepath.ToSlash(filepath.Join(dir, name+".mp4"))
		os.WriteFile(p, []byte("video"), 0600)
		return lib, p
	}
	run := func(t *testing.T, lib, p string, cfg datatype.Config_ScanDisk) []string {
		t.Helper()
		w, err := (ImportData{}).ScanDiskImportDataResult(context.Background(), lib, p, cfg)
		if err != nil {
			t.Fatal(err)
		}
		return w
	}
	lookup := func(t *testing.T, lib, p string) models.Resources {
		t.Helper()
		ds, err := (models.ResourcesDramaSeries{}).ListBySrcInLibrary(core.DBS(), lib, p)
		if err != nil || len(*ds) != 1 {
			t.Fatalf("path must have exactly one owner: %v %+v", err, ds)
		}
		var r models.Resources
		if err := core.DBS().First(&r, "id = ?", (*ds)[0].ResourcesID).Error; err != nil {
			t.Fatal(err)
		}
		return r
	}
	for _, mode := range []datatype.ResourceNamingMode{"fileName", "dirName", "dirFileName", "fullPathName"} {
		t.Run("naming/"+string(mode), func(t *testing.T) {
			lib, p := fixture(t, "movie")
			cfg := base
			cfg.ResourceNamingMode = mode
			run(t, lib, p, cfg)
			r := lookup(t, lib, p)
			want := "movie"
			switch mode {
			case "dirName":
				want = lib
			case "dirFileName":
				want = lib + "movie"
			case "fullPathName":
				want = p
			}
			if r.Title != want {
				t.Fatal(r.Title, want)
			}
		})
	}
	for _, tc := range []struct {
		name, image                    string
		names                          []string
		fuzzy, random, crop, bad, want bool
	}{
		{name: "exact", image: "movie", want: true}, {name: "configured", image: "poster", names: []string{"poster"}, want: true},
		{name: "filename-token", image: "movie", names: []string{"fileName"}, want: true}, {name: "regex", image: "movie-poster", names: []string{"regex:^@filename-poster$"}, want: true},
		{name: "fuzzy", image: "movie-poster", fuzzy: true, want: true}, {name: "strict-miss", image: "other"},
		{name: "random", image: "other", random: true, want: true}, {name: "crop", image: "movie", crop: true, want: true},
		{name: "invalid-crop", image: "movie", crop: true, bad: true},
	} {
		t.Run("poster/"+tc.name, func(t *testing.T) {
			lib, p := fixture(t, "movie")
			os.WriteFile(filepath.Join(filepath.Dir(p), tc.image+".jpg"), poster.Bytes(), 0600)
			cfg := base
			cfg.CoverPosterMatchName = tc.names
			cfg.CoverPosterFuzzyMatch = tc.fuzzy
			cfg.CoverPosterUseRandomImageIfNoMatch = tc.random
			if tc.crop {
				cfg.CoverPosterType = 0
				cfg.CoverPosterWidth = 40
				cfg.CoverPosterHeight = 60
				if tc.bad {
					cfg.CoverPosterWidth = 0
				}
			}
			w := run(t, lib, p, cfg)
			r := lookup(t, lib, p)
			if (r.CoverPoster != "") != tc.want || (len(w) > 0) != tc.bad {
				t.Fatalf("unexpected poster/warnings: %+v %v", r, w)
			}
		})
	}
	for _, enabled := range []bool{false, true} {
		name := "off"
		if enabled {
			name = "on"
		}
		t.Run("nfo/"+name, func(t *testing.T) {
			lib, p := fixture(t, "movie")
			os.WriteFile(filepath.Join(filepath.Dir(p), "movie.nfo"), []byte(`<movie><title lang="zh">NFO title</title><rating>85</rating></movie>`), 0600)
			cfg := base
			cfg.Nfo = datatype.Config_ScanDisk_Nfo{NfoStatus: enabled, Roots: []string{"movie"}, Titles: []string{"title"}, Score: []string{"rating"}}
			run(t, lib, p, cfg)
			r := lookup(t, lib, p)
			want := "movie"
			if enabled {
				want = "NFO title"
			}
			if r.Title != want || (enabled && r.Score != 8.5) {
				t.Fatal(r.Title, r.Score)
			}
		})
	}
	t.Run("append-and-cover-fallback", func(t *testing.T) {
		lib, p := fixture(t, "movie")
		photo := filepath.Join(filepath.Dir(p), "movie.jpg")
		os.WriteFile(photo, poster.Bytes(), 0600)
		run(t, lib, p, base)
		old := lookup(t, lib, p)
		core.DBS().Model(&old).Updates(map[string]any{"title": "manual", "subtitle": "subtitle", "definition": "4K"})
		w := run(t, lib, p, base)
		if len(w) != 1 || lookup(t, lib, p).Title != "manual" {
			t.Fatal("append overwrote existing resource")
		}
		os.WriteFile(photo, []byte("broken"), 0600)
		cfg := base
		cfg.ImportMode = datatype.ImportMode_cover
		w = run(t, lib, p, cfg)
		r := lookup(t, lib, p)
		if len(w) != 1 || r.CoverPoster != old.CoverPoster || r.CoverPosterWidth != 80 || r.Subtitle != "subtitle" || r.Definition != "4K" {
			t.Fatal("cover fallback lost metadata", r, w)
		}
	})
	t.Run("similar-name-second-group", func(t *testing.T) {
		lib, p := fixture(t, "Alpha 01")
		run(t, lib, p, base)
		b := filepath.ToSlash(filepath.Join(filepath.Dir(p), "Beta 01.mp4"))
		os.WriteFile(b, []byte("video"), 0600)
		run(t, lib, b, base)
		candidates, err := (ResourcesDramaSeries{}).FindDramaSeriesSlcBySearchPath(lib, filepath.ToSlash(filepath.Dir(p)))
		if err != nil {
			t.Fatal(err)
		}
		if len(*candidates) == 0 {
			t.Fatal("missing grouping candidates")
		}
		target := p
		if (*candidates)[0].ResourcesID == lookup(t, lib, p).ID {
			target = b
		}
		name := "Alpha 02"
		if target == b {
			name = "Beta 02"
		}
		next := filepath.ToSlash(filepath.Join(filepath.Dir(p), name+".mp4"))
		os.WriteFile(next, []byte("video"), 0600)
		cfg := base
		cfg.SimilarNameToSeries = true
		run(t, lib, next, cfg)
		if lookup(t, lib, next).ID != lookup(t, lib, target).ID {
			t.Fatal("similar-name grouping ignored second resource in directory")
		}
	})
	t.Run("cover-folder-existing-owner", func(t *testing.T) {
		lib, p := fixture(t, "Alpha")
		run(t, lib, p, base)
		b := filepath.ToSlash(filepath.Join(filepath.Dir(p), "Beta.mp4"))
		os.WriteFile(b, []byte("video"), 0600)
		run(t, lib, b, base)
		aID, bID := lookup(t, lib, p).ID, lookup(t, lib, b).ID
		cfg := base
		cfg.ImportMode = datatype.ImportMode_cover
		cfg.FolderToSeries = true
		run(t, lib, p, cfg)
		run(t, lib, b, cfg)
		if lookup(t, lib, p).ID != aID || lookup(t, lib, b).ID != bID {
			t.Fatal("cover changed existing ownership")
		}
	})
	for _, enabled := range []bool{false, true} {
		name := "disabled"
		if enabled {
			name = "enabled"
		}
		t.Run("nfo-create-flags/"+name, func(t *testing.T) {
			lib, _ := fixture(t, "movie")
			pb := core.GenerateUniqueID()
			if err := core.DBS().Create(&models.PerformerBases{ID: pb, Name: pb}).Error; err != nil {
				t.Fatal(err)
			}
			if err := core.DBS().Create(&models.FilesRelatedPerformerBases{ID: core.GenerateUniqueID(), FilesBasesID: lib, PerformerBasesID: pb, Main: true}).Error; err != nil {
				t.Fatal(err)
			}
			root := map[string]interface{}{"movie": map[string]interface{}{"genre": "new-tag", "actor": "new-actor"}}
			cfg := datatype.Config_ScanDisk_Nfo{Roots: []string{"movie"}, Tags: []string{"genre"}, TagAutoCreate: enabled, PerformerNames: []string{"actor"}, PerformerAutoCreate: enabled}
			par := datatype.ReqParam_Resource{}
			if err := (ImportData{}).NfoExecData(lib, "", root, cfg, &par); err != nil {
				t.Fatal(err)
			}
			if (len(par.Tags) > 0) != enabled || (len(par.Performers) > 0) != enabled {
				t.Fatalf("creation flags ignored: tags=%v performers=%v", par.Tags, par.Performers)
			}
			if enabled {
				cfg.TagAutoCreate = false
				cfg.PerformerAutoCreate = false
				par = datatype.ReqParam_Resource{}
				(ImportData{}).NfoExecData(lib, "", root, cfg, &par)
				if len(par.Tags) != 1 || len(par.Performers) != 1 {
					t.Fatal("disabled creation must still match existing entries")
				}
			}
		})
		t.Run("nfo-alias/"+name, func(t *testing.T) {
			lib, _ := fixture(t, "movie")
			pb := core.GenerateUniqueID()
			actor := models.Performer{ID: core.GenerateUniqueID(), PerformerBasesID: pb, Name: "real-name", AliasName: "alias"}
			if err := core.DBS().Create(&models.PerformerBases{ID: pb, Name: pb}).Error; err != nil {
				t.Fatal(err)
			}
			if err := core.DBS().Create(&models.FilesRelatedPerformerBases{ID: core.GenerateUniqueID(), FilesBasesID: lib, PerformerBasesID: pb, Main: true}).Error; err != nil {
				t.Fatal(err)
			}
			if err := core.DBS().Create(&actor).Error; err != nil {
				t.Fatal(err)
			}
			root := map[string]interface{}{"movie": map[string]interface{}{"actor": "alias"}}
			par := datatype.ReqParam_Resource{}
			cfg := datatype.Config_ScanDisk_Nfo{Roots: []string{"movie"}, PerformerNames: []string{"actor"}, PerformerMatchAliasName: enabled}
			if err := (ImportData{}).NfoExecData(lib, "", root, cfg, &par); err != nil {
				t.Fatal(err)
			}
			if (len(par.Performers) > 0) != enabled {
				t.Fatal("alias flag ignored", par.Performers)
			}
		})
	}
	// Cross-check the changed branches against their original callers and policies.
	for _, matchAlias := range []bool{false, true} {
		for _, autoCreate := range []bool{false, true} {
			name := fmt.Sprintf("actor-libraries/alias=%t,create=%t", matchAlias, autoCreate)
			t.Run(name, func(t *testing.T) {
				lib, _ := fixture(t, "movie")
				main, related := core.GenerateUniqueID(), core.GenerateUniqueID()
				for _, id := range []string{related, main} {
					if err := core.DBS().Create(&models.PerformerBases{ID: id, Name: id}).Error; err != nil {
						t.Fatal(err)
					}
					if err := core.DBS().Create(&models.FilesRelatedPerformerBases{ID: core.GenerateUniqueID(), FilesBasesID: lib, PerformerBasesID: id, Main: id == main}).Error; err != nil {
						t.Fatal(err)
					}
				}
				mainActor := models.Performer{ID: core.GenerateUniqueID(), PerformerBasesID: main, Name: "shared", AliasName: "main-alias"}
				relatedActor := models.Performer{ID: core.GenerateUniqueID(), PerformerBasesID: related, Name: "related-only", AliasName: "related-alias"}
				for _, actor := range []models.Performer{mainActor, relatedActor, {ID: core.GenerateUniqueID(), PerformerBasesID: related, Name: "shared"}} {
					if err := core.DBS().Create(&actor).Error; err != nil {
						t.Fatal(err)
					}
				}
				service := Performer{}
				for _, pair := range []struct{ name, id string }{{"shared", mainActor.ID}, {"related-only", relatedActor.ID}} {
					actor, err := service.performerInfoForImport(lib, pair.name, "", matchAlias, autoCreate)
					if err != nil || actor.ID != pair.id {
						t.Fatalf("library lookup changed: %v %+v", err, actor)
					}
				}
				actor, err := service.performerInfoForImport(lib, "related-alias", "", matchAlias, autoCreate)
				if matchAlias {
					if err != nil || actor.ID != relatedActor.ID {
						t.Fatal("related alias not matched", err)
					}
				} else if autoCreate {
					if err != nil || actor.PerformerBasesID != main || actor.Name != "related-alias" {
						t.Fatal("new actor not created in main library", err)
					}
				} else if err != errorMessage.Err_performer_Not_Found {
					t.Fatal("disabled alias/create returned unexpected result", actor, err)
				}
				// Existing helper retains automatic creation and alias matching.
				actor, err = service.PerformerInfoByNameNotFoundCreate(lib, "main-alias", "")
				if err != nil || actor.ID != mainActor.ID {
					t.Fatal("legacy alias behavior changed", err)
				}
				actor, err = service.PerformerInfoByNameNotFoundCreate(lib, "legacy-new", "")
				if err != nil || actor.PerformerBasesID != main {
					t.Fatal("legacy create behavior changed", err)
				}
				var count int64
				if err := core.DBS().Model(&models.Performer{}).Where("performerBases_id IN ?", []string{main, related}).Count(&count).Error; err != nil {
					t.Fatal(err)
				}
				want := int64(4)
				if !matchAlias && autoCreate {
					want++
				}
				if count != want {
					t.Fatal("unexpected duplicate actor creation", count, want)
				}
			})
		}
	}
	t.Run("actor-without-main-library", func(t *testing.T) {
		lib, _ := fixture(t, "movie")
		if _, err := (Performer{}).performerInfoForImport(lib, "unknown", "", true, true); err != errorMessage.Err_performer_Not_Found {
			t.Fatal(err)
		}
	})
	for _, policy := range []string{"folder-priority", "simple-episode", "different-directory"} {
		t.Run("group-isolation/"+policy, func(t *testing.T) {
			lib, p := fixture(t, "Other 01")
			run(t, lib, p, base)
			b := filepath.ToSlash(filepath.Join(filepath.Dir(p), "01.mp4"))
			if err := os.WriteFile(b, []byte("video"), 0600); err != nil {
				t.Fatal(err)
			}
			run(t, lib, b, base)
			cfg := base
			cfg.SimilarNameToSeries = true
			next := filepath.ToSlash(filepath.Join(filepath.Dir(p), "02.mp4"))
			wantID := lookup(t, lib, b).ID
			if policy == "folder-priority" {
				cfg.FolderToSeries = true
				groups, err := (ResourcesDramaSeries{}).FindDramaSeriesSlcBySearchPath(lib, filepath.ToSlash(filepath.Dir(p)))
				if err != nil || len(*groups) == 0 {
					t.Fatal(err)
				}
				wantID = (*groups)[0].ResourcesID
				// Force a name that would select the other group in similar-only mode.
				if wantID == lookup(t, lib, b).ID {
					next = filepath.ToSlash(filepath.Join(filepath.Dir(p), "Other 02.mp4"))
				}
			} else if policy == "different-directory" {
				dir := filepath.Join(filepath.Dir(p), "subdir")
				if err := os.MkdirAll(dir, 0700); err != nil {
					t.Fatal(err)
				}
				next = filepath.ToSlash(filepath.Join(dir, "02.mp4"))
			}
			if err := os.WriteFile(next, []byte("video"), 0600); err != nil {
				t.Fatal(err)
			}
			run(t, lib, next, cfg)
			got := lookup(t, lib, next).ID
			if policy == "different-directory" {
				if got == wantID || got == lookup(t, lib, p).ID {
					t.Fatal("cross-directory merge")
				}
			} else if got != wantID {
				t.Fatal("group policy changed", got, wantID)
			}
		})
	}
	t.Run("cover-existing-updates-nfo-and-poster", func(t *testing.T) {
		lib, p := fixture(t, "Show 01")
		cfg := base
		cfg.FolderToSeries = true
		run(t, lib, p, cfg)
		next := filepath.ToSlash(filepath.Join(filepath.Dir(p), "Show 02.mp4"))
		if err := os.WriteFile(next, []byte("video"), 0600); err != nil {
			t.Fatal(err)
		}
		run(t, lib, next, cfg)
		original := lookup(t, lib, p)
		if err := os.WriteFile(filepath.Join(filepath.Dir(p), "Show 01.nfo"), []byte(`<movie><title>Updated title</title><genre>existing-tag</genre><actor>existing-actor</actor></movie>`), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(filepath.Dir(p), "Show 01.jpg"), poster.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
		cfg.ImportMode = datatype.ImportMode_cover
		cfg.SimilarNameToSeries = true
		cfg.Nfo = datatype.Config_ScanDisk_Nfo{NfoStatus: true, Roots: []string{"movie"}, Titles: []string{"title"}, Tags: []string{"genre"}, PerformerNames: []string{"actor"}}
		tag, err := (Tag{}).TagInfoByNameNotFoundCreate(lib, "existing-tag")
		if err != nil {
			t.Fatal(err)
		}
		pb := core.GenerateUniqueID()
		actor := models.Performer{ID: core.GenerateUniqueID(), PerformerBasesID: pb, Name: "existing-actor"}
		for _, row := range []any{&models.PerformerBases{ID: pb, Name: pb}, &models.FilesRelatedPerformerBases{ID: core.GenerateUniqueID(), FilesBasesID: lib, PerformerBasesID: pb, Main: true}, &actor} {
			if err := core.DBS().Create(row).Error; err != nil {
				t.Fatal(err)
			}
		}
		run(t, lib, p, cfg)
		updated := lookup(t, lib, p)
		if updated.ID != original.ID || lookup(t, lib, next).ID != original.ID || updated.Title != "Updated title" || updated.CoverPoster == "" || updated.CoverPosterWidth != 80 {
			t.Fatal("cover skipped metadata or changed ownership", updated)
		}
		info, err := (Resources{}).Info(updated.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(info.ResourcesDramaSeries) != 2 || len(info.Tags) != 1 || info.Tags[0].ID != tag.ID || len(info.Performers) != 1 || info.Performers[0].ID != actor.ID {
			t.Fatal("cover lost episodes or existing associations", info)
		}
	})
	for _, mode := range []datatype.SeriesSortMode{"keep", "nameAsc", "nameDesc", "sizeAsc", "sizeDesc", "legacy"} {
		t.Run("series-sort/"+string(mode), func(t *testing.T) {
			lib, p := fixture(t, "Show 02")
			os.WriteFile(p, bytes.Repeat([]byte("a"), 100), 0600)
			cfg := base
			cfg.FolderToSeries = true
			cfg.FolderToSeriesSortMode = mode
			if mode == "legacy" {
				cfg.FolderToSeriesSortMode = ""
				cfg.FolderToSeriesSort = true
			}
			run(t, lib, p, cfg)
			next := filepath.ToSlash(filepath.Join(filepath.Dir(p), "Show 01.mp4"))
			os.WriteFile(next, []byte("a"), 0600)
			run(t, lib, next, cfg)
			r := lookup(t, lib, p)
			var ds []models.ResourcesDramaSeries
			if err := core.DBS().Where("resources_id = ?", r.ID).Order("sort ASC").Find(&ds).Error; err != nil {
				t.Fatal(err)
			}
			want := p
			if mode == "nameAsc" || mode == "sizeAsc" || mode == "legacy" {
				want = next
			}
			if len(ds) != 2 || ds[0].Src != want {
				t.Fatal("unexpected episode order", ds)
			}
		})
	}
	for _, mode := range []string{"strict", "fuzzy", "fallback"} {
		t.Run("nfo-match/"+mode, func(t *testing.T) {
			lib, p := fixture(t, "movie")
			name := "movie-extra"
			if mode == "fallback" {
				name = "unrelated"
			}
			os.WriteFile(filepath.Join(filepath.Dir(p), name+".nfo"), []byte(`<movie><title>NFO match</title></movie>`), 0600)
			cfg := base
			cfg.Nfo = datatype.Config_ScanDisk_Nfo{NfoStatus: true, Roots: []string{"movie"}, Titles: []string{"title"}}
			cfg.EnableNfoFuzzyMatch = mode == "fuzzy"
			cfg.UseRandomNfoIfNoneMatch = mode == "fallback"
			run(t, lib, p, cfg)
			want := "NFO match"
			if mode == "strict" {
				want = "movie"
			}
			if lookup(t, lib, p).Title != want {
				t.Fatal("NFO matching option ignored")
			}
		})
	}
	t.Run("probe-failure-preserves-definition", func(t *testing.T) {
		lib, p := fixture(t, "movie")
		run(t, lib, p, base)
		r := lookup(t, lib, p)
		core.DBS().Model(&r).Update("definition", "4K")
		cfg := base
		cfg.ImportMode = datatype.ImportMode_cover
		cfg.AutoGetVideoDefinition = true
		w := run(t, lib, p, cfg)
		if len(w) == 0 || lookup(t, lib, p).Definition != "4K" {
			t.Fatal("probe failure cleared definition", w)
		}
	})
	for _, mode := range []datatype.ImportMode{datatype.ImportMode_append, datatype.ImportMode_cover} {
		for _, group := range []string{"none", "folder", "similar", "both"} {
			t.Run(string(mode)+"/"+group, func(t *testing.T) {
				lib, p := fixture(t, "Show 01")
				cfg := base
				cfg.ImportMode = mode
				cfg.FolderToSeries = group == "folder" || group == "both"
				cfg.SimilarNameToSeries = group == "similar" || group == "both"
				run(t, lib, p, cfg)
				next := filepath.ToSlash(filepath.Join(filepath.Dir(p), "Show 02.mp4"))
				os.WriteFile(next, []byte("video"), 0600)
				run(t, lib, next, cfg)
				if (lookup(t, lib, p).ID == lookup(t, lib, next).ID) != (group != "none") {
					t.Fatal("incorrect grouping")
				}
			})
		}
	}
}
