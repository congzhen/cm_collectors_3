import request from "@/assets/request";
import { E_performerCareerType } from "@/dataType/app.dataType";
import type { I_performer, I_performerBasic, I_search_performer } from "@/dataType/performer.dataType";
const routerGroupUri = '';
export const performerServer = {
  dataList: async (performerBasesId: string, fetchCount: boolean, page: number, limit: number, searchCondition: I_search_performer) => {
    return await request<{ total: number, dataList: I_performer[] }>({
      url: `${routerGroupUri}/performer/dataList/${performerBasesId}/${fetchCount}/${page}/${limit}`,
      method: 'get',
      params: searchCondition
    })
  },
  listTopPreferredPerformers: async (preferredIds: string[], mainPerformerBasesId: string, shieldNoPerformerPhoto: boolean, limit: number) => {
    return await request<I_performer[]>({
      url: `${routerGroupUri}/performer/list/top/preferred`,
      method: 'post',
      data: {
        preferredIds,
        mainPerformerBasesId,
        shieldNoPerformerPhoto,
        limit,
      }
    });
  },
  basicList: async (performerBasesIds: string[], careerType: E_performerCareerType) => {
    let careerPerformer = false;
    let careerDirector = false;
    switch (careerType) {
      case E_performerCareerType.All:
        break;
      case E_performerCareerType.Director:
        careerDirector = true;
        break;
      case E_performerCareerType.Performer:
        careerPerformer = true;
        break;
    }
    return await request<I_performerBasic[]>({
      url: `${routerGroupUri}/performer/basicList`,
      method: 'post',
      data: {
        performerBasesIds,
        careerPerformer,
        careerDirector,
      }
    });
  },
  dataListByIds: async (ids: string[]) => {
    return await request<I_performer[]>({
      url: `${routerGroupUri}/performer/dataList/ids/${ids.join(',')}`,
    });
  },
  infoById: async (id: string) => {
    return await request<I_performer>({
      url: `${routerGroupUri}/performer/info/${id}`,
    });
  },

  create: async (performer: I_performer, photoBase64: string) => {
    return await request<I_performer>({
      url: `${routerGroupUri}/performer/create`,
      method: 'post',
      data: {
        performer,
        photoBase64,
      },
    });
  },
  update: async (performer: I_performer, photoBase64: string) => {
    return await request<I_performer>({
      url: `${routerGroupUri}/performer/update`,
      method: 'put',
      data: {
        performer,
        photoBase64,
      },
    });
  },
  updateStatus: async (id: string, status: boolean) => {
    return await request<boolean>({
      url: `${routerGroupUri}/performer/updateStatus`,
      method: 'put',
      data: {
        id,
        status,
      },
    });
  },
  recycleBin: async (performerBasesId: string) => {
    return await request<I_performer[]>({
      url: `${routerGroupUri}/performer/recycleBin/${performerBasesId}`,
    });
  },
  delete: async (id: string) => {
    return await request<boolean>({
      url: `${routerGroupUri}/performer/delete/${id}`,
      method: 'delete',
    });
  },
  migrate: async (performerId: string, performerBasesId: string) => {
    return await request<boolean>({
      url: `${routerGroupUri}/performer/migrate`,
      method: 'post',
      data: {
        performerId,
        performerBasesId,
      },
    });
  }
}
