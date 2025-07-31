
import type { I_performerBases } from "@/dataType/performer.dataType";
import { defineStore } from "pinia";
import { filesBasesStoreData } from "@/storeData/filesBases.storeData";

export const performerBasesStoreData = defineStore('performerBases', {
  state: () => ({
    performerBases: [] as I_performerBases[],
  }),
  getters: {
    activeFirstPerformerBases: function (state) {
      return state.performerBases.find(item => item.status) || null;
    },
    activeFirstPerformerBasesId: function (state) {
      return state.performerBases.find(item => item.status)?.id || '';
    },
  },
  actions: {
    init: function (dataList: I_performerBases[]) {
      this.performerBases = dataList;
    },
    getPerformerBasesById: function (id: string) {
      return this.performerBases.find(item => item.id === id) || null;
    },
    listByIds: function (ids: string[]) {
      return this.performerBases.filter(item => ids.includes(item.id));
    },
    listByFilesBasesId: function (filesBasesId: string) {
      const _filesBasesStore = filesBasesStoreData();
      const ids = _filesBasesStore.getPerformerBasesIdsByFilesBasesId(filesBasesId)
      return this.performerBases.filter(item => ids.includes(item.id));
    },
    update: function (data: I_performerBases) {
      for (let i = 0; i < this.performerBases.length; i++) {
        if (this.performerBases[i].id === data.id) {
          this.performerBases[i] = data;
          break;
        }
      }
    }
  }
})
