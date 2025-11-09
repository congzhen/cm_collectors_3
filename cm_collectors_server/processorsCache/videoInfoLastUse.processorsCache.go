package processorscache

import (
	processorsffmpeg "cm_collectors_server/processorsFFmpeg"
	"fmt"
	"time"
)

// 文件句柄释放时间
const videoInfoHandleExpiration time.Duration = 30 * time.Second

// 文件句柄清理时间
const videoInfoHandleCleanupInterval time.Duration = 60 * time.Second

var videoInfoLastUseHandle = NewGenericLastUseCache[processorsffmpeg.VideoFormatInfo]("videoInfo", videoInfoHandleExpiration, videoInfoHandleCleanupInterval)

type CacheVideoInfoLastUse struct {
}

// 获取缓存的文件句柄或创建新的
func (CacheVideoInfoLastUse) GetVideoInfoHandle(src string) (processorsffmpeg.VideoFormatInfo, error) {
	if value, ok := videoInfoLastUseHandle.Get(src); ok {
		fmt.Println("######################### 视频信息缓存:", src)
		return value, nil
	}
	// 获取
	videoFormatInfo, err := processorsffmpeg.VideoInfo{}.GetVideoFormatInfo(src)
	if err != nil {
		return processorsffmpeg.VideoFormatInfo{}, err
	}
	fmt.Println("######################### 视频信息缓存:", src)
	videoInfoLastUseHandle.Set(src, videoFormatInfo)
	return videoFormatInfo, nil
}
