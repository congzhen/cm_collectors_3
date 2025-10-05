package processors

import (
	cmscraper "cm_collectors_server/api/cm_scraper"
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	"cm_collectors_server/utils"
	"context"
	"encoding/json"
	"fmt"
	"path"
	"time"
)

type Scraper struct {
}

func (Scraper) getConfigPath(fileName string) string {
	return path.Join("./scraper", fileName+".json")
}
func (Scraper) getNfoPath(filePath, nfoName string) string {
	return path.Join(utils.GetDirPathFromFilePath(filePath), nfoName+".nfo")
}
func (Scraper) getSaveImagePath(filePath, imageName string) string {
	return path.Join(utils.GetDirPathFromFilePath(filePath), imageName)
}

// Pretreatment 预处理函数，用于获取待刮削的文件列表并保存配置信息
//
// 参数:
//   - filesBasesId: 文件库ID，用于关联配置信息
//   - config: 刮削器配置对象，包含扫描路径、文件后缀等配置项
//
// 返回值:
//   - []string: 待处理的文件路径列表
//   - error: 错误信息，如果处理过程中发生错误则返回相应错误

func (t Scraper) Pretreatment(filesBasesId string, config datatype.Config_Scraper) ([]string, error) {
	// 获取待处理文件列表
	filesPaths, err := utils.GetFilesByExtensions(config.ScanDiskPaths, config.VideoSuffixName, true)
	if err != nil {
		return nil, err
	}
	// 按时间排序
	filesPaths = utils.SortFilesByOrder(filesPaths, utils.FileTimeAsc)
	db := core.DBS()
	jsonBytes, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	// 转换为字符串
	configJsonString := string(jsonBytes)
	settingModel := models.FilesBasesSetting{
		ScraperJsonData: configJsonString,
	}
	// 更新配置信息到数据库
	err = settingModel.Update(db, filesBasesId, &settingModel, []string{"scraper_json_data"})
	if err != nil {
		return nil, err
	}
	// 如果设置为跳过已存在NFO的文件，则直接返回待处理文件列表
	if !config.SkipIfNfoExists {
		return filesPaths, nil
	}
	// 如果没有配置 scraper_configs，则直接返回待处理文件列表
	if len(config.ScraperConfigs) == 0 {
		return filesPaths, nil
	}
	// 获取配置文件路径
	configPath := t.getConfigPath(config.ScraperConfigs[0])
	// 加载配置
	scraperConfig, err := cmscraper.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	// 筛选待处理文件
	newFilesPaths := make([]string, 0)
	for _, filePath := range filesPaths {
		// 获取 NFO 文件名
		nfoName := cmscraper.ParseID(filePath, scraperConfig)
		// 获取 NFO 文件路径
		nfoPath := t.getNfoPath(filePath, nfoName)
		// 检查 NFO 文件是否存在
		if !utils.FileExists(nfoPath) {
			// 如果不存在，则添加到待处理列表中
			newFilesPaths = append(newFilesPaths, filePath)
		}
	}
	return newFilesPaths, nil
}

func (t Scraper) ScraperDataProcess(filesBasesId, filePath string, config datatype.Config_Scraper) error {
	var processErr error = nil
	for _, scraperConfigName := range config.ScraperConfigs {
		// 重试次数
		retryCount := config.RetryCount
		for i := 0; i < retryCount; i++ {
			// 加载配置
			scraperConfig, err := cmscraper.LoadConfig(t.getConfigPath(scraperConfigName))
			processErr = err
			if err != nil {
				continue
			}
			// 创建刮削器
			scraperSL := cmscraper.NewScraper(scraperConfig, time.Duration(5), 1)
			id := cmscraper.ParseID(filePath, scraperConfig)
			ctx := context.Background()
			metadata, pageUrl, err := scraperSL.Scrape(ctx, id)
			processErr = err
			if err != nil {
				continue
			}
			// 是否下载图片
			if config.EnableDownloadImages {
				// 获取元数据的base64图片数据map
				images, err := cmscraper.GetMetadataImages(ctx, pageUrl, metadata, config.UseTagAsImageName, config.EnableUserSimulation, 1.0)
				if err == nil && len(images) > 0 {
					// 保存图片
					for imageName, base64Data := range images {
						// 保存图片
						saveImagePath := t.getSaveImagePath(filePath, imageName)
						if err := utils.SaveBase64AsImage(base64Data, saveImagePath, true); err != nil {
							fmt.Println("保存图片失败 %s: %v", imageName, err)
						}
					}
				}
			}
			// 是否 保存 NFO
			if config.SaveNfo && cmscraper.IsValidMetadata(metadata, scraperConfig) {
				saveNfoPath := t.getNfoPath(filePath, id)
				nfo := cmscraper.ToNFO(metadata, &scraperConfig.Sites[len(scraperConfig.Sites)-1])
				processErr = utils.WriteStringToFile(saveNfoPath, nfo)
			}
			// 如果没有错误,且元数据里有数据，则返回
			if processErr == nil && cmscraper.IsValidMetadata(metadata, scraperConfig) {
				return nil
			}
			// 等待重试
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}
	return processErr
}
