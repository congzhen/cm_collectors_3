package processors

import (
	"archive/zip"
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	"cm_collectors_server/utils"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

type DatabaseTool struct {
}

func (t DatabaseTool) GetDBBackupPath() (string, string, error) {
	backupDataPath := core.Config.System.FilePath
	// 将相对路径转换为绝对路径
	absBackupDataPath, err := filepath.Abs(backupDataPath)
	if err != nil {
		return "", "", err
	}
	// 获取backupDataPath的父目录
	parentDir := filepath.Dir(absBackupDataPath)
	// 获取backupDataPath的目录名
	dirName := filepath.Base(absBackupDataPath)
	// 创建同级目录，名字是原目录名加上"_backup"
	savePath := filepath.Join(parentDir, dirName+"_backup")
	return absBackupDataPath, savePath, nil
}

func (t DatabaseTool) GetDBBackupList() ([]string, error) {
	_, savePath, err := t.GetDBBackupPath()
	if err != nil {
		return nil, err
	}
	return utils.GetFilesByExtensions([]string{savePath}, []string{".zip"}, false)
}

func (t DatabaseTool) DBBackup() error {
	backupDataPath, savePath, err := t.GetDBBackupPath()
	if err != nil {
		return err
	}
	// 确保保存路径存在
	err = os.MkdirAll(savePath, os.ModePerm)
	if err != nil {
		return err
	}

	// 生成带时间戳的文件名
	timestamp := time.Now().Format("20060102_150405")
	zipFileName := fmt.Sprintf("db_backup_%s.zip", timestamp)
	zipFilePath := filepath.Join(savePath, zipFileName)

	// 创建zip文件
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	// 创建zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 递归遍历文件夹并添加到zip
	return filepath.Walk(backupDataPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目标备份目录本身
		if path == savePath || path == savePath+"/" {
			return nil
		}

		// 计算相对路径
		relPath, err := filepath.Rel(backupDataPath, path)
		if err != nil {
			return err
		}

		// 跳过相对路径为"."的情况（即源目录本身）
		if relPath == "." {
			return nil
		}

		// 如果是目录，则在zip中创建目录条目
		if info.IsDir() {
			_, err = zipWriter.Create(filepath.ToSlash(relPath) + "/")
			return err
		}

		// 如果是文件，则读取内容并写入zip
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// 在zip中创建文件
		writer, err := zipWriter.Create(filepath.ToSlash(relPath))
		if err != nil {
			return err
		}

		// 复制文件内容到zip
		_, err = io.Copy(writer, file)
		return err
	})
}

func (t DatabaseTool) DeleteDbBackup(fileName string) error {
	_, savePath, err := t.GetDBBackupPath()
	if err != nil {
		return err
	}
	fileName = filepath.Clean(fileName)
	filePath := filepath.Join(savePath, fileName+".zip")
	return os.Remove(filePath)
}

