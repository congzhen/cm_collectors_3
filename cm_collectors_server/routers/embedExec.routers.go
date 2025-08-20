package routers

import (
	"embed"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type EmbeddedFolderCreateStaticRoutes struct {
	Router     *gin.Engine
	Efs        embed.FS
	FolderName string
}

func (ef EmbeddedFolderCreateStaticRoutes) Register() error {
	return ef.registerStaticFiles(ef.FolderName)
}

func (ef EmbeddedFolderCreateStaticRoutes) registerStaticFiles(rootPath string) error {
	// 遍历当前目录
	entries, err := ef.Efs.ReadDir(rootPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return err
		}
		embedFilePath := ef.embedFilePathJoin(rootPath, info.Name())

		if info.IsDir() {
			// 遍历子目录
			if err := ef.registerStaticFiles(embedFilePath); err != nil {
				return err
			}
		} else {
			// 注册静态文件,去掉目录名称(folderName)
			routerPath := embedFilePath[len(ef.FolderName):]
			ef.Router.GET(routerPath, func(c *gin.Context) {
				// 读取文件内容
				content, err := ef.Efs.ReadFile(embedFilePath)
				if err != nil {
					c.Status(http.StatusInternalServerError)
					return
				}

				// 设置响应头
				c.Header("Content-Type", ef.getContentType(info.Name()))
				c.Data(http.StatusOK, ef.getContentType(info.Name()), content)

			})
		}
	}

	return nil
}

func (EmbeddedFolderCreateStaticRoutes) embedFilePathJoin(s1, s2 string) string {
	return s1 + "/" + s2
}
func (EmbeddedFolderCreateStaticRoutes) getContentType(fileName string) string {
	ext := filepath.Ext(fileName)
	switch ext {
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}
