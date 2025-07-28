import { playResource } from "@/common/play";
import request, { type IRequest } from "@/assets/request";
import type { I_appData } from "@/dataType/app.dataType";
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
  playOpenResourceFolder: async (resourceId: string) => {
    return await request<boolean>({
      url: `${routerGroupUri}/play/open/resource/folder/${resourceId}`,
      method: 'get',
    });
  },
}
