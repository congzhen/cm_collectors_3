import request from "@/assets/request";
import type { I_config_app } from "@/dataType/config.dataType";
import type { I_filesBases, I_filesBases_base } from "@/dataType/filesBases.dataType";
const routerGroupUri = '';
export const filesBasesServer = {
  infoById: async (id: string) => {
    return await request<I_filesBases>({
      url: `${routerGroupUri}/filesBases/info/${id}`,
      method: 'get',
    });
  },
  setData: async (id: string, info: I_filesBases_base, config: I_config_app, mainPerformerBasesId: string, relatedPerformerBases: string[]) => {
    return await request<boolean>({
      url: `${routerGroupUri}/filesBases/setData`,
      method: 'put',
      data: {
        id,
        info,
        config: JSON.stringify(config),
        mainPerformerBasesId,
        relatedPerformerBases
      }
    });
  },
  create: async (name: string, mainPerformerBasesId: string, relatedPerformerBasesIds: string[]) => {
    return await request<I_filesBases>({
      url: `${routerGroupUri}/filesBases/create`,
      method: 'post',
      data: {
        name,
        mainPerformerBasesId,
        relatedPerformerBasesIds,
      },
    });
  },
}
