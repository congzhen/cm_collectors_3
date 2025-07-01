package models

import (
	"cm_collectors_server/utils"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func DB_Init(db *gorm.DB) error {
	RegJoinTable(db)
	return AutoDatabase(db)
}

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

func RegJoinTable(db *gorm.DB) {
	// 注册中间表模型
	db.SetupJoinTable(&Resources{}, "Tags", &ResourcesTags{})
	db.SetupJoinTable(&Resources{}, "Performers", &ResourcesPerformers{})
	db.SetupJoinTable(&Resources{}, "Directors", &ResourcesDirectors{})
	//db.SetupJoinTable(&Tag{}, "Resources", &ResourcesTags{})
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
		{
			ID: "20250625-001-update_performer_keywords",
			Migrate: func(tx *gorm.DB) error {
				// 查询所有 performer 记录
				var performers []Performer
				if err := tx.Find(&performers).Error; err != nil {
					return err
				}

				for _, p := range performers {
					// 拼接 name + aliasName 的拼音首字母
					nameInitials := utils.PinyinInitials(p.Name)
					aliasInitials := utils.PinyinInitials(p.AliasName)
					keywords := nameInitials + aliasInitials

					// 更新 keywords 字段
					if err := tx.Model(&p).Update("KeyWords", keywords).Error; err != nil {
						return err
					}
				}

				return nil
			},
		},
		{
			ID: "20250625-002-update_resources_keywords",
			Migrate: func(tx *gorm.DB) error {
				// 查询所有 performer 记录
				var resources []Resources
				if err := tx.Find(&resources).Error; err != nil {
					return err
				}

				for _, p := range resources {
					//  Title 首字母
					keywords := utils.PinyinInitials(p.Title)
					// 更新 keywords 字段
					if err := tx.Model(&p).Update("KeyWords", keywords).Error; err != nil {
						return err
					}
				}

				return nil
			},
		},
		{
			ID: "20250701-001-update_tag_keywords",
			Migrate: func(tx *gorm.DB) error {
				// 查询所有 performer 记录
				var tags []Tag
				if err := tx.Find(&tags).Error; err != nil {
					return err
				}

				for _, p := range tags {
					//  Title 首字母
					keywords := utils.PinyinInitials(p.Name)
					// 更新 keywords 字段
					if err := tx.Model(&p).Update("KeyWords", keywords).Error; err != nil {
						return err
					}
				}

				return nil
			},
		},
	})
	return m.Migrate()
}
