export type VideoMetadataScopeMode = 'selected' | 'all';
export type VideoMetadataRunMode = 'missing' | 'missing_stale' | 'failed' | 'failed_force' | 'all';

export interface I_videoMetadata {
  dramaSeriesId: string;
  metadataVersion: number;
  probeStatus: 'processing' | 'success' | 'failed' | 'stale';
  probeTime: string;
  nextRetryTime: string;
  retryCount: number;
  errorCode: string;
  errorMessage: string;
  fileSize: number;
  fileModifiedTime: number;
  width: number;
  height: number;
  frameRate: number;
  frameRateRaw: string;
  videoCodec: string;
  videoProfile: string;
  pixelFormat: string;
  bitDepth: number;
  videoBitRate: number;
  containerFormat: string;
  audioCodec: string;
  audioChannels: number;
  audioSampleRate: number;
}

export interface I_videoMetadataSetting {
  id: string;
  collectOnNewOrChanged: boolean;
  collectOnDetailOrPlay: boolean;
  collectOnList: boolean;
  idleBackfillEnabled: boolean;
  idleScopeMode: VideoMetadataScopeMode;
  idleWaitMinutes: number;
  probeIntervalMilliseconds: number;
  idleBatchSize: number;
  paused: boolean;
}

export interface I_videoMetadataSettingData {
  setting: I_videoMetadataSetting;
  filesBasesIds: string[];
}

export interface I_videoMetadataStats {
  filesBasesId: string;
  name: string;
  total: number;
  completed: number;
  pending: number;
  failed: number;
  processing: number;
  stale: number;
}

export interface I_videoMetadataRunRequest {
  scopeMode: VideoMetadataScopeMode;
  filesBasesIds: string[];
  runMode: VideoMetadataRunMode;
  maxItemsPerRun?: number;
}

export interface I_videoMetadataBatchTask {
  id: string;
  scopeMode: VideoMetadataScopeMode;
  runMode: VideoMetadataRunMode;
  status: 'running' | 'paused' | 'stopped' | 'completed' | '';
  total: number;
  success: number;
  failed: number;
  skipped: number;
  currentSrc: string;
  lastError: string;
  createdAt: string;
  startedAt: string;
  finishedAt: string;
}
