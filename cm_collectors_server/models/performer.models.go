package models

import (
	"cm_collectors_server/datatype"
	"sort"
	"strings"

	"gorm.io/gorm"
)

type Performer struct {
	ID                    string               `json:"id" gorm:"primaryKey;type:char(20);"`
	PerformerBasesID      string               `json:"performerBases_id" gorm:"column:performerBases_id;type:char(20);;index:idx_performer_performerBasesID;"`
	Name                  string               `json:"name" gorm:"type:varchar(200);"`
	AliasName             string               `json:"aliasName" gorm:"column:aliasName;type:varchar(500);"`
	KeyWords              string               `json:"keyWords" gorm:"column:keyWords;type:varchar(500);"`
	Birthday              string               `json:"birthday" gorm:"column:birthday;type:varchar(10);"`
	Nationality           string               `json:"nationality" gorm:"type:varchar(200);"`
	CareerPerformer       *bool                `json:"careerPerformer" gorm:"column:careerPerformer;type:tinyint(1);default:1"`
	CareerDirector        *bool                `json:"careerDirector" gorm:"column:careerDirector;type:tinyint(1);default:0"`
	Photo                 string               `json:"photo" gorm:"type:varchar(100);"`
	Introduction          string               `json:"introduction" gorm:"type:text;"`
	Cup                   string               `json:"cup" gorm:"type:varchar(10);index:idx_performer_cup;"`
	Bust                  string               `json:"bust" gorm:"type:varchar(10);"`
	Waist                 string               `json:"waist" gorm:"type:varchar(10);"`
	Hip                   string               `json:"hip" gorm:"type:varchar(10);"`
	Stars                 int                  `json:"stars" gorm:"type:int;"`
	RetreatStatus         bool                 `json:"retreatStatus" gorm:"column:retreatStatus;type:tinyint(1);default:0"`
	CreatedAt             *datatype.CustomTime `json:"-" gorm:"column:addTime;type:datetime"`
	LastScraperUpdateTime *datatype.CustomDate `json:"lastScraperUpdateTime" gorm:"column:lastScraperUpdateTime;type:date;default:NULL"`
	Status                bool                 `json:"status" gorm:"type:tinyint(1);default:1"`
	ResourceCount         int64                `json:"resourceCount" gorm:"column:resourceCount;->;-:migration"`
}

type PerformerBasic struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AliasName string `json:"aliasName"  gorm:"column:aliasName;"`
	KeyWords  string `json:"keyWords" gorm:"column:keyWords;"`
}

func (Performer) TableName() string {
	return "performer"
}
func (PerformerBasic) TableName() string {
	return "performer"
}

type performerResourceCount struct {
	PerformerID   string `gorm:"column:performer_id"`
	ResourceCount int64  `gorm:"column:resourceCount"`
}

