import type { I_filesBases } from "./filesBases.dataType";
import type { I_performerBases } from "./performer.dataType";

//header模式
export enum E_headerMode {
  Index = 'index',
  GoBack = 'goBack',
}

// 资源显示模式
export enum E_resourceShowMode {
  Index = 'index',
  Name = 'name'
}


export enum E_detailsDramaSeriesMode {
  fileName = 'fileName',
  digit = 'digit',
}

// 资源剧集类型
export enum E_resourceDramaSeriesType {
  Movies = 'movies',              // 电影
  Comic = 'comic',                // 动漫
  Atlas = 'atlas',                // 画本
  Files = 'files',                // 文件
  VideoLink = 'videoLink',        // 视频链接
  NetDisk = 'netDisk',          //网盘
}

//资源打开方式
export enum E_resourceOpenMode {
  Soft = 'soft',
  CloundPlay = 'cloundPlay',
  System = 'system',
}
export enum E_resourceOpenMode_SoftType {
  Windows = 'windows',
  Dialog = 'dialog',
}

export enum E_resourceStorageLocation {
  Local = 'local',
  Server = 'server',
  NetDisk = 'netDisk',
}

// 标签类型
export enum E_tagType {
  Sort = 'sort',
  Country = 'country',
  Definition = 'definition',
  Year = 'year',
  Star = 'starRating',
  DiyTag = 'diyTag',
  Performer = 'performer',
  Cup = 'cup',
}

// performer 职业类型
export enum E_performerCareerType {
  All = 'all',
  Performer = 'performer',
  Director = 'director',
}

export type T_VideoPlayMode = 'mp4' | 'm3u8';

// 标签数据
export interface I_tagData {
  id: string;
  name: string;
  status: boolean;
}


export interface I_appData {
  appConfig: I_appConfig;
  filesBases: I_filesBases[];
  performerBases: I_performerBases[];
}


export interface I_appConfig {
  logoName: string;
  isAdminLogin: boolean;
  theme: string;
}

export interface I_appConfig_scraper {
  useBrowserPath: boolean;
  browserPath: string;

}

export interface I_appSystemVideoRateLimit {
  enabled: boolean;
  requestsPerSecond: number;
  burst: number;
}
export interface I_taryMenu {
  name: string;
  path: string;
}

export interface I_appSystemConfig extends I_appConfig {
  adminPassword: string;
  isAutoCreateM3u8: boolean;
  language: string;
  notAllowServerOpenFile: boolean;
  playVideoFormats: string[];
  playAudioFormats: string[];
  serverFileManagementRootPath: string[];
  videoRateLimit: I_appSystemVideoRateLimit;
  scraper: I_appConfig_scraper;
  taryMenu: I_taryMenu[];
}
export interface I_video_basic_info {
  width: number;
  height: number;
  duration: string;
  bit_rate: string;
  size: string;
}

export interface I_playVideoInfo {
  video_basic_info: I_video_basic_info;
  is_web: boolean;
}
export interface I_playVideoData {
  playUrl: string;
  playType: T_VideoPlayMode;
}
