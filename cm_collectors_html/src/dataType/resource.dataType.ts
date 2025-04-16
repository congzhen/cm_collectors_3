import type { E_resourceDramaSeriesType } from "./app.dataType";

export interface I_resource {
  id: string;
  files_bases_id: string;
  mode: E_resourceDramaSeriesType;
  title: string;
  issue_number: string;
  cover_poster: string;
  cover_poster_mode: number;
  cover_poster_width: string;
  cover_poster_height: string;
  issuing_date: string;
  country: string;
  definition: string;
  stars: number;
  abstract: string;
  tags: { [key: string]: Array<string> };
  directors: string[];
  performers: string[];
  drama_series: I_resourceDramaSeries[];
}


export interface I_resourceDramaSeries {
  id: string;
  resources_id: string;
  type: E_resourceDramaSeriesType;
  src: string;
}