// resourceCountMap 批量统计一组演员在指定文件库中的资源数量。
//
// 这里不在 performer 列表查询中直接写逐行子查询，是为了避免一页演员触发多次
// COUNT/EXISTS。先拿到当前页的 performerIds，再用一次 UNION + GROUP BY 统一统计，
// 可以把“每个演员一次计数查询”压缩成“一批演员一次计数查询”。
//
// 统计规则：
// 1. 只统计 status = 1 的资源。
// 2. 演员表 resourcesPerformers 和导演表 resourcesDirectors 都计入。
// 3. 同一个资源如果同时出现在演员和导演关联里，只算一次。
// 4. filesBasesId 非空时，严格限定 resources.filesBases_id = filesBasesId。
// 5. filesBasesId 为空时，按演员所属 performerBases_id 关联的文件库范围统计。
func (Performer) resourceCountMap(db *gorm.DB, performerIds []string, filesBasesId string) (map[string]int64, error) {
	countMap := make(map[string]int64, len(performerIds))
	if len(performerIds) == 0 {
		return countMap, nil
	}

	var list []performerResourceCount
	if filesBasesId != "" {
		err := db.Raw(`
			SELECT performer_id, COUNT(DISTINCT resources_id) AS resourceCount
			FROM (
				SELECT resourcesPerformers.performer_id, resourcesPerformers.resources_id
				FROM resourcesPerformers
				INNER JOIN resources ON resources.id = resourcesPerformers.resources_id
				WHERE resources.status = 1
					AND resources.filesBases_id = ?
					AND resourcesPerformers.performer_id IN ?
				UNION
				SELECT resourcesDirectors.performer_id, resourcesDirectors.resources_id
				FROM resourcesDirectors
				INNER JOIN resources ON resources.id = resourcesDirectors.resources_id
				WHERE resources.status = 1
					AND resources.filesBases_id = ?
					AND resourcesDirectors.performer_id IN ?
			) AS performerResources
			GROUP BY performer_id
		`, filesBasesId, performerIds, filesBasesId, performerIds).Scan(&list).Error
		if err != nil {
			return countMap, err
		}
	} else {
		err := db.Raw(`
			SELECT performer_id, COUNT(DISTINCT resources_id) AS resourceCount
			FROM (
				SELECT resourcesPerformers.performer_id, resourcesPerformers.resources_id
				FROM resourcesPerformers
				INNER JOIN resources ON resources.id = resourcesPerformers.resources_id
				INNER JOIN performer ON performer.id = resourcesPerformers.performer_id
				WHERE resources.status = 1
					AND resourcesPerformers.performer_id IN ?
					AND EXISTS (
						SELECT 1 FROM filesRelatedPerformerBases
						WHERE filesRelatedPerformerBases.filesBases_id = resources.filesBases_id
							AND filesRelatedPerformerBases.performerBases_id = performer.performerBases_id
					)
				UNION
				SELECT resourcesDirectors.performer_id, resourcesDirectors.resources_id
				FROM resourcesDirectors
				INNER JOIN resources ON resources.id = resourcesDirectors.resources_id
				INNER JOIN performer ON performer.id = resourcesDirectors.performer_id
				WHERE resources.status = 1
					AND resourcesDirectors.performer_id IN ?
					AND EXISTS (
						SELECT 1 FROM filesRelatedPerformerBases
						WHERE filesRelatedPerformerBases.filesBases_id = resources.filesBases_id
							AND filesRelatedPerformerBases.performerBases_id = performer.performerBases_id
					)
			) AS performerResources
			GROUP BY performer_id
		`, performerIds, performerIds).Scan(&list).Error
		if err != nil {
			return countMap, err
		}
	}

	for _, item := range list {
		countMap[item.PerformerID] = item.ResourceCount
	}
	return countMap, nil
}

// fillResourceCounts 将批量统计出的资源数量回填到演员切片。
// dataList 是值切片，内部通过索引写回 ResourceCount，因此调用方传入的切片内容会被更新。
func (t Performer) fillResourceCounts(db *gorm.DB, dataList []Performer, filesBasesId string) error {
	ids := make([]string, 0, len(dataList))
	for _, performer := range dataList {
		ids = append(ids, performer.ID)
	}
	countMap, err := t.resourceCountMap(db, ids, filesBasesId)
	if err != nil {
		return err
	}
	for i := range dataList {
		dataList[i].ResourceCount = countMap[dataList[i].ID]
	}
	return nil
}

// fillResourceCountsForResources 给资源列表中预加载出来的演员/导演回填资源数量。
//
// 资源列表可能来自首页、详情页、历史记录、热门等接口。这里按资源自身的 filesBases_id
// 分组后再批量统计，避免把不同文件库的同一演员数量混在一起，也避免在 Preload 阶段
// 对每个演员做逐行统计。
func (t Performer) fillResourceCountsForResources(db *gorm.DB, dataList []Resources) error {
	performerMap := make(map[string][]*Performer)
	for i := range dataList {
		filesBasesId := dataList[i].FilesBasesID
		for j := range dataList[i].Performers {
			performerMap[filesBasesId] = append(performerMap[filesBasesId], &dataList[i].Performers[j])
		}
		for j := range dataList[i].Directors {
			performerMap[filesBasesId] = append(performerMap[filesBasesId], &dataList[i].Directors[j])
		}
	}
	for filesBasesId, performers := range performerMap {
		ids := make([]string, 0, len(performers))
		exists := make(map[string]bool, len(performers))
		for _, performer := range performers {
			if !exists[performer.ID] {
				ids = append(ids, performer.ID)
				exists[performer.ID] = true
			}
		}
		countMap, err := t.resourceCountMap(db, ids, filesBasesId)
		if err != nil {
			return err
		}
		for _, performer := range performers {
			performer.ResourceCount = countMap[performer.ID]
		}
	}
	return nil
}

