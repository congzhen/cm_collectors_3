package response

import (
	"cm_collectors_server/core"
)

type ResErrorCodeMsg map[int]string

func GetErrorMsg(code int) string {
	msg, ok := ErrorMap[code][getLanguage()]
	if !ok {
		msg = ErrorMap[1][getLanguage()]
	}
	return msg
}
func getLanguage() string {
	switch core.Config.System.ResponseMsgLanguage {
	case "zh":
		return "ZH"
	case "en":
		return "EN"
	default:
		return "ZH"
	}
}

var ErrorMap = map[int]map[string]string{
	1: {
		"ZH": "未知错误",
		"EN": "Unknown error",
	},
	2: {
		"ZH": "系统错误",
		"EN": "System error",
	},
	3: {
		"ZH": "用户未登录",
		"EN": "User not logged in",
	},
	4: {
		"ZH": "模型数据错误",
		"EN": "Models data error",
	},
	5: {
		"ZH": "参数错误",
		"EN": "Parameter error",
	},
	6: {
		"ZH": "事务错误",
		"EN": "Transaction error",
	},
	100: {
		"ZH": "创建Token失败",
		"EN": "Failed to create Token",
	},
	101: {
		"ZH": "令牌无效或过期",
		"EN": "Token invalid or expired",
	},
}
