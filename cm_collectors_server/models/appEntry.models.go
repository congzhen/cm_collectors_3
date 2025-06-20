package models

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// AutoMigrate 自动迁移数据库表结构，确保数据库中的表与指定的模型结构一致。
// 参数:
//
//	db - *gorm.DB: GORM 数据库连接实例，用于执行迁移操作。
//
// 返回值:
//
//	error: 如果迁移过程中发生错误，则返回错误信息；否则返回 nil。
func autoMigrate(db *gorm.DB) error {
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

// AutoDatabase 执行数据库迁移。
// 该函数使用 gormigrate 包来管理数据库模式的迁移，确保数据库结构与应用保持同步。
// 参数:
//
//	db *gorm.DB - 一个指向 gorm.DB 实例的指针，用于执行数据库操作。
//
// 返回值:
//
//	error - 如果迁移过程中发生错误，则返回该错误。
func AutoDatabase(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "20250620-001-initApp",
			Migrate: func(tx *gorm.DB) error {
				return autoMigrate(tx)
			},
		},
	})
	return m.Migrate()
}
