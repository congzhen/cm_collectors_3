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
