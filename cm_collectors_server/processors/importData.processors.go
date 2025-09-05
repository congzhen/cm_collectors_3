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
	"os"
	"path"
	"strings"
)

type ImportData struct {
}

// ScanDiskImportPaths 扫描磁盘路径中的视频文件，并更新扫描配置
//
// 该函数会扫描指定路径下的所有视频文件，将扫描配置保存到数据库，
// 并过滤出尚未在数据库中记录的新文件路径。
//
// 参数:
//   - filesBasesId: 文件库ID，用于关联扫描配置和过滤已存在的文件
//   - config: 磁盘扫描配置，包含扫描路径和视频文件后缀等信息
//
// 返回值:
//   - []string: 不存在于数据库中的新文件路径列表
//   - error: 执行过程中可能出现的错误
func (ImportData) ScanDiskImportPaths(filesBasesId string, config datatype.Config_ScanDisk) ([]string, error) {
	filesPaths, err := utils.GetFilesByExtensions(config.ScanDiskPaths, config.VideoSuffixName, true)
	if err != nil {
		return nil, err
	}
	filesPaths = utils.SortFilesByOrder(filesPaths, utils.FileTimeAsc)
	db := core.DBS()
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
	if !config.CheckPath {
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
	// 查找符合后缀名的图片文件
	imagePaths, err := utils.GetFilesByExtensions([]string{fileDir}, config.CoverPosterSuffixName, false)
	if err != nil {
		return err
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

	// 资源标题
	resourceTitle := fileName
	dirName := utils.GetDirNameFromFilePath(filePath)
	switch config.ResourceNamingMode {
	case datatype.ResourceNamingModeDirName:
		resourceTitle = dirName
	case datatype.ResourceNamingModeDirFileName:
		resourceTitle = dirName + fileName
	case datatype.ResourceNamingModeFullPathName:
		resourceTitle = filePath
	}

	resourceDataParam := datatype.ReqParam_Resource{
		Resource: datatype.ReqParam_ResourceBase{
			FilesBasesID:      filesBasesId,
			Title:             resourceTitle,
			Mode:              datatype.E_resourceMode_Movies,
			CoverPosterMode:   -1,
			CoverPosterWidth:  coverPosterWidth,
			CoverPosterHeight: coverPosterHeight,
		},
		PhotoBase64: coverPosterBase64,
		DramaSeries: []datatype.ReqParam_resourceDramaSeries_Base{
			{Src: filePath},
		},
	}

	if config.AutoCreatePoster {
		videoBasicInfo, err := processorsffmpeg.VideoInfo{}.GetVideoBasicInfo(filePath)
		if err == nil {
			videoDefinition := processorsffmpeg.VideoInfo{}.GetVideoDefinition(videoBasicInfo.Width, videoBasicInfo.Height)
			resourceDataParam.Resource.Definition = string(videoDefinition)
		}
	}

	t.nfo(filesBasesId, path.Join(fileDir, fileName+".nfo"), config.Nfo, &resourceDataParam)

	// 创建资源
	_, err = Resources{}.CreateResource(&resourceDataParam)
	return err
}

func (ImportData) nfo(filesBasesId, nfoPath string, nfoConfig datatype.Config_ScanDisk_Nfo, data *datatype.ReqParam_Resource) error {
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
	rootElement, err := parseXMLToMap(decoder)
	if err != nil {
		return err
	}

	// 根据Roots配置获取根节点
	var xmlData map[string]interface{}
	if len(nfoConfig.Roots) > 0 {
		// 有指定根节点
		for _, rootPath := range nfoConfig.Roots {
			if data := getXMLValueByPath(rootElement, rootPath); data != nil {
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
			if value := getXMLValueByPath(xmlData, titlePath); value != "" {
				data.Resource.Title = value.(string)
				break
			}
		}
	}

	// 提取番号
	if len(nfoConfig.IssueNumbers) > 0 {
		for _, issueNumberPath := range nfoConfig.IssueNumbers {
			if value := getXMLValueByPath(xmlData, issueNumberPath); value != "" {
				data.Resource.IssueNumber = value.(string)
				break
			}
		}
	}

	// 提取发行日期
	if len(nfoConfig.IssuingDates) > 0 {
		for _, issuingDatePath := range nfoConfig.IssuingDates {
			if value := getXMLValueByPath(xmlData, issuingDatePath); value != "" {
				data.Resource.IssuingDate = value.(string)
				break
			}
		}
	}

	// 提取简介
	if len(nfoConfig.Abstracts) > 0 {
		for _, abstractPath := range nfoConfig.Abstracts {
			if value := getXMLValueByPath(xmlData, abstractPath); value != "" {
				data.Resource.Abstract = value.(string)
				break
			}
		}
	}
	// 提取标签
	if len(nfoConfig.Tags) > 0 {
		for _, tagPath := range nfoConfig.Tags {
			if values := getXMLValuesByPath(xmlData, tagPath); len(values) > 0 {
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

	/*
		// 提取演员
		if len(nfoConfig.PerformerNames) > 0 {
			for _, performerPath := range nfoConfig.PerformerNames {
				if values := getXMLValuesByPath(xmlData, performerPath); len(values) > 0 {
					data.Performers = values
					break
				}
			}
		}


	*/
	return nil
}

// 将XML解析为map[string]interface{}结构
func parseXMLToMap(decoder *xml.Decoder) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		switch se := token.(type) {
		case xml.StartElement:
			name := se.Name.Local
			value, err := parseXMLElement(decoder, se)
			if err != nil {
				return nil, err
			}

			// 处理重复元素名，存储为数组
			if existing, exists := result[name]; exists {
				switch existing := existing.(type) {
				case []interface{}:
					result[name] = append(existing, value)
				default:
					result[name] = []interface{}{existing, value}
				}
			} else {
				result[name] = value
			}
		}
	}

	return result, nil
}

// 解析单个XML元素
func parseXMLElement(decoder *xml.Decoder, start xml.StartElement) (interface{}, error) {
	elementData := make(map[string]interface{})

	// 解析元素的属性
	for _, attr := range start.Attr {
		elementData["@"+attr.Name.Local] = attr.Value
	}

	// 处理元素内容和子元素
	content := ""
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			// 递归解析子元素
			childValue, err := parseXMLElement(decoder, t)
			if err != nil {
				return nil, err
			}

			name := t.Name.Local
			if existing, exists := elementData[name]; exists {
				switch existing := existing.(type) {
				case []interface{}:
					elementData[name] = append(existing, childValue)
				default:
					elementData[name] = []interface{}{existing, childValue}
				}
			} else {
				elementData[name] = childValue
			}

		case xml.CharData:
			// 处理文本内容
			content += string(t)

		case xml.EndElement:
			// 元素结束
			if t.Name.Local == start.Name.Local {
				// 如果只有文本内容且没有子元素，则返回文本内容
				if len(elementData) == len(start.Attr) { // 只有属性
					trimmedContent := strings.TrimSpace(content)
					if trimmedContent != "" {
						if len(start.Attr) == 0 { // 没有属性
							return trimmedContent, nil
						}
						// 有属性的情况
						elementData["#text"] = trimmedContent
					} else if len(start.Attr) == 0 {
						// 没有内容也没有属性
						return "", nil
					}
				}
				return elementData, nil
			}
		}
	}
}

// 根据路径获取XML中的值，支持嵌套节点（如：movie.title）
func getXMLValueByPath(data map[string]interface{}, path string) interface{} {
	// 分割路径
	paths := strings.Split(path, ".")
	current := data

	// 遍历路径
	for i, p := range paths {
		// 如果是最后一个元素，直接获取值
		if i == len(paths)-1 {
			if val, ok := current[p]; ok {
				return val
			}
			return ""
		}

		// 否则继续深入
		if val, ok := current[p]; ok {
			switch v := val.(type) {
			case map[string]interface{}:
				current = v
			case []interface{}:
				// 如果是数组，取第一个元素
				if len(v) > 0 {
					if next, ok := v[0].(map[string]interface{}); ok {
						current = next
					} else {
						return ""
					}
				} else {
					return ""
				}
			default:
				return ""
			}
		} else {
			return ""
		}
	}

	return ""
}

// 根据路径获取XML中的值数组，支持嵌套节点（如：movie.actors.actor）
func getXMLValuesByPath(data map[string]interface{}, path string) []string {
	paths := strings.Split(path, ".")
	current := data

	// 遍历到倒数第二个路径
	for i := 0; i < len(paths)-1; i++ {
		p := paths[i]
		if val, ok := current[p]; ok {
			switch v := val.(type) {
			case map[string]interface{}:
				current = v
			case []interface{}:
				// 如果是数组，取第一个元素
				if len(v) > 0 {
					if next, ok := v[0].(map[string]interface{}); ok {
						current = next
					} else {
						return []string{}
					}
				} else {
					return []string{}
				}
			default:
				return []string{}
			}
		} else {
			return []string{}
		}
	}

	// 获取最后一个路径的值
	lastPath := paths[len(paths)-1]
	if val, ok := current[lastPath]; ok {
		// 如果是数组
		if arr, ok := val.([]interface{}); ok {
			var result []string
			for _, item := range arr {
				if str, ok := item.(string); ok {
					result = append(result, str)
				} else if mapped, ok := item.(map[string]interface{}); ok {
					// 如果是map，尝试获取#text字段
					if text, exists := mapped["#text"]; exists {
						if str, ok := text.(string); ok {
							result = append(result, str)
						}
					}
				}
			}
			return result
		}
		// 如果是单个值
		if str, ok := val.(string); ok {
			return []string{str}
		} else if mapped, ok := val.(map[string]interface{}); ok {
			// 如果是map，尝试获取#text字段
			if text, exists := mapped["#text"]; exists {
				if str, ok := text.(string); ok {
					return []string{str}
				}
			}
		}
	}

	return []string{}
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
