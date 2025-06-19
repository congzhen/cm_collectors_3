package models

import "gorm.io/gorm"

// AutoMigrate 自动迁移数据库表结构，确保数据库中的表与指定的模型结构一致。
// 参数:
//   db - *gorm.DB: GORM 数据库连接实例，用于执行迁移操作。
// 返回值:
//   error: 如果迁移过程中发生错误，则返回错误信息；否则返回 nil。
func AutoMigrate(db *gorm.DB) error {
	// 调用 GORM 的 AutoMigrate 方法，自动创建或更新数据库表结构
	return db.AutoMigrate(
		&FilesBases{},
		&FilesBasesSetting{},
		&FilesRelatedPerformerBases{},
		&Performer{},
		&PerformerBases{},
		&Resources{},
		&ResourcesDirectors{},
		&ResourcesDramaSeries{},
		&ResourcesPerformers{},
		&ResourcesTags{},
		&Tag{},
		&TagClass{},
	)
}
