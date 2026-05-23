import request from "@/assets/request";
import type { I_performerBases } from "@/dataType/performer.dataType";
const routerGroupUri = '';
export const performerBasesServer = {
  create: async (name: string) => {
    return await request<I_performerBases>({
      url: `${routerGroupUri}/performerBases/create`,
      method: 'post',
      data: {
        name,
      },
    });
  },
  update: async (obj: I_performerBases) => {
    return await request<boolean>({
      url: `${routerGroupUri}/performerBases/update`,
      method: 'put',
      data: obj
    });
  },
  // 真实删除演员库。
  // 后端会检查该演员库是否还有演员，以及是否仍被文件库关联。
  delete: async (id: string) => {
    return await request<boolean>({
      url: `${routerGroupUri}/performerBases/delete/${id}`,
      method: 'delete',
    });
  },
  export: async (id: string) => {
    return await request<string>({
      url: `${routerGroupUri}/performerBases/export/${id}`,
      method: 'get',
    });
  },
  import: async (performerDatabaseId: string, fileName: string, content: string, reconstructId: boolean) => {
    return await request<boolean>({
      url: `${routerGroupUri}/performerBases/import`,
      method: 'post',
      data: {
        performerDatabaseId,
        fileName,
        content,
        reconstructId,
      }
    });
  },
}
