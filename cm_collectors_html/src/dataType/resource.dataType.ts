import type { E_resourceDramaSeriesType } from "./app.dataType";
import type { I_performer } from "./performer.dataType";
import type { I_tag } from "./tag.dataType";

export interface I_resource {
  id: string;
  filesBases_id: string;
  mode: E_resourceDramaSeriesType;
  title: string;
  issueNumber: string;
  coverPoster: string;
  coverPosterMode: number;
  coverPosterWidth: number;
  coverPosterHeight: number;
  issuingDate: string;
  country: string;
  definition: string;
  stars: number;
  hot: number;
  lastPlayTime: string;
  lastPlayFile: string;
  abstract: string;
  addTime: string;
  status: boolean;
  directors: I_performer[];
  performers: I_performer[];
  tags: I_tag[];
  dramaSeries: I_resourceDramaSeries[];
}


export interface I_resourceDramaSeries {
  id: string;
  resources_id: string;
  type: E_resourceDramaSeriesType;
  src: string;
  sort: number;
  m3u8BuilderTime: string;
  m3u8BuilderStatus: boolean;
}

export interface I_resourceDisplayTag {
  tag: I_tag;
  textColor: string;
  bgColor: string;
}
