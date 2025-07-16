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

func (ct *CustomTime) MarshalJSON() ([]byte, error) {
	if ct == nil || time.Time(*ct).IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", time.Time(*ct).Format("2006-01-02 15:04:05"))), nil
}

// SetValue 设置 CustomTime 的值，支持 string 或 time.Time 类型，空字符串或 nil 将被设置为零值
func (ct *CustomTime) SetValue(value interface{}) error {
	if value == nil {
		*ct = CustomTime(time.Time{})
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*ct = CustomTime(v)
	case string:
		if v == "" {
			*ct = CustomTime(time.Time{})
			return nil
		}
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

// Value 实现 driver.Valuer 接口
func (ct *CustomTime) Value() (driver.Value, error) {
	if ct.IsZero() {
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
			*ct = CustomTime(time.Time{}) // 设置为零值
			return nil
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

func (cd *CustomDate) MarshalJSON() ([]byte, error) {
	if cd == nil || time.Time(*cd).IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", time.Time(*cd).Format("2006-01-02"))), nil
}

// SetValue 设置 CustomDate 的值，支持 string 或 time.Time 类型，空字符串或 nil 将被设置为零值
func (cd *CustomDate) SetValue(value interface{}) error {
	if value == nil {
		*cd = CustomDate(time.Time{})
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*cd = CustomDate(v)
	case string:
		if v == "" {
			*cd = CustomDate(time.Time{})
			return nil
		}
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

// Value 实现 driver.Valuer 接口
func (cd *CustomDate) Value() (driver.Value, error) {
	if cd.IsZero() {
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
			*cd = CustomDate(time.Time{}) // 设置为零值
			return nil
		}
		*cd = CustomDate(t)
	default:
		return fmt.Errorf("unsupported type for CustomDate: %T", value)
	}
	return nil
}
