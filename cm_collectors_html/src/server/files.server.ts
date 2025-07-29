import request from "@/assets/request";
const routerGroupUri = '';
export const filesServer = {
  filesDListByDramaSeriesId_Image: async (dramaSeriesId: string) => {
    return await request<string[]>({
      url: `${routerGroupUri}/files/list/image/${dramaSeriesId}`,
      method: 'get',
    });
  }
}
