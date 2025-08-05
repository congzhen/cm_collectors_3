import type { E_resourceDramaSeriesType, E_resourceStorageLocation } from "./app.dataType";
import type { I_performer } from "./performer.dataType";
import type { I_tag } from "./tag.dataType";

export interface I_resource_base {
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
  abstract: string;
  status: boolean;
}

export interface I_resource extends I_resource_base {
  hot: number;
  lastPlayTime: string;
  lastPlayFile: string;
  addTime: string;
  directors: I_performer[];
  performers: I_performer[];
  tags: I_tag[];
  dramaSeries: I_resourceDramaSeries[];
}


export interface I_resourceDramaSeries_base {
  id: string;
  src: string;
}

export interface I_resourceDramaSeries extends I_resourceDramaSeries_base {
  resources_id: string;
  type: E_resourceDramaSeriesType;
  sort: number;
  storageLocation: E_resourceStorageLocation;
  m3u8BuilderTime: string;
  m3u8BuilderStatus: boolean;
}

export interface I_resourceDisplayTag {
  tag: I_tag;
  textColor: string;
  bgColor: string;
}

export interface I_dramaSeriesWithResource {
  id: string;
  resources_id: string;
  src: string;
  title: string;
}
