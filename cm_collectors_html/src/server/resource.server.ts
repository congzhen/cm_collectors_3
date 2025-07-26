import request from "@/assets/request";
import type { I_resource, I_resource_base, I_resourceDramaSeries_base } from "@/dataType/resource.dataType";
import type { I_searchData } from "@/dataType/search.dataType";
import { de } from "element-plus/es/locales.mjs";
const routerGroupUri = '';
export const resourceServer = {
  dataList: async (filesBasesId: string, fetchCount: boolean, page: number, limit: number, searchData: I_searchData) => {
    return await request<{ total: number, dataList: I_resource[] }>({
      url: `${routerGroupUri}/resource/dataList`,
      method: 'post',
      data: {
        fetchCount,
        page,
        limit,
        filesBasesId,
        searchData,
      }
    })
  },
  create: async (resource: I_resource_base, photoBase64: string, performers: string[], directors: string[], tags: string[], dramaSeries: I_resourceDramaSeries_base[]) => {
    return await request<I_resource>({
      url: `${routerGroupUri}/resource/create`,
      method: 'post',
      data: {
        resource,
        photoBase64,
        performers,
        directors,
        tags,
        dramaSeries,
      },
    });
  },
  update: async (resource: I_resource_base, photoBase64: string, performers: string[], directors: string[], tags: string[], dramaSeries: I_resourceDramaSeries_base[]) => {
    return await request<I_resource>({
      url: `${routerGroupUri}/resource/update`,
      method: 'put',
      data: {
        resource,
        photoBase64,
        performers,
        directors,
        tags,
        dramaSeries,
      },
    });
  },
  delete: async (id: string) => {
    return await request<boolean>({
      url: `${routerGroupUri}/resource/delete/${id}`,
      method: 'delete',
    });
  }
}
