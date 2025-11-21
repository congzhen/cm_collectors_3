package processors

import (
	cmscraper "cm_collectors_server/api/cm_scraper"
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	"cm_collectors_server/utils"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Scraper struct {
}

func (t Scraper) UpdateConfig(filesBasesId, defaultConfigJson string, field datatype.Scraper_UpdateConfig_Field) error {
	return t.UpdateConfig_DB(core.DBS(), filesBasesId, defaultConfigJson, field)
}
func (Scraper) UpdateConfig_DB(db *gorm.DB, filesBasesId, defaultConfigJson string, field datatype.Scraper_UpdateConfig_Field) error {
	switch field {
	case datatype.E_Scraper_UpdateConfig_Type_Resource:
		settingModel := models.FilesBasesSetting{
			ScraperJsonData: defaultConfigJson,
		}
		// 更新配置信息到数据库
		return settingModel.Update(db, filesBasesId, &settingModel, []string{"scraper_json_data"})
	case datatype.E_Scraper_UpdateConfig_Type_Performer:
		settingModel := models.FilesBasesSetting{
			ScraperPerformerJsonData: defaultConfigJson,
		}
		// 更新配置信息到数据库
		return settingModel.Update(db, filesBasesId, &settingModel, []string{"scraper_performer_json_data"})
	}
	return errors.New("无效的配置类型")
}

func (t Scraper) UpdatePerformerConfigByDataType(filesBasesId string, config datatype.ReqParam_SearchScraperPerformerConfig) error {
	jsonBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}
	// 转换为字符串
	configJsonString := string(jsonBytes)
	return t.UpdatePerformerConfig(filesBasesId, configJsonString)
}
func (t Scraper) UpdatePerformerConfig(filesBasesId, defaultConfigJson string) error {
	return t.UpdatePerformerConfig_DB(core.DBS(), filesBasesId, defaultConfigJson)
}
func (Scraper) UpdatePerformerConfig_DB(db *gorm.DB, filesBasesId, defaultConfigJson string) error {
	settingModel := models.FilesBasesSetting{
		ScraperPerformerJsonData: defaultConfigJson,
	}
	// 更新配置信息到数据库
	return settingModel.Update(db, filesBasesId, &settingModel, []string{"scraper_performer_json_data"})
}

func (Scraper) GetBrowserPath() string {
	if core.Config.Scraper.UseBrowserPath {
		return core.Config.Scraper.BrowserPath
	} else {
		return ""
	}
}

func (Scraper) getConfigPath(fileName string) string {
	return path.Join("./scraper", fileName+".json")
}
func (Scraper) getNfoPath(filePath string) string {
	//将路径文件的后缀名替换成nfo
	return path.Join(utils.GetDirPathFromFilePath(filePath), utils.GetFileNameFromPath(filePath, false)+".nfo")
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
	// 更新配置信息到数据库
	err = t.UpdateConfig_DB(db, filesBasesId, configJsonString, datatype.E_Scraper_UpdateConfig_Type_Resource)
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
	// 筛选待处理文件
	newFilesPaths := make([]string, 0)
	for _, filePath := range filesPaths {
		// 获取 NFO 文件路径
		nfoPath := t.getNfoPath(filePath)
		// 检查 NFO 文件是否存在
		if !utils.FileExists(nfoPath) {
			// 如果不存在，则添加到待处理列表中
			newFilesPaths = append(newFilesPaths, filePath)
		}
	}
	return newFilesPaths, nil
}

