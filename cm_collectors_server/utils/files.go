package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
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

var FileImageExtensions = []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}

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

// TrimBasePath 从文件路径中移除基础路径部分
// 参数:
// fullPath: 完整的文件路径
// basePath: 要移除的基础路径
// 返回值:
// string: 移除基础路径后的相对路径部分
func TrimBasePath(fullPath, basePath string) string {
	// 标准化路径分隔符
	fullPath = filepath.ToSlash(fullPath)
	basePath = filepath.ToSlash(basePath)

	// 确保basePath以斜杠结尾
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}

	// 移除基础路径部分
	return strings.TrimPrefix(fullPath, basePath)
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

// MoveFile 移动文件从源路径到目标路径
//
// 参数:
//   - srcPath: 源文件路径
//   - dstPath: 目标文件路径
//
// 返回值:
//   - error: 如果移动过程中出现错误则返回错误信息，否则返回 nil
func MoveFile(srcPath, dstPath string) error {
	// 确保目标目录存在
	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}
	// 执行文件移动操作
	return os.Rename(srcPath, dstPath)
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

// FileMD5 计算文件的MD5哈希值
// 参数:
// filePath: 文件的完整路径
// 返回值:
// string: 文件的MD5哈希值
// error: 错误信息，如果计算成功则为nil
func FileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	hashInBytes := hash.Sum(nil)[:16]
	return fmt.Sprintf("%x", hashInBytes), nil
}

// FileSHA256 计算文件的SHA256哈希值
// 参数:
// filePath: 文件的完整路径
// 返回值:
// string: 文件的SHA256哈希值
// error: 错误信息，如果计算成功则为nil
func FileSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	hashInBytes := hash.Sum(nil)
	return fmt.Sprintf("%x", hashInBytes), nil
}

// FileExt 获取文件路径的扩展名
// 参数:
// filePath: 文件的完整路径
// 返回值:
// string: 文件的扩展名，包括前导点号（如".txt"），如果文件没有扩展名则返回空字符串
func FileExt(filePath string) string {
	return strings.ToLower(filepath.Ext(filePath))
}

// WriteStringToFile 将字符串写入指定文件
// 参数:
// filePath: 文件的完整路径
// content: 要写入的字符串内容
// 返回值:
// error: 错误信息，如果写入成功则为nil
func WriteStringToFile(filePath string, content string) error {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 创建或截断文件并写入内容
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

// SanitizePath 对路径进行安全处理，防止路径遍历攻击
// 参数:
// path: 需要处理的路径
// 返回值:
// string: 安全处理后的路径
func SanitizePath(path string) string {
	// 使用filepath.Clean清理路径，移除冗余的元素如..和.
	cleanPath := filepath.Clean(path)

	// 移除路径开头的分隔符，防止绝对路径
	cleanPath = strings.TrimPrefix(cleanPath, string(filepath.Separator))

	// 移除路径开头的../等相对路径元素
	for strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) ||
		strings.HasPrefix(cleanPath, "..") {
		if strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) {
			cleanPath = strings.TrimPrefix(cleanPath, ".."+string(filepath.Separator))
		} else if strings.HasPrefix(cleanPath, "..") {
			cleanPath = strings.TrimPrefix(cleanPath, "..")
		}
	}

	return cleanPath
}

// ReadFile 读取指定路径的文件内容
// 参数:
// fullPath: 文件路径
// 返回值:
// []byte: 文件内容
// error: 错误信息
func ReadFile(fullPath string) ([]byte, error) {
	// 检查文件是否存在
	if !FileExists(fullPath) {
		return nil, fmt.Errorf("file does not exist")
	}
	// 读取文件内容
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file")
	}

	return content, nil
}
