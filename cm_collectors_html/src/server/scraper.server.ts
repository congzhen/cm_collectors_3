import request from "@/assets/request";
import { defualtConfigScraperData, type I_config_scraperData } from "@/dataType/config.dataType";
import type { I_performer, I_performerBasic } from "@/dataType/performer.dataType";
import type { I_resource } from "@/dataType/resource.dataType";
const routerGroupUri = '';
export const scraperDataServer = {
  configs: async () => {
    return await request<string[]>({
      url: `${routerGroupUri}/scraper/configs`,
      method: 'get',
    });
  },
  updateConfig: async (filesBases_id: string, config: I_config_scraperData | null = null) => {
    if (config == null) {
      config = defualtConfigScraperData;
    }
    const configJson = JSON.stringify(config);
    return await request<boolean>({
      url: `${routerGroupUri}/scraper/updateConfig`,
      method: 'post',
      data: {
        filesBases_id,
        configJson,
      }
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
  },
  searchScraperPerformerData: async (performerBases_id: string, lastScraperUpdateTime: string) => {
    return await request<I_performerBasic[]>({
      url: `${routerGroupUri}/scraper/searchScraperPerformer`,
      method: 'post',
      data: {
        performerBases_id,
        lastScraperUpdateTime,
      }
    });
  },
  scraperPerformerDataProcess: async (performerBases_id: string, performerId: string, performerName: string, scraperConfig: string, operate: string) => {
    return await request<I_performerBasic[]>({
      url: `${routerGroupUri}/scraper/scraperPerformerDataProcess`,
      method: 'post',
      data: {
        performerBases_id,
        performerId,
        performerName,
        scraperConfig,
        operate,
      }
    });
  },
  scraperOneResourceDataProcess: async (resourdId: string, filesBases_id: string, scraperConfig: string, timeout: number, operate: string, saveNfo: boolean, saveImage: boolean, cutPoster: boolean, useExistNfo: boolean, title: string, issueNumber: string, dramaSeriesSrc: string) => {
    return await request<I_resource>({
      url: `${routerGroupUri}/scraper/scraperOneResourceDataProcess`,
      method: 'post',
      data: {
        resourdId,
        filesBases_id,
        scraperConfig,
        timeout,
        operate,
        saveNfo,
        saveImage,
        cutPoster,
        useExistNfo,
        title,
        issueNumber,
        dramaSeriesSrc,
      }
    });
  },
  scraperOnePerformerDataProcess: async (performerId: string, performerName: string, performerBases_id: string, scraperConfig: string, timeout: number, operate: string) => {
    return await request<I_performer>({
      url: `${routerGroupUri}/scraper/scraperOnePerformerDataProcess`,
      method: 'post',
      data: {
        performerId,
        performerName,
        performerBases_id,
        scraperConfig,
        timeout,
        operate,
      }
    });
  }
}
