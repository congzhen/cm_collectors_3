import request from '@/assets/request';
import type { I_performerTag, I_performerTagClass, I_performerTagData } from '@/dataType/performerTag.dataType';

export const performerTagServer = {
  data: async (performerBasesId: string, includeDisabled = false) => request<I_performerTagData>({
    url: `/performerTag/data/${performerBasesId}`,
    params: { includeDisabled },
  }),
  createClass: async (data: Omit<I_performerTagClass, 'id'>) => request<I_performerTagClass>({
    url: '/performerTagClass/create', method: 'post', data,
  }),
  updateClass: async (data: I_performerTagClass) => request<boolean>({
    url: '/performerTagClass/update', method: 'put', data,
  }),
  deleteClass: async (id: string) => request<boolean>({
    url: `/performerTagClass/delete/${id}`, method: 'delete',
  }),
  createTag: async (data: Omit<I_performerTag, 'id' | 'performerCount'>) => request<I_performerTag>({
    url: '/performerTag/create', method: 'post', data,
  }),
  updateTag: async (data: I_performerTag) => request<boolean>({
    url: '/performerTag/update', method: 'put', data,
  }),
  deleteTag: async (id: string) => request<boolean>({
    url: `/performerTag/delete/${id}`, method: 'delete',
  }),
  updateSort: async (tagClassSort: Array<{ id: string; sort: number }>, tagSort: Array<{ id: string; sort: number }>) => request<boolean>({
    url: '/performerTag/update/sort', method: 'put', data: { tagClassSort, tagSort },
  }),
};
