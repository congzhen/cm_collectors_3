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
	//查找在文件夹中，是否有相似文件名的图片
	imagePaths, err := utils.GetFilesByExtensions([]string{fileDir}, config.CoverPosterSuffixName, false)
	if err != nil {
		return err
	}
	coverPosterPath := t.findSimilarImage(imagePaths, fileName)
	var coverPosterBytes []byte
	var coverPosterBase64 string
	var coverPosterWidth int
	var coverPosterHeight int

	// 如果未找到相似图片，则尝试自动创建海报
	if coverPosterPath == "" && config.AutoCreatePoster {
		// 自动创建海报
		coverPosterBytes, err = processorsffmpeg.KeyFrame{}.ExtractKeyframePoster(filePath, 10)
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

	resourceDataParam := datatype.ReqParam_Resource{
		Resource: datatype.ReqParam_ResourceBase{
			FilesBasesID:      filesBasesId,
			Title:             fileName,
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

// findSimilarImage 在图片路径列表中查找与目标文件名相似的图片
//
// 该函数通过比较文件名来查找相似的图片文件，支持完全匹配和包含关系匹配
//
// 参数:
//   - imagePaths: 图片文件路径列表
//   - targetFileName: 目标文件名（不包含扩展名）
//
// 返回值:
//   - string: 找到的相似图片路径，如果未找到则返回空字符串
func (ImportData) findSimilarImage(imagePaths []string, targetFileName string) string {
	// 查找与targetFileName相近的图片文件名
	for _, imagePath := range imagePaths {
		imageName := utils.GetFileNameFromPath(imagePath, false)
		// 如果文件名完全匹配 或包含关系
		if imageName == targetFileName || strings.Contains(imageName, targetFileName) {
			return imagePath
		}
	}
	return ""
}
