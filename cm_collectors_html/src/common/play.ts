import { E_resourceDramaSeriesType, E_resourceOpenMode } from "@/dataType/app.dataType";
import type { I_resource } from "@/dataType/resource.dataType";
import { appDataServer } from "@/server/app.server";
import { appStoreData } from "@/storeData/app.storeData";
import { ElMessage, ElNotification } from "element-plus";
import router from '@/router';

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
      break;
    case E_resourceDramaSeriesType.NetDisk:
      break;
    case E_resourceDramaSeriesType.VideoLink:
      break;
    default:
      ElMessage.error('未知的资源类型');
      return;
  }

  if (openMode == E_resourceOpenMode.Soft) {
    router.push(`/play/${E_resourceDramaSeriesType.Movies}/${resource.id}` + (dramaSeriesId != '' ? `/${dramaSeriesId}` : ''));
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

export const playOpenResourceFolder = async (resourceId: string) => {
  const result = await appDataServer.playOpenResourceFolder(resourceId);
  if (!result || !result.status) {
    ElMessage.error(result.msg);
    return;
  }
}
