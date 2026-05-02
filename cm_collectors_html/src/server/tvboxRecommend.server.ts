import request from '@/assets/request'
import type { I_tvboxRecommend, I_tvboxRecommendSort } from '@/dataType/tvboxRecommend.dataType'

export const tvboxRecommendServer = {
  list: async () => {
    return await request<I_tvboxRecommend[]>({
      url: '/tvbox/recommend/list',
      method: 'get',
    })
  },
  add: async (resourceId: string) => {
    return await request<boolean>({
      url: `/tvbox/recommend/add/${resourceId}`,
      method: 'post',
    })
  },
  delete: async (id: string) => {
    return await request<boolean>({
      url: `/tvbox/recommend/delete/${id}`,
      method: 'delete',
    })
  },
  updateSort: async (sortItems: I_tvboxRecommendSort[]) => {
    return await request<boolean>({
      url: '/tvbox/recommend/sort',
      method: 'put',
      data: sortItems,
    })
  },
}
