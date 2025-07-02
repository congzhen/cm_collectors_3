package utils

import (
	"fmt"
	"reflect"
)

// ExtractArrayField 提取数组中的指定字段。
// 该函数接收一个泛型数组和一个函数，该函数用于从数组的每个元素中提取所需字段。
// 参数:
//
//	items: 一个泛型数组，表示需要处理的项集合。
//	field: 一个函数，用于从数组的每个元素中提取所需字段。
//
// 返回值:
//
//	一个包含从每个元素中提取的字段值的切片。
func ExtractArrayField[T any, V any](items []T, field func(T) V) []V {
	// 初始化结果切片，用于存储提取的字段值。
	result := make([]V, 0, len(items))
	// 遍历数组中的每个元素。
	for _, item := range items {
		// 使用提供的函数从当前元素中提取字段值，并将其添加到结果切片中。
		result = append(result, field(item))
	}
	// 返回包含所有提取字段值的切片。
	return result
}

// 定义一个泛型函数，用于判断给定元素是否存在于数组中
// IsElementInArray 检查一个给定的元素是否存在于一个数组中。
// 这个函数采用泛型编程的方式，可以处理任何可比较类型的数组和元素。
// 参数:
//
//	arr - 一个泛型数组，包含待检查的元素。
//	elem - 需要查找的元素。
//
// 返回值:
//
//	如果数组中存在该元素，则返回true；否则返回false。
func IsElementInArray[T comparable](arr []T, elem T) bool {
	// 遍历数组中的每个元素，与给定元素进行比较。
	for _, item := range arr {
		// 如果找到匹配的元素，则立即返回true。
		if item == elem {
			return true
		}
	}
	// 如果遍历完数组后没有找到匹配的元素，则返回false。
	return false
}

// SliceToMap 将切片转换为映射，使用反射从每个元素中提取键。
// arr 是输入的切片。
// keyField 是切片元素结构体中用作键的字段名称。
// 返回值是一个映射，键是从元素中提取的字段值，值是原始元素。
func SliceToMap[T any](arr []T, keyField string) (map[string]T, error) {
	result := make(map[string]T)
	for _, item := range arr {
		value := reflect.ValueOf(item)
		field := value.FieldByName(keyField)
		if !field.IsValid() {
			return nil, fmt.Errorf("field %s not found in element", keyField)
		}
		result[field.String()] = item
	}
	return result, nil
}

// ReverseArray 反转一个包含任何类型元素的数组。
// 该函数采用泛型数组作为输入参数，允许处理任何类型的数组。
// 参数:
//
//	arr []T: 待反转的数组，其中T可以是任何类型。
//
// 返回值:
//
//	[]T: 反转后的数组。
func ReverseArray[T any](arr []T) []T {
	// 获取数组长度
	n := len(arr)
	// 通过交换数组对称位置的元素来实现反转
	for i := 0; i < n/2; i++ {
		// 将对称位置的两个元素进行交换
		arr[i], arr[n-1-i] = arr[n-1-i], arr[i]
	}

	// 返回反转后的数组
	return arr
}

// InSliceByFieldGeneric 通过反射检查一个值的指定字段是否存在于给定的切片中。
// 参数:
// slice: 要检查的切片。
// fieldName: 要比较的字段名。
// value: 要查找的值。
// 返回值:
// bool: 如果找到则返回true，否则返回false。
// 该函数支持任意类型的切片和字段值，利用了反射来动态获取字段值进行比较。
// 示例:
// InSliceByFieldGeneric(userList, "Id", 2) // 输出: true or false
func InSliceByFieldGeneric[T any](slice []T, fieldName string, value interface{}) bool {
	// 获取待比较值的反射值
	fieldVal := reflect.ValueOf(value)
	// 遍历切片中的每个元素
	for _, elem := range slice {
		// 获取元素的反射值
		elemVal := reflect.ValueOf(elem)
		// 如果元素是指针，则获取指向的值
		if elemVal.Kind() == reflect.Ptr {
			elemVal = elemVal.Elem()
		}
		// 尝试获取元素中指定名称的字段
		field := elemVal.FieldByName(fieldName)
		// 如果字段存在且字段值与待比较值相等，则返回true
		if field.IsValid() && field.Interface() == fieldVal.Interface() {
			return true
		}
	}
	// 如果遍历完成后没有找到匹配的字段值，则返回false
	return false
}

// ArrayIntersectDiff 计算两个切片的交集、差集1和差集2。
// 其中，交集是同时存在于slice1和slice2中的元素组成的切片。
// 差集1是存在于slice1中但不存在于slice2中的元素组成的切片。
// 差集2是存在于slice2中但不存在于slice1中的元素组成的切片。
// 参数slice1和slice2是待比较的两个切片，类型T必须实现comparable接口。
// 返回值是交集、差集1和差集2组成的三个切片[交集, 第一个切片所特有的元素, 第二个切片所独有的元素]。[更新，删除，添加]
func ArrayIntersectDiff[T comparable](slice1, slice2 []T) ([]T, []T, []T) {
	// 初始化交集、差集1和差集2的切片。
	intersection := make([]T, 0)
	diff1 := make([]T, 0)
	diff2 := make([]T, 0)

	// 创建一个映射，用于快速检查slice1中的元素是否存在于slice2中。
	// 创建一个映射以存储slice1中元素的出现情况，用于O（1）查找。
	elemMap := make(map[T]bool)
	for _, v := range slice1 {
		elemMap[v] = true
	}

	// 遍历slice2，找出交集和差集2。
	// 计算交集和差值。
	for _, v := range slice2 {
		if elemMap[v] {
			intersection = append(intersection, v)
			elemMap[v] = false // 标记为false，表示该元素已被处理，避免重复计数。
		} else {
			diff2 = append(diff2, v)
		}
	}

	// 遍历slice1，找出差集1。
	// slice1中不在slice2中的元素现在在elemMap中标记为false。
	for _, v := range slice1 {
		if elemMap[v] {
			diff1 = append(diff1, v)
		}
	}

	// 返回交集、差集1和差集2。
	return intersection, diff1, diff2
}
