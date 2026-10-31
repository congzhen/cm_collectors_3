import request from '@/assets/request';
import type { I_resource, I_resourceGroupSummary } from '@/dataType/resource.dataType';

export const resourceGroupServer = {
  summaries: (filesBasesId: string, groupIds: string[] = [], search = '', page = 1) =>
    request<{ dataList: I_resourceGroupSummary[]; total: number }>({
      url: '/resource/groups/summaries', method: 'post', data: { filesBasesId, groupIds, search, page, limit: 20 },
    }),
  info: (filesBasesId: string, groupId: string) => request<I_resource[]>({
    url: `/resource/groups/${encodeURIComponent(filesBasesId)}/${encodeURIComponent(groupId)}`, method: 'get',
  }),
  mutate: (action: 'create' | 'add' | 'remove' | 'reorder' | 'dissolve', filesBasesId: string, resourceIds: string[] = [], groupId = '') =>
    request<{ id: string }>({ url: `/resource/groups/manage/${action}`, method: 'post', data: { filesBasesId, groupId, resourceIds } }),
  setDisplay: (filesBasesId: string, enabled: boolean) => request<boolean>({
    url: '/filesBases/resourceGroupDisplay', method: 'put', data: { filesBasesId, enabled },
  }),
};
