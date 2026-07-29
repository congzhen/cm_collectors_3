import request from '@/assets/request';
import type {
  I_videoMetadata,
  I_videoMetadataBatchTask,
  I_videoMetadataRunRequest,
  I_videoMetadataSettingData,
  I_videoMetadataStats,
} from '@/dataType/videoMetadata.dataType';

const routerGroupUri = '/videoMetadata';

export const videoMetadataServer = {
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
