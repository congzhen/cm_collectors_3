package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	processorsffmpeg "cm_collectors_server/processorsFFmpeg"
	"cm_collectors_server/utils"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type ImportData struct {
}

func (ImportData) UpdateScanDiskConfig(filesBasesId, defaultConfigJson string) error {
	settingModel := models.FilesBasesSetting{
		ScanDiskJsonData: defaultConfigJson,
	}
	return settingModel.Update(core.DBS(), filesBasesId, &settingModel, []string{"scan_disk_json_data"})
}

// ScanDiskImportPaths 扫描磁盘路径中的视频文件，并更新扫描配置
//
// 该函数会扫描指定路径下的所有视频文件，将扫描配置保存到数据库，
// 并过滤出尚未在数据库中记录的新文件路径。
//
// 参数:
//   - filesBasesId: 文件库ID，用于关联扫描配置和过滤已存在的文件
//   - config: 磁盘扫描配置，包含扫描路径和视频文件后缀等信息
//   - saveConfig: 是否保存配置
//
// 返回值:
//   - []string: 不存在于数据库中的新文件路径列表
//   - error: 执行过程中可能出现的错误
func (ImportData) ScanDiskImportPaths(filesBasesId string, config datatype.Config_ScanDisk, saveConfig bool) ([]string, error) {
	filesPaths, err := utils.GetFilesByExtensions(config.ScanDiskPaths, config.VideoSuffixName, true)
	if err != nil {
		return nil, err
	}
	filesPaths = utils.SortFilesByOrder(filesPaths, utils.FileTimeAsc)
	db := core.DBS()

	if saveConfig {
		jsonBytes, err := json.Marshal(config)
		if err != nil {
			return nil, err
		}
		// 转换为字符串
		configJsonString := string(jsonBytes)
		settingModel := models.FilesBasesSetting{
			ScanDiskJsonData: configJsonString,
		}
		err = settingModel.Update(db, filesBasesId, &settingModel, []string{"scan_disk_json_data"})
		if err != nil {
			return nil, err
		}
	}

	if config.ImportMode == datatype.ImportMode_cover {
		return filesPaths, nil
	}
	nonExistingSrcPaths, err := models.ResourcesDramaSeries{}.FilterNonExistingSrcPaths(db, filesBasesId, filesPaths)
	if err != nil {
		return nil, err
	}

	return nonExistingSrcPaths, nil
}

