import request from "@/assets/request";
import type { E_config_type, I_config_app } from "@/dataType/config.dataType";
import type { I_filesBases, I_filesBases_base, I_filesBases_sort } from "@/dataType/filesBases.dataType";
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
  // 真实删除文件库。
  // 后端会检查该文件库是否仍有资源记录，前端调用方只需要传当前选中的文件库 ID。
  delete: async (id: string) => {
    return await request<boolean>({
      url: `${routerGroupUri}/filesBases/delete/${id}`,
      method: 'delete',
    });
  },
  sort: async (sortObj: I_filesBases_sort[]) => {
    return await request<boolean>({
      url: `${routerGroupUri}/filesBases/sort`,
      method: 'put',
      data: {
        sortData: sortObj,
      },
    });
  },
  getConfigById: async (id: string, configType: E_config_type) => {
    return await request<string>({
      url: `${routerGroupUri}/filesBases/config/${id}/${configType}`,
      method: 'get',
    });
  },
  setFilesBasesConfigById: async (id: string, config: I_config_app) => {
    return await request<boolean>({
      url: `${routerGroupUri}/filesBases/setConfig/filesBases`,
      method: 'put',
      data: {
        id,
        config: JSON.stringify(config),
      }
    });
  }
}
