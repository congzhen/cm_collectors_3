export interface I_searchData {
  searchTextSlc: string[]; // 搜索文本
  sort: E_searchSort; // 排序
  country: I_searchGroup;  // 国家
  definition: I_searchGroup; // 清晰度
  year: I_searchGroup; // 年份
  star: I_searchGroup; // 评星
  performer: I_searchGroup; // 演员
  cup: I_searchGroup; // cup
  tag: Record<string, I_searchGroup>; // 标签
}

export interface I_searchGroup {
  logic: E_searchLogic;
  options: string[];
}

export enum E_searchLogic {
  Single = 'single',         // 单选
  MultiAnd = 'multiAnd',     // 多选（与）
  MultiOr = 'multiOr',       // 多选（或）
  Not = 'not'                // 非
}

export enum E_searchSort {
  AddTimeDesc = 'addTimeDesc', //时间倒叙
  AddTimeAsc = 'addTimeAsc', //时间正序
  IssuingDateDesc = 'issuingDateDesc', //发行时间倒叙
  IssuingDateAsc = 'issuingDateAsc', //发行时间正序
  IssueNumberDesc = 'issueNumberDesc', //发行倒叙
  IssueNumberAsc = 'issueNumberAsc',  //发行正序
  ScoreDesc = 'scoreDesc', //评分倒叙
  ScoreAsc = 'scoreAsc', //评分正序
  StarDesc = 'starDesc', //评星倒叙
  StarAsc = 'starAsc', //评星正序
  TitleDesc = 'titleDesc', //标题倒叙
  TitleAsc = 'titleAsc',  //标题正序
  History = 'history', //历史记录
  Hot = 'hot', //当前热度
  YouLike = 'youLike', //猜你喜欢
}