// ScanDiskImportData 扫描磁盘导入数据，创建资源记录
//
// 该函数会检查指定路径的视频文件是否存在，查找相似的图片作为海报，
// 如果未找到图片且配置允许自动创建，则从视频中提取关键帧作为海报，
// 最后创建资源记录。
//
// 参数:
//   - filesBasesId: 文件库ID，用于关联创建的资源
//   - filePath: 视频文件的完整路径
//   - config: 磁盘扫描配置，包含封面海报后缀名和自动创建海报等配置项
//
// 返回值:
//   - error: 执行过程中可能出现的错误
func (t ImportData) ScanDiskImportData(filesBasesId, filePath string, config datatype.Config_ScanDisk) error {
	exists := utils.FileExists(filePath)
	if !exists {
		return fmt.Errorf("文件不存在")
	}
	fileDir := utils.GetDirPathFromFilePath(filePath)
	fileName := utils.GetFileNameFromPath(filePath, false)

	// 如果启用了“文件夹转系列”，则先去查找是否已经有资源
	if config.FolderToSeries {
		resourcesDramaSeries, err := ResourcesDramaSeries{}.FindDramaSeriesSlcBySearchPath(filesBasesId, fileDir)
		if err == nil && len(*resourcesDramaSeries) > 0 {
			existsDramaSeriesFilePath := false
			//判断filepath是否已在改资源的剧集中，则不处理
			for _, resourcesDramaSeries := range *resourcesDramaSeries {
				if resourcesDramaSeries.Src == filePath {
					existsDramaSeriesFilePath = true
					break
				}
			}
			if !existsDramaSeriesFilePath {
				resourcesID := (*resourcesDramaSeries)[0].ResourcesID
				// 直接写入剧集信息
				err := ResourcesDramaSeries{}.Create(core.DBS(), resourcesID, filePath, len(*resourcesDramaSeries))
				if err != nil {
					return err
				}
				// 是否按名称重新排序剧集
				if config.FolderToSeriesSort {
					err := ResourcesDramaSeries{}.SortBySrc(resourcesID)
					if err != nil {
						return err
					}
				}
			}
			//如果是追加操作，这里已完成
			if config.ImportMode == datatype.ImportMode_append {
				return nil
			}
		}
	}

	coverPosterBase64, coverPosterWidth, coverPosterHeight, err := ImportData{}.GetCoverPosterBase64(filePath, config)

	// 资源标题
	resourceTitle := t.GetResourceTitle(filePath, config)

	resourceDataParam := datatype.ReqParam_Resource{
		Resource: datatype.ReqParam_ResourceBase{
			FilesBasesID:      filesBasesId,
			Title:             resourceTitle,
			Mode:              datatype.E_resourceMode_Movies,
			CoverPosterMode:   config.CoverPosterType,
			CoverPosterWidth:  coverPosterWidth,
			CoverPosterHeight: coverPosterHeight,
		},
		PhotoBase64: coverPosterBase64,
		DramaSeries: []datatype.ReqParam_resourceDramaSeries_Base{
			{Src: filePath},
		},
	}

	if config.AutoCreatePoster {
		resourceDataParam.Resource.Definition = t.VideoDefinition(filePath, config)
	}

	nfoPath := path.Join(fileDir, fileName+".nfo")
	// 如果nfo文件不存在，则从文件夹下所有nfo文件中查找
	if !utils.FileExists(nfoPath) && config.EnableNfoFuzzyMatch {
		//读取文件夹下所有nfo文件
		nfos, err := utils.GetFilesByExtensions([]string{fileDir}, []string{"nfo"}, false)
		if err != nil {
			return err
		}
		// 遍历所有nfo文件，判断文件名是否包含fileName
		for _, nfo := range nfos {
			tmpFileName := utils.GetFileNameFromPath(nfo, false)
			if strings.Contains(tmpFileName, fileName) {
				nfoPath = path.Join(fileDir, tmpFileName+".nfo")
				break
			}
		}
	}
	// 如果nfo文件不存在，则使用随机nfo文件
	if !utils.FileExists(nfoPath) && config.UseRandomNfoIfNoneMatch {
		nfos, err := utils.GetFilesByExtensions([]string{fileDir}, []string{"nfo"}, false)
		if err != nil {
			return err
		}
		if len(nfos) > 0 {
			nfoPath = nfos[0]
		}
	}

	t.Nfo(filesBasesId, nfoPath, config.Nfo, &resourceDataParam)

	//根据filePath地址，查询资源是否已经存在
	resourcesDramaSeriesScl, err := ResourcesDramaSeries{}.ListBySrc(filePath)
	if err != nil {
		return err
	}
	if len(*resourcesDramaSeriesScl) > 0 {
		//如果已经存在，则更新
		resourceID := (*resourcesDramaSeriesScl)[0].ResourcesID
		resourceDataParam.Resource.ID = resourceID
		_, err := Resources{}.UpdateResource(&resourceDataParam, false)
		return err
	} else {
		// 创建资源
		_, err = Resources{}.CreateResource(&resourceDataParam)
		return err
	}
}

// GetResourceTitle 根据配置的命名模式生成资源标题
//
// 该函数根据不同的资源命名模式配置，从文件路径中提取相应的部分作为资源标题。
// 支持目录名、文件名、目录+文件名以及完整路径四种命名方式。
//
// 参数:
//   - filePath: 视频文件的完整路径
//   - config: 磁盘扫描配置，包含资源命名模式等配置项
//
// 返回值:
//   - string: 根据配置生成的资源标题
func (ImportData) GetResourceTitle(filePath string, config datatype.Config_ScanDisk) string {
	// 资源标题
	dirName := utils.GetDirNameFromFilePath(filePath)
	fileName := utils.GetFileNameFromPath(filePath, false)
	resourceTitle := fileName
	switch config.ResourceNamingMode {
	case datatype.ResourceNamingModeDirName:
		resourceTitle = dirName
	case datatype.ResourceNamingModeDirFileName:
		resourceTitle = dirName + fileName
	case datatype.ResourceNamingModeFullPathName:
		resourceTitle = filePath
	}
	return resourceTitle
}

