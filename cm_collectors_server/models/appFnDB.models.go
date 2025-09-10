package models

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// BatchUpdate 批量更新指定字段
//
//	示例：BatchUpdate(db, "tags", "id", []string{"sort"}, tagSortItems, func(item TagSort) map[string]interface{} {
//	  return map[string]interface{}{"sort": item.Sort}
//	})
func BatchUpdate[T any](db *gorm.DB, table string, pkField string, fields []string, items []T, toMap func(T) map[string]interface{}) error {
	if len(items) == 0 {
		return nil
	}

	// 构建 CASE WHEN SQL 片段
	setClauses := make([]string, len(fields))
	params := make([]interface{}, 0)

	for i, field := range fields {
		caseClauses := make([]string, len(items))
		for j, item := range items {
			vals := toMap(item)
			caseClauses[j] = "WHEN ? THEN ?"
			params = append(params, vals[pkField], vals[field])
		}
		setClause := fmt.Sprintf("%s = CASE %s %s END", field, pkField, strings.Join(caseClauses, " "))
		setClauses[i] = setClause
	}

	// 构建主键 IN 条件
	ids := make([]interface{}, len(items))
	for i, item := range items {
		ids[i] = toMap(item)[pkField]
	}
	whereClause := fmt.Sprintf("%s IN (?)", pkField)

	// 组装最终 SQL
	sql := fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, strings.Join(setClauses, ", "), whereClause)
	params = append(params, ids)

	// 执行 SQL
	return db.Exec(sql, params...).Error
}
