import request from '@/assets/request'
import type { I_resource } from '@/dataType/resource.dataType'
import type { I_searchData } from '@/dataType/search.dataType'
import type { I_config_app } from '@/dataType/config.dataType'
export type CategoryBrowseSettings = Pick<
  I_config_app,
  | 'categoryBrowseEnabled'
  | 'categoryBrowseSource'
  | 'categoryBrowseSort'
  | 'categoryBrowseGroupLimit'
  | 'categoryBrowseCoverStyle'
  | 'categoryBrowseExpandStyle'
>
export interface CategoryGroup {
  categoryId: string
  name: string
  photo: string
  part: number
  parts: number
  count: number
  total: number
  preview: I_resource[]
}
export interface CategoryQuery {
  filesBasesId: string
  source: string
  groupLimit: number
  categorySort?: CategoryBrowseSettings['categoryBrowseSort']
  searchData: I_searchData
  page: number
  limit: number
  categoryId?: string
  part?: number
}
export const categoryBrowseServer = {
  summaries: (data: CategoryQuery) =>
    request<{
      dataList: CategoryGroup[]
      total: number
      resourceTotal: number
      categoryTotal: number
    }>({ url: '/resource/categoryBrowse/summaries', method: 'post', data }),
  members: (data: CategoryQuery) =>
    request<{ dataList: I_resource[]; total: number }>({
      url: '/resource/categoryBrowse/members',
      method: 'post',
      data,
    }),
  settings: (filesBasesId: string, data: CategoryBrowseSettings) =>
    request<boolean>({
      url: '/filesBases/categoryBrowse',
      method: 'put',
      data: { filesBasesId, ...data },
    }),
}
export const categorySettings = (config: I_config_app): CategoryBrowseSettings => ({
  categoryBrowseEnabled: config.categoryBrowseEnabled === true,
  categoryBrowseSource: config.categoryBrowseSource || 'performer',
  categoryBrowseSort: config.categoryBrowseSort || 'count',
  categoryBrowseGroupLimit: config.categoryBrowseGroupLimit ?? 12,
  categoryBrowseCoverStyle: config.categoryBrowseCoverStyle || 'stack',
  categoryBrowseExpandStyle: config.categoryBrowseExpandStyle || 'fan',
})
