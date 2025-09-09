package utils

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SortOrder 定义文件排序方式的枚举类型
type FilesSortOrder int

const (
	// FileNameAsc 按文件名正序排序
	FileNameAsc FilesSortOrder = iota
	// FileNameDesc 按文件名倒序排序
	FileNameDesc
	// FileTimeAsc 按文件创建时间正序排序
	FileTimeAsc
	// FileTimeDesc 按文件创建时间倒序排序
	FileTimeDesc
)

// GetFilesByExtensions 根据指定的文件扩展名筛选目录中的文件
// 参数:
// dirPaths: 要搜索的目录路径数组，每个路径都是完整的绝对路径
// extensions: 文件扩展名数组，例如 [".mp4", ".avi"] 或 ["mp4", "avi"]
// recursive: 是否递归遍历子目录，true表示递归遍历，false表示只遍历当前目录
// 返回值:
// []string: 匹配的文件绝对路径数组
// error: 错误信息，如果提取成功则为nil
func GetFilesByExtensions(dirPaths []string, extensions []string, recursive bool) ([]string, error) {
	var result []string

	if len(dirPaths) == 0 || len(extensions) == 0 {
		return result, nil
	}

	// 标准化扩展名，确保它们以点号开头
	normalizedExtensions := make([]string, len(extensions))
	for i, ext := range extensions {
		if !strings.HasPrefix(ext, ".") {
			normalizedExtensions[i] = "." + ext
		} else {
			normalizedExtensions[i] = ext
		}
	}

	// 处理每个目录
	for _, dirPath := range dirPaths {
		// 根据是否递归选择不同的遍历方式
		var err error
		if recursive {
			// 递归遍历目录及其子目录
			err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				// 跳过目录
				if info.IsDir() {
					return nil
				}

				// 检查文件扩展名是否匹配指定的任何扩展名
				fileExt := strings.ToLower(filepath.Ext(path))
				for _, ext := range normalizedExtensions {
					if fileExt == strings.ToLower(ext) {
						// 将路径转换为标准路径
						path = filepath.ToSlash(path)
						result = append(result, path)
						break
					}
				}

				return nil
			})
		} else {
			// 只遍历当前目录，不递归子目录
			entries, readErr := os.ReadDir(dirPath)
			if readErr != nil {
				err = readErr
			} else {
				for _, entry := range entries {
					// 跳过目录
					if entry.IsDir() {
						continue
					}

					// 检查文件扩展名是否匹配指定的任何扩展名
					fileExt := strings.ToLower(filepath.Ext(entry.Name()))
					for _, ext := range normalizedExtensions {
						if fileExt == strings.ToLower(ext) {
							// 将路径转换为标准路径
							path := filepath.ToSlash(filepath.Join(dirPath, entry.Name()))
							result = append(result, path)
							break
						}
					}
				}
			}
		}

		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// SortFilesByOrder 根据指定的排序方式对文件路径列表进行排序
// 参数:
// filePaths: 要排序的文件路径数组
// sortOrder: 排序方式（FileNameAsc、FileNameDesc、FileTimeAsc、FileTimeDesc）
// 返回值:
// []string: 排序后的文件路径数组
func SortFilesByOrder(filePaths []string, sortOrder FilesSortOrder) []string {
	switch sortOrder {
	case FileNameAsc:
		sort.Strings(filePaths)
	case FileNameDesc:
		sort.Sort(sort.Reverse(sort.StringSlice(filePaths)))
	case FileTimeAsc, FileTimeDesc:
		// 创建包含文件路径和信息的结构体切片
		filesWithInfo := make([]struct {
			path string
			info os.FileInfo
		}, 0, len(filePaths))

		// 获取所有文件的信息
		for _, path := range filePaths {
			if info, err := os.Stat(path); err == nil {
				filesWithInfo = append(filesWithInfo, struct {
					path string
					info os.FileInfo
				}{path: path, info: info})
			}
		}

		// 按时间排序
		sort.Slice(filesWithInfo, func(i, j int) bool {
			if sortOrder == FileTimeAsc {
				return filesWithInfo[i].info.ModTime().Before(filesWithInfo[j].info.ModTime())
			}
			return filesWithInfo[i].info.ModTime().After(filesWithInfo[j].info.ModTime())
		})

		// 提取排序后的路径
		for i := range filesWithInfo {
			filePaths[i] = filesWithInfo[i].path
		}
	}

	return filePaths
}

// GetFileNameFromPath 从给定的文件路径中提取文件名
// 参数:
// path: 文件的完整路径
// withExtension: 是否包含扩展名，true表示包含，false表示不包含
// 返回值:
// string: 文件名
func GetFileNameFromPath(path string, withExtension bool) string {
	filename := filepath.Base(path)
	if withExtension {
		return filename
	}

	// 移除扩展名
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}

// GetDirPathFromFilePath 根据文件路径获取文件所在的文件夹路径
// 参数:
// filePath: 文件的完整路径
// 返回值:
// string: 文件所在的文件夹路径
func GetDirPathFromFilePath(filePath string) string {
	return filepath.ToSlash(filepath.Dir(filePath))
}

// FileExists 判断文件或目录是否存在
// 参数:
// path: 要检查的文件或目录路径
// 返回值:
// bool: 文件或目录存在返回true，否则返回false
func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// GetDirNameFromFilePath 根据文件路径获取文件所在的文件夹名称
// 参数:
// filePath: 文件的完整路径
// 返回值:
// string: 文件所在的文件夹名称
func GetDirNameFromFilePath(filePath string) string {
	dirPath := filepath.Dir(filePath)
	return filepath.Base(dirPath)
}

// IsSameDirectory 判断两个路径是否在同一个目录下
// 参数:
// path1: 第一个路径（文件或文件夹）
// path2: 第二个路径（文件或文件夹）
// 返回值:
// bool: 在同一目录下返回true，否则返回false
func IsSameDirectory(path1, path2 string) bool {
	// 获取path1的目录路径
	var dir1 string
	if info, err := os.Stat(path1); err == nil && info.IsDir() {
		// path1是目录
		dir1 = filepath.Clean(path1)
	} else {
		// path1是文件
		dir1 = filepath.Dir(path1)
	}

	// 获取path2的目录路径
	var dir2 string
	if info, err := os.Stat(path2); err == nil && info.IsDir() {
		// path2是目录
		dir2 = filepath.Clean(path2)
	} else {
		// path2是文件
		dir2 = filepath.Dir(path2)
	}

	return filepath.Clean(dir1) == filepath.Clean(dir2)
}
