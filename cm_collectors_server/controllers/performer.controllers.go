package controllers

import (
	"cm_collectors_server/datatype"
	"cm_collectors_server/processors"
	"cm_collectors_server/response"
	"strings"

	"github.com/gin-gonic/gin"
)

type Performer struct{}

func (Performer) Info(c *gin.Context) {
	id := c.Param("id")
	info, err := processors.Performer{}.InfoByID(id)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(info, c)
}
func (Performer) BasicList(c *gin.Context) {
	var par datatype.ReqParam_PerformersList
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	dataList, err := processors.Performer{}.BasicList(par.PerformerBasesIds, par.CareerPerformer, par.CareerDirector)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(dataList, c)
}
func (Performer) DataListByIds(c *gin.Context) {
	idsStr := c.Param("ids")
	if idsStr == "" {
		response.OkWithData([]any{}, c)
	}
	ids := strings.Split(idsStr, ",")
	if len(ids) == 0 {
		response.OkWithData([]any{}, c)
	}
	dataList, err := processors.Performer{}.DataListByIds(ids, true)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(dataList, c)
}
func (Performer) DataList(c *gin.Context) {
	performerBasesId := c.Param("performerBasesId")
	var fetchCount bool
	var page int
	var limit int
	GetUrlParameter_Param(c, "fetchCount", &fetchCount)
	GetUrlParameter_Param(c, "page", &page)
	GetUrlParameter_Param(c, "limit", &limit)
	search := c.Query("search")
	star := c.Query("star")
	cup := c.Query("cup")
	charIndex := c.Query("charIndex")
	// countFilesBasesId 只影响 resourceCount 的统计范围，不影响演员列表本身。
	// 例如同一个演员集被多个文件库关联时，前端可传当前 filesBasesId，让角标只显示当前文件库资源数。
	countFilesBasesId := c.Query("countFilesBasesId")
	dataList, total, err := processors.Performer{}.DataList(performerBasesId, fetchCount, page, limit, search, star, cup, charIndex, countFilesBasesId)
	if err := ResError(c, err); err != nil {
		return
	}
	resDataList := datatype.ResDataList{
		DataList: dataList,
		Total:    total,
	}
	response.OkWithData(resDataList, c)
}

func (Performer) ListTopPreferredPerformers(c *gin.Context) {
	var par datatype.ReqParam_TopPreferredPerformers
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	// 常用演员接口也可能出现在某个具体文件库首页，需要按当前文件库统计角标数量。
	countFilesBasesId := c.Query("countFilesBasesId")
	dataList, err := processors.Performer{}.ListTopPreferredPerformers(par.PreferredIds, par.MainPerformerBasesId, par.ShieldNoPerformerPhoto, par.Limit, countFilesBasesId)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(dataList, c)
}

func (Performer) RecycleBin(c *gin.Context) {
	performerBasesId := c.Param("performerBasesId")
	dataList, err := processors.Performer{}.RecycleBin(performerBasesId)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(dataList, c)
}

func (Performer) CreatePerformer(c *gin.Context) {
	var par datatype.ReqParam_PerformerData
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	info, err := processors.Performer{}.Create(&par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(info, c)
}
func (Performer) UpdatePerformer(c *gin.Context) {
	var par datatype.ReqParam_PerformerData
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	info, err := processors.Performer{}.Update(&par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(info, c)
}
func (Performer) UpdatePerformerStatus(c *gin.Context) {
	var par datatype.ReqParam_PerformerStatus
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.Performer{}.UpdatePerformerStatus(par.ID, par.Status)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (Performer) PerformerBasesUpdate(c *gin.Context) {
	var par datatype.ReqParam_UpdatePerformerBases
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.PerformerBases{}.Update(&par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (Performer) ExportPerformerBases(c *gin.Context) {
	id := c.Param("id")
	jsonData, err := processors.PerformerBases{}.Export(id)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(jsonData, c)
}
func (Performer) ImportPerformerBases(c *gin.Context) {
	var par datatype.ReqParam_ImportPerformerBases
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	importNum, err := processors.PerformerBases{}.Import(par.PerformerDatabaseId, par.Content, par.ReconstructId)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(importNum, c)
}

func (Performer) CreatePerformerBases(c *gin.Context) {
	var par datatype.ReqParam_CreatePerformerBases
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	id, err := processors.PerformerBases{}.Create(par.Name)
	if err := ResError(c, err); err != nil {
		return
	}
	info, err := processors.PerformerBases{}.InfoById(id)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(info, c)
}
func (Performer) DeletePerformer(c *gin.Context) {
	id := c.Param("id")
	err := processors.Performer{}.DeleteByID(id)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}
func (Performer) MigratePerformer(c *gin.Context) {
	var par datatype.ReqParam_MigratePerformer
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.Performer{}.MigratePerformer(par.PerformerId, par.PerformerBasesId)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}
