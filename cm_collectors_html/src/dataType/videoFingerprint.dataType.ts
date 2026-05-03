export interface I_VideoFingerprintStats {
  total: number;
  pending: number;
  done: number;
  failed: number;
  drama_total: number;
}

export interface I_VideoFingerprintTaskStatus {
  running: boolean;
  files_bases_id: string;
  started_at: string;
  last_finished_at: string;
  last_error: string;
  last_success: number;
  last_failed: number;
}

export interface I_DuplicateItem {
  drama_series_id: string;
  resources_id: string;
  files_bases_id: string;
  src: string;
  resource_title: string;
  cover_poster: string;
  duration: number;
}

export interface I_DuplicateGroup {
  items: I_DuplicateItem[];
  avg_distance: number;
  matched_count: number;
}

export interface I_DuplicateResult {
  dataList: I_DuplicateGroup[];
  total: number;
}
