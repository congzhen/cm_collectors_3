import { E_detailsDramaSeriesMode, E_resourceOpenMode, E_resourceOpenMode_SoftType, type E_tagType } from "./app.dataType";

import dataset from "@/assets/dataset";

export enum E_config_type {
  app = 'app',
  importScanDisk = 'importScanDisk',
  importNfo = 'importNfo',
  importSimple = 'importSimple',
}


export interface I_config_app {
  country: string[]; // 国家
  coverDisplayTagAttribute: string[]; // 封面显示标签属性
  coverDisplayTag: string[]; // 封面显示标签
  coverDisplayTagColor: string; // 封面显示标签颜色
  coverDisplayTagColors: string[]; // 封面显示标签颜色
  coverDisplayTagRgba: string; // 封面显示标签颜色
  coverDisplayTagRgbas: string[]; // 封面显示标签颜色
  coverPosterData: I_coverPosterData[]; // 封面数据
  coverPosterDataDefaultSelect: number; // 封面数据默认选择
  coverPosterWidthBase: number; // 封面宽度
  coverPosterWidthStatus: boolean; // 封面宽度状态
  coverPosterHeightBase: number;  // 封面高度
  coverPosterHeightStatus: boolean; // 封面高度状态
  definition: string[]; // 清晰度
  definitionFontColor: string; // 清晰度颜色
  definitionRgba: string; // 清晰度颜色
  detailsDramaSeriesMode: E_detailsDramaSeriesMode; // 剧集显示模式
  director_Text: string; // 导演显示文字
  historyModule: boolean; // 历史记录是否开启
  historyNumber: number; // 历史记录数量
  hotModule: boolean; // 热门资源是否开启
  hotNumber: number; // 热门资源数量
  leftColumnMode: string; // 左侧栏模式
  leftColumnWidth: number; // 左侧栏宽度
  leftDisplay: E_tagType[]; // 左侧栏显示内容
  pageLimit: number; // 分页数量
  performerPhoto: boolean; // 左侧栏-演员图片是否开启
  performerPreferred: string[]; // 左侧栏-优先显示演员
  performerShowNum: number; // 左侧栏-演员显示数量
  performer_Text: string; // 演员显示名称
  performer_photo: string; // 默认演员照片
  playAtlasImageWidth: number; // 图集-图片宽度
  playAtlasMode: string; // 图集-模式
  playAtlasPageLimit: number; // 图集-分页数量
  playAtlasThumbnail: boolean; // 图集-缩略图
  playComicMode: string; // 漫画-模式
  playComicrReadingMode: boolean; // 漫画-阅读模式
  plugInUnit_Cup: boolean; // 插件单元-cup
  plugInUnit_Cup_Text: string; // 插件单元-cup-文字
  previewImageFolder: string; // 预览-图片文件夹
  randomPosterAutoSize: boolean; // 随机海报-自动大小
  randomPosterWidth: number; // 随机海报-宽度
  randomPosterHeight: number; // 随机海报-高度
  randomPosterPath: string; // 随机海报-路径
  randomPosterStatus: boolean; // 随机海报-状态
  resourceDetailsShowMode: string; // 资源详情-显示模式
  resourcesShowMode: string; // 资源-显示模式
  coverPosterBoxInfoWidth: number; // 资源-显示模式-封面海报盒子-信息宽度
  coverPosterWaterfallColumn: number; // 资源-显示模式-封面海报瀑布流-列数
  routeConversion: I_routeConversion[];// 路由转换
  shieldNoPerformerPhoto: boolean; // 屏蔽无照片演员
  showPreviewImage: boolean; // 显示预览图片
  sortMode: string; // 排序模式
  tagMode: string; // 标签模式
  youLikeModule: boolean; // 猜你喜欢-模块
  youLikeNumber: number; // 猜你喜欢-数量
  youLikeTagClass: string[]; // 猜你喜欢-标签类
  youLikeWordNumber: number; // 猜你喜欢-匹配词数量

  openResModeMovies: E_resourceOpenMode; // 视频 - 打开方式
  openResModeMovies_SoftType: E_resourceOpenMode_SoftType; // 软件内置播放器类型
  openResModeComic: E_resourceOpenMode;  // 漫画 - 打开方式
  openResModeAtlas: E_resourceOpenMode; // 图集 - 打开方式
}