// ScraperDataProcess 对指定文件进行刮削处理，包括获取元数据、下载图片和保存NFO文件
// filesBasesId: 文件基础ID，用于标识处理的文件组
// filePath: 要处理的文件路径
// config: 刮削器配置信息
// 返回值: 处理过程中发生的错误，如果没有错误则返回nil
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
			scraperSL := cmscraper.NewScraper(scraperConfig, t.GetBrowserPath(), core.Config.Scraper.Headless, time.Duration(config.Timeout), 1, core.Config.Scraper.LogStatus, core.Config.Scraper.LogPath)
			// 关闭日志
			defer cmscraper.CloseGlobalLogger()
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
				images, err := cmscraper.GetMetadataImages(ctx, scraperConfig, pageUrl, metadata, config.UseTagAsImageName, t.GetBrowserPath(), core.Config.Scraper.Headless, core.Config.Scraper.VisitHome, config.EnableUserSimulation, 1.0)
				if err == nil && len(images) > 0 {
					// 保存图片
					for imageName, base64Data := range images {
						// 保存图片
						saveImagePath := t.getSaveImagePath(filePath, imageName)
						cmscraper.SaveBase64AsImage(base64Data, saveImagePath, true)
					}
				} else {
					processErr = fmt.Errorf("下载图片失败: %w", err)
					cmscraper.LogError(processErr.Error())
					continue
				}
			}
			// 是否 保存 NFO
			if config.SaveNfo && cmscraper.IsValidMetadata(metadata, scraperConfig) {
				saveNfoPath := t.getNfoPath(filePath)
				nfo := cmscraper.ToNFO(metadata, &scraperConfig.Sites[len(scraperConfig.Sites)-1])
				processErr = cmscraper.SaveNfoFile(saveNfoPath, []byte(nfo))
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

// ScraperPerformerDataProcess 根据演员名称从刮削源获取演员信息并更新到数据库
//
// 参数:
//   - par: 包含演员基本信息和刮削配置的参数对象
//   - PerformerBasesId: 演员库ID
//   - PerformerId: 演员ID
//   - PerformerName: 演员名称（用于刮削搜索）
//   - ScraperConfig: 使用的刮削器配置文件名
//   - Operate: 更新操作类型 ("update" 或 "cover")
//
// 返回值:
//   - error: 如果处理过程中出现错误则返回错误信息，否则返回nil
func (t Scraper) ScraperPerformerDataProcess(par *datatype.ReqParam_ScraperPerformerDataProcess) error {
	performerModels, perforomerPhotoBase64, err := t.execScraperPerformerData(par.PerformerId, par.PerformerName, par.ScraperConfig, core.Config.Scraper.Timeout)
	if err != nil {
		return err
	}
	return Performer{}.UpdateScraperByModels(par.PerformerId, performerModels, perforomerPhotoBase64, par.Operate)
}
func (t Scraper) ScraperOnePerformerDataProcess(par *datatype.ReqParam_ScraperOnePerformerDataProcess) (*models.Performer, error) {
	mode := "add"
	performerId := par.PerformerId
	if par.PerformerId != "" {
		mode = "update"
	} else {
		performerId = core.GenerateUniqueID()
	}

	performerModels, perforomerPhotoBase64, err := t.execScraperPerformerData(performerId, par.PerformerName, par.ScraperConfig, par.Timeout)
	if err != nil {
		return nil, err
	}
	if mode == "update" {
		err = Performer{}.UpdateScraperByModels(performerId, performerModels, perforomerPhotoBase64, par.Operate)
	} else {
		performerModels.PerformerBasesID = par.PerformerBases_Id
		err = Performer{}.CreateScraperByModels(performerId, performerModels, perforomerPhotoBase64)
	}
	if err != nil {
		return nil, err
	}
	return Performer{}.InfoByID(performerId)
}

func (t Scraper) execScraperPerformerData(performerId, performerName, scraperConfigName string, timeout int) (models.Performer, string, error) {
	// 加载配置
	scraperConfig, err := cmscraper.LoadConfig(t.getConfigPath(scraperConfigName))
	if err != nil {
		return models.Performer{}, "", err
	}
	// 创建刮削器
	scraperSL := cmscraper.NewScraper(scraperConfig, t.GetBrowserPath(), core.Config.Scraper.Headless, time.Duration(timeout), 1, core.Config.Scraper.LogStatus, core.Config.Scraper.LogPath)
	// 关闭日志
	defer cmscraper.CloseGlobalLogger()
	ctx := context.Background()
	metadata, pageUrl, err := scraperSL.Scrape(ctx, performerName)
	if err != nil {
		return models.Performer{}, "", err
	}
	images, err := cmscraper.GetMetadataImages(ctx, scraperConfig, pageUrl, metadata, true, t.GetBrowserPath(), core.Config.Scraper.Headless, core.Config.Scraper.VisitHome, false, 1.0)
	perforomerPhotoBase64 := ""
	if err == nil {
		// 判断image中是否有 avatar ，photo , cover ，如果有则使用第一个
		imageTags := []string{"avatar", "photo", "cover"}
		for _, tag := range imageTags {
			if image, ok := images[fmt.Sprintf("%s.jpg", tag)]; ok {
				perforomerPhotoBase64 = image
				break
			}
		}
	}

	performerModels := models.Performer{
		ID: performerId,
	}
	nameTages := []string{"name", "title"}
	aliasNameTags := []string{"aliasName", "alias", "alias_name"}
	birthdayTags := []string{"birthday", "birth", "birthday_date"}
	nationalityTags := []string{"nationality", "nation", "nation_name"}
	introductionTags := []string{"introduction", "intro", "desc", "description"}
	cupTags := []string{"cup", "cup_size", "cup_size_name"}
	bustTags := []string{"bust", "bust_size", "bust_size_name"}
	waistTags := []string{"waist", "waist_size", "waist_size_name"}
	hipTags := []string{"hip", "hip_size", "hip_size_name"}
	performerModels.Name = t.getMetadataContentByTags(metadata, nameTages)
	performerModels.AliasName = t.getMetadataContentByTags(metadata, aliasNameTags)
	performerModels.Birthday = t.getMetadataContentByTags(metadata, birthdayTags)
	performerModels.Nationality = t.getMetadataContentByTags(metadata, nationalityTags)
	performerModels.Introduction = t.getMetadataContentByTags(metadata, introductionTags)
	performerModels.Cup = t.getMetadataContentByTags(metadata, cupTags)
	performerModels.Bust = t.getMetadataContentByTags(metadata, bustTags)
	performerModels.Waist = t.getMetadataContentByTags(metadata, waistTags)
	performerModels.Hip = t.getMetadataContentByTags(metadata, hipTags)
	return performerModels, perforomerPhotoBase64, nil
}

// getMetadataContentByTags 从元数据中按标签顺序查找第一个存在的字符串值
//
// 参数:
//   - metadata: 指向包含元数据的map指针，键为字符串，值为任意类型
//   - tags: 字符串标签数组，按优先级顺序排列
//
// 返回值:
//   - string: 找到的第一个标签对应的字符串值，如果未找到则返回空字符串
func (Scraper) getMetadataContentByTags(metadata *map[string]any, tags []string) string {
	for _, tag := range tags {
		if data, ok := (*metadata)[tag]; ok {
			if dataStr, ok := data.(string); ok {
				return dataStr
			}
		}
	}
	return ""
}

// ScraperOneResourceDataProcess 处理单个资源的数据刮削
//
// 该函数根据提供的参数对单个资源进行元数据刮削处理，包括从现有NFO文件读取或通过网络刮削获取元数据，
// 提取并处理封面海报，最后创建或更新资源记录。
//
// 参数:
//   - par: 包含刮削所需参数的结构体指针，主要字段包括:
//   - ResourdId: 资源ID（为空时表示新增资源）
//   - FilesBases_ID: 文件库ID
//   - ScraperConfig: 刮削器配置名称
//   - Timeout: 刮削超时时间
//   - UseExistNfo: 是否使用现有的NFO文件
//   - SaveNfo: 是否保存NFO文件
//   - SaveImage: 是否保存图片
//   - CutPoster: 是否裁剪海报
//   - Title: 标题
//   - IssueNumber: 出版号
//   - DramaSeriesSrc: 剧集文件路径
//
// 返回值:
//   - *models.Resources: 处理完成的资源对象
//   - error: 错误信息，如果处理成功则为nil
func (t Scraper) ScraperOneResourceDataProcess(par *datatype.ReqParam_ScraperOneResourceDataProcess) (*models.Resources, error) {
	// 是否已存在NFO文件
	var useExistNfo = false
	// 文件路径
	var filePath = par.DramaSeriesSrc
	// 文件是否存在
	var filePathExists = false
	if filePath != "" {
		filePathExists = utils.FileExists(filePath)
	}
	// 元数据
	var metadata *map[string]any
	// XML 元数据
	var xmlmetadata *map[string]any
	// 元数据图片
	var metadataImages map[string]string

	//配置文件ScanDisk json解构
	configScanDisk, err := FilesBases{}.Config_ScanDisk(par.FilesBases_ID)
	if err != nil {
		return nil, err
	}

	//是否可以直接调用已存在的nfo
	if par.UseExistNfo && par.DramaSeriesSrc != "" {
		// 获取 NFO 文件路径
		nfoPath := t.getNfoPath(par.DramaSeriesSrc)
		// 读取 NFO 文件内容
		nfoBytes, err := utils.ReadFile(nfoPath)
		if err == nil && nfoBytes != nil {
			useExistNfo = true
			// 创建XML解码器
			decoder := xml.NewDecoder(strings.NewReader(string(nfoBytes)))
			// 解析为通用结构
			rootElement, err := utils.XML_parseXMLToMap(decoder)
			if err != nil {
				return nil, err
			}
			xmlmetadata = &rootElement
		}
	}
	// 如果不使用已存在的NFO文件，则进行刮削
	if !useExistNfo {
		// 加载配置
		scraperConfig, err := cmscraper.LoadConfig(t.getConfigPath(par.ScraperConfig))
		if err != nil {
			return nil, err
		}
		// 创建刮削器
		scraperSL := cmscraper.NewScraper(scraperConfig, t.GetBrowserPath(), core.Config.Scraper.Headless, time.Duration(par.Timeout), 1, core.Config.Scraper.LogStatus, core.Config.Scraper.LogPath)
		// 关闭日志
		defer cmscraper.CloseGlobalLogger()
		var scrapeIDS = [3]string{}
		if par.IssueNumber != "" {
			scrapeIDS[0] = par.IssueNumber
		}
		if par.Title != "" {
			scrapeIDS[1] = par.Title
		}
		if filePathExists {
			scrapeIDS[2] = cmscraper.ParseID(par.DramaSeriesSrc, scraperConfig)
		}
		var pageUrl string = ""
		// 刮削数据
		ctx := context.Background()
		for _, id := range scrapeIDS {
			if id == "" {
				continue
			}
			metadata, pageUrl, err = scraperSL.Scrape(ctx, id)
			if err == nil {
				break
			}
		}
		if err != nil {
			return nil, err
		}
		// 检查元数据是否有效
		if !cmscraper.IsValidMetadata(metadata, scraperConfig) {
			return nil, errors.New("没有找到匹配的元数据")
		}
		// 获取图片
		metadataImages, err = cmscraper.GetMetadataImages(ctx, scraperConfig, pageUrl, metadata, true, t.GetBrowserPath(), core.Config.Scraper.Headless, core.Config.Scraper.VisitHome, false, 1.0)
		if err == nil && len(metadataImages) > 0 && filePathExists && par.SaveImage {
			// 保存图片
			for imageName, base64Data := range metadataImages {
				saveImagePath := t.getSaveImagePath(filePath, imageName)
				cmscraper.SaveBase64AsImage(base64Data, saveImagePath, true)
			}
		}
		// 生成 NFO
		nfo := cmscraper.ToNFO(metadata, &scraperConfig.Sites[len(scraperConfig.Sites)-1])
		// 创建XML解码器
		decoder := xml.NewDecoder(strings.NewReader(nfo))
		// 解析为通用结构
		rootElement, err := utils.XML_parseXMLToMap(decoder)
		if err != nil {
			return nil, err
		}
		xmlmetadata = &rootElement
		// 保存 NFO
		if filePathExists && par.SaveNfo && cmscraper.IsValidMetadata(metadata, scraperConfig) {
			saveNfoPath := t.getNfoPath(filePath)
			err := cmscraper.SaveNfoFile(saveNfoPath, []byte(nfo))
			if err != nil {
				return nil, err
			}
		}
	}

	var coverPosterBase64 string
	var coverPosterWidth int
	var coverPosterHeight int
	// 是否已存在本地nfo文件
	if useExistNfo {
		// 获取封面海报- 获取本地封面海报
		coverPosterBase64, coverPosterWidth, coverPosterHeight, err = ImportData{}.GetCoverPosterBase64(filePath, configScanDisk)
	} else {
		// 获取封面海报- 获取元数据中的封面海报
		if len(configScanDisk.CoverPosterMatchName) > 0 && len(metadataImages) > 0 {
			// 匹配封面海报
			for _, matchName := range configScanDisk.CoverPosterMatchName {
				for imageName, _ := range metadataImages {
					// 模糊匹配
					if imageName == matchName || strings.Contains(imageName, matchName) || strings.Contains(matchName, imageName) {
						coverPosterBase64 = metadataImages[imageName]
						break
					}
				}
			}
		}
		if coverPosterBase64 != "" {
			// 如果封面海报存在，转成bytes
			coverPosterBytes, err := utils.Base64ToBytes(coverPosterBase64, true)
			if err == nil {
				// 获取图片尺寸
				coverPosterWidth, coverPosterHeight, err = utils.GetImageDimensionsFromBytes(coverPosterBytes)
				if err != nil {
					coverPosterWidth = 200
					coverPosterHeight = 200
				}
			}
			// 如果设置了封面海报类型，则进行裁剪
			if configScanDisk.CoverPosterType >= 0 {
				coverPosterWidth = configScanDisk.CoverPosterWidth
				coverPosterHeight = configScanDisk.CoverPosterHeight

				if par.CutPoster {
					// 裁剪图片
					cropImageBytes, err := utils.ResizeAndCropImage(coverPosterBytes, coverPosterWidth, coverPosterHeight)
					if err == nil {
						coverPosterWidth = configScanDisk.CoverPosterWidth
						coverPosterHeight = configScanDisk.CoverPosterHeight
						// 转成base64
						cropImageBytesBase64, err := utils.ImageBytesToBase64(cropImageBytes)
						if err == nil {
							// 设置裁切后的图片base64
							coverPosterBase64 = cropImageBytesBase64
						}
					}
				}
			}

		}

	}
	// 资源标题
	resourceTitle := ImportData{}.GetResourceTitle(filePath, configScanDisk)
	lastScraperUpdateTime := datatype.CustomDate{}
	lastScraperUpdateTime.SetValue(core.TimeNow())
	resourceDataParam := datatype.ReqParam_Resource{
		Resource: datatype.ReqParam_ResourceBase{
			ID:                    par.ResourdId,
			FilesBasesID:          par.FilesBases_ID,
			Title:                 resourceTitle,
			Mode:                  datatype.E_resourceMode_Movies,
			CoverPosterMode:       configScanDisk.CoverPosterType,
			CoverPosterWidth:      coverPosterWidth,
			CoverPosterHeight:     coverPosterHeight,
			LastScraperUpdateTime: &lastScraperUpdateTime,
		},
		PhotoBase64: coverPosterBase64,
		DramaSeries: []datatype.ReqParam_resourceDramaSeries_Base{
			{Src: filePath},
		},
	}
	if configScanDisk.AutoCreatePoster {
		resourceDataParam.Resource.Definition = ImportData{}.VideoDefinition(filePath, configScanDisk)
	}

	ImportData{}.NfoExecData(par.FilesBases_ID, filePath, *xmlmetadata, configScanDisk.Nfo, &resourceDataParam)
	if resourceDataParam.Resource.Title == "" {
		return nil, errors.New("刮削数据失败")
	}

	if par.ResourdId == "" {
		//添加新资源
		return Resources{}.CreateResource(&resourceDataParam)
	} else {
		//如果是更新操作
		if par.Operate == datatype.E_PerformerUpdateOperate_Update {
			// 读取原始资源信息
			resourceInfo, err := Resources{}.Info(par.ResourdId)
			if err == nil {
				//覆盖信息
				if resourceInfo.Title != "" {
					resourceDataParam.Resource.Title = resourceInfo.Title
				}
				if resourceInfo.IssueNumber != "" {
					resourceDataParam.Resource.IssueNumber = resourceInfo.IssueNumber
				}
				if resourceInfo.IssuingDate.IsZero() == false {
					issuingDate, err := resourceInfo.IssuingDate.Value()
					if err == nil {
						resourceDataParam.Resource.IssuingDate = issuingDate.(string)
					}
				}
				if resourceInfo.Country != "" {
					resourceDataParam.Resource.Country = resourceInfo.Country
				}
				if resourceInfo.Definition != "" {
					resourceDataParam.Resource.Definition = resourceInfo.Definition
				}
				if resourceInfo.Stars != 0 {
					resourceDataParam.Resource.Stars = resourceInfo.Stars
				}
				if resourceInfo.Abstract != "" {
					resourceDataParam.Resource.Abstract = resourceInfo.Abstract
				}
				if len(resourceInfo.Tags) > 0 {
					for _, tag := range resourceInfo.Tags {
						// 判断tag.id 是否在 resourceDataParam.Tags 中，不在则添加
						if !utils.IsElementInArray(resourceDataParam.Tags, tag.ID) {
							resourceDataParam.Tags = append(resourceDataParam.Tags, tag.ID)
						}
					}
				}
				if len(resourceInfo.Performers) > 0 {
					for _, performer := range resourceInfo.Performers {
						// 判断performer.id 是否在 resourceDataParam.Performers 中，不在则添加
						if !utils.IsElementInArray(resourceDataParam.Performers, performer.ID) {
							resourceDataParam.Performers = append(resourceDataParam.Performers, performer.ID)
						}
					}
				}
				if len(resourceInfo.Directors) > 0 {
					for _, director := range resourceInfo.Directors {
						// 判断director.id 是否在 resourceDataParam.Directors 中，不在则添加
						if !utils.IsElementInArray(resourceDataParam.Directors, director.ID) {
							resourceDataParam.Directors = append(resourceDataParam.Directors, director.ID)
						}
					}
				}
			}
		}
		//修改资源
		return Resources{}.UpdateResource(&resourceDataParam, false)
	}
}
