package controllers

import (
	"cm_collectors_server/processors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type TVBox struct{}

// Home TVBox主页配置接口
func (TVBox) Home(c *gin.Context) {
	// 获取推荐内容
	recommendVideos := processors.TVBox{}.RecommendVideos(c.Request.Host, 30) // 获取30个推荐视频

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
		"flags":     []string{},
		"ijk":       []interface{}{},
		"ads":       []string{},
		"recommend": recommendVideos,
	}

	c.JSON(http.StatusOK, home)
}

// Videos 提供视频分类和列表
func (t TVBox) Videos(c *gin.Context) {

	id := c.Query("ids")
	if id != "" {
		ids := strings.Split(id, ",")
		videos, err := processors.TVBox{}.VideoDetail(c.Request.Host, ids)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.JSON(http.StatusOK, videos)
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
	videos, err := processors.TVBox{}.Videos(c.Request.Host, typeId, keyword, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, videos)
}
