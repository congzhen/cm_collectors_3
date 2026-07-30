import request from '@/assets/request';
import type {
  I_searchFavorite,
  I_searchFavoriteCreate,
  I_searchFavoriteUpdate,
} from '@/dataType/searchFavorite.dataType';

export const searchFavoriteServer = {
  list: async (filesBasesId: string) => {
    return await request<I_searchFavorite[]>({
      url: `/searchFavorites/${filesBasesId}`,
      method: 'get',
    });
  },
  create: async (data: I_searchFavoriteCreate) => {
    return await request<I_searchFavorite>({
      url: '/searchFavorites',
      method: 'post',
      data,
    });
  },
  update: async (id: string, data: I_searchFavoriteUpdate) => {
    return await request<I_searchFavorite>({
      url: `/searchFavorites/${id}`,
      method: 'put',
      data,
    });
  },
  delete: async (id: string) => {
    return await request<boolean>({
      url: `/searchFavorites/${id}`,
      method: 'delete',
    });
  },
};
