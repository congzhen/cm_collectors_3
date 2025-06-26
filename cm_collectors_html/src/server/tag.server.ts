import request from "@/assets/request";
import type { I_tagData } from "@/dataType/tag.dataType";
const routerGroupUri = '';
export const tagServer = {
  tagDataByFilesBasesId: async (filesBases_id: string) => {
    return await request<I_tagData>({
      url: `${routerGroupUri}/tag/data/${filesBases_id}`,
      method: 'get',
    });
  },
}
