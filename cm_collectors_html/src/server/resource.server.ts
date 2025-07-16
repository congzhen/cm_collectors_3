import request from "@/assets/request";
import type { I_resource, I_resource_base, I_resourceDramaSeries_base } from "@/dataType/resource.dataType";
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
}
