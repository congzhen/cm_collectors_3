package serverfilemanagement

import (
	"archive/zip"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// 文件条目结构体
type ServerFileManagement_FileEntry struct {
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	IsDir       bool       `json:"is_dir"`
	Permissions string     `json:"permissions"`
	Size        int64      `json:"size"`
	ModifiedAt  *time.Time `json:"modified_at"`
	Path        string     `json:"path"`
}

type ServerFileManagement struct {
	RootPath []ServerFileManagement_PathEntry
}

type ServerFileManagement_PathEntry struct {
	RealPath    string // 真实路径
	VirtualPath string // 虚拟路径
	cleanRoot   string // 预处理后的标准化根路径
}

// 预处理根路径（消除多余斜杠并标准化）
func (s *ServerFileManagement) preprocessRoots() {
	for i := range s.RootPath {
		root := &s.RootPath[i] // 获取指向原始切片元素的指针

		// 获取绝对路径
		absRealPath, err := filepath.Abs(root.RealPath)
		if err != nil {
			continue // 或记录错误
		}

		_cleanRoot := filepath.Clean(absRealPath)
		_cleanRoot = filepath.ToSlash(_cleanRoot)

		// 确保以分隔符结尾
		if !strings.HasSuffix(_cleanRoot, "/") {
			_cleanRoot += "/"
		}
		root.cleanRoot = _cleanRoot // 直接修改原始切片中的值
	}
}

// ConvertCleanRootToVirtualPath 将 cleanRoot 转换为 VirtualPath
func (s *ServerFileManagement) convertCleanRootToVirtualPath(cleanRoot string) string {
	// 标准化 cleanRoot 路径
	cleanRoot = filepath.ToSlash(filepath.Clean(cleanRoot))
	if !strings.HasSuffix(cleanRoot, "/") {
		cleanRoot += "/"
	}

	// 遍历所有预处理后的根路径，找到匹配的 cleanRoot
	for _, root := range s.RootPath {
		// 确保 cleanRoot 以 root.cleanRoot 开头
		compareRoot := filepath.ToSlash(root.cleanRoot)
		if strings.HasPrefix(cleanRoot, compareRoot) {
			// 移除 cleanRoot 的公共前缀，得到 VirtualPath
			virtualPath := strings.TrimPrefix(cleanRoot, compareRoot)
			return filepath.ToSlash(path.Join(root.VirtualPath, virtualPath))
		}
	}
	return cleanRoot
}

// ConvertVirtualPathToCleanRoot 将 VirtualPath 转换为 cleanRoot
func (s *ServerFileManagement) convertVirtualPathToCleanRoot(virtualPath string) string {
	// 遍历所有预处理后的根路径，找到匹配的 VirtualPath 前缀
	for _, root := range s.RootPath {
		compareVirtualPath := filepath.ToSlash(root.VirtualPath)
		if strings.HasPrefix(virtualPath, compareVirtualPath) {
			// 移除 VirtualPath 的公共前缀，得到相对路径
			relativePath := strings.TrimPrefix(virtualPath, compareVirtualPath)
			// 手动拼接 cleanRoot 和相对路径，确保以 / 结尾
			cleanRoot := root.cleanRoot
			if !strings.HasSuffix(cleanRoot, "/") {
				cleanRoot += "/"
			}
			return cleanRoot + strings.TrimPrefix(relativePath, "/")
		}
	}
	return virtualPath
}

func (s ServerFileManagement) IsAbs(path string) error {
	if s.isWindowsDriveRoot(path) {
		path += "/"
	}
	if !filepath.IsAbs(path) {
		return Err_ServerFileManagement_NotAbsolutePath
	}
	return nil
}
func (s ServerFileManagement) AbsPath(path string) (string, error) {
	if s.isWindowsDriveRoot(path) {
		path += "/"
	}
	return filepath.Abs(path)
}
func (s ServerFileManagement) isWindowsDriveRoot(path string) bool {
	// 匹配单个驱动器字母后跟冒号的模式
	matched, _ := regexp.MatchString(`^[a-zA-Z]:$`, path)
	return matched
}

// GetValidatePath 验证绝对路径合法性
func (s ServerFileManagement) GetValidatePath(_path string) (string, error) {
	if len(s.RootPath) == 0 {
		return "", Err_ServerFileManagement_NoRootPaths
	}

	// 处理空路径的特殊情况
	if _path == "" {
		return "", nil // 特殊标记表示要显示根目录
	}
	_path = s.convertVirtualPathToCleanRoot(_path)
	// 绝对路径检查
	err := s.IsAbs(_path)
	if err != nil {
		return "", err
	}

	// 转换为绝对路径
	absPath, err := s.AbsPath(_path)
	if err != nil {
		return "", err
	}

	// 标准化路径（使用filepath处理系统相关路径）
	cleanPath := filepath.Clean(absPath)

	// 统一分隔符为系统分隔符
	cleanPath = filepath.ToSlash(cleanPath)

	// 遍历所有预处理后的根路径
	for _, root := range s.RootPath {
		//  路径比较时添加分隔符避免部分匹配
		compareRoot := filepath.ToSlash(root.cleanRoot)
		// 确保以分隔符结尾
		if !strings.HasSuffix(compareRoot, "/") {
			compareRoot += "/"
		}
		// 路径匹配
		if strings.HasPrefix(cleanPath+"/", compareRoot) {
			return cleanPath, nil
		}
	}
	return "", Err_ServerFileManagement_InvalidPath
}

// 获取目录内容（按文件夹优先、名称排序）
func (s ServerFileManagement) PathDir(_path string) ([]ServerFileManagement_FileEntry, error) {
	validPath, err := s.GetValidatePath(_path)
	if err != nil {
		return nil, err
	}
	// 处理空路径的特殊情况
	if _path == "" {
		return s.getRootEntries(), nil
	}

	dir, err := os.Open(validPath)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	entries := make([]ServerFileManagement_FileEntry, 0, len(files))
	for _, fileInfo := range files {
		// 忽略当前目录和父目录
		if fileInfo.Name() == "." || fileInfo.Name() == ".." {
			continue
		}
		modifiedAt := fileInfo.ModTime()
		entry := ServerFileManagement_FileEntry{
			Name:        fileInfo.Name(),
			IsDir:       fileInfo.IsDir(),
			Type:        s.getType(fileInfo),
			Permissions: s.getPermissions(fileInfo),
			Size:        s.getFileSize(fileInfo),
			ModifiedAt:  &modifiedAt,
			Path:        s.convertCleanRootToVirtualPath(filepath.Join(validPath, fileInfo.Name())),
		}

		entries = append(entries, entry)
	}

	// 排序：文件夹优先 → 名称排序
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir && !entries[j].IsDir {
			return true
		}
		if !entries[i].IsDir && entries[j].IsDir {
			return false
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})

	return entries, nil
}

