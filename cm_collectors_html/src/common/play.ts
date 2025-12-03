import { E_resourceDramaSeriesType, E_resourceOpenMode, E_resourceOpenMode_SoftType, type I_playVideoData, type T_VideoPlayMode } from "@/dataType/app.dataType";
import type { I_resource } from "@/dataType/resource.dataType";
import { appDataServer } from "@/server/app.server";
import { appStoreData } from "@/storeData/app.storeData";
import { ElMessage, ElNotification } from "element-plus";
import router from '@/router';
import { eventBus } from "@/main";
import { isMobile } from "@/assets/mobile";

export const playResource = async (resource: I_resource, dramaSeriesId: string = '') => {
  const store = {
    appStoreData: appStoreData(),
  }
  let openMode: E_resourceOpenMode = E_resourceOpenMode.Soft;
  switch (resource.mode) {
    case E_resourceDramaSeriesType.Movies:
      openMode = store.appStoreData.currentConfigApp.openResModeMovies;
      break;
    case E_resourceDramaSeriesType.Comic:
      openMode = store.appStoreData.currentConfigApp.openResModeComic;
      break;
    case E_resourceDramaSeriesType.Atlas:
      openMode = store.appStoreData.currentConfigApp.openResModeAtlas;
      break;
    case E_resourceDramaSeriesType.Files:
      openMode = E_resourceOpenMode.System;
      break;
    case E_resourceDramaSeriesType.NetDisk:
      break;
    case E_resourceDramaSeriesType.VideoLink:
      break;
    default:
      ElMessage.error('未知的资源类型');
      return;
  }

  if (dramaSeriesId == '' && resource.dramaSeries && resource.dramaSeries.length > 0) {
    dramaSeriesId = resource.dramaSeries[0].id;
  }

  playUpdate(resource.id, dramaSeriesId)
  if (isMobile()) {
    router.push(`/play/${resource.mode}Mobile/${resource.id}` + (dramaSeriesId != '' ? `/${dramaSeriesId}` : ''));
    return
  }

  if (openMode == E_resourceOpenMode.Soft) {
    if (store.appStoreData.currentConfigApp.openResModeMovies_SoftType == E_resourceOpenMode_SoftType.Dialog) {
      eventBus.emit('resource-dialog-play-start', { resourceId: resource.id, dramaSeriesId });
    } else {
      router.push(`/play/${resource.mode}/${resource.id}` + (dramaSeriesId != '' ? `/${dramaSeriesId}` : ''));
    }
  } else if (openMode == E_resourceOpenMode.CloundPlay) {
    if (dramaSeriesId == '') {
      ElNotification({
        message: `无播放源`,
        type: 'error',
      })
      return;
    }
    eventBus.emit('playClound', { resourceId: resource.id, dramaSeriesId, playSrc: getPlayVideoURL(dramaSeriesId, 'mp4') });
  } else if (openMode == E_resourceOpenMode.System) {
    const eln = ElNotification({
      message: `正在打开资源 ${resource.title} ...`,
      type: 'success',
    })
    const result = await appDataServer.playOpenResource(resource.id, dramaSeriesId);
    if (!result || !result.status) {
      eln.close();
      ElNotification({
        message: `${result.msg} ${resource.title}`,
        type: 'error',
      })
      return;
    }
  }
  return;
}

export const playUpdate = (resourceId: string, dramaSeriesId: string) => {
  return appDataServer.playUpdate(resourceId, dramaSeriesId);
}

export const openInPlayerDramaSeries = async (dramaSeriesId: string): Promise<boolean> => {
  if (dramaSeriesId == '') return false;
  const eln = ElNotification({
    message: `正在打开频 ...`,
    type: 'success',
  })
  const result = await appDataServer.playOpenDramaSeries(dramaSeriesId);
  if (!result || !result.status) {
    eln.close();
    ElNotification({
      message: `打开失败: ${result.msg}`,
      type: 'error',
    })
    return false;
  } else {
    setTimeout(() => {
      eln.close();
    }, 1500);
  }
  return true;
}

export const playOpenResourceFolder = async (resourceId: string) => {
  const result = await appDataServer.playOpenResourceFolder(resourceId);
  if (!result || !result.status) {
    ElMessage.error(result.msg);
    return;
  }
}

export const getPlayVideoURLAndType = async (dramaSeriesId: string): Promise<I_playVideoData> => {
  const result = await appDataServer.playVideoInfo(dramaSeriesId);
  if (!result || !result.status) {
    ElMessage.error(result.msg);
    return {
      playUrl: '',
      playType: 'mp4',
    };
  }
  let playType: T_VideoPlayMode = 'mp4';
  if (!result.data.is_web) {
    playType = 'm3u8';
  }
  return {
    playUrl: getPlayVideoURL(dramaSeriesId, playType),
    playType: playType,
  }
}

export const getPlayVideoURL = (dramaSeriesId: string, type: T_VideoPlayMode = 'mp4', fileName: string = 'v') => {
  if (type == 'mp4') {
    return `/api/video/mp4/${dramaSeriesId}/${fileName}.mp4`;
  } else {
    return `/api/video/m3u8/${dramaSeriesId}/${fileName}.m3u8`;
  }
}

// 云播检查
export const playCloudCheck = () => {
  const playCloudCheck = localStorage.getItem('playCloudCheck')
  return playCloudCheck ? true : false;
}
// 设置云播检查完成
export const setPlayCloudCheckComplete = () => {
  localStorage.setItem('playCloudCheck', 'true');
}
