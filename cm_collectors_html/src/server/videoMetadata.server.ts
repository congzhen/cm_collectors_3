import request from '@/assets/request';
import type {
  I_videoMetadata,
  I_videoMetadataBatchTask,
  I_videoMetadataFailureList,
  I_videoMetadataFailureQuery,
  I_videoMetadataManualRequest,
  I_videoMetadataRunRequest,
  I_videoMetadataSettingData,
  I_videoMetadataStats,
} from '@/dataType/videoMetadata.dataType';

const routerGroupUri = '/videoMetadata';

export const videoMetadataServer = {
  bulk: (ids: string[], action: 'retry' | 'nonvideo' | 'video') => request<{ succeeded: number; failed: Record<string, string> }>({ url: `${routerGroupUri}/bulk`, method: 'post', data: { ids, action } }),
  cleanupPreview: () => request<{ total: number; byExtension: Record<string, number>; token: string }>({ url: `${routerGroupUri}/cleanupPreview`, method: 'get' }),
  startCleanup: (token: string) => request<boolean>({ url: `${routerGroupUri}/cleanup`, method: 'post', data: { token } }),
  cleanupStatus: () => request<{ status: string; processed: number; total: number; error: string }>({ url: `${routerGroupUri}/cleanupStatus`, method: 'get' }),
  setting: async () => request<I_videoMetadataSettingData>({
    url: `${routerGroupUri}/setting`,
    method: 'get',
  }),
  saveSetting: async (data: I_videoMetadataSettingData) => request<I_videoMetadataSettingData>({
    url: `${routerGroupUri}/setting`,
    method: 'put',
    data,
  }),
  stats: async () => request<I_videoMetadataStats[]>({
    url: `${routerGroupUri}/stats`,
    method: 'get',
  }),
  info: async (dramaSeriesId: string) => request<I_videoMetadata>({
    url: `${routerGroupUri}/info/${dramaSeriesId}`,
    method: 'get',
  }),
  failures: async (data: I_videoMetadataFailureQuery) => request<I_videoMetadataFailureList>({
    url: `${routerGroupUri}/failures`,
    method: 'post',
    data,
  }),
  retryFailure: async (dramaSeriesId: string) => request<boolean>({
    url: `${routerGroupUri}/retry/${dramaSeriesId}`,
    method: 'post',
  }),
  setClassification: async (dramaSeriesId: string, isVideo: boolean) => request<boolean>({
    url: `${routerGroupUri}/classification`,
    method: 'put',
    data: { dramaSeriesId, isVideo },
  }),
  saveManual: async (data: I_videoMetadataManualRequest) => request<I_videoMetadata>({
    url: `${routerGroupUri}/manual`,
    method: 'put',
    data,
  }),
  run: async (data: I_videoMetadataRunRequest) => request<I_videoMetadataBatchTask>({
    url: `${routerGroupUri}/run`,
    method: 'post',
    data,
  }),
  taskStatus: async () => request<I_videoMetadataBatchTask>({
    url: `${routerGroupUri}/taskStatus`,
    method: 'get',
  }),
  pause: async () => request<boolean>({ url: `${routerGroupUri}/pause`, method: 'post' }),
  resume: async () => request<boolean>({ url: `${routerGroupUri}/resume`, method: 'post' }),
  stop: async () => request<boolean>({ url: `${routerGroupUri}/stop`, method: 'post' }),
};