// SearchFiles 搜索指定路径下的文件和文件夹（包括子目录）
func (s ServerFileManagement) SearchFiles(_path string, searchQuery string) ([]ServerFileManagement_FileEntry, error) {
	validPath, err := s.GetValidatePath(_path)
	if err != nil {
		return nil, err
	}

	// 将搜索内容转换为小写，支持大小写不敏感匹配
	searchQuery = strings.ToLower(searchQuery)
	var results = make([]ServerFileManagement_FileEntry, 0)
	err = s.searchFilesRecursive(validPath, searchQuery, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// 递归搜索文件和文件夹
func (s ServerFileManagement) searchFilesRecursive(currentPath string, searchQuery string, results *[]ServerFileManagement_FileEntry) error {
	// 打开当前目录
	dir, err := os.Open(currentPath)
	if err != nil {
		// 捕获权限不足等错误，跳过该文件夹
		if os.IsPermission(err) {
			return nil
		}
		// 其他错误直接返回
		return err
	}
	defer dir.Close()

	// 读取目录内容
	files, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	for _, fileInfo := range files {
		// 忽略当前目录和父目录
		if fileInfo.Name() == "." || fileInfo.Name() == ".." {
			continue
		}

		// 构建完整路径
		fullPath := filepath.Join(currentPath, fileInfo.Name())
		// 检查是否匹配搜索内容
		if strings.Contains(strings.ToLower(fileInfo.Name()), searchQuery) {
			modifiedAt := fileInfo.ModTime()
			entry := ServerFileManagement_FileEntry{
				Name:        fileInfo.Name(),
				Type:        s.getType(fileInfo),
				IsDir:       fileInfo.IsDir(),
				Permissions: s.getPermissions(fileInfo),
				Size:        s.getFileSize(fileInfo),
				ModifiedAt:  &modifiedAt,
				Path:        s.convertCleanRootToVirtualPath(fullPath),
			}
			*results = append(*results, entry)
		}

		// 如果是目录，递归搜索子目录
		if fileInfo.IsDir() {
			if err := s.searchFilesRecursive(fullPath, searchQuery, results); err != nil {
				return err
			}
		}
	}

	return nil
}

// 创建文件
func (s ServerFileManagement) CreateFile(_fileName, _path string) (bool, error) {
	validPath, err := s.GetValidatePath(_path)
	if err != nil {
		return false, err
	}

	// 创建文件
	file, err := os.Create(filepath.Join(validPath, _fileName))
	if err != nil {
		return false, err
	}
	defer file.Close()

	return true, nil
}

// 创建文件夹
func (s ServerFileManagement) CreateFolder(_folderName, _path, permissions string) (bool, error) {
	validPath, err := s.GetValidatePath(_path)
	if err != nil {
		return false, err
	}
	// 权限字符串解析
	perm, err := strconv.ParseUint(permissions, 8, 32)
	if err != nil {
		return false, fmt.Errorf("%v: %v", Err_InvalidPermissionsFormat, err)
	}
	if err := os.Mkdir(filepath.Join(validPath, _folderName), os.FileMode(perm)); err != nil {
		return false, err
	}
	return true, nil
}

// 打开文件
func (s ServerFileManagement) OpenFile(_path string) ([]byte, error) {
	validPath, err := s.GetValidatePath(_path)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(validPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return content, nil
}

// 打开文件并转换为base64
func (s ServerFileManagement) OpenFileToBase64(_path string) (string, error) {
	content, err := s.OpenFile(_path)
	if err != nil {
		return "", err
	}

	// 1. 检测文件 MIME 类型
	mimeType := http.DetectContentType(content[:min(len(content), 512)])
	if mimeType == "" {
		// 备用方案：通过文件扩展名获取 MIME 类型
		ext := filepath.Ext(_path)
		mimeType = mime.TypeByExtension(ext)
	}

	// 2. 编码为 Base64 字符串
	encoded := base64.StdEncoding.EncodeToString(content)

	// 3. 组合成完整的 Data URL 格式
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded), nil
}

// 保存文件
func (s ServerFileManagement) SaveFile(_filePath, _content string) (bool, error) {
	validPath, err := s.GetValidatePath(_filePath)
	if err != nil {
		return false, err
	}

	// 验证目标路径是否为目录（可选）
	if _, err := os.Stat(validPath); err == nil {
		fileInfo, _ := os.Stat(validPath)
		if fileInfo.IsDir() {
			return false, Err_SaveFile_TargetIsDirectory
		}
	}

	file, err := os.Create(validPath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// 使用io.WriteString避免自动添加换行符
	_, err = io.WriteString(file, _content)
	if err != nil {
		return false, err
	}

	return true, nil
}

// 重命名文件/目录
func (s ServerFileManagement) RenameFile(_fileName, _path string) (bool, error) {
	// 1. 验证原路径合法性
	validSrcPath, err := s.GetValidatePath(_path)
	if err != nil {
		return false, err
	}
	// 2. 拆分目录与新文件名
	dir := filepath.Dir(validSrcPath)
	validDstPath := filepath.Join(dir, _fileName)

	// 3. 验证目标路径合法性（确保在允许范围内）
	_, err = s.GetValidatePath(validDstPath)
	if err != nil {
		return false, err
	}

	// 4. 执行重命名操作
	if err := os.Rename(validSrcPath, validDstPath); err != nil {
		return false, err
	}

	return true, nil
}

// 上传文件
func (s ServerFileManagement) UploadFile(fileObj *multipart.FileHeader, filePath, uploadPath string) (bool, error) {
	rootFilePath := path.Join(uploadPath, filePath)
	validPath, err := s.GetValidatePath(rootFilePath)
	if err != nil {
		return false, err
	}
	file, err := fileObj.Open()
	if err != nil {
		return false, err
	}
	defer file.Close()

	//创建多级目录
	dir := filepath.Dir(validPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return false, err
	}

	// 创建目标文件
	targetFile, err := os.Create(validPath)
	if err != nil {
		return false, err
	}
	defer targetFile.Close()

	// 复制文件内容
	_, err = io.Copy(targetFile, file)
	if err != nil {
		return false, err
	}
	return true, nil
}

// 复制文件
func (s ServerFileManagement) Copy(_path string, _files []string) (bool, error) {
	validDstPath, err := s.GetValidatePath(_path)
	if err != nil {
		return false, err
	}
	for _, file := range _files {
		validSrcPath, err := s.GetValidatePath(file)
		if err != nil {
			return false, err
		}

		// 判断是文件还是目录
		info, err := os.Stat(validSrcPath)
		if err != nil {
			return false, err
		}
		validDstPathFile := filepath.Join(validDstPath, info.Name())
		if filepath.ToSlash(validSrcPath) == filepath.ToSlash(validDstPathFile) {
			return false, fmt.Errorf("%w: %s", Err_CopySamePath, validSrcPath)
		}

		if info.IsDir() {
			// 复制目录
			dstDir := filepath.Join(validDstPath, info.Name())
			err = s.copyDir(validSrcPath, dstDir)
		} else {
			// 复制文件
			dstFile := filepath.Join(validDstPath, info.Name())
			err = s.copyFile(validSrcPath, dstFile)
		}

		if err != nil {
			return false, err
		}

	}
	return true, nil
}

// 移动文件
func (s ServerFileManagement) Move(_dstPath string, _files []string) (bool, error) {
	// 验证目标路径
	validDstPath, err := s.GetValidatePath(_dstPath)
	if err != nil {
		return false, err
	}

	for _, srcPath := range _files {
		validSrcPath, err := s.GetValidatePath(srcPath)
		if err != nil {
			return false, err
		}

		// 获取源路径信息
		info, err := os.Stat(validSrcPath)
		if err != nil {
			return false, err
		}

		// 构建目标路径（保留原文件/目录名）
		dstPath := filepath.Join(validDstPath, info.Name())

		if filepath.ToSlash(validSrcPath) == filepath.ToSlash(dstPath) {
			return false, fmt.Errorf("%w: %s", Err_MoveSamePath, validSrcPath)
		}

		// 1. 先执行复制操作
		if err := s.moveInternal(validSrcPath, dstPath); err != nil {
			return false, err
		}

		// 2. 复制成功后删除原文件/目录
		if err := os.RemoveAll(validSrcPath); err != nil {
			return false, fmt.Errorf("%v: %v", Err_MoveDeleteFailed, err)
		}
	}

	return true, nil
}

// 压缩文件
func (s ServerFileManagement) CompressFile(_dstPath string, _files []string) (bool, error) {
	validDstPath, err := s.GetValidatePath(_dstPath)
	if err != nil {
		return false, err
	}

	// 生成压缩包路径（如：compressed_files_20231010.zip）
	timestamp := time.Now().Format("20060102150405")
	zipFileName := fmt.Sprintf("compressed_files_%s.zip", timestamp)
	zipFilePath := filepath.Join(validDstPath, zipFileName)

	// 创建 ZIP 文件
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return false, err
	}
	defer zipFile.Close()

	// 初始化 ZIP 写入器
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, srcPathStr := range _files {
		validSrcPath, err := s.GetValidatePath(srcPathStr)
		if err != nil {
			return false, err
		}

		fileInfo, err := os.Stat(validSrcPath)
		if err != nil {
			return false, err
		}

		// 根据文件类型添加到 ZIP
		if fileInfo.IsDir() {
			// 递归添加目录内容
			basePath := filepath.Dir(validSrcPath)
			if err := s.addDirToZip(zipWriter, validSrcPath, basePath); err != nil {
				return false, err
			}
		} else {
			// 添加单个文件
			filename := filepath.Base(validSrcPath)
			if err := s.addFileToZip(zipWriter, validSrcPath, filename); err != nil {
				return false, err
			}
		}
	}

	return true, nil
}

// 设置权限
func (s ServerFileManagement) PermissionsFile(_files []string, _permissions string, subFiles bool) (bool, error) {
	// 1. 解析权限模式
	perm, err := strconv.ParseUint(_permissions, 8, 32)
	if err != nil {
		return false, fmt.Errorf("%v: %v", Err_InvalidPermissionsFormat, err)
	}

	// 2. 遍历所有目标路径
	for _, filePath := range _files {
		validPath, err := s.GetValidatePath(filePath)
		if err != nil {
			return false, err
		}

		// 3. 设置权限（递归处理子项）
		if err := s.setPermissionsRecursive(validPath, os.FileMode(perm), subFiles); err != nil {
			return false, fmt.Errorf("%v: %v", Err_SetPermissionsFailed, err)
		}
	}

	return true, nil
}

// 解压缩文件到当前目录
func (s ServerFileManagement) UnCompressFile(_path string) (bool, error) {
	// 1. 验证路径合法性
	validPath, err := s.GetValidatePath(_path)
	if err != nil {
		return false, err
	}

	// 2. 检查是否是有效文件
	fileInfo, err := os.Stat(validPath)
	if err != nil || fileInfo.IsDir() || fileInfo.Size() == 0 {
		return false, Err_InvalidZipFile
	}

	// 3. 获取解压目标目录（压缩包所在目录）
	extractDir := filepath.Dir(validPath)

	// 4. 打开ZIP文件
	zipReader, err := zip.OpenReader(validPath)
	if err != nil {
		return false, err
	}
	defer zipReader.Close()

	// 5. 遍历所有条目进行解压
	for _, file := range zipReader.File {
		// 构建目标路径
		targetPath := filepath.Join(extractDir, file.Name)

		// 安全验证（防止路径遍历攻击）
		_, err := s.GetValidatePath(targetPath)
		if err != nil {
			return false, fmt.Errorf("%v: %s", Err_InvalidZipEntryPath, file.Name)
		}

		if file.FileInfo().IsDir() {
			// 创建目录
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return false, err
			}
		} else {
			// 确保父目录存在
			dir := filepath.Dir(targetPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return false, err
			}

			// 写入文件内容
			srcFile, err := file.Open()
			if err != nil {
				return false, err
			}
			defer srcFile.Close()

			dstFile, err := os.Create(targetPath)
			if err != nil {
				return false, err
			}
			defer dstFile.Close()

			if _, err := io.Copy(dstFile, srcFile); err != nil {
				return false, err
			}
		}
	}

	return true, nil
}

// 删除文件
func (s ServerFileManagement) DeleteFile(_files []string) (bool, error) {
	for _, file := range _files {
		validPath, err := s.GetValidatePath(file)
		if err != nil {
			return false, err
		}
		if err := os.RemoveAll(validPath); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (s *ServerFileManagement) copyFile(src, dst string) error {
	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// 创建目标文件（覆盖模式）
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// 复制文件内容
	_, err = io.Copy(dstFile, srcFile)
	return err
}
func (s *ServerFileManagement) copyDir(src, dst string) error {
	// 创建目标目录
	err := os.MkdirAll(dst, 0755)
	if err != nil {
		return err
	}

	// 遍历源目录
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		info, err := os.Stat(srcPath)
		if err != nil {
			return err
		}

		if info.IsDir() {
			// 递归复制子目录
			err = s.copyDir(srcPath, dstPath)
		} else {
			// 复制文件
			err = s.copyFile(srcPath, dstPath)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *ServerFileManagement) moveInternal(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return s.copyDir(src, dst)
	}
	return s.copyFile(src, dst)
}

// 递归添加目录及其子文件到 ZIP
func (s *ServerFileManagement) addDirToZip(zipWriter *zip.Writer, srcDir, basePath string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		relativePath := strings.TrimPrefix(srcPath, basePath)
		relativePath = filepath.ToSlash(relativePath) // 统一分隔符为 /

		if entry.IsDir() {
			// 递归处理子目录
			err = s.addDirToZip(zipWriter, srcPath, basePath)
		} else {
			// 添加文件
			err = s.addFileToZip(zipWriter, srcPath, relativePath)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// 添加单个文件到 ZIP
func (s *ServerFileManagement) addFileToZip(zipWriter *zip.Writer, srcPath, relativePath string) error {
	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建 ZIP 文件条目
	zipFile, err := zipWriter.Create(relativePath)
	if err != nil {
		return err
	}

	// 复制文件内容到 ZIP
	_, err = io.Copy(zipFile, file)
	return err
}

// 递归设置权限的辅助函数
func (s ServerFileManagement) setPermissionsRecursive(path string, perm os.FileMode, recursive bool) error {
	// 基础权限设置
	if err := os.Chmod(path, perm); err != nil {
		return err
	}

	// 递归处理子项（仅目录且开启递归时）
	if recursive {
		fileInfo, err := os.Stat(path)
		if err != nil {
			return err
		}
		if fileInfo.IsDir() {
			entries, err := os.ReadDir(path)
			if err != nil {
				return err
			}
			for _, entry := range entries {
				subPath := filepath.Join(path, entry.Name())
				if err := s.setPermissionsRecursive(subPath, perm, recursive); err != nil {
					return err // 停止传播错误（或收集错误）
				}
			}
		}
	}

	return nil
}

// 空路径的特殊情况,返回根目录信息
func (s *ServerFileManagement) getRootEntries() []ServerFileManagement_FileEntry {
	entries := make([]ServerFileManagement_FileEntry, 0, len(s.RootPath))

	for _, root := range s.RootPath {
		// 获取根目录实际信息
		fi, err := os.Stat(root.cleanRoot)
		if err != nil {
			continue // 跳过无效路径
		}
		modifiedAt := fi.ModTime()
		rcr := s.convertCleanRootToVirtualPath(root.cleanRoot)
		entries = append(entries, ServerFileManagement_FileEntry{
			Name:        rcr,
			Type:        "directory",
			IsDir:       true,
			Permissions: s.getPermissions(fi),
			Size:        0,
			ModifiedAt:  &modifiedAt,
			Path:        rcr,
		})
	}
	return entries
}

// 辅助方法：获取文件类型字符串
func (s ServerFileManagement) getType(fileInfo os.FileInfo) string {
	if fileInfo.IsDir() {
		return "directory"
	}
	return "file"
}

// 辅助方法：获取权限字符串（八进制）
func (s ServerFileManagement) getPermissions(fileInfo os.FileInfo) string {
	mode := fileInfo.Mode().Perm()
	return fmt.Sprintf("%03o", mode)
}

// 辅助方法：获取文件大小（文件夹返回0）
func (s ServerFileManagement) getFileSize(fileInfo os.FileInfo) int64 {
	if fileInfo.IsDir() {
		return 0
	}
	return fileInfo.Size()
}
