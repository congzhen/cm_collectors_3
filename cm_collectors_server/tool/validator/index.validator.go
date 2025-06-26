package validator

import (
	"html"
	"reflect"
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func RegValidator() {
	// 注册
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("xss", xss)
		v.RegisterValidation("uniqueID", uniqueID)
	}
}

func xss(fl validator.FieldLevel) bool {
	field := fl.Field()

	// 确保字段是字符串类型
	if field.Kind() != reflect.String {
		return false
	}
	escaped := html.EscapeString(field.String())
	return escaped == field.String()
}

func uniqueID(fl validator.FieldLevel) bool {
	field := fl.Field()
	// 确保字段是字符串类型
	if field.Kind() != reflect.String {
		return false
	}
	uniqueID := field.String()
	xidPattern := `^[0-9a-z]{20}$`
	//xidPattern := `^[0-9A-Za-z]{19,24}$`
	match, _ := regexp.MatchString(xidPattern, uniqueID)
	return match
}
