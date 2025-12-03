package controllers

import (
	"cm_collectors_server/datatype"
	"cm_collectors_server/processors"
	"cm_collectors_server/response"

	"github.com/gin-gonic/gin"
)

type CronJobs struct {
}

func (CronJobs) List(c *gin.Context) {
	list, err := processors.CronJobs{}.DataList()
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(list, c)
}
func (CronJobs) Exec(c *gin.Context) {
	id := c.Param("cronJobsId")
	err := processors.CronJobs{}.Exec(id)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}
func (CronJobs) Create(c *gin.Context) {
	var par datatype.ReqParam_CreateCronJobs
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	info, err := processors.CronJobs{}.Create(par.FilesBasesId, par.JobsType, par.CronExpression)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(info, c)
}
func (CronJobs) Update(c *gin.Context) {
	var par datatype.ReqParam_UpdateCronJobs
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	info, err := processors.CronJobs{}.Update(par.ID, par.FilesBasesId, par.JobsType, par.CronExpression)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(info, c)
}
func (CronJobs) Delete(c *gin.Context) {
	id := c.Param("cronJobsId")
	err := processors.CronJobs{}.Delete(id)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}
