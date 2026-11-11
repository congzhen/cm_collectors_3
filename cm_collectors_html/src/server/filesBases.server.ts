import { ElMessageBox } from 'element-plus';
import { sharedConfigServer } from './sharedConfig.server';
import request from "@/assets/request";
import type { E_config_type, I_config_app } from "@/dataType/config.dataType";
import type { I_filesBases, I_filesBases_base, I_filesBases_sort } from "@/dataType/filesBases.dataType";
const routerGroupUri = '';
export const filesBasesServer = {
  infoById: async (id: string) => {
    return await request<I_filesBases>({
      url: `${routerGroupUri}/filesBases/info/${id}`,
      method: 'get',
    });
  },
  setData: async (id: string, info: I_filesBases_base, config: I_config_app, mainPerformerBasesId: string, relatedPerformerBases: string[]) => {
    return await request<boolean>({
      url: `${routerGroupUri}/filesBases/setData`,
      method: 'put',
      data: {
        id,
        info,
        config: JSON.stringify(config),
        mainPerformerBasesId,
        relatedPerformerBases
      }
    });
  },
  create: async (name: string, mainPerformerBasesId: string, relatedPerformerBasesIds: string[], followModules: string[] = []) => {
    return await request<I_filesBases>({
      url: `${routerGroupUri}/filesBases/create`,
      method: 'post',
      data: {
        name,
        mainPerformerBasesId,
        relatedPerformerBasesIds,
        followModules,
      },
    });
  },
  // 真实删除文件库。
  // 后端会检查该文件库是否仍有资源记录，前端调用方只需要传当前选中的文件库 ID。
  delete: async (id: string) => {
    return await request<boolean>({
      url: `${routerGroupUri}/filesBases/delete/${id}`,
      method: 'delete',
    });
  },
  sort: async (sortObj: I_filesBases_sort[]) => {
    return await request<boolean>({
      url: `${routerGroupUri}/filesBases/sort`,
      method: 'put',
      data: {
        sortData: sortObj,
      },
    });
  },
  getConfigById: async (id: string, configType: E_config_type) => {
    return await request<string>({
      url: `${routerGroupUri}/filesBases/config/${id}/${configType}`,
      method: 'get',
    });
  },
  setFilesBasesConfigById: async (id: string, config: I_config_app, options?: { localOnly: boolean; baseline: I_config_app }) => {
    const status = await sharedConfigServer.status(id, 'display');
    if (!status.status) return { status: false, msg: status.msg, data: false };
    const submitted = { ...config } as unknown as Record<string, unknown>;
    const baseline = options?.baseline as unknown as Record<string, unknown> | undefined;
    const changesShared = status.data.fields.some(key =>
      // Legacy shared configurations may omit fields supplied by frontend defaults.
      (baseline || Object.prototype.hasOwnProperty.call(status.data.config, key)) &&
      JSON.stringify(submitted[key]) !== JSON.stringify(baseline ? baseline[key] : status.data.config[key]));
    if (status.data.following && changesShared) {
      if (!options?.localOnly) try {
        await ElMessageBox.confirm('当前基础展示设置跟随公共配置。继续将关闭本库的跟随并保存此次调整，其他库不受影响。', '改为本库自定义');
      } catch { return { status: false, msg: '已取消修改', data: false }; }
      const result = await sharedConfigServer.follow(id, 'display', false, status.data.revision);
      if (!result.status) return result;
    }
    // Independent quick-setting changes keep following and use the latest shared values.
    if (status.data.following && !changesShared) {
      for (const key of status.data.fields) {
        if (Object.prototype.hasOwnProperty.call(status.data.config, key)) submitted[key] = status.data.config[key];
      }
    }
    return await request<boolean>({
      url: `${routerGroupUri}/filesBases/setConfig/filesBases`,
      method: 'put',
      data: {
        id,
        config: JSON.stringify(submitted),
      }
    });
  }
}
