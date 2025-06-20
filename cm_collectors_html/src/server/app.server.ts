import request from "@/assets/request";
import type { I_appData } from "@/dataType/app.dataType";
const routerGroupUri = '';
export const appDataServer = {
  init: async () => {
    return await request<I_appData>({
      url: `${routerGroupUri}/app/data`,
      method: 'get',
    });
  },
}