// VideoDefinition 获取视频文件的清晰度信息
//
// 该函数通过FFmpeg工具获取视频文件的基本信息，包括宽度和高度，
// 然后根据这些尺寸信息确定视频的清晰度等级并返回
//
// 参数:
//   - filePath: 视频文件的完整路径
//   - config: 磁盘扫描配置信息
//
// 返回值:
//   - string: 视频清晰度标识符（如"4K"、"1080P"等），如果获取失败则返回空字符串

func (ImportData) VideoDefinition(filePath string, config datatype.Config_ScanDisk) string {
	videoBasicInfo, err := processorsffmpeg.VideoInfo{}.GetVideoBasicInfo(filePath)
	if err == nil {
		videoDefinition := processorsffmpeg.VideoInfo{}.GetVideoDefinition(videoBasicInfo.Width, videoBasicInfo.Height)
		return string(videoDefinition)
	}
	return ""
}

// GetCoverPosterBase64 获取视频文件的封面海报Base64编码及相关尺寸信息
//
// 该函数首先在视频文件所在目录查找匹配的图片文件作为封面海报，
// 如果未找到且配置允许自动创建，则从视频中提取关键帧作为海报，
// 然后根据配置对海报进行尺寸调整和裁剪，最后转换为Base64编码返回
//
// 参数:
//   - filePath: 视频文件的完整路径
//   - config: 磁盘扫描配置，包含封面海报后缀名、匹配规则和处理选项等配置项
//
// 返回值:
//   - string: 封面海报的Base64编码字符串
//   - int: 封面海报的宽度
//   - int: 封面海报的高度
//   - error: 执行过程中可能出现的错误
func (t ImportData) GetCoverPosterBase64(filePath string, config datatype.Config_ScanDisk) (string, int, int, error) {
	fileDir := utils.GetDirPathFromFilePath(filePath)
	fileName := utils.GetFileNameFromPath(filePath, false)
	// 查找符合后缀名的图片文件
	imagePaths, err := utils.GetFilesByExtensions([]string{fileDir}, config.CoverPosterSuffixName, false)
	if err != nil {
		return "", 0, 0, err
	}
	// 匹配图片文件
	coverPosterPath := t.findCoverPoster(imagePaths, fileName, config.CoverPosterMatchName, config.CoverPosterFuzzyMatch, config.CoverPosterUseRandomImageIfNoMatch)
	var coverPosterBytes []byte
	var coverPosterBase64 string
	var coverPosterWidth int
	var coverPosterHeight int
	// 如果未找到相似图片，则尝试自动创建海报
	if coverPosterPath == "" && config.AutoCreatePoster {
		// 自动创建海报
		coverPosterBytes, err = processorsffmpeg.Thumbnail{}.ExtractThumbnailPoster(filePath)
		if err != nil {
			core.LogErr(err)
		}
	} else {
		coverPosterBytes, _ = utils.ImageToBytes(coverPosterPath)
	}
	// 获取图片尺寸
	coverPosterWidth, coverPosterHeight, err = utils.GetImageDimensionsFromBytes(coverPosterBytes)
	if err != nil {
		coverPosterWidth = 200
		coverPosterHeight = 200
	}
	// 如果是自动创建的海报，宽度大于400，则缩放海报
	if coverPosterPath == "" && config.AutoCreatePoster && coverPosterWidth > 400 {
		resizeImageBytes, resizeImageWidth, resizeImageHeight, err := utils.ResizeImageByMaxWidth(coverPosterBytes, 400)
		if err == nil {
			// 缩放后的海报数据
			coverPosterBytes = resizeImageBytes
			coverPosterWidth = resizeImageWidth
			coverPosterHeight = resizeImageHeight
		}
	}

	// 封面进行封面尺寸适应裁切
	if config.CoverPosterType >= 0 {
		cropImageBytes, err := utils.ResizeAndCropImage(coverPosterBytes, config.CoverPosterWidth, config.CoverPosterHeight)
		if err == nil {
			coverPosterBytes = cropImageBytes
			coverPosterWidth = config.CoverPosterWidth
			coverPosterHeight = config.CoverPosterHeight
		}
	}

	// 转换为Base64
	coverPosterBase64, _ = utils.ImageBytesToBase64(coverPosterBytes)
	return coverPosterBase64, coverPosterWidth, coverPosterHeight, err
}

