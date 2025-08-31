package processors

import (
	"cm_collectors_server/errorMessage"
	"cm_collectors_server/models"
	"os"
	"path/filepath"

	"github.com/skratchdot/open-golang/open"
)

type Play struct{}

// PlayOpenResource 打开指定资源进行播放
func (p Play) PlayOpenResource(resourceId, dramaSeriesId string) error {
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

	err = Resources{}.UpdateResourcePlay(resourceInfo, playSrc)
	if err != nil {
		return err
	}

	// 调用系统执行该路径文件
	return open.Run(playSrc)
}

func (p Play) PlayOpenDramaSeries(dramaSeriesId string) error {
	info, err := ResourcesDramaSeries{}.Info(dramaSeriesId)
	if err != nil {
		return err
	}
	return open.Run(info.Src)
}

// PlayOpenResourceFolder 打开资源所在文件夹
func (p Play) PlayOpenResourceFolder(resourceId string) error {
	resourceInfo, err := Resources{}.Info(resourceId)
	if err != nil {
		return err
	}

	playSrc, err := p.getPlaySource(resourceInfo, "")
	if err != nil {
		return err
	}

	// 检查播放源路径是否存在
	err = p.checkPlaySourceExists(playSrc)
	if err != nil {
		return err
	}

	// 获取文件所在的目录路径
	folderPath := filepath.Dir(playSrc)

	// 调用系统命令打开文件夹
	return open.Run(folderPath)
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
