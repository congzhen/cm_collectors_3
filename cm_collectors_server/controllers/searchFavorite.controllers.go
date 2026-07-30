package controllers

import (
	"cm_collectors_server/processors"
	"cm_collectors_server/response"

	"github.com/gin-gonic/gin"
)

type SearchFavorite struct{}

func (SearchFavorite) List(c *gin.Context) {
	list, err := processors.SearchFavorite{}.List(c.Param("filesBasesId"))
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(list, c)
}

func (SearchFavorite) Create(c *gin.Context) {
	var input processors.SearchFavoriteInput
	if err := ParameterHandleShouldBindJSON(c, &input); err != nil {
		return
	}
	favorite, err := processors.SearchFavorite{}.Create(input)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(favorite, c)
}

func (SearchFavorite) Update(c *gin.Context) {
	var input processors.SearchFavoriteUpdateInput
	if err := ParameterHandleShouldBindJSON(c, &input); err != nil {
		return
	}
	favorite, err := processors.SearchFavorite{}.Update(c.Param("id"), input)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(favorite, c)
}

func (SearchFavorite) Delete(c *gin.Context) {
	if err := (processors.SearchFavorite{}).Delete(c.Param("id")); err != nil {
		ResError(c, err)
		return
	}
	response.OkWithData(true, c)
}
