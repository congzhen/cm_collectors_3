export type PerformerAvatarStrategy = 'recommended' | 'original' | 'aiFix';
export type PerformerAvatarDialogStrategy = PerformerAvatarStrategy | 'manual';

export interface I_performerAvatarLibrarySetting {
  customBaseUrl: string;
  defaultStrategy: PerformerAvatarStrategy;
}

export interface I_performerAvatarLibraryConfig extends I_performerAvatarLibrarySetting {
  cachePath: string;
}

export interface I_performerAvatarLibraryStatus {
  ready: boolean;
  updating: boolean;
  fileSize: number;
  updatedAt: string;
  dataTimestamp: string;
  totalNum: string;
  totalSize: string;
  activeBaseUrl: string;
  setting: I_performerAvatarLibrarySetting;
}

export interface I_performerAvatarCandidate {
  id: string;
  source: string;
  fileName: string;
  aiFixed: boolean;
  rank: number;
}

export interface I_performerAvatarBatchPreview {
  total: number;
  matched: number;
  unmatched: number;
  skippedExisting: number;
  multipleCandidates: number;
}

export interface I_performerAvatarBatchProgress {
  batchId: string;
  total: number;
  completed: number;
  matched: number;
  success: number;
  failed: number;
  unmatched: number;
  skippedExisting: number;
  multipleCandidates: number;
  currentActors: string[];
  failures: Array<{ performerId: string; name: string; error: string }>;
  done: boolean;
}

export interface I_performerAvatarBatchActor {
  id: string;
  performerBasesId: string;
  name: string;
  aliasName: string;
  photo: string;
  hasPhoto: boolean;
}

export interface I_performerAvatarBatchActorPage {
  dataList: I_performerAvatarBatchActor[];
  total: number;
}