// findCoverPoster 在给定的图片路径中查找匹配的封面海报
//
// 该函数根据提供的匹配规则在图片路径中查找与目标文件名匹配的封面海报。
// 可以使用预定义的匹配名称或直接使用目标文件名进行匹配，支持模糊匹配和严格匹配两种模式。
//
// 参数:
//   - imagePaths: 图片文件路径数组，用于在其中查找匹配的封面海报
//   - targetFileName: 目标文件名，当coverPosterMatchName为空时使用此名称进行匹配
//   - coverPosterMatchName: 预定义的封面海报匹配名称数组，用于指定匹配规则
//   - fuzzyMatch: 是否启用模糊匹配模式，true表示启用模糊匹配，false表示严格匹配
//   - coverPosterUseRandomImageIfNoMatch: 是否使用随机图片作为封面海报，true表示使用随机图片，false表示使用默认图片
//
// 返回值:
//   - string: 匹配到的图片文件路径，未找到匹配项时返回空字符串
func (ImportData) findCoverPoster(imagePaths []string, targetFileName string, coverPosterMatchName []string, fuzzyMatch bool, coverPosterUseRandomImageIfNoMatch bool) string {
	// 如果 coverPosterMatchName 为空，则使用 targetFileName 进行匹配
	if len(coverPosterMatchName) == 0 {
		// 查找与targetFileName相近的图片文件名
		for _, imagePath := range imagePaths {
			imageName := utils.GetFileNameFromPath(imagePath, false)
			if fuzzyMatch {
				// 模糊匹配：文件名完全匹配或者包含关系
				if imageName == targetFileName || strings.Contains(imageName, targetFileName) || strings.Contains(targetFileName, imageName) {
					return imagePath
				}
			} else {
				// 严格匹配：文件名完全匹配
				if imageName == targetFileName {
					return imagePath
				}
			}
		}
	} else {
		// 使用 coverPosterMatchName 的值做匹配
		for _, _matchName := range coverPosterMatchName {
			for _, imagePath := range imagePaths {
				imageName := utils.GetFileNameFromPath(imagePath, false)
				matchName := string(_matchName)

				//如果以regex:开头，代表使用这则表达式匹配，后面是正则表达式，其中@filename或者@fileName要被替换成文件名targetFileName
				if strings.HasPrefix(matchName, "regex:") {
					regexPattern := strings.TrimPrefix(matchName, "regex:")
					regex := strings.ReplaceAll(regexPattern, "@fileName", targetFileName)
					regex = strings.ReplaceAll(regex, "@filename", targetFileName)
					match, _ := regexp.MatchString(regex, imageName)
					if match {
						return imagePath
					}
				}
				// 如果是fileName，则将matchName替换成targetFileName，用以匹配文件名
				if _matchName == datatype.CoverPosterMatchName_fileName {
					matchName = targetFileName
				}
				if fuzzyMatch {
					// 模糊匹配：文件名完全匹配或者包含关系
					if imageName == matchName || strings.Contains(imageName, matchName) || strings.Contains(matchName, imageName) {
						return imagePath
					}
				} else {
					// 严格匹配：文件名完全匹配
					if imageName == matchName {
						return imagePath
					}
				}
			}
		}
	}
	if coverPosterUseRandomImageIfNoMatch && len(imagePaths) > 0 {
		// 随机取数组中的一个元素
		randIndex := utils.Rand_Intn(len(imagePaths))
		return imagePaths[randIndex]
	}
	return ""
}

