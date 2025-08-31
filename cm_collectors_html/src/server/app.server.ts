import request, { type IRequest } from "@/assets/request";
import type { I_appSystemConfig, I_appData } from "@/dataType/app.dataType";
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
}
