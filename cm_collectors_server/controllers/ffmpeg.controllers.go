package controllers

import (
	"cm_collectors_server/datatype"
	"cm_collectors_server/internal/coverpreview"
	processorsffmpeg "cm_collectors_server/processorsFFmpeg"
	"cm_collectors_server/response"

	"github.com/gin-gonic/gin"
)

type FFmpeg struct{}

var coverPreviewCache = coverpreview.Cache{Root: "./runtime/cache/coverThumbnails"}

func (FFmpeg) GetVideoThumbnails(c *gin.Context) {
	var par datatype.ReqParam_FFmpeg_VideoKeyFramePosters
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	base64Slc, err := coverPreviewCache.Get(par.VideoPath, par.FrameCount, par.Refresh, func() ([]string, error) {
		return processorsffmpeg.Thumbnail{}.ExtractThumbnailAsBase64(par.VideoPath, par.FrameCount, 1)
	})
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(base64Slc, c)
}

func (FFmpeg) CoverPreviewSettings(c *gin.Context) {
	value, err := coverPreviewCache.Settings()
	if ResError(c, err) == nil {
		response.OkWithData(value, c)
	}
}
func (FFmpeg) SaveCoverPreviewSettings(c *gin.Context) {
	var value coverpreview.Settings
	if ParameterHandleShouldBindJSON(c, &value) != nil {
		return
	}
	if ResError(c, coverPreviewCache.SaveSettings(value)) == nil {
		response.OkWithData(true, c)
	}
}
func (FFmpeg) CoverPreviewStatus(c *gin.Context) {
	value, err := coverPreviewCache.Status()
	if ResError(c, err) == nil {
		response.OkWithData(value, c)
	}
}
func (FFmpeg) ClearCoverPreview(c *gin.Context) {
	if ResError(c, coverPreviewCache.Clear()) == nil {
		response.OkWithData(true, c)
	}
}
