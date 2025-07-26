import type { I_filesBases } from "@/dataType/filesBases.dataType";
import { defineStore } from "pinia";
export const filesBasesStoreData = defineStore('filesBases', {
  state: () => ({
    filesBases: [] as I_filesBases[],
  }),
  getters: {
    filesBasesStatus: (state) => {
      return state.filesBases.filter(item => item.status)
    },
    filesBasesFirst: (state) => {
      if (!state.filesBases || state.filesBases.filter(item => item.status).length == 0) return null;
      return state.filesBases.filter(item => item.status)[0];
    },
  },
  actions: {
    init: function (dataList: I_filesBases[]) {
      this.filesBases = dataList;
    },
    getPerformerRelatedFilesBases: function (performerBasesId: string): I_filesBases[] {
      return this.filesBases.filter(item => item.filesRelatedPerformerBases.find(item => item.performerBases_id === performerBasesId))
    },
    getPerformerRelatedFilesBasesIds: function (performerBasesId: string): string[] {
      return this.getPerformerRelatedFilesBases(performerBasesId).map(item => item.id)
    },
    // 获取文件集关联的主演员集id
    getMainPerformerBasesIdByFilesBasesId: function (filesBasesId: string): string {
      return this.filesBases.find(item => item.id === filesBasesId)?.filesRelatedPerformerBases.find(item => item.main)?.performerBases_id || ''
    },
    // 获取文件集关联的演员集id数组
    getPerformerBasesIdsByFilesBasesId: function (filesBasesId: string): string[] {
      return this.filesBases.find(item => item.id === filesBasesId)?.filesRelatedPerformerBases.map(item => item.performerBases_id) || []
    },
  }
})
