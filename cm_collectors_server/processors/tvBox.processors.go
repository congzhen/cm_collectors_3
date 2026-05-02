package processors

import (
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type TVBox struct{}

const TVBox_TitleTrunNum = 17

func (t TVBox) Videos(host, typeId, searchText string, page, limit int) (map[string]interface{}, error) {
	classes, err := t.getClasses()
	if err != nil {
		return nil, err
	}
	resources, total, err := t.getResources(typeId, searchText, page, limit)
	if err != nil {
		return nil, err
	}
	list := t.buildResourceList(resources, host)

	pageCount := int(total) / limit
	if int(total)%limit != 0 {
		pageCount++
	}
	if pageCount == 0 {
		pageCount = 1
	}

	videos := map[string]interface{}{
		"class":     classes,
		"list":      list,
		"page":      page,
		"pagecount": pageCount,
		"limit":     limit,
		"total":     total,
	}
	return videos, nil
}


func (t TVBox) VideoDetail(host string, resourceIds []string) (map[string]interface{}, error) {
	list := make([]map[string]interface{}, 0)

	resourceSlc, err := Resources{}.DataListByIds(resourceIds)
	if err != nil {
		return nil, err
	}
	if len(*resourceSlc) == 0 {
		return nil, fmt.Errorf("未找到资源")
	}

	// 建立 id → resource 的索引，按 resourceIds 顺序输出
	resourceMap := make(map[string]models.Resources, len(*resourceSlc))
	for _, r := range *resourceSlc {
		resourceMap[r.ID] = r
	}

	for _, id := range resourceIds {
		resource, ok := resourceMap[id]
		if !ok {
			continue
		}
		performerNames := t.getPerformerNames(resource.Performers)
		directorNames := t.getDirectorNames(resource.Directors)
		tagNames := t.getTagNames(resource.Tags)
		playFrom, playUrls := t.buildPlayUrls(resource.ResourcesDramaSeries, host)
		area := resource.Country
		year := t.getYearString(resource.IssuingDate)

		list = append(list, map[string]interface{}{
			"vod_id":        resource.ID,
			"vod_name":      t.truncateString(resource.Title, TVBox_TitleTrunNum),
			"vod_pic":       t.coverPosterURL(host, resource.FilesBasesID, resource.CoverPoster),
			"type_name":     "视频",
			"vod_year":      year,
			"vod_area":      area,
			"vod_remarks":   fmt.Sprintf("全%d集", len(resource.ResourcesDramaSeries)),
			"vod_actor":     performerNames,
			"vod_director":  directorNames,
			"vod_content":   fmt.Sprintf("%s # %s", tagNames, resource.Abstract),
			"vod_play_from": playFrom,
			"vod_play_url":  playUrls,
			"vod_tag":       "local",
			"vod_class":     "本地视频",
		})
	}

	if len(list) == 0 {
		return nil, fmt.Errorf("未找到视频资源")
	}

	detail := map[string]interface{}{
		"list": list,
	}

	return detail, nil
}

// getClasses 获取分类信息
func (TVBox) getClasses() ([]map[string]interface{}, error) {
	filesBasesSlc, err := FilesBases{}.DataList()
	if err != nil {
		return nil, err
	}

	classes := make([]map[string]interface{}, 0)
	for _, filesBases := range *filesBasesSlc {
		if filesBases.Status {
			classes = append(classes, map[string]interface{}{
				"type_id":   filesBases.ID,
				"type_name": filesBases.Name,
			})
		}
	}
	return classes, nil
}

// getResources 获取资源列表
func (TVBox) getResources(typeId, searchText string, page, limit int) (*[]models.Resources, int64, error) {
	if typeId != "" {
		var searchTextSlc []string
		if searchText != "" {
			searchTextSlc = []string{searchText}
		}
		req := datatype.ReqParam_ResourcesList{
			FilesBasesId: typeId,
			ParPaging: datatype.ParPaging{
				FetchCount: true,
				Limit:      limit,
				Page:       page,
			},
			SearchData: datatype.ReqParam_SearchData{
				SearchTextSlc: searchTextSlc,
			},
		}
		return Resources{}.DataList(&req)
	}
	// typeId 为空且无搜索词时，返回 TvboxRecommend 表中的推荐资源
	if searchText == "" {
		recommends, err := TvboxRecommend{}.List()
		if err != nil {
			return nil, 0, err
		}
		if recommends != nil && len(*recommends) > 0 {
			resources := make([]models.Resources, 0, len(*recommends))
			for _, rec := range *recommends {
				resources = append(resources, rec.Resource)
			}
			total := int64(len(resources))
			return &resources, total, nil
		}
	}
	// typeId 为空时跨分类查询（含全局搜索）
	return Resources{}.DataListAllSearch(searchText, page, limit)
}

// buildResourceList 构造资源列表
func (t TVBox) buildResourceList(resources *[]models.Resources, host string) []map[string]interface{} {
	list := make([]map[string]interface{}, 0)
	for _, r := range *resources {
		resource := r
		// 获取演员信息
		performerNames := t.getPerformerNames(resource.Performers)

		// 获取导演信息
		directorNames := t.getDirectorNames(resource.Directors)

		// 获取标签信息
		tagNames := t.getTagNames(resource.Tags)

		remarks := resource.Definition
		area := resource.Country
		year := t.getYearString(resource.IssuingDate)

		list = append(list, map[string]interface{}{
			"vod_id":       resource.ID,
			"vod_name":     t.truncateString(resource.Title, TVBox_TitleTrunNum),
			"vod_pic":      t.coverPosterURL(host, resource.FilesBasesID, resource.CoverPoster),
			"vod_remarks":  remarks,
			"vod_year":     year,
			"vod_area":     area,
			"vod_actor":    performerNames,
			"vod_director": directorNames,
			"vod_content":  fmt.Sprintf("%s 标签：%s", resource.Abstract, tagNames),
		})
	}
	return list
}

// getPerformerNames 获取演员名字列表
func (TVBox) getPerformerNames(performers []models.Performer) string {
	names := make([]string, len(performers))
	for i, performer := range performers {
		names[i] = performer.Name
	}
	return strings.Join(names, ",")
}

// getDirectorNames 获取导演名字列表
func (TVBox) getDirectorNames(directors []models.Performer) string {
	names := make([]string, len(directors))
	for i, director := range directors {
		names[i] = director.Name
	}
	return strings.Join(names, ",")
}

// getTagNames 获取标签名字列表
func (TVBox) getTagNames(tags []models.Tag) string {
	names := make([]string, len(tags))
	for i, tag := range tags {
		names[i] = tag.Name
	}
	return strings.Join(names, ",")
}

// buildPlayUrls 构造两路播放源：m3u8（默认）和 mp4，格式符合 TVBox 多源规范
func (TVBox) buildPlayUrls(dramaSeries []models.ResourcesDramaSeries, host string) (playFrom string, playUrl string) {
	m3u8Plays := make([]string, len(dramaSeries))
	mp4Plays := make([]string, len(dramaSeries))
	for i, drama := range dramaSeries {
		title := fmt.Sprintf("第%d集", i+1)
		m3u8Plays[i] = title + "$" + fmt.Sprintf("http://%s/api/video/m3u8/%s/v.m3u8", host, drama.ID)
		mp4Plays[i] = title + "$" + fmt.Sprintf("http://%s/api/video/mp4/%s/v.mp4", host, drama.ID)
	}
	playFrom = "m3u8$$$mp4"
	playUrl = strings.Join(m3u8Plays, "#") + "$$$" + strings.Join(mp4Plays, "#")
	return
}

// getYearString 获取年份字符串
func (TVBox) getYearString(issuingDate *datatype.CustomDate) string {
	if issuingDate != nil {
		return strconv.Itoa(time.Time(*issuingDate).Year())
	}
	return ""
}

func (TVBox) coverPosterURL(host, filesBasesID, coverPoster string) string {
	if coverPoster == "" {
		return ""
	}
	return fmt.Sprintf("http://%s/api/resCoverPoster/%s/%s", host, filesBasesID, coverPoster)
}

func (TVBox) truncateString(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxLen]) + "..."
}
