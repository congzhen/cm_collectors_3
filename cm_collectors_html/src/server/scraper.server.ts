import request from "@/assets/request";
import type { I_config_scraperData } from "@/dataType/config.dataType";
const routerGroupUri = '';
export const scraperDataServer = {
  configs: async () => {
    return await request<string[]>({
      url: `${routerGroupUri}/scraper/configs`,
      method: 'get',
    });
  },
  pretreatment: async (filesBases_id: string, config: I_config_scraperData) => {
    return await request<string[]>({
      url: `${routerGroupUri}/scraper/pretreatment`,
      method: 'post',
      data: {
        filesBases_id,
        config,
      }
    });
  },
  scraperDataProcess: async (filesBases_id: string, filePath: string, config: I_config_scraperData) => {
    return await request<boolean>({
      url: `${routerGroupUri}/scraper/scraperDataProcess`,
      method: 'post',
      data: {
        filesBases_id,
        filePath,
        config,
      }
    });
  }
}
