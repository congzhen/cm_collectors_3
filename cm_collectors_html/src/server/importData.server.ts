import request from "@/assets/request";
import type { I_config_scanDisk } from "@/dataType/config.dataType";
const routerGroupUri = '';
export const importDataServer = {
  scanDisk: async (filesBasesId: string, config: I_config_scanDisk) => {
    return await request({
      url: `${routerGroupUri}/importData/scanDisk`,
      method: 'post',
      data: {
        filesBasesId,
        config,
      }
    });
  },
}