func (t DatabaseTool) DatabaseCleanup(filesBasesIds []string, clearItems []datatype.DatabaseCleanupClearItem) error {
	var err error
	err = t.DBBackup()
	if err != nil {
		return err
	}
	var filesBasesList *[]models.FilesBases

	if len(filesBasesIds) == 0 {
		filesBasesList, err = FilesBases{}.DataList()
		if err != nil {
			return err
		}
	} else {
		filesBasesList, err = FilesBases{}.DataListByIds(filesBasesIds)
		if err != nil {
			return err
		}
	}
	db := core.DBS()
	return db.Transaction(func(tx *gorm.DB) error {
		for _, filesBases := range *filesBasesList {
			for _, clearItem := range clearItems {
				switch clearItem {
				case datatype.ENUM_DatabaseCleanupClearItem_Resource:
					t.clear_resource(tx, &filesBases)
				case datatype.ENUM_DatabaseCleanupClearItem_Performer:
					t.clear_performer(tx, &filesBases)
				case datatype.ENUM_DatabaseCleanupClearItem_Tags:
					t.clear_tags(tx, &filesBases)
				case datatype.ENUM_DatabaseCleanupClearItem_TagClass:
					t.clear_tagClass(tx, &filesBases)
				case datatype.ENUM_DatabaseCleanupClearItem_FileDatabaseConfig:
					FilesBases{}.ClearFilesBasesConfig(tx, filesBases.ID, []datatype.Config_Field{datatype.Config_Field_FilesBases})
				case datatype.ENUM_DatabaseCleanupClearItem_ImportConfig:
					FilesBases{}.ClearFilesBasesConfig(tx, filesBases.ID, []datatype.Config_Field{datatype.Config_Field_ScanDisk})
				case datatype.ENUM_DatabaseCleanupClearItem_ResourceScraperConfig:
					FilesBases{}.ClearFilesBasesConfig(tx, filesBases.ID, []datatype.Config_Field{datatype.Config_Field_Scraper})
				case datatype.ENUM_DatabaseCleanupClearItem_PerformerScraperConfig:
					FilesBases{}.ClearFilesBasesConfig(tx, filesBases.ID, []datatype.Config_Field{datatype.Config_Field_ScraperPerformer})
				case datatype.ENUM_DatabaseCleanupClearItem_GeneralConfig:
					core.ResetConfig()
				case datatype.ENUM_DatabaseCleanupClearItem_CronJobs:
					CronJobs{}.DeleteByFilesBasesID(tx, filesBases.ID)
				}
			}
		}
		return nil
	})

}
func (DatabaseTool) clear_resource(db *gorm.DB, filesBases *models.FilesBases) error {
	return db.Transaction(func(tx *gorm.DB) error {
		coverPosterSlc, err := Resources{}.CoverPosterSlcByFilesBasesID_DB(tx, filesBases.ID)
		if err != nil {
			return err
		}
		return Resources{}.DeleteByFilesBasesID(tx, filesBases.ID, coverPosterSlc)
	})
}
func (DatabaseTool) clear_performer(db *gorm.DB, filesBases *models.FilesBases) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var PerformerBasesIds []string
		PerformerBasesPhotosMap := make(map[string][]string)
		for _, filesRelatedPerformerBases := range filesBases.FilesRelatedPerformerBases {
			PerformerBasesIds = append(PerformerBasesIds, filesRelatedPerformerBases.PerformerBasesID)
			Photos, err := Performer{}.PhotosByPerformerBasesId_DB(tx, filesRelatedPerformerBases.PerformerBasesID)
			if err != nil {
				return err
			}
			PerformerBasesPhotosMap[filesRelatedPerformerBases.PerformerBasesID] = Photos
		}
		err := Performer{}.DeleteByPerformerBasesIds(tx, PerformerBasesIds)
		if err != nil {
			return err
		}
		for performerBasesID, photos := range PerformerBasesPhotosMap {
			for _, photo := range photos {
				Performer{}.DeletePerformerPhoto(performerBasesID, photo)
			}
		}
		return nil
	})
}
func (DatabaseTool) clear_tags(db *gorm.DB, filesBases *models.FilesBases) error {
	return db.Transaction(func(tx *gorm.DB) error {
		tagData, err := Tag{}.TagData(filesBases.ID)
		if err != nil {
			return err
		}
		var tagClassSlc []string
		var tagIds []string
		for _, item := range *tagData.Tag {
			tagIds = append(tagIds, item.ID)
		}
		for _, item := range *tagData.TagClass {
			tagClassSlc = append(tagClassSlc, item.ID)
		}
		err = Tag{}.DeleteTagByTagClassSlc(tx, tagClassSlc)
		if err != nil {
			return err
		}
		return ResourcesTags{}.DeleteByTagIDS(tx, tagIds)
	})
}
func (t DatabaseTool) clear_tagClass(db *gorm.DB, filesBases *models.FilesBases) error {
	return db.Transaction(func(tx *gorm.DB) error {
		err := t.clear_tags(tx, filesBases)
		if err != nil {
			return err
		}
		return TagClass{}.DeleteByFilesBasesID_DB(tx, filesBases.ID)
	})
}
