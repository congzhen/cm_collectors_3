import request from "@/assets/request";
import type { I_resource } from "@/dataType/resource.dataType";
const routerGroupUri = '';
export const resourceServer = {
  dataList: async (filesBasesId: string, fetchCount: boolean, page: number, limit: number) => {
    return await request<{ total: number, dataList: I_resource[] }>({
      url: `${routerGroupUri}/resource/dataList`,
      method: 'post',
      data: {
        fetchCount,
        page,
        limit,
        filesBasesId,
      }
    })
  },

}
