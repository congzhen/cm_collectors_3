package controllers

import (
	"cm_collectors_server/core"
	"cm_collectors_server/errorMessage"
	"cm_collectors_server/response"
	"cm_collectors_server/tool/filter"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"runtime"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetUrlParameter_Param 从 HTTP 请求的 URL 参数中提取指定名称的参数，并将其解析到目标对象中。
// 该函数是为了解决如何将 URL 参数中的值映射到目标对象的问题，特别是当参数值的类型可能变化时。
// 参数:
//
//	c *gin.Context: Gin 框架的上下文，用于处理 HTTP 请求和响应。
//	perName string: 需要提取的 URL 参数的名称。
//	recipient *T: 一个指向需要存储解析后参数值的对象的指针，类型为泛型 T，T 可以是任何类型。
//
// 返回值:
//
//	error: 如果参数解析过程中发生错误，则返回错误信息，否则返回 nil。
func GetUrlParameter_Param[T any](c *gin.Context, perName string, recipient *T) error {
	// 通过 Gin 上下文的 Param 方法获取 URL 参数中的值。
	par := c.Param(perName)
	// 调用辅助函数 getUrlParameter 来进一步处理获取的参数值，将其解析并赋值给目标对象。
	// 这里的 getUrlParameter 函数是一个假设存在的函数，具体实现依赖于实际的业务逻辑。
	return getUrlParameter[T](c, par, recipient)
}

// GetUrlParameter_Query 从 HTTP 请求的查询参数中获取指定名称的参数，并将其解析到指定的接收者对象中。
// 该函数是泛型的，可以处理任何类型的接收者对象。
// 参数:
//
//	c *gin.Context: Gin 框架的上下文对象，包含 HTTP 请求和响应的相关信息。
//	perName string: 查询参数的名称，用于从请求中提取对应的参数值。
//	recipient *T: 接收解析后参数值的对象指针，T 是任何类型。
//
// 返回值:
//
//	error: 如果参数解析过程中发生错误，则返回错误信息，否则返回 nil。
func GetUrlParameter_Query[T any](c *gin.Context, perName string, recipient *T) error {
	// 使用 Gin 框架的 Query 方法从请求中获取指定名称的查询参数。
	par := c.Query(perName)
	// 调用 getUrlParameter 函数将获取的参数值解析到接收者对象中。
	// getUrlParameter 是一个泛型函数，能够处理任何类型的接收者对象。
	return getUrlParameter[T](c, par, recipient)
}

// getUrlParameter 是一个通用函数，用于将URL参数转换为指定类型。
// 它支持字符串、整数、无符号整数、浮点数、布尔值等类型的转换。
// 如果转换失败，它会根据环境（调试或生产）返回不同的错误信息。
// 参数:
//
//	c *gin.Context: Gin框架的上下文，用于处理HTTP请求和响应。
//	par string: URL参数的字符串值。
//	recipient *T: 一个指向要转换类型的指针。
//
// 返回值:
//
//	error: 如果转换过程中发生错误，返回该错误。
func getUrlParameter[T any](c *gin.Context, par string, recipient *T) error {
	var err error
	if par == "" {
		// 如果参数为空，可以选择返回错误或设置默认值
		err = errors.New("parameter is empty")
	}

	// 获取recipient的类型
	recipientType := reflect.TypeOf(recipient).Elem()

	// 创建一个零值的recipientType
	var zeroValue T
	zeroValueReflect := reflect.ValueOf(&zeroValue).Elem()

	// 根据recipientType的种类进行类型转换
	switch recipientType.Kind() {
	case reflect.String:
		zeroValueReflect.SetString(par)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(par, 10, 64)
		if err == nil {
			zeroValueReflect.SetInt(intValue)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(par, 10, 64)
		if err == nil {
			zeroValueReflect.SetUint(uintValue)
		}

	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(par, 64)
		if err == nil {
			zeroValueReflect.SetFloat(floatValue)
		}

	case reflect.Bool:
		boolValue, err := strconv.ParseBool(par)
		if err == nil {
			zeroValueReflect.SetBool(boolValue)
		}

	default:
		// 不支持的类型
		err = fmt.Errorf("unsupported type: %v", recipientType)
	}

	if err != nil {
		if core.Config.System.Env == "debug" {
			response.FailWithError(5, err, c)
		} else {
			// 在非调试环境下，仅返回错误代码。
			response.FailWithCode(5, c)
		}
		return err
	}

	// 将转换后的值写入recipient
	reflect.ValueOf(recipient).Elem().Set(zeroValueReflect)
	return nil
}

// ParameterHandle 参数处理函数
// 该函数负责对请求参数进行序列化和过滤处理
// 参数:
//
//	c *gin.Context: Gin框架的上下文，用于处理HTTP请求
//	structHandle any: 传入的结构体，用于绑定和处理请求参数
//
// 返回值:
//
//	error: 如果参数处理过程中发生错误，返回该错误
func ParameterHandle(c *gin.Context, structHandle any) error {
	// 尝试将请求体绑定到structHandle结构体，确保其为JSON格式
	err := ParameterHandleShouldBindJSON(c, structHandle)
	if err != nil {
		return err
	}
	// 对已绑定的请求参数进行过滤处理
	err = ParameterHandleFilter(c, structHandle)
	return err
}

// ParameterHandleShouldBindJSON 使用gin框架处理JSON参数绑定。
// 该函数尝试将请求体中的JSON数据绑定到指定的结构体。
// 如果绑定失败，它将根据环境决定返回详细的错误信息还是仅返回错误代码。
// 参数:
//
//	c *gin.Context: gin的上下文，用于处理HTTP请求和响应。
//	structHandle any: 将JSON数据绑定到该结构体。需要在调用前定义和初始化。
//
// 返回值:
//
//	error: 如果JSON绑定失败，则返回绑定错误；否则返回nil。
func ParameterHandleShouldBindJSON(c *gin.Context, structHandle any) error {
	// 尝试绑定JSON数据到结构体。
	bindErr := c.ShouldBindJSON(structHandle)
	if bindErr != nil {
		// 在调试环境下，返回详细的错误信息。
		if core.Config.System.Env == "debug" {
			response.FailWithError(5, bindErr, c)
		} else {
			// 在非调试环境下，仅返回错误代码。
			response.FailWithCode(5, c)
		}

		// 返回绑定错误，终止处理请求。
		return bindErr
	}
	// 如果没有绑定错误，返回nil，继续处理请求。
	return nil
}

// ParameterHandleFilter 参数处理过滤器
// 该函数用于对给定的结构体进行过滤处理，如果环境为调试模式(debug)，则返回详细的错误信息；
// 否则返回一个通用的错误代码。该函数旨在确保结构体数据的正确性和有效性。
// 参数:
//
//	c *gin.Context: Gin框架的上下文，用于处理HTTP请求和响应。
//	structHandle any: 任意类型的结构体，需要进行过滤处理。
//
// 返回值:
//
//	error: 在过滤过程中如果发生错误，返回该错误；如果没有错误，返回nil。
func ParameterHandleFilter(c *gin.Context, structHandle any) error {
	// 对结构体进行过滤处理
	filteErr := filter.FilterStruct(structHandle)
	if filteErr != nil {
		// 如果系统环境为调试模式，返回详细的错误信息
		if core.Config.System.Env == "debug" {
			response.FailWithError(5, filteErr, c)
		} else {
			// 如果系统环境不为调试模式，返回一个通用的错误代码
			response.FailWithCode(5, c)
		}
		return filteErr
	}
	return nil
}

func ResError(c *gin.Context, err error) error {
	// 如果没有错误，直接返回，不进行后续处理
	if err == nil {
		return nil
	}

	// 检查错误是否为自定义错误类型
	errorData, codeStatus := errorMessage.GetErrorData(err)
	if codeStatus {
		response.FailWithCodeMsg(errorData.GetCode(), errorData.GetMsg(), c)
	} else {
		// 获取调用者的信息
		_, file, line, ok := runtime.Caller(1)
		if !ok {
			file = "unknown"
			line = 0
		}
		// 记录错误日志，包括文件名和行号
		logrus.Errorf("Error:%v; Api:%s; at %s:%d", err, c.Request.URL.String(), file, line)
		response.FailWithCode(8, c)
	}
	return err
}

func UrlDecode(base64Str string) (string, error) {
	strBytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return "", err
	}
	str := string(strBytes)
	str, err = url.QueryUnescape(str)
	if err != nil {
		return "", err
	}
	return str, nil
}
