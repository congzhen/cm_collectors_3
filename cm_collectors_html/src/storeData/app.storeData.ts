import type { I_filesBases } from "@/dataType/filesBases.dataType";
import { defineStore } from "pinia";
import type { I_tagClass, I_tag } from "@/dataType/tag.dataType";
import { createDefaultConfigApp, type I_config_app } from "@/dataType/config.dataType";
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
    runtimeBridgeStatus: false, // runtimeBridge是否可用
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
    tagDataRequestSerial: 0,
    filesBasesRequestSerial: 0,
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
    // 是否显示管理员功能
    displayAdminFn(): boolean {
      if (!this.isAdminLogin || this.isAdminLoginStatus) {
        return true;
      }
      return false;
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
    // 所有依赖准备完成后再提交，失败时保留当前库；只接受最后一次初始化。
    async initCurrentFilesBases(filesBasesId: string): Promise<{ status: boolean; message: string; stale?: boolean }> {
      const requestSerial = ++this.filesBasesRequestSerial;
      const isStale = () => requestSerial !== this.filesBasesRequestSerial;
      const staleResult = { status: true, message: 'stale files base request ignored', stale: true };
      try {
        const info = await filesBasesServer.infoById(filesBasesId);
        if (isStale()) return staleResult;
        if (!info?.status) return { status: false, message: info?.msg || '文件库加载失败' };

        const library = info.data;
        const performerBasesIds = library.filesRelatedPerformerBases.map(item => item.performerBases_id);
        const mainPerformerBasesId = [...library.filesRelatedPerformerBases].reverse().find(item => item.main)?.performerBases_id
          || performerBasesIds[0] || '';
        const config = {
          ...createDefaultConfigApp(),
          ...(library.filesBasesSetting.config_json_data ? JSON.parse(library.filesBasesSetting.config_json_data) : {}),
        } as I_config_app;
        const [tagResult, performersResult] = await Promise.all([
          tagServer.tagDataByFilesBasesId(filesBasesId, config.showCustomTagResourceCount !== false),
          performerServer.listTopPreferredPerformers(
            config.performerPreferredEnabled !== false ? config.performerPreferred : [],
            mainPerformerBasesId, config.shieldNoPerformerPhoto, config.performerShowNum,
            filesBasesId, config.performerSortMode, config.performerRecentDays,
          ),
        ]);
        if (isStale()) return staleResult;
        if (!tagResult?.status) return { status: false, message: tagResult?.msg || '标签加载失败' };
        if (!performersResult?.status) return { status: false, message: performersResult?.msg || '演员加载失败' };

        // 使提交前发出的标签刷新失效，避免同库重载时旧请求覆盖新配置的结果。
        // 当前库最后更新，让监听库 ID 的同步逻辑读取到完整的新配置。
        // 使用函数式 patch 整体替换对象，避免递归合并时库 ID 先变而名称、关联仍是旧值。
        this.$patch(state => Object.assign(state, {
          tagDataRequestSerial: this.tagDataRequestSerial + 1,
          currentPerformerBasesIds: performerBasesIds,
          currentMainPerformerBasesId: mainPerformerBasesId,
          currentConfigApp: config,
          currentTagClass: tagResult.data.tagClass,
          currentTag: tagResult.data.tag,
          currentTopPreferredPerformers: performersResult.data,
          currentFilesBases: library,
        }));
        return { status: true, message: 'success' };
      } catch (error) {
        if (isStale()) return staleResult;
        return { status: false, message: error instanceof Error ? error.message : '文件库加载失败，请重试' };
      }
    },
    async initTagData(filesBasesId: string) {
      const requestSerial = ++this.tagDataRequestSerial;
      // 获取标签数据
      const includeResourceCount = this.currentConfigApp.showCustomTagResourceCount !== false;
      const tagDataResult = await tagServer.tagDataByFilesBasesId(filesBasesId, includeResourceCount)
      if (!tagDataResult.status) {
        return {
          status: false,
          message: tagDataResult.msg
        };
      }

      if (requestSerial !== this.tagDataRequestSerial || this.currentFilesBases.id !== filesBasesId) {
        return {
          status: true,
          message: 'stale tag request ignored'
        };
      }

      // 设置当前标签与标签分类
      this.currentTagClass = tagDataResult.data.tagClass;
      this.currentTag = tagDataResult.data.tag;
      return {
        status: true,
        message: 'success'
      };
    },
    async refreshCurrentTagData() {
      if (!this.currentFilesBases.id) return;
      try {
        return await this.initTagData(this.currentFilesBases.id);
      } catch (error) {
        console.warn('刷新自定义标签资源数量失败', error);
        return {
          status: false,
          message: 'refresh tag data failed'
        };
      }
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
