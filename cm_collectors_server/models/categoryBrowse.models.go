package models

import (
	"cm_collectors_server/datatype"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CategoryBrowseRequest struct {
	FilesBasesID string                       `json:"filesBasesId" binding:"required,max=20"`
	Source       string                       `json:"source" binding:"required,max=24"`
	GroupLimit   int                          `json:"groupLimit"`
	CategorySort string                       `json:"categorySort"`
	Page         int                          `json:"page"`
	Limit        int                          `json:"limit"`
	SearchData   datatype.ReqParam_SearchData `json:"searchData"`
	CategoryID   string                       `json:"categoryId" binding:"max=20"`
	Part         int                          `json:"part"` // 0 查看整类，正整数查看对应小组
}
type CategoryBrowseGroup struct {
	CategoryID string      `json:"categoryId"`
	Name       string      `json:"name"`
	Photo      string      `json:"photo"`
	Part       int         `json:"part"`
	Parts      int         `json:"parts"`
	Count      int64       `json:"count"`
	Total      int64       `json:"total"`
	Preview    []Resources `json:"preview"`
}
type CategoryBrowsePage struct {
	DataList      []CategoryBrowseGroup `json:"dataList"`
	Total         int64                 `json:"total"`         // 展示组数量
	ResourceTotal int64                 `json:"resourceTotal"` // 去重后的筛选命中资源数量
	CategoryTotal int                   `json:"categoryTotal"`
}
type categoryBrowseCount struct {
	ID               string
	Name             string
	Photo            string
	Count            int64
	PerformerBasesID string
}
type CategoryBrowse struct{}

func categoryBrowseValidate(db *gorm.DB, par *CategoryBrowseRequest) error {
	if par.CategorySort == "" {
		par.CategorySort = "count"
	}
	switch par.CategorySort {
	case "count", "name", "default":
	default:
		return errors.New("分类排序方式无效")
	}
	if par.FilesBasesID == "" {
		return errors.New("请选择文件库")
	}
	if par.GroupLimit < 0 || par.GroupLimit > 100 {
		return errors.New("每组资源上限必须在 0 到 100 之间，0 表示不限制")
	}
	if par.GroupLimit == 0 && par.Part > 1 {
		return errors.New("不限量分类只有一个展示组")
	}
	if par.Page < 1 {
		par.Page = 1
	}
	if par.Page > 1000000 {
		return errors.New("页码超出范围")
	}
	if par.Limit < 1 {
		par.Limit = 24
	}
	if par.Limit > 100 {
		par.Limit = 100
	}
	if par.Part < 0 || par.Part > 1000000 {
		return errors.New("分组序号无效")
	}
	if par.Source == "performer" {
		return nil
	}
	if !strings.HasPrefix(par.Source, "tag:") {
		return errors.New("不支持的分类来源")
	}
	var count int64
	if err := db.Model(&TagClass{}).Where("id = ? AND filesBases_id = ? AND status = ?", strings.TrimPrefix(par.Source, "tag:"), par.FilesBasesID, true).Count(&count).Error; err != nil {
		return err
	}
	if count != 1 {
		return errors.New("所选标签分类已删除、禁用或不属于当前文件库，请重新选择")
	}
	return nil
}
func categoryBrowseBase(db *gorm.DB, par *CategoryBrowseRequest) *gorm.DB {
	return (Resources{}).setDbSearchData(db.Model(&Resources{}).Where("filesBases_id = ?", par.FilesBasesID), &par.SearchData)
}
func categoryBrowseMembership(par *CategoryBrowseRequest) (string, []interface{}) {
	if par.Source == "performer" {
		return "SELECT 1 FROM resourcesPerformers cb_link JOIN performer cb_category ON cb_category.id = cb_link.performer_id WHERE cb_link.resources_id = resources.id AND cb_category.status = ?", []interface{}{true}
	}
	return "SELECT 1 FROM resourcesTags cb_link JOIN tag cb_category ON cb_category.id = cb_link.tag_id WHERE cb_link.resources_id = resources.id AND cb_category.status = ? AND cb_category.tagClass_id = ?", []interface{}{true, strings.TrimPrefix(par.Source, "tag:")}
}
func categoryBrowseMembers(db *gorm.DB, par *CategoryBrowseRequest, category string) *gorm.DB {
	query := categoryBrowseBase(db, par)
	membership, args := categoryBrowseMembership(par)
	if category == "" {
		return query.Where("NOT EXISTS ("+membership+")", args...)
	}
	args = append(args, category)
	return query.Where("EXISTS ("+membership+" AND cb_category.id = ?)", args...)
}
func categoryBrowseOrder(db *gorm.DB, par *CategoryBrowseRequest) *gorm.DB {
	order := par.SearchData.Sort
	if order == datatype.E_searchSort_random || order == "" {
		order = datatype.E_searchSort_addTimeDesc
	}
	return (Resources{}).setDbSearchDataOrder(db, order, par.FilesBasesID).Order("resources.id ASC")
}

// Aggregate library totals using SQL candidate subqueries instead of binding all
// category IDs or repeatedly querying batches. Only the ranking is unfiltered.
func categoryBrowseRankCounts(db *gorm.DB, par *CategoryBrowseRequest) (map[string]int64, error) {
	ids := categoryBrowseBase(db, par).Select("resources.id")
	type rankCount struct {
		ID    string
		Count int64
	}
	var rows []rankCount
	if par.Source == "performer" {
		candidates := db.Table("resourcesPerformers").Select("performer_id").Where("resources_id IN (?)", ids)
		links := func(table string) *gorm.DB {
			return db.Table(table+" l").Select("l.performer_id, l.resources_id").
				Joins("JOIN resources r ON r.id = l.resources_id").
				Joins("JOIN performer p ON p.id = l.performer_id").
				Where("r.status = ? AND l.performer_id IN (?)", true, candidates).
				Where("EXISTS (SELECT 1 FROM filesRelatedPerformerBases f WHERE f.filesBases_id = r.filesBases_id AND f.performerBases_id = p.performerBases_id)")
		}
		if err := db.Raw("SELECT performer_id AS id, COUNT(DISTINCT resources_id) AS count FROM (? UNION ALL ?) links GROUP BY performer_id", links("resourcesPerformers"), links("resourcesDirectors")).Scan(&rows).Error; err != nil {
			return nil, err
		}
	} else {
		if err := db.Table("resourcesTags l").Select("l.tag_id AS id, COUNT(DISTINCT l.resources_id) AS count").
			Joins("JOIN resources r ON r.id = l.resources_id").
			Joins("JOIN tag t ON t.id = l.tag_id").
			Where("r.status = ? AND t.tagClass_id = ? AND t.status = ?", true, strings.TrimPrefix(par.Source, "tag:"), true).
			Group("l.tag_id").Scan(&rows).Error; err != nil {
			return nil, err
		}
	}
	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[row.ID] = row.Count
	}
	return counts, nil
}

