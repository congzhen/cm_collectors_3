import type { I_filesBases } from "./filesBases.dataType";

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

// 资源剧集类型
export enum E_resourceDramaSeriesType {
  Movies = 'movies',              // 电影
  Comic = 'comic',                // 动漫
  Atlas = 'atlas',                // 画本
  Files = 'files',                // 文件
  VideoLink = 'videoLink',        // 视频链接
  NetDisk = 'netDisk',          //网盘
}

// 标签类型
export enum E_tagType {
  DiyTag = 'diyTag',
  Performer = 'performer',
  Star = 'star',
}

// 演员职业
export enum E_performerCareer {
  Performer = 'performer',
  Director = 'director',
}

//cup
export interface I_cupBWH {
  bust: number;
  waist: string;
  hip: string;
}


export interface I_appData {
  filesBases: I_filesBases[];
}
