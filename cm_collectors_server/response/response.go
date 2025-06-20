package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	Success = 0
	Error   = 1
)

type Response struct {
	Status     bool   `json:"status"`
	StatusCode int    `json:"statusCode"`
	Msg        string `json:"msg"`
	Data       any    `json:"data"`
}

func Result(status bool, statusCode int, msg string, data any, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Status:     status,
		StatusCode: statusCode,
		Msg:        msg,
		Data:       data,
	})
}

func Ok(data any, msg string, c *gin.Context) {
	Result(true, Success, msg, data, c)
}

func OkWithData(data any, c *gin.Context) {
	Ok(data, "OK", c)
}

func Fail(data any, msg string, c *gin.Context) {
	Result(false, Error, msg, data, c)
}

func FailWithMessage(msg string, c *gin.Context) {
	Fail(map[string]any{}, msg, c)
}

func FailWithCode(code int, c *gin.Context) {
	Result(false, int(code), GetErrorMsg(code), map[string]any{}, c)
}

func FailWithError(code int, err error, c *gin.Context) {
	Result(false, int(code), GetErrorMsg(code)+" "+err.Error(), map[string]any{}, c)
}

func FailPermissions(c *gin.Context) {
	c.JSON(401, Response{
		Status:     false,
		StatusCode: 401,
		Msg:        "Unauthorized",
		Data:       false,
	})
}
