package datatype

import (
	"database/sql/driver"
	"fmt"
	"time"
)

/*
条件	是否写入数据库
CreatedAt = nil	❌ 不写入
CreatedAt = &CustomTime{}（零值）	✅ 写入 NULL 或 0000-00-00 00:00:00（取决于数据库配置）
CreatedAt = NewCustomTime(time.Now())	✅ 正常写入
*/

// CustomTime 时间类型，支持 nil 判断
type CustomTime time.Time

// IsZero 判断是否为空时间
func (ct *CustomTime) IsZero() bool {
	return ct == nil || time.Time(*ct).IsZero()
}

// Value 实现 driver.Valuer 接口
func (ct *CustomTime) Value() (driver.Value, error) {
	if ct == nil {
		return nil, nil
	}
	return time.Time(*ct).Format("2006-01-02 15:04:05"), nil
}

// Scan 实现 sql.Scanner 接口
func (ct *CustomTime) Scan(value interface{}) error {
	if value == nil {
		*ct = CustomTime(time.Time{})
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*ct = CustomTime(v)
	case string:
		t, err := time.Parse("2006-01-02 15:04:05", v)
		if err != nil {
			return err
		}
		*ct = CustomTime(t)
	default:
		return fmt.Errorf("unsupported type for CustomTime: %T", value)
	}
	return nil
}

// CustomDate 日期类型，支持 nil 判断
type CustomDate time.Time

// IsZero 判断是否为空日期
func (cd *CustomDate) IsZero() bool {
	return cd == nil || time.Time(*cd).IsZero()
}

// Value 实现 driver.Valuer 接口
func (cd *CustomDate) Value() (driver.Value, error) {
	if cd == nil {
		return nil, nil
	}
	return time.Time(*cd).Format("2006-01-02"), nil
}

// Scan 实现 sql.Scanner 接口
func (cd *CustomDate) Scan(value interface{}) error {
	if value == nil {
		*cd = CustomDate(time.Time{})
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*cd = CustomDate(v)
	case string:
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return err
		}
		*cd = CustomDate(t)
	default:
		return fmt.Errorf("unsupported type for CustomDate: %T", value)
	}
	return nil
}
