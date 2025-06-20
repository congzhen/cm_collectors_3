import type { I_appData } from "@/dataType/app.dataType";
import { defineStore } from "pinia";
export const appStoreData = defineStore('app', {
  state: () => ({
    appData: {} as I_appData,
  }),
  getters: {

  },
  actions: {
    init: function (obj: I_appData) {
      obj.filesBases.forEach(filesBase => {
        if (filesBase.filesBasesSetting.config_json_data != '') {
          filesBase.filesBasesSetting.config_app = JSON.parse(filesBase.filesBasesSetting.config_json_data);
          filesBase.filesBasesSetting.config_json_data = '';
        }
        if (filesBase.filesBasesSetting.nfo_json_data != '') {
          filesBase.filesBasesSetting.config_nfo = JSON.parse(filesBase.filesBasesSetting.nfo_json_data);
          filesBase.filesBasesSetting.nfo_json_data = '';
        }
        if (filesBase.filesBasesSetting.simple_json_data != '') {
          filesBase.filesBasesSetting.config_simple = JSON.parse(filesBase.filesBasesSetting.simple_json_data);
          filesBase.filesBasesSetting.simple_json_data = '';
        }
      });
      console.log(666, obj);
      this.appData = obj;
    },

  }
})
