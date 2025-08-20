import type { I_filesBases } from "@/dataType/filesBases.dataType";
import { defineStore } from "pinia";
import type { I_tagClass, I_tag } from "@/dataType/tag.dataType";
import { defualtConfigApp, type I_config_app } from "@/dataType/config.dataType";
import { tagServer } from "@/server/tag.server";
import { filesBasesServer } from "@/server/filesBases.server";
import type { I_performer } from "@/dataType/performer.dataType";
import { performerServer } from "@/server/performer.server";
import { appDataServer } from "@/server/app.server";

import { filesBasesStoreData } from '@/storeData/filesBases.storeData'
import { performerBasesStoreData } from '@/storeData/performerBases.storeData';
import type { I_appConfig } from "@/dataType/app.dataType";


export const appStoreData = defineStore('app', {
  state: () => ({
    adminResourceStatus: false,
    adminLoginStatus: false, // 管理员是否已登录
    appConfig: {} as I_appConfig,
    currentFilesBases: {} as I_filesBases,
    currentMainPerformerBasesId: "",
    currentPerformerBasesIds: [] as string[],
    currentConfigApp: {} as I_config_app,
    //currentConfigNfo: {} as I_config_nfo,
    //currentConfigSimple: {} as I_config_simple,
    currentTag: [] as I_tag[],
    currentTagClass: [] as I_tagClass[],
    currentTopPreferredPerformers: [] as I_performer[],
  }),
  getters: {
    getLogoName(state): string {
      return !state.appConfig || !state.appConfig.logoName || state.appConfig.logoName == '' ? 'CM File Collectors' : state.appConfig.logoName;
    },
    // 是否需要管理员登录
    isAdminLogin(state): boolean {
      return state.appConfig && state.appConfig.isAdminLogin;
    },
    // 是否需要管理员登录 且  管理员以登陆
    isAdminLoginStatus(state): boolean {
      return this.isAdminLogin && state.adminLoginStatus
    },
    // 详情页是否显示
    detailsViewStatus(state): boolean {
      return state.adminResourceStatus ? false : true;
    },
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
      this.appConfig = appResult.data.appConfig;
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
        const configApp = JSON.parse(info.data.filesBasesSetting.config_json_data);
        this.currentConfigApp = { ...defualtConfigApp, ...configApp };
      } else {
        this.currentConfigApp = defualtConfigApp;
      }
      /*
      if (info.data.filesBasesSetting.nfo_json_data != '') {
        const configNfo = JSON.parse(info.data.filesBasesSetting.nfo_json_data);
        this.currentConfigNfo = { ...defualtConfigApp, ...configNfo };
      } else {
        this.currentConfigNfo = defualtConfigNfo;
      }
      if (info.data.filesBasesSetting.simple_json_data != '') {
        const configSimple = JSON.parse(info.data.filesBasesSetting.simple_json_data);
        this.currentConfigNfo = { ...defualtConfigApp, ...configSimple };
      } else {
        this.currentConfigSimple = defualtConfigSimple;
      }
      */
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
    // 根据标签分类获取当前标签列表
    currentTagsByTagClassId(tagClassId: string) {
      return this.currentTag.filter(tag => tag.tagClass_id == tagClassId && tag.status);
    },
    //根据标签id获取标签信息
    currentTagInfoById(tagId: string) {
      return this.currentTag.find(tag => tag.id == tagId);
    },
    //获取cup文字
    cupText(cup: string, separator = ' ') {
      return `${cup}${separator}${this.currentCupText}`;
    },
  }
})
