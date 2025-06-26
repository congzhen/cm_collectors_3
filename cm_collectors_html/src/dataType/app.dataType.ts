import type { I_filesBases } from "./filesBases.dataType";
import type { I_performerBases } from "./performer.dataType";
import type { I_tag, I_tagClass } from "./tag.dataType";

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



// 标签数据
export interface I_tagData {
  id: string;
  name: string;
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
  performerBases: I_performerBases[];
}
