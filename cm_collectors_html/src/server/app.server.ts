import request, { type IRequest } from "@/assets/request";
import type { I_appSystemConfig, I_appData, I_playVideoInfo } from "@/dataType/app.dataType";
import type { I_databaseCleanupForm } from "@/dataType/other.dataType";
const routerGroupUri = '';
export const appDataServer = {
  init: async () => {
    return await request<I_appData>({
      url: `${routerGroupUri}/app/data`,
      method: 'get',
    });
  },
  playOpenResource: async (resourceId: string, dramaSeriesId: string = '') => {
    const obj: IRequest = {
      url: `${routerGroupUri}/play/open/resource/${resourceId}`,
      method: 'get',
    }
    if (dramaSeriesId != '') {
      obj.params = {
        dramaSeriesId
      }
    }
    return await request<boolean>(obj);
  },
  playOpenDramaSeries: async (dramaSeriesId: string) => {
    const obj: IRequest = {
      url: `${routerGroupUri}/play/open/dramaSeries/${dramaSeriesId}`,
      method: 'get',
    }
    return await request<boolean>(obj);
  },
  playOpenResourceFolder: async (resourceId: string) => {
    return await request<boolean>({
      url: `${routerGroupUri}/play/open/resource/folder/${resourceId}`,
      method: 'get',
    });
  },
  playUpdate: async (resourceId: string, dramaSeriesId: string) => {
    const obj: IRequest = {
      url: `${routerGroupUri}/play/update/${resourceId}`,
      method: 'get',
    }
    if (dramaSeriesId != '') {
      obj.params = {
        dramaSeriesId
      }
    }
    return await request<boolean>(obj);
  },
  playVideoInfo: async (dramaSeriesId: string) => {
    return await request<I_playVideoInfo>({
      url: `${routerGroupUri}/play/video/info/${dramaSeriesId}`,
      method: 'get',
    });
  },
  getAppConfig: async () => {
    return await request<I_appSystemConfig>({
      url: `${routerGroupUri}/app/getConfig`,
      method: 'get',
    });
  },
  setAppConfig: async (config: I_appSystemConfig) => {
    return await request<boolean>({
      url: `${routerGroupUri}/app/setConfig`,
      method: 'put',
      data: config,
    });
  },
  databaseCleanup: async (form: I_databaseCleanupForm) => {
    return await request<boolean>({
      url: `${routerGroupUri}/database/cleanup`,
      method: 'post',
      data: form,
    });
  },
  dbBackupList: async () => {
    return await request<string[]>({
      url: `${routerGroupUri}/database/dbBackupList`,
      method: 'get',
    });
  },
  deleteDbBackup: async (fileName: string) => {
    return await request<boolean>({
      url: `${routerGroupUri}/database/deleteDbBackup/${fileName}`,
      method: 'delete',
    });
  },
}
