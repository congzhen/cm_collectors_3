package core

import (
	"encoding/json"
	"io"
	"os"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func logInit(logPath, logLevel string) {
	//设置日志等级
	switch logLevel {
	case "Error":
		logrus.SetLevel(logrus.DebugLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
	file, _ := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	writers := []io.Writer{
		file,
		os.Stdout}
	//同时写文件和屏幕
	fileAndStdoutWriter := io.MultiWriter(writers...)
	logrus.SetOutput(fileAndStdoutWriter)
	//显示行号
	logrus.SetReportCaller(true)
}

func LogErr(err error) {
	// 获取调用者的信息
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}
	// 记录错误日志，包括文件名和行号
	logrus.Errorf("Error:%v; at %s:%d", err, file, line)
}

func LogErrorRequest(err error, c *gin.Context) {
	LogErrorRequestData(err, c, nil)
}

func LogErrorRequestData(err error, c *gin.Context, data any) {
	// 获取访问日志信息
	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery
	if raw != "" {
		path = path + "?" + raw
	}
	clientIP := c.ClientIP()
	method := c.Request.Method
	// 获取请求头信息
	requestHeaders := c.Request.Header

	// 获取查询参数和POST表单数据（注意：根据实际情况选择合适的方法，如PostForm、GetQuery等）
	queryParams := c.Request.URL.Query()
	postData := c.Request.PostForm

	jsonData, _ := json.Marshal(data)

	// 创建一个包含访问日志和错误信息的日志条目
	fields := logrus.Fields{
		"client_ip":       clientIP,
		"method":          method,
		"path":            path,
		"params_query":    queryParams,
		"params_post":     postData,
		"params_data":     string(jsonData),
		"request_headers": requestHeaders,
	}

	logrus.WithFields(fields).Errorln(err)
}
