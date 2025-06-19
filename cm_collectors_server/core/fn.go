package core

import (
	"time"

	"gorm.io/gorm"
)

// DBS 返回一个 *gorm.DB 类型的数据库连接实例。
// 该函数没有输入参数。
// 返回值是 *gorm.DB 类型，是数据库操作的基础对象。
func DBS() *gorm.DB {
	return dbs()
}

// GenerateUniqueID 生成一个唯一的标识符。
//
// 该函数通过调用 getXid 函数来生成一个唯一的标识符。这个标识符可以用于
// 各种目的，比如作为数据库记录的主键、会话标识符等。使用 getXid 而不是
// 直接在该函数中实现生成逻辑，可以提高代码的可维护性和可测试性。
func GenerateUniqueID() string {
	return getXid()
}

// TimeNow 返回当前的系统时间。
// 该函数无需参数。
// 返回值为当前的系统时间，类型为time.Time。
func TimeNow() time.Time {
	return time.Now()
}

// TimeNowSecond 返回当前时间，精确到秒。
// 该函数通过调用 time.Now().Truncate(time.Second) 来获取当前时间并将其截断到最近的秒。
// 这样做是为了去除时间中的毫秒和纳秒部分，确保时间值仅包含年、月、日、时、分、秒。
// 返回值: 当前时间，精确到秒。
func TimeNowSecond() time.Time {
	return time.Now().Truncate(time.Second)
}

// TimeParse 根据指定的布局和时间值解析时间
// 该函数旨在简化时间解析过程，通过封装 time.ParseInLocation 函数的调用，
// 使其更易于使用和阅读。它总是使用本地时间进行解析。
//
// 参数:
//
//	layout: 时间的布局字符串，遵循 time 包中定义的格式。
//	value: 需要解析的时间值字符串，应与 layout 参数指定的格式匹配。
//
// 返回值:
//
//	time.Time: 解析后的时间对象，如果解析失败则为零值。
//	error: 如果解析过程中发生错误，返回错误信息。
func TimeParse(layout string, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, time.Local)
}
