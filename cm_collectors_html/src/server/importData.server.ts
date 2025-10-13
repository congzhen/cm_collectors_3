import request from "@/assets/request";
import { defualtConfigScanDisk, type I_config_scanDisk } from "@/dataType/config.dataType";
const routerGroupUri = '';
export const importDataServer = {
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
  scanDiskImportData: async (filesBases_id: string, filePath: string, config: I_config_scanDisk) => {
    return await request<boolean>({
      url: `${routerGroupUri}/importData/scanDiskImportData`,
      method: 'post',
      data: {
        filesBases_id,
        filePath,
        config,
      }
    });
  },
  createScanDiskDefaultConfig: async (filesBases_id: string) => {
    const defaultConfigJson = JSON.stringify(defualtConfigScanDisk);
    return await request<I_config_scanDisk>({
      url: `${routerGroupUri}/importData/updateScanDiskConfig`,
      method: 'post',
      data: {
        filesBases_id,
        defaultConfigJson,
      }
    });
  },
}