func categoryBrowsePreviews(db *gorm.DB, par *CategoryBrowseRequest, groups []CategoryBrowseGroup) error {
	for first := 0; first < len(groups); {
		end := first + 1
		// Merge adjacent small parts, avoiding repeated filtering/sorting of the
		// same category. Bound temporary cover rows and skip large-part gaps.
		if par.GroupLimit > 0 && par.GroupLimit <= 16 {
			for end < len(groups) && groups[end].CategoryID == groups[first].CategoryID && end-first < 256/par.GroupLimit {
				end++
			}
		}
		last := groups[end-1]
		span := (last.Part-groups[first].Part)*par.GroupLimit + min(7, int(last.Count))
		covers := []Resources{}
		q := categoryBrowseOrder(categoryBrowseMembers(db, par, groups[first].CategoryID), par)
		if err := q.Select("resources.id, resources.filesBases_id, resources.title, resources.subtitle, resources.coverPoster, resources.coverPosterWidth, resources.coverPosterHeight").
			Offset((groups[first].Part - 1) * par.GroupLimit).Limit(span).Find(&covers).Error; err != nil {
			return err
		}
		for i := first; i < end; i++ {
			start := min((groups[i].Part-groups[first].Part)*par.GroupLimit, len(covers))
			stop := min(start+min(7, int(groups[i].Count)), len(covers))
			groups[i].Preview = append([]Resources{}, covers[start:stop]...)
		}
		first = end
	}
	return nil
}

