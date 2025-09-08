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
	// 获取分类信息
	classes, err := t.getClasses()
	if err != nil {
		return nil, err
	}
	resources, err := t.getResources(typeId, searchText, page, limit)
	if err != nil {
		return nil, err
	}
	// 构造资源列表
	list := t.buildResourceList(resources, host)

	videos := map[string]interface{}{
		"class": classes,
		"list":  list,
	}
	return videos, nil
}

// RecommendVideos 获取推荐视频列表
func (t TVBox) RecommendVideos(host string, limit int) []map[string]interface{} {
	resources, err := Resources{}.DataListAll(1, limit)
	if err != nil || resources == nil || len(*resources) == 0 {
		return []map[string]interface{}{}
	}

	// 取前limit个资源作为推荐
	count := limit
	if len(*resources) < limit {
		count = len(*resources)
	}

	recommendList := make([]map[string]interface{}, 0, count)
	for i := 0; i < count && i < len(*resources); i++ {
		resource := (*resources)[i]
		performerNames := t.getPerformerNames(resource.Performers)
		directorNames := t.getDirectorNames(resource.Directors)
		tagNames := t.getTagNames(resource.Tags)
		year := t.getYearString(resource.IssuingDate)

		recommendList = append(recommendList, map[string]interface{}{
			"vod_id":       resource.ID,
			"vod_name":     t.truncateString(resource.Title, TVBox_TitleTrunNum),
			"vod_pic":      fmt.Sprintf("http://%s/api/resCoverPoster/%s/%s", host, resource.FilesBasesID, resource.CoverPoster),
			"vod_remarks":  resource.Definition,
			"vod_year":     year,
			"vod_area":     resource.Country,
			"vod_actor":    performerNames,
			"vod_director": directorNames,
			"vod_content":  fmt.Sprintf("%s 标签：%s", resource.Abstract, tagNames),
		})
	}

	return recommendList
}

// ... existing code ...

func (t TVBox) VideoDetail(host string, resourceIds []string) (map[string]interface{}, error) {
	list := make([]map[string]interface{}, 0)

	resourceSlc, err := Resources{}.DataListByIds(resourceIds)
	if err != nil {
		return nil, err
	}
	if len(*resourceSlc) == 0 {
		return nil, fmt.Errorf("未找到资源")
	}
	for _, _resource := range *resourceSlc {
		resource := _resource
		performerNames := t.getPerformerNames(resource.Performers)
		// 获取导演信息
		directorNames := t.getDirectorNames(resource.Directors)

		// 获取标签信息
		tagNames := t.getTagNames(resource.Tags)

		// 构造播放链接
		playUrls := t.buildPlayUrls(resource.ResourcesDramaSeries, host)

		area := resource.Country
		year := t.getYearString(resource.IssuingDate)

		detail := map[string]interface{}{
			"vod_id":        resource.ID,
			"vod_name":      t.truncateString(resource.Title, TVBox_TitleTrunNum),
			"vod_pic":       fmt.Sprintf("http://%s/api/resCoverPoster/%s/%s", host, resource.FilesBasesID, resource.CoverPoster),
			"type_name":     "视频",
			"vod_year":      year,
			"vod_area":      area,
			"vod_remarks":   fmt.Sprintf("全%d集", len(resource.ResourcesDramaSeries)),
			"vod_actor":     performerNames,
			"vod_director":  directorNames,
			"vod_content":   fmt.Sprintf("%s # %s", tagNames, resource.Abstract),
			"vod_play_from": "本地播放",
			"vod_play_url":  playUrls,
			"vod_tag":       "local",
			"vod_class":     "本地视频",
		}

		list = append(list, detail)
	}

	if len(list) == 0 {
		return nil, fmt.Errorf("未找到视频资源")
	}

	detail := map[string]interface{}{
		"list": list,
	}

	return detail, nil
}

// ... existing code ...

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
func (TVBox) getResources(typeId, searchText string, page, limit int) (*[]models.Resources, error) {
	if typeId != "" {
		req := datatype.ReqParam_ResourcesList{
			FilesBasesId: typeId,
			ParPaging: datatype.ParPaging{
				FetchCount: false,
				Limit:      limit,
				Page:       page,
			},
			SearchData: datatype.ReqParam_SearchData{
				SearchTextSlc: []string{searchText},
			},
		}
		resources, _, err := Resources{}.DataList(&req)
		return resources, err
	}

	resources, err := Resources{}.DataListAll(page, limit)
	return resources, err
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
			"vod_pic":      fmt.Sprintf("http://%s/api/resCoverPoster/%s/%s", host, resource.FilesBasesID, resource.CoverPoster),
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

// buildPlayUrls 构造播放链接
func (TVBox) buildPlayUrls(dramaSeries []models.ResourcesDramaSeries, host string) string {
	plays := make([]string, len(dramaSeries))
	for i, drama := range dramaSeries {
		title := fmt.Sprintf("第%d集", i+1)
		plays[i] = fmt.Sprintf("%s$http://%s/api/video/mp4/%s/v.mp4", title, host, drama.ID)
	}
	return strings.Join(plays, "#")
}

// getYearString 获取年份字符串
func (TVBox) getYearString(issuingDate *datatype.CustomDate) string {
	if issuingDate != nil {
		return strconv.Itoa(time.Time(*issuingDate).Year())
	}
	return ""
}

func (TVBox) truncateString(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxLen]) + "..."
}
