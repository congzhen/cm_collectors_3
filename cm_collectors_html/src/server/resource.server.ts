import request, { type IRequest } from "@/assets/request";
import type { I_dramaSeriesWithResource, I_resource, I_resource_base, I_resourceDramaSeries_base } from "@/dataType/resource.dataType";
import type { I_searchData } from "@/dataType/search.dataType";
const routerGroupUri = '';
export const resourceServer = {
  info: async (id: string) => {
    return await request<I_resource>({
      url: `${routerGroupUri}/resource/info/${id}`,
      method: 'get',
    })
  },
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
  sampleImages: async (id: string, sampleFolder: string) => {
    const obj: IRequest = {
      url: `${routerGroupUri}/resource/sampleImages/${id}`,
      method: 'get',
    }
    sampleFolder = encodeURIComponent(sampleFolder);
    if (sampleFolder != '') {
      obj.params = {
        q: sampleFolder
      }
    }
    return await request<string[]>(obj);
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
  updateTag: async (resourceId: string, tags: string[]) => {
    return await request<I_resource>({
      url: `${routerGroupUri}/resource/update/tag`,
      method: 'put',
      data: {
        resourceId,
        tags
      },
    });
  },
  batchAddTag: async (resourceIds: string[], tags: string[]) => {
    return await request<I_resource[]>({
      url: `${routerGroupUri}/resource/batchAddTag`,
      method: 'put',
      data: {
        resourceIds,
        tags
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

export const resourcesDramaSeriesServer = {
  searchPath: async (filesBasesIds: string[], searchPath: string) => {
    return await request<I_dramaSeriesWithResource[]>({
      url: `${routerGroupUri}/resourceDramaSeries/searchPath`,
      method: 'post',
      data: {
        filesBasesIds,
        searchPath
      }
    })
  },
  replacePath: async (filesBasesIds: string[], searchPath: string, replacePath: string) => {
    return await request<I_dramaSeriesWithResource[]>({
      url: `${routerGroupUri}/resourceDramaSeries/replacePath`,
      method: 'post',
      data: {
        filesBasesIds,
        searchPath,
        replacePath
      }
    })
  },
}
