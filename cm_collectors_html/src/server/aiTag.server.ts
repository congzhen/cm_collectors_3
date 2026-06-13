import request from "@/assets/request";
import type {
  I_aiTagFilesBasesSetting,
  I_aiTagModelTestResult,
  I_aiTagRecordList,
  I_aiTagResetResult,
  I_aiTagRunResult,
  I_aiTagSetting,
  I_aiTagStats
} from "@/dataType/aiTag.dataType";

const routerGroupUri = '/aiTag';

export const aiTagServer = {
  setting: async () => {
    return await request<I_aiTagSetting>({
      url: `${routerGroupUri}/setting`,
      method: 'get',
    });
  },
  saveSetting: async (setting: I_aiTagSetting) => {
    return await request<I_aiTagSetting>({
      url: `${routerGroupUri}/setting`,
      method: 'put',
      data: setting,
    });
  },
  filesBases: async () => {
    return await request<I_aiTagFilesBasesSetting[]>({
      url: `${routerGroupUri}/filesBases`,
      method: 'get',
    });
  },
  saveFilesBases: async (items: I_aiTagFilesBasesSetting[]) => {
    return await request<boolean>({
      url: `${routerGroupUri}/filesBases`,
      method: 'put',
      data: { items },
    });
  },
  stats: async (filesBasesId = '') => {
    return await request<I_aiTagStats>({
      url: `${routerGroupUri}/stats`,
      method: 'get',
      params: { files_bases_id: filesBasesId },
    });
  },
  records: async (filesBasesId = '', status = '', page = 1, limit = 20) => {
    return await request<I_aiTagRecordList>({
      url: `${routerGroupUri}/records`,
      method: 'get',
      params: { files_bases_id: filesBasesId, status, page, limit },
    });
  },
  runOnce: async (filesBasesId = '') => {
    return await request<I_aiTagRunResult>({
      url: `${routerGroupUri}/runOnce`,
      method: 'post',
      params: { files_bases_id: filesBasesId },
    });
  },
  resetFailed: async (filesBasesId = '') => {
    return await request<I_aiTagResetResult>({
      url: `${routerGroupUri}/resetFailed`,
      method: 'post',
      params: { files_bases_id: filesBasesId },
    });
  },
  resetProcessing: async (filesBasesId = '') => {
    return await request<I_aiTagResetResult>({
      url: `${routerGroupUri}/resetProcessing`,
      method: 'post',
      params: { files_bases_id: filesBasesId },
    });
  },
  pause: async () => {
    return await request<I_aiTagSetting>({
      url: `${routerGroupUri}/pause`,
      method: 'post',
    });
  },
  resume: async () => {
    return await request<I_aiTagSetting>({
      url: `${routerGroupUri}/resume`,
      method: 'post',
    });
  },
  rescan: async (filesBasesId = '') => {
    return await request<boolean>({
      url: `${routerGroupUri}/rescan`,
      method: 'post',
      params: { files_bases_id: filesBasesId },
    });
  },
  testConnection: async (setting: I_aiTagSetting) => {
    return await request<I_aiTagModelTestResult>({
      url: `${routerGroupUri}/testConnection`,
      method: 'post',
      data: setting,
    });
  },
  testService: async () => {
    return await request<boolean>({
      url: `${routerGroupUri}/testService`,
      method: 'post',
    });
  },
}
