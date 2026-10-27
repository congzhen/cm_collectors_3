package controllers

import (
	"cm_collectors_server/internal/selfupdate"
	"cm_collectors_server/response"
	"github.com/gin-gonic/gin"
)

type Update struct{}

func updater(c *gin.Context) *selfupdate.Manager {
	if selfupdate.Default == nil {
		response.FailWithMessage("更新服务不可用，请检查程序目录写权限和启动日志", c)
	}
	return selfupdate.Default
}
func (Update) Check(c *gin.Context) {
	m := updater(c)
	if m == nil {
		return
	}
	v, e := m.Check(c.Request.Context(), c.Query("manual") == "1")
	if e != nil {
		response.FailWithMessage("检测更新失败: "+e.Error(), c)
		return
	}
	response.OkWithData(v, c)
}
func (Update) Status(c *gin.Context) {
	m := updater(c)
	if m == nil {
		return
	}
	response.OkWithData(m.State(), c)
}
func (Update) Download(c *gin.Context) {
	m := updater(c)
	if m == nil {
		return
	}
	var p struct {
		Version string `json:"version"`
	}
	if e := c.ShouldBindJSON(&p); e != nil {
		response.FailWithMessage("无效的版本参数", c)
		return
	}
	if e := m.Download(p.Version); e != nil {
		response.FailWithMessage(e.Error(), c)
		return
	}
	response.OkWithData(m.State(), c)
}
func (Update) Install(c *gin.Context) {
	m := updater(c)
	if m == nil {
		return
	}
	var p struct {
		Version string `json:"version"`
		Job     string `json:"job"`
	}
	if e := c.ShouldBindJSON(&p); e != nil {
		response.FailWithMessage("无效的安装参数", c)
		return
	}
	if e := m.Install(p.Version, p.Job); e != nil {
		response.FailWithMessage(e.Error(), c)
		return
	}
	response.OkWithData(m.State(), c)
}
func (Update) Cancel(c *gin.Context) {
	m := updater(c)
	if m == nil {
		return
	}
	if e := m.CancelWait(); e != nil {
		response.FailWithMessage(e.Error(), c)
		return
	}
	response.OkWithData(m.State(), c)
}
