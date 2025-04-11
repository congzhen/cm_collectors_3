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
