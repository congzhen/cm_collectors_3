package controllers

import (
	"cm_collectors_server/datatype"
	"cm_collectors_server/processors"
	"cm_collectors_server/response"

	"github.com/gin-gonic/gin"
)

type Login struct {
}

func (Login) setCookie(c *gin.Context, cookieName, token string) {
	c.SetCookie(cookieName, token, 0, "/", "", false, true)
}
func (t Login) Admin(c *gin.Context) {
	var par datatype.ReqParam_AdminLogin
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	token, err := processors.Login{}.AdminLogin(par.Password)
	if err := ResError(c, err); err != nil {
		return
	}
	t.setCookie(c, "adminToken", token)
	response.OkWithData(token, c)
}
