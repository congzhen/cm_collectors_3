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
	dataList, total, err := processors.Performer{}.DataList(performerBasesId, fetchCount, page, limit, search, star, cup, charIndex)
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
	dataList, err := processors.Performer{}.ListTopPreferredPerformers(par.PreferredIds, par.MainPerformerBasesId, par.ShieldNoPerformerPhoto, par.Limit)
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

func (Performer) Create(c *gin.Context) {
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
func (Performer) Delete(c *gin.Context) {
	id := c.Param("id")
	err := processors.Performer{}.DeleteByID(id)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}