export const defualtConfigApp: I_config_app = {
  leftDisplay: ['sort', 'country', 'definition', 'year', 'starRating', 'performer', 'diyTag'] as E_tagType[],
  leftColumnWidth: 319,
  leftColumnMode: 'fixed',
  country: ['China', 'Japan', 'SouthKorea', 'America', 'England', 'France', 'OtherCountry'],
  definition: ['8K', '4K', '2K', '1080P', '720P', 'HighDefinition', 'StandardDefinition'],
  tagMode: 'fixed',
  performerPhoto: true,
  shieldNoPerformerPhoto: true,
  performerShowNum: 12,
  performerPreferred: [],
  pageLimit: 32,
  sortMode: 'desc',
  resourcesShowMode: 'coverPoster',
  coverPosterBoxInfoWidth: 200,
  coverPosterWaterfallColumn: 8,
  detailsDramaSeriesMode: E_detailsDramaSeriesMode.fileName,
  historyModule: true,
  hotModule: true,
  youLikeModule: true,
  historyNumber: 10,
  hotNumber: 10,
  youLikeNumber: 10,
  youLikeWordNumber: 3,
  youLikeTagClass: [],
  resourceDetailsShowMode: 'right',
  showPreviewImage: false,
  previewImageFolder: '/',
  plugInUnit_Cup: false,
  plugInUnit_Cup_Text: 'Cup',
  coverPosterData: [
    { name: 'Default 1', width: 300, height: 420, type: 'default' },
    { name: 'Default 2', width: 400, height: 275, type: 'default' },
    { name: 'Default 3', width: 524, height: 270, type: 'default' },
    { name: 'Default 4', width: 480, height: 270, type: 'default' },
  ],
  coverPosterDataDefaultSelect: 0,
  coverPosterWidthStatus: false,
  coverPosterWidthBase: 316,
  coverPosterHeightStatus: true,
  coverPosterHeightBase: 218,
  playAtlasImageWidth: 150,
  playAtlasMode: 'waterfall',
  playAtlasPageLimit: 100,
  playAtlasThumbnail: true,
  playComicMode: 'fullScreen',
  playComicrReadingMode: false,
  routeConversion: [],
  definitionRgba: 'rgba(155, 88, 182,0.5)',
  definitionFontColor: '#F3F3F3',
  coverDisplayTagAttribute: [],
  coverDisplayTag: [],
  coverDisplayTagRgba: 'rgba(244, 54, 16,0.5)',
  coverDisplayTagRgbas: ['rgba(244, 54, 16, 0.75)'],
  coverDisplayTagColor: '#F3F3F3',
  coverDisplayTagColors: ['#F3F3F3'],
  randomPosterStatus: false,
  randomPosterPath: '',
  randomPosterAutoSize: false,
  randomPosterWidth: 156,
  randomPosterHeight: 218,
  performer_Text: '',
  director_Text: '',
  performer_photo: '',

  openResModeMovies: E_resourceOpenMode.Soft,
  openResModeMovies_SoftType: E_resourceOpenMode_SoftType.Windows,
  openResModeComic: E_resourceOpenMode.Soft,
  openResModeAtlas: E_resourceOpenMode.Soft,
}

export interface I_coverPosterData {
  name: string;
  width: number;
  height: number;
  type: string;
}

export interface I_routeConversion {
  from: string;
  to: string;
}


export interface I_config_scanDisk {
  scanDiskPaths: string[]; // 扫描磁盘路径
  videoSuffixName: string[]; // 视频文件后缀名
  resourceNamingMode: 'fileName' | 'dirName' | 'dirFileName'; // 资源命名模式
  coverPosterMatchName: string[]; // 封面海报匹配名称
  coverPosterFuzzyMatch: boolean; // 封面海报模糊匹配
  coverPosterUseRandomImageIfNoMatch: boolean; // 无匹配时使用目录下随机图片
  coverPosterSuffixName: string[]; // 封面海报文件后缀名
  coverPosterType: number; // 封面海报类型
  autoCreatePoster: boolean; // 自动创建封面海报
  checkPath: boolean; // 检查路径
}

export const defualtConfigScanDisk: I_config_scanDisk = {
  scanDiskPaths: [],
  videoSuffixName: dataset.videoSuffixName,
  coverPosterMatchName: dataset.coverPosterMatchName,
  coverPosterFuzzyMatch: true,
  coverPosterUseRandomImageIfNoMatch: false,
  resourceNamingMode: 'fileName',
  coverPosterSuffixName: dataset.imageSuffixName,
  coverPosterType: -1,
  autoCreatePoster: true,
  checkPath: true,
}


export interface I_config_nfo {
  abstract: string; // 描述
  autoConverSeries: boolean; // 自动转换剧集
  country: string; // 国家
  cover: string; // 封面
  coverPosterUsesPreSetDimensions: boolean; // 封面-使用预设尺寸
  coverSuffix: string; // 封面-后缀
  coverUrl: string; // 封面-URL
  importCheckTitleAlready: boolean; // 导入-检查标题是否已存在
  issueNumber: string; // 发行编号
  performer: string; // 演员
  performerName: string; // 演员-名称
  performerThumb: string; // 演员-缩略图
  removedTag: string; // 删除的标签
  root: string; // 根节点名称
  star: string; // 评分
  suffix: string; // 导入的后缀名
  tag: string; // 标签
  title: string; // 标题
  year: string; // 年份
}

export const defualtConfigNfo: I_config_nfo = {
  autoConverSeries: false,
  importCheckTitleAlready: true,
  coverPosterUsesPreSetDimensions: false,
  suffix: '.mp4|.avi|.rmvb|.wmv|.mov|.mkv|.flv|.ts|.webm|.iso|.mpg|.m4v',
  root: 'movie',
  title: 'originaltitle|title|sorttitle',
  issueNumber: 'num',
  year: 'releasedate|premiered|year',
  cover: 'poster|thumb|fanart',
  coverSuffix: 'jpg|jpeg|png',
  coverUrl: 'cover',
  tag: 'tag|genre',
  removedTag: '',
  abstract: 'outline|plot',
  country: 'country',
  star: 'star',
  performer: 'actor',
  performerName: 'name',
  performerThumb: 'thumb',
}

//简单导入配置
export interface I_config_simple {
  cover: string; // 封面
  coverPosterUsesPreSetDimensions: boolean; // 封面-使用预设尺寸
  coverSuffix: string; // 封面-后缀名
  importCheckTitleAlready: boolean; // 导入时检查标题是否已存在
  suffix: string; // 导入的后缀名
  title: string; // 标题
}
export const defualtConfigSimple: I_config_simple = {
  importCheckTitleAlready: true,
  coverPosterUsesPreSetDimensions: false,
  suffix: '.mp4|.avi|.rmvb|.wmv|.mov|.mkv|.flv|.ts|.webm|.iso|.mpg|.m4v',
  title: 'file|folder',
  cover: 'poster|thumb|fanart',
  coverSuffix: 'jpg|jpeg|png|webp',
}
