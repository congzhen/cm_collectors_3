package controllers

import (
	serverfilemanagement "cm_collectors_server/api/serverFileManagement"
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/response"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"
)

type ServerFileManagement struct{}

func (ServerFileManagement) DefaultPathEntrySlc() []serverfilemanagement.ServerFileManagement_PathEntry {
	pathEntrySlc := []serverfilemanagement.ServerFileManagement_PathEntry{}
	if runtime.GOOS == "windows" {
		// 简化处理：列出 A-Z 盘符，判断是否存在
		for c := 'C'; c <= 'Z'; c++ {
			path := string(c) + `:\`
			if _, err := os.Stat(path); err == nil {
				path = filepath.ToSlash(path)
				pathEntrySlc = append(pathEntrySlc, serverfilemanagement.ServerFileManagement_PathEntry{
					RealPath:    path,
					VirtualPath: path,
				})
			}
		}
	} else {
		// Linux/macOS: 列出根目录下内容
		files, err := os.ReadDir("/")
		if err != nil {
			return pathEntrySlc
		}
		for _, f := range files {
			path := filepath.ToSlash("/" + f.Name())
			pathEntrySlc = append(pathEntrySlc, serverfilemanagement.ServerFileManagement_PathEntry{
				RealPath:    path,
				VirtualPath: path,
			})
		}
	}
	return pathEntrySlc
}

func (t ServerFileManagement) pathEntrySlc() []serverfilemanagement.ServerFileManagement_PathEntry {
	rootPathLen := len(core.Config.ServerFileManagement.RootPath)
	if rootPathLen == 0 {
		return t.DefaultPathEntrySlc()
	}
	pathEntrySlc := make([]serverfilemanagement.ServerFileManagement_PathEntry, rootPathLen)
	for i, v := range core.Config.ServerFileManagement.RootPath {
		pathEntrySlc[i].RealPath = v.RealPath
		pathEntrySlc[i].VirtualPath = v.VirtualPath
	}
	return pathEntrySlc
}
func (t ServerFileManagement) newServerFileManagement() *serverfilemanagement.ServerFileManagement {
	return serverfilemanagement.NewServerFileManagement(t.pathEntrySlc())
}

func (t ServerFileManagement) PathDir(c *gin.Context) {
	var par datatype.ServerFileManagement_PathFiles

	if err := c.ShouldBindJSON(&par); err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	sfm := t.newServerFileManagement()
	list, err := sfm.PathDir(par.Path)
	if err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	response.Result(true, 0, "success", list, c)
}

func (t ServerFileManagement) SearchFiles(c *gin.Context) {
	var par datatype.ServerFileManagement_SearchFiles

	if err := c.ShouldBindJSON(&par); err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	sfm := t.newServerFileManagement()
	list, err := sfm.SearchFiles(par.Path, par.SearchQuery)
	if err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	response.Result(true, 0, "success", list, c)
}

func (t ServerFileManagement) CreateFile(c *gin.Context) {
	var par datatype.ServerFileManagement_CreateFile

	if err := c.ShouldBindJSON(&par); err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	sfm := t.newServerFileManagement()
	b, err := sfm.CreateFile(par.Name, par.Path)
	if err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	response.Result(true, 0, "success", b, c)
}

func (t ServerFileManagement) CreateFolder(c *gin.Context) {
	var par datatype.ServerFileManagement_CreateFolder

	if err := c.ShouldBindJSON(&par); err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	sfm := t.newServerFileManagement()
	b, err := sfm.CreateFolder(par.Name, par.Path, par.Permissions)
	if err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	response.Result(true, 0, "success", b, c)
}

func (t ServerFileManagement) OpenFile(c *gin.Context) {
	var par datatype.ServerFileManagement_OpenFile
	if err := c.ShouldBindJSON(&par); err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	sfm := t.newServerFileManagement()
	var data interface{}
	switch par.ReturnType {
	case "base64":
		base64, err := sfm.OpenFileToBase64(par.FilePath)
		if err != nil {
			response.Result(false, 100, err.Error(), nil, c)
			return
		}
		data = base64
	case "text", "":
		// 默认返回文本，显式指定编码
		b, err := sfm.OpenFile(par.FilePath)
		if err != nil {
			response.Result(false, 100, err.Error(), nil, c)
			return
		}
		data = string(b)
	default:
		// 直接返回原始二进制数据（如图片、二进制文件）
		b, err := sfm.OpenFile(par.FilePath)
		if err != nil {
			response.Result(false, 100, err.Error(), nil, c)
			return
		}
		data = b
	}
	response.Result(true, 0, "success", data, c)
}
func (t ServerFileManagement) SaveFile(c *gin.Context) {
	var par datatype.ServerFileManagement_SaveFile
	if err := c.ShouldBindJSON(&par); err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	sfm := t.newServerFileManagement()
	b, err := sfm.SaveFile(par.FilePath, par.Content)
	if err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	response.Result(true, 0, "success", b, c)
}

func (t ServerFileManagement) UploadFile(c *gin.Context) {
	fileObj, err := c.FormFile("file")
	if err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	filePath := c.PostForm("file_path")
	uploadPath := c.PostForm("upload_path")
	sfm := t.newServerFileManagement()
	b, err := sfm.UploadFile(fileObj, filePath, uploadPath)
	if err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	response.Result(true, 0, "success", b, c)
}

func (t ServerFileManagement) RenameFile(c *gin.Context) {
	var par datatype.ServerFileManagement_RenameFile
	if err := c.ShouldBindJSON(&par); err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	sfm := t.newServerFileManagement()
	b, err := sfm.RenameFile(par.Name, par.Path)
	if err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	response.Result(true, 0, "success", b, c)
}

func (t ServerFileManagement) DownloadFile(c *gin.Context) {
	filePath := c.Query("filePath")
	if filePath == "" {
		response.Result(false, 100, "path is empty", nil, c)
		return
	}
	sfm := t.newServerFileManagement()
	validPath, err := sfm.GetValidatePath(filePath)
	if err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	// 设置响应头触发下载
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(validPath)))
	c.Header("Accept-Ranges", "bytes") // 支持断点续传

	// 直接使用标准库自动处理Range请求
	http.ServeFile(c.Writer, c.Request, validPath)

}

func (t ServerFileManagement) CopyFile(c *gin.Context) {
	var par datatype.ServerFileManagement_Action
	if err := c.ShouldBindJSON(&par); err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	sfm := t.newServerFileManagement()
	b, err := sfm.Copy(par.Path, par.Files)
	if err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	response.Result(true, 0, "success", b, c)
}

func (t ServerFileManagement) MoveFile(c *gin.Context) {
	var par datatype.ServerFileManagement_Action
	if err := c.ShouldBindJSON(&par); err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	sfm := t.newServerFileManagement()
	b, err := sfm.Move(par.Path, par.Files)
	if err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	response.Result(true, 0, "success", b, c)
}

func (t ServerFileManagement) CompressFile(c *gin.Context) {
	var par datatype.ServerFileManagement_Action
	if err := c.ShouldBindJSON(&par); err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	sfm := t.newServerFileManagement()
	b, err := sfm.CompressFile(par.Path, par.Files)
	if err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	response.Result(true, 0, "success", b, c)
}

func (t ServerFileManagement) PermissionsFile(c *gin.Context) {
	var par datatype.ServerFileManagement_Permissions
	if err := c.ShouldBindJSON(&par); err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	sfm := t.newServerFileManagement()
	b, err := sfm.PermissionsFile(par.Files, par.Permissions, par.SubFiles)
	if err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	response.Result(true, 0, "success", b, c)
}

func (t ServerFileManagement) UnCompressFile(c *gin.Context) {
	var par datatype.ServerFileManagement_UnCompressFile
	if err := c.ShouldBindJSON(&par); err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	sfm := t.newServerFileManagement()
	b, err := sfm.UnCompressFile(par.File)
	if err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	response.Result(true, 0, "success", b, c)
}

func (t ServerFileManagement) DeleteFile(c *gin.Context) {
	var par datatype.ServerFileManagement_Action
	if err := c.ShouldBindJSON(&par); err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	sfm := t.newServerFileManagement()
	b, err := sfm.DeleteFile(par.Files)
	if err != nil {
		response.Result(false, 100, err.Error(), nil, c)
		return
	}
	response.Result(true, 0, "success", b, c)
}
