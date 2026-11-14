import request from "@/assets/request";
import { defualtConfigScanDisk, type I_config_scanDisk } from "@/dataType/config.dataType";
const routerGroupUri = '';
export interface ImportResult { warnings: string[] }
export const importDataServer = {
  batch: (action: 'begin' | 'heartbeat' | 'end', filesBases_id: string, batchId = '', signal?: AbortSignal) =>
    request<string | boolean>({ url: '/importData/batch', method: 'post', timeout: 10000, signal, data: { action, filesBases_id, batchId } }),
  scanDiskImportPaths: async (filesBases_id: string, config: I_config_scanDisk) => {
    return await request<string[]>({
      url: `${routerGroupUri}/importData/scanDiskImportPaths`,
      method: 'post',
      data: {
        filesBases_id,
        config,
      }
    });
  },
  scanDiskImportData: async (filesBases_id: string, filePath: string, config: I_config_scanDisk, batchId = '', signal?: AbortSignal) => {
    return await request<ImportResult | boolean>({
      signal, timeout: 90000,
      url: `${routerGroupUri}/importData/scanDiskImportData`,
      method: 'post',
      data: {
        filesBases_id,
        filePath,
        config,
        batchId,
      }
    });
  },
  updateScanDiskConfig: async (filesBases_id: string, config: I_config_scanDisk | null = null) => {
    if (config == null) {
      config = defualtConfigScanDisk;
    }
    const configJson = JSON.stringify(config);
    return await request<boolean>({
      url: `${routerGroupUri}/importData/updateScanDiskConfig`,
      method: 'post',
      data: {
        filesBases_id,
        configJson,
      }
    });
  },
}
