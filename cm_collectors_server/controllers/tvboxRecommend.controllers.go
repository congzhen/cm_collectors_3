package controllers

import (
	"cm_collectors_server/datatype"
	"cm_collectors_server/processors"
	"cm_collectors_server/response"

	"github.com/gin-gonic/gin"
)

type TvboxRecommend struct{}

func (TvboxRecommend) List(c *gin.Context) {
	list, err := processors.TvboxRecommend{}.List()
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(list, c)
}

func (TvboxRecommend) Add(c *gin.Context) {
	resourceId := c.Param("resourceId")
	err := processors.TvboxRecommend{}.Add(resourceId)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (TvboxRecommend) Delete(c *gin.Context) {
	id := c.Param("id")
	err := processors.TvboxRecommend{}.Delete(id)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (TvboxRecommend) UpdateSort(c *gin.Context) {
	var par []datatype.ReqParam_TvboxRecommendSort
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.TvboxRecommend{}.UpdateSort(par)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}