// Nfo 从NFO文件中解析元数据并应用到资源数据
//
// 该函数读取指定路径的NFO文件(XML格式)，解析其中的元数据，
// 并根据配置将解析的数据应用到资源数据结构中。
//
// 参数:
//   - filesBasesId: 文件库ID
//   - nfoPath: NFO文件的完整路径
//   - nfoConfig: NFO配置信息，包含是否启用NFO功能及解析规则
//   - data: 资源数据指针，用于存储从NFO文件解析出的元数据
//
// 返回值:
//   - error: 执行过程中可能出现的错误
func (t ImportData) Nfo(filesBasesId, nfoPath string, nfoConfig datatype.Config_ScanDisk_Nfo, data *datatype.ReqParam_Resource) error {
	if !nfoConfig.NfoStatus {
		return nil
	}
	// 检查文件是否存在
	if !utils.FileExists(nfoPath) {
		return nil
	}
	// 读取XML文件
	xmlFile, err := os.Open(nfoPath)
	if err != nil {
		return nil
	}
	defer xmlFile.Close()

	// 解析XML
	byteValue, _ := io.ReadAll(xmlFile)

	// 创建XML解码器
	decoder := xml.NewDecoder(strings.NewReader(string(byteValue)))

	// 解析为通用结构
	rootElement, err := utils.XML_parseXMLToMap(decoder)
	if err != nil {
		return err
	}

	return t.NfoExecData(filesBasesId, nfoPath, rootElement, nfoConfig, data)
}

