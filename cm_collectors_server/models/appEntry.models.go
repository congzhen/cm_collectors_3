package models

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/utils"
	"errors"

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
			ID: "coverPosterModeConvert",
			Migrate: func(tx *gorm.DB) error {
				// 首先检查 resources 表是否存在
				if !tx.Migrator().HasTable(&Resources{}) {
					// 表不存在，直接返回
					return nil
				}
				// 更新 resources 表中 coverPosterMode 字段，将非数字值替换为 0
				err := tx.Exec(`
					UPDATE resources 
					SET coverPosterMode = CASE 
						WHEN typeof(coverPosterMode) = 'integer' THEN coverPosterMode 
						WHEN length(trim(coverPosterMode, '0123456789')) < length(coverPosterMode) THEN coverPosterMode 
						ELSE '0' 
					END;
				`).Error
				if err != nil {
					core.LogErr(err)
					return err
				}
				return nil
			},
		},
		{
			ID: "initApp",
			Migrate: func(tx *gorm.DB) error {
				err := autoMigrate(tx)
				if err != nil {
					core.LogErr(err)
					return err
				}
				return nil
			},
		},
		{
			ID: "initData",
			Migrate: func(tx *gorm.DB) error {
				//获取文件集
				filesBasesDataList, err := FilesBases{}.DataList(tx)
				if err != nil {
					core.LogErr(err)
					return err
				}
				var firstFilesBasesId string
				createdAt := datatype.CustomTime(core.TimeNow())
				// 文件集不存在 创建一个文件集
				if len(*filesBasesDataList) > 0 {
					firstFilesBasesId = (*filesBasesDataList)[0].ID
				} else {
					// 创建默认文件集
					firstFilesBasesId = core.GenerateUniqueID()
					err := FilesBases{}.Create(tx, &FilesBases{ID: firstFilesBasesId, Name: "Default", Sort: 0, CreatedAt: &createdAt, Status: true})
					if err != nil {
						core.LogErr(err)
						return err
					}
				}
				// 获取文件集设置
				_, err = FilesBasesSetting{}.InfoByFilesBasesID(tx, firstFilesBasesId)
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					core.LogErr(err)
					return err
				}
				// 文件集设置不存在时，创建
				if err == gorm.ErrRecordNotFound {
					err := FilesBasesSetting{}.CreateNull(tx, firstFilesBasesId)
					if err != nil {
						core.LogErr(err)
						return err
					}
				}
				// 获取演员集
				performerBasesDataList, err := PerformerBases{}.DataList(tx)
				if err != nil {
					core.LogErr(err)
					return err
				}
				var firstPerformerBasesId string
				// 演员集不存在时，创建
				if len(*performerBasesDataList) > 0 {
					firstPerformerBasesId = (*performerBasesDataList)[0].ID
				} else {
					// 创建默认演员集
					firstPerformerBasesId = core.GenerateUniqueID()
					err := PerformerBases{}.Create(tx, &PerformerBases{ID: firstPerformerBasesId, Name: "Default", Sort: 0, CreatedAt: &createdAt, Status: true})
					if err != nil {
						core.LogErr(err)
						return err
					}
				}
				// 获取指定文件集关联的演员集
				filesRelatedPerformerBasesDataList, err := FilesRelatedPerformerBases{}.ListByFilesBasesID(tx, firstFilesBasesId)
				if err != nil {
					core.LogErr(err)
					return err
				}
				// 如果没有关联的演员集
				if len(*filesRelatedPerformerBasesDataList) == 0 {
					// 创建关联记录
					filesRelatedPerformerBasesRecords := []FilesRelatedPerformerBases{
						{ID: core.GenerateUniqueID(), FilesBasesID: firstFilesBasesId, PerformerBasesID: firstPerformerBasesId, Main: true},
					}
					err := FilesRelatedPerformerBases{}.Creates(tx, &filesRelatedPerformerBasesRecords)
					if err != nil {
						core.LogErr(err)
						return err
					}
				}
				return nil
			},
		},
		{
			ID: "update_performer_keywords",
			Migrate: func(tx *gorm.DB) error {
				// 查询所有 performer 记录
				var performers []Performer
				if err := tx.Find(&performers).Error; err != nil {
					core.LogErr(err)
					return err
				}

				for _, p := range performers {
					// 拼接 name + aliasName 的拼音首字母
					nameInitials := utils.PinyinInitials(p.Name)
					aliasInitials := utils.PinyinInitials(p.AliasName)
					keywords := nameInitials + aliasInitials

					// 更新 keywords 字段
					if err := tx.Model(&p).Update("KeyWords", keywords).Error; err != nil {
						core.LogErr(err)
						return err
					}
				}

				return nil
			},
		},
		{
			ID: "update_resources_keywords",
			Migrate: func(tx *gorm.DB) error {
				// 查询所有 performer 记录
				var resources []Resources
				if err := tx.Find(&resources).Error; err != nil {
					core.LogErr(err)
					return err
				}

				for _, p := range resources {
					//  Title 首字母
					keywords := utils.PinyinInitials(p.Title)
					// 更新 keywords 字段
					if err := tx.Model(&p).Update("KeyWords", keywords).Error; err != nil {
						core.LogErr(err)
						return err
					}
				}

				return nil
			},
		},
		{
			ID: "update_tag_keywords",
			Migrate: func(tx *gorm.DB) error {
				// 查询所有 performer 记录
				var tags []Tag
				if err := tx.Find(&tags).Error; err != nil {
					core.LogErr(err)
					return err
				}

				for _, p := range tags {
					//  Title 首字母
					keywords := utils.PinyinInitials(p.Name)
					// 更新 keywords 字段
					if err := tx.Model(&p).Update("KeyWords", keywords).Error; err != nil {
						core.LogErr(err)
						return err
					}
				}

				return nil
			},
		},
		{
			ID: "scraper",
			Migrate: func(tx *gorm.DB) error {
				err := tx.AutoMigrate(
					&FilesBasesSetting{},
					&Performer{},
					&Resources{},
				)
				if err != nil {
					core.LogErr(err)
					return err
				}
				return nil
			},
		},
		{
			ID: "resources_score",
			Migrate: func(tx *gorm.DB) error {
				err := tx.AutoMigrate(
					&Resources{},
				)
				if err != nil {
					core.LogErr(err)
					return err
				}
				return nil
			},
		},
	})
	errMigrate := m.Migrate()
	if errMigrate != nil {
		core.LogErr(errMigrate)
	}
	return errMigrate
}
