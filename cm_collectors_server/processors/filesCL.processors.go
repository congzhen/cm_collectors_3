package processors

import (
	"cm_collectors_server/errorMessage"
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type FilesCL struct{}

// FileType 定义文件类型枚举
type FileType int

const (
	FilesCL_All      FileType = iota // 所有文件
	FilesCL_Image                    // 图片文件
	FilesCL_Video                    // 视频文件
	FilesCL_Audio                    // 音频文件
	FilesCL_Document                 // 文档文件
)

// fileExtensions 定义各种文件类型的扩展名
var fileExtensions = map[FileType][]string{
	FilesCL_Image:    {".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp", ".svg"},
	FilesCL_Video:    {".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm", ".m4v"},
	FilesCL_Audio:    {".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma", ".m4a"},
	FilesCL_Document: {".pdf", ".doc", ".docx", ".txt", ".xls", ".xlsx", ".ppt", ".pptx"},
}

func (t FilesCL) FilesListByDramaSeriesId(dramaSeriesId string, fileType FileType) ([]string, error) {
	src, err := ResourcesDramaSeries{}.GetSrc(dramaSeriesId)
	if err != nil {
		return nil, err
	}
	return t.getAllFilesInDirectory(src, fileType)
}

func (t FilesCL) FilesImage(dramaSeriesId, fileName string) ([]byte, error) {
	src, err := ResourcesDramaSeries{}.GetSrc(dramaSeriesId)
	if err != nil {
		return nil, err
	}
	dirPath, err := t.checkSrcFolder(src)
	if err != nil {
		return nil, err
	}
	filePath := path.Join(dirPath, fileName)
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, errorMessage.WrapError(errorMessage.Err_Resources_Play_Src_Error, errors.New(fileName))
	}
	// 读取文件内容
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return fileData, nil
}

// 检测源路径最终的文件夹是否存在，并返回文件夹地址
func (FilesCL) checkSrcFolder(src string) (string, error) {
	// 获取文件信息
	fileInfo, err := os.Stat(src)
	if err != nil {
		// 如果路径不存在，返回错误
		if os.IsNotExist(err) {
			return "", errorMessage.Err_Resources_Play_Src_Error
		}
		// 其他错误
		return "", err
	}

	// 检查是否为目录
	if !fileInfo.IsDir() {
		// 如果不是目录，获取其所在目录
		dir := filepath.Dir(src)
		// 检查目录是否存在
		if _, err := os.Stat(dir); err != nil {
			if os.IsNotExist(err) {
				return "", errorMessage.Err_Resources_Play_Src_Error
			}
			return "", err
		}
		return dir, nil
	}

	// 如果是目录，直接返回
	return src, nil
}

// 获取文件夹下的所有文件，支持按文件类型筛选，只返回文件名
func (t FilesCL) getAllFilesInDirectory(directoryPath string, fileType FileType) ([]string, error) {
	// 检查目录是否存在并获取正确的目录路径
	dirPath, err := t.checkSrcFolder(directoryPath)
	if err != nil {
		return nil, err
	}

	var files []string

	// 使用filepath.Walk遍历目录
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		// 如果遍历过程中出现错误，返回错误
		if err != nil {
			return err
		}

		// 跳过目录，只处理文件
		if info.IsDir() {
			return nil
		}

		// 如果选择所有文件类型，则包含所有文件
		if fileType == FilesCL_All {
			// 只返回文件名，不返回完整路径
			files = append(files, info.Name())
			return nil
		}

		// 检查文件扩展名是否匹配指定的类型
		ext := strings.ToLower(filepath.Ext(path))
		if extensions, exists := fileExtensions[fileType]; exists {
			for _, validExt := range extensions {
				if ext == strings.ToLower(validExt) {
					// 只返回文件名，不返回完整路径
					files = append(files, info.Name())
					return nil
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
