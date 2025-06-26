package filter

import (
	"fmt"
	"reflect"
	"strings"
)

var filterFuncMap = map[string]func(string) string{
	"SafeHTML": SafeHTML,
}

// FilterStruct 遍历结构体并应用过滤器
//
// 参数:
// obj interface{} - 需要遍历并应用过滤器的结构体指针。
//
// 返回值:
// error - 如果遍历或应用过滤器过程中发生错误，将返回error；否则返回nil。
//
// 使用方法：其会查询结构体中filter，并使用其中的函数进行过滤。可以在filter包中添加更多的过滤器函数，并将新过滤器函数添加到filterFuncMap中。
//
//		type parc struct {
//			Title    string `json:"title" binding:"required" filter:"SafeHTML"`
//			SubTitle string `json:"sub_title" filter:"SafeHTML"`
//			ColumnId string `json:"column_id" binding:"required" filter:"SafeHTML"`
//			Cotnent  string `json:"content" binding:"required" filter:"SafeHTML"`
//	        dArray  []string `json:"d_array" binding:"required" filter:"SafeHTML"`
//		}
//		var par parc
//		err := filter.FilterStruct(&par)
func FilterStruct(obj interface{}) error {
	// 获取对象的反射值
	value := reflect.ValueOf(obj)

	// 检查是否是指针，并且指向的是一个结构体
	if value.Kind() != reflect.Ptr || value.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to a struct, got %s", value.Type())
	}

	// 获取指向的结构体的反射值
	value = value.Elem()

	// 遍历结构体的所有字段
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		tags := value.Type().Field(i).Tag.Get("filter")

		// 如果字段有 filter 标签，则进行过滤操作
		if tags != "" {
			tagList := strings.Split(tags, ",")

			// 遍历标签列表
			for _, tag := range tagList {
				if funcName, ok := filterFuncMap[tag]; ok {
					// 根据字段类型进行不同的处理
					switch field.Kind() {
					case reflect.String:
						// 字段为 string 类型
						fieldValue := field.Addr().Interface().(*string)
						*fieldValue = funcName(*fieldValue)
					case reflect.Slice:
						// 字段为 slice 类型
						if field.Type().Elem().Kind() != reflect.String {
							return fmt.Errorf("filter function expected slice of string field, got %s", field.Type())
						}
						sliceValue := field.Addr().Interface().(*[]string)
						*sliceValue = applyFilterToSlice(*sliceValue, funcName)
					default:
						// 不支持的字段类型
						return fmt.Errorf("unsupported field type: %s", field.Type())
					}
				} else {
					// 未找到对应的过滤函数
					return fmt.Errorf("no such filter function for tag: %s", tag)
				}
			}
		}

		// 处理嵌套结构体
		if field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct {
			// 如果当前字段是结构体指针类型
			// field 是当前字段的反射值
			// field.Kind() 为 reflect.Ptr 表示这是一个指针类型
			// field.Elem() 获取指针指向的值的反射值
			// field.Elem().Kind() 为 reflect.Struct 表示指针指向的是一个结构体

			// 调用 FilterStruct 函数递归处理嵌套的结构体
			err := FilterStruct(field.Addr().Interface())
			if err != nil {
				// 如果处理过程中出现错误，则立即返回错误
				return err
			}
		} else if field.Kind() == reflect.Struct {
			// 如果当前字段是直接嵌套的结构体类型
			// field 是当前字段的反射值
			// field.Kind() 为 reflect.Struct 表示这是一个结构体

			// 调用 FilterStruct 函数递归处理嵌套的结构体
			// 注意这里需要将 field 地址化，因为 FilterStruct 需要一个指向结构体的指针
			err := FilterStruct(field.Addr().Interface())
			if err != nil {
				// 如果处理过程中出现错误，则立即返回错误
				return err
			}
		}

	}

	// 所有字段处理完毕，返回无错误
	return nil
}

// applyFilterToSlice 应用过滤函数到字符串切片
func applyFilterToSlice(slice []string, filterFunc func(string) string) []string {
	// 遍历字符串切片并应用过滤函数
	for i, item := range slice {
		slice[i] = filterFunc(item)
	}
	// 返回处理后的切片
	return slice
}
