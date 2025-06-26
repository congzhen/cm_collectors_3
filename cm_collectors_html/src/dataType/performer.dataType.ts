export interface I_performer {
  id: string;
  performerBases_id: string;
  name: string;
  aliasName: string;
  birthday: string;
  nationality: string;
  careerPerformer: boolean; // 职业演员
  careerDirector: boolean; // 职业导演
  photo: string; // 照片
  introduction: string; // 简介
  cup: string; // cup
  bust: string; // 胸围
  waist: string; // 腰围
  hip: string; // 臀围
  stars: number; // 评分
  retreatStatus: boolean; // 是否息影
  status: boolean;
}

export interface I_search_performer {
  search: string;
  star: string;
  cup: string;
}


export interface I_performerBases {
  id: string;
  name: string;
  sort: number;
  status: boolean;
}
