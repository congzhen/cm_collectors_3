export type VideoTranscodeStatus =
  | 'draft'
  | 'queued'
  | 'probing'
  | 'transcoding'
  | 'verifying'
  | 'replacing'
  | 'refreshing_metadata'
  | 'success'
  | 'failed'
  | 'cancelled'
  | 'interrupted'
  | 'rollback_failed';

export interface I_videoTranscodeConfig {
  container: 'mp4' | 'mkv';
  videoCodec: 'copy' | 'h264' | 'h265';
  qualityMode: 'crf' | 'bitrate';
  crf: number;
  videoBitrateKbps: number;
  preset: 'ultrafast' | 'fast' | 'medium' | 'slow';
  resolutionHeight: number;
  frameRate: number;
  audioCodec: 'copy' | 'aac';
  audioBitrateKbps: number;
  threads: number;
  keepBackup: boolean;
  gpuEncoder: '' | 'nvenc' | 'qsv' | 'amf';
}

export interface I_videoTranscodeTask {
  id: string;
  dramaSeriesId: string;
  resourceId: string;
  resourceTitle: string;
  sourcePath: string;
  sourceSize: number;
  sourceModifiedTime: number;
  sourceDuration: number;
  sourceWidth: number;
  sourceHeight: number;
  sourceFrameRate: number;
  sourceVideoCodec: string;
  sourceAudioCodec: string;
  sourceVideoBitRate: number;
  status: VideoTranscodeStatus;
  progress: number;
  processedSeconds: number;
  speed: string;
  temporaryPath: string;
  outputPath: string;
  backupPath: string;
  outputSize: number;
  outputDuration: number;
  outputWidth: number;
  outputHeight: number;
  outputFrameRate: number;
  outputVideoCodec: string;
  outputAudioCodec: string;
  outputVideoBitRate: number;
  errorMessage: string;
  warningMessage: string;
  createdAt: string;
  startedAt: string;
  finishedAt: string;
  config: I_videoTranscodeConfig;
  filesBasesId: string;
  coverPoster: string;
}

export interface I_videoTranscodeAddResult {
  added: number;
  skippedDuplicate: number;
  skippedMissing: number;
}

export interface I_videoTranscodeResetResult {
  reset: number;
  skipped: number;
}

export interface I_videoTranscodeQueueStatus {
  paused: boolean;
  currentId: string;
}

export interface I_videoTranscodeGPUEncoder {
  id: 'nvenc' | 'qsv' | 'amf';
  label: string;
  videoCodecs: Array<'h264' | 'h265'>;
}

export interface I_videoTranscodeCapabilities {
  gpuEncoders: I_videoTranscodeGPUEncoder[];
}
