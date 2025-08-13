package controllers

import (
	"cm_collectors_server/datatype"
	processorsffmpeg "cm_collectors_server/processorsFFmpeg"
	"cm_collectors_server/response"

	"github.com/gin-gonic/gin"
)

type FFmpeg struct{}

func (FFmpeg) GetVideoKeyFramePosters(c *gin.Context) {
	var par datatype.ReqParam_FFmpeg_VideoKeyFramePosters
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	base64Slc, err := processorsffmpeg.KeyFrame{}.ExtractKeyframesAsBase64(par.VideoPath, par.FrameCount, 5)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(base64Slc, c)
}
