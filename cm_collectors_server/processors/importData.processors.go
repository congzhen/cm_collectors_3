package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	processorsffmpeg "cm_collectors_server/processorsFFmpeg"
	"cm_collectors_server/utils"
	"encoding/json"
	"fmt"
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
	// 创建资源
	_, err = Resources{}.CreateResource(&resourceDataParam)
	return err
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
func (ImportData) findCoverPoster(imagePaths []string, targetFileName string, coverPosterMatchName []datatype.CoverPosterMatchName, fuzzyMatch bool, coverPosterUseRandomImageIfNoMatch bool) string {
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
