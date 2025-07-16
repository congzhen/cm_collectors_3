import request from "@/assets/request";
import type { I_tag, I_tagClass, I_tagData, I_tagSort } from "@/dataType/tag.dataType";
const routerGroupUri = '';
export const tagServer = {
  tagDataByFilesBasesId: async (filesBases_id: string) => {
    return await request<I_tagData>({
      url: `${routerGroupUri}/tag/data/${filesBases_id}`,
      method: 'get',
    });
  },
  tagListByFilesBasesId: async (filesBases_id: string) => {
    return await request<I_tag[]>({
      url: `${routerGroupUri}/tag/list/filesBasesId/${filesBases_id}`,
      method: 'get',
    });
  },
  tagListByTagClassId: async (tagClassId: string) => {
    return await request<I_tag[]>({
      url: `${routerGroupUri}/tag/list/tagClassId/${tagClassId}`,
      method: 'get',
    });
  },
  tagClassListByFilesBasesId: async (filesBases_id: string) => {
    return await request<I_tagClass[]>({
      url: `${routerGroupUri}/tagClass/list/${filesBases_id}`,
      method: 'get',
    });
  },
  createTag: async (tag: I_tag) => {
    return await request<I_tag>({
      url: `${routerGroupUri}/tag/create`,
      method: 'post',
      data: tag,
    });
  },
  updateTag: async (tag: I_tag) => {
    return await request<I_tag>({
      url: `${routerGroupUri}/tag/update`,
      method: 'put',
      data: tag,
    });
  },
  createTagClass: async (tagClass: I_tagClass) => {
    return await request<boolean>({
      url: `${routerGroupUri}/tagClass/create`,
      method: 'post',
      data: tagClass,
    });
  },
  updateTagClass: async (tagClass: I_tagClass) => {
    return await request<boolean>({
      url: `${routerGroupUri}/tagClass/update`,
      method: 'put',
      data: tagClass,
    });
  },
  setTagDataSort: async (tagClassSort: I_tagSort[], tagSort: I_tagSort[]) => {
    return await request<boolean>({
      url: `${routerGroupUri}/tag/update/sort/`,
      method: 'put',
      data: {
        tagClassSort,
        tagSort,
      }
    });
  }
}
