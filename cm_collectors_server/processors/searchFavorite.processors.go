package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	"errors"
	"strings"

	"gorm.io/gorm"
)

const (
	searchFavoriteSchemaVersion = 1
	searchFavoriteLimit         = 100
)

type SearchFavorite struct{}

type SearchFavoriteInput struct {
	FilesBasesID string                       `json:"filesBasesId"`
	SearchData   datatype.ReqParam_SearchData `json:"searchData"`
}

type SearchFavoriteUpdateInput struct {
	SearchData *datatype.ReqParam_SearchData `json:"searchData"`
}

func (SearchFavorite) List(filesBasesID string) ([]models.SearchFavorite, error) {
	if strings.TrimSpace(filesBasesID) == "" {
		return nil, errors.New("文件库不能为空")
	}
	list, err := (models.SearchFavorite{}).List(core.DBS(), filesBasesID)
	if err != nil {
		return nil, err
	}
	for i := range list {
		normalizeSearchFavoriteData(&list[i].SearchData)
		invalid, err := sanitizeSearchFavoriteConditions(core.DBS(), &list[i].SearchData)
		if err != nil {
			return nil, err
		}
		list[i].InvalidConditions = invalid
		list[i].OptionLabels, err = searchFavoriteOptionLabels(core.DBS(), list[i].SearchData)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (SearchFavorite) Create(input SearchFavoriteInput) (*models.SearchFavorite, error) {
	input.FilesBasesID = strings.TrimSpace(input.FilesBasesID)
	normalizeSearchFavoriteData(&input.SearchData)
	if !hasSearchFavoriteConditions(input.SearchData) {
		return nil, errors.New("当前没有可保存的搜索条件")
	}
	db := core.DBS()
	var filesBasesCount int64
	if err := db.Model(&models.FilesBases{}).Where("id = ?", input.FilesBasesID).Count(&filesBasesCount).Error; err != nil {
		return nil, err
	}
	if filesBasesCount == 0 {
		return nil, errors.New("文件库不存在")
	}
	count, err := (models.SearchFavorite{}).CountByFilesBasesID(db, input.FilesBasesID)
	if err != nil {
		return nil, err
	}
	if count >= searchFavoriteLimit {
		return nil, errors.New("每个文件库最多保存100条搜索收藏")
	}
	id := core.GenerateUniqueID()
	sortValue, err := (models.SearchFavorite{}).MaxSort(db, input.FilesBasesID)
	if err != nil {
		return nil, err
	}
	now := core.TimeNow()
	favorite := &models.SearchFavorite{
		ID:            id,
		FilesBasesID:  input.FilesBasesID,
		SearchData:    input.SearchData,
		SchemaVersion: searchFavoriteSchemaVersion,
		Sort:          sortValue + 1,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := (models.SearchFavorite{}).Create(db, favorite); err != nil {
		return nil, err
	}
	favorite.OptionLabels, err = searchFavoriteOptionLabels(db, favorite.SearchData)
	if err != nil {
		return nil, err
	}
	return favorite, nil
}

func (SearchFavorite) Update(id string, input SearchFavoriteUpdateInput) (*models.SearchFavorite, error) {
	db := core.DBS()
	favorite, err := (models.SearchFavorite{}).Get(db, id)
	if err != nil {
		return nil, err
	}
	if input.SearchData == nil {
		return nil, errors.New("搜索条件不能为空")
	}
	normalizeSearchFavoriteData(input.SearchData)
	if !hasSearchFavoriteConditions(*input.SearchData) {
		return nil, errors.New("当前没有可保存的搜索条件")
	}
	favorite.SearchData = *input.SearchData
	favorite.SchemaVersion = searchFavoriteSchemaVersion
	favorite.UpdatedAt = core.TimeNow()
	if err := (models.SearchFavorite{}).Save(db, favorite); err != nil {
		return nil, err
	}
	normalizeSearchFavoriteData(&favorite.SearchData)
	invalid, err := sanitizeSearchFavoriteConditions(db, &favorite.SearchData)
	if err != nil {
		return nil, err
	}
	favorite.InvalidConditions = invalid
	favorite.OptionLabels, err = searchFavoriteOptionLabels(db, favorite.SearchData)
	if err != nil {
		return nil, err
	}
	return favorite, nil
}

func (SearchFavorite) Delete(id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("搜索收藏不能为空")
	}
	return (models.SearchFavorite{}).Delete(core.DBS(), id)
}

func normalizeSearchFavoriteData(data *datatype.ReqParam_SearchData) {
	if data.SearchTextSlc == nil {
		data.SearchTextSlc = []string{}
	}
	data.SearchTextSlc = uniqueNonEmptyStrings(data.SearchTextSlc)
	if data.Sort == "" {
		data.Sort = datatype.E_searchSort("addTimeDesc")
	}
	normalizeSearchFavoriteGroup(&data.Country)
	normalizeSearchFavoriteGroup(&data.Definition)
	normalizeSearchFavoriteGroup(&data.VideoCodec)
	normalizeSearchFavoriteGroup(&data.Year)
	normalizeSearchFavoriteGroup(&data.Star)
	normalizeSearchFavoriteGroup(&data.Performer)
	normalizeSearchFavoriteGroup(&data.Cup)
	if data.Tag == nil {
		data.Tag = map[string]datatype.I_searchGroup{}
	}
	for id, group := range data.Tag {
		normalizeSearchFavoriteGroup(&group)
		data.Tag[id] = group
	}
}

func normalizeSearchFavoriteGroup(group *datatype.I_searchGroup) {
	switch group.Logic {
	case datatype.E_searchLogic("single"),
		datatype.E_searchLogic("multiAnd"),
		datatype.E_searchLogic("multiOr"),
		datatype.E_searchLogic("not"):
	default:
		group.Logic = datatype.E_searchLogic("single")
	}
	group.Options = uniqueNonEmptyStrings(group.Options)
}

func uniqueNonEmptyStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func hasSearchFavoriteConditions(data datatype.ReqParam_SearchData) bool {
	if len(data.SearchTextSlc) > 0 {
		return true
	}
	groups := []datatype.I_searchGroup{
		data.Country, data.Definition, data.VideoCodec, data.Year,
		data.Star, data.Performer, data.Cup,
	}
	for _, group := range groups {
		if len(group.Options) > 0 {
			return true
		}
	}
	for _, group := range data.Tag {
		if len(group.Options) > 0 {
			return true
		}
	}
	return false
}

func sanitizeSearchFavoriteConditions(db *gorm.DB, data *datatype.ReqParam_SearchData) (int, error) {
	invalid := 0
	tagIDs := make([]string, 0)
	for _, group := range data.Tag {
		tagIDs = append(tagIDs, group.Options...)
	}
	validTagIDs, err := existingSearchFavoriteIDs(db, &models.Tag{}, tagIDs)
	if err != nil {
		return 0, err
	}
	for classID, group := range data.Tag {
		group.Options, invalid = filterSearchFavoriteIDs(group.Options, validTagIDs, invalid, false)
		data.Tag[classID] = group
	}

	validPerformerIDs, err := existingSearchFavoriteIDs(db, &models.Performer{}, data.Performer.Options)
	if err != nil {
		return 0, err
	}
	data.Performer.Options, invalid = filterSearchFavoriteIDs(
		data.Performer.Options, validPerformerIDs, invalid, true,
	)
	return invalid, nil
}

func existingSearchFavoriteIDs(db *gorm.DB, model interface{}, ids []string) (map[string]struct{}, error) {
	result := make(map[string]struct{})
	if len(ids) == 0 {
		return result, nil
	}
	var existing []string
	if err := db.Model(model).Where("id IN ?", ids).Pluck("id", &existing).Error; err != nil {
		return nil, err
	}
	for _, id := range existing {
		result[id] = struct{}{}
	}
	return result, nil
}

func filterSearchFavoriteIDs(values []string, valid map[string]struct{}, invalid int, preserveNot bool) ([]string, int) {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if preserveNot && value == datatype.V_Search_Not {
			result = append(result, value)
			continue
		}
		if _, ok := valid[value]; ok {
			result = append(result, value)
		} else {
			invalid++
		}
	}
	return result, invalid
}

func searchFavoriteOptionLabels(db *gorm.DB, data datatype.ReqParam_SearchData) (map[string]string, error) {
	labels := make(map[string]string)
	tagIDs := make([]string, 0)
	for _, group := range data.Tag {
		tagIDs = append(tagIDs, group.Options...)
	}
	type optionLabel struct {
		ID   string
		Name string
	}
	if len(tagIDs) > 0 {
		var tags []optionLabel
		if err := db.Model(&models.Tag{}).Select("id", "name").Where("id IN ?", tagIDs).Scan(&tags).Error; err != nil {
			return nil, err
		}
		for _, tag := range tags {
			labels[tag.ID] = tag.Name
		}
	}
	performerIDs := make([]string, 0, len(data.Performer.Options))
	for _, id := range data.Performer.Options {
		if id != datatype.V_Search_Not {
			performerIDs = append(performerIDs, id)
		}
	}
	if len(performerIDs) > 0 {
		var performers []optionLabel
		if err := db.Model(&models.Performer{}).Select("id", "name").Where("id IN ?", performerIDs).Scan(&performers).Error; err != nil {
			return nil, err
		}
		for _, performer := range performers {
			labels[performer.ID] = performer.Name
		}
	}
	return labels, nil
}
