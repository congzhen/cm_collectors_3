package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/errorMessage"
	"cm_collectors_server/models"
	processorsFFmpeg "cm_collectors_server/processorsFFmpeg"
	"cm_collectors_server/utils"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/skratchdot/open-golang/open"
)

type Play struct{}

type PlayVideoInfo struct {
	VideoBasicInfo processorsFFmpeg.VideoBasicInfo `json:"video_basic_info"`
	IsWeb          bool                            `json:"is_web"`
}

// AllowServerOpenFile 检测是否允许服务器打开文件
func (Play) AllowServerOpenFile() error {
	if core.Config.General.NotAllowServerOpenFile {
		return errorMessage.Err_Current_Server_Has_Been_Set_To_Disallow_This_Peration
	}
	return nil
}

func (p Play) PlayVideoInfo(dramaSeriesId string) (*PlayVideoInfo, error) {
	playSrc, err := ResourcesDramaSeries{}.GetSrc(dramaSeriesId)
	if err != nil {
		return nil, err
	}
	if !utils.FileExists(playSrc) {
		return nil, errorMessage.Err_Resources_Play_Src_Error
	}
	pf_videoInfo := processorsFFmpeg.VideoInfo{}
	// 设置支持的编解码器
	pf_videoInfo.SetSupportedVideoCodecs(core.Config.Play.PlayVideoFormats)
	pf_videoInfo.SetSupportedAudioCodecs(core.Config.Play.PlayAudioFormats)
	videoFormatInfo, err := pf_videoInfo.GetVideoFormatInfo(playSrc)
	if err != nil {
		return nil, err
	}
	isWeb := pf_videoInfo.IsWebCompatible(videoFormatInfo)
	videoBasicInfo := pf_videoInfo.GetVideoBasicInfoByVideoFormatInfo(videoFormatInfo)
	return &PlayVideoInfo{VideoBasicInfo: videoBasicInfo, IsWeb: isWeb}, nil
}

func (p Play) PlayUpdate(resourceId, dramaSeriesId string) error {
	resourceInfo, err := Resources{}.Info(resourceId)
	if err != nil {
		return err
	}
	var dramaSeries = models.ResourcesDramaSeries{}
	if dramaSeriesId != "" {
		for _, v := range resourceInfo.ResourcesDramaSeries {
			if v.ID == dramaSeriesId {
				dramaSeries = v
			}
		}
	} else if len(resourceInfo.ResourcesDramaSeries) > 0 {
		dramaSeries = (resourceInfo.ResourcesDramaSeries)[0]
	} else {
		return errors.New("没有找到播放剧集")
	}
	err = Resources{}.UpdateResourcePlay(resourceInfo, dramaSeries.Src)
	if err != nil {
		return err
	}
	return nil
}

// PlayOpenResource 打开指定资源进行播放
func (p Play) PlayOpenResource(resourceId, dramaSeriesId string) error {
	err := p.AllowServerOpenFile()
	if err != nil {
		return err
	}
	resourceInfo, err := Resources{}.Info(resourceId)
	if err != nil {
		return err
	}

	playSrc, err := p.getPlaySource(resourceInfo, dramaSeriesId)
	if err != nil {
		return err
	}

	// 检查播放源路径是否存在
	err = p.checkPlaySourceExists(playSrc)
	if err != nil {
		return err
	}

	// 调用系统执行该路径文件
	return open.Run(playSrc)
}

func (p Play) PlayOpenDramaSeries(dramaSeriesId string) error {
	err := p.AllowServerOpenFile()
	if err != nil {
		return err
	}
	info, err := ResourcesDramaSeries{}.Info(dramaSeriesId)
	if err != nil {
		return err
	}
	return open.Run(info.Src)
}

// PlayOpenResourceFolder 打开资源所在文件夹
func (p Play) PlayOpenResourceFolder(resourceId string) error {
	err := p.AllowServerOpenFile()
	if err != nil {
		return err
	}
	resourceInfo, err := Resources{}.Info(resourceId)
	if err != nil {
		return err
	}

	playSrc, err := p.getPlaySource(resourceInfo, "")
	if err != nil {
		return err
	}
	// 清理路径
	playSrc = filepath.Clean(playSrc)
	// 检查播放源路径是否存在
	err = p.checkPlaySourceExists(playSrc)
	if err != nil {
		folderPath := filepath.Dir(playSrc)
		return open.Run(folderPath)
	}
	// 使用系统命令直接打开文件夹并选中文件，更好地处理特殊字符
	err = p.openFolderAndSelectFile(playSrc)
	if err != nil {
		folderPath := filepath.Dir(playSrc)
		return open.Run(folderPath)
	}
	return nil
}

func (p Play) openFolderAndSelectFile(filePath string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Windows下使用 explorer 命令并添加 /select 参数来选中文件
		cmd = exec.Command("explorer", "/select,", filePath)
	case "darwin":
		// macOS下使用 open 命令打开文件所在文件夹
		// macOS没有直接选中文件的命令行选项
		folderPath := filepath.Dir(filePath)
		cmd = exec.Command("open", folderPath)
	default:
		// Linux下使用 xdg-open 命令打开文件所在文件夹
		// 大多数Linux文件管理器没有直接选中文件的命令行选项
		folderPath := filepath.Dir(filePath)
		cmd = exec.Command("xdg-open", folderPath)
	}

	// 添加超时控制，防止进程挂起
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd = exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)

	// 执行命令
	err := cmd.Run()

	// 特殊处理Windows下的explorer命令
	// Windows的explorer命令有时会返回exit status 1，即使成功执行
	if runtime.GOOS == "windows" {
		if exitError, ok := err.(*exec.ExitError); ok {
			// 如果是exit status 1，我们将其视为成功
			if exitError.ExitCode() == 1 {
				return nil
			}
		}
	}

	// 确保命令执行完成后资源得到释放
	if cmd != nil && cmd.Process != nil {
		// 等待一段时间确保进程完全退出
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case <-done:
			// 正常完成
		case <-time.After(1 * time.Second):
			// 超时，强制杀死进程
			cmd.Process.Kill()
		}
	}

	return err
}

// 检测源路径是否存在
func (p Play) checkPlaySourceExists(playSrc string) error {
	if _, err := os.Stat(playSrc); os.IsNotExist(err) {
		return errorMessage.Err_Resources_Play_Src_Error
	}
	return nil
}

// getPlaySource 获取播放源地址
func (p Play) getPlaySource(resourceInfo *models.Resources, dramaSeriesId string) (string, error) {
	playSrc := ""

	// 获取播放源
	if dramaSeriesId == "" {
		if len(resourceInfo.ResourcesDramaSeries) > 0 {
			playSrc = resourceInfo.ResourcesDramaSeries[0].Src
		}
	} else {
		for _, v := range resourceInfo.ResourcesDramaSeries {
			if v.ID == dramaSeriesId {
				playSrc = v.Src
				break
			}
		}
	}

	// 播放源不存在
	if playSrc == "" {
		return "", errorMessage.Err_Resources_Play_DramaSeries_Not_Found
	}

	return playSrc, nil
}
