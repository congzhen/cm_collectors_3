package controllers

import (
	"cm_collectors_server/datatype"
	"cm_collectors_server/processors"
	"cm_collectors_server/response"

	"github.com/gin-gonic/gin"
)

type Performer struct{}

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
	dataList, total, err := processors.Performer{}.DataList(performerBasesId, fetchCount, page, limit, search, star, cup)
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
