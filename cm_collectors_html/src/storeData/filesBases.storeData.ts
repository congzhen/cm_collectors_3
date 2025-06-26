import type { I_filesBases } from "@/dataType/filesBases.dataType";
import { defineStore } from "pinia";
export const filesBasesStoreData = defineStore('filesBases', {
  state: () => ({
    filesBases: [] as I_filesBases[],
  }),
  getters: {
    filesBasesFirst: (state) => {
      if (!state.filesBases || state.filesBases.length == 0) return null;
      return state.filesBases[0];
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
    }
  }
})
