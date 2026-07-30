import request from '@/assets/request';
import type {
  I_videoTranscodeAddResult,
  I_videoTranscodeConfig,
  I_videoTranscodeCapabilities,
  I_videoTranscodeQueueStatus,
  I_videoTranscodeResetResult,
  I_videoTranscodeTask,
} from '@/dataType/videoTranscode.dataType';

const routerGroupUri = '/videoTranscode';

export const videoTranscodeServer = {
  list: async () => request<I_videoTranscodeTask[]>({
    url: `${routerGroupUri}/list`,
    method: 'get',
  }),
  capabilities: async () => request<I_videoTranscodeCapabilities>({
    url: `${routerGroupUri}/capabilities`,
    method: 'get',
  }),
  add: async (data: {
    resourceIds?: string[];
    dramaSeriesIds?: string[];
    config?: I_videoTranscodeConfig;
  }) => request<I_videoTranscodeAddResult>({
    url: `${routerGroupUri}/add`,
    method: 'post',
    data,
  }),
  updateConfig: async (ids: string[], config: I_videoTranscodeConfig) => request<boolean>({
    url: `${routerGroupUri}/config`,
    method: 'put',
    data: { ids, config },
  }),
  start: async (ids: string[] = []) => request<boolean>({
    url: `${routerGroupUri}/start`,
    method: 'post',
    data: { ids },
  }),
  resetBatch: async (ids: string[]) => request<I_videoTranscodeResetResult>({
    url: `${routerGroupUri}/resetBatch`,
    method: 'post',
    data: { ids },
  }),
  pause: async () => request<boolean>({ url: `${routerGroupUri}/pause`, method: 'post' }),
  resume: async () => request<boolean>({ url: `${routerGroupUri}/resume`, method: 'post' }),
  status: async () => request<I_videoTranscodeQueueStatus>({
    url: `${routerGroupUri}/status`,
    method: 'get',
  }),
  cancel: async (id: string) => request<boolean>({
    url: `${routerGroupUri}/cancel/${id}`,
    method: 'post',
  }),
  delete: async (id: string) => request<boolean>({
    url: `${routerGroupUri}/delete/${id}`,
    method: 'delete',
  }),
  deleteBatch: async (ids: string[]) => request<number>({
    url: `${routerGroupUri}/deleteBatch`,
    method: 'post',
    data: { ids },
  }),
};
