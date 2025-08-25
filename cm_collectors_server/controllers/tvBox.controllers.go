package controllers

import (
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	"cm_collectors_server/processors"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

type TVBox struct{}

// Home TVBox主页配置接口
func (TVBox) Home(c *gin.Context) {
	home := map[string]interface{}{
		"sites": []map[string]interface{}{
			{
				"key":         "local_videos",
				"name":        "CM_Collectors_3",
				"type":        1,
				"api":         "http://" + c.Request.Host + "/api/tvbox/sites/videos",
				"searchable":  1,
				"quickSearch": 1,
				"filterable":  1,
			},
		},
		"lives": []interface{}{},
		"parses": []map[string]interface{}{
			{
				"name": "本地解析",
				"type": 0,
				"url":  "",
			},
		},
		"flags": []string{},
		"ijk":   []interface{}{},
		"ads":   []string{},
	}

	c.JSON(http.StatusOK, home)
}

// Videos 提供视频分类和列表
func (t TVBox) Videos(c *gin.Context) {

	id := c.Query("ids")
	if id != "" {
		t.VideoDetail(c, id)
		return
	}

	// 获取分类参数
	typeId := c.Query("t")
	if typeId == "" {
		typeId = c.Query("type")
	}

	// 获取分页参数
	pageStr := c.Query("pg")
	if pageStr == "" {
		pageStr = "1"
	}
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	limit := 60

	// 获取搜索关键词参数
	keyword := c.Query("wd")
	if keyword == "" {
		keyword = c.Query("search")
	}

	filesBasesSlc, err := processors.FilesBases{}.DataList()
	// 构造分类信息
	classes := make([]map[string]interface{}, 0)
	for _, filesBases := range *filesBasesSlc {
		if filesBases.Status {
			classes = append(classes, map[string]interface{}{
				"type_id":   filesBases.ID,
				"type_name": filesBases.Name,
			})
		}

	}
	// 获取资源列表
	var resources *[]models.Resources
	// 获取资源列表
	if typeId != "" {
		req := datatype.ReqParam_ResourcesList{
			FilesBasesId: typeId,
			ParPaging: datatype.ParPaging{
				FetchCount: false,
				Limit:      limit,
				Page:       page,
			},
		}
		resources, _, err = processors.Resources{}.DataList(&req)
	} else {

		resources, err = processors.Resources{}.DataListAll()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 构造资源列表
	list := make([]map[string]interface{}, 0)
	for _, r := range *resources {
		resource := r
		// 获取演员信息
		performerNames := ""
		for i, performer := range resource.Performers {
			if i > 0 {
				performerNames += ","
			}
			performerNames += performer.Name
		}

		// 获取导演信息
		directorNames := ""
		for i, director := range resource.Directors {
			if i > 0 {
				directorNames += ","
			}
			directorNames += director.Name
		}

		// 获取标签信息
		tagNames := ""
		for i, tag := range resource.Tags {
			if i > 0 {
				tagNames += ","
			}
			tagNames += tag.Name
		}

		remarks := resource.Definition
		area := resource.Country
		year := ""
		if resource.IssuingDate != nil {
			year = strconv.Itoa(time.Time(*resource.IssuingDate).Year())
		}

		list = append(list, map[string]interface{}{
			"vod_id":       resource.ID,
			"vod_name":     truncateString(resource.Title, 16),
			"vod_pic":      "http://" + c.Request.Host + "/api/resCoverPoster/" + resource.FilesBasesID + "/" + resource.CoverPoster,
			"vod_remarks":  remarks,
			"vod_year":     year,
			"vod_area":     area,
			"vod_actor":    performerNames,
			"vod_director": directorNames,
			"vod_content":  resource.Abstract + " 标签：" + tagNames,
		})
	}

	videos := map[string]interface{}{
		"class": classes,
		"list":  list,
	}

	c.JSON(http.StatusOK, videos)
}

// VideoDetail 提供视频详情和播放链接
func (TVBox) VideoDetail(c *gin.Context, resourceId string) {

	// 获取资源详情
	resource, err := processors.Resources{}.Info(resourceId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到视频资源"})
		return
	}

	// 获取演员信息
	performerNames := ""
	for i, performer := range resource.Performers {
		if i > 0 {
			performerNames += ","
		}
		performerNames += performer.Name
	}

	// 获取导演信息
	directorNames := ""
	for i, director := range resource.Directors {
		if i > 0 {
			directorNames += ","
		}
		directorNames += director.Name
	}

	// 获取标签信息
	tagNames := ""
	for i, tag := range resource.Tags {
		if i > 0 {
			tagNames += ","
		}
		tagNames += tag.Name
	}

	// 构造播放链接
	playUrls := ""
	if len(resource.ResourcesDramaSeries) > 0 {
		for i, drama := range resource.ResourcesDramaSeries {
			if i > 0 {
				playUrls += "#"
			}
			title := fmt.Sprintf("第%d集", i+1)
			playUrls += title + "$http://" + c.Request.Host + "/api/video/mp4/" + drama.ID + "/v.mp4"
		}
	}

	area := resource.Country
	year := ""
	if resource.IssuingDate != nil {
		year = strconv.Itoa(time.Time(*resource.IssuingDate).Year())
	}

	detail := map[string]interface{}{
		"list": []map[string]interface{}{
			{
				"vod_id":        resource.ID,
				"vod_name":      truncateString(resource.Title, 16),
				"vod_pic":       "http://" + c.Request.Host + "/api/resCoverPoster/" + resource.FilesBasesID + "/" + resource.CoverPoster,
				"type_name":     "视频",
				"vod_year":      year,
				"vod_area":      area,
				"vod_remarks":   "全" + strconv.Itoa(len(resource.ResourcesDramaSeries)) + "集",
				"vod_actor":     performerNames,
				"vod_director":  directorNames,
				"vod_content":   tagNames + " # " + resource.Abstract,
				"vod_play_from": "本地播放",
				"vod_play_url":  playUrls,
				"vod_tag":       "local",
				"vod_class":     "本地视频",
			},
		},
	}

	c.JSON(http.StatusOK, detail)
}

func truncateString(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxLen]) + "..."
}
