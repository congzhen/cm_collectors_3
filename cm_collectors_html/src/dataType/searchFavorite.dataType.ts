import type { I_searchData } from './search.dataType';

export interface I_searchFavorite {
  id: string;
  filesBasesId: string;
  searchData: I_searchData;
  schemaVersion: number;
  sort: number;
  invalidConditions: number;
  optionLabels: Record<string, string>;
  createdAt: string;
  updatedAt: string;
}

export interface I_searchFavoriteCreate {
  filesBasesId: string;
  searchData: I_searchData;
}

export interface I_searchFavoriteUpdate {
  searchData?: I_searchData;
}
