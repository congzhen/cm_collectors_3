import { E_resourceDramaSeriesType, E_resourceOpenMode, E_resourceOpenMode_SoftType } from "@/dataType/app.dataType";
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

export const getPlayVideoURL = (dramaSeriesId: string, type = 'mp4') => {
  return `/api/video/${type}/${dramaSeriesId}/v.${type}`;
}
