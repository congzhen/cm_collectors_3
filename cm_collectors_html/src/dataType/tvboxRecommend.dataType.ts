import type { I_resource_base } from './resource.dataType'

export interface I_tvboxRecommend {
  id: string
  resourceId: string
  sort: number
  resource: I_resource_base
}

export interface I_tvboxRecommendSort {
  id: string
  sort: number
}
