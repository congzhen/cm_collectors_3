import request from "@/assets/request";
import type { I_performerBases } from "@/dataType/performer.dataType";
const routerGroupUri = '';
export const performerBasesServer = {
  create: async (name: string) => {
    return await request<I_performerBases>({
      url: `${routerGroupUri}/performerBases/create`,
      method: 'post',
      data: {
        name,
      },
    });
  },
  update: async (obj: I_performerBases) => {
    return await request<boolean>({
      url: `${routerGroupUri}/performerBases/update`,
      method: 'put',
      data: obj
    });
  }
}