func (CategoryBrowse) Summaries(db *gorm.DB, par CategoryBrowseRequest) (*CategoryBrowsePage, error) {
	result := &CategoryBrowsePage{DataList: []CategoryBrowseGroup{}}
	if err := categoryBrowseValidate(db, &par); err != nil {
		return nil, err
	}
	if err := categoryBrowseBase(db, &par).Count(&result.ResourceTotal).Error; err != nil {
		return nil, err
	}
	if result.ResourceTotal == 0 {
		return result, nil
	}
	ids := categoryBrowseBase(db, &par).Select("resources.id")
	var categories []categoryBrowseCount
	var query *gorm.DB
	groupBy := "c.id, c.name"
	if par.Source == "performer" {
		query = db.Table("performer c").Select("c.id, c.name, c.photo, c.performerBases_id AS performer_bases_id, COUNT(DISTINCT l.resources_id) AS count").Joins("JOIN resourcesPerformers l ON l.performer_id = c.id").Where("c.status = ?", true)
		groupBy = "c.id, c.name, c.photo, c.performerBases_id"
	} else {
		query = db.Table("tag c").Select("c.id, c.name, '' AS photo, COUNT(DISTINCT l.resources_id) AS count").Joins("JOIN resourcesTags l ON l.tag_id = c.id").Where("c.status = ? AND c.tagClass_id = ?", true, strings.TrimPrefix(par.Source, "tag:"))
	}
	order := "c.addTime DESC, c.id DESC"
	switch par.CategorySort {
	case "name":
		order = "c.name ASC, c.id ASC"
	case "default":
		order = "c.addTime DESC, c.id DESC"
	}
	if err := query.Where("l.resources_id IN (?)", ids).Group(groupBy + ", c.addTime").Order(order).Scan(&categories).Error; err != nil {
		return nil, err
	}
	if par.CategorySort == "count" && len(categories) > 0 {
		// Rank by the same lifetime totals used by the performer/tag libraries.
		// Category.Count remains the filtered member count for cards and splitting.
		counts, err := categoryBrowseRankCounts(db, &par)
		if err != nil {
			return nil, err
		}
		sort.SliceStable(categories, func(i, j int) bool {
			return counts[categories[i].ID] > counts[categories[j].ID]
		})
	}
	var unclassified int64
	if err := categoryBrowseMembers(db, &par, "").Count(&unclassified).Error; err != nil {
		return nil, err
	}
	if unclassified > 0 {
		name := "未关联演员"
		if par.Source != "performer" {
			name = "未设置分类标签"
		}
		categories = append(categories, categoryBrowseCount{Name: name, Count: unclassified})
	}
	result.CategoryTotal = len(categories)
	start, end := int64(par.Page-1)*int64(par.Limit), int64(par.Page)*int64(par.Limit)
	var offset int64
	for _, category := range categories {
		if category.Photo != "" {
			category.Photo = fmt.Sprintf("/api/performerFace/%s/%s", category.PerformerBasesID, category.Photo)
		}
		groupSize := int64(par.GroupLimit)
		if groupSize == 0 {
			groupSize = category.Count
		}
		parts := (category.Count + groupSize - 1) / groupSize
		for part := max(int64(0), start-offset); part < parts && offset+part < end; part++ {
			count := min(groupSize, category.Count-part*groupSize)
			group := CategoryBrowseGroup{CategoryID: category.ID, Name: category.Name, Photo: category.Photo, Part: int(part + 1), Parts: int(parts), Count: count, Total: category.Count, Preview: []Resources{}}
			result.DataList = append(result.DataList, group)
		}
		offset += parts
	}
	result.Total = offset
	if err := categoryBrowsePreviews(db, &par, result.DataList); err != nil {
		return nil, err
	}
	return result, nil
}
func (CategoryBrowse) Members(db *gorm.DB, par CategoryBrowseRequest) (*[]Resources, int64, error) {
	if err := categoryBrowseValidate(db, &par); err != nil {
		return nil, 0, err
	}
	query := categoryBrowseMembers(db, &par, par.CategoryID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset, limit := (par.Page-1)*par.Limit, par.Limit
	if par.Part > 0 {
		offset = (par.Part - 1) * par.GroupLimit
		limit = par.GroupLimit
		if par.GroupLimit == 0 {
			limit = -1 // 不限量分类展开整个分类；part=0 的资源列表仍分页。
		}
	}
	data := []Resources{}
	q := categoryBrowseOrder(categoryBrowseMembers(db, &par, par.CategoryID), &par)
	if err := (Resources{}).Preload(q).Offset(offset).Limit(limit).Find(&data).Error; err != nil {
		return nil, 0, err
	}
	for i := range data {
		if data[i].Tags == nil {
			data[i].Tags = []Tag{}
		}
		if data[i].Performers == nil {
			data[i].Performers = []Performer{}
		}
		if data[i].Directors == nil {
			data[i].Directors = []Performer{}
		}
		if data[i].ResourcesDramaSeries == nil {
			data[i].ResourcesDramaSeries = []ResourcesDramaSeries{}
		}
	}
	if err := (Performer{}).fillResourceCountsForResources(db, data); err != nil {
		return nil, 0, err
	}
	return &data, total, nil
}

type CategoryBrowseSettings struct {
	Enabled         bool   `json:"categoryBrowseEnabled"`
	Source          string `json:"categoryBrowseSource"`
	GroupLimit      int    `json:"categoryBrowseGroupLimit"`
	Sort            string `json:"categoryBrowseSort"`
	CoverStyle      string `json:"categoryBrowseCoverStyle"`
	HideSingle bool `json:"categoryBrowseHideSingle"`
	AdaptCoverRatio bool   `json:"categoryBrowseAdaptCoverRatio"`
	ExpandStyle     string `json:"categoryBrowseExpandStyle"`
}

func SetCategoryBrowseSettings(db *gorm.DB, library string, settings CategoryBrowseSettings) error {
	if settings.Source == "" {
		settings.Source = "performer"
	}
	validationSource := settings.Source
	if !settings.Enabled {
		validationSource = "performer"
	}
	if settings.Sort == "" {
		settings.Sort = "count"
	}
	if err := categoryBrowseValidate(db, &CategoryBrowseRequest{FilesBasesID: library, Source: validationSource, GroupLimit: settings.GroupLimit, CategorySort: settings.Sort}); err != nil {
		return err
	}
	if settings.CoverStyle == "" {
		settings.CoverStyle = "stack"
	}
	if settings.ExpandStyle == "" {
		settings.ExpandStyle = "fan"
	}
	switch settings.CoverStyle {
	case "stack", "triptych", "book", "adaptive":
	default:
		return errors.New("分类浏览封面样式无效")
	}
	switch settings.ExpandStyle {
	case "fan", "strip", "fold", "spine", "magazine":
	default:
		return errors.New("分类浏览展开样式无效")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var setting FilesBasesSetting
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("filesBases_id = ?", library).First(&setting).Error; err != nil {
			return err
		}
		config, err := parseConfig(setting.ConfigJsonData)
		if err != nil {
			return err
		}
		raw, _ := json.Marshal(settings)
		var fields map[string]interface{}
		if err := json.Unmarshal(raw, &fields); err != nil {
			return err
		}
		for key, value := range fields {
			config[key] = value
		}
		raw, err = json.Marshal(config)
		if err != nil {
			return err
		}
		return tx.Model(&setting).Update("config_json_data", string(raw)).Error
	})
}
