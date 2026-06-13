export interface I_aiTagSetting {
  id: string;
  enabled: boolean;
  paused: boolean;
  provider: string;
  baseUrl: string;
  apiKey: string;
  model: string;
  requestTimeoutSeconds: number;
  maxResourcesPerRun: number;
  maxFramesPerResource: number;
  maxFramesPerVideo: number;
  maxVideosPerResource: number;
  maxImagesPerAiRequest: number;
  frameStrategy: string;
  imageResizeMode: string;
  fallbackImageMaxWidth: number;
  imageJpegQuality: number;
  minConfidence: number;
  maxTagsPerResource: number;
  writeMode: string;
}

export interface I_aiTagFilesBasesSetting {
  filesBasesId: string;
  enabled: boolean;
  includeTagClassIds: string[];
  excludeTagClassIds: string[];
}

export interface I_aiTagStats {
  pending: number;
  processing: number;
  success: number;
  failed: number;
  skipped: number;
}

export interface I_aiTagRecord {
  id: string;
  resourcesId: string;
  filesBasesId: string;
  status: string;
  srcHash: string;
  tagVersionHash: string;
  recommendedTagIds: string;
  writtenTagIds: string;
  writtenTagNames: string[];
  writtenTagText: string;
  resultJson: string;
  evidenceJson: string;
  failReason: string;
  analyzedAt: string;
  filesBasesName: string;
  resourceTitle: string;
  resourceIssueNumber: string;
  resourceName: string;
  created_at: string;
  updated_at: string;
}

export interface I_aiTagRecordList {
  total: number;
  dataList: I_aiTagRecord[];
}

export interface I_aiTagResetResult {
  reset: number;
}

export interface I_aiTagRunResult {
  processed: number;
  success: number;
  skipped: number;
  failed: number;
  started: boolean;
  running: boolean;
}

export interface I_aiTagModelTestMetrics {
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
  usageReturned: boolean;
  elapsedMs: number;
  estimatedTokensPerSecond: number;
  serviceTokensPerSecond: number;
  servicePromptPerSecond: number;
  serviceGeneratedPerSecond: number;
}

export interface I_aiTagSettingRecommendation {
  field: keyof I_aiTagSetting;
  label: string;
  currentValue: string | number | boolean;
  recommendedValue: string | number | boolean;
  reason: string;
  impact: string;
}

export interface I_aiTagModelTestResult {
  success: boolean;
  model: string;
  endpoint: string;
  responseFormat: string;
  fallbackUsed: boolean;
  finishReason: string;
  summary: string;
  content: string;
  error: string;
  firstError: string;
  metrics: I_aiTagModelTestMetrics;
  recommendations: I_aiTagSettingRecommendation[];
}