func (Performer) BasicList(db *gorm.DB, performerBasesIds []string, careerPerformer, careerDirector bool) (*[]PerformerBasic, error) {
	var list []PerformerBasic
	db = db.Model(&Performer{})
	if len(performerBasesIds) > 0 {
		db = db.Where("performerBases_id in (?) and status = 1", performerBasesIds)
	}
	if careerPerformer {
		db = db.Where("careerPerformer = 1")
	}
	if careerDirector {
		db = db.Where("careerDirector = 1")
	}
	err := db.Order("addTime desc").Find(&list).Error
	return &list, err
}
func (Performer) PhotosByPerformerBasesId_DB(db *gorm.DB, performerBasesId string) ([]string, error) {
	var photos []string
	err := db.Model(&Performer{}).Where("performerBases_id = ?", performerBasesId).Pluck("photo", &photos).Error
	return photos, err
}

func (Performer) SearchLastScraperUpdateTime(db *gorm.DB, performerBasesId, lastScraperUpdateTime string) (*[]PerformerBasic, error) {
	var list []PerformerBasic
	db = db.Model(&Performer{}).Where("performerBases_id = ?", performerBasesId)
	if lastScraperUpdateTime != "" {
		//转换成日期
		lastData := datatype.CustomDate{}
		lastData.SetValue(lastScraperUpdateTime)
		lastDataValue, err := lastData.Value()
		if err == nil {
			db = db.Where("lastScraperUpdateTime < ?", lastDataValue)
		}
	} else {
		db = db.Where("lastScraperUpdateTime is null")
	}
	err := db.Order("addTime desc").Find(&list).Error
	return &list, err
}

func (Performer) InfoByName(db *gorm.DB, performerBasesID, name string, searchAliasName bool) (*Performer, error) {
	var performer Performer
	if searchAliasName {
		db = db.Where("performerBases_id = ? and (name = ? or aliasName like ?)", performerBasesID, name, "%"+name+"%")
	} else {
		db = db.Where("performerBases_id = ? and name = ?", performerBasesID, name)
	}

	err := db.First(&performer).Error
	return &performer, err
}
func (Performer) InfoByID(db *gorm.DB, id string) (*Performer, error) {
	var performer Performer
	err := db.Where("id = ?", id).First(&performer).Error
	if err != nil {
		return &performer, err
	}
	dataList := []Performer{performer}
	err = Performer{}.fillResourceCounts(db, dataList, "")
	performer = dataList[0]
	return &performer, err
}

func (Performer) DataListByPerformerBasesId(db *gorm.DB, performerBasesId string) (*[]Performer, error) {
	var dataList []Performer
	err := db.Model(Performer{}).Where("performerBases_id = ?", performerBasesId).Find(&dataList).Error
	if err != nil {
		return &dataList, err
	}
	err = Performer{}.fillResourceCounts(db, dataList, "")
	return &dataList, err
}

