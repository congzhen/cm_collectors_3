import type { I_filesBases } from "@/dataType/filesBases.dataType";
import { defineStore } from "pinia";
import type { I_tagClass, I_tag } from "@/dataType/tag.dataType";
import { defualtConfigApp, defualtConfigNfo, defualtConfigSimple, type I_config_app, type I_config_nfo, type I_config_simple } from "@/dataType/config.dataType";
import { tagServer } from "@/server/tag.server";
import { filesBasesServer } from "@/server/filesBases.server";
import type { I_performer } from "@/dataType/performer.dataType";
import { performerServer } from "@/server/performer.server";
import { appDataServer } from "@/server/app.server";

import { filesBasesStoreData } from '@/storeData/filesBases.storeData'
import { performerBasesStoreData } from '@/storeData/performerBases.storeData';



export const appStoreData = defineStore('app', {
  state: () => ({
    currentFilesBases: {} as I_filesBases,
    currentMainPerformerBasesId: "",
    currentPerformerBasesIds: [] as string[],
    currentConfigApp: {} as I_config_app,
    currentConfigNfo: {} as I_config_nfo,
    currentConfigSimple: {} as I_config_simple,
    currentTag: [] as I_tag[],
    currentTagClass: [] as I_tagClass[],
    currentTopPreferredPerformers: [] as I_performer[],
  }),
  getters: {
    // 获取当前filesBases的app config
    currentFilesBasesAppConfig(state): I_config_app {
      return state.currentConfigApp;
    },
    // 获取当前filesBases的演员显示文字
    currentPerformerText(state): string {
      const text = state.currentConfigApp.performer_Text;
      return text == '' ? 'performer' : text;
    },
    // 获取当前filesBases的cup显示文字
    currentCupText(state): string {
      const text = state.currentConfigApp.plugInUnit_Cup_Text;
      return text == '' ? 'Cup' : text;
    }
  },
  actions: {
    async initApp() {
      const appResult = await appDataServer.init()
      if (!appResult || !appResult.status) {
        return {
          status: false,
          message: appResult.msg
        };
      }
      filesBasesStoreData().init(appResult.data.filesBases)
      performerBasesStoreData().init(appResult.data.performerBases)
      return {
        status: true,
        message: 'success'
      };
    },
    // 初始化
    async initCurrentFilesBases(filesBasesId: string) {
      // 获取filesBases信息
      const info = await filesBasesServer.infoById(filesBasesId);
      if (!info.status) {
        return {
          status: false,
          message: info.msg
        };
      }
      // 设置当前filesBases
      this.currentFilesBases = info.data;

      // 设置currentMainPerformerBasesId 与 currentPerformerBasesIds
      info.data.filesRelatedPerformerBases.forEach((item) => {
        if (item.main) {
          this.currentMainPerformerBasesId = item.performerBases_id;
        }
        this.currentPerformerBasesIds.push(item.performerBases_id);
      });
      if (this.currentMainPerformerBasesId == '') {
        this.currentMainPerformerBasesId = this.currentPerformerBasesIds[0];
      }

      // 解析配置文件
      if (info.data.filesBasesSetting.config_json_data != '') {
        this.currentConfigApp = JSON.parse(info.data.filesBasesSetting.config_json_data);
      } else {
        this.currentConfigApp = defualtConfigApp;
      }
      if (info.data.filesBasesSetting.nfo_json_data != '') {
        this.currentConfigNfo = JSON.parse(info.data.filesBasesSetting.nfo_json_data);
      } else {
        this.currentConfigNfo = defualtConfigNfo;
      }
      if (info.data.filesBasesSetting.simple_json_data != '') {
        this.currentConfigSimple = JSON.parse(info.data.filesBasesSetting.simple_json_data);
      } else {
        this.currentConfigSimple = defualtConfigSimple;
      }

      console.log(this.currentConfigApp);

      this.initTagData(filesBasesId)

      // 获取优先显示演员数据
      const topPreferredPerformersResult = await performerServer.listTopPreferredPerformers(
        this.currentConfigApp.performerPreferred,
        this.currentMainPerformerBasesId,
        this.currentConfigApp.shieldNoPerformerPhoto,
        this.currentConfigApp.performerShowNum
      );
      if (!topPreferredPerformersResult.status) {
        return {
          status: false,
          message: topPreferredPerformersResult.msg
        };
      }
      this.currentTopPreferredPerformers = topPreferredPerformersResult.data;

      return {
        status: true,
        message: 'success'
      };
    },
    async initTagData(filesBasesId: string) {
      // 获取标签数据
      const tagDataResult = await tagServer.tagDataByFilesBasesId(filesBasesId)
      if (!tagDataResult.status) {
        return {
          status: false,
          message: tagDataResult.msg
        };
      }

      // 设置当前标签与标签分类
      this.currentTagClass = tagDataResult.data.tagClass;
      this.currentTag = tagDataResult.data.tag;
    },
    currentTagsByTagClassId(tagClassId: string) {
      return this.currentTag.filter(tag => tag.tagClass_id == tagClassId && tag.status);
    }

  }
})
