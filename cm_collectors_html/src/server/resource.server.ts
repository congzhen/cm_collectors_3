import request, { type IRequest } from "@/assets/request";
import type { I_dramaSeriesWithResource, I_resource, I_resource_base, I_resourceDramaSeries, I_resourceDramaSeries_base } from "@/dataType/resource.dataType";
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
  // 获取随便看看
  dataListCasualView: async (filesBasesId: string, quantity: number) => {
    return await request<I_resource[]>({
      url: `${routerGroupUri}/resource/dataListCasualView/${filesBasesId}/${quantity}`,
      method: 'get',
    })
  },
  //获取历史记录
  dataListHistory: async (filesBasesId: string, quantity: number) => {
    return await request<I_resource[]>({
      url: `${routerGroupUri}/resource/dataListHistory/${filesBasesId}/${quantity}`,
      method: 'get',
    })
  },
  //获取热门资源
  dataListHot: async (filesBasesId: string, quantity: number) => {
    return await request<I_resource[]>({
      url: `${routerGroupUri}/resource/dataListHot/${filesBasesId}/${quantity}`,
      method: 'get',
    })
  },
  dataListByIds: async (ids: string[]) => {
    return await request<I_resource[]>({
      url: `${routerGroupUri}/resource/list/ids`,
      method: 'post',
      data: {
        ids,
      }
    })
  },
  deletedDataList: async (filesBasesIds: string[]) => {
    return await request<I_resource[]>({
      url: `${routerGroupUri}/resource/deleted/list?filesBasesIds=${filesBasesIds.join(',')}`,
    })
  },
  dataCountByPerformerId: async (filesBasesId: string, performerId: string) => {
    return await request<number>({
      url: `${routerGroupUri}/resource/count/${filesBasesId}/${performerId}`,
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
  updatePerformer: async (resourceId: string, performers: string[]) => {
    return await request<I_resource>({
      url: `${routerGroupUri}/resource/update/performer`,
      method: 'put',
      data: {
        resourceId,
        performers
      },
    });
  },
  batchSetPerformer: async (resourceIds: string[], performersIds: string[]) => {
    return await request<I_resource[]>({
      url: `${routerGroupUri}/resource/batchSetPerformer`,
      method: 'post',
      data: {
        resourceIds,
        performersIds
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
  batchSetTag: async (mode: 'add' | 'remove', resourceIds: string[], tags: string[]) => {
    return await request<I_resource[]>({
      url: `${routerGroupUri}/resource/batchSetTag`,
      method: 'put',
      data: {
        mode,
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
  infoById: async (id: string) => {
    return await request<I_resourceDramaSeries>({
      url: `${routerGroupUri}/resourceDramaSeries/info/${id}`,
      method: 'get',
    })
  },
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
