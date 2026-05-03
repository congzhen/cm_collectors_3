import request from "@/assets/request";
import type {
  I_DuplicateResult,
  I_VideoFingerprintStats,
  I_VideoFingerprintTaskStatus,
} from "@/dataType/videoFingerprint.dataType";

const routerGroupUri = '/videoFingerprint';

export const videoFingerprintServer = {
  stats: async (files_bases_id = '') => {
    return await request<I_VideoFingerprintStats>({
      url: `${routerGroupUri}/stats`,
      method: 'get',
      params: { files_bases_id },
    });
  },
  queryDuplicates: async (params: {
    files_bases_id?: string;
    match_mode?: string;
    threshold?: number;
    duration_first?: boolean;
    duration_tolerance?: number;
    page?: number;
    limit?: number;
  }) => {
    return await request<I_DuplicateResult>({
      url: `${routerGroupUri}/queryDuplicates`,
      method: 'get',
      params,
    });
  },
  triggerCompute: async (batch_size = 50, files_bases_id = '') => {
    return await request<boolean>({
      url: `${routerGroupUri}/triggerCompute`,
      method: 'get',
      params: { batch_size, files_bases_id },
    });
  },
  reScan: async (files_bases_id = '') => {
    return await request<{ added: number }>({
      url: `${routerGroupUri}/reScan`,
      method: 'get',
      params: { files_bases_id },
    });
  },
  taskStatus: async () => {
    return await request<I_VideoFingerprintTaskStatus>({
      url: `${routerGroupUri}/taskStatus`,
      method: 'get',
    });
  },
  resetFailed: async (files_bases_id = '') => {
    return await request<{ reset: number }>({
      url: `${routerGroupUri}/resetFailed`,
      method: 'put',
      params: { files_bases_id },
    });
  },
  deleteDramaSeries: async (drama_series_ids: string[], delete_file: boolean) => {
    return await request<boolean>({
      url: `${routerGroupUri}/deleteDramaSeries`,
      method: 'post',
      data: { drama_series_ids, delete_file },
    });
  },
  resetAll: async () => {
    return await request<boolean>({
      url: `${routerGroupUri}/resetAll`,
      method: 'delete',
    });
  },
};