func (Performer) DataList(db *gorm.DB, performerBasesId string, fetchCount bool, page, limit int, search, star, cup, charIndex, countFilesBasesId string) (*[]Performer, int64, error) {
	var dataList []Performer
	var total int64
	offset := (page - 1) * limit
	query := db.Model(Performer{}).Where("performerBases_id = ? and status = 1", performerBasesId)
	if charIndex != "" {
		charIndex = strings.ToLower(charIndex)
		if charIndex != "all" {
			query = query.Where("name like ? or keyWords like ?", charIndex+"%", charIndex+"%")
		}
	}
	if search != "" {
		searchQuery := "%" + search + "%"
		query = query.Where("keyWords like ? or name like ? or aliasName like ?", searchQuery, searchQuery, searchQuery)
	}
	if star != "" {
		query = query.Where("stars = ?", star)
	}
	if cup != "" && cup != "ALL" {
		if cup == "noCup" {
			query = query.Where("cup = ''")
		} else {
			query = query.Where("cup = ?", cup)
		}
	}
	if fetchCount {
		err := query.Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
	}
	query = query.Order("addTime desc").Limit(limit).Offset(offset)
	err := query.Find(&dataList).Error
	if err != nil {
		return nil, 0, err
	}
	err = Performer{}.fillResourceCounts(db, dataList, countFilesBasesId)
	if err != nil {
		return nil, 0, err
	}
	return &dataList, total, err
}
func (Performer) DataListByIds(db *gorm.DB, ids []string) (*[]Performer, error) {
	var dataList []Performer
	err := db.Model(Performer{}).Where("id in (?)", ids).Find(&dataList).Error
	if err != nil {
		return &dataList, err
	}
	err = Performer{}.fillResourceCounts(db, dataList, "")
	return &dataList, err
}

func (t Performer) ListTopPreferredPerformers(db *gorm.DB, preferredIds []string, mainPerformerBasesId string, shieldNoPerformerPhoto bool, limit int, countFilesBasesId string) (*[]Performer, error) {
	var dataList []Performer

	dataListByIds, err := t.ListByIds(db, preferredIds, countFilesBasesId)
	if err != nil {
		return nil, err
	}
	surplus := limit - len(*dataListByIds)

	if surplus > 0 {
		var surplusDataList []Performer
		query := db.Model(Performer{}).Where("performerBases_id = ? and status = 1", mainPerformerBasesId)
		if shieldNoPerformerPhoto {
			query = query.Where("photo != ''")
		}
		err := query.Order("addTime desc").Limit(surplus).Find(&surplusDataList).Error
		if err != nil {
			return nil, err
		}
		//组合数据
		err = t.fillResourceCounts(db, surplusDataList, countFilesBasesId)
		if err != nil {
			return nil, err
		}
		dataList = append(*dataListByIds, surplusDataList...)
	} else {
		dataList = *dataListByIds
	}

	return &dataList, err
}
func (t Performer) RecycleBin(db *gorm.DB, performerBasesId string) (*[]Performer, error) {
	var dataList []Performer
	err := db.Where("performerBases_id = ?", performerBasesId).Where("status = ?", false).Find(&dataList).Error
	return &dataList, err
}

func (Performer) ListByIds(db *gorm.DB, ids []string, countFilesBasesId string) (*[]Performer, error) {
	var dataList []Performer
	err := db.Model(Performer{}).Where("id in (?)", ids).Find(&dataList).Error
	if err != nil {
		return &dataList, err
	}
	err = Performer{}.fillResourceCounts(db, dataList, countFilesBasesId)
	// 构建 id 到索引的映射
	idIndexMap := make(map[string]int)
	for i, id := range ids {
		idIndexMap[id] = i
	}
	// 按 ids 的顺序排序
	sort.Slice(dataList, func(i, j int) bool {
		return idIndexMap[dataList[i].ID] < idIndexMap[dataList[j].ID]
	})
	return &dataList, err
}

func (Performer) Update(db *gorm.DB, performer *Performer, fields []string) error {
	result := db.Model(&performer).Select(fields).Updates(performer)
	if result.RowsAffected == 0 {
		return nil
	}
	return result.Error
}
func (Performer) Create(db *gorm.DB, performer *Performer) error {
	return db.Create(&performer).Error
}

func (Performer) DeleteByPerformerBasesIds(db *gorm.DB, performerBasesIds []string) error {
	return db.Unscoped().Where("performerBases_id in (?)", performerBasesIds).Delete(&Performer{}).Error
}
func (Performer) DeleteById(db *gorm.DB, id string) error {
	return db.Unscoped().Where("id = ?", id).Delete(&Performer{}).Error
}