// NfoExecData 从NFO配置中提取XML数据并填充到资源数据结构中
//
// 该函数根据NFO配置中指定的XPath路径从XML数据中提取各种媒体信息，
// 包括标题、番号、发行日期、简介、标签、演员等，并将这些信息填充到资源数据结构中。
//
// 参数:
//   - filesBasesId: 文件库ID，用于关联标签和演员信息
//   - nfoPath: NFO文件的完整路径，用于定位相关资源文件
//   - rootElement: 已解析的XML根元素数据映射
//   - nfoConfig: NFO配置信息，包含各类字段的XPath路径配置
//   - data: 资源数据指针，用于存储从XML中提取的信息
//
// 返回值:
//   - error: 执行过程中可能出现的错误
func (ImportData) NfoExecData(filesBasesId, nfoPath string, rootElement map[string]interface{}, nfoConfig datatype.Config_ScanDisk_Nfo, data *datatype.ReqParam_Resource) error {
	var xmlData map[string]interface{}

	// 根据Roots配置获取根节点
	if len(nfoConfig.Roots) > 0 {
		// 有指定根节点
		for _, rootPath := range nfoConfig.Roots {
			if data := utils.XML_getXMLValueByPath(rootElement, rootPath); data != nil {
				if mappedData, ok := data.(map[string]interface{}); ok {
					xmlData = mappedData
					break
				}
			}
		}
	} else {
		// 没有指定根节点，使用整个文档
		xmlData = rootElement
	}

	// 如果没有找到有效的数据节点，直接返回
	if xmlData == nil {
		return nil
	}

	// 提取标题
	if len(nfoConfig.Titles) > 0 {
		for _, titlePath := range nfoConfig.Titles {
			if value := utils.XML_getXMLValueByPath(xmlData, titlePath); value != "" {
				data.Resource.Title = value.(string)
				break
			}
		}
	}

	// 提取番号
	if len(nfoConfig.IssueNumbers) > 0 {
		for _, issueNumberPath := range nfoConfig.IssueNumbers {
			if value := utils.XML_getXMLValueByPath(xmlData, issueNumberPath); value != "" {
				data.Resource.IssueNumber = value.(string)
				break
			}
		}
	}

	// 提取发行日期
	if len(nfoConfig.IssuingDates) > 0 {
		for _, issuingDatePath := range nfoConfig.IssuingDates {
			if value := utils.XML_getXMLValueByPath(xmlData, issuingDatePath); value != "" {
				data.Resource.IssuingDate = value.(string)
				break
			}
		}
	}

	// 提取评分
	if len(nfoConfig.Score) > 0 {
		for _, scorePath := range nfoConfig.Score {
			if value := utils.XML_getXMLValueByPath(xmlData, scorePath); value != "" {
				// 将字符串转换为浮点数
				scoreStr := value.(string)
				score, err := strconv.ParseFloat(scoreStr, 64)
				if err != nil {
					// 转换失败，跳过该评分
					continue
				}

				// 规范化评分到0-10区间
				if score >= 0 && score <= 10 {
					// 0-10区间直接使用，保留一位小数
					data.Resource.Score = math.Round(score*10) / 10
				} else if score > 10 && score <= 100 {
					// 10-100区间除以10，保留一位小数
					normalizedScore := score / 10.0
					data.Resource.Score = math.Round(normalizedScore*10) / 10
				} else {
					// 不在有效区间内，设为默认值0
					data.Resource.Score = 0
				}
				break
			}
		}
	}

	// 提取简介
	if len(nfoConfig.Abstracts) > 0 {
		for _, abstractPath := range nfoConfig.Abstracts {
			if value := utils.XML_getXMLValueByPath(xmlData, abstractPath); value != "" {
				data.Resource.Abstract = value.(string)
				break
			}
		}
	}
	// 提取标签
	if len(nfoConfig.Tags) > 0 {
		for _, tagPath := range nfoConfig.Tags {
			if values := utils.XML_getXMLValuesByPath(xmlData, tagPath); len(values) > 0 {
				for _, value := range values {
					tag, err := Tag{}.TagInfoByNameNotFoundCreate(filesBasesId, value)
					if err != nil {
						continue
					}
					data.Tags = append(data.Tags, tag.ID)
				}
				break
			}
		}
	}

	performerPhotoBase64 := ""
	// 提取演员照片
	if len(nfoConfig.PerformerThumbs) > 0 && nfoPath != "" {
		for _, performerThumbPath := range nfoConfig.PerformerThumbs {
			if values := utils.XML_getXMLValuesByPath(xmlData, performerThumbPath); len(values) > 0 {
				for _, value := range values {
					// 获取nfo文件所在目录
					nfoDir := filepath.Dir(nfoPath)
					// 调用我们新创建的函数处理演员图片
					base64Image, err := getPerformerImage(value, nfoDir)
					if err != nil {
						core.LogErr(err)
						// 处理错误，可以选择记录日志或者忽略
						continue
					}
					if base64Image != "" {
						// 将获取到的base64图片数据赋值给performerPhotoBase64
						performerPhotoBase64 = base64Image
						// 找到第一个有效的图片就跳出循环
						break
					}
				}
				// 如果已经获取到图片，就不再处理其他路径
				if performerPhotoBase64 != "" {
					break
				}
			}
		}
	}

	// 提取演员
	if len(nfoConfig.PerformerNames) > 0 {
		for _, performerPath := range nfoConfig.PerformerNames {
			if values := utils.XML_getXMLValuesByPath(xmlData, performerPath); len(values) > 0 {
				for _, value := range values {
					performer, err := Performer{}.PerformerInfoByNameNotFoundCreate(filesBasesId, value, performerPhotoBase64)
					if err != nil {
						continue
					}
					data.Performers = append(data.Performers, performer.ID)
				}
				break
			}
		}
	}

	return nil
}

// CRHandlePerformerImage 处理演员图片，根据value判断是链接还是本地文件路径
// 如果是链接，则调用爬虫获取图片数据；如果是本地路径，则直接读取文件
// 最后将图片数据转换为base64格式
// 参数:
//   - value: 图片的URL或者相对路径
//   - nfoDir: nfo文件所在的目录路径
//
// 返回值:
//   - string: base64编码的图片数据
//   - error: 错误信息
func getPerformerImage(value string, nfoDir string) (string, error) {
	var imageData []byte
	var err error

	// 不是链接，认为是相对路径，拼接完整路径
	fullPath := filepath.Join(nfoDir, value)
	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", fmt.Errorf("本地图片文件不存在: %s", fullPath)
	}

	// 读取本地图片文件
	imageData, err = os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("读取本地图片文件失败: %v", err)
	}

	// 将图片数据转换为base64格式
	base64Data, err := utils.ImageBytesToBase64(imageData)
	if err != nil {
		return "", fmt.Errorf("转换图片为base64失败: %v", err)
	}

	return base64Data, nil
}
