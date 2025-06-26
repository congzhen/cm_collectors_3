import request from "@/assets/request";
import type { I_filesBases } from "@/dataType/filesBases.dataType";
const routerGroupUri = '';
export const filesBasesServer = {
  infoById: async (id: string) => {
    return await request<I_filesBases>({
      url: `${routerGroupUri}/filesBases/info/${id}`,
      method: 'get',
    });
  },
}
